// Copyright Envoy AI Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream"
	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	"github.com/envoyproxy/ai-gateway/internal/apischema/awsbedrock"
	"github.com/envoyproxy/ai-gateway/internal/apischema/openai"
)

func TestOpenAIToAWSBedrockTranslatorV1ChatCompletion_RequestBody(t *testing.T) {
	tests := []struct {
		name   string
		output awsbedrock.ConverseInput
		input  openai.ChatCompletionRequest
	}{
		{
			name: "basic test",
			input: openai.ChatCompletionRequest{
				Stream: false,
				Model:  "gpt-4o",
				Messages: []openai.ChatCompletionMessageParamUnion{
					{
						OfSystem: &openai.ChatCompletionSystemMessageParam{
							Content: openai.StringOrArray{
								Value: "from-system",
							},
							Role: openai.ChatMessageRoleSystem,
						},
					},
					{
						OfDeveloper: &openai.ChatCompletionDeveloperMessageParam{
							Content: openai.StringOrArray{
								Value: "from-developer",
							},
							Role: openai.ChatMessageRoleDeveloper,
						},
					},
					{
						OfUser: &openai.ChatCompletionUserMessageParam{
							Content: openai.StringOrUserRoleContentUnion{
								Value: "from-user",
							},
							Role: openai.ChatMessageRoleUser,
						},
					},
					{
						OfUser: &openai.ChatCompletionUserMessageParam{
							Content: openai.StringOrUserRoleContentUnion{
								Value: "part1",
							},
							Role: openai.ChatMessageRoleUser,
						},
					},
					{
						OfUser: &openai.ChatCompletionUserMessageParam{
							Content: openai.StringOrUserRoleContentUnion{
								Value: "part2",
							},
							Role: openai.ChatMessageRoleUser,
						},
					},
					{
						OfTool: &openai.ChatCompletionToolMessageParam{
							Content: openai.StringOrArray{
								Value: "Weather in Queens, NY is 70F and clear skies.",
							},
							ToolCallID: "call_6g7a",
							Role:       openai.ChatMessageRoleTool,
						},
					},
					{
						OfAssistant: &openai.ChatCompletionAssistantMessageParam{
							Content: openai.StringOrAssistantRoleContentUnion{
								Value: openai.ChatCompletionAssistantMessageParamContent{
									Type: openai.ChatCompletionAssistantMessageParamContentTypeText,
									Text: ptr.To("I dunno"),
								},
							},
							ToolCalls: []openai.ChatCompletionMessageToolCallParam{
								{
									ID: ptr.To("call_6g7a"),
									Function: openai.ChatCompletionMessageToolCallFunctionParam{
										Arguments: "{\"code_block\":\"from playwright.sync_api import sync_playwright\\n\"}",
										Name:      "exec_python_code",
									},
									Type: openai.ChatCompletionMessageToolCallTypeFunction,
								},
							},
							Role: openai.ChatMessageRoleAssistant,
						},
					},
					{
						OfAssistant: &openai.ChatCompletionAssistantMessageParam{
							Content: openai.StringOrAssistantRoleContentUnion{
								Value: "I also dunno",
							},
							Role: openai.ChatMessageRoleAssistant,
						},
					},
					{
						OfAssistant: &openai.ChatCompletionAssistantMessageParam{
							Content: openai.StringOrAssistantRoleContentUnion{
								Value: "",
							},
							Role: openai.ChatMessageRoleAssistant,
						},
					},
				},
			},
			output: awsbedrock.ConverseInput{
				InferenceConfig: &awsbedrock.InferenceConfiguration{},
				System: []*awsbedrock.SystemContentBlock{
					{
						Text: "from-system",
					},
					{
						Text: "from-developer",
					},
				},
				Messages: []*awsbedrock.Message{
					{
						Role: openai.ChatMessageRoleUser,
						Content: []*awsbedrock.ContentBlock{
							{
								Text: ptr.To("from-user"),
							},
						},
					},
					{
						Role: openai.ChatMessageRoleUser,
						Content: []*awsbedrock.ContentBlock{
							{
								Text: ptr.To("part1"),
							},
						},
					},
					{
						Role: openai.ChatMessageRoleUser,
						Content: []*awsbedrock.ContentBlock{
							{
								Text: ptr.To("part2"),
							},
						},
					},
					{
						Role: openai.ChatMessageRoleUser,
						Content: []*awsbedrock.ContentBlock{
							{
								ToolResult: &awsbedrock.ToolResultBlock{
									ToolUseID: ptr.To("call_6g7a"),
									Content: []*awsbedrock.ToolResultContentBlock{
										{
											Text: ptr.To("Weather in Queens, NY is 70F and clear skies."),
										},
									},
								},
							},
						},
					},
					{
						Role: openai.ChatMessageRoleAssistant,
						Content: []*awsbedrock.ContentBlock{
							{
								Text: ptr.To("I dunno"),
							},
							{
								ToolUse: &awsbedrock.ToolUseBlock{
									Name:      "exec_python_code",
									ToolUseID: "call_6g7a",
									Input:     map[string]any{"code_block": "from playwright.sync_api import sync_playwright\n"},
								},
							},
						},
					},
					{
						Role: openai.ChatMessageRoleAssistant,
						Content: []*awsbedrock.ContentBlock{
							{
								Text: ptr.To("I also dunno"),
							},
						},
					},
					{
						Role:    openai.ChatMessageRoleAssistant,
						Content: []*awsbedrock.ContentBlock{},
					},
				},
			},
		},
		{
			name: "test content array",
			input: openai.ChatCompletionRequest{
				Stream: false,
				Model:  "gpt-4o",
				Messages: []openai.ChatCompletionMessageParamUnion{
					{
						OfSystem: &openai.ChatCompletionSystemMessageParam{
							Content: openai.StringOrArray{
								Value: []openai.ChatCompletionContentPartTextParam{
									{Text: "from-system"},
								},
							},
							Role: openai.ChatMessageRoleSystem,
						},
					},
					{
						OfDeveloper: &openai.ChatCompletionDeveloperMessageParam{
							Content: openai.StringOrArray{
								Value: []openai.ChatCompletionContentPartTextParam{
									{Text: "from-developer"},
								},
							},
							Role: openai.ChatMessageRoleDeveloper,
						},
					},
					{
						OfUser: &openai.ChatCompletionUserMessageParam{
							Content: openai.StringOrUserRoleContentUnion{
								Value: []openai.ChatCompletionContentPartUserUnionParam{
									{OfText: &openai.ChatCompletionContentPartTextParam{Text: "from-user"}},
								},
							},
							Role: openai.ChatMessageRoleUser,
						},
					},
					{
						OfUser: &openai.ChatCompletionUserMessageParam{
							Content: openai.StringOrUserRoleContentUnion{
								Value: []openai.ChatCompletionContentPartUserUnionParam{
									{OfText: &openai.ChatCompletionContentPartTextParam{Text: "user1"}},
								},
							},
							Role: openai.ChatMessageRoleUser,
						},
					},
					{
						OfUser: &openai.ChatCompletionUserMessageParam{
							Content: openai.StringOrUserRoleContentUnion{
								Value: []openai.ChatCompletionContentPartUserUnionParam{
									{OfText: &openai.ChatCompletionContentPartTextParam{Text: "user2"}},
								},
							},
							Role: openai.ChatMessageRoleUser,
						},
					},
				},
			},
			output: awsbedrock.ConverseInput{
				InferenceConfig: &awsbedrock.InferenceConfiguration{},
				System: []*awsbedrock.SystemContentBlock{
					{
						Text: "from-system",
					},
					{
						Text: "from-developer",
					},
				},
				Messages: []*awsbedrock.Message{
					{
						Role: openai.ChatMessageRoleUser,
						Content: []*awsbedrock.ContentBlock{
							{
								Text: ptr.To("from-user"),
							},
						},
					},
					{
						Role: openai.ChatMessageRoleUser,
						Content: []*awsbedrock.ContentBlock{
							{
								Text: ptr.To("user1"),
							},
						},
					},
					{
						Role: openai.ChatMessageRoleUser,
						Content: []*awsbedrock.ContentBlock{
							{
								Text: ptr.To("user2"),
							},
						},
					},
				},
			},
		},
		{
			name: "test image",
			input: openai.ChatCompletionRequest{
				Stream: false,
				Model:  "gpt-4o",
				Messages: []openai.ChatCompletionMessageParamUnion{
					{
						OfSystem: &openai.ChatCompletionSystemMessageParam{
							Content: openai.StringOrArray{
								Value: []openai.ChatCompletionContentPartTextParam{
									{Text: "from-system"},
								},
							},
							Role: openai.ChatMessageRoleSystem,
						},
					},
					{
						OfUser: &openai.ChatCompletionUserMessageParam{
							Content: openai.StringOrUserRoleContentUnion{
								Value: []openai.ChatCompletionContentPartUserUnionParam{
									{OfImageURL: &openai.ChatCompletionContentPartImageParam{
										ImageURL: openai.ChatCompletionContentPartImageImageURLParam{
											URL: "data:image/jpeg;base64,dGVzdA==",
										},
									}},
								},
							},
							Role: openai.ChatMessageRoleUser,
						},
					},
				},
			},
			output: awsbedrock.ConverseInput{
				InferenceConfig: &awsbedrock.InferenceConfiguration{},
				System: []*awsbedrock.SystemContentBlock{
					{
						Text: "from-system",
					},
				},
				Messages: []*awsbedrock.Message{
					{
						Role: openai.ChatMessageRoleUser,
						Content: []*awsbedrock.ContentBlock{
							{
								Image: &awsbedrock.ImageBlock{
									Source: awsbedrock.ImageSource{
										Bytes: []byte("test"),
									},
									Format: "jpeg",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "test parameters",
			input: openai.ChatCompletionRequest{
				Stream:      false,
				Model:       "gpt-4o",
				MaxTokens:   ptr.To(int64(10)),
				TopP:        ptr.To(float64(1)),
				Temperature: ptr.To(0.7),
				Messages: []openai.ChatCompletionMessageParamUnion{
					{
						OfUser: &openai.ChatCompletionUserMessageParam{
							Content: openai.StringOrUserRoleContentUnion{
								Value: "from-user",
							},
							Role: openai.ChatMessageRoleUser,
						},
					},
				},
			},
			output: awsbedrock.ConverseInput{
				InferenceConfig: &awsbedrock.InferenceConfiguration{
					MaxTokens:   ptr.To(int64(10)),
					TopP:        ptr.To(float64(1)),
					Temperature: ptr.To(0.7),
				},
				Messages: []*awsbedrock.Message{
					{
						Role: openai.ChatMessageRoleUser,
						Content: []*awsbedrock.ContentBlock{
							{
								Text: ptr.To("from-user"),
							},
						},
					},
				},
			},
		},
		{
			name: "test tools function calling with empty tool choice",
			input: openai.ChatCompletionRequest{
				Stream:      false,
				Model:       "gpt-4o",
				MaxTokens:   ptr.To(int64(10)),
				TopP:        ptr.To(float64(1)),
				Temperature: ptr.To(0.7),
				Messages: []openai.ChatCompletionMessageParamUnion{
					{
						OfUser: &openai.ChatCompletionUserMessageParam{
							Content: openai.StringOrUserRoleContentUnion{
								Value: "from-user",
							},
							Role: openai.ChatMessageRoleUser,
						},
					},
				},
				Tools: []openai.Tool{
					{
						Type: "function",
						Function: &openai.FunctionDefinition{
							Name:        "get_current_weather",
							Description: "Get the current weather in a given location",
							Parameters: map[string]any{
								"type": "object",
								"properties": map[string]any{
									"location": map[string]any{
										"type":        "string",
										"description": "The city and state, e.g. San Francisco, CA",
									},
									"unit": map[string]any{
										"type": "string",
										"enum": []string{"celsius", "fahrenheit"},
									},
								},
								"required": []string{"location"},
							},
						},
					},
				},
			},
			output: awsbedrock.ConverseInput{
				InferenceConfig: &awsbedrock.InferenceConfiguration{
					MaxTokens:   ptr.To(int64(10)),
					TopP:        ptr.To(float64(1)),
					Temperature: ptr.To(0.7),
				},
				Messages: []*awsbedrock.Message{
					{
						Role: openai.ChatMessageRoleUser,
						Content: []*awsbedrock.ContentBlock{
							{
								Text: ptr.To("from-user"),
							},
						},
					},
				},
				ToolConfig: &awsbedrock.ToolConfiguration{
					Tools: []*awsbedrock.Tool{
						{
							ToolSpec: &awsbedrock.ToolSpecification{
								Name:        ptr.To("get_current_weather"),
								Description: ptr.To("Get the current weather in a given location"),
								InputSchema: &awsbedrock.ToolInputSchema{
									JSON: map[string]any{
										"type": "object",
										"properties": map[string]any{
											"location": map[string]any{
												"type":        "string",
												"description": "The city and state, e.g. San Francisco, CA",
											},
											"unit": map[string]any{
												"type": "string",
												"enum": []any{"celsius", "fahrenheit"},
											},
										},
										"required": []any{"location"},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "test auto tool choice",
			input: openai.ChatCompletionRequest{
				Model: "gpt-4o",
				Messages: []openai.ChatCompletionMessageParamUnion{
					{
						OfUser: &openai.ChatCompletionUserMessageParam{
							Content: openai.StringOrUserRoleContentUnion{
								Value: "from-user",
							},
							Role: openai.ChatMessageRoleUser,
						},
					},
				},
				Tools: []openai.Tool{
					{
						Type: "function",
						Function: &openai.FunctionDefinition{
							Name:        "get_current_weather",
							Description: "Get the current weather in a given location",
						},
					},
				},
				ToolChoice: "auto",
			},
			output: awsbedrock.ConverseInput{
				InferenceConfig: &awsbedrock.InferenceConfiguration{},
				Messages: []*awsbedrock.Message{
					{
						Role: openai.ChatMessageRoleUser,
						Content: []*awsbedrock.ContentBlock{
							{
								Text: ptr.To("from-user"),
							},
						},
					},
				},
				ToolConfig: &awsbedrock.ToolConfiguration{
					Tools: []*awsbedrock.Tool{
						{
							ToolSpec: &awsbedrock.ToolSpecification{
								Name:        ptr.To("get_current_weather"),
								Description: ptr.To("Get the current weather in a given location"),
								InputSchema: &awsbedrock.ToolInputSchema{},
							},
						},
					},
					ToolChoice: &awsbedrock.ToolChoice{Auto: &awsbedrock.AutoToolChoice{}},
				},
			},
		},
		{
			name: "test required tool choice",
			input: openai.ChatCompletionRequest{
				Model: "gpt-4o",
				Messages: []openai.ChatCompletionMessageParamUnion{
					{
						OfUser: &openai.ChatCompletionUserMessageParam{
							Content: openai.StringOrUserRoleContentUnion{
								Value: "from-user",
							},
							Role: openai.ChatMessageRoleUser,
						},
					},
				},
				Tools: []openai.Tool{
					{
						Type: "function",
						Function: &openai.FunctionDefinition{
							Name:        "get_current_weather",
							Description: "Get the current weather in a given location",
						},
					},
				},
				ToolChoice: "required",
			},
			output: awsbedrock.ConverseInput{
				InferenceConfig: &awsbedrock.InferenceConfiguration{},
				Messages: []*awsbedrock.Message{
					{
						Role: openai.ChatMessageRoleUser,
						Content: []*awsbedrock.ContentBlock{
							{
								Text: ptr.To("from-user"),
							},
						},
					},
				},
				ToolConfig: &awsbedrock.ToolConfiguration{
					Tools: []*awsbedrock.Tool{
						{
							ToolSpec: &awsbedrock.ToolSpecification{
								Name:        ptr.To("get_current_weather"),
								Description: ptr.To("Get the current weather in a given location"),
								InputSchema: &awsbedrock.ToolInputSchema{},
							},
						},
					},
					ToolChoice: &awsbedrock.ToolChoice{Any: &awsbedrock.AnyToolChoice{}},
				},
			},
		},
		{
			name: "test tool choice for anthropic claude model",
			input: openai.ChatCompletionRequest{
				Model: "bedrock.anthropic.claude-3-5-sonnet-20240620-v1:0",
				Messages: []openai.ChatCompletionMessageParamUnion{
					{
						OfUser: &openai.ChatCompletionUserMessageParam{
							Content: openai.StringOrUserRoleContentUnion{
								Value: "from-user",
							},
							Role: openai.ChatMessageRoleUser,
						},
					},
				},
				Tools: []openai.Tool{
					{
						Type: "function",
						Function: &openai.FunctionDefinition{
							Name:        "get_current_weather",
							Description: "Get the current weather in a given location",
						},
					},
				},
				ToolChoice: "some-tools",
			},
			output: awsbedrock.ConverseInput{
				InferenceConfig: &awsbedrock.InferenceConfiguration{},
				Messages: []*awsbedrock.Message{
					{
						Role: openai.ChatMessageRoleUser,
						Content: []*awsbedrock.ContentBlock{
							{
								Text: ptr.To("from-user"),
							},
						},
					},
				},
				ToolConfig: &awsbedrock.ToolConfiguration{
					Tools: []*awsbedrock.Tool{
						{
							ToolSpec: &awsbedrock.ToolSpecification{
								Name:        ptr.To("get_current_weather"),
								Description: ptr.To("Get the current weather in a given location"),
								InputSchema: &awsbedrock.ToolInputSchema{},
							},
						},
					},
					ToolChoice: &awsbedrock.ToolChoice{
						Tool: &awsbedrock.SpecificToolChoice{
							Name: ptr.To("some-tools"),
						},
					},
				},
			},
		},
		{
			name: "test tool choices for anthropic claude model",
			input: openai.ChatCompletionRequest{
				Model: "bedrock.anthropic.claude-3-5-sonnet-20240620-v1:0",
				Messages: []openai.ChatCompletionMessageParamUnion{
					{
						OfUser: &openai.ChatCompletionUserMessageParam{
							Content: openai.StringOrUserRoleContentUnion{
								Value: "from-user",
							},
							Role: openai.ChatMessageRoleUser,
						},
					},
				},
				Tools: []openai.Tool{
					{
						Type: "function",
						Function: &openai.FunctionDefinition{
							Name:        "get_current_weather",
							Description: "Get the current weather in a given location",
						},
					},
				},
				ToolChoice: openai.ToolChoice{
					Type: openai.ToolType("function"),
					Function: openai.ToolFunction{
						Name: "my_function",
					},
				},
			},
			output: awsbedrock.ConverseInput{
				InferenceConfig: &awsbedrock.InferenceConfiguration{},
				Messages: []*awsbedrock.Message{
					{
						Role: openai.ChatMessageRoleUser,
						Content: []*awsbedrock.ContentBlock{
							{
								Text: ptr.To("from-user"),
							},
						},
					},
				},
				ToolConfig: &awsbedrock.ToolConfiguration{
					Tools: []*awsbedrock.Tool{
						{
							ToolSpec: &awsbedrock.ToolSpecification{
								Name:        ptr.To("get_current_weather"),
								Description: ptr.To("Get the current weather in a given location"),
								InputSchema: &awsbedrock.ToolInputSchema{},
							},
						},
					},
					ToolChoice: &awsbedrock.ToolChoice{
						Tool: &awsbedrock.SpecificToolChoice{
							Name: ptr.To("function"),
						},
					},
				},
			},
		},
		{
			name: "test single stop word",
			input: openai.ChatCompletionRequest{
				Model: "gpt-4o",
				Messages: []openai.ChatCompletionMessageParamUnion{
					{
						OfUser: &openai.ChatCompletionUserMessageParam{
							Content: openai.StringOrUserRoleContentUnion{
								Value: "from-user",
							},
							Role: openai.ChatMessageRoleUser,
						},
					},
				},
				Stop: []*string{ptr.To("stop_only")},
			},
			output: awsbedrock.ConverseInput{
				InferenceConfig: &awsbedrock.InferenceConfiguration{
					StopSequences: []*string{ptr.To("stop_only")},
				},
				Messages: []*awsbedrock.Message{
					{
						Role: openai.ChatMessageRoleUser,
						Content: []*awsbedrock.ContentBlock{
							{
								Text: ptr.To("from-user"),
							},
						},
					},
				},
			},
		},
		{
			name: "test parallel tool calls for anthropic claude model",
			input: openai.ChatCompletionRequest{
				Model: "bedrock.anthropic.claude-3-5-sonnet-20240620-v1:0",
				Messages: []openai.ChatCompletionMessageParamUnion{
					{
						OfUser: &openai.ChatCompletionUserMessageParam{
							Role: openai.ChatMessageRoleUser,
							Content: openai.StringOrUserRoleContentUnion{
								Value: "What is the weather in Dallas, Texas and Orlando, Florida in Fahrenheit?",
							},
						},
					},
					{
						OfAssistant: &openai.ChatCompletionAssistantMessageParam{
							Role: openai.ChatMessageRoleAssistant,
							ToolCalls: []openai.ChatCompletionMessageToolCallParam{
								{
									ID: ptr.To("tool-1"),
									Function: openai.ChatCompletionMessageToolCallFunctionParam{
										Name:      "get_current_weather",
										Arguments: "{\"city\": \"Dallas\", \"state\": \"TX\", \"unit\": \"fahrenheit\"}",
									},
									Type: openai.ChatCompletionMessageToolCallType(openai.ToolTypeFunction),
								},
								{
									ID: ptr.To("tool-2"),
									Function: openai.ChatCompletionMessageToolCallFunctionParam{
										Name:      "get_current_weather",
										Arguments: "{\"city\": \"Orlando\", \"state\": \"FL\", \"unit\": \"fahrenheit\"}",
									},
									Type: openai.ChatCompletionMessageToolCallType(openai.ToolTypeFunction),
								},
							},
						},
					},
					{
						OfTool: &openai.ChatCompletionToolMessageParam{
							Content: openai.StringOrArray{
								Value: "The weather in Dallas TX is 98 degrees fahrenheit with mostly cloudy skies and a change of rain in the evening.",
							},
							Role:       openai.ChatMessageRoleTool,
							ToolCallID: "tool-1",
						},
					},
					{
						OfTool: &openai.ChatCompletionToolMessageParam{
							Content: openai.StringOrArray{
								Value: "The weather in Orlando FL is 78 degrees fahrenheit with clear skies.",
							},
							Role:       openai.ChatMessageRoleTool,
							ToolCallID: "tool-2",
						},
					},
				},
			},
			output: awsbedrock.ConverseInput{
				InferenceConfig: &awsbedrock.InferenceConfiguration{},
				Messages: []*awsbedrock.Message{
					{
						Role: openai.ChatMessageRoleUser,
						Content: []*awsbedrock.ContentBlock{
							{
								Text: ptr.To("What is the weather in Dallas, Texas and Orlando, Florida in Fahrenheit?"),
							},
						},
					},
					{
						Role: openai.ChatMessageRoleAssistant,
						Content: []*awsbedrock.ContentBlock{
							{
								ToolUse: &awsbedrock.ToolUseBlock{
									Name:      "get_current_weather",
									ToolUseID: "tool-1",
									Input:     map[string]any{"city": "Dallas", "state": "TX", "unit": "fahrenheit"},
								},
							},
							{
								ToolUse: &awsbedrock.ToolUseBlock{
									Name:      "get_current_weather",
									ToolUseID: "tool-2",
									Input:     map[string]any{"city": "Orlando", "state": "FL", "unit": "fahrenheit"},
								},
							},
						},
					},
					{
						Role: awsbedrock.ConversationRoleUser,
						Content: []*awsbedrock.ContentBlock{
							{
								ToolResult: &awsbedrock.ToolResultBlock{
									Content: []*awsbedrock.ToolResultContentBlock{
										{Text: ptr.To("The weather in Dallas TX is 98 degrees fahrenheit with mostly cloudy skies and a change of rain in the evening.")},
									},
									ToolUseID: ptr.To("tool-1"),
								},
							},
							{
								ToolResult: &awsbedrock.ToolResultBlock{
									Content: []*awsbedrock.ToolResultContentBlock{
										{Text: ptr.To("The weather in Orlando FL is 78 degrees fahrenheit with clear skies.")},
									},
									ToolUseID: ptr.To("tool-2"),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "test thinking parameter for anthropic claude model",
			input: openai.ChatCompletionRequest{
				Model: "anthropic.claude-3-sonnet-20240229-v1:0",
				Messages: []openai.ChatCompletionMessageParamUnion{
					{
						OfUser: &openai.ChatCompletionUserMessageParam{
							Content: openai.StringOrUserRoleContentUnion{
								Value: "Hello",
							},
							Role: openai.ChatMessageRoleUser,
						},
					},
				},
				AnthropicVendorFields: &openai.AnthropicVendorFields{
					Thinking: &anthropic.ThinkingConfigParamUnion{
						OfEnabled: &anthropic.ThinkingConfigEnabledParam{
							BudgetTokens: int64(1024),
						},
					},
				},
			},
			output: awsbedrock.ConverseInput{
				AdditionalModelRequestFields: map[string]interface{}{
					"thinking": map[string]interface{}{"type": "enabled", "budget_tokens": float64(1024)},
				},
				InferenceConfig: &awsbedrock.InferenceConfiguration{},
				Messages: []*awsbedrock.Message{
					{
						Role: openai.ChatMessageRoleUser,
						Content: []*awsbedrock.ContentBlock{
							{
								Text: ptr.To("Hello"),
							},
						},
					},
				},
			},
		},
		{
			name: "test assistant message with thinking content",
			input: openai.ChatCompletionRequest{
				Model: "unit-test-model",
				Messages: []openai.ChatCompletionMessageParamUnion{
					{
						OfUser: &openai.ChatCompletionUserMessageParam{
							Role:    openai.ChatMessageRoleUser,
							Content: openai.StringOrUserRoleContentUnion{Value: "How do I list prime numbers?"},
						},
					},
					{
						OfAssistant: &openai.ChatCompletionAssistantMessageParam{
							Role: openai.ChatMessageRoleAssistant,
							Content: openai.StringOrAssistantRoleContentUnion{
								Value: []openai.ChatCompletionAssistantMessageParamContent{
									{
										Type: openai.ChatCompletionAssistantMessageParamContentTypeThinking,
										Text: ptr.To("Let me think"),
									},
								},
							},
						},
					},
				},
			},
			output: awsbedrock.ConverseInput{
				InferenceConfig: &awsbedrock.InferenceConfiguration{},
				Messages: []*awsbedrock.Message{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: []*awsbedrock.ContentBlock{{Text: ptr.To("How do I list prime numbers?")}},
					},
					{
						Role: openai.ChatMessageRoleAssistant,
						Content: []*awsbedrock.ContentBlock{
							{
								ReasoningContent: &awsbedrock.ReasoningContentBlock{
									ReasoningText: &awsbedrock.ReasoningTextBlock{
										Text: "Let me think",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "test assistant message with redacted thinking content",
			input: openai.ChatCompletionRequest{
				Model: "unit-test-model",
				Messages: []openai.ChatCompletionMessageParamUnion{
					{
						OfUser: &openai.ChatCompletionUserMessageParam{
							Role:    openai.ChatMessageRoleUser,
							Content: openai.StringOrUserRoleContentUnion{Value: "How do I list prime numbers?"},
						},
					},
					{
						OfAssistant: &openai.ChatCompletionAssistantMessageParam{
							Role: openai.ChatMessageRoleAssistant,
							Content: openai.StringOrAssistantRoleContentUnion{
								Value: []openai.ChatCompletionAssistantMessageParamContent{
									{
										Type:            openai.ChatCompletionAssistantMessageParamContentTypeRedactedThinking,
										RedactedContent: []byte{104, 101, 108, 108, 111},
									},
								},
							},
						},
					},
				},
			},
			output: awsbedrock.ConverseInput{
				InferenceConfig: &awsbedrock.InferenceConfiguration{},
				Messages: []*awsbedrock.Message{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: []*awsbedrock.ContentBlock{{Text: ptr.To("How do I list prime numbers?")}},
					},
					{
						Role: openai.ChatMessageRoleAssistant,
						Content: []*awsbedrock.ContentBlock{
							{
								ReasoningContent: &awsbedrock.ReasoningContentBlock{
									RedactedContent: []byte{104, 101, 108, 108, 111},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "test multi-turn with tool use and thinking content",
			input: openai.ChatCompletionRequest{
				Model: "unit-test-model",
				Messages: []openai.ChatCompletionMessageParamUnion{
					{
						OfUser: &openai.ChatCompletionUserMessageParam{
							Role:    openai.ChatMessageRoleUser,
							Content: openai.StringOrUserRoleContentUnion{Value: "What's the weather in Paris?"},
						},
					},
					{
						OfAssistant: &openai.ChatCompletionAssistantMessageParam{
							Role: openai.ChatMessageRoleAssistant,
							ToolCalls: []openai.ChatCompletionMessageToolCallParam{
								{
									ID:   ptr.To("tool_call_123"),
									Type: openai.ChatCompletionMessageToolCallTypeFunction,
									Function: openai.ChatCompletionMessageToolCallFunctionParam{
										Name:      "get_weather",
										Arguments: "{\"location\":\"Paris\"}",
									},
								},
							},
							Content: openai.StringOrAssistantRoleContentUnion{
								Value: []openai.ChatCompletionAssistantMessageParamContent{
									{
										Type:      openai.ChatCompletionAssistantMessageParamContentTypeThinking,
										Text:      ptr.To("I need to call the get_weather tool for Paris."),
										Signature: ptr.To("sig_12345"),
									},
								},
							},
						},
					},
					{
						OfTool: &openai.ChatCompletionToolMessageParam{
							Role:       openai.ChatMessageRoleTool,
							ToolCallID: "tool_call_123",
							Content:    openai.StringOrArray{Value: "{\"temperature\": 88}"},
						},
					},
				},
			},
			output: awsbedrock.ConverseInput{
				InferenceConfig: &awsbedrock.InferenceConfiguration{},
				Messages: []*awsbedrock.Message{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: []*awsbedrock.ContentBlock{{Text: ptr.To("What's the weather in Paris?")}},
					},
					{
						Role: openai.ChatMessageRoleAssistant,
						Content: []*awsbedrock.ContentBlock{
							{
								ReasoningContent: &awsbedrock.ReasoningContentBlock{
									ReasoningText: &awsbedrock.ReasoningTextBlock{
										Text:      "I need to call the get_weather tool for Paris.",
										Signature: "sig_12345",
									},
								},
							},
							{
								ToolUse: &awsbedrock.ToolUseBlock{
									ToolUseID: "tool_call_123",
									Name:      "get_weather",
									Input:     map[string]any{"location": "Paris"},
								},
							},
						},
					},
					{
						Role: awsbedrock.ConversationRoleUser,
						Content: []*awsbedrock.ContentBlock{
							{
								ToolResult: &awsbedrock.ToolResultBlock{
									ToolUseID: ptr.To("tool_call_123"),
									Content: []*awsbedrock.ToolResultContentBlock{
										{Text: ptr.To("{\"temperature\": 88}")},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "test thinking enabled config",
			input: openai.ChatCompletionRequest{
				Model: "bedrock.unit-test-model",
				Messages: []openai.ChatCompletionMessageParamUnion{
					{
						OfUser: &openai.ChatCompletionUserMessageParam{
							Role:    openai.ChatMessageRoleUser,
							Content: openai.StringOrUserRoleContentUnion{Value: "Hello"},
						},
					},
				},
				AnthropicVendorFields: &openai.AnthropicVendorFields{
					Thinking: &anthropic.ThinkingConfigParamUnion{
						OfEnabled: &anthropic.ThinkingConfigEnabledParam{
							Type:         "enabled",
							BudgetTokens: 1024,
						},
					},
				},
			},
			output: awsbedrock.ConverseInput{
				AdditionalModelRequestFields: map[string]interface{}{
					"thinking": map[string]interface{}{"type": "enabled", "budget_tokens": float64(1024)},
				},
				InferenceConfig: &awsbedrock.InferenceConfiguration{},
				Messages: []*awsbedrock.Message{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: []*awsbedrock.ContentBlock{{Text: ptr.To("Hello")}},
					},
				},
			},
		},
		{
			name: "test thinking disabled config",
			input: openai.ChatCompletionRequest{
				Model: "bedrock.unit-test-model",
				Messages: []openai.ChatCompletionMessageParamUnion{
					{
						OfUser: &openai.ChatCompletionUserMessageParam{
							Role:    openai.ChatMessageRoleUser,
							Content: openai.StringOrUserRoleContentUnion{Value: "Hello"},
						},
					},
				},
				AnthropicVendorFields: &openai.AnthropicVendorFields{
					Thinking: &anthropic.ThinkingConfigParamUnion{
						OfDisabled: &anthropic.ThinkingConfigDisabledParam{
							Type: "disabled",
						},
					},
				},
			},
			output: awsbedrock.ConverseInput{
				AdditionalModelRequestFields: map[string]interface{}{
					"thinking": map[string]interface{}{"type": "disabled"},
				},
				InferenceConfig: &awsbedrock.InferenceConfiguration{},
				Messages: []*awsbedrock.Message{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: []*awsbedrock.ContentBlock{{Text: ptr.To("Hello")}},
					},
				},
			},
		},
		{
			name: "test assistant message with mixed text and thinking content",
			input: openai.ChatCompletionRequest{
				Model: "unit-test-model",
				Messages: []openai.ChatCompletionMessageParamUnion{
					{
						OfAssistant: &openai.ChatCompletionAssistantMessageParam{
							Role: openai.ChatMessageRoleAssistant,
							Content: openai.StringOrAssistantRoleContentUnion{
								Value: []openai.ChatCompletionAssistantMessageParamContent{
									{
										Type: openai.ChatCompletionAssistantMessageParamContentTypeText,
										Text: ptr.To("This is a standard text part."),
									},
									{
										Type: openai.ChatCompletionAssistantMessageParamContentTypeThinking,
										Text: ptr.To("This is a thinking part."),
									},
								},
							},
						},
					},
				},
			},
			output: awsbedrock.ConverseInput{
				InferenceConfig: &awsbedrock.InferenceConfiguration{},
				Messages: []*awsbedrock.Message{
					{
						Role: openai.ChatMessageRoleAssistant,
						Content: []*awsbedrock.ContentBlock{
							{
								Text: ptr.To("This is a standard text part."),
							},
							{
								ReasoningContent: &awsbedrock.ReasoningContentBlock{
									ReasoningText: &awsbedrock.ReasoningTextBlock{
										Text: "This is a thinking part.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &openAIToAWSBedrockTranslatorV1ChatCompletion{}
			originalReq := tt.input
			hm, bm, err := o.RequestBody(nil, &originalReq, false)
			var expPath string
			require.Equal(t, tt.input.Stream, o.stream)
			if tt.input.Stream {
				expPath = fmt.Sprintf("/model/%s/converse-stream", tt.input.Model)
			} else {
				expPath = fmt.Sprintf("/model/%s/converse", tt.input.Model)
			}
			require.NoError(t, err)
			require.NotNil(t, hm)
			require.NotNil(t, hm.SetHeaders)
			require.Len(t, hm.SetHeaders, 2)
			require.Equal(t, ":path", hm.SetHeaders[0].Header.Key)
			require.Equal(t, expPath, string(hm.SetHeaders[0].Header.RawValue))
			require.Equal(t, "content-length", hm.SetHeaders[1].Header.Key)
			newBody := bm.Mutation.(*extprocv3.BodyMutation_Body).Body
			require.Equal(t, strconv.Itoa(len(newBody)), string(hm.SetHeaders[1].Header.RawValue))

			var awsReq awsbedrock.ConverseInput
			err = json.Unmarshal(newBody, &awsReq)
			require.NoError(t, err)
			if !cmp.Equal(awsReq, tt.output) {
				t.Errorf("ConvertOpenAIToBedrock(), diff(got, expected) = %s\n", cmp.Diff(awsReq, tt.output))
			}
		})
	}

	t.Run("model override", func(t *testing.T) {
		modelNameOverride := "bedrock.anthropic.claude-3-5-sonnet-20240620-v1:0"
		o := &openAIToAWSBedrockTranslatorV1ChatCompletion{modelNameOverride: modelNameOverride}
		originalReq := openai.ChatCompletionRequest{
			Model:    "claude-3-5-sonnet",
			Messages: []openai.ChatCompletionMessageParamUnion{},
		}
		hm, _, err := o.RequestBody(nil, &originalReq, false)
		require.NoError(t, err)
		require.NotNil(t, hm)
		require.NotNil(t, hm.SetHeaders)
		require.Len(t, hm.SetHeaders, 2)
		require.Equal(t, ":path", hm.SetHeaders[0].Header.Key)
		require.Equal(t, "/model/"+modelNameOverride+"/converse", string(hm.SetHeaders[0].Header.RawValue))
	})
}

func TestOpenAIToAWSBedrockTranslatorV1ChatCompletion_ResponseHeaders(t *testing.T) {
	t.Run("streaming", func(t *testing.T) {
		o := &openAIToAWSBedrockTranslatorV1ChatCompletion{stream: true}
		hm, err := o.ResponseHeaders(map[string]string{
			"content-type": "application/vnd.amazon.eventstream",
		})
		require.NoError(t, err)
		require.NotNil(t, hm)
		require.NotNil(t, hm.SetHeaders)
		require.Len(t, hm.SetHeaders, 1)
		require.Equal(t, "content-type", hm.SetHeaders[0].Header.Key)
		require.Equal(t, "text/event-stream", hm.SetHeaders[0].Header.Value)
	})
	t.Run("non-streaming", func(t *testing.T) {
		o := &openAIToAWSBedrockTranslatorV1ChatCompletion{}
		hm, err := o.ResponseHeaders(nil)
		require.NoError(t, err)
		require.Nil(t, hm)
	})
}

func TestOpenAIToAWSBedrockTranslatorV1ChatCompletion_Streaming_ResponseBody(t *testing.T) {
	t.Run("streaming", func(t *testing.T) {
		o := &openAIToAWSBedrockTranslatorV1ChatCompletion{stream: true}
		buf, err := base64.StdEncoding.DecodeString(base64RealStreamingEvents)
		require.NoError(t, err)

		var results []string
		for i := range buf {
			hm, bm, tokenUsage, _, err := o.ResponseBody(nil, bytes.NewBuffer([]byte{buf[i]}), i == len(buf)-1, nil)
			require.NoError(t, err)
			require.Nil(t, hm)
			require.NotNil(t, bm)
			require.NotNil(t, bm.Mutation)
			newBody := bm.Mutation.(*extprocv3.BodyMutation_Body).Body
			if len(newBody) > 0 {
				results = append(results, string(newBody))
			}
			if tokenUsage.OutputTokens > 0 {
				require.Equal(t, uint32(75), tokenUsage.OutputTokens)
			}
		}

		result := strings.Join(results, "")

		require.Equal(t,
			`data: {"choices":[{"index":0,"delta":{"content":"","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":"To","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":" calculate the cosine","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":" of 7,","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":" we can use the","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":" \"","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":"cosine\" function","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":" that","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":" is","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":" available to","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":" us.","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":" Let","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":"'s use","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":" this","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":" function to","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":" get","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":" the result","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":".","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"id":"tooluse_QklrEHKjRu6Oc4BQUfy7ZQ","function":{"arguments":"","name":"cosine"},"type":"function"}]}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"id":null,"function":{"arguments":"","name":""},"type":"function"}]}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"id":null,"function":{"arguments":"{\"x\": 7}","name":""},"type":"function"}]}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":"","role":"assistant"},"finish_reason":"tool_calls"}],"object":"chat.completion.chunk"}

data: {"object":"chat.completion.chunk","usage":{"completion_tokens":75,"prompt_tokens":386,"total_tokens":461}}

data: [DONE]
`, result)
	})
}

func TestOpenAIToAWSBedrockTranslator_ResponseError(t *testing.T) {
	tests := []struct {
		name            string
		responseHeaders map[string]string
		input           io.Reader
		output          openai.Error
	}{
		{
			name: "test unhealthy upstream",
			responseHeaders: map[string]string{
				":status":      "503",
				"content-type": "text/plain",
			},
			input: bytes.NewBuffer([]byte("service not available")),
			output: openai.Error{
				Type: "error",
				Error: openai.ErrorType{
					Type:    awsBedrockBackendError,
					Code:    ptr.To("503"),
					Message: "service not available",
				},
			},
		},
		{
			name: "test AWS throttled error response",
			responseHeaders: map[string]string{
				":status":              "429",
				"content-type":         "application/json",
				awsErrorTypeHeaderName: "ThrottledException",
			},
			input: bytes.NewBuffer([]byte(`{"message": "aws bedrock rate limit exceeded"}`)),
			output: openai.Error{
				Type: "error",
				Error: openai.ErrorType{
					Type:    "ThrottledException",
					Code:    ptr.To("429"),
					Message: "aws bedrock rate limit exceeded",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := json.Marshal(tt.input)
			require.NoError(t, err)

			o := &openAIToAWSBedrockTranslatorV1ChatCompletion{}
			hm, bm, err := o.ResponseError(tt.responseHeaders, tt.input)
			require.NoError(t, err)
			require.NotNil(t, bm)
			require.NotNil(t, bm.Mutation)
			require.NotNil(t, bm.Mutation.(*extprocv3.BodyMutation_Body))
			newBody := bm.Mutation.(*extprocv3.BodyMutation_Body).Body
			require.NotNil(t, newBody)
			require.NotNil(t, hm)
			require.NotNil(t, hm.SetHeaders)
			require.Len(t, hm.SetHeaders, 1)
			require.Equal(t, "content-length", hm.SetHeaders[0].Header.Key)
			require.Equal(t, strconv.Itoa(len(newBody)), string(hm.SetHeaders[0].Header.RawValue))

			var openAIError openai.Error
			err = json.Unmarshal(newBody, &openAIError)
			require.NoError(t, err)
			if !cmp.Equal(openAIError, tt.output) {
				t.Errorf("ConvertAWSBedrockErrorResp(), diff(got, expected) = %s\n", cmp.Diff(openAIError, tt.output))
			}
		})
	}
}

func TestOpenAIToAWSBedrockTranslatorV1ChatCompletion_ResponseBody(t *testing.T) {
	t.Run("invalid body", func(t *testing.T) {
		o := &openAIToAWSBedrockTranslatorV1ChatCompletion{}
		_, _, _, _, err := o.ResponseBody(nil, bytes.NewBuffer([]byte("invalid")), false, nil)
		require.Error(t, err)
	})
	tests := []struct {
		name   string
		input  awsbedrock.ConverseResponse
		output openai.ChatCompletionResponse
	}{
		{
			name: "basic_testing",
			input: awsbedrock.ConverseResponse{
				Usage: &awsbedrock.TokenUsage{
					InputTokens:  10,
					OutputTokens: 20,
					TotalTokens:  30,
				},
				Output: &awsbedrock.ConverseOutput{
					Message: awsbedrock.Message{
						Role: "assistant",
						Content: []*awsbedrock.ContentBlock{
							{Text: ptr.To("response")},
							{Text: ptr.To("from")},
							{Text: ptr.To("assistant")},
						},
					},
				},
			},
			output: openai.ChatCompletionResponse{
				Object: "chat.completion",
				Usage: openai.ChatCompletionResponseUsage{
					TotalTokens:      30,
					PromptTokens:     10,
					CompletionTokens: 20,
				},
				Choices: []openai.ChatCompletionResponseChoice{
					{
						Index: 0,
						Message: openai.ChatCompletionResponseChoiceMessage{
							Content: ptr.To("response"),
							Role:    awsbedrock.ConversationRoleAssistant,
						},
						FinishReason: openai.ChatCompletionChoicesFinishReasonStop,
					},
				},
			},
		},
		{
			name: "test stop reason",
			input: awsbedrock.ConverseResponse{
				Usage: &awsbedrock.TokenUsage{
					InputTokens:  10,
					OutputTokens: 20,
					TotalTokens:  30,
				},
				StopReason: ptr.To("stop_sequence"),
				Output: &awsbedrock.ConverseOutput{
					Message: awsbedrock.Message{
						Role: awsbedrock.ConversationRoleAssistant,
						Content: []*awsbedrock.ContentBlock{
							{Text: ptr.To("response")},
						},
					},
				},
			},
			output: openai.ChatCompletionResponse{
				Object: "chat.completion",
				Usage: openai.ChatCompletionResponseUsage{
					TotalTokens:      30,
					PromptTokens:     10,
					CompletionTokens: 20,
				},
				Choices: []openai.ChatCompletionResponseChoice{
					{
						Index:        0,
						FinishReason: openai.ChatCompletionChoicesFinishReasonStop,
						Message: openai.ChatCompletionResponseChoiceMessage{
							Content: ptr.To("response"),
							Role:    awsbedrock.ConversationRoleAssistant,
						},
					},
				},
			},
		},
		{
			name: "test tool use",
			input: awsbedrock.ConverseResponse{
				StopReason: ptr.To(awsbedrock.StopReasonToolUse),
				Output: &awsbedrock.ConverseOutput{
					Message: awsbedrock.Message{
						Role: awsbedrock.ConversationRoleAssistant,
						// Text and ToolUse are sent in two different content blocks for AWS Bedrock, OpenAI merges them in one message.
						Content: []*awsbedrock.ContentBlock{
							{
								Text: ptr.To("response"),
							},
							{
								ToolUse: &awsbedrock.ToolUseBlock{
									Name:      "exec_python_code",
									ToolUseID: "call_6g7a",
									Input:     map[string]any{"code_block": "from playwright.sync_api import sync_playwright\n"},
								},
							},
						},
					},
				},
			},
			output: openai.ChatCompletionResponse{
				Object: "chat.completion",
				Choices: []openai.ChatCompletionResponseChoice{
					{
						Index:        0,
						FinishReason: openai.ChatCompletionChoicesFinishReasonToolCalls,
						Message: openai.ChatCompletionResponseChoiceMessage{
							Content: ptr.To("response"),
							Role:    awsbedrock.ConversationRoleAssistant,
							ToolCalls: []openai.ChatCompletionMessageToolCallParam{
								{
									ID: ptr.To("call_6g7a"),
									Function: openai.ChatCompletionMessageToolCallFunctionParam{
										Name:      "exec_python_code",
										Arguments: "{\"code_block\":\"from playwright.sync_api import sync_playwright\\n\"}",
									},
									Type: openai.ChatCompletionMessageToolCallTypeFunction,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "merge content",
			input: awsbedrock.ConverseResponse{
				Usage: &awsbedrock.TokenUsage{
					InputTokens:  10,
					OutputTokens: 20,
					TotalTokens:  30,
				},
				Output: &awsbedrock.ConverseOutput{
					Message: awsbedrock.Message{
						Role: awsbedrock.ConversationRoleAssistant,
						Content: []*awsbedrock.ContentBlock{
							{Text: ptr.To("response")},
							{ToolUse: &awsbedrock.ToolUseBlock{
								Name:      "exec_python_code",
								ToolUseID: "call_6g7a",
								Input:     map[string]any{"code_block": "from playwright.sync_api import sync_playwright\n"},
							}},
						},
					},
				},
			},
			output: openai.ChatCompletionResponse{
				Object: "chat.completion",
				Usage: openai.ChatCompletionResponseUsage{
					TotalTokens:      30,
					PromptTokens:     10,
					CompletionTokens: 20,
				},
				Choices: []openai.ChatCompletionResponseChoice{
					{
						Index: 0,
						Message: openai.ChatCompletionResponseChoiceMessage{
							Content: ptr.To("response"),
							Role:    awsbedrock.ConversationRoleAssistant,
							ToolCalls: []openai.ChatCompletionMessageToolCallParam{
								{
									ID: ptr.To("call_6g7a"),
									Function: openai.ChatCompletionMessageToolCallFunctionParam{
										Name:      "exec_python_code",
										Arguments: "{\"code_block\":\"from playwright.sync_api import sync_playwright\\n\"}",
									},
									Type: openai.ChatCompletionMessageToolCallTypeFunction,
								},
							},
						},
						FinishReason: openai.ChatCompletionChoicesFinishReasonStop,
					},
				},
			},
		},
		{
			name: "response with reasoning content",
			input: awsbedrock.ConverseResponse{
				StopReason: ptr.To(awsbedrock.StopReasonEndTurn),
				Output: &awsbedrock.ConverseOutput{
					Message: awsbedrock.Message{
						Role: awsbedrock.ConversationRoleAssistant,
						Content: []*awsbedrock.ContentBlock{
							{
								ReasoningContent: &awsbedrock.ReasoningContentBlock{
									ReasoningText: &awsbedrock.ReasoningTextBlock{
										Text: "This is the model's thought process.",
									},
								},
							},
							{
								Text: ptr.To("This is the final answer."),
							},
						},
					},
				},
			},
			output: openai.ChatCompletionResponse{
				Object: "chat.completion",
				Choices: []openai.ChatCompletionResponseChoice{
					{
						Index:        0,
						FinishReason: openai.ChatCompletionChoicesFinishReasonStop,
						Message: openai.ChatCompletionResponseChoiceMessage{
							Role:    awsbedrock.ConversationRoleAssistant,
							Content: ptr.To("This is the final answer."),
							ReasoningContent: &openai.ReasoningContentUnion{
								Value: &openai.AWSBedrockReasoningContent{
									ReasoningContent: &awsbedrock.ReasoningContentBlock{
										ReasoningText: &awsbedrock.ReasoningTextBlock{
											Text: "This is the model's thought process.",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.input)
			require.NoError(t, err)

			o := &openAIToAWSBedrockTranslatorV1ChatCompletion{}
			hm, bm, usedToken, _, err := o.ResponseBody(nil, bytes.NewBuffer(body), false, nil)
			require.NoError(t, err)
			require.NotNil(t, bm)
			require.NotNil(t, bm.Mutation)
			require.NotNil(t, bm.Mutation.(*extprocv3.BodyMutation_Body))
			newBody := bm.Mutation.(*extprocv3.BodyMutation_Body).Body
			require.NotNil(t, newBody)
			require.NotNil(t, hm)
			require.NotNil(t, hm.SetHeaders)
			require.Len(t, hm.SetHeaders, 1)
			require.Equal(t, "content-length", hm.SetHeaders[0].Header.Key)
			require.Equal(t, strconv.Itoa(len(newBody)), string(hm.SetHeaders[0].Header.RawValue))

			expectedBody, err := json.Marshal(tt.output)
			require.NoError(t, err)
			require.JSONEq(t, string(expectedBody), string(newBody))
			require.Equal(t,
				LLMTokenUsage{
					InputTokens:  uint32(tt.output.Usage.PromptTokens),     //nolint:gosec
					OutputTokens: uint32(tt.output.Usage.CompletionTokens), //nolint:gosec
					TotalTokens:  uint32(tt.output.Usage.TotalTokens),      //nolint:gosec
				}, usedToken)
		})
	}
}

// base64RealStreamingEvents is the base64 encoded raw binary response from bedrock anthropic.claude model.
// The request is to find the cosine of number 7 with a tool configuration.
const base64RealStreamingEvents = "AAAAmwAAAFJGkfmwCzpldmVudC10eXBlBwAMbWVzc2FnZVN0YXJ0DTpjb250ZW50LXR5cGUHABBhcHBsaWNhdGlvbi9qc29uDTptZXNzYWdlLXR5cGUHAAVldmVudHsicCI6ImFiY2RlZmdoaWprbG1ub3BxcnN0dXZ3eHl6QUJDRCIsInJvbGUiOiJhc3Npc3RhbnQifbCidJ0AAACpAAAAV+0a5tkLOmV2ZW50LXR5cGUHABFjb250ZW50QmxvY2tEZWx0YQ06Y29udGVudC10eXBlBwAQYXBwbGljYXRpb24vanNvbg06bWVzc2FnZS10eXBlBwAFZXZlbnR7ImNvbnRlbnRCbG9ja0luZGV4IjowLCJkZWx0YSI6eyJ0ZXh0IjoiVG8ifSwicCI6ImFiY2RlZmdoaWprbG1uIn0rY75JAAAAsQAAAFe9ijqaCzpldmVudC10eXBlBwARY29udGVudEJsb2NrRGVsdGENOmNvbnRlbnQtdHlwZQcAEGFwcGxpY2F0aW9uL2pzb24NOm1lc3NhZ2UtdHlwZQcABWV2ZW50eyJjb250ZW50QmxvY2tJbmRleCI6MCwiZGVsdGEiOnsidGV4dCI6IiBjYWxjdWxhdGUgdGhlIGNvc2luZSJ9LCJwIjoiYWJjIn3hywqfAAAA2gAAAFdTaHzGCzpldmVudC10eXBlBwARY29udGVudEJsb2NrRGVsdGENOmNvbnRlbnQtdHlwZQcAEGFwcGxpY2F0aW9uL2pzb24NOm1lc3NhZ2UtdHlwZQcABWV2ZW50eyJjb250ZW50QmxvY2tJbmRleCI6MCwiZGVsdGEiOnsidGV4dCI6IiBvZiA3LCJ9LCJwIjoiYWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXpBQkNERUZHSElKS0xNTk9QUVJTVFVWV1hZWjAxMjM0NTYifUsRwHsAAADXAAAAV6v4uHcLOmV2ZW50LXR5cGUHABFjb250ZW50QmxvY2tEZWx0YQ06Y29udGVudC10eXBlBwAQYXBwbGljYXRpb24vanNvbg06bWVzc2FnZS10eXBlBwAFZXZlbnR7ImNvbnRlbnRCbG9ja0luZGV4IjowLCJkZWx0YSI6eyJ0ZXh0IjoiIHdlIGNhbiB1c2UgdGhlIn0sInAiOiJhYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5ekFCQ0RFRkdISUpLTE1OT1BRUlNUVSJ9jBuxjAAAALwAAABXRRr+Kws6ZXZlbnQtdHlwZQcAEWNvbnRlbnRCbG9ja0RlbHRhDTpjb250ZW50LXR5cGUHABBhcHBsaWNhdGlvbi9qc29uDTptZXNzYWdlLXR5cGUHAAVldmVudHsiY29udGVudEJsb2NrSW5kZXgiOjAsImRlbHRhIjp7InRleHQiOiIgXCIifSwicCI6ImFiY2RlZmdoaWprbG1ub3BxcnN0dXZ3eHl6QUJDREVGIn3SOp66AAAA2wAAAFduCFV2CzpldmVudC10eXBlBwARY29udGVudEJsb2NrRGVsdGENOmNvbnRlbnQtdHlwZQcAEGFwcGxpY2F0aW9uL2pzb24NOm1lc3NhZ2UtdHlwZQcABWV2ZW50eyJjb250ZW50QmxvY2tJbmRleCI6MCwiZGVsdGEiOnsidGV4dCI6ImNvc2luZVwiIGZ1bmN0aW9uIn0sInAiOiJhYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5ekFCQ0RFRkdISUpLTE1OT1BRUlNUVVZXIn2f+1UQAAAA2QAAAFcUyAYWCzpldmVudC10eXBlBwARY29udGVudEJsb2NrRGVsdGENOmNvbnRlbnQtdHlwZQcAEGFwcGxpY2F0aW9uL2pzb24NOm1lc3NhZ2UtdHlwZQcABWV2ZW50eyJjb250ZW50QmxvY2tJbmRleCI6MCwiZGVsdGEiOnsidGV4dCI6IiB0aGF0In0sInAiOiJhYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5ekFCQ0RFRkdISUpLTE1OT1BRUlNUVVZXWFlaMDEyMzQ1NiJ9uD7t8wAAAM8AAABX+2hkNAs6ZXZlbnQtdHlwZQcAEWNvbnRlbnRCbG9ja0RlbHRhDTpjb250ZW50LXR5cGUHABBhcHBsaWNhdGlvbi9qc29uDTptZXNzYWdlLXR5cGUHAAVldmVudHsiY29udGVudEJsb2NrSW5kZXgiOjAsImRlbHRhIjp7InRleHQiOiIgaXMifSwicCI6ImFiY2RlZmdoaWprbG1ub3BxcnN0dXZ3eHl6QUJDREVGR0hJSktMTU5PUFFSU1RVVldYWSJ9p52nrQAAAMQAAABXjLhVJQs6ZXZlbnQtdHlwZQcAEWNvbnRlbnRCbG9ja0RlbHRhDTpjb250ZW50LXR5cGUHABBhcHBsaWNhdGlvbi9qc29uDTptZXNzYWdlLXR5cGUHAAVldmVudHsiY29udGVudEJsb2NrSW5kZXgiOjAsImRlbHRhIjp7InRleHQiOiIgYXZhaWxhYmxlIHRvIn0sInAiOiJhYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5ekFCQ0QifYC08b0AAADTAAAAV154HrcLOmV2ZW50LXR5cGUHABFjb250ZW50QmxvY2tEZWx0YQ06Y29udGVudC10eXBlBwAQYXBwbGljYXRpb24vanNvbg06bWVzc2FnZS10eXBlBwAFZXZlbnR7ImNvbnRlbnRCbG9ja0luZGV4IjowLCJkZWx0YSI6eyJ0ZXh0IjoiIHVzLiJ9LCJwIjoiYWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXpBQkNERUZHSElKS0xNTk9QUVJTVFVWV1hZWjAxIn0mTm4jAAAAtAAAAFd1arXqCzpldmVudC10eXBlBwARY29udGVudEJsb2NrRGVsdGENOmNvbnRlbnQtdHlwZQcAEGFwcGxpY2F0aW9uL2pzb24NOm1lc3NhZ2UtdHlwZQcABWV2ZW50eyJjb250ZW50QmxvY2tJbmRleCI6MCwiZGVsdGEiOnsidGV4dCI6IiBMZXQifSwicCI6ImFiY2RlZmdoaWprbG1ub3BxcnN0dXZ3In34BFwTAAAA0AAAAFcZ2GRnCzpldmVudC10eXBlBwARY29udGVudEJsb2NrRGVsdGENOmNvbnRlbnQtdHlwZQcAEGFwcGxpY2F0aW9uL2pzb24NOm1lc3NhZ2UtdHlwZQcABWV2ZW50eyJjb250ZW50QmxvY2tJbmRleCI6MCwiZGVsdGEiOnsidGV4dCI6IidzIHVzZSJ9LCJwIjoiYWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXpBQkNERUZHSElKS0xNTk9QUVJTVFVWVyJ9vfdBjQAAALwAAABXRRr+Kws6ZXZlbnQtdHlwZQcAEWNvbnRlbnRCbG9ja0RlbHRhDTpjb250ZW50LXR5cGUHABBhcHBsaWNhdGlvbi9qc29uDTptZXNzYWdlLXR5cGUHAAVldmVudHsiY29udGVudEJsb2NrSW5kZXgiOjAsImRlbHRhIjp7InRleHQiOiIgdGhpcyJ9LCJwIjoiYWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXpBQkNEIn1Xtb4jAAAAuAAAAFewmljrCzpldmVudC10eXBlBwARY29udGVudEJsb2NrRGVsdGENOmNvbnRlbnQtdHlwZQcAEGFwcGxpY2F0aW9uL2pzb24NOm1lc3NhZ2UtdHlwZQcABWV2ZW50eyJjb250ZW50QmxvY2tJbmRleCI6MCwiZGVsdGEiOnsidGV4dCI6IiBmdW5jdGlvbiB0byJ9LCJwIjoiYWJjZGVmZ2hpamtsbW5vcHFycyJ9GYv84AAAALQAAABXdWq16gs6ZXZlbnQtdHlwZQcAEWNvbnRlbnRCbG9ja0RlbHRhDTpjb250ZW50LXR5cGUHABBhcHBsaWNhdGlvbi9qc29uDTptZXNzYWdlLXR5cGUHAAVldmVudHsiY29udGVudEJsb2NrSW5kZXgiOjAsImRlbHRhIjp7InRleHQiOiIgZ2V0In0sInAiOiJhYmNkZWZnaGlqa2xtbm9wcXJzdHV2dyJ99bdUOgAAAN4AAABXpujaBgs6ZXZlbnQtdHlwZQcAEWNvbnRlbnRCbG9ja0RlbHRhDTpjb250ZW50LXR5cGUHABBhcHBsaWNhdGlvbi9qc29uDTptZXNzYWdlLXR5cGUHAAVldmVudHsiY29udGVudEJsb2NrSW5kZXgiOjAsImRlbHRhIjp7InRleHQiOiIgdGhlIHJlc3VsdCJ9LCJwIjoiYWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXpBQkNERUZHSElKS0xNTk9QUVJTVFVWV1hZWjAxMjM0NSJ9niPS/gAAAM0AAABXgag3VAs6ZXZlbnQtdHlwZQcAEWNvbnRlbnRCbG9ja0RlbHRhDTpjb250ZW50LXR5cGUHABBhcHBsaWNhdGlvbi9qc29uDTptZXNzYWdlLXR5cGUHAAVldmVudHsiY29udGVudEJsb2NrSW5kZXgiOjAsImRlbHRhIjp7InRleHQiOiIuIn0sInAiOiJhYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5ekFCQ0RFRkdISUpLTE1OT1BRUlNUVVZXWFkifRc68JQAAACuAAAAVig9Cl8LOmV2ZW50LXR5cGUHABBjb250ZW50QmxvY2tTdG9wDTpjb250ZW50LXR5cGUHABBhcHBsaWNhdGlvbi9qc29uDTptZXNzYWdlLXR5cGUHAAVldmVudHsiY29udGVudEJsb2NrSW5kZXgiOjAsInAiOiJhYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5ekFCQ0RFRkdISUpLTE1OT1AifY2eizoAAAEEAAAAV67xblsLOmV2ZW50LXR5cGUHABFjb250ZW50QmxvY2tTdGFydA06Y29udGVudC10eXBlBwAQYXBwbGljYXRpb24vanNvbg06bWVzc2FnZS10eXBlBwAFZXZlbnR7ImNvbnRlbnRCbG9ja0luZGV4IjoxLCJwIjoiYWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXpBQkNERUZHSElKS0xNTk9QUVIiLCJzdGFydCI6eyJ0b29sVXNlIjp7Im5hbWUiOiJjb3NpbmUiLCJ0b29sVXNlSWQiOiJ0b29sdXNlX1FrbHJFSEtqUnU2T2M0QlFVZnk3WlEifX19kpNGawAAAK0AAABXGJpAGQs6ZXZlbnQtdHlwZQcAEWNvbnRlbnRCbG9ja0RlbHRhDTpjb250ZW50LXR5cGUHABBhcHBsaWNhdGlvbi9qc29uDTptZXNzYWdlLXR5cGUHAAVldmVudHsiY29udGVudEJsb2NrSW5kZXgiOjEsImRlbHRhIjp7InRvb2xVc2UiOnsiaW5wdXQiOiIifX0sInAiOiJhYmNkZWZnIn3XeK+kAAAAswAAAFfHSmn6CzpldmVudC10eXBlBwARY29udGVudEJsb2NrRGVsdGENOmNvbnRlbnQtdHlwZQcAEGFwcGxpY2F0aW9uL2pzb24NOm1lc3NhZ2UtdHlwZQcABWV2ZW50eyJjb250ZW50QmxvY2tJbmRleCI6MSwiZGVsdGEiOnsidG9vbFVzZSI6eyJpbnB1dCI6IntcInhcIjogN30ifX0sInAiOiJhYmMifaN4jhsAAACxAAAAVsqNCgwLOmV2ZW50LXR5cGUHABBjb250ZW50QmxvY2tTdG9wDTpjb250ZW50LXR5cGUHABBhcHBsaWNhdGlvbi9qc29uDTptZXNzYWdlLXR5cGUHAAVldmVudHsiY29udGVudEJsb2NrSW5kZXgiOjEsInAiOiJhYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5ekFCQ0RFRkdISUpLTE1OT1BRUlMifUJp3UkAAACFAAAAUQBIgekLOmV2ZW50LXR5cGUHAAttZXNzYWdlU3RvcA06Y29udGVudC10eXBlBwAQYXBwbGljYXRpb24vanNvbg06bWVzc2FnZS10eXBlBwAFZXZlbnR7InAiOiJhYmNkIiwic3RvcFJlYXNvbiI6InRvb2xfdXNlIn3ejv14AAAAygAAAE5X40OECzpldmVudC10eXBlBwAIbWV0YWRhdGENOmNvbnRlbnQtdHlwZQcAEGFwcGxpY2F0aW9uL2pzb24NOm1lc3NhZ2UtdHlwZQcABWV2ZW50eyJtZXRyaWNzIjp7ImxhdGVuY3lNcyI6MTk1N30sInAiOiJhYmNkZWZnIiwidXNhZ2UiOnsiaW5wdXRUb2tlbnMiOjM4Niwib3V0cHV0VG9rZW5zIjo3NSwidG90YWxUb2tlbnMiOjQ2MX19Ke/W4Q=="

func TestOpenAIToAWSBedrockTranslatorExtractAmazonEventStreamEvents(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	e := eventstream.NewEncoder()
	var offsets []int
	for _, data := range []awsbedrock.ConverseStreamEvent{
		{Delta: &awsbedrock.ConverseStreamEventContentBlockDelta{Text: ptr.To("1")}},
		{Delta: &awsbedrock.ConverseStreamEventContentBlockDelta{Text: ptr.To("2")}},
		{Delta: &awsbedrock.ConverseStreamEventContentBlockDelta{Text: ptr.To("3")}},
	} {
		offsets = append(offsets, buf.Len())
		eventPayload, err := json.Marshal(data)
		require.NoError(t, err)
		err = e.Encode(buf, eventstream.Message{
			Headers: eventstream.Headers{{Name: "event-type", Value: eventstream.StringValue("content")}},
			Payload: eventPayload,
		})
		require.NoError(t, err)
	}

	eventBytes := buf.Bytes()

	t.Run("all-at-once", func(t *testing.T) {
		o := &openAIToAWSBedrockTranslatorV1ChatCompletion{}
		o.bufferedBody = eventBytes
		o.extractAmazonEventStreamEvents()
		require.Len(t, o.events, 3)
		require.Empty(t, o.bufferedBody)
		for i, text := range []string{"1", "2", "3"} {
			require.Equal(t, text, *o.events[i].Delta.Text)
		}
	})

	t.Run("in-chunks", func(t *testing.T) {
		o := &openAIToAWSBedrockTranslatorV1ChatCompletion{}
		o.bufferedBody = eventBytes[0:1]
		o.extractAmazonEventStreamEvents()
		require.Empty(t, o.events)
		require.Len(t, o.bufferedBody, 1)

		o.bufferedBody = eventBytes[0 : offsets[1]+5]
		o.extractAmazonEventStreamEvents()
		require.Len(t, o.events, 1)
		require.Equal(t, eventBytes[offsets[1]:offsets[1]+5], o.bufferedBody)

		clear(o.events)
		o.events = o.events[:0]
		o.bufferedBody = eventBytes[0 : offsets[2]+5]
		o.extractAmazonEventStreamEvents()
		require.Len(t, o.events, 2)
		require.Equal(t, eventBytes[offsets[2]:offsets[2]+5], o.bufferedBody)
	})

	t.Run("real events", func(t *testing.T) {
		o := &openAIToAWSBedrockTranslatorV1ChatCompletion{}
		var err error
		o.bufferedBody, err = base64.StdEncoding.DecodeString(base64RealStreamingEvents)
		require.NoError(t, err)
		o.extractAmazonEventStreamEvents()

		var texts []string
		var usage *awsbedrock.TokenUsage
		for _, event := range o.events {
			if delta := event.Delta; delta != nil && delta.Text != nil && *delta.Text != "" {
				texts = append(texts, *event.Delta.Text)
			}
			if u := event.Usage; u != nil {
				usage = u
			}
		}
		require.Equal(t,
			"To calculate the cosine of 7, we can use the \"cosine\" function that is available to us. Let's use this function to get the result.",
			strings.Join(texts, ""),
		)
		require.NotNil(t, usage)
		require.Equal(t, 461, usage.TotalTokens)
	})
}

func TestOpenAIToAWSBedrockTranslator_convertEvent(t *testing.T) {
	ptrOf := func(s string) *string { return &s }
	for _, tc := range []struct {
		name string
		in   awsbedrock.ConverseStreamEvent
		out  *openai.ChatCompletionResponseChunk
	}{
		{
			name: "usage",
			in: awsbedrock.ConverseStreamEvent{
				Usage: &awsbedrock.TokenUsage{
					InputTokens:  10,
					OutputTokens: 20,
					TotalTokens:  30,
				},
			},
			out: &openai.ChatCompletionResponseChunk{
				Object: "chat.completion.chunk",
				Usage: &openai.ChatCompletionResponseUsage{
					TotalTokens:      30,
					PromptTokens:     10,
					CompletionTokens: 20,
				},
			},
		},
		{
			name: "role",
			in: awsbedrock.ConverseStreamEvent{
				Role: ptrOf(awsbedrock.ConversationRoleAssistant),
			},
			out: &openai.ChatCompletionResponseChunk{
				Object: "chat.completion.chunk",
				Choices: []openai.ChatCompletionResponseChunkChoice{
					{
						Delta: &openai.ChatCompletionResponseChunkChoiceDelta{
							Role:    awsbedrock.ConversationRoleAssistant,
							Content: &emptyString,
						},
					},
				},
			},
		},
		{
			name: "delta",
			in: awsbedrock.ConverseStreamEvent{
				Delta: &awsbedrock.ConverseStreamEventContentBlockDelta{Text: ptr.To("response")},
			},
			out: &openai.ChatCompletionResponseChunk{
				Object: "chat.completion.chunk",
				Choices: []openai.ChatCompletionResponseChunkChoice{
					{
						Delta: &openai.ChatCompletionResponseChunkChoiceDelta{
							Content: ptrOf("response"),
						},
					},
				},
			},
		},
		{
			name: "reasoning delta",
			in: awsbedrock.ConverseStreamEvent{
				Delta: &awsbedrock.ConverseStreamEventContentBlockDelta{
					ReasoningContent: &awsbedrock.ReasoningContentBlockDelta{
						Text: "thinking...",
					},
				},
			},
			out: &openai.ChatCompletionResponseChunk{
				Object: "chat.completion.chunk",
				Choices: []openai.ChatCompletionResponseChunkChoice{
					{
						Delta: &openai.ChatCompletionResponseChunkChoiceDelta{
							ReasoningContent: &openai.AWSBedrockStreamReasoningContent{
								Text: "thinking...",
							},
						},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := &openAIToAWSBedrockTranslatorV1ChatCompletion{}
			chunk, ok := o.convertEvent(&tc.in)
			if tc.out == nil {
				require.False(t, ok)
			} else {
				require.Equal(t, tc.out, chunk)
			}
		})
	}
}

func TestOpenAIToAWSBedrockTranslatorV1ChatCompletion_Streaming_WithReasoning(t *testing.T) {
	inputEvents := []awsbedrock.ConverseStreamEvent{
		{Role: ptr.To(awsbedrock.ConversationRoleAssistant)},
		{
			ContentBlockIndex: 0,
			Delta: &awsbedrock.ConverseStreamEventContentBlockDelta{
				ReasoningContent: &awsbedrock.ReasoningContentBlockDelta{
					Text: "Okay, 27 * 453. ",
				},
			},
		},
		{
			ContentBlockIndex: 0,
			Delta: &awsbedrock.ConverseStreamEventContentBlockDelta{
				ReasoningContent: &awsbedrock.ReasoningContentBlockDelta{
					Text: "Let's do the math...",
				},
			},
		},
		{
			ContentBlockIndex: 0,
			Delta: &awsbedrock.ConverseStreamEventContentBlockDelta{
				Text: ptr.To("The result of 27 multiplied by 453 is "),
			},
		},
		{
			ContentBlockIndex: 0,
			Delta: &awsbedrock.ConverseStreamEventContentBlockDelta{
				Text: ptr.To("12231."),
			},
		},
		{StopReason: ptr.To(awsbedrock.StopReasonEndTurn)},
	}

	buf := bytes.NewBuffer(nil)
	encoder := eventstream.NewEncoder()
	for _, event := range inputEvents {
		payload, err := json.Marshal(event)
		require.NoError(t, err)

		// This header is a simplification; the real stream has more, but this is sufficient for the decoder.
		err = encoder.Encode(buf, eventstream.Message{
			Headers: eventstream.Headers{
				{Name: ":event-type", Value: eventstream.StringValue("chunk")},
			},
			Payload: payload,
		})
		require.NoError(t, err)
	}
	// Process the entire encoded stream through the translator.
	o := &openAIToAWSBedrockTranslatorV1ChatCompletion{stream: true}
	_, bm, _, _, err := o.ResponseBody(nil, buf, true, nil)
	require.NoError(t, err)
	require.NotNil(t, bm)

	// Parse the translated SSE (Server-Sent Events) output.
	outputBody := bm.Mutation.(*extprocv3.BodyMutation_Body).Body
	lines := strings.Split(string(outputBody), "\n")

	var generationChunks []string
	var reasoningChunks []string

	for _, line := range lines {
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if strings.Contains(data, "[DONE]") {
				continue
			}

			var chunk openai.ChatCompletionResponseChunk
			err := json.Unmarshal([]byte(data), &chunk)
			require.NoError(t, err)

			if len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta
				if delta != nil {
					if delta.Content != nil && *delta.Content != "" {
						generationChunks = append(generationChunks, *delta.Content)
					}
					if delta.ReasoningContent != nil && delta.ReasoningContent.Text != "" {
						reasoningChunks = append(reasoningChunks, delta.ReasoningContent.Text)

						var untypedChunk map[string]interface{}
						err = json.Unmarshal([]byte(data), &untypedChunk)
						require.NoError(t, err)

						choices, _ := untypedChunk["choices"].([]interface{})
						choice, _ := choices[0].(map[string]interface{})
						deltaMap, _ := choice["delta"].(map[string]interface{})
						reasoningContent, ok := deltaMap["reasoning_content"].(map[string]interface{})
						require.True(t, ok, "Delta should have a 'reasoning_content' map")
						_, textOk := reasoningContent["text"]
						require.True(t, textOk, "Reasoning content should have a 'text' key")
					}
				}
			}
		}
	}

	generation := strings.Join(generationChunks, "")
	reasoning := strings.Join(reasoningChunks, "")

	require.Equal(t, "The result of 27 multiplied by 453 is 12231.", generation)
	require.Equal(t, "Okay, 27 * 453. Let's do the math...", reasoning)

	require.NotEmpty(t, generation, "Generation content should not be empty")
	require.NotEmpty(t, reasoning, "Reasoning content should not be empty")
}

func TestOpenAIToAWSBedrockTranslatorV1ChatCompletion_ResponseBody_WithReasoning(t *testing.T) {
	// Define a mock AWS Bedrock non-streaming response.
	// This response contains two content blocks: one for reasoning and one for the final text answer.
	mockBedrockResponse := awsbedrock.ConverseResponse{
		Output: &awsbedrock.ConverseOutput{
			Message: awsbedrock.Message{
				Role: awsbedrock.ConversationRoleAssistant,
				Content: []*awsbedrock.ContentBlock{
					{
						ReasoningContent: &awsbedrock.ReasoningContentBlock{
							ReasoningText: &awsbedrock.ReasoningTextBlock{
								Text: "The user wants to compare two numbers. 9.11 is larger than 9.8.",
							},
						},
					},
					{
						Text: ptr.To("9.11 is greater than 9.8."),
					},
				},
			},
		},
		StopReason: ptr.To(awsbedrock.StopReasonEndTurn),
	}

	body, err := json.Marshal(mockBedrockResponse)
	require.NoError(t, err)

	o := &openAIToAWSBedrockTranslatorV1ChatCompletion{}
	_, bm, _, _, err := o.ResponseBody(nil, bytes.NewBuffer(body), false, nil)
	require.NoError(t, err)
	require.NotNil(t, bm)

	outputBody := bm.Mutation.(*extprocv3.BodyMutation_Body).Body
	var openAIResponse openai.ChatCompletionResponse
	err = json.Unmarshal(outputBody, &openAIResponse)
	require.NoError(t, err)

	// Assert that the translated response contains both content and reasoning.
	require.Len(t, openAIResponse.Choices, 1, "There should be exactly one choice in the response")
	message := openAIResponse.Choices[0].Message

	require.NotNil(t, message.Content, "Message content should not be nil")
	require.NotEmpty(t, *message.Content, "Message content should not be an empty string")
	require.Equal(t, "9.11 is greater than 9.8.", *message.Content)

	require.NotNil(t, message.ReasoningContent, "Reasoning content should not be nil")
	reasoningBlock, _ := message.ReasoningContent.Value.(*openai.AWSBedrockReasoningContent)
	require.NotNil(t, reasoningBlock, "The nested reasoning content block should not be nil")
	require.NotEmpty(t, reasoningBlock.ReasoningContent.ReasoningText.Text, "The reasoning text itself should not be empty")

	var untypedResponse map[string]interface{}
	err = json.Unmarshal(outputBody, &untypedResponse)
	require.NoError(t, err)

	// Traverse the map to verify the structure: reasoning_content["reasoningContent"].
	choices, ok := untypedResponse["choices"].([]interface{})
	require.True(t, ok, "JSON should have a 'choices' array")
	require.Len(t, choices, 1)

	choice, ok := choices[0].(map[string]interface{})
	require.True(t, ok, "Choice item should be a map")

	messageMap, ok := choice["message"].(map[string]interface{})
	require.True(t, ok, "Choice should have a 'message' map")
	reasoningContent, ok := messageMap["reasoning_content"].(map[string]interface{})
	require.True(t, ok, "Message should have a 'reasoning_content' map")

	_, ok = reasoningContent["reasoningContent"]
	require.True(t, ok, "The 'reasoning_content' object should have a nested 'reasoningContent' key")
}

func TestOpenAIToAWSBedrockTranslatorV1ChatCompletion_Streaming_WithRedactedContent(t *testing.T) {
	redactedBytes := []byte("a redacted thought")
	inputEvents := []awsbedrock.ConverseStreamEvent{
		{Role: ptr.To(awsbedrock.ConversationRoleAssistant)},
		{Delta: &awsbedrock.ConverseStreamEventContentBlockDelta{
			ReasoningContent: &awsbedrock.ReasoningContentBlockDelta{
				RedactedContent: redactedBytes,
			},
		}},
		{Delta: &awsbedrock.ConverseStreamEventContentBlockDelta{
			Text: ptr.To("This is the final answer."),
		}},
		{StopReason: ptr.To(awsbedrock.StopReasonEndTurn)},
	}

	buf := bytes.NewBuffer(nil)
	encoder := eventstream.NewEncoder()
	for _, event := range inputEvents {
		payload, err := json.Marshal(event)
		require.NoError(t, err)
		err = encoder.Encode(buf, eventstream.Message{
			Headers: eventstream.Headers{{Name: ":event-type", Value: eventstream.StringValue("chunk")}},
			Payload: payload,
		})
		require.NoError(t, err)
	}

	o := &openAIToAWSBedrockTranslatorV1ChatCompletion{stream: true}
	_, bm, _, _, err := o.ResponseBody(nil, buf, true, nil)
	require.NoError(t, err)

	lines := strings.Split(string(bm.Mutation.(*extprocv3.BodyMutation_Body).Body), "\n")
	var foundReasoningChunk bool
	for _, line := range lines {
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if strings.Contains(data, "[DONE]") {
				continue
			}
			var chunk openai.ChatCompletionResponseChunk
			err := json.Unmarshal([]byte(data), &chunk)
			require.NoError(t, err)

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil && chunk.Choices[0].Delta.ReasoningContent != nil {
				reasoning := chunk.Choices[0].Delta.ReasoningContent
				require.Equal(t, redactedBytes, reasoning.RedactedContent)
				require.Empty(t, reasoning.Text)
				foundReasoningChunk = true

				var untypedChunk map[string]interface{}
				err = json.Unmarshal([]byte(data), &untypedChunk)
				require.NoError(t, err)

				choices, _ := untypedChunk["choices"].([]interface{})
				choice, _ := choices[0].(map[string]interface{})
				deltaMap, _ := choice["delta"].(map[string]interface{})
				reasoningContent, ok := deltaMap["reasoning_content"].(map[string]interface{})
				require.True(t, ok, "Delta should have a 'reasoning_content' map")
				_, redactedOk := reasoningContent["redactedContent"]
				require.True(t, redactedOk, "Reasoning content should have a 'redactedContent' key")
			}
		}
	}
	require.True(t, foundReasoningChunk, "A reasoning chunk with redacted content should have been found")
}

// TestResponseModel_AWSBedrock tests that AWS Bedrock returns the request model (no virtualization)
func TestResponseModel_AWSBedrock(t *testing.T) {
	modelName := "anthropic.claude-3-5-sonnet-20241022-v2:0"
	translator := NewChatCompletionOpenAIToAWSBedrockTranslator(modelName)

	// Initialize translator with the model
	req := &openai.ChatCompletionRequest{
		Model: "claude-3-5-sonnet",
	}
	reqBody, _ := json.Marshal(req)
	_, _, err := translator.RequestBody(reqBody, req, false)
	require.NoError(t, err)

	// AWS Bedrock response doesn't have model field
	bedrockResponse := `{
		"output": {
			"message": {
				"content": [{"text": "Hello"}],
				"role": "assistant"
			}
		},
		"usage": {
			"inputTokens": 10,
			"outputTokens": 5,
			"totalTokens": 15
		}
	}`

	_, _, tokenUsage, responseModel, err := translator.ResponseBody(nil, bytes.NewReader([]byte(bedrockResponse)), true, nil)
	require.NoError(t, err)
	require.Equal(t, modelName, responseModel) // Returns the request model since no virtualization
	require.Equal(t, uint32(10), tokenUsage.InputTokens)
	require.Equal(t, uint32(5), tokenUsage.OutputTokens)
}
