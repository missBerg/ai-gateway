---
id: request-mutation
title: Request Mutation
sidebar_position: 9
---

# Request Mutation

Envoy AI Gateway allows you to modify HTTP headers and JSON body fields before requests are sent to upstream AI providers. This enables customization without changing client code.

## Overview

Request mutation supports two types of modifications:

| Mutation Type | What It Does | Use Cases |
|---------------|--------------|-----------|
| **Header Mutation** | Add, modify, or remove HTTP headers | Custom routing headers, remove sensitive headers, inject auth tokens |
| **Body Mutation** | Add, modify, or remove JSON body fields | Override model parameters, inject metadata, remove internal flags |

:::info How It Works
Under the hood, Envoy AI Gateway implements request mutations using Envoy's [External Processing filter](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_proc_filter) (`ext_proc`). This filter sits in the HTTP filter chain after authentication filters but before the router, allowing mutations to be applied before requests reach upstream providers.
:::

## Configuration Levels

Mutations can be configured at two levels:

1. **Backend Level** (`AIServiceBackend`) - Applied to all requests going to a specific backend
2. **Route Level** (`AIGatewayRouteRuleBackendRef`) - Applied to requests matching a specific route rule

:::info Precedence
When both levels define mutations, **route-level takes precedence** for conflicting operations. Non-conflicting mutations from both levels are merged.
:::

## Header Mutation

Use `headerMutation` to modify HTTP request headers.

### Configuration

```yaml
headerMutation:
  set:
    - name: "header-name"
      value: "header-value"
  remove:
    - "header-to-remove"
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `set` | `[]HTTPHeader` | Headers to add or overwrite. Max 16 items per mutation. |
| `remove` | `[]string` | Header names to remove (case-insensitive). Max 16 items per mutation. |

:::note
The 16-item limit is an Envoy AI Gateway design constraint to ensure predictable performance, not an underlying Envoy limitation.
:::

### Example: Backend-Level Header Mutation

Add custom headers to all requests going to a specific backend:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIServiceBackend
metadata:
  name: openai-backend
spec:
  schema:
    name: OpenAI
  backendRef:
    name: openai
    kind: Backend
    group: gateway.envoyproxy.io
  headerMutation:
    set:
      - name: "x-custom-tenant"
        value: "tenant-123"
      - name: "x-request-source"
        value: "ai-gateway"
    remove:
      - "x-internal-debug"
```

### Example: Route-Level Header Mutation

Apply different headers based on routing rules:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGatewayRoute
metadata:
  name: my-route
spec:
  parentRefs:
    - name: my-gateway
      kind: Gateway
      group: gateway.networking.k8s.io
  rules:
    - matches:
        - headers:
            - type: Exact
              name: x-ai-eg-model
              value: gpt-4o
      backendRefs:
        - name: openai-backend
          headerMutation:
            set:
              - name: "x-priority"
                value: "high"
```

## Body Mutation

Use `bodyMutation` to modify JSON fields in the request body.

### Configuration

```yaml
bodyMutation:
  set:
    - path: "field_name"
      value: '"string_value"'  # JSON-encoded value
  remove:
    - "field_to_remove"
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `set` | `[]HTTPBodyField` | Fields to add or overwrite. Max 16 items per mutation. |
| `remove` | `[]string` | Top-level field names to remove. Max 16 items per mutation. |

### Value Encoding

The `value` field must be a valid JSON value as a string:

| Type | Example Value | Result in Body |
|------|---------------|----------------|
| String | `"\"scale\""` | `"scale"` |
| Number | `"42"` | `42` |
| Boolean | `"true"` | `true` |
| Object | `"{\"key\": \"value\"}"` | `{"key": "value"}` |
| Array | `"[1, 2, 3]"` | `[1, 2, 3]` |
| Null | `"null"` | `null` |

### Example: Override Service Tier

Force all requests to use a specific OpenAI service tier:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIServiceBackend
metadata:
  name: openai-backend
spec:
  schema:
    name: OpenAI
  backendRef:
    name: openai
    kind: Backend
    group: gateway.envoyproxy.io
  bodyMutation:
    set:
      - path: "service_tier"
        value: '"scale"'
```

**Input request:**
```json
{
  "model": "gpt-4o",
  "messages": [{"role": "user", "content": "Hello"}],
  "service_tier": "default"
}
```

**Output to backend:**
```json
{
  "model": "gpt-4o",
  "messages": [{"role": "user", "content": "Hello"}],
  "service_tier": "scale"
}
```

### Example: Remove Sensitive Fields

Strip internal fields before sending to upstream:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIServiceBackend
metadata:
  name: openai-backend
spec:
  schema:
    name: OpenAI
  backendRef:
    name: openai
    kind: Backend
    group: gateway.envoyproxy.io
  bodyMutation:
    remove:
      - "internal_request_id"
      - "debug_mode"
      - "internal_metadata"
```

### Example: Inject Default Parameters

Add default parameters that clients don't need to specify:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIServiceBackend
metadata:
  name: openai-backend
spec:
  schema:
    name: OpenAI
  backendRef:
    name: openai
    kind: Backend
    group: gateway.envoyproxy.io
  bodyMutation:
    set:
      - path: "max_tokens"
        value: "4096"
      - path: "temperature"
        value: "0.7"
      - path: "user"
        value: '"ai-gateway-user"'
```

## Combined Example

Apply both header and body mutations at the route level:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGatewayRoute
metadata:
  name: production-route
spec:
  parentRefs:
    - name: my-gateway
      kind: Gateway
      group: gateway.networking.k8s.io
  rules:
    - matches:
        - headers:
            - type: Exact
              name: x-ai-eg-model
              value: gpt-4o
      backendRefs:
        - name: openai-production
          headerMutation:
            set:
              - name: "x-environment"
                value: "production"
            remove:
              - "x-debug-token"
          bodyMutation:
            set:
              - path: "service_tier"
                value: '"scale"'
              - path: "stream_options"
                value: '{"include_usage": true}'
            remove:
              - "internal_trace_id"
```

## Limitations

- **Top-level fields only**: Body mutation currently supports only top-level JSON fields. Nested field paths (e.g., `messages[0].content`) are not supported.
- **InferencePool backends**: Header and body mutations are ignored when the backend references an `InferencePool` resource, as InferencePool handles its own request processing.
- **JSON bodies only**: Body mutation only works with JSON request bodies. Non-JSON content types are passed through unmodified.
- **Filter chain position**: Mutations are applied at the External Processing filter's position in the filter chain, after authentication but before upstream routing.

## Use Cases

| Use Case | Mutation Type | Example |
|----------|---------------|---------|
| Custom routing headers | Header | Add `x-tenant-id` for multi-tenant routing |
| Remove debug headers | Header | Strip `x-debug-*` headers before upstream |
| Force service tier | Body | Set `service_tier` to `"scale"` |
| Inject user tracking | Body | Add `user` field for usage tracking |
| Remove internal fields | Body | Strip `internal_*` fields |
| Default parameters | Body | Set `max_tokens`, `temperature` defaults |
| Compliance filtering | Body | Remove PII-related custom fields |

## See Also

- [Model Name Virtualization](./model-virtualization.md) - Override model names per backend
- [Provider Fallback](./provider-fallback.md) - Configure fallback between providers
- [AIServiceBackend API Reference](/api/api.mdx#aiservicebackend) - Full API specification
