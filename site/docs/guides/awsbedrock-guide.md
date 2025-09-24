# AWS Bedrock Reasoning Stream Support Guide

This guide explains how to use the new AWS Bedrock reasoning stream support feature in AI Gateway.

## Overview

AI Gateway now supports streaming responses with reasoning capabilities from AWS Bedrock models. This feature allows you to receive real-time streaming responses that include the model's reasoning process, providing transparency into how the AI arrives at its conclusions.

## Prerequisites

- AI Gateway installed and configured
- AWS Bedrock access with appropriate permissions
- A compatible AWS Bedrock model that supports reasoning (e.g., Claude models)

## Configuration

### Basic Setup

Configure your AI Gateway to connect to AWS Bedrock with reasoning stream support:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIServiceBackend
metadata:
  name: bedrock-reasoning-backend
spec:
  provider:
    type: awsBedrock
    awsBedrock:
      region: us-east-1
      model: anthropic.claude-3-5-sonnet-20241022-v2:0
      auth:
        accessKey:
          name: aws-credentials
          key: access-key
        secretKey:
          name: aws-credentials
          key: secret-key
```

### Route Configuration

Set up routing to enable reasoning stream responses:

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
    - name: bedrock-reasoning-backend
      weight: 100
```

## Usage Examples

### Streaming Request with Reasoning

Send a chat completion request that will return a streaming response with reasoning:

```bash
curl -X POST http://your-gateway-url/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "anthropic.claude-3-5-sonnet-20241022-v2:0",
    "messages": [
      {
        "role": "user",
        "content": "Explain the process of photosynthesis and why it is important for life on Earth."
      }
    ],
    "stream": true,
    "max_tokens": 1000
  }'
```

### Response Format

The streaming response will include reasoning information in the following format:

```json
data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677652288,"model":"anthropic.claude-3-5-sonnet-20241022-v2:0","choices":[{"index":0,"delta":{"content":"","reasoning":"The user is asking about photosynthesis, which is a fundamental biological process. I should explain both the mechanism and its significance comprehensively."},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677652288,"model":"anthropic.claude-3-5-sonnet-20241022-v2:0","choices":[{"index":0,"delta":{"content":"Photosynthesis is the process by which plants..."},"finish_reason":null}]}
```

## Key Features

### Reasoning Transparency

The reasoning stream provides insight into the model's thought process:

- **Reasoning Field**: Contains the model's internal reasoning before generating the response
- **Content Field**: Contains the actual response content
- **Real-time Streaming**: Both reasoning and content are streamed in real-time

### Error Handling

The feature includes robust error handling for:

- Invalid AWS credentials
- Unsupported model types
- Network connectivity issues
- Malformed requests

## Best Practices

### Performance Optimization

1. **Connection Pooling**: Use persistent connections for better performance
2. **Request Batching**: Consider batching multiple requests when appropriate
3. **Timeout Configuration**: Set appropriate timeouts for streaming responses

### Security Considerations

1. **Credential Management**: Store AWS credentials securely using Kubernetes secrets
2. **Network Security**: Use TLS encryption for all communications
3. **Access Control**: Implement proper RBAC for AI Gateway resources

### Monitoring

Monitor your reasoning stream usage:

```yaml
# Example metrics to track
- bedrock_reasoning_requests_total
- bedrock_reasoning_stream_duration
- bedrock_reasoning_errors_total
```

## Troubleshooting

### Common Issues

**Stream Connection Drops**
- Check network connectivity to AWS Bedrock
- Verify timeout configurations
- Review AWS service limits

**Missing Reasoning Data**
- Ensure the model supports reasoning capabilities
- Check request format and parameters
- Verify AI Gateway version compatibility

**Authentication Errors**
- Validate AWS credentials
- Check IAM permissions for Bedrock access
- Verify region configuration

### Debug Mode

Enable debug logging to troubleshoot issues:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ai-gateway-config
data:
  log-level: debug
```

## Limitations

- Reasoning stream support is currently available for select AWS Bedrock models
- Requires AI Gateway version 0.4.0 or later
- Some advanced reasoning features may vary by model

## Related Documentation

- [AWS Bedrock Integration Guide](connect-providers/aws-bedrock.md)
- [Streaming Responses](../capabilities/traffic/streaming.md)
- [Monitoring and Observability](../capabilities/observability/index.md)