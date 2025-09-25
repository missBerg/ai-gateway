---
id: environment-configuration
title: Environment Variable Configuration
sidebar_position: 9
---

# Environment Variable Configuration

## Overview

Envoy AI Gateway supports flexible environment variable configuration for the external processor (ext_proc) component, including secure handling of sensitive data through Kubernetes secrets.

## Basic Environment Variable Configuration

Environment variables can be configured in the AI Gateway Helm chart using the `extProc.extraEnvVars` parameter. Each environment variable can be set as either a literal value or by referencing a Kubernetes secret.

### Literal Values

```yaml
# values.yaml
extProc:
  extraEnvVars:
    - name: OTEL_SERVICE_NAME
      value: "my-ai-gateway"
    - name: OTEL_TRACES_EXPORTER
      value: "otlp"
    - name: DEBUG_LEVEL
      value: "info"
```

Or via Helm command line:

```shell
helm upgrade ai-eg oci://docker.io/envoyproxy/ai-gateway-helm \
  --version v0.0.0-latest \
  --namespace envoy-ai-gateway-system \
  --set "extProc.extraEnvVars[0].name=OTEL_SERVICE_NAME" \
  --set "extProc.extraEnvVars[0].value=my-ai-gateway" \
  --set "extProc.extraEnvVars[1].name=OTEL_TRACES_EXPORTER" \
  --set "extProc.extraEnvVars[1].value=otlp"
```

## Secret Reference Configuration

For sensitive configuration data like API keys, authentication tokens, or private endpoints, you can reference Kubernetes secrets using the `secretKeyRef` syntax.

### SecretKeyRef Syntax

When using the CLI `aigw run` command, environment variables support a special `secretKeyRef` syntax:

```
VARIABLE_NAME=secretKeyRef {"name":"SECRET_NAME","key":"SECRET_KEY"}
```

**Key points:**
- The JSON object must have valid JSON syntax with quoted keys and values
- Whitespace around the JSON object is allowed and will be trimmed
- The secret must exist in the same namespace as the AI Gateway

### Examples

#### OpenTelemetry Headers with Authentication

```shell
# Create a secret with OTEL headers that include authentication
kubectl create secret generic otel-headers \
  --from-literal=headers='authorization=Bearer secret-token,x-custom-header=sensitive-value' \
  -n envoy-ai-gateway-system

# Use the secret in aigw run
OTEL_EXPORTER_OTLP_HEADERS=secretKeyRef {"name":"otel-headers","key":"headers"} aigw run
```

#### Multiple Environment Variables with Mixed Types

```shell
# Mix literal values and secret references
OTEL_SERVICE_NAME=ai-gateway;OTEL_EXPORTER_OTLP_HEADERS=secretKeyRef {"name":"otel-headers","key":"headers"}
```

#### Flexible JSON Formatting

The secretKeyRef syntax accepts various JSON formatting styles:

```shell
# Compact format
OTEL_HEADERS=secretKeyRef {"name":"otel-secret","key":"headers"}

# With spaces
OTEL_HEADERS=secretKeyRef { "name": "otel-secret", "key": "headers" }

# Trailing semicolon (ignored)
OTEL_HEADERS=secretKeyRef {"name":"otel-secret","key":"headers"};
```

## Helm Chart Secret References

When using the Helm chart, you can reference secrets using standard Kubernetes `valueFrom` syntax:

```yaml
# values.yaml
extProc:
  extraEnvVars:
    - name: OTEL_SERVICE_NAME
      value: "my-ai-gateway"
    - name: OTEL_EXPORTER_OTLP_HEADERS
      valueFrom:
        secretKeyRef:
          name: otel-headers
          key: headers
    - name: API_KEY
      valueFrom:
        secretKeyRef:
          name: api-credentials
          key: key
```

## Common Use Cases

### OpenTelemetry Authentication

Configure authenticated OTEL export:

```shell
# Create secret with authentication headers
kubectl create secret generic otel-auth \
  --from-literal=headers='authorization=Bearer eyJhbGc...' \
  -n envoy-ai-gateway-system

# Use in environment configuration
OTEL_EXPORTER_OTLP_ENDPOINT=https://collector.example.com;OTEL_EXPORTER_OTLP_HEADERS=secretKeyRef {"name":"otel-auth","key":"headers"}
```

### External Service Integration

Reference API keys for external services:

```yaml
extProc:
  extraEnvVars:
    - name: EXTERNAL_SERVICE_URL
      value: "https://api.example.com"
    - name: EXTERNAL_SERVICE_KEY
      valueFrom:
        secretKeyRef:
          name: external-api
          key: api-key
```

### Debug Configuration

Combine public and private configuration:

```shell
# Public configuration with private debug endpoint
DEBUG_MODE=true;DEBUG_ENDPOINT=secretKeyRef {"name":"debug-config","key":"endpoint"}
```

## Error Handling

### Invalid JSON Format

If the JSON in secretKeyRef is malformed, the environment variable parsing will fail:

```shell
# ❌ Invalid - unquoted keys
OTEL_HEADERS=secretKeyRef {name:"secret",key:"key"}

# ✅ Valid - properly quoted
OTEL_HEADERS=secretKeyRef {"name":"secret","key":"key"}
```

### Missing Secret or Key

If the referenced secret or key doesn't exist, the pod will fail to start with an error indicating the missing secret.

### Empty Name or Key

The secret name and key cannot be empty:

```shell
# ❌ Invalid - empty name
OTEL_HEADERS=secretKeyRef {"name":"","key":"headers"}

# ❌ Invalid - missing key
OTEL_HEADERS=secretKeyRef {"name":"otel-secret"}
```

## Best Practices

### Security

- Always use secret references for sensitive data like API keys, tokens, and credentials
- Never commit secrets to version control; use Kubernetes secrets or external secret management
- Regularly rotate secrets and update references accordingly
- Use appropriate RBAC to control access to secrets

### Organization

- Group related configuration into the same secret when possible
- Use descriptive secret and key names
- Document secret dependencies in your deployment documentation

### Validation

- Test your configuration in a development environment before deploying to production
- Verify that all referenced secrets exist and contain the expected data
- Monitor pod startup logs for configuration errors

## See Also

- [Upstream Authentication](./upstream-auth.mdx) - Configure authentication to AI providers
- [GenAI Tracing](../observability/tracing.md) - OpenTelemetry configuration examples
- [Kubernetes Secrets Documentation](https://kubernetes.io/docs/concepts/configuration/secret/)