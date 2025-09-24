# AWS Bedrock Reasoning Stream Support

This guide covers the new reasoning stream support for AWS Bedrock models, which enables streaming responses with reasoning traces for models that support this capability.

## Overview

AWS Bedrock reasoning stream support allows you to receive streaming responses that include the model's reasoning process. This feature is particularly useful for:

- Understanding how the model arrived at its conclusions
- Debugging complex reasoning chains
- Building applications that need to expose the model's thought process

## Supported Models

Currently, reasoning stream support is available for AWS Bedrock models that support reasoning capabilities, including:

- Claude models with reasoning features
- Other Bedrock models that provide reasoning traces

## Configuration

To enable reasoning stream support, configure your `AIServiceBackend` to use AWS Bedrock with the appropriate model:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIServiceBackend
metadata:
  name: bedrock-reasoning
spec:
  provider: awsbedrock
  config:
    model: "anthropic.claude-3-5-sonnet-20241022-v2:0"
    region: "us-east-1"
```

## Using Reasoning Streams

### Basic Request

When making requests to a Bedrock model with reasoning support, use the standard OpenAI-compatible API format:

```bash
curl -X POST http://your-gateway/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "anthropic.claude-3-5-sonnet-20241022-v2:0",
    "messages": [
      {
        "role": "user",
        "content": "Solve this math problem: If a train travels 120 miles in 2 hours, what is its average speed?"
      }
    ],
    "stream": true
  }'
```

### Response Format

When reasoning is available, the streaming response will include reasoning traces in addition to the regular content:

```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion.chunk",
  "created": 1677652288,
  "model": "anthropic.claude-3-5-sonnet-20241022-v2:0",
  "choices": [
    {
      "index": 0,
      "delta": {
        "reasoning": "I need to calculate the average speed using the formula: speed = distance / time. Given: distance = 120 miles, time = 2 hours."
      },
      "finish_reason": null
    }
  ]
}
```

## Implementation Details

The reasoning stream support works by:

1. **Request Translation**: Converting OpenAI-format requests to Bedrock's native format while preserving reasoning parameters
2. **Stream Processing**: Parsing Bedrock's streaming response format to extract both content and reasoning traces
3. **Response Translation**: Converting Bedrock responses back to OpenAI-compatible format with reasoning information included

## Error Handling

If reasoning is not supported by the selected model, the gateway will:

- Continue to provide regular streaming responses
- Log a warning about reasoning not being available
- Not fail the request due to missing reasoning capability

## Best Practices

### Model Selection

Choose models that explicitly support reasoning features:

```yaml
spec:
  config:
    model: "anthropic.claude-3-5-sonnet-20241022-v2:0"  # Supports reasoning
```

### Request Parameters

For optimal reasoning output, consider:

- Using detailed prompts that benefit from step-by-step reasoning
- Enabling streaming to see reasoning traces in real-time
- Setting appropriate temperature values for reasoning tasks

### Response Processing

When processing responses with reasoning:

```javascript
// Example client-side processing
const processStreamChunk = (chunk) => {
  if (chunk.choices[0].delta.reasoning) {
    console.log('Reasoning:', chunk.choices[0].delta.reasoning);
  }
  if (chunk.choices[0].delta.content) {
    console.log('Content:', chunk.choices[0].delta.content);
  }
};
```

## Troubleshooting

### Common Issues

**Reasoning not appearing in responses:**
- Verify the selected model supports reasoning features
- Check that streaming is enabled in your request
- Ensure your AWS Bedrock configuration is correct

**Stream parsing errors:**
- Verify your client correctly handles Server-Sent Events (SSE)
- Check that Content-Type headers are properly set
- Ensure your network configuration allows streaming responses

### Debugging

Enable debug logging to see the reasoning stream processing:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ai-gateway-config
data:
  log_level: "debug"
```

## Migration Notes

This feature is backward compatible. Existing configurations will continue to work without modification, and reasoning information will only be included when:

1. The underlying model supports it
2. The request format allows for it
3. Streaming is enabled

No changes are required to existing `AIServiceBackend` or `AIGatewayRoute` configurations.