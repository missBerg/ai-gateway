# Response Processing: Gzip Content Handling

This document describes improvements made to the AI Gateway's response processing system to ensure consistent handling of compressed HTTP responses across all endpoints.

## Overview

The AI Gateway processes responses from various AI service providers to extract token usage information and perform other transformations. A bug was discovered where gzipped responses from upstream services weren't being properly decompressed before processing, leading to incorrect or missing token usage extraction for certain endpoints.

## The Problem

Prior to this change, the AI Gateway had inconsistent handling of compressed responses:

- **Chat Completion endpoint** (`/v1/chat/completions`): Properly handled gzip decompression
- **Messages endpoint** (`/v1/messages`): Missing gzip handling, causing token usage extraction failures
- **Embeddings endpoint**: Missing gzip handling

When upstream services returned gzipped responses, the processors would attempt to parse the compressed binary data as JSON, resulting in:
- Failed token usage extraction
- Incomplete metrics collection
- Potential processing errors

## The Solution

A unified approach was implemented using shared utility functions to handle content decompression consistently across all response processors.

### New Utility Functions

The following utility functions were added to `internal/extproc/util.go`:

#### `decodeContentIfNeeded`

```go
func decodeContentIfNeeded(body []byte, contentEncoding string) (contentDecodingResult, error)
```

This function:
- Checks the `content-encoding` header value
- Decompresses gzipped content when needed
- Returns a reader for the (potentially decompressed) body
- Provides metadata about whether the content was encoded

#### `removeContentEncodingIfNeeded`

```go
func removeContentEncodingIfNeeded(headerMutation *extprocv3.HeaderMutation, bodyMutation *extprocv3.BodyMutation, isEncoded bool) *extprocv3.HeaderMutation
```

This function:
- Removes the `content-encoding` header when the response body is modified
- Prevents client-side decompression errors when the gateway has already decompressed content
- Only removes the header when body mutations occur

### Implementation Details

The solution involves these key changes:

1. **Centralized decompression logic**: All processors now use the same decompression utilities
2. **Header management**: Automatic removal of encoding headers when content is modified
3. **Error handling**: Proper error reporting for decompression failures
4. **Extensible design**: Easy to add support for additional compression formats

### Usage Pattern

All response processors now follow this pattern:

```go
// Decompress the body if needed using common utility
decodingResult, err := decodeContentIfNeeded(body.Body, c.responseEncoding)
if err != nil {
    return nil, err
}

// Process the decompressed content
headerMutation, bodyMutation, tokenUsage, err := c.translator.ResponseBody(
    c.responseHeaders, 
    decodingResult.reader,  // Use decompressed reader
    body.EndOfStream
)
if err != nil {
    return nil, fmt.Errorf("failed to transform response: %w", err)
}

// Remove content-encoding header if body was mutated and originally encoded
headerMutation = removeContentEncodingIfNeeded(headerMutation, bodyMutation, decodingResult.isEncoded)
```

## Benefits

This implementation provides:

1. **Consistency**: All endpoints handle compressed responses uniformly
2. **Reliability**: Token usage extraction works correctly regardless of compression
3. **Maintainability**: Centralized compression handling reduces code duplication
4. **Extensibility**: Easy to add support for additional compression formats (brotli, deflate, etc.)
5. **Performance**: Only decompresses when necessary

## Supported Compression Formats

Currently supported:
- **gzip**: Full support for gzip-compressed responses
- **identity** (no compression): Handled transparently

Future expansion can easily add support for:
- brotli
- deflate
- Other standard HTTP compression formats

## Configuration

No configuration changes are required. The compression handling is automatic based on the `content-encoding` header in upstream responses.

## Monitoring

The fix ensures that:
- Token usage metrics are accurately collected for all response types
- Observability data includes correct token counts
- Cost tracking works properly regardless of response compression

This improvement ensures reliable operation of the AI Gateway's response processing pipeline across all supported AI service endpoints.