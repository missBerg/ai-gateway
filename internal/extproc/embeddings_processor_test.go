// Copyright Envoy AI Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package extproc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/ai-gateway/internal/apischema/openai"
	"github.com/envoyproxy/ai-gateway/internal/extproc/headermutator"
	"github.com/envoyproxy/ai-gateway/internal/extproc/translator"
	"github.com/envoyproxy/ai-gateway/internal/filterapi"
	"github.com/envoyproxy/ai-gateway/internal/internalapi"
	"github.com/envoyproxy/ai-gateway/internal/llmcostcel"
	tracing "github.com/envoyproxy/ai-gateway/internal/tracing/api"
)

func TestEmbeddings_Schema(t *testing.T) {
	t.Run("supported openai / on route", func(t *testing.T) {
		cfg := &processorConfig{}
		routeFilter, err := EmbeddingsProcessorFactory(nil)(cfg, nil, slog.Default(), tracing.NoopTracing{}, false)
		require.NoError(t, err)
		require.NotNil(t, routeFilter)
		require.IsType(t, &embeddingsProcessorRouterFilter{}, routeFilter)
	})
	t.Run("supported openai / on upstream", func(t *testing.T) {
		cfg := &processorConfig{}
		routeFilter, err := EmbeddingsProcessorFactory(nil)(cfg, nil, slog.Default(), tracing.NoopTracing{}, true)
		require.NoError(t, err)
		require.NotNil(t, routeFilter)
		require.IsType(t, &embeddingsProcessorUpstreamFilter{}, routeFilter)
	})
}

func Test_embeddingsProcessorUpstreamFilter_SelectTranslator(t *testing.T) {
	e := &embeddingsProcessorUpstreamFilter{}
	t.Run("unsupported", func(t *testing.T) {
		err := e.selectTranslator(filterapi.VersionedAPISchema{Name: "Bar", Version: "v123"})
		require.ErrorContains(t, err, "unsupported API schema: backend={Bar v123}")
	})
	t.Run("supported openai", func(t *testing.T) {
		err := e.selectTranslator(filterapi.VersionedAPISchema{Name: filterapi.APISchemaOpenAI})
		require.NoError(t, err)
		require.NotNil(t, e.translator)
	})
}

func Test_embeddingsProcessorRouterFilter_ProcessRequestBody(t *testing.T) {
	t.Run("body parser error", func(t *testing.T) {
		p := &embeddingsProcessorRouterFilter{}
		_, err := p.ProcessRequestBody(t.Context(), &extprocv3.HttpBody{Body: []byte("nonjson")})
		require.ErrorContains(t, err, "invalid character 'o' in literal null")
	})

	t.Run("ok", func(t *testing.T) {
		headers := map[string]string{":path": "/foo"}
		const modelKey = "x-ai-gateway-model-key"
		p := &embeddingsProcessorRouterFilter{
			config:         &processorConfig{modelNameHeaderKey: modelKey},
			requestHeaders: headers,
			logger:         slog.Default(),
		}
		resp, err := p.ProcessRequestBody(t.Context(), &extprocv3.HttpBody{Body: embeddingBodyFromModel(t, "some-model")})
		require.NoError(t, err)
		require.NotNil(t, resp)
		re, ok := resp.Response.(*extprocv3.ProcessingResponse_RequestBody)
		require.True(t, ok)
		require.NotNil(t, re)
		require.NotNil(t, re.RequestBody)
		setHeaders := re.RequestBody.GetResponse().GetHeaderMutation().SetHeaders
		require.Len(t, setHeaders, 2)
		require.Equal(t, modelKey, setHeaders[0].Header.Key)
		require.Equal(t, "some-model", string(setHeaders[0].Header.RawValue))
		require.Equal(t, "x-ai-eg-original-path", setHeaders[1].Header.Key)
		require.Equal(t, "/foo", string(setHeaders[1].Header.RawValue))
	})
}

func Test_embeddingsProcessorUpstreamFilter_ProcessResponseHeaders(t *testing.T) {
	t.Run("error translation", func(t *testing.T) {
		mm := &mockEmbeddingsMetrics{}
		mt := &mockEmbeddingTranslator{t: t, expHeaders: make(map[string]string)}
		p := &embeddingsProcessorUpstreamFilter{
			translator: mt,
			metrics:    mm,
		}
		mt.retErr = errors.New("test error")
		_, err := p.ProcessResponseHeaders(t.Context(), nil)
		require.ErrorContains(t, err, "test error")
		mm.RequireRequestFailure(t)
	})
	t.Run("ok", func(t *testing.T) {
		inHeaders := &corev3.HeaderMap{
			Headers: []*corev3.HeaderValue{{Key: "foo", Value: "bar"}, {Key: "dog", RawValue: []byte("cat")}},
		}
		expHeaders := map[string]string{"foo": "bar", "dog": "cat"}
		mm := &mockEmbeddingsMetrics{}
		mt := &mockEmbeddingTranslator{t: t, expHeaders: expHeaders}
		p := &embeddingsProcessorUpstreamFilter{
			translator: mt,
			metrics:    mm,
		}
		res, err := p.ProcessResponseHeaders(t.Context(), inHeaders)
		require.NoError(t, err)
		commonRes := res.Response.(*extprocv3.ProcessingResponse_ResponseHeaders).ResponseHeaders.Response
		require.Equal(t, mt.retHeaderMutation, commonRes.HeaderMutation)
		mm.RequireRequestNotCompleted(t)
	})
}

func embeddingBodyFromModel(_ *testing.T, model string) []byte {
	return fmt.Appendf(nil, `{"model":"%s","input":"test input"}`, model)
}

func Test_embeddingsProcessorUpstreamFilter_ProcessResponseBody(t *testing.T) {
	t.Run("error translation", func(t *testing.T) {
		mm := &mockEmbeddingsMetrics{}
		mt := &mockEmbeddingTranslator{t: t}
		p := &embeddingsProcessorUpstreamFilter{
			translator:      mt,
			metrics:         mm,
			responseHeaders: map[string]string{":status": "200"},
		}
		mt.retErr = errors.New("test error")
		_, err := p.ProcessResponseBody(t.Context(), &extprocv3.HttpBody{})
		require.ErrorContains(t, err, "test error")
		mm.RequireRequestFailure(t)
		mm.RequireTokenUsage(t, 0)
	})
	t.Run("ok", func(t *testing.T) {
		inBody := &extprocv3.HttpBody{Body: []byte("some-body"), EndOfStream: true}
		expBodyMut := &extprocv3.BodyMutation{}
		expHeadMut := &extprocv3.HeaderMutation{}
		mm := &mockEmbeddingsMetrics{}
		mt := &mockEmbeddingTranslator{
			t: t, expResponseBody: inBody,
			retBodyMutation: expBodyMut, retHeaderMutation: expHeadMut,
			retUsedToken: translator.LLMTokenUsage{InputTokens: 123, TotalTokens: 123},
		}

		celProgInt, err := llmcostcel.NewProgram("54321")
		require.NoError(t, err)
		celProgUint, err := llmcostcel.NewProgram("uint(9999)")
		require.NoError(t, err)
		p := &embeddingsProcessorUpstreamFilter{
			translator: mt,
			logger:     slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})),
			metrics:    mm,
			config: &processorConfig{
				metadataNamespace:  "ai_gateway_llm_ns",
				modelNameHeaderKey: "x-aigw-model",
				requestCosts: []processorConfigRequestCost{
					{LLMRequestCost: &filterapi.LLMRequestCost{Type: filterapi.LLMRequestCostTypeInputToken, MetadataKey: "input_token_usage"}},
					{LLMRequestCost: &filterapi.LLMRequestCost{Type: filterapi.LLMRequestCostTypeTotalToken, MetadataKey: "total_token_usage"}},
					{
						celProg:        celProgInt,
						LLMRequestCost: &filterapi.LLMRequestCost{Type: filterapi.LLMRequestCostTypeCEL, MetadataKey: "cel_int"},
					},
					{
						celProg:        celProgUint,
						LLMRequestCost: &filterapi.LLMRequestCost{Type: filterapi.LLMRequestCostTypeCEL, MetadataKey: "cel_uint"},
					},
				},
			},
			requestHeaders:    map[string]string{"x-aigw-model": "some_model"},
			backendName:       "some_backend",
			modelNameOverride: "some_model",
			responseHeaders:   map[string]string{":status": "200"},
		}
		res, err := p.ProcessResponseBody(t.Context(), inBody)
		require.NoError(t, err)
		commonRes := res.Response.(*extprocv3.ProcessingResponse_ResponseBody).ResponseBody.Response
		require.Equal(t, expBodyMut, commonRes.BodyMutation)
		require.Equal(t, expHeadMut, commonRes.HeaderMutation)
		mm.RequireRequestSuccess(t)
		mm.RequireTokenUsage(t, 123)

		md := res.DynamicMetadata
		require.NotNil(t, md)
		require.Equal(t, float64(123), md.Fields["ai_gateway_llm_ns"].
			GetStructValue().Fields["input_token_usage"].GetNumberValue())
		require.Equal(t, float64(123), md.Fields["ai_gateway_llm_ns"].
			GetStructValue().Fields["total_token_usage"].GetNumberValue())
		require.Equal(t, float64(54321), md.Fields["ai_gateway_llm_ns"].
			GetStructValue().Fields["cel_int"].GetNumberValue())
		require.Equal(t, float64(9999), md.Fields["ai_gateway_llm_ns"].
			GetStructValue().Fields["cel_uint"].GetNumberValue())
		require.Equal(t, "some_backend", md.Fields["ai_gateway_llm_ns"].
			GetStructValue().Fields["backend_name"].GetStringValue())
		require.Equal(t, "some_model", md.Fields["ai_gateway_llm_ns"].
			GetStructValue().Fields["model_name_override"].GetStringValue())
	})

	t.Run("error/streaming", func(t *testing.T) {
		inBody := &extprocv3.HttpBody{Body: []byte("some-body"), EndOfStream: true}
		mm := &mockEmbeddingsMetrics{}
		mt := &mockEmbeddingTranslator{t: t, expResponseBody: inBody}
		p := &embeddingsProcessorUpstreamFilter{
			translator:        mt,
			logger:            slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})),
			metrics:           mm,
			config:            &processorConfig{},
			backendName:       "some_backend",
			modelNameOverride: "some_model",
			responseHeaders:   map[string]string{":status": "500"},
		}
		res, err := p.ProcessResponseBody(t.Context(), inBody)
		require.NoError(t, err)
		commonRes := res.Response.(*extprocv3.ProcessingResponse_ResponseBody).ResponseBody.Response
		require.NotNil(t, commonRes)
		require.True(t, mt.responseErrorCalled)
		// Ensure failure metric recorded for non-2xx.
		mm.RequireRequestFailure(t)
	})

	// Success should be recorded only when EndOfStream is true.
	t.Run("completion only at end", func(t *testing.T) {
		mm := &mockEmbeddingsMetrics{}
		mt := &mockEmbeddingTranslator{t: t}
		p := &embeddingsProcessorUpstreamFilter{
			translator:        mt,
			logger:            slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})),
			metrics:           mm,
			config:            &processorConfig{},
			backendName:       "some_backend",
			modelNameOverride: "some_model",
			responseHeaders:   map[string]string{":status": "200"},
		}

		// First chunk (not end of stream) should not complete the request.
		chunk := &extprocv3.HttpBody{Body: []byte("chunk-1"), EndOfStream: false}
		mt.expResponseBody = chunk
		_, err := p.ProcessResponseBody(t.Context(), chunk)
		require.NoError(t, err)
		mm.RequireRequestNotCompleted(t)

		// Final chunk should mark success.
		final := &extprocv3.HttpBody{Body: []byte("chunk-final"), EndOfStream: true}
		mt.expResponseBody = final
		_, err = p.ProcessResponseBody(t.Context(), final)
		require.NoError(t, err)
		mm.RequireRequestSuccess(t)
	})
}

func Test_embeddingsProcessorUpstreamFilter_SetBackend(t *testing.T) {
	headers := map[string]string{":path": "/foo"}
	mm := &mockEmbeddingsMetrics{}
	p := &embeddingsProcessorUpstreamFilter{
		config:         &processorConfig{},
		requestHeaders: headers,
		logger:         slog.Default(),
		metrics:        mm,
	}
	err := p.SetBackend(t.Context(), &filterapi.Backend{
		Name:   "some-backend",
		Schema: filterapi.VersionedAPISchema{Name: "some-schema", Version: "v10.0"},
	}, nil, &embeddingsProcessorRouterFilter{})
	require.ErrorContains(t, err, "unsupported API schema: backend={some-schema v10.0}")
	mm.RequireRequestFailure(t)
	mm.RequireTokenUsage(t, 0)
	mm.RequireSelectedBackend(t, "some-backend")
}

func Test_embeddingsProcessorUpstreamFilter_SetBackend_Success(t *testing.T) {
	headers := map[string]string{":path": "/foo", "x-model-name": "some-model"}
	mm := &mockEmbeddingsMetrics{}
	p := &embeddingsProcessorUpstreamFilter{
		config:         &processorConfig{modelNameHeaderKey: "x-model-name"},
		requestHeaders: headers,
		logger:         slog.Default(),
		metrics:        mm,
	}
	rp := &embeddingsProcessorRouterFilter{
		originalRequestBody: &openai.EmbeddingRequest{},
	}
	err := p.SetBackend(t.Context(), &filterapi.Backend{
		Name:              "openai",
		Schema:            filterapi.VersionedAPISchema{Name: filterapi.APISchemaOpenAI, Version: "v1"},
		ModelNameOverride: "override-model",
	}, nil, rp)
	require.NoError(t, err)
	mm.RequireSelectedBackend(t, "openai")
	require.Equal(t, "override-model", p.requestHeaders["x-model-name"])
	require.NotNil(t, p.translator)
}

func Test_embeddingsProcessorUpstreamFilter_ProcessRequestHeaders(t *testing.T) {
	const modelKey = "x-ai-gateway-model-key"
	t.Run("translator error", func(t *testing.T) {
		headers := map[string]string{":path": "/foo", modelKey: "some-model"}
		someBody := embeddingBodyFromModel(t, "some-model")
		var body openai.EmbeddingRequest
		require.NoError(t, json.Unmarshal(someBody, &body))
		tr := &mockEmbeddingTranslator{t: t, retErr: errors.New("test error"), expRequestBody: &body}
		mm := &mockEmbeddingsMetrics{}
		p := &embeddingsProcessorUpstreamFilter{
			config: &processorConfig{
				modelNameHeaderKey: modelKey,
			},
			requestHeaders:         headers,
			logger:                 slog.Default(),
			metrics:                mm,
			translator:             tr,
			originalRequestBodyRaw: someBody,
			originalRequestBody:    &body,
		}
		_, err := p.ProcessRequestHeaders(t.Context(), nil)
		require.ErrorContains(t, err, "failed to transform request: test error")
		mm.RequireRequestFailure(t)
		mm.RequireTokenUsage(t, 0)
		// Verify request model was set even though processing failed
		require.Equal(t, "some-model", mm.requestModel)
		require.Empty(t, mm.responseModel)
	})
	t.Run("ok", func(t *testing.T) {
		someBody := embeddingBodyFromModel(t, "some-model")
		headers := map[string]string{":path": "/foo", modelKey: "some-model"}
		headerMut := &extprocv3.HeaderMutation{
			SetHeaders: []*corev3.HeaderValueOption{{Header: &corev3.HeaderValue{Key: "foo", RawValue: []byte("bar")}}},
		}
		bodyMut := &extprocv3.BodyMutation{Mutation: &extprocv3.BodyMutation_Body{Body: []byte("some body")}}

		var expBody openai.EmbeddingRequest
		require.NoError(t, json.Unmarshal(someBody, &expBody))
		mt := &mockEmbeddingTranslator{t: t, expRequestBody: &expBody, retHeaderMutation: headerMut, retBodyMutation: bodyMut}
		mm := &mockEmbeddingsMetrics{}
		p := &embeddingsProcessorUpstreamFilter{
			config:                 &processorConfig{modelNameHeaderKey: modelKey},
			requestHeaders:         headers,
			logger:                 slog.Default(),
			metrics:                mm,
			translator:             mt,
			originalRequestBodyRaw: someBody,
			originalRequestBody:    &expBody,
			handler:                &mockBackendAuthHandler{},
		}
		resp, err := p.ProcessRequestHeaders(t.Context(), nil)
		require.NoError(t, err)
		require.Equal(t, mt, p.translator)
		require.NotNil(t, resp)
		commonRes := resp.Response.(*extprocv3.ProcessingResponse_RequestHeaders).RequestHeaders.Response
		require.Equal(t, headerMut, commonRes.HeaderMutation)
		require.Equal(t, bodyMut, commonRes.BodyMutation)

		mm.RequireRequestNotCompleted(t)
		// Verify request model was set
		require.Equal(t, "some-model", mm.requestModel)
		// Response model not set yet - only set when we get actual response
		require.Empty(t, mm.responseModel)
	})
}

func TestEmbeddings_ProcessRequestHeaders_SetsRequestModel(t *testing.T) {
	const modelKey = internalapi.ModelNameHeaderKeyDefault
	headers := map[string]string{":path": "/v1/embeddings", modelKey: "header-model"}
	body := openai.EmbeddingRequest{Model: "body-model"}
	raw, _ := json.Marshal(body)
	mm := &mockEmbeddingsMetrics{}
	p := &embeddingsProcessorUpstreamFilter{
		config:                 &processorConfig{modelNameHeaderKey: modelKey},
		requestHeaders:         headers,
		logger:                 slog.Default(),
		metrics:                mm,
		translator:             &mockEmbeddingTranslator{t: t, expRequestBody: &body},
		originalRequestBodyRaw: raw,
		originalRequestBody:    &body,
	}
	_, _ = p.ProcessRequestHeaders(t.Context(), nil)
	// Should use the override model from the header, as that's what is sent upstream.
	require.Equal(t, "header-model", mm.requestModel)
	// Response model is not set until we get actual response
	require.Empty(t, mm.responseModel)
}

func TestEmbeddings_ProcessResponseBody_OverridesHeaderModelWithResponseModel(t *testing.T) {
	const modelKey = internalapi.ModelNameHeaderKeyDefault
	headers := map[string]string{":path": "/v1/embeddings", modelKey: "header-model"}
	body := openai.EmbeddingRequest{
		Model: "body-model",
		Input: openai.StringOrArray{Value: "test"},
	}
	raw, _ := json.Marshal(body)
	mm := &mockEmbeddingsMetrics{}

	// Create a mock translator that returns token usage with response model
	mt := &mockEmbeddingTranslator{
		t:              t,
		expRequestBody: &body,
		expHeaders:     map[string]string{":status": "200"},
		retUsedToken: translator.LLMTokenUsage{
			InputTokens: 15,
		},
		retResponseModel: "actual-embedding-model",
	}

	p := &embeddingsProcessorUpstreamFilter{
		config:                 &processorConfig{modelNameHeaderKey: modelKey},
		requestHeaders:         headers,
		logger:                 slog.Default(),
		metrics:                mm,
		translator:             mt,
		originalRequestBodyRaw: raw,
		originalRequestBody:    &body,
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

	// Now process response body (should override with actual response model)
	responseBytes := []byte(`{"model":"actual-embedding-model","data":[{"embedding":[0.1,0.2]}],"usage":{"prompt_tokens":15,"total_tokens":15}}`)
	_, err = p.ProcessResponseBody(t.Context(), &extprocv3.HttpBody{
		Body:        responseBytes,
		EndOfStream: true,
	})
	require.NoError(t, err)

	// Should use the override model from the header, as that's what is sent upstream.
	mm.RequireSelectedModel(t, "header-model", "actual-embedding-model")
	mm.RequireTokenUsage(t, 15)
	mm.RequireRequestSuccess(t)
}

func TestEmbeddings_ParseBody(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		jsonBody := `{"model":"text-embedding-ada-002","input":"test input"}`
		modelName, rb, err := parseOpenAIEmbeddingBody(&extprocv3.HttpBody{Body: []byte(jsonBody)})
		require.NoError(t, err)
		require.Equal(t, "text-embedding-ada-002", modelName)
		require.NotNil(t, rb)
		require.Equal(t, "text-embedding-ada-002", rb.Model)
		require.Equal(t, "test input", rb.Input.Value)
	})
	t.Run("error", func(t *testing.T) {
		modelName, rb, err := parseOpenAIEmbeddingBody(&extprocv3.HttpBody{})
		require.Error(t, err)
		require.Empty(t, modelName)
		require.Nil(t, rb)
	})
}

func TestEmbeddingsProcessorRouterFilter_ProcessResponseHeaders_ProcessResponseBody(t *testing.T) {
	t.Run("no ok path with passthrough", func(t *testing.T) {
		p := &embeddingsProcessorRouterFilter{}
		_, err := p.ProcessResponseHeaders(t.Context(), nil)
		require.NoError(t, err)
		_, err = p.ProcessResponseBody(t.Context(), nil)
		require.NoError(t, err)
	})
	t.Run("ok path with upstream filter", func(t *testing.T) {
		p := &embeddingsProcessorRouterFilter{
			upstreamFilter: &embeddingsProcessorUpstreamFilter{
				translator: &mockEmbeddingTranslator{t: t, expHeaders: map[string]string{}},
				logger:     slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})),
				metrics:    &mockEmbeddingsMetrics{},
				config:     &processorConfig{metadataNamespace: ""},
			},
		}
		resp, err := p.ProcessResponseHeaders(t.Context(), &corev3.HeaderMap{Headers: []*corev3.HeaderValue{}})
		require.NoError(t, err)
		require.NotNil(t, resp)

		resp, err = p.ProcessResponseBody(t.Context(), &extprocv3.HttpBody{Body: []byte("some body")})
		require.NoError(t, err)
		require.NotNil(t, resp)
		re, ok := resp.Response.(*extprocv3.ProcessingResponse_ResponseBody)
		require.True(t, ok)
		require.NotNil(t, re)
		require.NotNil(t, re.ResponseBody)
		require.NotNil(t, re.ResponseBody.Response)
		require.IsType(t, &extprocv3.BodyMutation{}, re.ResponseBody.Response.BodyMutation)
		require.IsType(t, &extprocv3.HeaderMutation{}, re.ResponseBody.Response.HeaderMutation)
	})
}

func TestEmbeddingsProcessorUpstreamFilter_ProcessRequestHeaders_WithHeaderMutations(t *testing.T) {
	const testModelKey = "x-ai-gateway-model-key"
	t.Run("header mutations applied correctly", func(t *testing.T) {
		headers := map[string]string{
			":path":         "/v1/embeddings",
			testModelKey:    "some-model",
			"authorization": "bearer token123",
			"x-api-key":     "secret-key",
			"x-custom":      "custom-value",
		}
		someBody := embeddingBodyFromModel(t, "some-model")
		var body openai.EmbeddingRequest
		require.NoError(t, json.Unmarshal(someBody, &body))

		// Create header mutations.
		headerMutations := &filterapi.HTTPHeaderMutation{
			Remove: []string{"authorization", "x-api-key"},
			Set:    []filterapi.HTTPHeader{{Name: "x-new-header", Value: "new-value"}},
		}

		mt := &mockEmbeddingTranslator{t: t, expRequestBody: &body}
		mm := &mockEmbeddingsMetrics{}
		p := &embeddingsProcessorUpstreamFilter{
			config:                 &processorConfig{modelNameHeaderKey: testModelKey},
			requestHeaders:         headers,
			logger:                 slog.Default(),
			metrics:                mm,
			translator:             mt,
			originalRequestBodyRaw: someBody,
			originalRequestBody:    &body,
			handler:                &mockBackendAuthHandler{},
		}

		// Set header mutator.
		originalHeaders := map[string]string{
			"authorization": "bearer original-token",
			"x-api-key":     "original-secret",
		}
		p.headerMutator = headermutator.NewHeaderMutator(headerMutations, originalHeaders)

		resp, err := p.ProcessRequestHeaders(t.Context(), nil)
		require.NoError(t, err)
		require.NotNil(t, resp)

		commonRes := resp.Response.(*extprocv3.ProcessingResponse_RequestHeaders).RequestHeaders.Response

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
			":path":      "/v1/embeddings",
			testModelKey: "some-model",
			// "x-custom" is not present in current headers, so it can be restored.
			"x-new-header": "new-value", // Already set from previous mutation.
		}
		someBody := embeddingBodyFromModel(t, "some-model")
		var body openai.EmbeddingRequest
		require.NoError(t, json.Unmarshal(someBody, &body))

		// Create header mutations that don't remove x-custom (so it can be restored).
		headerMutations := &filterapi.HTTPHeaderMutation{
			Remove: []string{"authorization", "x-api-key"},
			Set:    []filterapi.HTTPHeader{{Name: "x-new-header", Value: "updated-value"}},
		}

		mt := &mockEmbeddingTranslator{t: t, expRequestBody: &body}
		mm := &mockEmbeddingsMetrics{}
		p := &embeddingsProcessorUpstreamFilter{
			config:                 &processorConfig{modelNameHeaderKey: testModelKey},
			requestHeaders:         headers,
			logger:                 slog.Default(),
			metrics:                mm,
			translator:             mt,
			originalRequestBodyRaw: someBody,
			originalRequestBody:    &body,
			handler:                &mockBackendAuthHandler{},
			onRetry:                true, // This is a retry request.
		}

		// Use the same headers map as the original headers (this simulates the router filter's requestHeaders).
		originalHeaders := map[string]string{
			":path":         "/v1/embeddings",
			testModelKey:    "some-model",
			"authorization": "bearer original-token", // This will be removed, so won't be restored.
			"x-api-key":     "original-secret",       // This will be removed, so won't be restored.
			"x-custom":      "original-custom",       // This won't be removed, so can be restored.
			"x-new-header":  "original-value",        // This will be set, so won't be restored.
		}
		p.headerMutator = headermutator.NewHeaderMutator(headerMutations, originalHeaders)

		resp, err := p.ProcessRequestHeaders(t.Context(), nil)
		require.NoError(t, err)
		require.NotNil(t, resp)

		commonRes := resp.Response.(*extprocv3.ProcessingResponse_RequestHeaders).RequestHeaders.Response

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
			":path":         "/v1/embeddings",
			testModelKey:    "some-model",
			"authorization": "bearer token123",
		}
		someBody := embeddingBodyFromModel(t, "some-model")
		var body openai.EmbeddingRequest
		require.NoError(t, json.Unmarshal(someBody, &body))

		mt := &mockEmbeddingTranslator{t: t, expRequestBody: &body}
		mm := &mockEmbeddingsMetrics{}
		p := &embeddingsProcessorUpstreamFilter{
			config:                 &processorConfig{modelNameHeaderKey: testModelKey},
			requestHeaders:         headers,
			logger:                 slog.Default(),
			metrics:                mm,
			translator:             mt,
			originalRequestBodyRaw: someBody,
			originalRequestBody:    &body,
			handler:                &mockBackendAuthHandler{},
			headerMutator:          nil, // No header mutator.
		}

		resp, err := p.ProcessRequestHeaders(t.Context(), nil)
		require.NoError(t, err)
		require.NotNil(t, resp)

		commonRes := resp.Response.(*extprocv3.ProcessingResponse_RequestHeaders).RequestHeaders.Response

		// Check that no header mutations were applied.
		require.NotNil(t, commonRes.HeaderMutation)
		require.Empty(t, commonRes.HeaderMutation.RemoveHeaders)
		require.Empty(t, commonRes.HeaderMutation.SetHeaders)

		// Check that original headers remain unchanged.
		require.Equal(t, "bearer token123", headers["authorization"])
	})
}

func TestEmbeddingsProcessorUpstreamFilter_SetBackend_WithHeaderMutations(t *testing.T) {
	t.Run("header mutator created correctly", func(t *testing.T) {
		headers := map[string]string{":path": "/foo"}
		mm := &mockEmbeddingsMetrics{}
		p := &embeddingsProcessorUpstreamFilter{
			config:         &processorConfig{},
			requestHeaders: headers,
			logger:         slog.Default(),
			metrics:        mm,
		}

		// Create backend with header mutations.
		headerMutations := &filterapi.HTTPHeaderMutation{
			Remove: []string{"x-sensitive"},
			Set:    []filterapi.HTTPHeader{{Name: "x-backend", Value: "backend-value"}},
		}

		rp := &embeddingsProcessorRouterFilter{
			requestHeaders: headers,
		}

		err := p.SetBackend(t.Context(), &filterapi.Backend{
			Name:           "test-backend",
			Schema:         filterapi.VersionedAPISchema{Name: filterapi.APISchemaOpenAI},
			HeaderMutation: headerMutations,
		}, nil, rp)
		require.NoError(t, err)

		// Verify header mutator was created.
		require.NotNil(t, p.headerMutator)

		// Test that the header mutator works correctly.
		testHeaders := map[string]string{
			"x-sensitive": "secret",
			"x-existing":  "value",
		}
		mutation := p.headerMutator.Mutate(testHeaders, false) // onRetry = false.

		require.NotNil(t, mutation)
		require.ElementsMatch(t, []string{"x-sensitive"}, mutation.RemoveHeaders)
		require.Len(t, mutation.SetHeaders, 1)
		require.Equal(t, "x-backend", mutation.SetHeaders[0].Header.Key)
		require.Equal(t, []byte("backend-value"), mutation.SetHeaders[0].Header.RawValue)
	})

	t.Run("header mutator with original headers", func(t *testing.T) {
		headers := map[string]string{":path": "/foo"}
		mm := &mockEmbeddingsMetrics{}
		p := &embeddingsProcessorUpstreamFilter{
			config:         &processorConfig{},
			requestHeaders: headers,
			logger:         slog.Default(),
			metrics:        mm,
		}

		// Create backend with header mutations that don't remove x-custom.
		headerMutations := &filterapi.HTTPHeaderMutation{
			Remove: []string{"authorization"},
		}

		// Original headers from router filter (simulate what would be in rp.requestHeaders).
		originalHeaders := map[string]string{
			":path":         "/foo",
			"authorization": "bearer original-token", // This will be removed, so won't be restored.
			"x-custom":      "original-value",        // This won't be removed, so can be restored.
			"x-existing":    "existing-value",        // This won't be removed, so can be restored.
		}

		rp := &embeddingsProcessorRouterFilter{
			requestHeaders: originalHeaders,
		}

		err := p.SetBackend(t.Context(), &filterapi.Backend{
			Name:           "test-backend",
			Schema:         filterapi.VersionedAPISchema{Name: filterapi.APISchemaOpenAI},
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
