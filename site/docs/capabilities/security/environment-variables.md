---
id: environment-variables
title: Secure Environment Variables
sidebar_position: 9
---

# Secure Environment Variables

## Overview

The Envoy AI Gateway supports referencing Kubernetes secrets in environment variable configurations to enhance security when handling sensitive data like API keys, tokens, and authentication headers. This prevents credentials from being exposed as plain text in configuration files or Helm values.

## secretKeyRef Support

When configuring the AI Gateway's external processor (ExtProc), you can reference values from Kubernetes secrets using the `secretKeyRef` mechanism in two ways:

### 1. Standard Kubernetes Format

Use the standard Kubernetes `secretKeyRef` format in your Helm values:

```yaml
extProc:
  extraEnvVars:
    - name: OTEL_EXPORTER_OTLP_HEADERS
      valueFrom:
        secretKeyRef:
          name: otel-headers-secret
          key: authorization-header
    - name: CUSTOM_API_KEY
      valueFrom:
        secretKeyRef:
          name: api-credentials
          key: custom-key
    - name: REGULAR_ENV_VAR
      value: "plain-text-value"
```

### 2. Inline secretKeyRef Syntax

For convenience when using the `extraEnvVars` string format (particularly useful with `helm --set`), you can use an inline syntax:

```bash
helm upgrade ai-eg oci://docker.io/envoyproxy/ai-gateway-helm \
  --version v0.0.0-latest \
  --namespace envoy-ai-gateway-system \
  --set 'extProc.extraEnvVars=OTEL_EXPORTER_OTLP_HEADERS=secretKeyRef {"name":"otel-secret","key":"headers"}'
```

#### Supported Inline Formats

The inline `secretKeyRef` syntax supports various JSON formatting styles:

```bash
# Standard format with spaces
OTEL_HEADERS=secretKeyRef {"name":"my-secret","key":"my-key"}

# Compact format without spaces
OTEL_HEADERS=secretKeyRef{"name":"my-secret","key":"my-key"}

# Pretty format with extra spaces
OTEL_HEADERS=secretKeyRef { "name": "my-secret", "key": "my-key" }
```

:::info JSON Format Requirements
The inline format requires valid JSON syntax with quoted field names and values. The `name` and `key` fields are mandatory.
:::

## Common Use Cases

### OpenTelemetry Authentication

Store OpenTelemetry collector authentication headers securely:

```bash
# Create the secret
kubectl create secret generic otel-auth \
  --from-literal=headers="Authorization=Bearer otel-token-12345" \
  -n envoy-ai-gateway-system

# Reference in configuration
helm upgrade ai-eg oci://docker.io/envoyproxy/ai-gateway-helm \
  --version v0.0.0-latest \
  --namespace envoy-ai-gateway-system \
  --set 'extProc.extraEnvVars=OTEL_EXPORTER_OTLP_ENDPOINT=http://collector:4317;OTEL_EXPORTER_OTLP_HEADERS=secretKeyRef {"name":"otel-auth","key":"headers"}'
```

### Custom Provider Credentials

Store custom API keys or tokens:

```bash
# Create the secret
kubectl create secret generic custom-provider-creds \
  --from-literal=api-key="sk-custom-api-key-12345" \
  --from-literal=webhook-secret="webhook-secret-67890" \
  -n envoy-ai-gateway-system

# Reference multiple secrets
helm upgrade ai-eg oci://docker.io/envoyproxy/ai-gateway-helm \
  --version v0.0.0-latest \
  --namespace envoy-ai-gateway-system \
  --set 'extProc.extraEnvVars=CUSTOM_API_KEY=secretKeyRef {"name":"custom-provider-creds","key":"api-key"};WEBHOOK_SECRET=secretKeyRef {"name":"custom-provider-creds","key":"webhook-secret"}'
```

### Mixed Configuration

Combine regular environment variables with secret references:

```bash
helm upgrade ai-eg oci://docker.io/envoyproxy/ai-gateway-helm \
  --version v0.0.0-latest \
  --namespace envoy-ai-gateway-system \
  --set 'extProc.extraEnvVars=OTEL_SERVICE_NAME=ai-gateway;DEBUG_LEVEL=info;OTEL_EXPORTER_OTLP_HEADERS=secretKeyRef {"name":"otel-secret","key":"headers"};CUSTOM_ENDPOINT=https://api.example.com'
```

## Creating Secrets

Use `kubectl` to create secrets with your sensitive data:

```bash
# Single key-value pair
kubectl create secret generic my-secret \
  --from-literal=api-key="your-secret-value" \
  -n envoy-ai-gateway-system

# Multiple key-value pairs
kubectl create secret generic multi-secret \
  --from-literal=key1="value1" \
  --from-literal=key2="value2" \
  -n envoy-ai-gateway-system

# From file
echo -n "your-secret-value" > /tmp/secret-file
kubectl create secret generic file-secret \
  --from-file=api-key=/tmp/secret-file \
  -n envoy-ai-gateway-system
rm /tmp/secret-file
```

## Security Best Practices

### 1. Use Secrets for Sensitive Data
Always store the following types of data in Kubernetes secrets:
- API keys and tokens
- Authentication headers
- Database passwords
- Certificate keys
- Any personally identifiable information (PII)

### 2. Apply Least Privilege Access
Ensure the AI Gateway service account has minimal permissions:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: ai-gateway-secret-reader
  namespace: envoy-ai-gateway-system
rules:
- apiGroups: [""]
  resources: ["secrets"]
  resourceNames: ["otel-auth", "custom-provider-creds"]  # Specific secrets only
  verbs: ["get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: ai-gateway-secret-reader-binding
  namespace: envoy-ai-gateway-system
subjects:
- kind: ServiceAccount
  name: ai-gateway-service-account
roleRef:
  kind: Role
  name: ai-gateway-secret-reader
  apiGroup: rbac.authorization.k8s.io
```

### 3. Implement Secret Rotation
Regularly rotate your secrets and update them in Kubernetes:

```bash
# Update existing secret
kubectl create secret generic otel-auth \
  --from-literal=headers="Authorization=Bearer new-token-12345" \
  --dry-run=client -o yaml | kubectl apply -f -
```

### 4. Monitor Secret Access
Use Kubernetes audit logging to monitor secret access:

```yaml
# In kube-apiserver audit policy
- level: Restricted
  resources:
  - group: ""
    resources: ["secrets"]
  namespaces: ["envoy-ai-gateway-system"]
```

### 5. Use Namespace Isolation
Keep secrets in the same namespace as the AI Gateway to maintain proper isolation and access controls.

## Troubleshooting

### Common Issues

**Error: Secret not found**
```
Error: secret "my-secret" not found
```
- Verify the secret exists: `kubectl get secrets -n envoy-ai-gateway-system`
- Check the secret name spelling in your configuration

**Error: Invalid JSON format**
```
Error: failed to parse secretKeyRef json
```
- Ensure proper JSON syntax with quoted field names and values
- Check for missing commas or brackets

**Error: Missing name or key**
```
Error: secretKeyRef missing name or key
```
- Both `name` and `key` fields are required in the JSON
- Verify the values are not empty strings

### Verification

Check if your configuration is working:

```bash
# Verify the pod has the expected environment variables
kubectl exec -n envoy-gateway-system deployment/envoy-ai-gateway-extproc -c extproc -- env | grep OTEL

# Check pod logs for configuration issues
kubectl logs -n envoy-gateway-system deployment/envoy-ai-gateway-extproc -c extproc
```

## See Also

- [Upstream Authentication](/docs/capabilities/security/upstream-auth) - Provider-specific authentication
- [OpenTelemetry Tracing](/docs/capabilities/observability/tracing) - Using secrets with tracing configuration
- [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/) - Official Kubernetes documentation