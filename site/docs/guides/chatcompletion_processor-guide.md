# Response Processing and Gzip Handling Guide

This guide covers how the AI Gateway handles response processing from upstream AI providers, with a focus on the gzip compression handling improvements.

## Overview

The AI Gateway processes responses from various AI providers to extract metrics, token usage, and other relevant information. This processing needs to handle compressed responses correctly to ensure accurate data extraction.

## Response Processing Architecture

The AI Gateway uses specialized processors for different types of AI endpoints:

- **Chat Completion Processor** - Handles `/v1/chat/completions` endpoints
- **Messages Processor** - Handles `/v1/messages` endpoints (Anthropic-style)
- **Embeddings Processor** - Handles `/v1/embeddings` endpoints

## Gzip Handling

### Background

Many AI providers return gzip-compressed responses to reduce bandwidth usage. The AI Gateway must decompress these responses before extracting token usage and other metrics.

### Implementation

All response processors now use a common utility function for handling gzipped content:

```go
func decompressGzipIfNeeded(body []byte, headers map[string]string) ([]byte, error) {
    // Check if Content-Encoding indicates gzip compression
    if contentEncoding, exists := headers["content-encoding"]; exists && contentEncoding == "gzip" {
        // Decompress the gzipped content
        reader, err := gzip.NewReader(bytes.NewReader(body))
        if err != nil {
            return nil, err
        }
        defer reader.Close()
        
        return io.ReadAll(reader)
    }
    return body, nil
}
```

### Affected Processors

#### Chat Completion Processor

Handles OpenAI-compatible chat completion responses:

```go
// Before processing, decompress if needed
decompressedBody, err := decompressGzipIfNeeded(body, headers)
if err != nil {
    return err
}

// Extract token usage from decompressed response
var response ChatCompletionResponse
if err := json.Unmarshal(decompressedBody, &response); err != nil {
    return err
}
```

#### Messages Processor

Handles Anthropic-style message responses:

```go
// Decompress response before token extraction
decompressedBody, err := decompressGzipIfNeeded(body, headers)
if err != nil {
    return err
}

// Process decompressed content for token metrics
var response MessagesResponse
if err := json.Unmarshal(decompressedBody, &response); err != nil {
    return err
}
```

#### Embeddings Processor

Handles embedding generation responses:

```go
// Handle gzipped embeddings responses
decompressedBody, err := decompressGzipIfNeeded(body, headers)
if err != nil {
    return err
}

// Extract embedding metrics from decompressed data
var response EmbeddingsResponse
if err := json.Unmarshal(decompressedBody, &response); err != nil {
    return err
}
```

## Benefits

### Accurate Token Counting

With proper gzip handling, the AI Gateway can now accurately extract token usage from compressed responses, ensuring:

- Correct billing metrics
- Accurate rate limiting based on token consumption
- Proper observability data

### Performance Impact

The gzip decompression adds minimal overhead while ensuring data accuracy:

- Decompression only occurs when `Content-Encoding: gzip` is present
- Processing continues normally for uncompressed responses
- Memory usage is optimized through streaming decompression

## Troubleshooting

### Common Issues

1. **Missing Token Metrics**: If token usage is not being recorded, check if the upstream provider is sending gzipped responses
2. **Processing Errors**: Ensure the response format matches the expected schema after decompression
3. **Performance Issues**: Monitor memory usage if processing very large compressed responses

### Debugging

Enable debug logging to see response processing details:

```yaml
logging:
  level: debug
```

Look for log entries related to response decompression and token extraction in the processor logs.

## Best Practices

1. **Monitor Metrics**: Track token usage metrics to ensure accurate extraction
2. **Test with Providers**: Verify that different AI providers' compressed responses are handled correctly
3. **Resource Monitoring**: Monitor memory usage when processing large compressed responses
4. **Error Handling**: Implement proper error handling for decompression failures