// Copyright Envoy AI Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package extproc

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric/noop"
	"google.golang.org/protobuf/types/known/structpb"

	anthropicschema "github.com/envoyproxy/ai-gateway/internal/apischema/anthropic"
	"github.com/envoyproxy/ai-gateway/internal/extproc/headermutator"
	"github.com/envoyproxy/ai-gateway/internal/extproc/translator"
	"github.com/envoyproxy/ai-gateway/internal/filterapi"
	"github.com/envoyproxy/ai-gateway/internal/internalapi"
	"github.com/envoyproxy/ai-gateway/internal/metrics"
	tracing "github.com/envoyproxy/ai-gateway/internal/tracing/api"
)

func TestMessagesProcessorFactory(t *testing.T) {
	chatMetrics := metrics.NewChatCompletion(noop.NewMeterProvider().Meter("test"), map[string]string{})
	factory := MessagesProcessorFactory(chatMetrics)
	require.NotNil(t, factory, "MessagesProcessorFactory should return a non-nil factory")

	// Test creating a router filter.
	config := &processorConfig{
		modelNameHeaderKey: "x-model",
	}
	headers := map[string]string{
		":path":         "/anthropic/v1/messages",
		"authorization": "Bearer token",
	}
	logger := slog.Default()

	routerProcessor, err := factory(config, headers, logger, tracing.NoopTracing{}, false)
	require.NoError(t, err, "Factory should create router processor without error")
	require.NotNil(t, routerProcessor, "Router processor should not be nil")
	require.IsType(t, &messagesProcessorRouterFilter{}, routerProcessor, "Should return router filter type")

	// Test creating an upstream filter.
	upstreamProcessor, err := factory(config, headers, logger, tracing.NoopTracing{}, true)
	require.NoError(t, err, "Factory should create upstream processor without error")
	require.NotNil(t, upstreamProcessor, "Upstream processor should not be nil")
	require.IsType(t, &messagesProcessorUpstreamFilter{}, upstreamProcessor, "Should return upstream filter type")
}

func TestParseAnthropicMessagesBody(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		expectError bool
		checkFields func(t *testing.T, req *anthropicschema.MessagesRequest)
	}{
		{
			name: "valid_anthropic_request",
			body: `{
				"model": "claude-3-sonnet",
				"max_tokens": 1000,
				"messages": [{"role": "user", "content": "Hello"}],
				"stream": false
			}`,
			expectError: false,
			checkFields: func(t *testing.T, req *anthropicschema.MessagesRequest) {
				require.Equal(t, "claude-3-sonnet", req.GetModel())
				require.Equal(t, 1000, req.GetMaxTokens())
				require.False(t, req.GetStream())
			},
		},
		{
			name: "streaming_request",
			body: `{
				"model": "claude-3-sonnet",
				"max_tokens": 1000,
				"messages": [{"role": "user", "content": "Hello"}],
				"stream": true
			}`,
			expectError: false,
			checkFields: func(t *testing.T, req *anthropicschema.MessagesRequest) {
				require.True(t, req.GetStream())
			},
		},
		{
			name:        "invalid_json",
			body:        `{"invalid": json}`,
			expectError: true,
		},
		{
			name:        "empty_body",
			body:        "",
			expectError: true,
		},
		{
			name: "request_with_tools",
			body: `{
				"model": "claude-3-sonnet",
				"max_tokens": 1000,
				"messages": [{"role": "user", "content": "Hello"}],
				"tools": [{"type": "function", "function": {"name": "test"}}]
			}`,
			expectError: false,
			checkFields: func(t *testing.T, req *anthropicschema.MessagesRequest) {
				tools, ok := (*req)["tools"]
				require.True(t, ok)
				require.NotNil(t, tools)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes := []byte(tt.body)
			body := &extprocv3.HttpBody{Body: bodyBytes}

			modelName, req, err := parseAnthropicMessagesBody(body)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, req)
			} else {
				require.NoError(t, err)
				require.NotNil(t, req)
				require.NotEmpty(t, modelName)
				if tt.checkFields != nil {
					tt.checkFields(t, req)
				}
			}
		})
	}
}

func TestMessagesProcessorRouterFilter_ProcessRequestHeaders(t *testing.T) {
	processor := &messagesProcessorRouterFilter{
		config: &processorConfig{},
		logger: slog.Default(),
	}

	ctx := context.Background()
	headers := &corev3.HeaderMap{}

	response, err := processor.ProcessRequestHeaders(ctx, headers)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, response.Response)
	require.NotNil(t, response.Response.(*extprocv3.ProcessingResponse_RequestHeaders))
}

func TestMessagesProcessorRouterFilter_ProcessRequestBody(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		expectError bool
		expectModel string
	}{
		{
			name: "valid anthropic request",
			body: `{
				"model": "claude-3-sonnet-20240229",
				"max_tokens": 1000,
				"messages": [{"role": "user", "content": "Hello"}]
			}`,
			expectError: false,
			expectModel: "claude-3-sonnet-20240229",
		},
		{
			name: "missing model field",
			body: `{
				"max_tokens": 1000,
				"messages": [{"role": "user", "content": "Hello"}]
			}`,
			expectError: true,
		},
		{
			name:        "invalid json",
			body:        `{invalid json`,
			expectError: true,
		},
		{
			name:        "empty body",
			body:        "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := &messagesProcessorRouterFilter{
				config: &processorConfig{
					modelNameHeaderKey: "x-model-name",
				},
				requestHeaders: make(map[string]string),
				logger:         slog.Default(),
			}

			ctx := context.Background()
			httpBody := &extprocv3.HttpBody{
				Body: []byte(tt.body),
			}

			response, err := processor.ProcessRequestBody(ctx, httpBody)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, response)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				require.Equal(t, tt.expectModel, processor.requestHeaders["x-model-name"])
				require.NotNil(t, processor.originalRequestBody)
				require.NotEmpty(t, processor.originalRequestBodyRaw)
			}
		})
	}
}

func TestMessagesProcessorRouterFilter_UnimplementedMethods(t *testing.T) {
	processor := &messagesProcessorRouterFilter{
		config: &processorConfig{},
		logger: slog.Default(),
	}

	ctx := context.Background()

	// Test ProcessResponseHeaders.
	respHeaders, err := processor.ProcessResponseHeaders(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, respHeaders)

	// Test ProcessResponseBody.
	respBody, err := processor.ProcessResponseBody(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, respBody)

	// Test SetBackend.
	err = processor.SetBackend(ctx, nil, nil, nil)
	require.NoError(t, err)
}

func TestMessagesProcessorUpstreamFilter_ProcessRequestBody_ShouldPanic(t *testing.T) {
	processor := &messagesProcessorUpstreamFilter{
		config: &processorConfig{},
		logger: slog.Default(),
	}

	ctx := context.Background()
	httpBody := &extprocv3.HttpBody{
		Body: []byte(`{"messages": []}`),
	}

	// This method should panic as upstream filters don't process request bodies.
	require.Panics(t, func() {
		_, _ = processor.ProcessRequestBody(ctx, httpBody)
	})
}

func TestSelectTranslator(t *testing.T) {
	processor := &messagesProcessorUpstreamFilter{
		config: &processorConfig{
			backends: map[string]*processorConfigBackend{
				"gcp": {
					b: &filterapi.Backend{
						Name: "gcp",
						Schema: filterapi.VersionedAPISchema{
							Name:    filterapi.APISchemaGCPAnthropic,
							Version: "vertex-2023-10-16",
						},
					},
				},
			},
		},
		logger: slog.Default(),
	}

	tests := []struct {
		name        string
		backend     string
		expectError bool
	}{
		{
			name:        "gcp backend",
			backend:     "gcp",
			expectError: false,
		},
		{
			name:        "unsupported backend",
			backend:     "unknown",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor.backendName = tt.backend
			backend := processor.config.backends[tt.backend]
			if backend != nil {
				err := processor.selectTranslator(backend.b.Schema)
				if tt.expectError {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					require.NotNil(t, processor.translator)
				}
			} else {
				// For unknown backend, we expect error.
				require.True(t, tt.expectError)
			}
		})
	}
}

// mockAnthropicTranslator implements [translator.AnthropicMessagesTranslator] for testing.
type mockAnthropicTranslator struct {
	t                           *testing.T
	expRequestBody              *anthropicschema.MessagesRequest
	expForceRequestBodyMutation bool
	retHeaderMutation           *extprocv3.HeaderMutation
	retBodyMutation             *extprocv3.BodyMutation
	retTokenUsage               translator.LLMTokenUsage
	retResponseModel            internalapi.ResponseModel
	retErr                      error
}

// RequestBody implements [translator.AnthropicMessagesTranslator].
func (m mockAnthropicTranslator) RequestBody(_ []byte, body *anthropicschema.MessagesRequest, forceRequestBodyMutation bool) (*extprocv3.HeaderMutation, *extprocv3.BodyMutation, error) {
	if m.expRequestBody != nil {
		require.Equal(m.t, m.expRequestBody, body)
	}
	require.Equal(m.t, m.expForceRequestBodyMutation, forceRequestBodyMutation)
	return m.retHeaderMutation, m.retBodyMutation, m.retErr
}

// ResponseHeaders implements [translator.AnthropicMessagesTranslator].
func (m mockAnthropicTranslator) ResponseHeaders(_ map[string]string) (*extprocv3.HeaderMutation, error) {
	return m.retHeaderMutation, m.retErr
}

// ResponseBody implements [translator.AnthropicMessagesTranslator].
func (m mockAnthropicTranslator) ResponseBody(_ map[string]string, _ io.Reader, _ bool) (*extprocv3.HeaderMutation, *extprocv3.BodyMutation, translator.LLMTokenUsage, string, error) {
	return m.retHeaderMutation, m.retBodyMutation, m.retTokenUsage, m.retResponseModel, m.retErr
}

func TestMessagesProcessorUpstreamFilter_ProcessRequestHeaders_WithMocks(t *testing.T) {
	tests := []struct {
		name               string
		translatorErr      error
		expectError        bool
		forcedIncludeUsage bool
	}{
		{
			name:               "successful processing",
			translatorErr:      nil,
			expectError:        false,
			forcedIncludeUsage: false,
		},
		{
			name:               "translator error",
			translatorErr:      errors.New("test translator error"),
			expectError:        true,
			forcedIncludeUsage: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := map[string]string{":path": "/anthropic/v1/messages", "x-model-name": "claude-3-sonnet"}

			// Create request body.
			requestBody := &anthropicschema.MessagesRequest{
				"model":      "claude-3-sonnet",
				"max_tokens": 1000,
				"messages":   []any{map[string]any{"role": "user", "content": "Hello"}},
			}
			requestBodyRaw := []byte(`{"model": "claude-3-sonnet", "max_tokens": 1000, "messages": [{"role": "user", "content": "Hello"}]}`)

			// Create mock translator.
			mockTranslator := mockAnthropicTranslator{
				t:                           t,
				expRequestBody:              requestBody,
				expForceRequestBodyMutation: tt.forcedIncludeUsage,
				retHeaderMutation:           &extprocv3.HeaderMutation{},
				retBodyMutation:             &extprocv3.BodyMutation{},
				retErr:                      tt.translatorErr,
			}

			// Create mock metrics.
			chatMetrics := metrics.NewChatCompletion(noop.NewMeterProvider().Meter("test"), map[string]string{})

			// Create processor.
			processor := &messagesProcessorUpstreamFilter{
				config: &processorConfig{
					modelNameHeaderKey: "x-model-name",
				},
				requestHeaders:         headers,
				logger:                 slog.Default(),
				metrics:                chatMetrics,
				translator:             mockTranslator,
				originalRequestBody:    requestBody,
				originalRequestBodyRaw: requestBodyRaw,
				onRetry:                tt.forcedIncludeUsage,
			}

			ctx := context.Background()
			response, err := processor.ProcessRequestHeaders(ctx, nil)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, response)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
			}
		})
	}
}

func TestMessagesProcessorUpstreamFilter_ProcessResponseHeaders_WithMocks(t *testing.T) {
	mockTranslator := mockAnthropicTranslator{
		t:                 t,
		retHeaderMutation: &extprocv3.HeaderMutation{},
		retErr:            nil,
	}

	chatMetrics := metrics.NewChatCompletion(noop.NewMeterProvider().Meter("test"), map[string]string{})
	processor := &messagesProcessorUpstreamFilter{
		config:         &processorConfig{},
		requestHeaders: make(map[string]string),
		logger:         slog.Default(),
		metrics:        chatMetrics,
		translator:     mockTranslator,
	}

	ctx := context.Background()
	response, err := processor.ProcessResponseHeaders(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, response)
}

func TestMessagesProcessorUpstreamFilter_ProcessResponseBody_WithMocks(t *testing.T) {
	// Create a simple test for the method that passes through.
	mockTranslator := mockAnthropicTranslator{
		t:                 t,
		retHeaderMutation: &extprocv3.HeaderMutation{},
		retBodyMutation:   &extprocv3.BodyMutation{},
		retResponseModel:  "test-model",
		retErr:            nil,
	}

	chatMetrics := metrics.NewChatCompletion(noop.NewMeterProvider().Meter("test"), map[string]string{})
	processor := &messagesProcessorUpstreamFilter{
		config:         &processorConfig{},
		requestHeaders: make(map[string]string),
		logger:         slog.Default(),
		metrics:        chatMetrics,
		translator:     mockTranslator,
	}

	ctx := context.Background()
	httpBody := &extprocv3.HttpBody{Body: []byte(`{"test": "response"}`)}
	response, err := processor.ProcessResponseBody(ctx, httpBody)
	require.NoError(t, err)
	require.NotNil(t, response)
}

func TestMessagesProcessorUpstreamFilter_ProcessResponseBody_ErrorRecordsFailure(t *testing.T) {
	// Translator returns error; ensure failure is recorded.
	mockTranslator := mockAnthropicTranslator{
		t:      t,
		retErr: errors.New("translate error"),
	}

	mm := &mockChatCompletionMetrics{}
	processor := &messagesProcessorUpstreamFilter{
		config:         &processorConfig{},
		requestHeaders: make(map[string]string),
		logger:         slog.Default(),
		metrics:        mm,
		translator:     mockTranslator,
	}

	ctx := context.Background()
	_, err := processor.ProcessResponseBody(ctx, &extprocv3.HttpBody{})
	require.Error(t, err)
	mm.RequireRequestFailure(t)
}

func TestMessagesProcessorUpstreamFilter_ProcessResponseBody_CompletionOnlyAtEnd(t *testing.T) {
	// Verify success is recorded only at EndOfStream by checking that no error occurs mid-stream
	// and the call completes successfully at end.
	mockTranslator := mockAnthropicTranslator{
		t:                 t,
		retHeaderMutation: &extprocv3.HeaderMutation{},
		retBodyMutation:   &extprocv3.BodyMutation{},
		retErr:            nil,
	}

	mm := &mockChatCompletionMetrics{}
	processor := &messagesProcessorUpstreamFilter{
		config:         &processorConfig{},
		requestHeaders: make(map[string]string),
		logger:         slog.Default(),
		metrics:        mm,
		translator:     mockTranslator,
		stream:         true,
	}

	ctx := context.Background()
	// Mid-stream.
	_, err := processor.ProcessResponseBody(ctx, &extprocv3.HttpBody{Body: []byte("chunk"), EndOfStream: false})
	require.NoError(t, err)
	mm.RequireRequestNotCompleted(t)

	// End-of-stream.
	_, err = processor.ProcessResponseBody(ctx, &extprocv3.HttpBody{Body: []byte("final"), EndOfStream: true})
	require.NoError(t, err)
	mm.RequireRequestSuccess(t)
}

func TestMessagesProcessorUpstreamFilter_MergeWithTokenLatencyMetadata(t *testing.T) {
	chatMetrics := metrics.NewChatCompletion(noop.NewMeterProvider().Meter("test"), map[string]string{})
	processor := &messagesProcessorUpstreamFilter{
		config: &processorConfig{
			metadataNamespace: "ai_gateway_llm_ns",
		},
		logger:  slog.Default(),
		metrics: chatMetrics,
		costs:   translator.LLMTokenUsage{InputTokens: 100, OutputTokens: 50},
	}

	// Test with valid metadata structure.
	metadata := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"ai_gateway_llm_ns": {
				Kind: &structpb.Value_StructValue{
					StructValue: &structpb.Struct{
						Fields: map[string]*structpb.Value{},
					},
				},
			},
		},
	}

	// This method doesn't return anything, just test it doesn't panic.
	require.NotPanics(t, func() {
		processor.mergeWithTokenLatencyMetadata(metadata)
	})
}

func TestMessagesProcessorUpstreamFilter_SetBackend(t *testing.T) {
	headers := map[string]string{":path": "/anthropic/v1/messages"}
	chatMetrics := metrics.NewChatCompletion(noop.NewMeterProvider().Meter("test"), map[string]string{})
	processor := &messagesProcessorUpstreamFilter{
		config: &processorConfig{
			requestCosts: []processorConfigRequestCost{
				{LLMRequestCost: &filterapi.LLMRequestCost{Type: filterapi.LLMRequestCostTypeOutputToken, MetadataKey: "output_token_usage", CEL: "15"}},
			},
		},
		requestHeaders: headers,
		logger:         slog.Default(),
		metrics:        chatMetrics,
	}

	// Test with unsupported schema (should error).
	err := processor.SetBackend(context.Background(), &filterapi.Backend{
		Name:              "some-backend",
		Schema:            filterapi.VersionedAPISchema{Name: "some-unsupported-schema", Version: "v10.0"},
		ModelNameOverride: "claude-override",
	}, nil, &messagesProcessorRouterFilter{
		config: &processorConfig{},
		logger: slog.Default(),
	})
	require.ErrorContains(t, err, "only supports backends that return native Anthropic format")
}

func Test_messagesProcessorUpstreamFilter_SetBackend_Success(t *testing.T) {
	const modelKey = "x-model-name"
	headers := map[string]string{":path": "/anthropic/v1/messages", modelKey: "claude"}
	chatMetrics := metrics.NewChatCompletion(noop.NewMeterProvider().Meter("test"), map[string]string{})
	p := &messagesProcessorUpstreamFilter{
		config:         &processorConfig{modelNameHeaderKey: modelKey},
		requestHeaders: headers,
		logger:         slog.Default(),
		metrics:        chatMetrics,
	}
	rp := &messagesProcessorRouterFilter{
		originalRequestBody: &anthropicschema.MessagesRequest{"model": "claude", "stream": true},
	}
	err := p.SetBackend(t.Context(), &filterapi.Backend{
		Name:              "gcp",
		Schema:            filterapi.VersionedAPISchema{Name: filterapi.APISchemaGCPAnthropic, Version: "vertex-2023-10-16"},
		ModelNameOverride: "claude-vertex",
	}, nil, rp)
	require.NoError(t, err)
	require.Equal(t, "claude-vertex", p.requestHeaders[modelKey])
	require.True(t, p.stream)
	require.NotNil(t, p.translator)
}

func TestMessages_ProcessRequestHeaders_SetsRequestModel(t *testing.T) {
	const modelKey = internalapi.ModelNameHeaderKeyDefault
	headers := map[string]string{":path": "/anthropic/v1/messages", modelKey: "header-model"}
	requestBody := &anthropicschema.MessagesRequest{"model": "body-model", "messages": []any{"hello"}}
	requestBodyRaw := []byte(`{"model":"body-model","messages":["hello"]}`)
	mm := &mockChatCompletionMetrics{}
	p := &messagesProcessorUpstreamFilter{
		config:                 &processorConfig{modelNameHeaderKey: modelKey},
		requestHeaders:         headers,
		logger:                 slog.Default(),
		metrics:                mm,
		translator:             mockAnthropicTranslator{t: t, expRequestBody: requestBody},
		originalRequestBodyRaw: requestBodyRaw,
		originalRequestBody:    requestBody,
	}
	_, _ = p.ProcessRequestHeaders(t.Context(), nil)
	// Should use the override model from the header, as that's what is sent upstream.
	require.Equal(t, "header-model", mm.requestModel)
	// Response model is not set until we get actual response
	require.Empty(t, mm.responseModel)
}

// TestMessages_ProcessResponseBody_UsesActualResponseModelOverHeaderOverride verifies that
// the actual response model from the API response is used for metrics, not the header override.
// This is important because the backend may return a more specific model version than what was
// requested (e.g., "claude-3-opus-20240229" instead of "claude-3-opus").
func TestMessages_ProcessResponseBody_UsesActualResponseModelOverHeaderOverride(t *testing.T) {
	const modelKey = internalapi.ModelNameHeaderKeyDefault
	headers := map[string]string{":path": "/v1/messages", modelKey: "header-model"}
	requestBody := &anthropicschema.MessagesRequest{"model": "body-model"}
	requestBodyRaw := []byte(`{"model": "body-model"}`)
	mm := &mockChatCompletionMetrics{}

	// Create a mock translator that returns token usage with response model
	mt := &mockAnthropicTranslator{
		t:              t,
		expRequestBody: requestBody,
		retTokenUsage: translator.LLMTokenUsage{
			InputTokens:  25,
			OutputTokens: 35,
		},
		retResponseModel: "actual-anthropic-model",
	}

	p := &messagesProcessorUpstreamFilter{
		config:                 &processorConfig{modelNameHeaderKey: modelKey},
		requestHeaders:         headers,
		logger:                 slog.Default(),
		metrics:                mm,
		translator:             mt,
		originalRequestBodyRaw: requestBodyRaw,
		originalRequestBody:    requestBody,
	}

	// First process request headers
	_, _ = p.ProcessRequestHeaders(t.Context(), nil)

	// Process response headers (required before body)
	responseHeaders := &corev3.HeaderMap{
		Headers: []*corev3.HeaderValue{
			{Key: ":status", Value: "200"},
		},
	}
	_, err := p.ProcessResponseHeaders(t.Context(), responseHeaders)
	require.NoError(t, err)

	// Now process response body (should set response model from response)
	responseBytes := []byte(`{"model": "actual-anthropic-model", "content": [{"type": "text", "text": "test"}], "usage": {"input_tokens": 25, "output_tokens": 35}}`)
	_, err = p.ProcessResponseBody(t.Context(), &extprocv3.HttpBody{
		Body:        responseBytes,
		EndOfStream: true,
	})
	require.NoError(t, err)

	// Should use the override model from the header, as that's what is sent upstream.
	mm.RequireSelectedModel(t, "header-model", "actual-anthropic-model")
	// For non-streaming, only usage is recorded, not latency
	require.Equal(t, 60, mm.tokenUsageCount)
	mm.RequireRequestSuccess(t)
}

func TestMessagesProcessorUpstreamFilter_ProcessRequestHeaders_WithHeaderMutations(t *testing.T) {
	t.Run("header mutations applied correctly", func(t *testing.T) {
		headers := map[string]string{
			":path":         "/anthropic/v1/messages",
			"x-model-name":  "claude-3-sonnet",
			"authorization": "bearer token123",
			"x-api-key":     "secret-key",
			"x-custom":      "custom-value",
		}

		// Create request body.
		requestBody := &anthropicschema.MessagesRequest{
			"model":      "claude-3-sonnet",
			"max_tokens": 1000,
			"messages":   []any{map[string]any{"role": "user", "content": "Hello"}},
		}
		requestBodyRaw := []byte(`{"model": "claude-3-sonnet", "max_tokens": 1000, "messages": [{"role": "user", "content": "Hello"}]}`)

		// Create header mutations.
		headerMutations := &filterapi.HTTPHeaderMutation{
			Remove: []string{"authorization", "x-api-key"},
			Set:    []filterapi.HTTPHeader{{Name: "x-new-header", Value: "new-value"}},
		}

		// Create mock translator.
		mockTranslator := mockAnthropicTranslator{
			t:                           t,
			expRequestBody:              requestBody,
			expForceRequestBodyMutation: false,
			retHeaderMutation:           &extprocv3.HeaderMutation{},
			retBodyMutation:             &extprocv3.BodyMutation{},
			retErr:                      nil,
		}

		// Create mock metrics.
		chatMetrics := metrics.NewChatCompletion(noop.NewMeterProvider().Meter("test"), map[string]string{})

		// Create processor.
		processor := &messagesProcessorUpstreamFilter{
			config: &processorConfig{
				modelNameHeaderKey: "x-model-name",
			},
			requestHeaders:         headers,
			logger:                 slog.Default(),
			metrics:                chatMetrics,
			translator:             mockTranslator,
			originalRequestBody:    requestBody,
			originalRequestBodyRaw: requestBodyRaw,
			handler:                &mockBackendAuthHandler{},
		}

		// Set header mutator.
		originalHeaders := map[string]string{
			"authorization": "bearer original-token",
			"x-api-key":     "original-secret",
		}
		processor.headerMutator = headermutator.NewHeaderMutator(headerMutations, originalHeaders)

		ctx := context.Background()
		response, err := processor.ProcessRequestHeaders(ctx, nil)

		require.NoError(t, err)
		require.NotNil(t, response)

		commonRes := response.Response.(*extprocv3.ProcessingResponse_RequestHeaders).RequestHeaders.Response

		// Check that header mutations were applied.
		require.NotNil(t, commonRes.HeaderMutation)
		require.ElementsMatch(t, []string{"authorization", "x-api-key"}, commonRes.HeaderMutation.RemoveHeaders)
		require.Len(t, commonRes.HeaderMutation.SetHeaders, 1)
		require.Equal(t, "x-new-header", commonRes.HeaderMutation.SetHeaders[0].Header.Key)
		require.Equal(t, []byte("new-value"), commonRes.HeaderMutation.SetHeaders[0].Header.RawValue)

		// Check that headers were modified in the request headers.
		require.Equal(t, "new-value", headers["x-new-header"])
		require.NotContains(t, headers, "authorization")
		require.NotContains(t, headers, "x-api-key")
		// x-custom remains unchanged since it wasn't in the mutations.
		require.Equal(t, "custom-value", headers["x-custom"])
	})

	t.Run("header mutations restored on retry", func(t *testing.T) {
		headers := map[string]string{
			":path":        "/anthropic/v1/messages",
			"x-model-name": "claude-3-sonnet",
			// "x-custom" is not present in current headers, so it can be restored.
			"x-new-header": "new-value", // Already set from previous mutation.
		}

		// Create request body.
		requestBody := &anthropicschema.MessagesRequest{
			"model":      "claude-3-sonnet",
			"max_tokens": 1000,
			"messages":   []any{map[string]any{"role": "user", "content": "Hello"}},
		}
		requestBodyRaw := []byte(`{"model": "claude-3-sonnet", "max_tokens": 1000, "messages": [{"role": "user", "content": "Hello"}]}`)

		// Create header mutations that don't remove x-custom (so it can be restored).
		headerMutations := &filterapi.HTTPHeaderMutation{
			Remove: []string{"authorization", "x-api-key"},
			Set:    []filterapi.HTTPHeader{{Name: "x-new-header", Value: "updated-value"}},
		}

		// Create mock translator.
		mockTranslator := mockAnthropicTranslator{
			t:                           t,
			expRequestBody:              requestBody,
			expForceRequestBodyMutation: true, // This is a retry request.
			retHeaderMutation:           &extprocv3.HeaderMutation{},
			retBodyMutation:             &extprocv3.BodyMutation{},
			retErr:                      nil,
		}

		// Create mock metrics.
		chatMetrics := metrics.NewChatCompletion(noop.NewMeterProvider().Meter("test"), map[string]string{})

		// Create processor.
		processor := &messagesProcessorUpstreamFilter{
			config: &processorConfig{
				modelNameHeaderKey: "x-model-name",
			},
			requestHeaders:         headers,
			logger:                 slog.Default(),
			metrics:                chatMetrics,
			translator:             mockTranslator,
			originalRequestBody:    requestBody,
			originalRequestBodyRaw: requestBodyRaw,
			handler:                &mockBackendAuthHandler{},
			onRetry:                true, // This is a retry request.
		}

		// Use the same headers map as the original headers (this simulates the router filter's requestHeaders).
		originalHeaders := map[string]string{
			":path":         "/anthropic/v1/messages",
			"x-model-name":  "claude-3-sonnet",
			"authorization": "bearer original-token", // This will be removed, so won't be restored.
			"x-api-key":     "original-secret",       // This will be removed, so won't be restored.
			"x-custom":      "original-custom",       // This won't be removed, so can be restored.
			"x-new-header":  "original-value",        // This will be set, so won't be restored.
		}
		processor.headerMutator = headermutator.NewHeaderMutator(headerMutations, originalHeaders)

		ctx := context.Background()
		response, err := processor.ProcessRequestHeaders(ctx, nil)

		require.NoError(t, err)
		require.NotNil(t, response)

		commonRes := response.Response.(*extprocv3.ProcessingResponse_RequestHeaders).RequestHeaders.Response

		// Check that header mutations were applied.
		require.NotNil(t, commonRes.HeaderMutation)
		// RemoveHeaders should be empty because authorization/x-api-key don't exist in current headers.
		require.Empty(t, commonRes.HeaderMutation.RemoveHeaders)
		require.Len(t, commonRes.HeaderMutation.SetHeaders, 2) // Updated header + restored header.

		// Check that x-custom header was restored on retry (it's not being removed or set).
		var restoredHeader *corev3.HeaderValueOption
		var updatedHeader *corev3.HeaderValueOption
		for _, h := range commonRes.HeaderMutation.SetHeaders {
			switch h.Header.Key {
			case "x-custom":
				restoredHeader = h
			case "x-new-header":
				updatedHeader = h
			}
		}
		require.NotNil(t, restoredHeader)
		require.Equal(t, []byte("original-custom"), restoredHeader.Header.RawValue)
		require.NotNil(t, updatedHeader)
		require.Equal(t, []byte("updated-value"), updatedHeader.Header.RawValue)

		// Check that headers were updated in the request headers.
		require.Equal(t, "updated-value", headers["x-new-header"])
		require.Equal(t, "original-custom", headers["x-custom"])
	})

	t.Run("no header mutations when mutator is nil", func(t *testing.T) {
		headers := map[string]string{
			":path":         "/anthropic/v1/messages",
			"x-model-name":  "claude-3-sonnet",
			"authorization": "bearer token123",
		}

		// Create request body.
		requestBody := &anthropicschema.MessagesRequest{
			"model":      "claude-3-sonnet",
			"max_tokens": 1000,
			"messages":   []any{map[string]any{"role": "user", "content": "Hello"}},
		}
		requestBodyRaw := []byte(`{"model": "claude-3-sonnet", "max_tokens": 1000, "messages": [{"role": "user", "content": "Hello"}]}`)

		// Create mock translator.
		mockTranslator := mockAnthropicTranslator{
			t:                           t,
			expRequestBody:              requestBody,
			expForceRequestBodyMutation: false,
			retHeaderMutation:           &extprocv3.HeaderMutation{},
			retBodyMutation:             &extprocv3.BodyMutation{},
			retErr:                      nil,
		}

		// Create mock metrics.
		chatMetrics := metrics.NewChatCompletion(noop.NewMeterProvider().Meter("test"), map[string]string{})

		// Create processor.
		processor := &messagesProcessorUpstreamFilter{
			config: &processorConfig{
				modelNameHeaderKey: "x-model-name",
			},
			requestHeaders:         headers,
			logger:                 slog.Default(),
			metrics:                chatMetrics,
			translator:             mockTranslator,
			originalRequestBody:    requestBody,
			originalRequestBodyRaw: requestBodyRaw,
			handler:                &mockBackendAuthHandler{},
			headerMutator:          nil, // No header mutator.
		}

		ctx := context.Background()
		response, err := processor.ProcessRequestHeaders(ctx, nil)

		require.NoError(t, err)
		require.NotNil(t, response)

		commonRes := response.Response.(*extprocv3.ProcessingResponse_RequestHeaders).RequestHeaders.Response

		// Check that no header mutations were applied.
		require.NotNil(t, commonRes.HeaderMutation)
		require.Empty(t, commonRes.HeaderMutation.RemoveHeaders)
		require.Empty(t, commonRes.HeaderMutation.SetHeaders)

		// Check that original headers remain unchanged.
		require.Equal(t, "bearer token123", headers["authorization"])
	})
}

func TestMessagesProcessorUpstreamFilter_SetBackend_WithHeaderMutations(t *testing.T) {
	t.Run("header mutator created correctly", func(t *testing.T) {
		headers := map[string]string{":path": "/anthropic/v1/messages"}
		chatMetrics := metrics.NewChatCompletion(noop.NewMeterProvider().Meter("test"), map[string]string{})
		p := &messagesProcessorUpstreamFilter{
			config:         &processorConfig{},
			requestHeaders: headers,
			logger:         slog.Default(),
			metrics:        chatMetrics,
		}

		// Create backend with header mutations.
		headerMutations := &filterapi.HTTPHeaderMutation{
			Remove: []string{"x-sensitive"},
			Set:    []filterapi.HTTPHeader{{Name: "x-backend", Value: "backend-value"}},
		}

		// Original headers from router filter.
		originalHeaders := map[string]string{
			":path":       "/anthropic/v1/messages",
			"x-sensitive": "original-secret",
			"x-existing":  "original-value",
		}

		rp := &messagesProcessorRouterFilter{
			requestHeaders: originalHeaders,
			originalRequestBody: &anthropicschema.MessagesRequest{
				"model":  "claude-3-sonnet",
				"stream": false,
			},
			upstreamFilterCount: 0,
		}

		err := p.SetBackend(t.Context(), &filterapi.Backend{
			Name:           "test-backend",
			Schema:         filterapi.VersionedAPISchema{Name: filterapi.APISchemaGCPAnthropic, Version: "vertex-2023-10-16"},
			HeaderMutation: headerMutations,
		}, nil, rp)
		require.NoError(t, err)

		// Verify header mutator was created.
		require.NotNil(t, p.headerMutator)

		// Test that the header mutator works correctly.
		testHeaders := map[string]string{
			"x-sensitive": "current-secret",
			"x-existing":  "current-value",
		}
		mutation := p.headerMutator.Mutate(testHeaders, false)

		require.NotNil(t, mutation)
		require.ElementsMatch(t, []string{"x-sensitive"}, mutation.RemoveHeaders)
		require.Len(t, mutation.SetHeaders, 1)
		require.Equal(t, "x-backend", mutation.SetHeaders[0].Header.Key)
		require.Equal(t, []byte("backend-value"), mutation.SetHeaders[0].Header.RawValue)
	})

	t.Run("header mutator with original headers", func(t *testing.T) {
		headers := map[string]string{":path": "/anthropic/v1/messages"}
		chatMetrics := metrics.NewChatCompletion(noop.NewMeterProvider().Meter("test"), map[string]string{})
		p := &messagesProcessorUpstreamFilter{
			config:         &processorConfig{},
			requestHeaders: headers,
			logger:         slog.Default(),
			metrics:        chatMetrics,
		}

		// Create backend with header mutations that don't remove x-custom.
		headerMutations := &filterapi.HTTPHeaderMutation{
			Remove: []string{"authorization"},
		}

		// Original headers from router filter (simulate what would be in rp.requestHeaders).
		originalHeaders := map[string]string{
			":path":         "/anthropic/v1/messages",
			"authorization": "bearer original-token", // This will be removed, so won't be restored.
			"x-custom":      "original-value",        // This won't be removed, so can be restored.
			"x-existing":    "existing-value",        // This won't be removed, so can be restored.
		}

		rp := &messagesProcessorRouterFilter{
			requestHeaders: originalHeaders,
			originalRequestBody: &anthropicschema.MessagesRequest{
				"model":  "claude-3-sonnet",
				"stream": false,
			},
			upstreamFilterCount: 0,
		}

		err := p.SetBackend(t.Context(), &filterapi.Backend{
			Name:           "test-backend",
			Schema:         filterapi.VersionedAPISchema{Name: filterapi.APISchemaGCPAnthropic, Version: "vertex-2023-10-16"},
			HeaderMutation: headerMutations,
		}, nil, rp)
		require.NoError(t, err)

		// Verify header mutator was created with original headers.
		require.NotNil(t, p.headerMutator)

		// Test retry scenario - original headers should be restored.
		testHeaders := map[string]string{
			"x-existing": "current-value", // This exists, so won't be restored.
		}
		mutation := p.headerMutator.Mutate(testHeaders, true) // onRetry = true.

		require.NotNil(t, mutation)
		// RemoveHeaders should be empty because authorization doesn't exist in testHeaders.
		require.Empty(t, mutation.RemoveHeaders)

		// Should restore x-custom header (not being removed and not already present).
		var restoredHeader *corev3.HeaderValueOption
		for _, h := range mutation.SetHeaders {
			if h.Header.Key == "x-custom" {
				restoredHeader = h
				break
			}
		}
		require.NotNil(t, restoredHeader)
		require.Equal(t, []byte("original-value"), restoredHeader.Header.RawValue)
		require.Equal(t, "original-value", testHeaders["x-custom"])
		// x-existing should not be restored because it already exists.
		require.Equal(t, "current-value", testHeaders["x-existing"])
	})
}
