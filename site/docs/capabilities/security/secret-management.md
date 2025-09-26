---
id: secret-management
title: Environment Variable Secrets
sidebar_position: 9
---

# Environment Variable Secrets

## Overview

Envoy AI Gateway supports referencing Kubernetes secrets in environment variables for secure credential management. This allows you to configure sensitive values like API keys, tokens, and other credentials without exposing them in your configuration files.

## Why Use Environment Variable Secrets

Environment variable secrets provide several security benefits:

1. **Secure Storage**: Sensitive values are stored in Kubernetes secrets rather than plain text configuration
2. **Access Control**: Leverage Kubernetes RBAC to control access to secrets
3. **Rotation**: Update secrets without modifying application configuration
4. **Audit Trail**: Track secret access and modifications through Kubernetes audit logs

## Syntax

When configuring environment variables, you can reference Kubernetes secrets using the `secretKeyRef` syntax:

```
VARIABLE_NAME=secretKeyRef {"name":"SECRET_NAME","key":"SECRET_KEY"}
```

**Format requirements:**
- Must start with `secretKeyRef` (case-sensitive)
- JSON object must be valid JSON with quoted keys and values
- `name`: The name of the Kubernetes secret
- `key`: The key within the secret to reference
- Spaces around the JSON are optional

**Valid formats:**
```
# Compact format
OTEL_API_KEY=secretKeyRef{"name":"otel-secret","key":"api-key"}

# Spaced format  
OTEL_API_KEY=secretKeyRef {"name":"otel-secret","key":"api-key"}

# Extra spacing
OTEL_API_KEY=secretKeyRef { "name": "otel-secret", "key": "api-key" }
```

## Examples

### OpenTelemetry Configuration with Secrets

Create a secret containing your OpenTelemetry configuration:

```bash
kubectl create secret generic otel-config \
  --from-literal=endpoint="http://collector.observability:4317" \
  --from-literal=api-key="your-secret-api-key" \
  -n envoy-ai-gateway-system
```

Configure AI Gateway to use the secret values:

```bash
helm upgrade ai-eg oci://docker.io/envoyproxy/ai-gateway-helm \
  --version v0.0.0-latest \
  --namespace envoy-ai-gateway-system \
  --set "extProc.extraEnvVars[0].name=OTEL_EXPORTER_OTLP_ENDPOINT" \
  --set "extProc.extraEnvVars[0].value=secretKeyRef {\"name\":\"otel-config\",\"key\":\"endpoint\"}" \
  --set "extProc.extraEnvVars[1].name=OTEL_EXPORTER_OTLP_HEADERS" \
  --set "extProc.extraEnvVars[1].value=secretKeyRef {\"name\":\"otel-config\",\"key\":\"api-key\"}"
```

### Mixed Environment Variables

You can mix regular string values with secret references:

```bash
helm upgrade ai-eg oci://docker.io/envoyproxy/ai-gateway-helm \
  --version v0.0.0-latest \
  --namespace envoy-ai-gateway-system \
  --set "extProc.extraEnvVars[0].name=OTEL_SERVICE_NAME" \
  --set "extProc.extraEnvVars[0].value=ai-gateway" \
  --set "extProc.extraEnvVars[1].name=OTEL_EXPORTER_OTLP_HEADERS" \
  --set "extProc.extraEnvVars[1].value=secretKeyRef {\"name\":\"otel-config\",\"key\":\"headers\"}"
```

### Using Helm Values File

For complex configurations, use a `values.yaml` file:

```yaml
# values.yaml
extProc:
  extraEnvVars:
    # Regular environment variable
    - name: OTEL_SERVICE_NAME
      value: "ai-gateway"
    
    # Secret reference
    - name: OTEL_EXPORTER_OTLP_ENDPOINT
      value: 'secretKeyRef {"name":"otel-config","key":"endpoint"}'
    
    # Another secret reference
    - name: OTEL_EXPORTER_OTLP_HEADERS
      value: 'secretKeyRef {"name":"otel-config","key":"headers"}'
    
    # Privacy controls
    - name: OPENINFERENCE_HIDE_INPUTS
      value: "false"
```

Apply the configuration:

```bash
helm upgrade ai-eg oci://docker.io/envoyproxy/ai-gateway-helm \
  --version v0.0.0-latest \
  --namespace envoy-ai-gateway-system \
  --values values.yaml
```

## Security Best Practices

### Secret Creation
- Create secrets in the same namespace as the AI Gateway deployment
- Use descriptive secret names that indicate their purpose
- Regularly rotate secret values

### Access Control
- Apply appropriate RBAC policies to limit secret access
- Use least privilege principles for service accounts
- Monitor secret access through audit logs

### Secret Management
- Use external secret management tools (e.g., External Secrets Operator) for production environments
- Implement secret rotation policies
- Avoid storing secrets in version control

## Troubleshooting

### Common Issues

**Invalid JSON format**
```
Error: failed to parse secretKeyRef json: invalid character
```
Solution: Ensure JSON is properly formatted with quoted keys and values.

**Secret not found**
```
Error: secret "otel-config" not found
```
Solution: Verify the secret exists in the correct namespace and the name matches exactly.

**Missing key in secret**
```
Error: key "api-key" not found in secret "otel-config"
```
Solution: Check that the specified key exists in the secret data.

### Validation

Verify your environment variables are properly configured:

```bash
# Check pod environment variables
kubectl describe pod -n envoy-ai-gateway-system -l app=ai-gateway-extproc

# Check secret contents (base64 encoded)
kubectl get secret otel-config -n envoy-ai-gateway-system -o yaml

# Decode secret values
kubectl get secret otel-config -n envoy-ai-gateway-system -o jsonpath='{.data.api-key}' | base64 -d
```

## Related Documentation

- [OpenTelemetry Tracing with Secrets](../observability/tracing.md) - Using secrets for observability configuration
- [Upstream Authentication](./upstream-auth.mdx) - Provider credential management
- [Envoy Gateway Security](https://gateway.envoyproxy.io/docs/tasks/security/) - Additional security configurations