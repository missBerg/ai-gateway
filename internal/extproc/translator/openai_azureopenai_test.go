// Copyright Envoy AI Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/ai-gateway/internal/apischema/openai"
)

func TestOpenAIToAzureOpenAITranslatorV1ChatCompletion_RequestBody(t *testing.T) {
	t.Run("valid body", func(t *testing.T) {
		for _, stream := range []bool{true, false} {
			t.Run(fmt.Sprintf("stream=%t", stream), func(t *testing.T) {
				originalReq := &openai.ChatCompletionRequest{Model: "foo-bar-ai", Stream: stream}

				o := &openAIToAzureOpenAITranslatorV1ChatCompletion{apiVersion: "some-version"}
				hm, bm, err := o.RequestBody(nil, originalReq, false)
				require.Nil(t, bm)
				require.NoError(t, err)
				require.Equal(t, stream, o.stream)
				require.NotNil(t, hm)

				require.Equal(t, ":path", hm.SetHeaders[0].Header.Key)
				require.Equal(t, "/openai/deployments/foo-bar-ai/chat/completions?api-version=some-version", string(hm.SetHeaders[0].Header.RawValue))
			})
		}
	})
	t.Run("model override", func(t *testing.T) {
		modelName := "gpt-4-turbo-2024-04-09"
		originalReq := &openai.ChatCompletionRequest{Model: "gpt-4-turbo", Stream: false}
		o := &openAIToAzureOpenAITranslatorV1ChatCompletion{
			apiVersion: "some-version",
			openAIToOpenAITranslatorV1ChatCompletion: openAIToOpenAITranslatorV1ChatCompletion{
				modelNameOverride: modelName,
			},
		}
		hm, bm, err := o.RequestBody(nil, originalReq, false)
		require.Nil(t, bm)
		require.NoError(t, err)
		require.NotNil(t, hm)
		require.Len(t, hm.SetHeaders, 1)
		require.Equal(t, ":path", hm.SetHeaders[0].Header.Key)
		require.Equal(t, "/openai/deployments/"+modelName+"/chat/completions?api-version=some-version", string(hm.SetHeaders[0].Header.RawValue))
	})
}

// TestResponseModel_AzureOpenAI tests that Azure OpenAI returns the actual model version from response
// Azure ignores the model field in the request and uses URI-based resolution
func TestResponseModel_AzureOpenAI(t *testing.T) {
	translator := NewChatCompletionOpenAIToAzureOpenAITranslator("2024-08-01-preview", "")

	// Azure OpenAI response includes the actual model version
	var resp openai.ChatCompletionResponse
	resp.Model = "gpt-4o-2024-11-20" // Azure returns the actual versioned model name
	resp.Usage.TotalTokens = 15
	resp.Usage.PromptTokens = 10
	resp.Usage.CompletionTokens = 5

	body, err := json.Marshal(resp)
	require.NoError(t, err)

	// Test that Azure returns the actual model from response (URI-based resolution)
	_, _, tokenUsage, responseModel, err := translator.ResponseBody(nil, bytes.NewBuffer(body), true, nil)
	require.NoError(t, err)
	require.Equal(t, "gpt-4o-2024-11-20", responseModel) // Uses response field as authoritative
	require.Equal(t, uint32(10), tokenUsage.InputTokens)
	require.Equal(t, uint32(5), tokenUsage.OutputTokens)
}

// TestResponseModel_AzureOpenAIStreaming tests Azure OpenAI streaming returns actual model version
// Azure uses URI-based resolution but still returns model in streaming responses
func TestResponseModel_AzureOpenAIStreaming(t *testing.T) {
	translator := NewChatCompletionOpenAIToAzureOpenAITranslator("2024-08-01-preview", "").(*openAIToAzureOpenAITranslatorV1ChatCompletion)

	// Initialize as streaming
	req := &openai.ChatCompletionRequest{
		Model:  "gpt-4o", // This is ignored by Azure
		Stream: true,
	}
	_, _, err := translator.RequestBody(nil, req, false)
	require.NoError(t, err)
	require.True(t, translator.stream)

	// Azure streaming response with model field
	sseChunks := `data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1234567890,"model":"gpt-4o-2024-11-20","choices":[{"index":0,"delta":{"role":"assistant","content":"Hello"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1234567890,"model":"gpt-4o-2024-11-20","choices":[{"index":0,"delta":{"content":" world"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1234567890,"model":"gpt-4o-2024-11-20","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}}

data: [DONE]

`

	_, _, tokenUsage, responseModel, err := translator.ResponseBody(nil, bytes.NewReader([]byte(sseChunks)), true, nil)
	require.NoError(t, err)
	require.Equal(t, "gpt-4o-2024-11-20", responseModel) // Returns actual versioned model from response
	require.Equal(t, uint32(10), tokenUsage.InputTokens)
	require.Equal(t, uint32(5), tokenUsage.OutputTokens)
}
