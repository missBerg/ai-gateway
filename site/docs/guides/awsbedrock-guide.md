# AWS Bedrock Reasoning Stream Support Guide

This guide explains how to use AWS Bedrock's reasoning stream support with Envoy AI Gateway, which enables streaming responses that include reasoning tokens for supported models.

## Overview

AWS Bedrock reasoning stream support allows you to receive streaming responses that include the model's reasoning process. This is particularly useful for models that support chain-of-thought reasoning, as you can observe how the model arrives at its conclusions in real-time.

## Supported Models

The reasoning stream feature is available for AWS Bedrock models that support reasoning tokens in their streaming responses. Check the AWS Bedrock documentation for the most up-to-date list of compatible models.

## Configuration

### AIServiceBackend Configuration

Configure your AWS Bedrock backend to enable reasoning stream support:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIServiceBackend
metadata:
  name: aws-bedrock-reasoning
spec:
  awsBedrock:
    region: us-west-2
    model: anthropic.claude-3-5-sonnet-20241022-v2:0
    accessKey:
      name: aws-credentials
      key: access-key
    secretKey:
      name: aws-credentials
      key: secret-key
```

### AIGatewayRoute Configuration

Set up your route to handle reasoning stream requests:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGatewayRoute
metadata:
  name: bedrock-reasoning-route
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: ai-gateway
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /v1/chat/completions
    backends:
    - name: aws-bedrock-reasoning
      weight: 1
```

## Using Reasoning Streams

### Basic Request

Send a chat completion request with streaming enabled:

```bash
curl -X POST http://your-gateway-endpoint/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "model": "anthropic.claude-3-5-sonnet-20241022-v2:0",
    "messages": [
      {
        "role": "user",
        "content": "Explain the reasoning behind solving this math problem: What is 2^10 * 3^5?"
      }
    ],
    "stream": true
  }'
```

### Response Format

The response will include both reasoning tokens and regular content tokens:

```
data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1234567890,"model":"anthropic.claude-3-5-sonnet-20241022-v2:0","choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1234567890,"model":"anthropic.claude-3-5-sonnet-20241022-v2:0","choices":[{"index":0,"delta":{"content":"<thinking>\nI need to calculate 2^10 * 3^5.\n\nFirst, let me calculate 2^10:\n2^10 = 1024\n\nNext, let me calculate 3^5:\n3^5 = 3 * 3 * 3 * 3 * 3 = 9 * 9 * 3 = 81 * 3 = 243\n\nSo the answer is 1024 * 243.\n</thinking>"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1234567890,"model":"anthropic.claude-3-5-sonnet-20241022-v2:0","choices":[{"index":0,"delta":{"content":"To solve 2^10 * 3^5, I'll calculate each part separately:\n\n**Step 1: Calculate 2^10**\n2^10 = 1,024\n\n**Step 2: Calculate 3^5**\n3^5 = 243\n\n**Step 3: Multiply the results**\n1,024 × 243 = 248,832\n\nTherefore, 2^10 * 3^5 = 248,832"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1234567890,"model":"anthropic.claude-3-5-sonnet-20241022-v2:0","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]
```

## Key Features

### Reasoning Token Handling

- **Transparent Processing**: Reasoning tokens are automatically handled and included in the stream
- **Format Preservation**: The original reasoning format from AWS Bedrock is maintained
- **Stream Continuity**: Reasoning and content tokens are seamlessly integrated in the response stream

### Performance Considerations

- **Low Latency**: Reasoning tokens are streamed as they become available
- **Memory Efficient**: Tokens are processed and forwarded without buffering the entire response
- **Connection Management**: Maintains persistent connections for optimal streaming performance

## Best Practices

### Client Implementation

1. **Handle Reasoning Content**: Be prepared to receive reasoning tokens that may contain model thinking process
2. **Stream Processing**: Process tokens incrementally rather than waiting for complete responses
3. **Error Handling**: Implement robust error handling for stream interruptions

### Request Optimization

1. **Model Selection**: Choose models that support reasoning for the best experience
2. **Prompt Design**: Structure prompts to encourage clear reasoning when beneficial
3. **Stream Management**: Monitor stream health and implement reconnection logic

## Troubleshooting

### Common Issues

**Stream Not Starting**
- Verify your AWS credentials have the necessary permissions
- Check that the selected model supports reasoning streams
- Ensure the `stream: true` parameter is included in your request

**Missing Reasoning Tokens**
- Confirm you're using a reasoning-capable model
- Check that your prompt encourages the model to show its reasoning
- Verify the gateway configuration includes proper backend setup

**Performance Issues**
- Monitor network connectivity between the gateway and AWS Bedrock
- Consider adjusting timeout settings for longer reasoning processes
- Check for any rate limiting on your AWS account

### Debugging

Enable debug logging to troubleshoot issues:

```yaml
# In your gateway configuration
spec:
  logging:
    level: debug
```

This will provide detailed information about the reasoning stream processing and any potential issues.

## Related Documentation

- [AWS Bedrock Integration Guide](connect-providers/aws-bedrock.md)
- [Streaming Responses](../capabilities/streaming.md)
- [Backend Configuration](../concepts/resources.md#aiservicebackend)