// Copyright Envoy AI Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package vcr

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/ai-gateway/tests/internal/testopenai"
	"github.com/envoyproxy/ai-gateway/tests/internal/testopeninference"
)

// otelEmbeddingsMetricsTestCase defines the expected behavior for each cassette.
type otelEmbeddingsMetricsTestCase struct {
	cassette testopenai.Cassette
	isError  bool // whether this is an error response.
}

// buildOtelEmbeddingsMetricsTestCases returns all test cases with their expected behaviors.
func buildOtelEmbeddingsMetricsTestCases() []otelEmbeddingsMetricsTestCase {
	var cases []otelEmbeddingsMetricsTestCase
	for _, cassette := range testopenai.EmbeddingsCassettes() {
		tc := otelEmbeddingsMetricsTestCase{cassette: cassette}
		switch cassette {
		case testopenai.CassetteEmbeddingsBadRequest, testopenai.CassetteEmbeddingsUnknownModel:
			tc.isError = true
		}
		cases = append(cases, tc)
	}
	return cases
}

// TestOtelOpenAIEmbeddings_metrics tests that metrics are properly exported via OTLP for embeddings completion requests.
func TestOtelOpenAIEmbeddings_metrics(t *testing.T) {
	env := setupOtelTestEnvironment(t)
	listenerPort := env.EnvoyListenerPort()
	was5xx := false

	for _, tc := range buildOtelEmbeddingsMetricsTestCases() {
		if was5xx {
			return // rather than also failing subsequent tests, which confuses root cause.
		}

		t.Run(tc.cassette.String(), func(t *testing.T) {
			// Send request.
			req, err := testopenai.NewRequest(t.Context(), fmt.Sprintf("http://localhost:%d/v1", listenerPort), tc.cassette)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			failIf5xx(t, resp, &was5xx)

			// Always read the content.
			_, err = io.ReadAll(resp.Body)
			require.NoError(t, err)

			// TODO: we have no span yet for embeddings, so compare against OpenInference python
			// See https://github.com/envoyproxy/ai-gateway/issues/1085
			span, err := testopeninference.GetSpan(t.Context(), io.Discard, tc.cassette)
			require.NoError(t, err)

			// Collect all metrics within the timeout period.
			// Embeddings should have 2 metrics: token usage + request duration
			allMetrics := env.collector.TakeMetrics(2)
			metrics := requireScopeMetrics(t, allMetrics)

			// Get expected model names from span
			requestModel := getInvocationModel(span.Attributes, "embedding.invocation_parameters")
			responseModel := getSpanAttributeString(span.Attributes, "embedding.model_name")

			// Verify each metric in separate functions.
			verifyTokenUsageMetrics(t, "embeddings", metrics, span, requestModel, responseModel, tc.isError)
			verifyRequestDurationMetrics(t, "embeddings", metrics, span, requestModel, responseModel, tc.isError)
		})
	}
}

func TestOtelOpenAIEmbeddings_metrics_modelNameOverride(t *testing.T) {
	env := setupOtelTestEnvironment(t)
	listenerPort := env.EnvoyListenerPort()

	req, err := testopenai.NewRequest(t.Context(), fmt.Sprintf("http://localhost:%d/v1", listenerPort), testopenai.CassetteEmbeddingsBasic)
	require.NoError(t, err)
	// Set the x-test-backend which envoy.yaml routes to the openai-embeddings-override
	// backend in extproc.yaml. This backend overrides the model to text-embedding-3-small.
	req.Header.Set("x-test-backend", "openai-embeddings-override")
	originalModel := "text-embedding-3"
	replaceRequestModel(t, req, originalModel)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode, "Response body: %s", string(body))

	// TODO: we have no span yet for embeddings, so compare against OpenInference python
	// See https://github.com/envoyproxy/ai-gateway/issues/1085
	span, err := testopeninference.GetSpan(t.Context(), io.Discard, testopenai.CassetteEmbeddingsBasic)
	require.NoError(t, err)

	expectedCount := 2 // token usage + request duration
	allMetrics := env.collector.TakeMetrics(expectedCount)
	metrics := requireScopeMetrics(t, allMetrics)

	// Get expected model names from span
	requestModel := getInvocationModel(span.Attributes, "embedding.invocation_parameters")
	responseModel := getSpanAttributeString(span.Attributes, "embedding.model_name")

	verifyTokenUsageMetrics(t, "embeddings", metrics, span, requestModel, responseModel, false)
	verifyRequestDurationMetrics(t, "embeddings", metrics, span, requestModel, responseModel, false)
}
