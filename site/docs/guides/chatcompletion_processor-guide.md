# Gzip Response Processing Guide

This guide explains how the AI Gateway handles gzip-compressed HTTP responses from upstream providers to ensure proper token usage extraction and response processing.

## Overview

The AI Gateway now includes unified utility functions for handling gzipped HTTP responses across all processor types. This enhancement ensures that token usage is correctly extracted from compressed responses, preventing metrics gaps and improving observability.

## Background

Previously, different processors handled gzip decompression independently, leading to inconsistencies and bugs where token usage couldn't be extracted from compressed responses. The streamlined approach centralizes gzip handling and applies it consistently across:

- Chat completion processor (`/v1/chat/completions`)
- Messages processor (`/v1/messages`) 
- Embeddings processor (`/v1/embeddings`)

## Key Components

### Content Decoding Utility

The gateway provides a common utility function `decodeContentIfNeeded()` that:

- Detects content encoding from response headers
- Automatically decompresses gzipped content
- Returns metadata about encoding status
- Supports extensible encoding types

```go
// Example usage in response processing
decodingResult, err := decodeContentIfNeeded(body.Body, c.responseEncoding)
if err != nil {
    return nil, err
}
```

### Header Management

When content is decompressed for processing, the gateway automatically:

- Removes the `content-encoding` header if the response body was modified
- Prevents mismatched headers that could cause client issues
- Maintains transparency for unmodified responses

## Implementation Details

### Response Processing Flow

1. **Header Processing**: Extract `content-encoding` header value
2. **Body Processing**: 
   - Decompress body content if gzipped
   - Process response for token extraction and transformations
   - Remove encoding headers if body was modified

### Supported Encodings

Currently supported compression formats:
- **gzip**: Full support for decompression and processing
- **identity/none**: Pass-through for uncompressed content

Future encoding support can be easily added by extending the utility functions.

### Processor Integration

All processors now use the standardized approach:

```go
// Common pattern across all processors
decodingResult, err := decodeContentIfNeeded(body.Body, c.responseEncoding)
if err != nil {
    return nil, err
}

// Process using decompressed content
headerMutation, bodyMutation, tokenUsage, err := c.translator.ResponseBody(
    c.responseHeaders, 
    decodingResult.reader, 
    body.EndOfStream
)

// Clean up headers if needed
headerMutation = removeContentEncodingIfNeeded(
    headerMutation, 
    bodyMutation, 
    decodingResult.isEncoded
)
```

## Benefits

### Consistent Token Extraction

- Token usage metrics are now accurately captured from gzipped responses
- Eliminates gaps in observability data
- Ensures billing and rate limiting work correctly

### Unified Codebase

- Reduces code duplication across processors
- Simplifies maintenance and testing
- Makes adding new encoding support easier

### Transparent Operation

- Clients receive properly formatted responses
- No impact on existing functionality
- Backward compatible with all configurations

## Configuration

No configuration changes are required. The gzip handling is automatically applied based on the `content-encoding` header from upstream responses.

## Troubleshooting

### Missing Token Usage Metrics

If you're not seeing token usage data for certain providers:

1. Verify the provider sends gzipped responses with proper headers
2. Check that the AI Gateway is processing the correct endpoint
3. Review logs for decompression errors

### Response Format Issues

If clients report malformed responses:

1. Ensure the upstream provider is sending valid gzipped content
2. Check for mismatched `content-encoding` headers in responses
3. Verify processor transformations aren't corrupting content

## Monitoring

The gzip handling integrates with existing observability:

- Token usage metrics include data from compressed responses
- Request/response latency measurements account for decompression time
- Error rates capture decompression failures

Monitor these metrics to ensure optimal performance and catch potential issues early.