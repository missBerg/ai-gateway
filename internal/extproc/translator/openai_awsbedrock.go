// Copyright Envoy AI Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"k8s.io/utils/ptr"

	"github.com/envoyproxy/ai-gateway/internal/apischema/awsbedrock"
	"github.com/envoyproxy/ai-gateway/internal/apischema/openai"
	"github.com/envoyproxy/ai-gateway/internal/internalapi"
	tracing "github.com/envoyproxy/ai-gateway/internal/tracing/api"
)

// NewChatCompletionOpenAIToAWSBedrockTranslator implements [Factory] for OpenAI to AWS Bedrock translation.
func NewChatCompletionOpenAIToAWSBedrockTranslator(modelNameOverride internalapi.ModelNameOverride) OpenAIChatCompletionTranslator {
	return &openAIToAWSBedrockTranslatorV1ChatCompletion{modelNameOverride: modelNameOverride}
}

// openAIToAWSBedrockTranslator translates OpenAI Chat Completions API requests to AWS Bedrock Converse API.
// Note: This uses the Converse API directly, not Bedrock's OpenAI-compatible API:
// https://docs.aws.amazon.com/bedrock/latest/APIReference/API_runtime_Converse.html
type openAIToAWSBedrockTranslatorV1ChatCompletion struct {
	modelNameOverride internalapi.ModelNameOverride
	stream            bool
	bufferedBody      []byte
	events            []awsbedrock.ConverseStreamEvent
	// role is from MessageStartEvent in chunked messages, and used for all openai chat completion chunk choices.
	// Translator is created for each request/response stream inside external processor, accordingly the role is not reused by multiple streams.
	role         string
	requestModel internalapi.RequestModel
}

// RequestBody implements [OpenAIChatCompletionTranslator.RequestBody].
func (o *openAIToAWSBedrockTranslatorV1ChatCompletion) RequestBody(_ []byte, openAIReq *openai.ChatCompletionRequest, _ bool) (
	headerMutation *extprocv3.HeaderMutation, bodyMutation *extprocv3.BodyMutation, err error,
) {
	var pathTemplate string
	if openAIReq.Stream {
		o.stream = true
		pathTemplate = "/model/%s/converse-stream"
	} else {
		pathTemplate = "/model/%s/converse"
	}

	o.requestModel = openAIReq.Model
	if o.modelNameOverride != "" {
		// Use modelName override if set.
		o.requestModel = o.modelNameOverride
	}

	headerMutation = &extprocv3.HeaderMutation{
		SetHeaders: []*corev3.HeaderValueOption{
			{Header: &corev3.HeaderValue{
				Key:      ":path",
				RawValue: fmt.Appendf(nil, pathTemplate, o.requestModel),
			}},
		},
	}

	var bedrockReq awsbedrock.ConverseInput
	// Convert InferenceConfiguration.
	bedrockReq.InferenceConfig = &awsbedrock.InferenceConfiguration{}
	bedrockReq.InferenceConfig.Temperature = openAIReq.Temperature
	bedrockReq.InferenceConfig.TopP = openAIReq.TopP

	bedrockReq.InferenceConfig.MaxTokens = cmp.Or(openAIReq.MaxCompletionTokens, openAIReq.MaxTokens)

	stopSequence, err := processStop(openAIReq.Stop)
	if err != nil {
		return
	}
	if len(stopSequence) > 0 {
		bedrockReq.InferenceConfig.StopSequences = stopSequence
	}

	// Handle Anthropic vendor fields if present. Currently only supports thinking fields.
	if openAIReq.AnthropicVendorFields != nil && openAIReq.Thinking != nil {
		if bedrockReq.AdditionalModelRequestFields == nil {
			bedrockReq.AdditionalModelRequestFields = make(map[string]interface{})
		}
		bedrockReq.AdditionalModelRequestFields["thinking"] = openAIReq.Thinking
	}

	// Convert Chat Completion messages.
	err = o.openAIMessageToBedrockMessage(openAIReq, &bedrockReq)
	if err != nil {
		return nil, nil, err
	}
	// Convert ToolConfiguration.
	if len(openAIReq.Tools) > 0 {
		err = o.openAIToolsToBedrockToolConfiguration(openAIReq, &bedrockReq)
		if err != nil {
			return nil, nil, err
		}
	}

	mut := &extprocv3.BodyMutation_Body{}
	if mut.Body, err = json.Marshal(bedrockReq); err != nil {
		return nil, nil, fmt.Errorf("failed to marshal body: %w", err)
	}
	setContentLength(headerMutation, mut.Body)
	return headerMutation, &extprocv3.BodyMutation{Mutation: mut}, nil
}

// openAIToolsToBedrockToolConfiguration converts openai ChatCompletion tools to aws bedrock tool configurations.
func (o *openAIToAWSBedrockTranslatorV1ChatCompletion) openAIToolsToBedrockToolConfiguration(openAIReq *openai.ChatCompletionRequest,
	bedrockReq *awsbedrock.ConverseInput,
) error {
	bedrockReq.ToolConfig = &awsbedrock.ToolConfiguration{}
	tools := make([]*awsbedrock.Tool, 0, len(openAIReq.Tools))
	for i := range openAIReq.Tools {
		toolDefinition := &openAIReq.Tools[i]
		if toolDefinition.Function != nil {
			toolName := toolDefinition.Function.Name
			var toolDesc *string
			if toolDefinition.Function.Description != "" {
				toolDesc = &toolDefinition.Function.Description
			}
			tool := &awsbedrock.Tool{
				ToolSpec: &awsbedrock.ToolSpecification{
					Name:        &toolName,
					Description: toolDesc,
					InputSchema: &awsbedrock.ToolInputSchema{
						JSON: toolDefinition.Function.Parameters,
					},
				},
			}
			tools = append(tools, tool)
		}
	}
	bedrockReq.ToolConfig.Tools = tools

	if openAIReq.ToolChoice != nil {
		if toolChoice, ok := openAIReq.ToolChoice.(string); ok {
			switch toolChoice {
			case "auto":
				bedrockReq.ToolConfig.ToolChoice = &awsbedrock.ToolChoice{
					Auto: &awsbedrock.AutoToolChoice{},
				}
			case "required":
				bedrockReq.ToolConfig.ToolChoice = &awsbedrock.ToolChoice{
					Any: &awsbedrock.AnyToolChoice{},
				}
			default:
				// Anthropic Claude supports tool_choice parameter with three options.
				// * `auto` allows Claude to decide whether to call any provided tools or not.
				// * `any` tells Claude that it must use one of the provided tools, but doesn't force a particular tool.
				// * `tool` allows us to force Claude to always use a particular tool.
				// The tool option is only applied to Anthropic Claude.
				if strings.Contains(openAIReq.Model, "anthropic") && strings.Contains(openAIReq.Model, "claude") {
					bedrockReq.ToolConfig.ToolChoice = &awsbedrock.ToolChoice{
						Tool: &awsbedrock.SpecificToolChoice{
							Name: &toolChoice,
						},
					}
				}
			}
		} else if toolChoice, ok := openAIReq.ToolChoice.(openai.ToolChoice); ok {
			tool := string(toolChoice.Type)
			bedrockReq.ToolConfig.ToolChoice = &awsbedrock.ToolChoice{
				Tool: &awsbedrock.SpecificToolChoice{
					Name: &tool,
				},
			}
		} else {
			return fmt.Errorf("unexpected type: %T", openAIReq.ToolChoice)
		}
	}
	return nil
}

// openAIMessageToBedrockMessageRoleUser converts openai user role message.
func (o *openAIToAWSBedrockTranslatorV1ChatCompletion) openAIMessageToBedrockMessageRoleUser(
	openAiMessage *openai.ChatCompletionUserMessageParam, role string,
) (*awsbedrock.Message, error) {
	if v, ok := openAiMessage.Content.Value.(string); ok {
		return &awsbedrock.Message{
			Role: role,
			Content: []*awsbedrock.ContentBlock{
				{Text: ptr.To(v)},
			},
		}, nil
	} else if contents, ok := openAiMessage.Content.Value.([]openai.ChatCompletionContentPartUserUnionParam); ok {
		chatMessage := &awsbedrock.Message{Role: role}
		chatMessage.Content = make([]*awsbedrock.ContentBlock, 0, len(contents))
		for i := range contents {
			contentPart := &contents[i]
			if contentPart.OfText != nil {
				textContentPart := contentPart.OfText
				chatMessage.Content = append(chatMessage.Content, &awsbedrock.ContentBlock{
					Text: &textContentPart.Text,
				})
			} else if contentPart.OfImageURL != nil {
				imageContentPart := contentPart.OfImageURL
				contentType, b, err := parseDataURI(imageContentPart.ImageURL.URL)
				if err != nil {
					return nil, fmt.Errorf("failed to parse image URL: %s %w", imageContentPart.ImageURL.URL, err)
				}
				var format string
				switch contentType {
				case mimeTypeImagePNG:
					format = "png"
				case mimeTypeImageJPEG:
					format = "jpeg"
				case mimeTypeImageGIF:
					format = "gif"
				case mimeTypeImageWEBP:
					format = "webp"
				default:
					return nil, fmt.Errorf("unsupported image type: %s please use one of [png, jpeg, gif, webp]",
						contentType)
				}

				chatMessage.Content = append(chatMessage.Content, &awsbedrock.ContentBlock{
					Image: &awsbedrock.ImageBlock{
						Format: format,
						Source: awsbedrock.ImageSource{
							Bytes: b, // Decoded data as bytes.
						},
					},
				})
			}
		}
		return chatMessage, nil
	}
	return nil, fmt.Errorf("unexpected content type")
}

// unmarshalToolCallArguments is a helper method to unmarshal tool call arguments.
func unmarshalToolCallArguments(arguments string) (map[string]any, error) {
	var input map[string]any
	if err := json.Unmarshal([]byte(arguments), &input); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool call arguments: %w", err)
	}
	return input, nil
}

// openAIMessageToBedrockMessageRoleAssistant converts openai assistant role message
// The tool content is appended to the bedrock message content list if tool_call is in openai message.
func (o *openAIToAWSBedrockTranslatorV1ChatCompletion) openAIMessageToBedrockMessageRoleAssistant(
	openAiMessage *openai.ChatCompletionAssistantMessageParam, role string,
) (*awsbedrock.Message, error) {
	bedrockMessage := &awsbedrock.Message{Role: role}
	contentBlocks := make([]*awsbedrock.ContentBlock, 0)

	var contentParts []openai.ChatCompletionAssistantMessageParamContent
	if v, ok := openAiMessage.Content.Value.(string); ok && len(v) > 0 {
		// Case 1: Content is a simple string.
		contentParts = append(contentParts, openai.ChatCompletionAssistantMessageParamContent{Type: openai.ChatCompletionAssistantMessageParamContentTypeText, Text: &v})
	} else if singleContent, ok := openAiMessage.Content.Value.(openai.ChatCompletionAssistantMessageParamContent); ok {
		// Case 2: Content is a single object.
		contentParts = append(contentParts, singleContent)
	} else if sliceContent, ok := openAiMessage.Content.Value.([]openai.ChatCompletionAssistantMessageParamContent); ok {
		// Case 3: Content is already a slice of objects.
		contentParts = sliceContent
	}

	for _, content := range contentParts {
		switch content.Type {
		case openai.ChatCompletionAssistantMessageParamContentTypeText:
			if content.Text != nil {
				contentBlocks = append(contentBlocks, &awsbedrock.ContentBlock{Text: content.Text})
			}
		case openai.ChatCompletionAssistantMessageParamContentTypeThinking:
			if content.Text != nil {
				reasoningText := &awsbedrock.ReasoningTextBlock{
					Text: *content.Text,
				}
				if content.Signature != nil {
					reasoningText.Signature = *content.Signature
				}
				contentBlocks = append(contentBlocks, &awsbedrock.ContentBlock{
					ReasoningContent: &awsbedrock.ReasoningContentBlock{
						ReasoningText: reasoningText,
					},
				})
			}
		case openai.ChatCompletionAssistantMessageParamContentTypeRedactedThinking:
			if content.RedactedContent != nil {
				contentBlocks = append(contentBlocks, &awsbedrock.ContentBlock{
					ReasoningContent: &awsbedrock.ReasoningContentBlock{
						RedactedContent: content.RedactedContent,
					},
				})
			}
		case openai.ChatCompletionAssistantMessageParamContentTypeRefusal:
			if content.Refusal != nil {
				contentBlocks = append(contentBlocks, &awsbedrock.ContentBlock{Text: content.Refusal})
			}
		}
	}

	bedrockMessage.Content = contentBlocks

	for i := range openAiMessage.ToolCalls {
		toolCall := &openAiMessage.ToolCalls[i]
		input, err := unmarshalToolCallArguments(toolCall.Function.Arguments)
		if err != nil {
			return nil, err
		}
		bedrockMessage.Content = append(bedrockMessage.Content,
			&awsbedrock.ContentBlock{
				ToolUse: &awsbedrock.ToolUseBlock{
					Name:      toolCall.Function.Name,
					ToolUseID: *toolCall.ID,
					Input:     input,
				},
			})
	}
	return bedrockMessage, nil
}

// openAIMessageToBedrockMessageRoleSystem converts openai system role message.
func (o *openAIToAWSBedrockTranslatorV1ChatCompletion) openAIMessageToBedrockMessageRoleSystem(
	openAiMessage *openai.ChatCompletionSystemMessageParam, bedrockSystem *[]*awsbedrock.SystemContentBlock,
) error {
	if v, ok := openAiMessage.Content.Value.(string); ok {
		*bedrockSystem = append(*bedrockSystem, &awsbedrock.SystemContentBlock{
			Text: v,
		})
	} else if contents, ok := openAiMessage.Content.Value.([]openai.ChatCompletionContentPartTextParam); ok {
		for i := range contents {
			contentPart := &contents[i]
			textContentPart := contentPart.Text
			*bedrockSystem = append(*bedrockSystem, &awsbedrock.SystemContentBlock{
				Text: textContentPart,
			})
		}
	} else {
		return fmt.Errorf("unexpected content type for system message")
	}
	return nil
}

// openAIMessageToBedrockMessageRoleTool converts openai tool role message.
func (o *openAIToAWSBedrockTranslatorV1ChatCompletion) openAIMessageToBedrockMessageRoleTool(
	openAiMessage *openai.ChatCompletionToolMessageParam, role string,
) (*awsbedrock.Message, error) {
	// Validate and cast the openai content value into bedrock content block.
	content := make([]*awsbedrock.ToolResultContentBlock, 0)

	switch v := openAiMessage.Content.Value.(type) {
	case string:
		content = []*awsbedrock.ToolResultContentBlock{
			{
				Text: &v,
			},
		}
	case []openai.ChatCompletionContentPartTextParam:
		for _, part := range v {
			content = append(content, &awsbedrock.ToolResultContentBlock{
				Text: &part.Text,
			})
		}

	default:
		return nil, fmt.Errorf("unexpected content type for tool message: %T", openAiMessage.Content.Value)
	}

	return &awsbedrock.Message{
		Role: role,
		Content: []*awsbedrock.ContentBlock{
			{
				ToolResult: &awsbedrock.ToolResultBlock{
					Content:   content,
					ToolUseID: &openAiMessage.ToolCallID,
				},
			},
		},
	}, nil
}

// openAIMessageToBedrockMessage converts openai ChatCompletion messages to aws bedrock messages.
func (o *openAIToAWSBedrockTranslatorV1ChatCompletion) openAIMessageToBedrockMessage(openAIReq *openai.ChatCompletionRequest,
	bedrockReq *awsbedrock.ConverseInput,
) error {
	// Convert Messages.
	bedrockReq.Messages = make([]*awsbedrock.Message, 0, len(openAIReq.Messages))
	openAIReqMessageLen, i := len(openAIReq.Messages), 0
	for i < openAIReqMessageLen {
		msg := &openAIReq.Messages[i]
		role := msg.ExtractMessgaeRole()
		switch {
		case msg.OfUser != nil:
			userMessage := msg.OfUser
			bedrockMessage, err := o.openAIMessageToBedrockMessageRoleUser(userMessage, role)
			if err != nil {
				return err
			}
			bedrockReq.Messages = append(bedrockReq.Messages, bedrockMessage)
		case msg.OfAssistant != nil:
			assistantMessage := msg.OfAssistant
			bedrockMessage, err := o.openAIMessageToBedrockMessageRoleAssistant(assistantMessage, role)
			if err != nil {
				return err
			}
			bedrockReq.Messages = append(bedrockReq.Messages, bedrockMessage)
		case msg.OfSystem != nil:
			if bedrockReq.System == nil {
				bedrockReq.System = make([]*awsbedrock.SystemContentBlock, 0)
			}
			systemMessage := msg.OfSystem
			err := o.openAIMessageToBedrockMessageRoleSystem(systemMessage, &bedrockReq.System)
			if err != nil {
				return err
			}
		case msg.OfDeveloper != nil:
			message := msg.OfDeveloper
			if bedrockReq.System == nil {
				bedrockReq.System = []*awsbedrock.SystemContentBlock{}
			}

			if text, ok := message.Content.Value.(string); ok {
				bedrockReq.System = append(bedrockReq.System, &awsbedrock.SystemContentBlock{
					Text: text,
				})
			} else {
				if contents, ok := message.Content.Value.([]openai.ChatCompletionContentPartTextParam); ok {
					for i := range contents {
						contentPart := &contents[i]
						textContentPart := contentPart.Text
						bedrockReq.System = append(bedrockReq.System, &awsbedrock.SystemContentBlock{
							Text: textContentPart,
						})
					}
				} else {
					return fmt.Errorf("unexpected content type for developer message")
				}
			}
		case msg.OfTool != nil:
			toolMessage := msg.OfTool
			// Bedrock does not support tool role, merging to the user role.
			bedrockMessage, err := o.openAIMessageToBedrockMessageRoleTool(toolMessage, awsbedrock.ConversationRoleUser)
			if err != nil {
				return err
			}
			// Coalesce consecutive tool messages following a user message.
			for i+1 < openAIReqMessageLen {
				nextMessage := &openAIReq.Messages[i+1]
				if nextMessage.ExtractMessgaeRole() != openai.ChatMessageRoleTool {
					break
				}

				nextToolMessage := nextMessage.OfTool
				nextBedrockMessage, err := o.openAIMessageToBedrockMessageRoleTool(nextToolMessage, awsbedrock.ConversationRoleUser)
				if err != nil {
					return err
				}
				if len(nextBedrockMessage.Content) > 0 {
					bedrockMessage.Content = append(bedrockMessage.Content, nextBedrockMessage.Content[0])
				}
				i++
			}

			bedrockReq.Messages = append(bedrockReq.Messages, bedrockMessage)
		default:
			return fmt.Errorf("unexpected role: %s", msg.ExtractMessgaeRole())
		}

		i++
	}
	return nil
}

// ResponseHeaders implements [OpenAIChatCompletionTranslator.ResponseHeaders].
func (o *openAIToAWSBedrockTranslatorV1ChatCompletion) ResponseHeaders(headers map[string]string) (
	headerMutation *extprocv3.HeaderMutation, err error,
) {
	if o.stream {
		contentType := headers["content-type"]
		if contentType == "application/vnd.amazon.eventstream" {
			// We need to change the content-type to text/event-stream for streaming responses.
			return &extprocv3.HeaderMutation{
				SetHeaders: []*corev3.HeaderValueOption{
					{Header: &corev3.HeaderValue{Key: "content-type", Value: "text/event-stream"}},
				},
			}, nil
		}
	}
	return nil, nil
}

func (o *openAIToAWSBedrockTranslatorV1ChatCompletion) bedrockStopReasonToOpenAIStopReason(
	stopReason *string,
) openai.ChatCompletionChoicesFinishReason {
	if stopReason == nil {
		return openai.ChatCompletionChoicesFinishReasonStop
	}

	switch *stopReason {
	case awsbedrock.StopReasonStopSequence, awsbedrock.StopReasonEndTurn:
		return openai.ChatCompletionChoicesFinishReasonStop
	case awsbedrock.StopReasonMaxTokens:
		return openai.ChatCompletionChoicesFinishReasonLength
	case awsbedrock.StopReasonContentFiltered:
		return openai.ChatCompletionChoicesFinishReasonContentFilter
	case awsbedrock.StopReasonToolUse:
		return openai.ChatCompletionChoicesFinishReasonToolCalls
	default:
		return openai.ChatCompletionChoicesFinishReasonStop
	}
}

func (o *openAIToAWSBedrockTranslatorV1ChatCompletion) bedrockToolUseToOpenAICalls(
	toolUse *awsbedrock.ToolUseBlock,
) *openai.ChatCompletionMessageToolCallParam {
	if toolUse == nil {
		return nil
	}
	arguments, err := json.Marshal(toolUse.Input)
	if err != nil {
		return nil
	}
	return &openai.ChatCompletionMessageToolCallParam{
		ID: &toolUse.ToolUseID,
		Function: openai.ChatCompletionMessageToolCallFunctionParam{
			Name:      toolUse.Name,
			Arguments: string(arguments),
		},
		Type: openai.ChatCompletionMessageToolCallTypeFunction,
	}
}

// ResponseError implements [OpenAIChatCompletionTranslator.ResponseError].
// Translate AWS Bedrock exceptions to OpenAI error type.
// The error type is stored in the "x-amzn-errortype" HTTP header for AWS error responses.
// If AWS Bedrock connection fails the error body is translated to OpenAI error type for events such as HTTP 503 or 504.
func (o *openAIToAWSBedrockTranslatorV1ChatCompletion) ResponseError(respHeaders map[string]string, body io.Reader) (
	headerMutation *extprocv3.HeaderMutation, bodyMutation *extprocv3.BodyMutation, err error,
) {
	statusCode := respHeaders[statusHeaderName]
	var openaiError openai.Error
	if v, ok := respHeaders[contentTypeHeaderName]; ok && v == jsonContentType {
		var bedrockError awsbedrock.BedrockException
		if err = json.NewDecoder(body).Decode(&bedrockError); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal error body: %w", err)
		}
		openaiError = openai.Error{
			Type: "error",
			Error: openai.ErrorType{
				Type:    respHeaders[awsErrorTypeHeaderName],
				Message: bedrockError.Message,
				Code:    &statusCode,
			},
		}
	} else {
		var buf []byte
		buf, err = io.ReadAll(body)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read error body: %w", err)
		}
		openaiError = openai.Error{
			Type: "error",
			Error: openai.ErrorType{
				Type:    awsBedrockBackendError,
				Message: string(buf),
				Code:    &statusCode,
			},
		}
	}
	mut := &extprocv3.BodyMutation_Body{}
	mut.Body, err = json.Marshal(openaiError)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal error body: %w", err)
	}
	headerMutation = &extprocv3.HeaderMutation{}
	setContentLength(headerMutation, mut.Body)
	return headerMutation, &extprocv3.BodyMutation{Mutation: mut}, nil
}

// ResponseBody implements [OpenAIChatCompletionTranslator.ResponseBody].
// AWS Bedrock uses static model execution without virtualization, where the requested model
// is exactly what gets executed. The response does not contain a model field, so we return
// the request model that was originally sent.
func (o *openAIToAWSBedrockTranslatorV1ChatCompletion) ResponseBody(_ map[string]string, body io.Reader, endOfStream bool, span tracing.ChatCompletionSpan) (
	headerMutation *extprocv3.HeaderMutation, bodyMutation *extprocv3.BodyMutation, tokenUsage LLMTokenUsage, responseModel string, err error,
) {
	mut := &extprocv3.BodyMutation_Body{}
	if o.stream {
		var buf []byte
		buf, err = io.ReadAll(body)
		if err != nil {
			return nil, nil, LLMTokenUsage{}, "", fmt.Errorf("failed to read body: %w", err)
		}
		o.bufferedBody = append(o.bufferedBody, buf...)
		o.extractAmazonEventStreamEvents()

		for i := range o.events {
			event := &o.events[i]
			if usage := event.Usage; usage != nil {
				tokenUsage = LLMTokenUsage{
					InputTokens:  uint32(usage.InputTokens),  //nolint:gosec
					OutputTokens: uint32(usage.OutputTokens), //nolint:gosec
					TotalTokens:  uint32(usage.TotalTokens),  //nolint:gosec
				}
			}
			oaiEvent, ok := o.convertEvent(event)
			if !ok {
				continue
			}
			var oaiEventBytes []byte
			oaiEventBytes, err = json.Marshal(oaiEvent)
			if err != nil {
				panic(fmt.Errorf("failed to marshal event: %w", err))
			}
			if span != nil {
				span.RecordResponseChunk(oaiEvent)
			}
			mut.Body = append(mut.Body, []byte("data: ")...)
			mut.Body = append(mut.Body, oaiEventBytes...)
			mut.Body = append(mut.Body, []byte("\n\n")...)
		}

		if endOfStream {
			mut.Body = append(mut.Body, []byte("data: [DONE]\n")...)
		}
		return headerMutation, &extprocv3.BodyMutation{Mutation: mut}, tokenUsage, o.requestModel, nil
	}

	var bedrockResp awsbedrock.ConverseResponse
	if err = json.NewDecoder(body).Decode(&bedrockResp); err != nil {
		return nil, nil, LLMTokenUsage{}, "", fmt.Errorf("failed to unmarshal body: %w", err)
	}
	openAIResp := &openai.ChatCompletionResponse{
		Object:  "chat.completion",
		Choices: make([]openai.ChatCompletionResponseChoice, 0),
	}
	// Convert token usage.
	if bedrockResp.Usage != nil {
		tokenUsage = LLMTokenUsage{
			InputTokens:  uint32(bedrockResp.Usage.InputTokens),  //nolint:gosec
			OutputTokens: uint32(bedrockResp.Usage.OutputTokens), //nolint:gosec
			TotalTokens:  uint32(bedrockResp.Usage.TotalTokens),  //nolint:gosec
		}
		openAIResp.Usage = openai.ChatCompletionResponseUsage{
			TotalTokens:      bedrockResp.Usage.TotalTokens,
			PromptTokens:     bedrockResp.Usage.InputTokens,
			CompletionTokens: bedrockResp.Usage.OutputTokens,
		}
	}

	// AWS Bedrock does not support N(multiple choices) > 0, so there could be only one choice.
	choice := openai.ChatCompletionResponseChoice{
		Index: (int64)(0),
		Message: openai.ChatCompletionResponseChoiceMessage{
			Role: bedrockResp.Output.Message.Role,
		},
		FinishReason: o.bedrockStopReasonToOpenAIStopReason(bedrockResp.StopReason),
	}

	for _, output := range bedrockResp.Output.Message.Content {
		// The AWS Content Block data type is a UNION,
		// so only one of the members can be specified when used or returned.
		// see: https: //docs.aws.amazon.com/bedrock/latest/APIReference/API_runtime_ContentBlock.html
		switch {
		case output.ToolUse != nil:
			toolCall := o.bedrockToolUseToOpenAICalls(output.ToolUse)
			choice.Message.ToolCalls = append(choice.Message.ToolCalls, *toolCall)
		case output.Text != nil:
			// We expect only one text content block in the response.
			if choice.Message.Content == nil {
				choice.Message.Content = output.Text
			}
		case output.ReasoningContent != nil:
			choice.Message.ReasoningContent = &openai.ReasoningContentUnion{
				Value: &openai.AWSBedrockReasoningContent{
					ReasoningContent: output.ReasoningContent,
				},
			}
		}
	}
	openAIResp.Choices = append(openAIResp.Choices, choice)

	mut.Body, err = json.Marshal(openAIResp)
	if err != nil {
		return nil, nil, LLMTokenUsage{}, "", fmt.Errorf("failed to marshal body: %w", err)
	}
	headerMutation = &extprocv3.HeaderMutation{}
	setContentLength(headerMutation, mut.Body)
	if span != nil {
		span.RecordResponse(openAIResp)
	}
	return headerMutation, &extprocv3.BodyMutation{Mutation: mut}, tokenUsage, o.requestModel, nil
}

// extractAmazonEventStreamEvents extracts [awsbedrock.ConverseStreamEvent] from the buffered body.
// The extracted events are stored in the processor's events field.
func (o *openAIToAWSBedrockTranslatorV1ChatCompletion) extractAmazonEventStreamEvents() {
	// TODO: Maybe reuse the reader and decoder.
	r := bytes.NewReader(o.bufferedBody)
	dec := eventstream.NewDecoder()
	clear(o.events)
	o.events = o.events[:0]
	var lastRead int64
	for {
		msg, err := dec.Decode(r, nil)
		if err != nil {
			o.bufferedBody = o.bufferedBody[lastRead:]
			return
		}
		var event awsbedrock.ConverseStreamEvent
		if err := json.Unmarshal(msg.Payload, &event); err == nil {
			o.events = append(o.events, event)
		}
		lastRead = r.Size() - int64(r.Len())
	}
}

var emptyString = ""

// convertEvent converts an [awsbedrock.ConverseStreamEvent] to an [openai.ChatCompletionResponseChunk].
// This is a static method and does not require a receiver, but defined as a method for namespacing.
func (o *openAIToAWSBedrockTranslatorV1ChatCompletion) convertEvent(event *awsbedrock.ConverseStreamEvent) (*openai.ChatCompletionResponseChunk, bool) {
	const object = "chat.completion.chunk"
	chunk := &openai.ChatCompletionResponseChunk{Object: object}

	switch {
	case event.Usage != nil:
		chunk.Usage = &openai.ChatCompletionResponseUsage{
			TotalTokens:      event.Usage.TotalTokens,
			PromptTokens:     event.Usage.InputTokens,
			CompletionTokens: event.Usage.OutputTokens,
		}
	case event.Role != nil:
		chunk.Choices = append(chunk.Choices, openai.ChatCompletionResponseChunkChoice{
			Index: 0,
			Delta: &openai.ChatCompletionResponseChunkChoiceDelta{
				Role:    *event.Role,
				Content: &emptyString,
			},
		})
		o.role = *event.Role
	case event.Delta != nil:
		switch {
		case event.Delta.Text != nil:
			chunk.Choices = append(chunk.Choices, openai.ChatCompletionResponseChunkChoice{
				Index: 0,
				Delta: &openai.ChatCompletionResponseChunkChoiceDelta{
					Role:    o.role,
					Content: event.Delta.Text,
				},
			})
		case event.Delta.ToolUse != nil:
			chunk.Choices = append(chunk.Choices, openai.ChatCompletionResponseChunkChoice{
				Index: 0,
				Delta: &openai.ChatCompletionResponseChunkChoiceDelta{
					Role: o.role,
					ToolCalls: []openai.ChatCompletionMessageToolCallParam{
						{
							Function: openai.ChatCompletionMessageToolCallFunctionParam{
								Arguments: event.Delta.ToolUse.Input,
							},
							Type: openai.ChatCompletionMessageToolCallTypeFunction,
						},
					},
				},
			})
		case event.Delta.ReasoningContent != nil:
			reasoningDelta := &openai.AWSBedrockStreamReasoningContent{}

			// Map all relevant fields from the Bedrock delta to our flattened OpenAI delta struct.
			if event.Delta.ReasoningContent != nil {
				reasoningDelta.Text = event.Delta.ReasoningContent.Text
				reasoningDelta.Signature = event.Delta.ReasoningContent.Signature
			}
			if event.Delta.ReasoningContent.RedactedContent != nil {
				reasoningDelta.RedactedContent = event.Delta.ReasoningContent.RedactedContent
			}

			chunk.Choices = append(chunk.Choices, openai.ChatCompletionResponseChunkChoice{
				Index: 0,
				Delta: &openai.ChatCompletionResponseChunkChoiceDelta{
					Role:             o.role,
					ReasoningContent: reasoningDelta,
				},
			})
		}
	case event.Start != nil:
		if event.Start.ToolUse != nil {
			chunk.Choices = append(chunk.Choices, openai.ChatCompletionResponseChunkChoice{
				Index: 0,
				Delta: &openai.ChatCompletionResponseChunkChoiceDelta{
					Role: o.role,
					ToolCalls: []openai.ChatCompletionMessageToolCallParam{
						{
							ID: &event.Start.ToolUse.ToolUseID,
							Function: openai.ChatCompletionMessageToolCallFunctionParam{
								Name: event.Start.ToolUse.Name,
							},
							Type: openai.ChatCompletionMessageToolCallTypeFunction,
						},
					},
				},
			})
		}
	case event.StopReason != nil:
		chunk.Choices = append(chunk.Choices, openai.ChatCompletionResponseChunkChoice{
			Index: 0,
			Delta: &openai.ChatCompletionResponseChunkChoiceDelta{
				Role:    o.role,
				Content: ptr.To(emptyString),
			},
			FinishReason: o.bedrockStopReasonToOpenAIStopReason(event.StopReason),
		})
	default:
		return chunk, false
	}
	return chunk, true
}
