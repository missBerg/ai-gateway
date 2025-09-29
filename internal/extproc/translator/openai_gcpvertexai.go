// Copyright Envoy AI Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"google.golang.org/genai"
	"k8s.io/utils/ptr"

	"github.com/envoyproxy/ai-gateway/internal/apischema/gcp"
	"github.com/envoyproxy/ai-gateway/internal/apischema/openai"
	"github.com/envoyproxy/ai-gateway/internal/internalapi"
	tracing "github.com/envoyproxy/ai-gateway/internal/tracing/api"
)

const (
	gcpVertexAIBackendError = "GCPVertexAIBackendError"
)

// gcpVertexAIError represents the structure of GCP Vertex AI error responses.
type gcpVertexAIError struct {
	Error gcpVertexAIErrorDetails `json:"error"`
}

type gcpVertexAIErrorDetails struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Status  string          `json:"status"`
	Details json.RawMessage `json:"details"`
}

// NewChatCompletionOpenAIToGCPVertexAITranslator implements [Factory] for OpenAI to GCP Gemini translation.
// This translator converts OpenAI ChatCompletion API requests to GCP Gemini API format.
func NewChatCompletionOpenAIToGCPVertexAITranslator(modelNameOverride internalapi.ModelNameOverride) OpenAIChatCompletionTranslator {
	return &openAIToGCPVertexAITranslatorV1ChatCompletion{modelNameOverride: modelNameOverride}
}

// openAIToGCPVertexAITranslatorV1ChatCompletion translates OpenAI Chat Completions API to GCP Vertex AI Gemini API.
// Note: This uses the Gemini native API directly, not Vertex AI's OpenAI-compatible API:
// https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/inference
type openAIToGCPVertexAITranslatorV1ChatCompletion struct {
	modelNameOverride internalapi.ModelNameOverride
	stream            bool   // Track if this is a streaming request.
	bufferedBody      []byte // Buffer for incomplete JSON chunks.
	requestModel      internalapi.RequestModel
}

// RequestBody implements [OpenAIChatCompletionTranslator.RequestBody] for GCP Gemini.
// This method translates an OpenAI ChatCompletion request to a GCP Gemini API request.
func (o *openAIToGCPVertexAITranslatorV1ChatCompletion) RequestBody(_ []byte, openAIReq *openai.ChatCompletionRequest, _ bool) (
	headerMutation *extprocv3.HeaderMutation, bodyMutation *extprocv3.BodyMutation, err error,
) {
	o.requestModel = openAIReq.Model
	if o.modelNameOverride != "" {
		// Use modelName override if set.
		o.requestModel = o.modelNameOverride
	}

	// Set streaming flag.
	o.stream = openAIReq.Stream

	// Choose the correct endpoint based on streaming.
	var pathSuffix string
	if o.stream {
		// For streaming requests, use the streamGenerateContent endpoint with SSE format.
		pathSuffix = buildGCPModelPathSuffix(gcpModelPublisherGoogle, o.requestModel, gcpMethodStreamGenerateContent, "alt=sse")
	} else {
		pathSuffix = buildGCPModelPathSuffix(gcpModelPublisherGoogle, o.requestModel, gcpMethodGenerateContent)
	}
	gcpReq, err := o.openAIMessageToGeminiMessage(openAIReq)
	if err != nil {
		return nil, nil, fmt.Errorf("error converting OpenAI request to Gemini request: %w", err)
	}
	gcpReqBody, err := json.Marshal(gcpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshaling Gemini request: %w", err)
	}

	headerMutation, bodyMutation = buildRequestMutations(pathSuffix, gcpReqBody)
	return headerMutation, bodyMutation, nil
}

// ResponseHeaders implements [OpenAIChatCompletionTranslator.ResponseHeaders].
func (o *openAIToGCPVertexAITranslatorV1ChatCompletion) ResponseHeaders(_ map[string]string) (
	headerMutation *extprocv3.HeaderMutation, err error,
) {
	if o.stream {
		// For streaming responses, set content-type to text/event-stream to match OpenAI API.
		headerMutation = &extprocv3.HeaderMutation{
			SetHeaders: []*corev3.HeaderValueOption{
				{Header: &corev3.HeaderValue{Key: contentTypeHeaderName, Value: eventStreamContentType}},
			},
		}
	}
	return
}

// ResponseBody implements [OpenAIChatCompletionTranslator.ResponseBody] for GCP Gemini.
// This method translates a GCP Gemini API response to the OpenAI ChatCompletion format.
// GCP Vertex AI uses deterministic model mapping without virtualization, where the requested model
// is exactly what gets executed. The response does not contain a model field, so we return
// the request model that was originally sent.
func (o *openAIToGCPVertexAITranslatorV1ChatCompletion) ResponseBody(_ map[string]string, body io.Reader, endOfStream bool, span tracing.ChatCompletionSpan) (
	headerMutation *extprocv3.HeaderMutation, bodyMutation *extprocv3.BodyMutation, tokenUsage LLMTokenUsage, responseModel string, err error,
) {
	if o.stream {
		return o.handleStreamingResponse(body, endOfStream, span)
	}

	// Non-streaming logic.
	var gcpResp genai.GenerateContentResponse
	if err = json.NewDecoder(body).Decode(&gcpResp); err != nil {
		return nil, nil, LLMTokenUsage{}, "", fmt.Errorf("error decoding GCP response: %w", err)
	}

	var openAIRespBytes []byte
	// Convert to OpenAI format.
	openAIResp, err := o.geminiResponseToOpenAIMessage(gcpResp)
	if err != nil {
		return nil, nil, LLMTokenUsage{}, "", fmt.Errorf("error converting GCP response to OpenAI format: %w", err)
	}

	// Marshal the OpenAI response.
	openAIRespBytes, err = json.Marshal(openAIResp)
	if err != nil {
		return nil, nil, LLMTokenUsage{}, "", fmt.Errorf("error marshaling OpenAI response: %w", err)
	}

	// Update token usage if available.
	var usage LLMTokenUsage
	if gcpResp.UsageMetadata != nil {
		usage = LLMTokenUsage{
			InputTokens:  uint32(gcpResp.UsageMetadata.PromptTokenCount),     // nolint:gosec
			OutputTokens: uint32(gcpResp.UsageMetadata.CandidatesTokenCount), // nolint:gosec
			TotalTokens:  uint32(gcpResp.UsageMetadata.TotalTokenCount),      // nolint:gosec
		}
	}

	headerMutation, bodyMutation = buildRequestMutations("", openAIRespBytes)
	if span != nil {
		span.RecordResponse(openAIResp)
	}
	return headerMutation, bodyMutation, usage, o.requestModel, nil
}

// handleStreamingResponse handles streaming responses from GCP Gemini API.
func (o *openAIToGCPVertexAITranslatorV1ChatCompletion) handleStreamingResponse(body io.Reader, endOfStream bool, span tracing.ChatCompletionSpan) (
	headerMutation *extprocv3.HeaderMutation, bodyMutation *extprocv3.BodyMutation, tokenUsage LLMTokenUsage, responseModel string, err error,
) {
	// Parse GCP streaming chunks from buffered body and current input.
	chunks, err := o.parseGCPStreamingChunks(body)
	if err != nil {
		return nil, nil, LLMTokenUsage{}, "", fmt.Errorf("error parsing GCP streaming chunks: %w", err)
	}

	sseChunkBuf := bytes.Buffer{}

	for _, chunk := range chunks {
		// Convert GCP chunk to OpenAI chunk.
		openAIChunk := o.convertGCPChunkToOpenAI(chunk)

		// Extract token usage if present in this chunk (typically in the last chunk).
		if chunk.UsageMetadata != nil {
			tokenUsage = LLMTokenUsage{
				InputTokens:  uint32(chunk.UsageMetadata.PromptTokenCount),     //nolint:gosec
				OutputTokens: uint32(chunk.UsageMetadata.CandidatesTokenCount), //nolint:gosec
				TotalTokens:  uint32(chunk.UsageMetadata.TotalTokenCount),      //nolint:gosec
			}
		}

		// Serialize to SSE format as expected by OpenAI API.
		var chunkBytes []byte
		chunkBytes, err = json.Marshal(openAIChunk)
		if err != nil {
			return nil, nil, LLMTokenUsage{}, "", fmt.Errorf("error marshaling OpenAI chunk: %w", err)
		}
		sseChunkBuf.WriteString("data: ")
		sseChunkBuf.Write(chunkBytes)
		sseChunkBuf.WriteString("\n\n")

		if span != nil {
			span.RecordResponseChunk(openAIChunk)
		}
	}
	mut := &extprocv3.BodyMutation_Body{
		Body: sseChunkBuf.Bytes(),
	}

	if endOfStream {
		// Add the [DONE] marker to indicate end of stream as per OpenAI API specification.
		mut.Body = append(mut.Body, []byte("data: [DONE]\n")...)
	}

	return headerMutation, &extprocv3.BodyMutation{Mutation: mut}, tokenUsage, o.requestModel, nil
}

// parseGCPStreamingChunks parses the buffered body to extract complete JSON chunks.
func (o *openAIToGCPVertexAITranslatorV1ChatCompletion) parseGCPStreamingChunks(body io.Reader) ([]genai.GenerateContentResponse, error) {
	var chunks []genai.GenerateContentResponse

	// Read buffered body and new input, then split into individual chunks.
	bodyBytes, err := io.ReadAll(io.MultiReader(bytes.NewReader(o.bufferedBody), body))
	if err != nil {
		return nil, err
	}
	lines := bytes.Split(bodyBytes, []byte("\n\n"))

	for idx, line := range lines {
		// Remove "data: " prefix from SSE format if present.
		line = bytes.TrimPrefix(line, []byte("data: "))

		// Try to parse as JSON.
		var chunk genai.GenerateContentResponse
		if err = json.Unmarshal(line, &chunk); err == nil {
			chunks = append(chunks, chunk)
		} else if idx == len(lines)-1 {
			// If we reach the last line, and it can't be parsed, keep it in the buffer
			// for the next call to handle incomplete JSON chunks.
			o.bufferedBody = line
		}
		// If this is not the last line and json unmarshal fails, we assume it's an invalid chunk and ignore it.
		//	TODO: Log this as a warning or error once logger is available in this context.
	}

	return chunks, nil
}

// convertGCPChunkToOpenAI converts a GCP streaming chunk to OpenAI streaming format.
func (o *openAIToGCPVertexAITranslatorV1ChatCompletion) convertGCPChunkToOpenAI(chunk genai.GenerateContentResponse) *openai.ChatCompletionResponseChunk {
	// Convert candidates to OpenAI choices for streaming.
	choices, err := geminiCandidatesToOpenAIStreamingChoices(chunk.Candidates)
	if err != nil {
		// For now, create empty choices on error to prevent breaking the stream.
		choices = []openai.ChatCompletionResponseChunkChoice{}
	}

	// Convert usage to pointer if available.
	var usage *openai.ChatCompletionResponseUsage
	if chunk.UsageMetadata != nil {
		usage = ptr.To(geminiUsageToOpenAIUsage(chunk.UsageMetadata))
	}

	return &openai.ChatCompletionResponseChunk{
		Object:  "chat.completion.chunk",
		Choices: choices,
		Usage:   usage,
	}
}

// openAIMessageToGeminiMessage converts an OpenAI ChatCompletionRequest to a GCP Gemini GenerateContentRequest.
func (o *openAIToGCPVertexAITranslatorV1ChatCompletion) openAIMessageToGeminiMessage(openAIReq *openai.ChatCompletionRequest) (*gcp.GenerateContentRequest, error) {
	// Convert OpenAI messages to Gemini Contents and SystemInstruction.
	contents, systemInstruction, err := openAIMessagesToGeminiContents(openAIReq.Messages)
	if err != nil {
		return nil, err
	}

	// Convert OpenAI tools to Gemini tools.
	tools, err := openAIToolsToGeminiTools(openAIReq.Tools)
	if err != nil {
		return nil, fmt.Errorf("error converting tools: %w", err)
	}

	// Convert tool config.
	toolConfig, err := openAIToolChoiceToGeminiToolConfig(openAIReq.ToolChoice)
	if err != nil {
		return nil, fmt.Errorf("error converting tool choice: %w", err)
	}

	// Convert generation config.
	generationConfig, err := openAIReqToGeminiGenerationConfig(openAIReq)
	if err != nil {
		return nil, fmt.Errorf("error converting generation config: %w", err)
	}

	gcr := gcp.GenerateContentRequest{
		Contents:          contents,
		Tools:             tools,
		ToolConfig:        toolConfig,
		GenerationConfig:  generationConfig,
		SystemInstruction: systemInstruction,
	}

	// Apply vendor-specific fields after standard OpenAI-to-Gemini translation.
	// Vendor fields take precedence over translated fields when conflicts occur.
	o.applyVendorSpecificFields(openAIReq, &gcr)

	return &gcr, nil
}

// applyVendorSpecificFields applies GCP Vertex AI vendor-specific fields to the Gemini request.
// These fields allow users to access advanced GCP-specific features not available in the OpenAI API.
// Vendor fields override any conflicting fields that were set during the standard translation process.
func (o *openAIToGCPVertexAITranslatorV1ChatCompletion) applyVendorSpecificFields(openAIReq *openai.ChatCompletionRequest, gcr *gcp.GenerateContentRequest) {
	// Early return if no vendor fields are specified.
	if openAIReq.GCPVertexAIVendorFields == nil {
		return
	}

	gcpVendorFields := openAIReq.GCPVertexAIVendorFields
	// Apply vendor-specific generation config if present.
	if vendorGenConfig := gcpVendorFields.GenerationConfig; vendorGenConfig != nil {
		if gcr.GenerationConfig == nil {
			gcr.GenerationConfig = &genai.GenerationConfig{}
		}
		if vendorGenConfig.ThinkingConfig != nil {
			gcr.GenerationConfig.ThinkingConfig = vendorGenConfig.ThinkingConfig
		}
	}
	if gcpVendorFields.SafetySettings != nil {
		gcr.SafetySettings = gcpVendorFields.SafetySettings
	}
}

func (o *openAIToGCPVertexAITranslatorV1ChatCompletion) geminiResponseToOpenAIMessage(gcr genai.GenerateContentResponse) (*openai.ChatCompletionResponse, error) {
	// Convert candidates to OpenAI choices.
	choices, err := geminiCandidatesToOpenAIChoices(gcr.Candidates)
	if err != nil {
		return nil, fmt.Errorf("error converting choices: %w", err)
	}

	// Set up the OpenAI response.
	openaiResp := &openai.ChatCompletionResponse{
		Choices: choices,
		Object:  "chat.completion",
		Usage:   geminiUsageToOpenAIUsage(gcr.UsageMetadata),
	}

	return openaiResp, nil
}

// ResponseError implements [OpenAIChatCompletionTranslator.ResponseError].
// Translate GCP Vertex AI exceptions to OpenAI error type.
// GCP error responses typically contain JSON with error details or plain text error messages.
func (o *openAIToGCPVertexAITranslatorV1ChatCompletion) ResponseError(respHeaders map[string]string, body io.Reader) (
	headerMutation *extprocv3.HeaderMutation, bodyMutation *extprocv3.BodyMutation, err error,
) {
	var buf []byte
	buf, err = io.ReadAll(body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read error body: %w", err)
	}

	// Assume all responses have a valid status code header.
	statusCode := respHeaders[statusHeaderName]

	openaiError := openai.Error{
		Type: "error",
		Error: openai.ErrorType{
			Type: gcpVertexAIBackendError,
			Code: &statusCode,
		},
	}

	var gcpError gcpVertexAIError
	// Try to parse as GCP error response structure.
	if err = json.Unmarshal(buf, &gcpError); err == nil {
		errMsg := gcpError.Error.Message
		if len(gcpError.Error.Details) > 0 {
			// If details are present and not null, append them to the error message.
			errMsg = fmt.Sprintf("Error: %s\nDetails: %s", errMsg, string(gcpError.Error.Details))
		}
		openaiError.Error.Type = gcpError.Error.Status
		openaiError.Error.Message = errMsg
	} else {
		// If not JSON, read the raw body as the error message.
		openaiError.Error.Message = string(buf)
	}

	errBdy, err := json.Marshal(openaiError)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal error body: %w", err)
	}
	headerMutation, bodyMutation = buildRequestMutations("", errBdy)
	return headerMutation, bodyMutation, nil
}
