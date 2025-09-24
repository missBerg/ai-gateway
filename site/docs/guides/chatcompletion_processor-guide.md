# Response Processing with Gzip Handling

This guide explains how Envoy AI Gateway handles gzipped HTTP responses when processing requests to extract token usage information and perform response transformations.

## Overview

The AI Gateway processes HTTP responses from various LLM providers to extract token usage metrics and transform responses between different API schemas. When upstream providers return gzipped responses, the gateway must properly decompress the content before processing to ensure accurate token usage extraction and response transformations.

## Problem Addressed

Prior to this improvement, token usage was not correctly extracted from gzipped HTTP responses for certain endpoints (specifically `v1/messages`). This occurred because the response processors attempted to parse compressed response bodies without first decompressing them, leading to:

- Inaccurate or missing token usage metrics
- Failed response transformations
- Incorrect billing and cost calculations

## Implementation

The gateway now includes centralized utility functions for handling gzipped content across all response processors:

### Core Utility Functions

The following utility functions are available in `internal/extproc/util.go`:

```go
// decodeContentIfNeeded decompresses response body based on content-encoding header
func decodeContentIfNeeded(body []byte, contentEncoding string) (contentDecodingResult, error)

// removeContentEncodingIfNeeded removes content-encoding header when body is modified
func removeContentEncodingIfNeeded(headerMutation *extprocv3.HeaderMutation, 
    bodyMutation *extprocv3.BodyMutation, isEncoded bool) *extprocv3.HeaderMutation
```

### Affected Processors

The gzip handling has been implemented across three main processors:

1. **Chat Completion Processor** (`chatcompletion_processor.go`)
2. **Embeddings Processor** (`embeddings_processor.go`) 
3. **Messages Processor** (`messages_processor.go`)

## Usage Examples

### Processing Response Bodies

All processors now follow this pattern when handling response bodies:

```go
func (c *processor) ProcessResponseBody(ctx context.Context, body *extprocv3.HttpBody) (*extprocv3.ProcessingResponse, error) {
    // Decompress the body if needed using common utility
    decodingResult, err := decodeContentIfNeeded(body.Body, c.responseEncoding)
    if err != nil {
        return nil, err
    }

    // Use the decompressed reader for processing
    headerMutation, bodyMutation, tokenUsage, err := c.translator.ResponseBody(
        c.responseHeaders, decodingResult.reader, body.EndOfStream)
    if err != nil {
        return nil, fmt.Errorf("failed to transform response: %w", err)
    }

    // Remove content-encoding header if body was modified and originally encoded
    headerMutation = removeContentEncodingIfNeeded(headerMutation, bodyMutation, decodingResult.isEncoded)

    // Continue with response processing...
}
```

### Content Encoding Detection

The processors detect gzipped content during response header processing:

```go
func (c *processor) ProcessResponseHeaders(ctx context.Context, headers *corev3.HeaderMap) (*extprocv3.ProcessingResponse, error) {
    c.responseHeaders = headersToMap(headers)
    if enc := c.responseHeaders["content-encoding"]; enc != "" {
        c.responseEncoding = enc
    }
    // Continue processing...
}
```

## Supported Encodings

Currently, the gateway supports:

- **gzip**: Full decompression and recompression handling
- **identity** (no encoding): Pass-through processing

The implementation is designed to be extensible for additional encoding formats in the future.

## Header Management

When the gateway modifies a response body that was originally gzipped:

1. The response is decompressed for processing
2. Token usage and transformations are applied to the decompressed content
3. The `content-encoding` header is removed from the response
4. The modified response is sent uncompressed to the client

> **Note**: The current implementation removes the `content-encoding` header rather than recompressing the modified content. This is noted as a potential future improvement where responses could be recompressed to maintain the original encoding.

## Benefits

This implementation provides:

- **Accurate Token Metrics**: Token usage is now correctly extracted from all response types, regardless of compression
- **Consistent Processing**: All endpoint processors handle compression uniformly
- **Transparent Operation**: Clients receive properly formatted responses without needing to handle decompression concerns
- **Extensible Design**: The utility functions can be easily extended to support additional compression formats

## Configuration

No additional configuration is required. The gzip handling is automatically applied based on the `content-encoding` header present in upstream responses.