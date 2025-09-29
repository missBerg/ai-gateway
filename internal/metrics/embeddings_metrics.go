// Copyright Envoy AI Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/envoyproxy/ai-gateway/internal/filterapi"
	"github.com/envoyproxy/ai-gateway/internal/internalapi"
)

// embeddings is the implementation for the embeddings AI Gateway metrics.
type embeddings struct {
	baseMetrics
}

// EmbeddingsMetrics is the interface for the embeddings AI Gateway metrics.
type EmbeddingsMetrics interface {
	// StartRequest initializes timing for a new request.
	StartRequest(headers map[string]string)
	// SetRequestModel sets the model from the request. This is usually called after parsing the request body.
	// Example: text-embedding-3-small
	SetRequestModel(requestModel internalapi.RequestModel)
	// SetResponseModel sets the model that ultimately generated the response.
	// Example: text-embedding-3-small-2025-02-18
	SetResponseModel(responseModel internalapi.ResponseModel)
	// SetBackend sets the selected backend when the routing decision has been made. This is usually called
	// after parsing the request body to determine the model and invoke the routing logic.
	SetBackend(backend *filterapi.Backend)

	// RecordTokenUsage records token usage metrics for embeddings (only input tokens are relevant).
	RecordTokenUsage(ctx context.Context, inputTokens uint32, requestHeaderLabelMapping map[string]string)
	// RecordRequestCompletion records latency metrics for the entire request.
	RecordRequestCompletion(ctx context.Context, success bool, requestHeaderLabelMapping map[string]string)
}

// NewEmbeddings creates a new Embeddings instance.
func NewEmbeddings(meter metric.Meter, requestHeaderLabelMapping map[string]string) EmbeddingsMetrics {
	return &embeddings{
		baseMetrics: newBaseMetrics(meter, genaiOperationEmbedding, requestHeaderLabelMapping),
	}
}

// RecordTokenUsage implements [EmbeddingsMetrics.RecordTokenUsage].
func (e *embeddings) RecordTokenUsage(ctx context.Context, inputTokens uint32, requestHeaders map[string]string) {
	attrs := e.buildBaseAttributes(requestHeaders)

	// Embeddings only consume input tokens to generate vector representations.
	e.metrics.tokenUsage.Record(ctx, float64(inputTokens),
		metric.WithAttributeSet(attrs),
		metric.WithAttributes(attribute.Key(genaiAttributeTokenType).String(genaiTokenTypeInput)),
	)
}
