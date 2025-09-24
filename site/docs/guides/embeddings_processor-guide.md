# Header Mutations for Embeddings and Messages Processors

This guide explains how to configure per-backend header mutations for the embeddings and messages processors in Envoy AI Gateway. Header mutations allow you to modify HTTP headers when routing requests to different backend services, providing enhanced security and customization capabilities.

## Overview

Header mutations enable you to:
- Remove sensitive headers (like authorization tokens) before forwarding requests
- Add custom headers specific to each backend
- Transform headers based on the target backend service

This feature is now available for:
- **Embeddings processor** - handles text embedding requests
- **Messages processor** - handles message-based API requests  
- **Chat completion processor** - already supported (existing feature)

## Configuration

Header mutations are configured at the backend level within your AI Gateway Route configuration. You can specify which headers to remove and which headers to add for each backend service.

### Basic Syntax

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGatewayRoute
metadata:
  name: example-route
spec:
  hostnames:
    - "api.example.com"
  rules:
    - backendRefs:
        - name: backend-service
          headerMutation:
            remove:
              - "authorization"
              - "x-api-key"
            set:
              - name: "x-custom-header"
                value: "custom-value"
              - name: "x-backend-id"
                value: "backend-1"
```

## Examples

### Embeddings Processor Configuration

Here's an example configuration for an embeddings service with header mutations:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGatewayRoute
metadata:
  name: embeddings-route
spec:
  hostnames:
    - "embeddings.ai-gateway.local"
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: "/v1/embeddings"
      backendRefs:
        - name: openai-embeddings-backend
          headerMutation:
            remove:
              - "authorization"
              - "x-forwarded-for"
            set:
              - name: "authorization"
                value: "Bearer ${OPENAI_API_KEY}"
              - name: "x-source"
                value: "ai-gateway"
        - name: azure-embeddings-backend
          headerMutation:
            remove:
              - "authorization"
            set:
              - name: "api-key"
                value: "${AZURE_API_KEY}"
              - name: "x-azure-region"
                value: "eastus"
```

### Messages Processor Configuration

Example configuration for a messages service with multiple backends:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGatewayRoute
metadata:
  name: messages-route
spec:
  hostnames:
    - "messages.ai-gateway.local"
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: "/v1/messages"
      backendRefs:
        - name: anthropic-backend
          headerMutation:
            remove:
              - "authorization"
              - "x-api-key"
            set:
              - name: "x-api-key"
                value: "${ANTHROPIC_API_KEY}"
              - name: "anthropic-version"
                value: "2023-06-01"
        - name: fallback-backend
          headerMutation:
            remove:
              - "x-api-key"
            set:
              - name: "authorization"
                value: "Bearer ${FALLBACK_API_KEY}"
```

## Header Mutation Options

### Remove Headers

The `remove` section specifies headers that should be stripped from the request before forwarding to the backend:

```yaml
headerMutation:
  remove:
    - "authorization"
    - "x-api-key"
    - "x-forwarded-for"
    - "user-agent"
```

### Set Headers

The `set` section allows you to add or override headers with specific values:

```yaml
headerMutation:
  set:
    - name: "authorization"
      value: "Bearer your-api-key"
    - name: "content-type"
      value: "application/json"
    - name: "x-custom-header"
      value: "custom-value"
```

## Security Considerations

### Removing Sensitive Headers

Always remove sensitive headers that should not be forwarded to backend services:

```yaml
headerMutation:
  remove:
    - "authorization"      # Remove client auth tokens
    - "x-api-key"         # Remove client API keys
    - "cookie"            # Remove session cookies
    - "x-forwarded-for"   # Remove client IP information
```

### Using Environment Variables

Use environment variables for sensitive values in header mutations:

```yaml
headerMutation:
  set:
    - name: "authorization"
      value: "Bearer ${BACKEND_API_KEY}"
    - name: "x-api-key"
      value: "${BACKEND_SECRET}"
```

## Best Practices

1. **Security First**: Always remove sensitive client headers before forwarding requests
2. **Environment Variables**: Use environment variables for API keys and secrets
3. **Backend-Specific Headers**: Set headers that are specific to each backend service
4. **Minimal Headers**: Only set headers that are required by the backend
5. **Consistent Naming**: Use consistent header naming conventions across backends

## Common Use Cases

### Multi-Provider Setup

Configure different authentication methods for different AI providers:

```yaml
backendRefs:
  - name: openai-backend
    headerMutation:
      remove: ["authorization"]
      set:
        - name: "authorization"
          value: "Bearer ${OPENAI_API_KEY}"
  
  - name: anthropic-backend
    headerMutation:
      remove: ["authorization"]
      set:
        - name: "x-api-key"
          value: "${ANTHROPIC_API_KEY}"
  
  - name: azure-backend
    headerMutation:
      remove: ["authorization"]
      set:
        - name: "api-key"
          value: "${AZURE_API_KEY}"
```

### Load Balancing with Custom Headers

Add backend identification headers for monitoring and debugging:

```yaml
backendRefs:
  - name: primary-backend
    headerMutation:
      set:
        - name: "x-backend-tier"
          value: "primary"
        - name: "x-routing-id"
          value: "route-001"
  
  - name: secondary-backend
    headerMutation:
      set:
        - name: "x-backend-tier"
          value: "secondary"
        - name: "x-routing-id"
          value: "route-002"
```

## Troubleshooting

### Debugging Header Issues

1. **Check Configuration**: Verify header mutation syntax in your AI Gateway Route
2. **Environment Variables**: Ensure environment variables are properly set
3. **Header Names**: Verify header names match backend API requirements
4. **Logs**: Check AI Gateway logs for header-related errors

### Common Issues

- **Authentication Failures**: Ensure backend API keys are correctly configured
- **Missing Headers**: Verify required headers are set for each backend
- **Case Sensitivity**: Some backends are case-sensitive for header names
- **Value Format**: Ensure header values match expected formats (e.g., "Bearer token" vs "token")

This feature brings the embeddings and messages processors to feature parity with the chat completion processor, providing consistent header mutation capabilities across all AI Gateway processors.