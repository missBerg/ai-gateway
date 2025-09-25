// Copyright Envoy AI Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/envoyproxy/ai-gateway/internal/filterapi"
	"github.com/envoyproxy/ai-gateway/internal/internalapi"
)

// baseMetrics provides shared functionality for AI Gateway metrics implementations.
type baseMetrics struct {
	metrics      *genAI
	operation    string
	requestStart time.Time
	// requestModel is the original model from the request body.
	requestModel string
	// responseModel is the model that ultimately generated the response (may differ due to backend override).
	responseModel             string
	backend                   string
	requestHeaderLabelMapping map[string]string // maps HTTP headers to metric label names.
}

// newBaseMetrics creates a new baseMetrics instance with the specified operation.
func newBaseMetrics(meter metric.Meter, operation string, requestHeaderLabelMapping map[string]string) baseMetrics {
	return baseMetrics{
		metrics:                   newGenAI(meter),
		operation:                 operation,
		requestModel:              "unknown",
		responseModel:             "unknown",
		backend:                   "unknown",
		requestHeaderLabelMapping: requestHeaderLabelMapping,
	}
}

// StartRequest initializes timing for a new request.
func (b *baseMetrics) StartRequest(_ map[string]string) {
	b.requestStart = time.Now()
}

// SetRequestModel sets the model the request. This is usually called after parsing the request body. e.g. gpt-5-nano
func (b *baseMetrics) SetRequestModel(requestModel internalapi.RequestModel) {
	b.requestModel = requestModel
}

// SetResponseModel is the model that ultimately generated the response. e.g. gpt-5-nano-2025-08-07
func (b *baseMetrics) SetResponseModel(responseModel internalapi.ResponseModel) {
	b.responseModel = responseModel
}

// SetBackend sets the name of the backend to be reported in the metrics according to:
// https://opentelemetry.io/docs/specs/semconv/gen-ai/gen-ai-metrics/
func (b *baseMetrics) SetBackend(backend *filterapi.Backend) {
	switch backend.Schema.Name {
	case filterapi.APISchemaOpenAI:
		b.backend = genaiProviderOpenAI
	case filterapi.APISchemaAWSBedrock:
		b.backend = genaiProviderAWSBedrock
	default:
		b.backend = backend.Name
	}
}

// buildBaseAttributes creates the base attributes for metrics recording.
func (b *baseMetrics) buildBaseAttributes(headers map[string]string) attribute.Set {
	opt := attribute.Key(genaiAttributeOperationName).String(b.operation)
	provider := attribute.Key(genaiAttributeProviderName).String(b.backend)
	reqModel := attribute.Key(genaiAttributeRequestModel).String(b.requestModel)
	respModel := attribute.Key(genaiAttributeResponseModel).String(b.responseModel)
	if len(b.requestHeaderLabelMapping) == 0 {
		return attribute.NewSet(opt, provider, reqModel, respModel)
	}

	// Add header values as attributes based on the header mapping if headers are provided.
	attrs := []attribute.KeyValue{opt, provider, reqModel, respModel}
	for headerName, labelName := range b.requestHeaderLabelMapping {
		if headerValue, exists := headers[headerName]; exists {
			attrs = append(attrs, attribute.Key(labelName).String(headerValue))
		}
	}
	return attribute.NewSet(attrs...)
}

// RecordRequestCompletion records the completion of a request with success/failure status.
func (b *baseMetrics) RecordRequestCompletion(ctx context.Context, success bool, requestHeaders map[string]string) {
	attrs := b.buildBaseAttributes(requestHeaders)

	if success {
		// According to the semantic conventions, the error attribute should not be added for successful operations.
		b.metrics.requestLatency.Record(ctx, time.Since(b.requestStart).Seconds(), metric.WithAttributeSet(attrs))
	} else {
		// We don't have a set of typed errors yet, or a set of low-cardinality values, so we can just set the value to the
		// placeholder one. See: https://opentelemetry.io/docs/specs/semconv/attributes-registry/error/#error-type
		b.metrics.requestLatency.Record(ctx, time.Since(b.requestStart).Seconds(),
			metric.WithAttributeSet(attrs),
			metric.WithAttributes(attribute.Key(genaiAttributeErrorType).String(genaiErrorTypeFallback)),
		)
	}
}
