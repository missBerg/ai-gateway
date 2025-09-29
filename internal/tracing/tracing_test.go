// Copyright Envoy AI Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package tracing

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"k8s.io/utils/ptr"

	"github.com/envoyproxy/ai-gateway/internal/apischema/openai"
	"github.com/envoyproxy/ai-gateway/internal/testing/testotel"
	tracing "github.com/envoyproxy/ai-gateway/internal/tracing/api"
	"github.com/envoyproxy/ai-gateway/internal/tracing/openinference"
	openaitracing "github.com/envoyproxy/ai-gateway/internal/tracing/openinference/openai"
)

// TestNewTracingFromEnv_DefaultServiceName tests that the service name.
// defaults to "ai-gateway" when OTEL_SERVICE_NAME is not set.
func TestNewTracingFromEnv_DefaultServiceName(t *testing.T) {
	tests := []struct {
		name              string
		env               map[string]string
		expectServiceName string
	}{
		{
			name: "default service name when OTEL_SERVICE_NAME not set",
			env: map[string]string{
				"OTEL_TRACES_EXPORTER": "console",
			},
			expectServiceName: "ai-gateway",
		},
		{
			name: "OTEL_SERVICE_NAME overrides default",
			env: map[string]string{
				"OTEL_TRACES_EXPORTER": "console",
				"OTEL_SERVICE_NAME":    "custom-service",
			},
			expectServiceName: "custom-service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			var stdout bytes.Buffer
			result, err := NewTracingFromEnv(t.Context(), &stdout)
			require.NoError(t, err)
			t.Cleanup(func() {
				_ = result.Shutdown(t.Context())
			})

			// Start a span to trigger output.
			span := startCompletionsSpan(t, result, nil)
			require.NotNil(t, span)
			span.EndSpan()

			// Check that the service name appears in the console output.
			output := stdout.String()
			require.Contains(t, output, `"service.name"`)
			require.Contains(t, output, tt.expectServiceName)
		})
	}
}

func TestNewTracingFromEnv_DisabledByEnv(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
	}{
		{
			name: "OTEL_SDK_DISABLED true",
			env: map[string]string{
				"OTEL_SDK_DISABLED":           "true",
				"OTEL_EXPORTER_OTLP_ENDPOINT": "http://localhost:4318", // Should be ignored.
			},
		},
		{
			name: "OTEL_TRACES_EXPORTER none",
			env: map[string]string{
				"OTEL_TRACES_EXPORTER":        "none",
				"OTEL_EXPORTER_OTLP_ENDPOINT": "http://localhost:4318", // Should be ignored.
			},
		},
		{
			name: "no endpoints or exporters configured",
			env:  map[string]string{},
		},
		{
			name: "no traces endpoint when only metrics endpoint is configured",
			env: map[string]string{
				"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT": "http://localhost:4318",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			result, err := NewTracingFromEnv(t.Context(), io.Discard)
			require.NoError(t, err)
			require.IsType(t, tracing.NoopTracing{}, result)
		})
	}
}

// TestNewTracingFromEnv_EndpointHierarchy tests the OTEL endpoint hierarchy.
// according to the OTEL spec where signal-specific endpoints override generic ones.
func TestNewTracingFromEnv_EndpointHierarchy(t *testing.T) {
	tests := []struct {
		name         string
		env          map[string]string
		expectActive bool
	}{
		{
			name: "uses generic OTLP endpoint when configured",
			env: map[string]string{
				"OTEL_EXPORTER_OTLP_ENDPOINT": "http://localhost:4317",
			},
			expectActive: true,
		},
		{
			name: "uses traces-specific endpoint when configured",
			env: map[string]string{
				"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT": "http://localhost:4318",
			},
			expectActive: true,
		},
		{
			name: "traces-specific endpoint overrides generic",
			env: map[string]string{
				"OTEL_EXPORTER_OTLP_ENDPOINT":        "http://localhost:4317",
				"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT": "http://localhost:4318",
			},
			expectActive: true,
		},
		{
			name: "explicit exporter overrides endpoint detection",
			env: map[string]string{
				"OTEL_EXPORTER_OTLP_ENDPOINT": "http://localhost:4317",
				"OTEL_TRACES_EXPORTER":        "console",
			},
			expectActive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			result, err := NewTracingFromEnv(t.Context(), io.Discard)
			require.NoError(t, err)

			if tt.expectActive {
				_, isNoop := result.(tracing.NoopTracing)
				require.False(t, isNoop, "expected active tracing")
			} else {
				require.IsType(t, tracing.NoopTracing{}, result)
			}

			_ = result.Shutdown(context.Background())
		})
	}
}

// TestNewTracingFromEnv_ConsoleExporter tests that console exporter works.
// without requiring OTLP endpoints and doesn't make network calls.
func TestNewTracingFromEnv_ConsoleExporter(t *testing.T) {
	tests := []struct {
		name                string
		env                 map[string]string
		expectNoop          bool
		expectConsoleOutput bool
	}{
		{
			name: "console exporter without any endpoints",
			env: map[string]string{
				"OTEL_TRACES_EXPORTER": "console",
			},
			expectConsoleOutput: true,
		},
		{
			name: "console exporter ignores OTLP endpoints",
			env: map[string]string{
				"OTEL_TRACES_EXPORTER":               "console",
				"OTEL_EXPORTER_OTLP_ENDPOINT":        "http://should-be-ignored:4317",
				"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT": "http://should-be-ignored:4318",
			},
			expectConsoleOutput: true,
		},
		{
			name: "console exporter with custom service name",
			env: map[string]string{
				"OTEL_TRACES_EXPORTER": "console",
				"OTEL_SERVICE_NAME":    "test-console-service",
			},
			expectConsoleOutput: true,
		},
		{
			name: "console exporter with sampling",
			env: map[string]string{
				"OTEL_TRACES_EXPORTER": "console",
				"OTEL_TRACES_SAMPLER":  "always_on",
			},
			expectConsoleOutput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			var stdout bytes.Buffer
			result, err := NewTracingFromEnv(t.Context(), &stdout)
			require.NoError(t, err)
			t.Cleanup(func() {
				_ = result.Shutdown(context.Background())
			})

			if tt.expectNoop {
				_, ok := result.(tracing.NoopTracing)
				require.True(t, ok, "expected NoopTracing")
				return
			}

			// Verify it's not noop.
			_, ok := result.(tracing.NoopTracing)
			require.False(t, ok, "expected non-noop tracing")

			// For console exporter, create a span and verify output.
			if tt.expectConsoleOutput {
				span := startCompletionsSpan(t, result, nil)
				require.NotNil(t, span, "expected span to be created")
				span.EndSpan()

				// Console exporter writes synchronously, so output should be immediate.
				output := stdout.String()
				require.Contains(t, output, "TraceID", "console output should contain TraceID")
				require.Contains(t, output, "SpanID", "console output should contain SpanID")

				// Verify service name if set.
				if serviceName := tt.env["OTEL_SERVICE_NAME"]; serviceName != "" {
					require.Contains(t, output, serviceName, "console output should contain custom service name")
				}
			}
		})
	}
}

// TestNewTracingFromEnv_Exporter tests that the OTEL_TRACES_EXPORTER env
// variable works.
// See: https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_traces_exporter
func TestNewTracingFromEnv_Exporter(t *testing.T) {
	// Just test 2 exporters to prove the SDK is wired up correctly.
	for _, exporter := range []string{"console", "otlp"} {
		t.Run(exporter, func(t *testing.T) {
			t.Setenv("OTEL_TRACES_EXPORTER", exporter)

			var stdout bytes.Buffer
			collector, tracing := newTracingFromEnvForTest(t, &stdout)

			// Create a test request to start a span.
			span := startCompletionsSpan(t, tracing, nil)
			require.NotNil(t, span, "expected span to be sampled")
			span.EndSpan()

			// Now, verify the actual ENV were honored.
			v1Span := collector.TakeSpan()
			switch exporter {
			case "otlp":
				require.NotNil(t, v1Span)
				require.Empty(t, stdout)
			case "console":
				require.Nil(t, v1Span)
				require.Contains(t, stdout.String(), "TraceID")
			}
		})
	}
}

// TestNewTracingFromEnv_TracesSampler tests that the OTEL_TRACES_SAMPLER env
// variable works.
// See: https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_traces_sampler
func TestNewTracingFromEnv_TracesSampler(t *testing.T) {
	// Just test 2 samplers to prove the SDK is wired up correctly.
	tests := []struct {
		sampler       string
		expectSampled bool
	}{
		{"always_on", true},
		{"always_off", false},
	}

	for _, tt := range tests {
		t.Run(tt.sampler, func(t *testing.T) {
			t.Setenv("OTEL_TRACES_SAMPLER", tt.sampler)
			collector, tracing := newTracingFromEnvForTest(t, io.Discard)

			span := startCompletionsSpan(t, tracing, nil)
			if tt.expectSampled {
				require.NotNil(t, span, "expected span to be sampled")
				span.EndSpan()
			} else {
				require.Nil(t, span, "expected span to not be sampled")
			}

			// Now, verify the actual ENV were honored.
			v1Span := collector.TakeSpan()
			if tt.expectSampled {
				require.NotNil(t, v1Span)
			} else {
				require.Nil(t, v1Span)
			}
		})
	}
}

// TestNewTracingFromEnv_OtelPropagators tests that the OTEL_PROPAGATORS env
// variable works.
// See: https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_propagators
func TestNewTracingFromEnv_OtelPropagators(t *testing.T) {
	// Just test 2 propagators to prove the SDK is wired up correctly.
	tests := []struct {
		propagator         string
		expectHeaderKey    string
		expectHeaderFormat func(string, string) string
	}{
		{
			propagator:         "b3",
			expectHeaderKey:    "b3",
			expectHeaderFormat: func(traceID, spanID string) string { return traceID + "-" + spanID + "-1" },
		},
		{
			propagator:         "tracecontext",
			expectHeaderKey:    "traceparent",
			expectHeaderFormat: func(traceID, spanID string) string { return "00-" + traceID + "-" + spanID + "-01" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.propagator, func(t *testing.T) {
			t.Setenv("OTEL_PROPAGATORS", tt.propagator)
			collector, tracing := newTracingFromEnvForTest(t, io.Discard)

			// Create headers for injection.
			headerMutation := &extprocv3.HeaderMutation{}

			// Start span and inject headers.
			span := startCompletionsSpan(t, tracing, headerMutation)
			require.NotNil(t, span)
			span.EndSpan()

			// Check that the expected header was injected.
			require.Len(t, headerMutation.SetHeaders, 1, "expected exactly one header to be set")
			header := headerMutation.SetHeaders[0].Header
			require.Equal(t, tt.expectHeaderKey, header.Key)

			// Get the span to check trace/span IDs.
			v1Span := collector.TakeSpan()
			require.NotNil(t, v1Span)

			// Convert IDs to hex strings.
			traceIDStr := fmt.Sprintf("%032x", v1Span.TraceId)
			spanIDStr := fmt.Sprintf("%016x", v1Span.SpanId)

			// Verify the header value format.
			expectedValue := tt.expectHeaderFormat(traceIDStr, spanIDStr)
			require.Equal(t, expectedValue, string(header.RawValue))
		})
	}
}

// TestNewTracingFromEnv_OpenInferenceRedaction tests that the OpenInference.
// environment variables (OPENINFERENCE_HIDE_INPUTS and OPENINFERENCE_HIDE_OUTPUTS)
// work correctly to redact sensitive data from spans, following the OpenInference.
// configuration specification.
func TestNewTracingFromEnv_OpenInferenceRedaction(t *testing.T) {
	tests := []struct {
		name        string
		hideInputs  bool
		hideOutputs bool
	}{
		{
			name:        "no redaction",
			hideInputs:  false,
			hideOutputs: false,
		},
		{
			name:        "hide inputs only",
			hideInputs:  true,
			hideOutputs: false,
		},
		{
			name:        "hide outputs only",
			hideInputs:  false,
			hideOutputs: true,
		},
		{
			name:        "hide inputs and outputs",
			hideInputs:  true,
			hideOutputs: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(openinference.EnvHideInputs, strconv.FormatBool(tt.hideInputs))
			t.Setenv(openinference.EnvHideOutputs, strconv.FormatBool(tt.hideOutputs))

			collector, tr := newTracingFromEnvForTest(t, io.Discard)
			tracer := tr.ChatCompletionTracer()

			// Create a test request with sensitive data.
			req := &openai.ChatCompletionRequest{
				Model: openai.ModelGPT5Nano,
				Messages: []openai.ChatCompletionMessageParamUnion{{
					OfUser: &openai.ChatCompletionUserMessageParam{
						Content: openai.StringOrUserRoleContentUnion{
							Value: "Hello, sensitive data!",
						},
						Role: openai.ChatMessageRoleUser,
					},
				}},
			}
			reqBody := []byte(`{"model":"gpt-4.1-nano","messages":[{"role":"user","content":"Hello, sensitive data!"}]}`)
			respBody := &openai.ChatCompletionResponse{
				ID:     "chatcmpl-abc123",
				Object: "chat.completion",
				Choices: []openai.ChatCompletionResponseChoice{{
					Message: openai.ChatCompletionResponseChoiceMessage{
						Role:    "assistant",
						Content: ptr.To("Response with sensitive data"),
					},
				}},
			}

			// Start a span and record request/response.
			span := tracer.StartSpanAndInjectHeaders(
				t.Context(),
				map[string]string{},
				&extprocv3.HeaderMutation{},
				req,
				reqBody,
			)
			require.NotNil(t, span)
			span.RecordResponse(respBody)
			span.EndSpan()

			// Check the recorded span.
			v1Span := collector.TakeSpan()
			require.NotNil(t, v1Span)

			// Check if inputs/outputs are redacted as expected.
			attrs := make(map[string]string)
			for _, kv := range v1Span.Attributes {
				attrs[kv.Key] = kv.Value.GetStringValue()
			}

			// Check input redaction.
			if tt.hideInputs {
				require.Equal(t, openinference.RedactedValue, attrs[openinference.InputValue])
				_, hasInputMessage := attrs[openinference.InputMessageAttribute(0, openinference.MessageContent)]
				require.False(t, hasInputMessage)
			} else {
				require.Contains(t, attrs[openinference.InputValue], "Hello, sensitive data!")
				require.Contains(t, attrs[openinference.InputValue], attrs[openinference.InputMessageAttribute(0, openinference.MessageContent)])
			}

			// Check output redaction.
			if tt.hideOutputs {
				require.Equal(t, openinference.RedactedValue, attrs[openinference.OutputValue])
				_, hasOutputMessage := attrs[openinference.OutputMessageAttribute(0, openinference.MessageContent)]
				require.False(t, hasOutputMessage)
			} else {
				require.Contains(t, attrs[openinference.OutputValue], "Response with sensitive data")
				require.Contains(t, attrs[openinference.OutputValue], attrs[openinference.OutputMessageAttribute(0, openinference.MessageContent)])
			}
		})
	}
}

func newTracingFromEnvForTest(t *testing.T, stdout io.Writer) (*testotel.OTLPCollector, tracing.Tracing) {
	collector := testotel.StartOTLPCollector()
	t.Cleanup(collector.Close)
	collector.SetEnv(t.Setenv)

	result, err := NewTracingFromEnv(t.Context(), stdout)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = result.Shutdown(context.Background())
	})

	return collector, result
}

func TestNewTracing(t *testing.T) {
	t.Run("with noop tracer", func(t *testing.T) {
		config := &tracing.TracingConfig{
			Tracer:                 noop.Tracer{},
			Propagator:             autoprop.NewTextMapPropagator(),
			ChatCompletionRecorder: openaitracing.NewChatCompletionRecorderFromEnv(),
		}

		result := NewTracing(config)
		require.IsType(t, tracing.NoopTracing{}, result)
	})

	t.Run("with real tracer", func(t *testing.T) {
		tp := trace.NewTracerProvider()
		t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

		config := &tracing.TracingConfig{
			Tracer:                 tp.Tracer("test"),
			Propagator:             autoprop.NewTextMapPropagator(),
			ChatCompletionRecorder: openaitracing.NewChatCompletionRecorderFromEnv(),
		}

		result := NewTracing(config)
		require.IsType(t, &tracingImpl{}, result)

		// Test that ChatCompletionTracer returns the expected tracer.
		tracer := result.ChatCompletionTracer()
		require.NotNil(t, tracer)

		// Test that Shutdown returns nil when tp wasn't created internally.
		err := result.Shutdown(t.Context())
		require.NoError(t, err)
	})
}

func TestNoopShutdown(t *testing.T) {
	ns := noopShutdown{}
	err := ns.Shutdown(t.Context())
	require.NoError(t, err)
}

// TestNewTracingFromEnv_OTLPHeaders tests that OTEL_EXPORTER_OTLP_HEADERS
// is properly handled by the autoexport package.
func TestNewTracingFromEnv_OTLPHeaders(t *testing.T) {
	expectedAuthorization := "ApiKey test-key-123"
	actualAuthorization := make(chan string, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualAuthorization <- r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	t.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "Authorization="+expectedAuthorization)
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", ts.URL)
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")

	result, err := NewTracingFromEnv(t.Context(), io.Discard)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = result.Shutdown(context.Background())
	})

	// Create span to trigger export
	span := startCompletionsSpan(t, result, nil)
	require.NotNil(t, span)
	span.EndSpan()

	// Force flush
	if impl, ok := result.(*tracingImpl); ok {
		_ = impl.shutdown(t.Context())
	}

	require.Equal(t, expectedAuthorization, <-actualAuthorization)
}

// startCompletionsSpan is a test helper that creates a span with a basic request.
// If headerMutation is nil, a new empty HeaderMutation will be created.
func startCompletionsSpan(t *testing.T, tracing tracing.Tracing, headerMutation *extprocv3.HeaderMutation) tracing.ChatCompletionSpan {
	if headerMutation == nil {
		headerMutation = &extprocv3.HeaderMutation{}
	}
	tracer := tracing.ChatCompletionTracer()
	req := &openai.ChatCompletionRequest{Model: openai.ModelGPT5Nano}
	return tracer.StartSpanAndInjectHeaders(t.Context(), nil, headerMutation, req, nil)
}
