# Header Mutations for Embeddings and Messages Processors

This guide explains how to configure per-backend header mutations for embeddings and messages processors in AI Gateway. Header mutations allow you to remove sensitive headers (like authorization tokens) and set custom headers when routing requests to different backend services.

## Overview

Header mutations provide fine-grained control over HTTP headers when requests are forwarded to backend AI services. This feature is particularly useful for:

- Removing sensitive headers like authorization tokens before forwarding requests
- Setting custom headers required by specific backend services
- Modifying headers based on routing configuration

This functionality is now available for:
- Chat completion processor
- Embeddings processor  
- Messages processor

## Configuration

Header mutations are configured in the AI Service Backend resource using the `headerMutations` field:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIServiceBackend
metadata:
  name: my-backend
spec:
  backendRef:
    name: my-service
    port: 8080
  headerMutations:
    remove:
      - "authorization"
      - "x-api-key"
    set:
      - name: "x-custom-header"
        value: "custom-value"
      - name: "x-backend-id"
        value: "backend-001"
```

## Configuration Fields

### `headerMutations`

The main configuration block for header mutations.

### `headerMutations.remove`

An array of header names to remove from the request before forwarding to the backend.

**Example:**
```yaml
headerMutations:
  remove:
    - "authorization"
    - "x-original-token"
```

### `headerMutations.set`

An array of headers to set on the request before forwarding to the backend.

**Fields:**
- `name`: The header name to set
- `value`: The header value to set

**Example:**
```yaml
headerMutations:
  set:
    - name: "x-api-version"
      value: "v1"
    - name: "x-service-name"
      value: "ai-gateway"
```

## Use Cases

### Removing Sensitive Headers

Remove authorization headers that shouldn't be forwarded to certain backends:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIServiceBackend
metadata:
  name: public-backend
spec:
  backendRef:
    name: public-ai-service
    port: 8080
  headerMutations:
    remove:
      - "authorization"
      - "x-api-key"
      - "x-user-token"
```

### Setting Backend-Specific Headers

Add headers required by specific AI service providers:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIServiceBackend
metadata:
  name: custom-provider
spec:
  backendRef:
    name: custom-ai-service
    port: 8080
  headerMutations:
    set:
      - name: "x-provider-auth"
        value: "bearer-token-123"
      - name: "x-model-version"
        value: "latest"
      - name: "content-type"
        value: "application/json"
```

### Combined Header Operations

Remove sensitive headers and add backend-specific ones:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIServiceBackend
metadata:
  name: secure-backend
spec:
  backendRef:
    name: secure-ai-service
    port: 443
  headerMutations:
    remove:
      - "x-original-auth"
      - "x-client-token"
    set:
      - name: "x-backend-auth"
        value: "service-specific-token"
      - name: "x-request-source"
        value: "ai-gateway"
```

## Processor Compatibility

Header mutations are supported by the following processors:

| Processor Type | Support Status |
|----------------|----------------|
| Chat Completion | ✅ Supported |
| Embeddings | ✅ Supported |
| Messages | ✅ Supported |

## Best Practices

1. **Security**: Always remove sensitive headers when routing to external or less trusted backends
2. **Consistency**: Use consistent header naming conventions across your backends
3. **Documentation**: Document the purpose of custom headers for team members
4. **Testing**: Test header mutations in development environments before deploying to production

## Example: Embeddings Service with Header Mutations

Here's a complete example showing an embeddings service with header mutations:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIServiceBackend
metadata:
  name: embeddings-backend
  namespace: ai-gateway
spec:
  backendRef:
    name: embeddings-service
    port: 8080
  headerMutations:
    remove:
      - "authorization"  # Remove client auth token
      - "x-user-id"      # Remove user identification
    set:
      - name: "x-service-auth"
        value: "embeddings-service-token"
      - name: "x-processor-type"
        value: "embeddings"
      - name: "accept"
        value: "application/json"
---
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGatewayRoute
metadata:
  name: embeddings-route
  namespace: ai-gateway
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: embeddings-http-route
  rules:
    - backendRefs:
        - name: embeddings-backend
          weight: 100
```

This configuration removes sensitive client headers and adds service-specific authentication and metadata headers when processing embeddings requests.