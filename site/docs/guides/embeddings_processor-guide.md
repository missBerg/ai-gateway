# Header Mutation for Embeddings and Messages Processors

## Overview

This guide explains the header mutation functionality that has been added to the embeddings and messages processors in Envoy AI Gateway. This feature allows you to manipulate HTTP headers when routing requests to backend services, providing enhanced security and customization capabilities.

## What is Header Mutation?

Header mutation enables you to:
- **Remove sensitive headers** like authorization tokens before forwarding requests to backends
- **Set custom headers** with specific values for different backend services
- **Restore original headers** during retry scenarios

This functionality was previously only available for chat completion endpoints and has now been extended to embeddings and messages processors for feature parity.

## How It Works

The header mutation functionality operates through the `HeaderMutator` component that:

1. **Removes specified headers** from requests before they reach the backend
2. **Sets new headers** with configured values
3. **Handles retry scenarios** by restoring original headers when appropriate
4. **Maintains header state** throughout the request lifecycle

## Configuration

Header mutations are configured per backend in your AI Gateway route configuration:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGatewayRoute
metadata:
  name: my-embeddings-route
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: my-embeddings-route
  rules:
    - matches:
        - headers:
            - name: x-ai-eg-model
              value: text-embedding-ada-002
      backendRefs:
        - name: my-backend
          headerMutation:
            remove:
              - "authorization"
              - "x-api-key"
            set:
              - name: "x-custom-backend"
                value: "embeddings-service"
              - name: "x-environment"
                value: "production"
```

## Supported Endpoints

Header mutation is now supported across all major AI Gateway processors:

### Embeddings Processor (`/v1/embeddings`)
- Supports OpenAI embedding requests
- Handles header mutations during request translation
- Compatible with various backend schemas

### Messages Processor (`/v1/messages`) 
- Supports Anthropic Messages API format
- Currently works with GCP Anthropic backends
- Maintains header mutations through streaming responses

### Chat Completion Processor (`/v1/chat/completions`)
- Full header mutation support (existing functionality)
- Works with all supported backend types
- Handles both streaming and non-streaming requests

## Common Use Cases

### Security: Removing Sensitive Headers

Remove authorization tokens that shouldn't be passed to specific backends:

```yaml
headerMutation:
  remove:
    - "authorization"
    - "x-api-key"
    - "x-auth-token"
```

### Backend Identification: Setting Custom Headers

Add headers to identify the routing path or environment:

```yaml
headerMutation:
  set:
    - name: "x-gateway-backend"
      value: "openai-production"
    - name: "x-request-source"
      value: "ai-gateway"
```

### Environment-Specific Configuration

Configure different headers for different environments:

```yaml
# Production backend
headerMutation:
  remove:
    - "x-debug-mode"
  set:
    - name: "x-environment"
      value: "prod"
    - name: "x-rate-limit-tier"
      value: "premium"

# Development backend  
headerMutation:
  set:
    - name: "x-environment"
      value: "dev"
    - name: "x-debug-enabled"
      value: "true"
```

## Implementation Details

### Request Processing Flow

1. **Router Filter Phase**: Original headers are captured and stored
2. **Backend Selection**: Header mutation configuration is loaded from the selected backend
3. **Header Processing**: 
   - Specified headers are removed from the request
   - New headers are set with configured values
   - Headers are updated in the request context
4. **Backend Authentication**: Authentication handlers can access the mutated headers
5. **Request Forwarding**: The modified request is sent to the backend

### Retry Handling

During retry scenarios:
- Original headers are restored if they were previously removed
- Set headers are reapplied consistently
- Header state is maintained across retry attempts

### Header Precedence

The header mutation follows this precedence order:
1. **Remove operations** are applied first
2. **Set operations** override any existing headers
3. **Original headers** are restored on retry (if not explicitly removed or set)

## Best Practices

### Security Considerations

- Always remove sensitive authentication headers when they're not needed by the backend
- Use specific header names rather than wildcards for security
- Regularly audit header mutation configurations

### Performance Optimization

- Minimize the number of header operations per request
- Cache header mutation configurations when possible
- Monitor header processing overhead in high-traffic scenarios

### Configuration Management

- Use consistent naming conventions for custom headers (e.g., `x-gateway-*`)
- Document header mutation purposes in your configuration
- Test header mutations in non-production environments first

## Troubleshooting

### Common Issues

**Headers not being removed:**
- Verify header names match exactly (case-sensitive)
- Check that the header exists in the original request

**Custom headers not appearing:**
- Ensure the header mutation is configured on the correct backend
- Verify the backend is being selected properly

**Retry behavior problems:**
- Check that original headers are being captured correctly
- Verify retry logic isn't conflicting with header mutations

### Debug Information

Enable debug logging to see header mutation operations:

```yaml
# In your AI Gateway configuration
logging:
  level: debug
  # Look for log entries with "processor" and "header-mutation" fields
```

## Migration from Chat Completion

If you're already using header mutations with chat completion endpoints, the configuration format is identical for embeddings and messages processors. Simply apply the same `headerMutation` configuration to your new backend definitions.

The functionality is fully compatible and maintains the same behavior patterns across all processor types.