# Header Mutations for Embeddings and Messages Processors

This guide explains how to use header mutations with the embeddings and messages processors in Envoy AI Gateway. Header mutations allow you to modify HTTP headers when routing requests to different backend services, providing enhanced security and customization capabilities.

## Overview

Header mutations enable you to:
- Remove sensitive headers (like authorization tokens) before forwarding requests
- Add custom headers required by specific backend services
- Transform headers to match backend service requirements

This feature is now available for:
- Chat completion processor (existing)
- Embeddings processor (new)
- Messages processor (new)

## Configuration

Header mutations are configured at the backend level using the `headerMutation` field in your AI Service Backend configuration.

### Basic Structure

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIServiceBackend
metadata:
  name: example-backend
spec:
  backend:
    # backend configuration
  headerMutation:
    remove:
      - "authorization"
      - "x-api-key"
    set:
      - name: "custom-header"
        value: "custom-value"
      - name: "x-backend-id"
        value: "backend-1"
```

## Embeddings Processor

Configure header mutations for embeddings endpoints:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIServiceBackend
metadata:
  name: embeddings-backend
spec:
  backend:
    type: "openai"
    endpoint: "https://api.openai.com/v1/embeddings"
  headerMutation:
    remove:
      - "x-original-auth"
    set:
      - name: "authorization"
        value: "Bearer ${OPENAI_API_KEY}"
      - name: "x-request-source"
        value: "ai-gateway"
---
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGatewayRoute
metadata:
  name: embeddings-route
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: embeddings-http-route
  rules:
    - backendRefs:
        - name: embeddings-backend
          port: 443
      matches:
        - path:
            type: PathPrefix
            value: "/v1/embeddings"
```

## Messages Processor

Configure header mutations for messages endpoints:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIServiceBackend
metadata:
  name: messages-backend
spec:
  backend:
    type: "anthropic"
    endpoint: "https://api.anthropic.com/v1/messages"
  headerMutation:
    remove:
      - "x-forwarded-for"
      - "x-original-token"
    set:
      - name: "x-api-key"
        value: "${ANTHROPIC_API_KEY}"
      - name: "anthropic-version"
        value: "2023-06-01"
---
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGatewayRoute
metadata:
  name: messages-route
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: messages-http-route
  rules:
    - backendRefs:
        - name: messages-backend
          port: 443
      matches:
        - path:
            type: PathPrefix
            value: "/v1/messages"
```

## Use Cases

### Security Enhancement

Remove sensitive headers before forwarding requests to prevent credential leakage:

```yaml
headerMutation:
  remove:
    - "authorization"
    - "x-api-key"
    - "x-auth-token"
  set:
    - name: "authorization"
      value: "Bearer ${BACKEND_SPECIFIC_TOKEN}"
```

### Backend-Specific Headers

Add headers required by specific AI service providers:

```yaml
# For Azure OpenAI
headerMutation:
  set:
    - name: "api-key"
      value: "${AZURE_OPENAI_KEY}"
    - name: "api-version"
      value: "2024-02-01"

# For Anthropic
headerMutation:
  set:
    - name: "x-api-key"
      value: "${ANTHROPIC_API_KEY}"
    - name: "anthropic-version"
      value: "2023-06-01"
```

### Request Tracking

Add custom headers for monitoring and debugging:

```yaml
headerMutation:
  set:
    - name: "x-gateway-version"
      value: "v0.3"
    - name: "x-processor-type"
      value: "embeddings"
    - name: "x-request-id"
      value: "${REQUEST_ID}"
```

## Best Practices

1. **Security First**: Always remove sensitive headers before adding new authentication headers
2. **Environment Variables**: Use environment variables for sensitive values like API keys
3. **Consistent Naming**: Use consistent header naming conventions across your backends
4. **Minimal Headers**: Only add headers that are necessary for the backend service
5. **Testing**: Test header mutations in a development environment before deploying to production

## Troubleshooting

### Headers Not Being Modified

- Verify the `headerMutation` configuration is properly indented under the backend spec
- Check that environment variables are properly set and accessible
- Ensure the AI Service Backend is correctly referenced in your AIGatewayRoute

### Authentication Failures

- Confirm that removed headers are not required by the backend
- Verify that new authentication headers match the backend service requirements
- Check that API keys and tokens are valid and have appropriate permissions

### Missing Custom Headers

- Ensure header names follow valid HTTP header naming conventions
- Verify that custom headers are not being overridden by other components
- Check the order of header operations (remove operations happen before set operations)

## Related Documentation

- [AI Service Backend Configuration](../capabilities/llm-integrations/connect-providers.md)
- [Supported Providers](../capabilities/llm-integrations/supported-providers.md)
- [Security Best Practices](../capabilities/security/index.md)