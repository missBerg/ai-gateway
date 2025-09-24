# AWS Bedrock Reasoning Stream Support Guide

This guide explains how to use the AWS Bedrock reasoning stream support feature in Envoy AI Gateway, which enables streaming responses with reasoning capabilities from AWS Bedrock models.

## Overview

The AWS Bedrock reasoning stream support allows you to receive streaming responses from AWS Bedrock models that include reasoning information. This feature is particularly useful for models that provide step-by-step reasoning in their responses, allowing you to see the model's thought process as it generates the final answer.

## Prerequisites

- Envoy AI Gateway installed and configured
- AWS Bedrock access with appropriate credentials
- A configured AIServiceBackend for AWS Bedrock
- An AIGatewayRoute configured to use the Bedrock backend

## Configuration

### AIServiceBackend Setup

Configure your AWS Bedrock backend to support reasoning streams:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIServiceBackend
metadata:
  name: bedrock-reasoning-backend
spec:
  type: awsBedrock
  awsBedrock:
    region: us-east-1
    model: anthropic.claude-3-sonnet-20240229-v1:0
    # Additional AWS Bedrock configuration
```

### AIGatewayRoute Configuration

Set up your route to handle streaming requests:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGatewayRoute
metadata:
  name: bedrock-reasoning-route
spec:
  hostnames:
  - "ai-gateway.example.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /v1/chat/completions
    backends:
    - name: bedrock-reasoning-backend
      weight: 100
```

## Usage Examples

### Basic Streaming Request

Send a chat completion request with streaming enabled:

```bash
curl -X POST https://ai-gateway.example.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "model": "anthropic.claude-3-sonnet-20240229-v1:0",
    "messages": [
      {
        "role": "user",
        "content": "Explain step by step how to solve: What is 15% of 240?"
      }
    ],
    "stream": true,
    "max_tokens": 1000
  }'
```

### Response Format

With reasoning stream support, you'll receive Server-Sent Events (SSE) that include both the reasoning process and the final response:

```
data: {"id":"msg-123","object":"chat.completion.chunk","created":1234567890,"model":"anthropic.claude-3-sonnet-20240229-v1:0","choices":[{"index":0,"delta":{"role":"assistant","content":"I need to calculate 15% of 240. Let me break this down:"},"finish_reason":null}]}

data: {"id":"msg-123","object":"chat.completion.chunk","created":1234567890,"model":"anthropic.claude-3-sonnet-20240229-v1:0","choices":[{"index":0,"delta":{"content":"\n\nStep 1: Convert the percentage to a decimal\n15% = 15/100 = 0.15"},"finish_reason":null}]}

data: {"id":"msg-123","object":"chat.completion.chunk","created":1234567890,"model":"anthropic.claude-3-sonnet-20240229-v1:0","choices":[{"index":0,"delta":{"content":"\n\nStep 2: Multiply by the number\n0.15 × 240 = 36"},"finish_reason":null}]}

data: {"id":"msg-123","object":"chat.completion.chunk","created":1234567890,"model":"anthropic.claude-3-sonnet-20240229-v1:0","choices":[{"index":0,"delta":{"content":"\n\nTherefore, 15% of 240 is 36."},"finish_reason":"stop"}]}

data: [DONE]
```

## Advanced Configuration

### Reasoning-Specific Parameters

You can control the reasoning behavior using model-specific parameters:

```json
{
  "model": "anthropic.claude-3-sonnet-20240229-v1:0",
  "messages": [...],
  "stream": true,
  "max_tokens": 2000,
  "temperature": 0.1,
  "reasoning": {
    "show_thinking": true,
    "max_reasoning_tokens": 1000
  }
}
```

### Error Handling

Handle streaming errors appropriately:

```javascript
const response = await fetch('/v1/chat/completions', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer your-token'
  },
  body: JSON.stringify({
    model: 'anthropic.claude-3-sonnet-20240229-v1:0',
    messages: [...],
    stream: true
  })
});

const reader = response.body.getReader();
const decoder = new TextDecoder();

try {
  while (true) {
    const { done, value } = await reader.read();
    if (done) break;
    
    const chunk = decoder.decode(value);
    const lines = chunk.split('\n');
    
    for (const line of lines) {
      if (line.startsWith('data: ')) {
        const data = line.slice(6);
        if (data === '[DONE]') return;
        
        try {
          const parsed = JSON.parse(data);
          // Handle the streaming chunk
          console.log(parsed.choices[0]?.delta?.content);
        } catch (e) {
          console.error('Failed to parse chunk:', e);
        }
      }
    }
  }
} catch (error) {
  console.error('Streaming error:', error);
}
```

## Troubleshooting

### Common Issues

1. **No reasoning in responses**: Ensure your model supports reasoning capabilities and that reasoning parameters are correctly configured.

2. **Stream interruption**: Check network connectivity and ensure your client can handle long-lived connections.

3. **Authentication errors**: Verify AWS credentials and permissions for Bedrock access.

### Debugging

Enable debug logging to troubleshoot issues:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ai-gateway-config
data:
  config.yaml: |
    logging:
      level: debug
    # Other configuration...
```

## Best Practices

1. **Token Management**: Set appropriate `max_tokens` limits to control response length and costs.

2. **Temperature Settings**: Use lower temperature values (0.1-0.3) for more consistent reasoning outputs.

3. **Error Handling**: Implement robust error handling for stream interruptions and parsing errors.

4. **Rate Limiting**: Consider implementing rate limiting to manage API usage and costs.

5. **Monitoring**: Monitor streaming performance and error rates through observability features.

## Related Documentation

- [AWS Bedrock Integration](../getting-started/connect-providers/aws-bedrock.md)
- [Streaming Responses](../capabilities/traffic/streaming.md)
- [Observability](../capabilities/observability/index.md)