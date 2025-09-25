---
id: extproc-configuration
title: ExtProc Configuration
sidebar_position: 4
---

# ExtProc Configuration

The Envoy AI Gateway uses an External Processor (ExtProc) container that runs as a sidecar to the Envoy Proxy to handle AI-specific processing logic. This page explains how to configure environment variables for the ExtProc container, including support for Kubernetes secrets.

## Environment Variable Configuration

The ExtProc container can be configured with additional environment variables using the AI Gateway controller's environment variable parsing functionality. This is particularly useful for:

- OpenTelemetry configuration
- Authentication credentials
- Feature flags and processing options
- Custom application settings

### Basic Environment Variables

Environment variables are configured using a semicolon-delimited format:

```yaml
# Example configuration in AI Gateway controller
EXTRA_ENV_VARS: "OTEL_SERVICE_NAME=ai-gateway;OTEL_TRACES_EXPORTER=otlp"
```

This creates the following environment variables in the ExtProc container:
- `OTEL_SERVICE_NAME=ai-gateway`
- `OTEL_TRACES_EXPORTER=otlp`

### Secret References

For sensitive configuration values, you can reference Kubernetes secrets using the `secretKeyRef` syntax:

```yaml
EXTRA_ENV_VARS: "OTEL_EXPORTER_OTLP_HEADERS=secretKeyRef {\"name\":\"otel-secret\",\"key\":\"auth-header\"}"
```

This configuration:
1. References a Kubernetes secret named `otel-secret`
2. Uses the value from the `auth-header` key in that secret
3. Sets the `OTEL_EXPORTER_OTLP_HEADERS` environment variable with the secret value

#### Secret Reference Format

The `secretKeyRef` syntax requires a JSON object with two fields:

```
secretKeyRef {"name":"SECRET_NAME","key":"SECRET_KEY"}
```

- **`name`**: The name of the Kubernetes secret
- **`key`**: The key within the secret to use as the value

**Important Requirements:**
- The JSON must be valid with proper quotes around keys and values
- Both `name` and `key` fields are required
- The secret must exist in the same namespace as the AI Gateway resources
- The AI Gateway controller must have permission to read the referenced secrets

### Mixed Configuration

You can combine regular environment variables with secret references:

```yaml
EXTRA_ENV_VARS: "OTEL_SERVICE_NAME=ai-gateway;OTEL_EXPORTER_OTLP_HEADERS=secretKeyRef {\"name\":\"otel-auth\",\"key\":\"bearer-token\"};DEBUG_LEVEL=info"
```

This creates:
- `OTEL_SERVICE_NAME` with the literal value `ai-gateway`
- `OTEL_EXPORTER_OTLP_HEADERS` with the value from the `otel-auth` secret's `bearer-token` key
- `DEBUG_LEVEL` with the literal value `info`

## Common Use Cases

### OpenTelemetry Authentication

Configure OpenTelemetry exporters with authentication headers stored in secrets:

```yaml
EXTRA_ENV_VARS: "OTEL_EXPORTER_OTLP_ENDPOINT=https://api.honeycomb.io;OTEL_EXPORTER_OTLP_HEADERS=secretKeyRef {\"name\":\"honeycomb-auth\",\"key\":\"api-key\"}"
```

Required secret:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: honeycomb-auth
type: Opaque
data:
  api-key: eC1ob25leWNvbWItdGVhbT1teS10ZWFt...  # base64 encoded API key
```

### Database Connection Strings

For ExtProc components that need database connectivity:

```yaml
EXTRA_ENV_VARS: "DB_CONNECTION_STRING=secretKeyRef {\"name\":\"db-credentials\",\"key\":\"connection-string\"};DB_TIMEOUT=30s"
```

### API Keys for External Services

Configure API keys for external service integrations:

```yaml
EXTRA_ENV_VARS: "THIRD_PARTY_API_KEY=secretKeyRef {\"name\":\"external-apis\",\"key\":\"service-key\"};API_RATE_LIMIT=1000"
```

## Validation and Error Handling

The environment variable parsing includes validation:

- **JSON Format**: Secret references must use valid JSON syntax
- **Required Fields**: Both `name` and `key` must be specified in secret references
- **Empty Values**: Environment variable names cannot be empty
- **Delimiter**: Use semicolons (`;`) to separate multiple environment variables

Invalid configurations will result in controller errors during resource reconciliation.

## Examples

### Complete Configuration Example

```yaml
# AI Gateway controller configuration
EXTRA_ENV_VARS: |
  OTEL_SERVICE_NAME=ai-gateway-extproc;
  OTEL_EXPORTER_OTLP_ENDPOINT=https://otel-collector:4317;
  OTEL_EXPORTER_OTLP_HEADERS=secretKeyRef {"name":"otel-auth","key":"authorization"};
  CUSTOM_FEATURE_FLAG=enabled;
  DB_PASSWORD=secretKeyRef {"name":"database-creds","key":"password"}
```

### Required Secrets

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: otel-auth
type: Opaque
data:
  authorization: QmVhcmVyIG15LWF1dGgtdG9rZW4=  # Bearer my-auth-token

---
apiVersion: v1
kind: Secret
metadata:
  name: database-creds
type: Opaque
data:
  password: bXktc2VjdXJlLXBhc3N3b3Jk  # my-secure-password
```

## Security Considerations

When using secret references:

1. **RBAC Permissions**: Ensure the AI Gateway controller has appropriate permissions to read the referenced secrets
2. **Secret Management**: Follow Kubernetes secret security best practices
3. **Audit Logging**: Monitor access to sensitive configuration values
4. **Rotation**: Implement secret rotation policies for credentials

## Troubleshooting

Common issues and solutions:

### Invalid JSON Format
```
Error: failed to parse secretKeyRef json: invalid character...
```
**Solution**: Ensure proper JSON formatting with quoted keys and values.

### Missing Secret Fields
```
Error: secretKeyRef missing name or key
```
**Solution**: Verify both `name` and `key` fields are present in the JSON object.

### Secret Not Found
Check that:
- The secret exists in the correct namespace
- The AI Gateway controller has read permissions for the secret
- The secret key specified in the configuration exists

## Next Steps

- Learn about [observability configuration](../capabilities/observability/index.md)
- Explore [OpenTelemetry integration](../capabilities/observability/tracing.md)
- Review [security policies](../capabilities/security/index.md)