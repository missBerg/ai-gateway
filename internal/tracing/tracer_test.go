// Copyright Envoy AI Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package tracing

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"k8s.io/utils/ptr"

	"github.com/envoyproxy/ai-gateway/internal/apischema/openai"
	tracing "github.com/envoyproxy/ai-gateway/internal/tracing/api"
)

var (
	startOpts = []oteltrace.SpanStartOption{oteltrace.WithSpanKind(oteltrace.SpanKindServer)}

	req = &openai.ChatCompletionRequest{
		Model: openai.ModelGPT5Nano,
		Messages: []openai.ChatCompletionMessageParamUnion{{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Content: openai.StringOrUserRoleContentUnion{Value: "Hello!"},
				Role:    openai.ChatMessageRoleUser,
			},
		}},
	}
)

func TestTracer_StartSpanAndInjectHeaders(t *testing.T) {
	respBody := &openai.ChatCompletionResponse{
		ID:     "chatcmpl-abc123",
		Object: "chat.completion",
		Model:  "gpt-4.1-nano",
		Choices: []openai.ChatCompletionResponseChoice{
			{
				Index: 0,
				Message: openai.ChatCompletionResponseChoiceMessage{
					Role:    "assistant",
					Content: ptr.To("hello world"),
				},
				FinishReason: "stop",
			},
		},
	}
	respBodyBytes, err := json.Marshal(respBody)
	require.NoError(t, err)
	bodyLen := len(respBodyBytes)

	reqStream := *req
	reqStream.Stream = true

	tests := []struct {
		name             string
		req              *openai.ChatCompletionRequest
		existingHeaders  map[string]string
		expectedSpanName string
		expectedAttrs    []attribute.KeyValue
		expectedTraceID  string
	}{
		{
			name:             "non-streaming request",
			req:              req,
			existingHeaders:  map[string]string{},
			expectedSpanName: "non-stream len: 70",
			expectedAttrs: []attribute.KeyValue{
				attribute.String("req", "stream: false"),
				attribute.Int("reqBodyLen", 70),
				attribute.Int("statusCode", 200),
				attribute.Int("respBodyLen", bodyLen),
			},
		},
		{
			name:             "streaming request",
			req:              &reqStream,
			existingHeaders:  map[string]string{},
			expectedSpanName: "stream len: 84",
			expectedAttrs: []attribute.KeyValue{
				attribute.String("req", "stream: true"),
				attribute.Int("reqBodyLen", 84),
				attribute.Int("statusCode", 200),
				attribute.Int("respBodyLen", bodyLen),
			},
		},
		{
			name: "with existing trace context",
			req:  req,
			existingHeaders: map[string]string{
				"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
			},
			expectedSpanName: "non-stream len: 70",
			expectedAttrs: []attribute.KeyValue{
				attribute.String("req", "stream: false"),
				attribute.Int("reqBodyLen", 70),
				attribute.Int("statusCode", 200),
				attribute.Int("respBodyLen", bodyLen),
			},
			expectedTraceID: "4bf92f3577b34da6a3ce929d0e0e4736",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exporter := tracetest.NewInMemoryExporter()
			tp := trace.NewTracerProvider(trace.WithSyncer(exporter))

			tracer := newChatCompletionTracer(tp.Tracer("test"), autoprop.NewTextMapPropagator(), testChatCompletionRecorder{})

			headerMutation := &extprocv3.HeaderMutation{}
			reqBody, err := json.Marshal(tt.req)
			require.NoError(t, err)

			span := tracer.StartSpanAndInjectHeaders(t.Context(),
				tt.existingHeaders,
				headerMutation,
				tt.req,
				reqBody,
			)
			require.IsType(t, &chatCompletionSpan{}, span)

			// End the span to export it.
			span.RecordResponse(respBody)
			span.EndSpan()

			spans := exporter.GetSpans()
			require.Len(t, spans, 1)
			actualSpan := spans[0]

			// Check span state.
			require.Equal(t, tt.expectedSpanName, actualSpan.Name)
			require.Equal(t, tt.expectedAttrs, actualSpan.Attributes)
			require.Empty(t, actualSpan.Events)

			// Check header mutation.
			traceID := actualSpan.SpanContext.TraceID().String()
			if tt.expectedTraceID != "" {
				require.Equal(t, tt.expectedTraceID, actualSpan.SpanContext.TraceID().String())
			}
			spanID := actualSpan.SpanContext.SpanID().String()
			require.Equal(t, &extprocv3.HeaderMutation{
				SetHeaders: []*corev3.HeaderValueOption{
					{
						Header: &corev3.HeaderValue{
							Key:      "traceparent",
							RawValue: []byte("00-" + traceID + "-" + spanID + "-01"),
						},
					},
				},
			}, headerMutation)
		})
	}
}

func TestNewChatCompletionTracer_Noop(t *testing.T) {
	// Use noop tracer.
	noopTracer := noop.Tracer{}

	tracer := newChatCompletionTracer(noopTracer, autoprop.NewTextMapPropagator(), testChatCompletionRecorder{})

	// Verify it returns NoopTracer.
	require.IsType(t, tracing.NoopChatCompletionTracer{}, tracer)

	// Test that noop tracer doesn't create spans.
	headers := map[string]string{}
	headerMutation := &extprocv3.HeaderMutation{}
	req := &openai.ChatCompletionRequest{Model: "test"}

	span := tracer.StartSpanAndInjectHeaders(t.Context(),
		headers,
		headerMutation,
		req,
		[]byte("{}"),
	)

	require.Nil(t, span)

	// Verify no headers were injected.
	require.Empty(t, headerMutation.SetHeaders)
}

func TestTracer_UnsampledSpan(t *testing.T) {
	// Use always_off sampler to ensure spans are not sampled.
	tracerProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.NeverSample()),
	)
	t.Cleanup(func() { _ = tracerProvider.Shutdown(context.Background()) })

	tracer := newChatCompletionTracer(tracerProvider.Tracer("test"), autoprop.NewTextMapPropagator(), testChatCompletionRecorder{})

	// Start a span that won't be sampled.
	headers := map[string]string{}
	headerMutation := &extprocv3.HeaderMutation{}
	req := &openai.ChatCompletionRequest{Model: "test"}

	span := tracer.StartSpanAndInjectHeaders(t.Context(),
		headers,
		headerMutation,
		req,
		[]byte("{}"),
	)

	// Span should be nil when not sampled.
	require.Nil(t, span)

	// Headers should still be injected for trace propagation.
	require.NotEmpty(t, headerMutation.SetHeaders)
}

func TestHeaderMutationCarrier(t *testing.T) {
	t.Run("Get panics", func(t *testing.T) {
		carrier := &headerMutationCarrier{m: &extprocv3.HeaderMutation{}}
		require.Panics(t, func() { carrier.Get("test-key") })
	})

	t.Run("Keys panics", func(t *testing.T) {
		carrier := &headerMutationCarrier{m: &extprocv3.HeaderMutation{}}
		require.Panics(t, func() { carrier.Keys() })
	})

	t.Run("Set headers", func(t *testing.T) {
		mutation := &extprocv3.HeaderMutation{}
		carrier := &headerMutationCarrier{m: mutation}

		carrier.Set("trace-id", "12345")
		carrier.Set("span-id", "67890")

		require.Equal(t, &extprocv3.HeaderMutation{
			SetHeaders: []*corev3.HeaderValueOption{
				{
					Header: &corev3.HeaderValue{
						Key:      "trace-id",
						RawValue: []byte("12345"),
					},
				},
				{
					Header: &corev3.HeaderValue{
						Key:      "span-id",
						RawValue: []byte("67890"),
					},
				},
			},
		}, mutation)
	})
}

var _ tracing.ChatCompletionRecorder = testChatCompletionRecorder{}

type testChatCompletionRecorder struct{}

func (r testChatCompletionRecorder) RecordResponseChunks(span oteltrace.Span, chunks []*openai.ChatCompletionResponseChunk) {
	span.SetAttributes(attribute.Int("eventCount", len(chunks)))
}

func (r testChatCompletionRecorder) RecordResponseOnError(span oteltrace.Span, statusCode int, body []byte) {
	span.SetAttributes(attribute.Int("statusCode", statusCode))
	span.SetAttributes(attribute.String("errorBody", string(body)))
}

func (testChatCompletionRecorder) StartParams(req *openai.ChatCompletionRequest, body []byte) (spanName string, opts []oteltrace.SpanStartOption) {
	if req.Stream {
		return fmt.Sprintf("stream len: %d", len(body)), startOpts
	}
	return fmt.Sprintf("non-stream len: %d", len(body)), startOpts
}

func (testChatCompletionRecorder) RecordRequest(span oteltrace.Span, req *openai.ChatCompletionRequest, body []byte) {
	span.SetAttributes(attribute.String("req", fmt.Sprintf("stream: %v", req.Stream)))
	span.SetAttributes(attribute.Int("reqBodyLen", len(body)))
}

func (testChatCompletionRecorder) RecordResponse(span oteltrace.Span, resp *openai.ChatCompletionResponse) {
	span.SetAttributes(attribute.Int("statusCode", 200))
	body, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	span.SetAttributes(attribute.Int("respBodyLen", len(body)))
}
