# Gzip Response Handling in AI Gateway

The AI Gateway now includes improved handling for gzipped HTTP responses from upstream providers. This enhancement ensures proper token usage extraction and response processing for compressed responses across all supported API endpoints.

## Overview

Previously, the AI Gateway's `/v1/messages` endpoint processor couldn't properly extract token usage from gzipped responses, while other endpoints like `/v1/chat/completions` handled compression correctly. This inconsistency has been resolved by implementing centralized utility functions for gzip handling.

## What Changed

### New Utility Functions

The AI Gateway now includes common utility functions in `internal/extproc/util.go` for handling compressed responses:

- `decodeContentIfNeeded()` - Automatically detects and decompresses gzipped response bodies
- `removeContentEncodingIfNeeded()` - Removes content-encoding headers when response body is modified

### Affected Endpoints

The following processors now support gzip decompression:

- **Chat Completion processor** (`/v1/chat/completions`)
- **Messages processor** (`/v1/messages`) 
- **Embeddings processor** (`/v1/embeddings`)

## Technical Implementation

### Content Decoding

When processing upstream responses, the AI Gateway now:

1. **Detects compression** - Checks the `content-encoding` header
2. **Decompresses automatically** - Uses gzip decompression for `content-encoding: gzip`
3. **Processes normally** - Extracts token usage and performs response transformations
4. **Cleans up headers** - Removes encoding headers when response is modified

```go
// Example usage in processor
decodingResult, err := decodeContentIfNeeded(body.Body, responseEncoding)
if err != nil {
    return nil, err
}

// Process the decompressed content
headerMutation, bodyMutation, tokenUsage, err := translator.ResponseBody(
    responseHeaders, 
    decodingResult.reader,  // Use decompressed reader
    body.EndOfStream
)

// Remove encoding header if body was modified
headerMutation = removeContentEncodingIfNeeded(headerMutation, bodyMutation, decodingResult.isEncoded)
```

### Error Handling

The implementation includes proper error handling for:

- Invalid gzip data
- Compression format detection
- Reader cleanup

## Benefits

### Consistent Token Tracking

All endpoints now correctly extract token usage from compressed responses, ensuring:

- Accurate billing and cost tracking
- Proper metrics collection 
- Consistent observability across providers

### Performance Optimization

Upstream providers often compress responses to reduce bandwidth usage. The AI Gateway now properly handles these optimized responses without losing functionality.

### Provider Compatibility

Many AI providers (especially cloud-based ones) automatically compress responses. This enhancement ensures the AI Gateway works seamlessly with all providers regardless of their compression settings.

## Migration Notes

This change is **backward compatible**. No configuration changes are required:

- Existing deployments will automatically benefit from improved gzip handling
- No breaking changes to API contracts
- All existing functionality remains unchanged

## Troubleshooting

If you encounter issues with compressed responses:

1. **Check logs** - Look for gzip decompression errors in AI Gateway logs
2. **Verify headers** - Ensure upstream responses include proper `content-encoding` headers  
3. **Test uncompressed** - Try disabling compression at the provider level to isolate issues

The AI Gateway will gracefully fall back to processing uncompressed content if decompression fails.