---
id: external-processor
title: External Processor Configuration  
sidebar_position: 1
---

# External Processor Configuration

The AI Gateway uses an external processor component to handle request/response transformation and routing logic. You can configure this component using environment variables through the Helm chart's `extProc.extraEnvVars` setting.

## Environment Variable Configuration

Environment variables are provided as a semicolon-delimited string where each variable follows the format `NAME=value`. This allows you to configure various aspects of the external processor behavior.

### Basic Syntax

```yaml
extProc:
  extraEnvVars:
    - name: VARIABLE_NAME
      value: "variable_value"
```

Or using Helm command line:
```bash
helm install ai-eg oci://docker.io/envoyproxy/ai-gateway-helm \
  --set "extProc.extraEnvVars[0].name=VARIABLE_NAME" \
  --set "extProc.extraEnvVars[0].value=variable_value"
```

### Multiple Variables

When configuring multiple environment variables with a single value string, separate them with semicolons:

```bash
--set 'extProc.extraEnvVars[0].value=OTEL_SERVICE_NAME=ai-gateway;OTEL_TRACES_EXPORTER=otlp'
```

## Secret References

For sensitive configuration values, you can reference Kubernetes secrets instead of storing values directly in Helm configuration.

### Secret Reference Syntax

```yaml
VARIABLE_NAME=secretKeyRef {"name":"SECRET_NAME","key":"SECRET_KEY"}
```

**Key points:**
- The JSON must be valid with quoted field names and values
- Spaces around the JSON are optional
- The secret must exist in the same namespace as the AI Gateway
- The external processor will receive the resolved secret value as an environment variable

### Example: OpenTelemetry with Authentication

1. **Create a secret with authentication headers:**
   ```bash
   kubectl create secret generic otel-collector-auth \
     --from-literal=auth-header="Authorization: Bearer your-api-token" \
     -n envoy-ai-gateway-system
   ```

2. **Reference the secret in Helm configuration:**
   ```yaml
   extProc:
     extraEnvVars:
       - name: OTEL_EXPORTER_OTLP_ENDPOINT
         value: "https://otel-collector.example.com:4318"
       - name: OTEL_EXPORTER_OTLP_HEADERS
         value: 'secretKeyRef {"name":"otel-collector-auth","key":"auth-header"}'
   ```

3. **Or using Helm command line:**
   ```bash
   helm upgrade ai-eg oci://docker.io/envoyproxy/ai-gateway-helm \
     --version v0.0.0-latest \
     --namespace envoy-ai-gateway-system \
     --set "extProc.extraEnvVars[0].name=OTEL_EXPORTER_OTLP_ENDPOINT" \
     --set "extProc.extraEnvVars[0].value=https://otel-collector.example.com:4318" \
     --set "extProc.extraEnvVars[1].name=OTEL_EXPORTER_OTLP_HEADERS" \
     --set 'extProc.extraEnvVars[1].value=secretKeyRef {"name":"otel-collector-auth","key":"auth-header"}'
   ```

## Common Configuration Examples

### OpenTelemetry Configuration

**Basic configuration:**
```yaml
extProc:
  extraEnvVars:
    - name: OTEL_EXPORTER_OTLP_ENDPOINT
      value: "http://jaeger:14268/api/traces"
    - name: OTEL_SERVICE_NAME
      value: "ai-gateway"
    - name: OTEL_TRACES_EXPORTER
      value: "otlp"
```

**With authentication:**
```yaml
extProc:
  extraEnvVars:
    - name: OTEL_EXPORTER_OTLP_ENDPOINT
      value: "https://api.honeycomb.io"
    - name: OTEL_EXPORTER_OTLP_HEADERS
      value: 'secretKeyRef {"name":"honeycomb-auth","key":"api-key"}'
```

### Multiple Environment Variables in One Value

You can combine multiple environment variables in a single value string:

```bash
helm install ai-eg oci://docker.io/envoyproxy/ai-gateway-helm \
  --set 'extProc.extraEnvVars[0].value=OTEL_SERVICE_NAME=ai-gateway;OTEL_TRACES_EXPORTER=otlp;OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger:14268'
```

### Mixing Secrets and Plain Values

```bash
helm install ai-eg oci://docker.io/envoyproxy/ai-gateway-helm \
  --set 'extProc.extraEnvVars[0].value=OTEL_SERVICE_NAME=ai-gateway;OTEL_EXPORTER_OTLP_HEADERS=secretKeyRef {"name":"otel-auth","key":"headers"}'
```

## Security Best Practices

1. **Use secrets for sensitive data:** Never store API keys, tokens, or passwords directly in Helm values.

2. **Limit secret access:** Ensure the AI Gateway service account has access only to necessary secrets.

3. **Rotate credentials regularly:** Update secrets periodically and restart the AI Gateway deployment.

4. **Use RBAC:** Configure Kubernetes RBAC to restrict secret access to authorized components only.

## Troubleshooting

### Common Issues

**Invalid JSON format:**
```bash
# ❌ Incorrect - missing quotes around field names
secretKeyRef {name:"secret",key:"key"}

# ✅ Correct - properly quoted JSON
secretKeyRef {"name":"secret","key":"key"}
```

**Secret not found:**
- Verify the secret exists: `kubectl get secret SECRET_NAME -n envoy-ai-gateway-system`
- Check secret key names: `kubectl describe secret SECRET_NAME -n envoy-ai-gateway-system`

**Environment variable not set:**
- Check pod environment variables: `kubectl describe pod -l app=ai-gateway-extproc -n envoy-ai-gateway-system`
- View pod logs: `kubectl logs -l app=ai-gateway-extproc -n envoy-ai-gateway-system`

### Validating Configuration

After deploying, verify your environment variables are correctly set:

```bash
# Check the external processor pod
kubectl get pods -l app=ai-gateway-extproc -n envoy-ai-gateway-system

# Inspect environment variables in the pod
kubectl exec -it POD_NAME -n envoy-ai-gateway-system -- env | grep OTEL
```

## Related Documentation

- [OpenTelemetry Tracing Configuration](../observability/tracing.md)
- [Upstream Authentication](../security/upstream-auth.mdx)
- [OpenTelemetry SDK Environment Variables](https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/)