// Copyright Envoy AI Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package extproc

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	openaigo "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/ai-gateway/internal/apischema/openai"
	"github.com/envoyproxy/ai-gateway/internal/filterapi"
	"github.com/envoyproxy/ai-gateway/tests/internal/testupstreamlib"
)

// failIf5xx because 5xx errors are likely a sign of a broken ExtProc or Envoy.
func failIf5xx(t *testing.T, resp *http.Response, was5xx *bool) {
	if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		*was5xx = true
		t.Fatalf("received %d response with body: %s", resp.StatusCode, string(body))
	}
}

// TestWithTestUpstream tests the end-to-end flow of the external processor with Envoy and the test upstream.
//
// This does not require any environment variables to be set as it relies on the test upstream.
func TestWithTestUpstream(t *testing.T) {
	now := time.Unix(int64(time.Now().Second()), 0).UTC()

	config := &filterapi.Config{
		MetadataNamespace: "ai_gateway_llm_ns",
		LLMRequestCosts: []filterapi.LLMRequestCost{
			{MetadataKey: "used_token", Type: filterapi.LLMRequestCostTypeInputToken},
		},
		ModelNameHeaderKey: "x-model-name",
		Backends: []filterapi.Backend{
			alwaysFailingBackend,
			testUpstreamOpenAIBackend,
			testUpstreamModelNameOverride,
			testUpstreamAAWSBackend,
			testUpstreamAzureBackend,
			testUpstreamGCPVertexAIBackend,
			testUpstreamGCPAnthropicAIBackend,
		},
		Models: []filterapi.Model{
			{Name: "some-model1", OwnedBy: "Envoy AI Gateway", CreatedAt: now},
			{Name: "some-model2", OwnedBy: "Envoy AI Gateway", CreatedAt: now},
			{Name: "some-model3", OwnedBy: "Envoy AI Gateway", CreatedAt: now},
		},
	}

	configBytes, err := yaml.Marshal(config)
	require.NoError(t, err)
	env := startTestEnvironment(t, string(configBytes), true, false)

	listenerPort := env.EnvoyListenerPort()

	expectedModels := openai.ModelList{
		Object: "list",
		Data: []openai.Model{
			{ID: "some-model1", Object: "model", OwnedBy: "Envoy AI Gateway", Created: openai.JSONUNIXTime(now)},
			{ID: "some-model2", Object: "model", OwnedBy: "Envoy AI Gateway", Created: openai.JSONUNIXTime(now)},
			{ID: "some-model3", Object: "model", OwnedBy: "Envoy AI Gateway", Created: openai.JSONUNIXTime(now)},
		},
	}

	was5xx := false
	for _, tc := range []struct {
		// name is the name of the test case.
		name,
		// backend is the backend to send the request to. Either "openai" or "aws-bedrock" (matching the headers in the config).
		backend,
		// path is the path to send the request to.
		path,
		// method is the HTTP method to use.
		method,
		// requestBody is the request requestBody.
		requestBody,
		// responseBody is the response body to return from the test upstream.
		responseBody,
		// responseType is either empty, "sse" or "aws-event-stream" as implemented by the test upstream.
		responseType,
		// responseStatus is the HTTP status code of the response returned by the test upstream.
		responseStatus,
		// responseHeaders are the headers sent in the HTTP response
		// The value is a base64 encoded string of comma separated key-value pairs.
		// E.g. "key1:value1,key2:value2".
		responseHeaders,
		// expRawQuery is the expected raw query to be sent to the test upstream.
		expRawQuery string
		// expPath is the expected path to be sent to the test upstream.
		expPath string
		// expHost is the expected host to be sent to the test upstream.
		expHost string
		// expHeaders are the expected headers to be sent to the test upstream.
		// The value is a base64 encoded string of comma separated key-value pairs.
		// E.g. "key1:value1,key2:value2".
		expHeaders map[string]string
		// expRequestBody is the expected body to be sent to the test upstream.
		// This can be used to test the request body translation.
		expRequestBody string
		// expStatus is the expected status code from the gateway.
		expStatus int
		// expResponseBody is the expected body from the gateway to the client.
		expResponseBody string
		// expResponseBodyFunc is a function to check the response body. This can be used instead of the expResponseBody field.
		expResponseBodyFunc func(require.TestingT, []byte)
	}{
		{
			name:            "unknown path",
			path:            "/unknown",
			requestBody:     `{"prompt": "hello"}`,
			expStatus:       http.StatusNotFound,
			expResponseBody: `unsupported path: /unknown`,
		},
		{
			name:            "aws system role - /v1/chat/completions",
			backend:         "aws-bedrock",
			path:            "/v1/chat/completions",
			requestBody:     `{"model":"something","messages":[{"role":"system","content":"You are a chatbot."}]}`,
			expPath:         "/model/something/converse",
			responseBody:    `{"output":{"message":{"content":[{"text":"response"},{"text":"from"},{"text":"assistant"}],"role":"assistant"}},"stopReason":null,"usage":{"inputTokens":10,"outputTokens":20,"totalTokens":30}}`,
			expRequestBody:  `{"inferenceConfig":{},"messages":[],"system":[{"text":"You are a chatbot."}]}`,
			expStatus:       http.StatusOK,
			expResponseBody: `{"choices":[{"finish_reason":"stop","index":0,"message":{"content":"response","role":"assistant"}}],"object":"chat.completion","usage":{"completion_tokens":20,"prompt_tokens":10,"total_tokens":30}}`,
		},
		{
			name:            "openai - /v1/chat/completions",
			backend:         "openai",
			path:            "/v1/chat/completions",
			method:          http.MethodPost,
			requestBody:     `{"model":"something","messages":[{"role":"system","content":"You are a chatbot."}]}`,
			expPath:         "/v1/chat/completions",
			responseBody:    `{"choices":[{"message":{"content":"This is a test."}}]}`,
			expStatus:       http.StatusOK,
			expResponseBody: `{"choices":[{"message":{"content":"This is a test."}}]}`,
		},
		{
			name:            "openai - /v1/chat/completions - gzip",
			backend:         "openai",
			responseType:    "gzip",
			path:            "/v1/chat/completions",
			method:          http.MethodPost,
			requestBody:     `{"model":"something","messages":[{"role":"system","content":"You are a chatbot."}]}`,
			expPath:         "/v1/chat/completions",
			responseBody:    `{"choices":[{"message":{"content":"This is a test."}}]}`,
			expStatus:       http.StatusOK,
			expResponseBody: `{"choices":[{"message":{"content":"This is a test."}}]}`,
		},
		{
			name:            "openai - /v1/chat/completions - tool call results",
			backend:         "openai",
			path:            "/v1/chat/completions",
			expPath:         "/v1/chat/completions",
			method:          http.MethodPost,
			requestBody:     toolCallResultsRequestBody,
			expRequestBody:  toolCallResultsRequestBody,
			responseBody:    `{"choices":[{"message":{"content":"This is a test."}}]}`,
			expResponseBody: `{"choices":[{"message":{"content":"This is a test."}}]}`,
			expStatus:       http.StatusOK,
		},
		{
			name:           "aws-bedrock - /v1/chat/completions - tool call results",
			backend:        "aws-bedrock",
			path:           "/v1/chat/completions",
			expPath:        "/model/gpt-4-0613/converse",
			method:         http.MethodPost,
			requestBody:    toolCallResultsRequestBody,
			expRequestBody: `{"inferenceConfig":{"maxTokens":1024},"messages":[{"content":[{"text":"List the files in the /tmp directory"}],"role":"user"},{"content":[{"toolUse":{"name":"list_files","input":{"path":"/tmp"},"toolUseId":"call_abc123"}}],"role":"assistant"},{"content":[{"toolResult":{"content":[{"text":"[\"foo.txt\", \"bar.log\", \"data.csv\"]"}],"status":null,"toolUseId":"call_abc123"}}],"role":"user"}]}`,
			responseBody:   `{"output":{"message":{"content":[{"text":"response"},{"text":"from"},{"text":"assistant"}],"role":"assistant"}},"stopReason":null,"usage":{"inputTokens":10,"outputTokens":20,"totalTokens":30}}`,
			expStatus:      http.StatusOK,
		},
		{
			name:            "gcp-anthropic - /v1/chat/completions - tool call results",
			backend:         "gcp-anthropicai",
			path:            "/v1/chat/completions",
			expPath:         "/v1/projects/gcp-project-name/locations/gcp-region/publishers/anthropic/models/gpt-4-0613:rawPredict",
			method:          http.MethodPost,
			requestBody:     toolCallResultsRequestBody,
			expRequestBody:  `{"max_tokens":1024,"messages":[{"content":[{"text":"List the files in the /tmp directory","type":"text"}],"role":"user"},{"content":[{"id":"call_abc123","input":{"path":"/tmp"},"name":"list_files","type":"tool_use"}],"role":"assistant"},{"content":[{"tool_use_id":"call_abc123","is_error":false,"content":[{"text":"[\"foo.txt\", \"bar.log\", \"data.csv\"]","type":"text"}],"type":"tool_result"}],"role":"user"}],"anthropic_version":"vertex-2023-10-16"}`,
			responseBody:    `{"id":"msg_123","type":"message","role":"assistant","stop_reason": "end_turn", "content":[{"type":"text","text":"Hello from Anthropic!"}],"usage":{"input_tokens":10,"output_tokens":25}}`,
			expResponseBody: `{"choices":[{"finish_reason":"stop","index":0,"message":{"content":"Hello from Anthropic!","role":"assistant"}}],"object":"chat.completion","usage":{"completion_tokens":25,"prompt_tokens":10,"total_tokens":35}}`,
			expStatus:       http.StatusOK,
		},
		{
			name:            "azure-openai - /v1/chat/completions",
			backend:         "azure-openai",
			path:            "/v1/chat/completions",
			method:          http.MethodPost,
			requestBody:     `{"model":"something","messages":[{"role":"system","content":"You are a chatbot."}]}`,
			expPath:         "/openai/deployments/something/chat/completions",
			responseBody:    `{"model":"gpt-4o-2024-08-01","choices":[{"message":{"content":"This is a test."}}]}`,
			expStatus:       http.StatusOK,
			expResponseBody: `{"model":"gpt-4o-2024-08-01","choices":[{"message":{"content":"This is a test."}}]}`,
		},
		{
			name:            "gcp-vertexai - /v1/chat/completions",
			backend:         "gcp-vertexai",
			path:            "/v1/chat/completions",
			method:          http.MethodPost,
			requestBody:     `{"model":"gemini-1.5-pro","messages":[{"role":"system","content":"You are a helpful assistant."}]}`,
			expRequestBody:  `{"contents":null,"tools":null,"generation_config":{},"system_instruction":{"parts":[{"text":"You are a helpful assistant."}]}}`,
			expHost:         "gcp-region-aiplatform.googleapis.com",
			expPath:         "/v1/projects/gcp-project-name/locations/gcp-region/publishers/google/models/gemini-1.5-pro:generateContent",
			expHeaders:      map[string]string{"Authorization": "Bearer " + fakeGCPAuthToken},
			responseStatus:  strconv.Itoa(http.StatusOK),
			responseBody:    `{"candidates":[{"content":{"parts":[{"text":"This is a test response from Gemini."}],"role":"model"},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":15,"candidatesTokenCount":10,"totalTokenCount":25}}`,
			expStatus:       http.StatusOK,
			expResponseBody: `{"choices":[{"finish_reason":"stop","index":0,"message":{"content":"This is a test response from Gemini.","role":"assistant"}}],"object":"chat.completion","usage":{"completion_tokens":10,"prompt_tokens":15,"total_tokens":25}}`,
		},
		{
			name:            "gcp-vertexai - /v1/chat/completions - tool use",
			backend:         "gcp-vertexai",
			path:            "/v1/chat/completions",
			method:          http.MethodPost,
			requestBody:     `{"model":"gemini-1.5-pro","messages":[{"role":"user","content":"tell me the delivery date for order 123"}],"tools":[{"type":"function","function":{"name":"get_delivery_date","description":"Get the delivery date for a customer's order. Call this whenever you need to know the delivery date, for example when a customer asks 'Where is my package'","parameters":{"type":"object","properties":{"order_id":{"type":"string","description":"The customer's order ID."}},"required":["order_id"]}}}]}`,
			expRequestBody:  `{"contents":[{"parts":[{"text":"tell me the delivery date for order 123"}],"role":"user"}],"tools":[{"functionDeclarations":[{"description":"Get the delivery date for a customer's order. Call this whenever you need to know the delivery date, for example when a customer asks 'Where is my package'","name":"get_delivery_date","parametersJsonSchema":{"properties":{"order_id":{"description":"The customer's order ID.","type":"string"}},"required":["order_id"],"type":"object"}}]}],"generation_config":{}}`,
			expHost:         "gcp-region-aiplatform.googleapis.com",
			expPath:         "/v1/projects/gcp-project-name/locations/gcp-region/publishers/google/models/gemini-1.5-pro:generateContent",
			expHeaders:      map[string]string{"Authorization": "Bearer " + fakeGCPAuthToken},
			responseStatus:  strconv.Itoa(http.StatusOK),
			responseBody:    `{"candidates":[{"content":{"role":"model","parts":[{"functionCall":{"name":"get_delivery_date","args":{"order_id":"123"}}}]},"finishReason":"STOP","avgLogprobs":0.000001220789272338152}],"usageMetadata":{"promptTokenCount":50,"candidatesTokenCount":11,"totalTokenCount":61,"trafficType":"ON_DEMAND","promptTokensDetails":[{"modality":"TEXT","tokenCount":50}],"candidatesTokensDetails":[{"modality":"TEXT","tokenCount":11}]},"modelVersion":"gemini-2.0-flash-001","createTime":"2025-07-11T22:15:44.956335Z","responseId":"EI5xaK-vOtqJm22IPmuCR14AI"}`,
			expStatus:       http.StatusOK,
			expResponseBody: `{"choices":[{"finish_reason":"stop","index":0,"message":{"role":"assistant","tool_calls":[{"id":"703482f8-2e5b-4dcc-a872-d74bd66c3866","function":{"arguments":"{\"order_id\":\"123\"}","name":"get_delivery_date"},"type":"function"}]}}],"object":"chat.completion","usage":{"completion_tokens":11,"prompt_tokens":50,"total_tokens":61}}`,
		},
		{
			name:            "gcp-anthropicai - /v1/chat/completions",
			backend:         "gcp-anthropicai",
			path:            "/v1/chat/completions",
			method:          http.MethodPost,
			requestBody:     `{"model":"claude-3-sonnet","max_completion_tokens":1024, "messages":[{"role":"system","content":"You are an Anthropic assistant."},{"role":"user","content":"Hello!"}]}`,
			expRequestBody:  `{"max_tokens":1024,"messages":[{"content":[{"text":"Hello!","type":"text"}],"role":"user"}],"system":[{"text":"You are an Anthropic assistant.","type":"text"}],"anthropic_version":"vertex-2023-10-16"}`,
			expHost:         "gcp-region-aiplatform.googleapis.com",
			expPath:         "/v1/projects/gcp-project-name/locations/gcp-region/publishers/anthropic/models/claude-3-sonnet:rawPredict",
			expHeaders:      map[string]string{"Authorization": "Bearer " + fakeGCPAuthToken},
			responseStatus:  strconv.Itoa(http.StatusOK),
			responseBody:    `{"id":"msg_123","type":"message","role":"assistant","stop_reason": "end_turn", "content":[{"type":"text","text":"Hello from Anthropic!"}],"usage":{"input_tokens":10,"output_tokens":25}}`,
			expStatus:       http.StatusOK,
			expResponseBody: `{"choices":[{"finish_reason":"stop","index":0,"message":{"content":"Hello from Anthropic!","role":"assistant"}}],"object":"chat.completion","usage":{"completion_tokens":25,"prompt_tokens":10,"total_tokens":35}}`,
		},
		{
			name:            "modelname-override - /v1/chat/completions",
			backend:         "modelname-override",
			path:            "/v1/chat/completions",
			method:          http.MethodPost,
			requestBody:     `{"model":"requested-model","messages":[{"role":"system","content":"You are a chatbot."}]}`,
			expRequestBody:  `{"model":"override-model","messages":[{"role":"system","content":"You are a chatbot."}]}`,
			expPath:         "/v1/chat/completions",
			responseBody:    `{"choices":[{"message":{"content":"This is a test."}}]}`,
			expStatus:       http.StatusOK,
			expResponseBody: `{"choices":[{"message":{"content":"This is a test."}}]}`,
		},
		{
			name:           "aws - /v1/chat/completions - streaming",
			backend:        "aws-bedrock",
			path:           "/v1/chat/completions",
			responseType:   "aws-event-stream",
			method:         http.MethodPost,
			requestBody:    `{"model":"something","messages":[{"role":"system","content":"You are a chatbot."}], "stream": true}`,
			expRequestBody: `{"inferenceConfig":{},"messages":[],"system":[{"text":"You are a chatbot."}]}`,
			expPath:        "/model/something/converse-stream",
			responseBody: `{"role":"assistant"}
{"start":{"toolUse":{"name":"cosine","toolUseId":"tooluse_QklrEHKjRu6Oc4BQUfy7ZQ"}}}
{"delta":{"text":"Don"}}
{"delta":{"text":"'t worry,  I'm here to help. It"}}
{"delta":{"text":" seems like you're testing my ability to respond appropriately"}}
{"stopReason":"tool_use"}
{"usage":{"inputTokens":41, "outputTokens":36, "totalTokens":77}}
`,
			expStatus: http.StatusOK,
			expResponseBody: `data: {"choices":[{"index":0,"delta":{"content":"","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"id":"tooluse_QklrEHKjRu6Oc4BQUfy7ZQ","function":{"arguments":"","name":"cosine"},"type":"function"}]}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":"Don","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":"'t worry,  I'm here to help. It","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":" seems like you're testing my ability to respond appropriately","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":"","role":"assistant"},"finish_reason":"tool_calls"}],"object":"chat.completion.chunk"}

data: {"object":"chat.completion.chunk","usage":{"completion_tokens":36,"prompt_tokens":41,"total_tokens":77}}

data: [DONE]
`,
		},
		{
			name:         "aws-bedrock - /v1/chat/completions - streaming with thinking config",
			backend:      "aws-bedrock",
			path:         "/v1/chat/completions",
			responseType: "aws-event-stream",
			method:       http.MethodPost,
			requestBody: `{
		"model":"something",
		"messages":[{"role":"system","content":"You are a chatbot."}],
		"stream": true,
		"thinking": {"type": "enabled", "budget_tokens": 4096}
	}`,
			expRequestBody: `{"additionalModelRequestFields":{"thinking":{"budget_tokens":4096,"type":"enabled"}},"inferenceConfig":{},"messages":[],"system":[{"text":"You are a chatbot."}]}`,
			expPath:        "/model/something/converse-stream",
			responseBody: `{"role":"assistant"}
	{"delta":{"reasoningContent":{"text":"First, I'll start by acknowledging the user..."}}}
	{"delta":{"text":"Hello!"}}
	{"stopReason":"end_turn"}`,
			expStatus: http.StatusOK,
			expResponseBody: `data: {"choices":[{"index":0,"delta":{"content":"","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":{"text":"First, I'll start by acknowledging the user..."}}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":"Hello!","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":"","role":"assistant"},"finish_reason":"stop"}],"object":"chat.completion.chunk"}

data: [DONE]
`,
		},
		{
			name:         "openai - /v1/chat/completions - streaming",
			backend:      "openai",
			path:         "/v1/chat/completions",
			responseType: "sse",
			method:       http.MethodPost,
			requestBody:  `{"model":"something","messages":[{"role":"system","content":"You are a chatbot."}], "stream": true}`,
			expPath:      "/v1/chat/completions",
			responseBody: `
{"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[{"index":0,"delta":{"role":"assistant","content":"","refusal":null},"logprobs":null,"finish_reason":null}],"usage":null}
{"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[],"usage":{"prompt_tokens":13,"completion_tokens":12,"total_tokens":25,"prompt_tokens_details":{"cached_tokens":0,"audio_tokens":0},"completion_tokens_details":{"reasoning_tokens":0,"audio_tokens":0,"accepted_prediction_tokens":0,"rejected_prediction_tokens":0}}}
[DONE]
`,
			expStatus: http.StatusOK,
			expResponseBody: `data: {"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[{"index":0,"delta":{"role":"assistant","content":"","refusal":null},"logprobs":null,"finish_reason":null}],"usage":null}

data: {"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[],"usage":{"prompt_tokens":13,"completion_tokens":12,"total_tokens":25,"prompt_tokens_details":{"cached_tokens":0,"audio_tokens":0},"completion_tokens_details":{"reasoning_tokens":0,"audio_tokens":0,"accepted_prediction_tokens":0,"rejected_prediction_tokens":0}}}

data: [DONE]

`,
		},
		{
			name:           "openai - /v1/chat/completions - streaming - forced to include usage",
			backend:        "openai",
			path:           "/v1/chat/completions",
			responseType:   "sse",
			method:         http.MethodPost,
			requestBody:    `{"model":"something","messages":[{"role":"system","content":"You are a chatbot."}], "stream": true, "stream_options": {"include_usage": false}}`,
			expRequestBody: `{"model":"something","messages":[{"role":"system","content":"You are a chatbot."}], "stream": true, "stream_options": {"include_usage": true}}`,
			expPath:        "/v1/chat/completions",
			responseBody: `
{"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[{"index":0,"delta":{"role":"assistant","content":"","refusal":null},"logprobs":null,"finish_reason":null}],"usage":null}
{"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[],"usage":{"prompt_tokens":13,"completion_tokens":12,"total_tokens":25,"prompt_tokens_details":{"cached_tokens":0,"audio_tokens":0},"completion_tokens_details":{"reasoning_tokens":0,"audio_tokens":0,"accepted_prediction_tokens":0,"rejected_prediction_tokens":0}}}
[DONE]
`,
			expStatus: http.StatusOK,
			expResponseBody: `data: {"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[{"index":0,"delta":{"role":"assistant","content":"","refusal":null},"logprobs":null,"finish_reason":null}],"usage":null}

data: {"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[],"usage":{"prompt_tokens":13,"completion_tokens":12,"total_tokens":25,"prompt_tokens_details":{"cached_tokens":0,"audio_tokens":0},"completion_tokens_details":{"reasoning_tokens":0,"audio_tokens":0,"accepted_prediction_tokens":0,"rejected_prediction_tokens":0}}}

data: [DONE]

`,
		},
		{
			name:           "openai - /v1/chat/completions - streaming - forced to include usage without steam_options",
			backend:        "openai",
			path:           "/v1/chat/completions",
			responseType:   "sse",
			method:         http.MethodPost,
			requestBody:    `{"model":"something","messages":[{"role":"system","content":"You are a chatbot."}], "stream": true}`,
			expRequestBody: `{"model":"something","messages":[{"role":"system","content":"You are a chatbot."}], "stream": true,"stream_options":{"include_usage":true}}`,
			expPath:        "/v1/chat/completions",
			responseBody: `
{"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[{"index":0,"delta":{"role":"assistant","content":"","refusal":null},"logprobs":null,"finish_reason":null}],"usage":null}
{"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[],"usage":{"prompt_tokens":13,"completion_tokens":12,"total_tokens":25,"prompt_tokens_details":{"cached_tokens":0,"audio_tokens":0},"completion_tokens_details":{"reasoning_tokens":0,"audio_tokens":0,"accepted_prediction_tokens":0,"rejected_prediction_tokens":0}}}
[DONE]
`,
			expStatus: http.StatusOK,
			expResponseBody: `data: {"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[{"index":0,"delta":{"role":"assistant","content":"","refusal":null},"logprobs":null,"finish_reason":null}],"usage":null}

data: {"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[],"usage":{"prompt_tokens":13,"completion_tokens":12,"total_tokens":25,"prompt_tokens_details":{"cached_tokens":0,"audio_tokens":0},"completion_tokens_details":{"reasoning_tokens":0,"audio_tokens":0,"accepted_prediction_tokens":0,"rejected_prediction_tokens":0}}}

data: [DONE]

`,
		},
		{
			name:           "openai - /v1/chat/completions - streaming - forced to include usage with model override",
			backend:        "modelname-override",
			path:           "/v1/chat/completions",
			responseType:   "sse",
			method:         http.MethodPost,
			requestBody:    `{"model":"requested-model","messages":[{"role":"system","content":"You are a chatbot."}], "stream": true, "stream_options": {"include_usage": false}}`,
			expRequestBody: `{"model":"override-model","messages":[{"role":"system","content":"You are a chatbot."}], "stream": true, "stream_options": {"include_usage": true}}`,
			expPath:        "/v1/chat/completions",
			responseBody: `
{"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[{"index":0,"delta":{"role":"assistant","content":"","refusal":null},"logprobs":null,"finish_reason":null}],"usage":null}
{"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[],"usage":{"prompt_tokens":13,"completion_tokens":12,"total_tokens":25,"prompt_tokens_details":{"cached_tokens":0,"audio_tokens":0},"completion_tokens_details":{"reasoning_tokens":0,"audio_tokens":0,"accepted_prediction_tokens":0,"rejected_prediction_tokens":0}}}
[DONE]
`,
			expStatus: http.StatusOK,
			expResponseBody: `data: {"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[{"index":0,"delta":{"role":"assistant","content":"","refusal":null},"logprobs":null,"finish_reason":null}],"usage":null}

data: {"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[],"usage":{"prompt_tokens":13,"completion_tokens":12,"total_tokens":25,"prompt_tokens_details":{"cached_tokens":0,"audio_tokens":0},"completion_tokens_details":{"reasoning_tokens":0,"audio_tokens":0,"accepted_prediction_tokens":0,"rejected_prediction_tokens":0}}}

data: [DONE]

`,
		},
		{
			name:           "openai - /v1/chat/completions - streaming - forced to include usage without steam_options with model override",
			backend:        "modelname-override",
			path:           "/v1/chat/completions",
			responseType:   "sse",
			method:         http.MethodPost,
			requestBody:    `{"model":"requested-model","messages":[{"role":"system","content":"You are a chatbot."}], "stream": true}`,
			expRequestBody: `{"model":"override-model","messages":[{"role":"system","content":"You are a chatbot."}], "stream": true,"stream_options":{"include_usage":true}}`,
			expPath:        "/v1/chat/completions",
			responseBody: `
{"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[{"index":0,"delta":{"role":"assistant","content":"","refusal":null},"logprobs":null,"finish_reason":null}],"usage":null}
{"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[],"usage":{"prompt_tokens":13,"completion_tokens":12,"total_tokens":25,"prompt_tokens_details":{"cached_tokens":0,"audio_tokens":0},"completion_tokens_details":{"reasoning_tokens":0,"audio_tokens":0,"accepted_prediction_tokens":0,"rejected_prediction_tokens":0}}}
[DONE]
`,
			expStatus: http.StatusOK,
			expResponseBody: `data: {"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[{"index":0,"delta":{"role":"assistant","content":"","refusal":null},"logprobs":null,"finish_reason":null}],"usage":null}

data: {"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[],"usage":{"prompt_tokens":13,"completion_tokens":12,"total_tokens":25,"prompt_tokens_details":{"cached_tokens":0,"audio_tokens":0},"completion_tokens_details":{"reasoning_tokens":0,"audio_tokens":0,"accepted_prediction_tokens":0,"rejected_prediction_tokens":0}}}

data: [DONE]

`,
		},
		{
			name:           "gcp-vertexai - /v1/chat/completions - streaming",
			backend:        "gcp-vertexai",
			path:           "/v1/chat/completions",
			responseType:   "sse",
			method:         http.MethodPost,
			requestBody:    `{"model":"gemini-1.5-pro","messages":[{"role":"system","content":"You are a helpful assistant."}], "stream": true}`,
			expRequestBody: `{"contents":null,"tools":null,"generation_config":{},"system_instruction":{"parts":[{"text":"You are a helpful assistant."}]}}`,
			expHost:        "gcp-region-aiplatform.googleapis.com",
			expPath:        "/v1/projects/gcp-project-name/locations/gcp-region/publishers/google/models/gemini-1.5-pro:streamGenerateContent",
			expRawQuery:    "alt=sse",
			expHeaders:     map[string]string{"Authorization": "Bearer " + fakeGCPAuthToken},
			responseStatus: strconv.Itoa(http.StatusOK),
			responseBody: `{"candidates":[{"content":{"parts":[{"text":"Hello"}],"role":"model"}}]}
{"candidates":[{"content":{"parts":[{"text":"! How"}],"role":"model"}}]}
{"candidates":[{"content":{"parts":[{"text":" can I"}],"role":"model"}}]}
{"candidates":[{"content":{"parts":[{"text":" help"}],"role":"model"}}]}
{"candidates":[{"content":{"parts":[{"text":" you"}],"role":"model"}}]}
{"candidates":[{"content":{"parts":[{"text":" today"}],"role":"model"}}]}
{"candidates":[{"content":{"parts":[{"text":"?"}],"role":"model"},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":10,"candidatesTokenCount":7,"totalTokenCount":17}}`,
			expStatus: http.StatusOK,
			expResponseBody: `data: {"choices":[{"index":0,"delta":{"content":"Hello","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":"! How","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":" can I","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":" help","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":" you","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":" today","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":"?","role":"assistant"},"finish_reason":"stop"}],"object":"chat.completion.chunk","usage":{"completion_tokens":7,"prompt_tokens":10,"total_tokens":17}}

data: [DONE]
`,
		},
		{
			name:           "gcp-anthropicai - /v1/chat/completions - streaming",
			backend:        "gcp-anthropicai",
			path:           "/v1/chat/completions",
			method:         http.MethodPost,
			responseType:   "sse",
			requestBody:    `{"model":"claude-3-sonnet","max_completion_tokens":1024, "messages":[{"role":"user","content":"Why is the sky blue?"}], "stream": true}`,
			expRequestBody: `{"max_tokens":1024,"messages":[{"content":[{"text":"Why is the sky blue?","type":"text"}],"role":"user"}],"anthropic_version":"vertex-2023-10-16"}`,
			expHost:        "gcp-region-aiplatform.googleapis.com",
			expPath:        "/v1/projects/gcp-project-name/locations/gcp-region/publishers/anthropic/models/claude-3-sonnet:streamRawPredict",
			expHeaders:     map[string]string{"Authorization": "Bearer " + fakeGCPAuthToken},
			responseStatus: strconv.Itoa(http.StatusOK),
			responseBody: `event: message_start
data: {"type": "message_start", "message": {"id": "msg_123", "usage": {"input_tokens": 15}}}

event: content_block_start
data: {"type": "content_block_start", "index": 0, "content_block": {"type": "text", "text": ""}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "The sky appears blue"}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text":" due to Rayleigh scattering."}}

event: content_block_stop
data: {"type": "content_block_stop", "index": 0}

event: message_delta
data: {"type": "message_delta", "delta": {"stop_reason": "end_turn"}, "usage": {"output_tokens": 12}}

event: message_stop
data: {"type": "message_stop"}
`,
			expStatus: http.StatusOK,
			expResponseBody: `data: {"choices":[{"index":0,"delta":{"content":"The sky appears blue","role":"assistant"}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"content":" due to Rayleigh scattering."}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"object":"chat.completion.chunk"}

data: {"object":"chat.completion.chunk","usage":{"completion_tokens":12,"prompt_tokens":15,"total_tokens":27}}

data: [DONE]

`,
		},
		{
			name:         "gcp-anthropicai - /v1/chat/completions - streaming tool use",
			backend:      "gcp-anthropicai",
			path:         "/v1/chat/completions",
			method:       http.MethodPost,
			responseType: "sse",
			requestBody: `{
		"model": "claude-3-sonnet",
		"max_tokens": 1024,
		"messages": [{"role": "user", "content": "What is the weather in Boston?"}],
		"stream": true,
		"tools": [{
			"type": "function",
			"function": {
				"name": "get_weather",
				"description": "Get the current weather in a given location",
				"parameters": {
					"type": "object",
					"properties": {
						"location": {"type": "string", "description": "The city and state, e.g. San Francisco, CA"}
					},
					"required": ["location"]
				}
			}
		}]
	}`,
			expHost:        "gcp-region-aiplatform.googleapis.com",
			expPath:        "/v1/projects/gcp-project-name/locations/gcp-region/publishers/anthropic/models/claude-3-sonnet:streamRawPredict",
			expHeaders:     map[string]string{"Authorization": "Bearer " + fakeGCPAuthToken},
			responseStatus: strconv.Itoa(http.StatusOK),
			responseBody: `event: message_start
data: {"type": "message_start", "message": {"id": "msg_123", "usage": {"input_tokens": 50}}}

event: content_block_start
data: {"type": "content_block_start", "index": 0, "content_block": {"type": "tool_use", "id": "toolu_abc123", "name": "get_weather", "input": {}}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "input_json_delta", "partial_json": "{\"location\":\"Bosto"}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "input_json_delta", "partial_json": "n, MA\"}"}}

event: content_block_stop
data: {"type": "content_block_stop", "index": 0}

event: message_delta
data: {"type": "message_delta", "delta": {"stop_reason": "tool_use"}, "usage": {"output_tokens": 20}}

event: message_stop
data: {"type": "message_stop"}`,
			expStatus: http.StatusOK,
			expResponseBody: `data: {"choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"index":0,"id":"toolu_abc123","function":{"arguments":"{}","name":"get_weather"},"type":"function"}]}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":null,"function":{"arguments":"{\"location\":\"Bosto","name":""}}]}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":null,"function":{"arguments":"n, MA\"}","name":""}}]}}],"object":"chat.completion.chunk"}

data: {"choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}],"object":"chat.completion.chunk"}

data: {"object":"chat.completion.chunk","usage":{"completion_tokens":20,"prompt_tokens":50,"total_tokens":70}}

data: [DONE]

`,
		},
		{
			name:            "openai - /v1/chat/completions - error response",
			backend:         "openai",
			path:            "/v1/chat/completions",
			responseType:    "",
			method:          http.MethodPost,
			requestBody:     `{"model":"something","messages":[{"role":"system","content":"You are a chatbot."}], "stream": true}`,
			expPath:         "/v1/chat/completions",
			responseStatus:  "400",
			expStatus:       http.StatusBadRequest,
			responseBody:    `{"error": {"message": "missing required field", "type": "BadRequestError", "code": "400"}}`,
			expResponseBody: `{"error": {"message": "missing required field", "type": "BadRequestError", "code": "400"}}`,
		},
		{
			name:            "aws-bedrock - /v1/chat/completions - error response",
			backend:         "aws-bedrock",
			path:            "/v1/chat/completions",
			responseType:    "",
			method:          http.MethodPost,
			requestBody:     `{"model":"something","messages":[{"role":"system","content":"You are a chatbot."}], "stream": true}`,
			expPath:         "/model/something/converse-stream",
			responseStatus:  "429",
			expStatus:       http.StatusTooManyRequests,
			responseHeaders: "x-amzn-errortype:ThrottledException",
			responseBody:    `{"message": "aws bedrock rate limit exceeded"}`,
			expResponseBody: `{"type":"error","error":{"type":"ThrottledException","code":"429","message":"aws bedrock rate limit exceeded"}}`,
		},
		{
			name:            "gcp-vertexai - /v1/chat/completions - error response",
			backend:         "gcp-vertexai",
			path:            "/v1/chat/completions",
			responseType:    "",
			method:          http.MethodPost,
			requestBody:     `{"model":"gemini-1.5-pro","messages":[{"role":"system","content":"You are a helpful assistant."}]}`,
			expPath:         "/v1/projects/gcp-project-name/locations/gcp-region/publishers/google/models/gemini-1.5-pro:generateContent",
			responseStatus:  "400",
			expStatus:       http.StatusBadRequest,
			responseBody:    `{"error":{"code":400,"message":"Invalid request: missing required field","status":"INVALID_ARGUMENT"}}`,
			expResponseBody: `{"type":"error","error":{"type":"INVALID_ARGUMENT","code":"400","message":"Invalid request: missing required field"}}`,
		},
		{
			name:            "gcp-anthropicai - /v1/chat/completions - error response",
			backend:         "gcp-anthropicai",
			path:            "/v1/chat/completions",
			responseType:    "",
			method:          http.MethodPost,
			requestBody:     `{"model":"claude-3-sonnet", "max_completion_tokens":1024, "messages":[{"role":"system","content":"You are a helpful assistant."}]}`,
			expPath:         "/v1/projects/gcp-project-name/locations/gcp-region/publishers/anthropic/models/claude-3-sonnet:rawPredict",
			responseStatus:  "400",
			expStatus:       http.StatusBadRequest,
			responseBody:    `{"error":{"type":"invalid_request_error","code":400,"message":"Invalid request: missing required field","status":"INVALID_ARGUMENT"}}`,
			expResponseBody: `{"type":"error","error":{"type":"invalid_request_error","code":"400","message":"Invalid request: missing required field"}}`,
		},
		{
			name:            "openai - /v1/embeddings",
			backend:         "openai",
			path:            "/v1/embeddings",
			method:          http.MethodPost,
			requestBody:     `{"model":"text-embedding-ada-002","input":"The food was delicious and the waiter..."}`,
			expPath:         "/v1/embeddings",
			responseBody:    `{"object":"list","data":[{"object":"embedding","embedding":[0.0023064255,-0.009327292,-0.0028842222],"index":0}],"model":"text-embedding-ada-002","usage":{"prompt_tokens":8,"total_tokens":8}}`,
			expStatus:       http.StatusOK,
			expResponseBody: `{"object":"list","data":[{"object":"embedding","embedding":[0.0023064255,-0.009327292,-0.0028842222],"index":0}],"model":"text-embedding-ada-002","usage":{"prompt_tokens":8,"total_tokens":8}}`,
		},
		{
			name:            "openai - /v1/embeddings - gzip",
			backend:         "openai",
			responseType:    "gzip",
			path:            "/v1/embeddings",
			method:          http.MethodPost,
			requestBody:     `{"model":"text-embedding-ada-002","input":"The food was delicious and the waiter..."}`,
			expPath:         "/v1/embeddings",
			responseBody:    `{"object":"list","data":[{"object":"embedding","embedding":[0.0023064255,-0.009327292,-0.0028842222],"index":0}],"model":"text-embedding-ada-002","usage":{"prompt_tokens":8,"total_tokens":8}}`,
			expStatus:       http.StatusOK,
			expResponseBody: `{"object":"list","data":[{"object":"embedding","embedding":[0.0023064255,-0.009327292,-0.0028842222],"index":0}],"model":"text-embedding-ada-002","usage":{"prompt_tokens":8,"total_tokens":8}}`,
		},
		{
			name:            "openai - /v1/embeddings - error response",
			backend:         "openai",
			path:            "/v1/embeddings",
			responseType:    "",
			method:          http.MethodPost,
			requestBody:     `{"model":"text-embedding-ada-002","input":""}`,
			expPath:         "/v1/embeddings",
			responseStatus:  "400",
			expStatus:       http.StatusBadRequest,
			responseBody:    `{"error": {"message": "input cannot be empty", "type": "BadRequestError", "code": "400"}}`,
			expResponseBody: `{"error": {"message": "input cannot be empty", "type": "BadRequestError", "code": "400"}}`,
		},
		{
			name:                "openai - /v1/models",
			backend:             "openai",
			path:                "/v1/models",
			method:              http.MethodGet,
			expStatus:           http.StatusOK,
			expResponseBodyFunc: checkModels(expectedModels),
		},
		{
			name:    "openai - /v1/chat/completions - assistant text content",
			backend: "openai",
			path:    "/v1/chat/completions",
			method:  http.MethodPost,
			requestBody: `
{
       "model": "whatever",
       "messages": [
               {"role": "user", "content": [{"type": "text", "text": "hi sir"}]},
               {"role": "assistant","content": [{"type": "text", "text": "Hello! How can I assist you today?"}]},
               {"role": "user", "content": [{"type": "text", "text": "what are you?"}]}
       ]
}`,
			expPath:         "/v1/chat/completions",
			responseBody:    `{"choices":[{"message":{"content":"This is a test."}}]}`,
			expStatus:       http.StatusOK,
			expResponseBody: `{"choices":[{"message":{"content":"This is a test."}}]}`,
		},
		{
			name:            "gcp-anthropicai - /anthropic/v1/messages",
			backend:         "gcp-anthropicai",
			path:            "/anthropic/v1/messages",
			method:          http.MethodPost,
			requestBody:     `{"model":"claude-3-sonnet","max_tokens":100,"messages":[{"role":"user","content":[{"type":"text","text":"Hello, just a simple test message."}]}],"stream":false}`,
			expRequestBody:  `{"anthropic_version":"vertex-2023-10-16","max_tokens":100,"messages":[{"content":[{"text":"Hello, just a simple test message.","type":"text"}],"role":"user"}],"stream":false}`,
			expHost:         "gcp-region-aiplatform.googleapis.com",
			expPath:         "/v1/projects/gcp-project-name/locations/gcp-region/publishers/anthropic/models/claude-3-sonnet:rawPredict",
			expHeaders:      map[string]string{"Authorization": "Bearer " + fakeGCPAuthToken},
			responseStatus:  strconv.Itoa(http.StatusOK),
			responseBody:    `{"id":"msg_123","type":"message","role":"assistant","stop_reason": "end_turn", "content":[{"type":"text","text":"Hello from native Anthropic API!"}],"usage":{"input_tokens":8,"output_tokens":15}}`,
			expStatus:       http.StatusOK,
			expResponseBody: `{"id":"msg_123","type":"message","role":"assistant","stop_reason": "end_turn", "content":[{"type":"text","text":"Hello from native Anthropic API!"}],"usage":{"input_tokens":8,"output_tokens":15}}`,
		},
		{
			name:           "gcp-anthropicai - /anthropic/v1/messages - streaming",
			backend:        "gcp-anthropicai",
			path:           "/anthropic/v1/messages",
			method:         http.MethodPost,
			responseType:   "sse",
			requestBody:    `{"model":"claude-3-sonnet","max_tokens":100,"messages":[{"role":"user","content":[{"type":"text","text":"Tell me a short joke"}]}],"stream":true}`,
			expRequestBody: `{"anthropic_version":"vertex-2023-10-16","max_tokens":100,"messages":[{"content":[{"text":"Tell me a short joke","type":"text"}],"role":"user"}],"stream":true}`,
			expHost:        "gcp-region-aiplatform.googleapis.com",
			expPath:        "/v1/projects/gcp-project-name/locations/gcp-region/publishers/anthropic/models/claude-3-sonnet:streamRawPredict",
			expHeaders:     map[string]string{"Authorization": "Bearer " + fakeGCPAuthToken},
			responseStatus: strconv.Itoa(http.StatusOK),
			responseBody: `event: message_start
data: {"type": "message_start", "message": {"id": "msg_789", "usage": {"input_tokens": 8}}}

event: content_block_start
data: {"type": "content_block_start", "index": 0, "content_block": {"type": "text", "text": ""}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "Why don't scientists trust atoms? Because they make up everything!"}}

event: content_block_stop
data: {"type": "content_block_stop", "index": 0}

event: message_delta
data: {"type": "message_delta", "delta": {"stop_reason": "end_turn"}, "usage": {"output_tokens": 15}}

event: message_stop
data: {"type": "message_stop"}

`,
			expStatus: http.StatusOK,
			expResponseBody: `event: message_start
data: {"type": "message_start", "message": {"id": "msg_789", "usage": {"input_tokens": 8}}}

event: content_block_start
data: {"type": "content_block_start", "index": 0, "content_block": {"type": "text", "text": ""}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "Why don't scientists trust atoms? Because they make up everything!"}}

event: content_block_stop
data: {"type": "content_block_stop", "index": 0}

event: message_delta
data: {"type": "message_delta", "delta": {"stop_reason": "end_turn"}, "usage": {"output_tokens": 15}}

event: message_stop
data: {"type": "message_stop"}

`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Eventually(t, func() bool {
				listenerAddress := fmt.Sprintf("http://localhost:%d", listenerPort)
				req, err := http.NewRequestWithContext(t.Context(), tc.method, listenerAddress+tc.path, strings.NewReader(tc.requestBody))
				require.NoError(t, err)
				req.Header.Set("x-test-backend", tc.backend)
				req.Header.Set(testupstreamlib.ResponseBodyHeaderKey, base64.StdEncoding.EncodeToString([]byte(tc.responseBody)))
				req.Header.Set(testupstreamlib.ExpectedPathHeaderKey, base64.StdEncoding.EncodeToString([]byte(tc.expPath)))
				req.Header.Set(testupstreamlib.ResponseStatusKey, tc.responseStatus)

				var expHeaders []string
				for k, v := range tc.expHeaders {
					expHeaders = append(expHeaders, fmt.Sprintf("%s:%s", k, v))
				}
				if len(expHeaders) > 0 {
					req.Header.Set(
						testupstreamlib.ExpectedHeadersKey,
						base64.StdEncoding.EncodeToString(
							[]byte(strings.Join(expHeaders, ","))),
					)
				}

				if tc.expRawQuery != "" {
					req.Header.Set(testupstreamlib.ExpectedRawQueryHeaderKey, tc.expRawQuery)
				}
				if tc.expHost != "" {
					req.Header.Set(testupstreamlib.ExpectedHostKey, tc.expHost)
				}
				if tc.responseType != "" {
					req.Header.Set(testupstreamlib.ResponseTypeKey, tc.responseType)
				}
				if tc.responseHeaders != "" {
					req.Header.Set(testupstreamlib.ResponseHeadersKey, base64.StdEncoding.EncodeToString([]byte(tc.responseHeaders)))
				}
				if tc.expRequestBody != "" {
					req.Header.Set(testupstreamlib.ExpectedRequestBodyHeaderKey, base64.StdEncoding.EncodeToString([]byte(tc.expRequestBody)))
				}

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					t.Logf("error: %v", err)
					return false
				}
				defer func() { _ = resp.Body.Close() }()

				failIf5xx(t, resp, &was5xx)

				if resp.StatusCode != tc.expStatus {
					t.Logf("unexpected status code: %d", resp.StatusCode)
					return false
				}

				if tc.expResponseBody != "" {
					bodyBytes, err := io.ReadAll(resp.Body)
					if err != nil {
						t.Logf("error reading response body: %v", err)
						return false
					}

					// Substitute any dynamically generated UUIDs in the response body with a placeholder
					// example generated UUID 703482f8-2e5b-4dcc-a872-d74bd66c386.
					m := regexp.MustCompile("[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}")
					body := m.ReplaceAllString(string(bodyBytes), "<UUID4-replaced>")
					expectedResponseBody := m.ReplaceAllString(tc.expResponseBody, "<UUID4-replaced>")
					if body != expectedResponseBody {
						t.Logf("unexpected response (-want +got):\n%s", cmp.Diff(tc.expResponseBody, body))
						return false
					}
				} else if tc.expResponseBodyFunc != nil {
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						t.Logf("error reading response body: %v", err)
						return false
					}
					tc.expResponseBodyFunc(t, body)
				}
				return true
			}, eventuallyTimeout, eventuallyInterval)
		})
	}

	t.Run("stream non blocking", func(t *testing.T) {
		if was5xx {
			return // rather than also failing subsequent tests, which confuses root cause.
		}

		// This receives a stream of 20 event messages. The testuptream server sleeps 200 ms between each message.
		// Therefore, if envoy fails to process the response in a streaming manner, the test will fail taking more than 4 seconds.
		listenerAddress := fmt.Sprintf("http://localhost:%d", listenerPort)
		client := openaigo.NewClient(
			option.WithBaseURL(listenerAddress+"/v1/"),
			option.WithHeader("x-test-backend", "openai"),
			option.WithHeader(testupstreamlib.ResponseTypeKey, "sse"),
			option.WithHeader(testupstreamlib.ResponseBodyHeaderKey,
				base64.StdEncoding.EncodeToString([]byte(
					`
{"id":"chatcmpl-B8ZKlXBoEXZVTtv3YBmewxuCpNW7b","object":"chat.completion.chunk","created":1741382147,"model":"gpt-4o-mini-2024-07-18","service_tier":"default","system_fingerprint":"fp_06737a9306","choices":[{"index":0,"delta":{"content":" This"},"logprobs":null,"finish_reason":null}],"usage":null}
{"id":"chatcmpl-B8ZKlXBoEXZVTtv3YBmewxuCpNW7b","object":"chat.completion.chunk","created":1741382147,"model":"gpt-4o-mini-2024-07-18","service_tier":"default","system_fingerprint":"fp_06737a9306","choices":[{"index":0,"delta":{"content":" is"},"logprobs":null,"finish_reason":null}],"usage":null}
{"id":"chatcmpl-B8ZKlXBoEXZVTtv3YBmewxuCpNW7b","object":"chat.completion.chunk","created":1741382147,"model":"gpt-4o-mini-2024-07-18","service_tier":"default","system_fingerprint":"fp_06737a9306","choices":[{"index":0,"delta":{"content":" a"},"logprobs":null,"finish_reason":null}],"usage":null}
{"id":"chatcmpl-B8ZKlXBoEXZVTtv3YBmewxuCpNW7b","object":"chat.completion.chunk","created":1741382147,"model":"gpt-4o-mini-2024-07-18","service_tier":"default","system_fingerprint":"fp_06737a9306","choices":[{"index":0,"delta":{"content":" test"},"logprobs":null,"finish_reason":null}],"usage":null}
{"id":"chatcmpl-B8ZKlXBoEXZVTtv3YBmewxuCpNW7b","object":"chat.completion.chunk","created":1741382147,"model":"gpt-4o-mini-2024-07-18","service_tier":"default","system_fingerprint":"fp_06737a9306","choices":[{"index":0,"delta":{"content":"."},"logprobs":null,"finish_reason":null}],"usage":null}
{"id":"chatcmpl-B8ZKlXBoEXZVTtv3YBmewxuCpNW7b","object":"chat.completion.chunk","created":1741382147,"model":"gpt-4o-mini-2024-07-18","service_tier":"default","system_fingerprint":"fp_06737a9306","choices":[{"index":0,"delta":{"content":" This"},"logprobs":null,"finish_reason":null}],"usage":null}
{"id":"chatcmpl-B8ZKlXBoEXZVTtv3YBmewxuCpNW7b","object":"chat.completion.chunk","created":1741382147,"model":"gpt-4o-mini-2024-07-18","service_tier":"default","system_fingerprint":"fp_06737a9306","choices":[{"index":0,"delta":{"content":" is"},"logprobs":null,"finish_reason":null}],"usage":null}
{"id":"chatcmpl-B8ZKlXBoEXZVTtv3YBmewxuCpNW7b","object":"chat.completion.chunk","created":1741382147,"model":"gpt-4o-mini-2024-07-18","service_tier":"default","system_fingerprint":"fp_06737a9306","choices":[{"index":0,"delta":{"content":" a"},"logprobs":null,"finish_reason":null}],"usage":null}
{"id":"chatcmpl-B8ZKlXBoEXZVTtv3YBmewxuCpNW7b","object":"chat.completion.chunk","created":1741382147,"model":"gpt-4o-mini-2024-07-18","service_tier":"default","system_fingerprint":"fp_06737a9306","choices":[{"index":0,"delta":{"content":" test"},"logprobs":null,"finish_reason":null}],"usage":null}
{"id":"chatcmpl-B8ZKlXBoEXZVTtv3YBmewxuCpNW7b","object":"chat.completion.chunk","created":1741382147,"model":"gpt-4o-mini-2024-07-18","service_tier":"default","system_fingerprint":"fp_06737a9306","choices":[{"index":0,"delta":{"content":"."},"logprobs":null,"finish_reason":null}],"usage":null}
{"id":"chatcmpl-B8ZKlXBoEXZVTtv3YBmewxuCpNW7b","object":"chat.completion.chunk","created":1741382147,"model":"gpt-4o-mini-2024-07-18","service_tier":"default","system_fingerprint":"fp_06737a9306","choices":[{"index":0,"delta":{"content":" This"},"logprobs":null,"finish_reason":null}],"usage":null}
{"id":"chatcmpl-B8ZKlXBoEXZVTtv3YBmewxuCpNW7b","object":"chat.completion.chunk","created":1741382147,"model":"gpt-4o-mini-2024-07-18","service_tier":"default","system_fingerprint":"fp_06737a9306","choices":[{"index":0,"delta":{"content":" is"},"logprobs":null,"finish_reason":null}],"usage":null}
{"id":"chatcmpl-B8ZKlXBoEXZVTtv3YBmewxuCpNW7b","object":"chat.completion.chunk","created":1741382147,"model":"gpt-4o-mini-2024-07-18","service_tier":"default","system_fingerprint":"fp_06737a9306","choices":[{"index":0,"delta":{"content":" a"},"logprobs":null,"finish_reason":null}],"usage":null}
{"id":"chatcmpl-B8ZKlXBoEXZVTtv3YBmewxuCpNW7b","object":"chat.completion.chunk","created":1741382147,"model":"gpt-4o-mini-2024-07-18","service_tier":"default","system_fingerprint":"fp_06737a9306","choices":[{"index":0,"delta":{"content":" test"},"logprobs":null,"finish_reason":null}],"usage":null}
{"id":"chatcmpl-B8ZKlXBoEXZVTtv3YBmewxuCpNW7b","object":"chat.completion.chunk","created":1741382147,"model":"gpt-4o-mini-2024-07-18","service_tier":"default","system_fingerprint":"fp_06737a9306","choices":[{"index":0,"delta":{"content":"."},"logprobs":null,"finish_reason":null}],"usage":null}
{"id":"chatcmpl-B8ZKlXBoEXZVTtv3YBmewxuCpNW7b","object":"chat.completion.chunk","created":1741382147,"model":"gpt-4o-mini-2024-07-18","service_tier":"default","system_fingerprint":"fp_06737a9306","choices":[{"index":0,"delta":{"content":" This"},"logprobs":null,"finish_reason":null}],"usage":null}
{"id":"chatcmpl-B8ZKlXBoEXZVTtv3YBmewxuCpNW7b","object":"chat.completion.chunk","created":1741382147,"model":"gpt-4o-mini-2024-07-18","service_tier":"default","system_fingerprint":"fp_06737a9306","choices":[{"index":0,"delta":{"content":" is"},"logprobs":null,"finish_reason":null}],"usage":null}
{"id":"chatcmpl-B8ZKlXBoEXZVTtv3YBmewxuCpNW7b","object":"chat.completion.chunk","created":1741382147,"model":"gpt-4o-mini-2024-07-18","service_tier":"default","system_fingerprint":"fp_06737a9306","choices":[{"index":0,"delta":{"content":" a"},"logprobs":null,"finish_reason":null}],"usage":null}
{"id":"chatcmpl-B8ZKlXBoEXZVTtv3YBmewxuCpNW7b","object":"chat.completion.chunk","created":1741382147,"model":"gpt-4o-mini-2024-07-18","service_tier":"default","system_fingerprint":"fp_06737a9306","choices":[{"index":0,"delta":{"content":" test"},"logprobs":null,"finish_reason":null}],"usage":null}
{"id":"chatcmpl-B8ZKlXBoEXZVTtv3YBmewxuCpNW7b","object":"chat.completion.chunk","created":1741382147,"model":"gpt-4o-mini-2024-07-18","service_tier":"default","system_fingerprint":"fp_06737a9306","choices":[{"index":0,"delta":{"content":"."},"logprobs":null,"finish_reason":null}],"usage":null}
{"id":"chatcmpl-B8ZKlXBoEXZVTtv3YBmewxuCpNW7b","object":"chat.completion.chunk","created":1741382147,"model":"gpt-4o-mini-2024-07-18","service_tier":"default","system_fingerprint":"fp_06737a9306","choices":[],"usage":{"prompt_tokens":25,"completion_tokens":61,"total_tokens":86,"prompt_tokens_details":{"cached_tokens":0,"audio_tokens":0},"completion_tokens_details":{"reasoning_tokens":0,"audio_tokens":0,"accepted_prediction_tokens":0,"rejected_prediction_tokens":0}}}
[DONE]
`,
				))),
		)

		// NewStreaming below will block until the first event is received, so take the time before calling it.
		start := time.Now()
		stream := client.Chat.Completions.NewStreaming(t.Context(), openaigo.ChatCompletionNewParams{
			Messages: []openaigo.ChatCompletionMessageParamUnion{
				openaigo.UserMessage("Say this is a test"),
			},
			Model: "something",
		})
		defer func() {
			_ = stream.Close()
		}()

		asserted := false
		for stream.Next() {
			chunk := stream.Current()
			if len(chunk.Choices) == 0 || chunk.Choices[0].Delta.Content == "" {
				continue
			}
			t.Logf("%v: %v", time.Now(), chunk.Choices[0].Delta.Content)
			// Check each event is received less than a second after the previous one.
			require.Less(t, time.Since(start), time.Second)
			start = time.Now()
			asserted = true
		}
		require.True(t, asserted)
		require.NoError(t, stream.Err())
	})
}

func checkModels(want openai.ModelList) func(t require.TestingT, body []byte) {
	return func(t require.TestingT, body []byte) {
		var models openai.ModelList
		require.NoError(t, json.Unmarshal(body, &models))
		require.Len(t, models.Data, len(want.Data))
		require.Equal(t, want, models)
	}
}

const (
	toolCallResultsRequestBody = `{
  "model": "gpt-4-0613",
  "max_completion_tokens":1024,
  "messages": [
    {"role": "user","content": "List the files in the /tmp directory"},
    {
      "role": "assistant",
      "tool_calls": [
        {
          "id": "call_abc123",
          "type": "function",
          "function": {
            "name": "list_files",
            "arguments": "{ \"path\": \"/tmp\" }"
          }
        }
      ]
    },
    {
      "role": "tool",
      "tool_call_id": "call_abc123",
      "content": "[\"foo.txt\", \"bar.log\", \"data.csv\"]"
    }
  ]
}`
)
