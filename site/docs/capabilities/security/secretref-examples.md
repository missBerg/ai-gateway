---
id: secretref-examples
title: SecretKeyRef Examples
sidebar_position: 10
---

# SecretKeyRef Configuration Examples

This page provides practical examples of using the `secretKeyRef` syntax for secure environment variable configuration in Envoy AI Gateway.

## Prerequisites

Before using these examples, ensure you have:
- A running Kubernetes cluster with Envoy AI Gateway installed
- Appropriate RBAC permissions to create and read secrets
- Understanding of [Environment Variable Configuration](./environment-configuration.md)

## Example 1: Authenticated OpenTelemetry Export

Configure OTEL tracing with authentication headers stored in a secret.

### Step 1: Create the secret

```shell
kubectl create secret generic otel-auth \
  --from-literal=headers='authorization=Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...' \
  -n envoy-ai-gateway-system
```

### Step 2: Configure the AI Gateway

**Using Helm values:**

```yaml
# values.yaml
extProc:
  extraEnvVars:
    - name: OTEL_EXPORTER_OTLP_ENDPOINT
      value: "https://phoenix.example.com:6006"
    - name: OTEL_EXPORTER_OTLP_HEADERS
      valueFrom:
        secretKeyRef:
          name: otel-auth
          key: headers
    - name: OTEL_TRACES_EXPORTER
      value: "otlp"
```

**Using CLI with aigw run:**

```shell
export OTEL_CONFIG="OTEL_EXPORTER_OTLP_ENDPOINT=https://phoenix.example.com:6006;OTEL_EXPORTER_OTLP_HEADERS=secretKeyRef {\"name\":\"otel-auth\",\"key\":\"headers\"};OTEL_TRACES_EXPORTER=otlp"
aigw run --env "$OTEL_CONFIG"
```

## Example 2: Multiple Provider API Keys

Store multiple API keys for different AI providers.

### Step 1: Create provider secrets

```shell
# OpenAI API key
kubectl create secret generic openai-creds \
  --from-literal=api-key='sk-proj-...' \
  -n envoy-ai-gateway-system

# Custom provider credentials
kubectl create secret generic custom-provider \
  --from-literal=api-key='cp_12345...' \
  --from-literal=endpoint='https://api.customprovider.com' \
  -n envoy-ai-gateway-system
```

### Step 2: Configure multiple environment variables

```yaml
# values.yaml
extProc:
  extraEnvVars:
    - name: OPENAI_API_KEY
      valueFrom:
        secretKeyRef:
          name: openai-creds
          key: api-key
    - name: CUSTOM_PROVIDER_KEY
      valueFrom:
        secretKeyRef:
          name: custom-provider
          key: api-key
    - name: CUSTOM_PROVIDER_ENDPOINT
      valueFrom:
        secretKeyRef:
          name: custom-provider
          key: endpoint
```

## Example 3: Complex OTEL Headers

Configure complex OTEL headers with multiple authentication and custom headers.

### Step 1: Create complex headers secret

```shell
# Create a secret with complex header configuration
kubectl create secret generic otel-complex \
  --from-literal=headers='authorization=Bearer token123,x-api-key=key456,x-tenant-id=tenant789' \
  -n envoy-ai-gateway-system
```

### Step 2: Use in configuration

**CLI format:**
```shell
OTEL_EXPORTER_OTLP_HEADERS=secretKeyRef {"name":"otel-complex","key":"headers"} aigw run
```

**Helm format:**
```yaml
extProc:
  extraEnvVars:
    - name: OTEL_EXPORTER_OTLP_HEADERS
      valueFrom:
        secretKeyRef:
          name: otel-complex
          key: headers
```

## Example 4: Development vs Production Configuration

Manage different configurations for different environments.

### Step 1: Create environment-specific secrets

```shell
# Development environment
kubectl create secret generic otel-dev \
  --from-literal=endpoint='http://localhost:4317' \
  --from-literal=headers='x-environment=dev' \
  -n envoy-ai-gateway-system

# Production environment  
kubectl create secret generic otel-prod \
  --from-literal=endpoint='https://otel.prod.example.com:4317' \
  --from-literal=headers='authorization=Bearer prod-token,x-environment=prod' \
  -n envoy-ai-gateway-system
```

### Step 2: Use environment-appropriate secrets

**Development:**
```yaml
extProc:
  extraEnvVars:
    - name: OTEL_EXPORTER_OTLP_ENDPOINT
      valueFrom:
        secretKeyRef:
          name: otel-dev
          key: endpoint
    - name: OTEL_EXPORTER_OTLP_HEADERS
      valueFrom:
        secretKeyRef:
          name: otel-dev
          key: headers
```

**Production:**
```yaml
extProc:
  extraEnvVars:
    - name: OTEL_EXPORTER_OTLP_ENDPOINT
      valueFrom:
        secretKeyRef:
          name: otel-prod
          key: endpoint
    - name: OTEL_EXPORTER_OTLP_HEADERS
      valueFrom:
        secretKeyRef:
          name: otel-prod
          key: headers
```

## Example 5: External Secret Management Integration

Integrate with external secret management systems like HashiCorp Vault or AWS Secrets Manager using the [External Secrets Operator](https://external-secrets.io/).

### Step 1: Configure External Secrets Operator

```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: vault-secret-store
  namespace: envoy-ai-gateway-system
spec:
  provider:
    vault:
      server: "https://vault.example.com"
      path: "secret"
      auth:
        kubernetes:
          mountPath: "kubernetes"
          role: "ai-gateway-role"
```

### Step 2: Create ExternalSecret

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: otel-external-auth
  namespace: envoy-ai-gateway-system
spec:
  refreshInterval: 5m
  secretStoreRef:
    name: vault-secret-store
    kind: SecretStore
  target:
    name: otel-auth-external
    creationPolicy: Owner
  data:
  - secretKey: headers
    remoteRef:
      key: ai-gateway/otel
      property: headers
```

### Step 3: Reference the externally-managed secret

```yaml
extProc:
  extraEnvVars:
    - name: OTEL_EXPORTER_OTLP_HEADERS
      valueFrom:
        secretKeyRef:
          name: otel-auth-external
          key: headers
```

## Troubleshooting

### Common Issues

**Secret not found:**
```
Error: couldn't find key headers in Secret envoy-ai-gateway-system/otel-auth
```

**Solution:** Verify the secret exists and contains the expected key:
```shell
kubectl get secret otel-auth -n envoy-ai-gateway-system -o yaml
```

**Invalid JSON in CLI:**
```shell
# ❌ Wrong - unquoted keys
OTEL_HEADERS=secretKeyRef {name:"secret",key:"key"}

# ✅ Correct - properly quoted JSON
OTEL_HEADERS=secretKeyRef {"name":"secret","key":"key"}
```

**Empty values:**
Check that your secret actually contains data:
```shell
kubectl get secret otel-auth -n envoy-ai-gateway-system -o jsonpath='{.data}' | base64 -d
```

### Validation Commands

**Test secret reference:**
```shell
# Check if secret exists
kubectl get secret otel-auth -n envoy-ai-gateway-system

# View secret data (base64 encoded)
kubectl get secret otel-auth -n envoy-ai-gateway-system -o yaml

# Decode secret value
kubectl get secret otel-auth -n envoy-ai-gateway-system -o jsonpath='{.data.headers}' | base64 -d
```

**Verify pod configuration:**
```shell
# Check environment variables in running pod
kubectl get pods -n envoy-gateway-system -l gateway.envoyproxy.io/owning-gateway-name=envoy-ai-gateway-basic
kubectl exec -it <pod-name> -n envoy-gateway-system -- env | grep OTEL
```

## Best Practices

1. **Use descriptive secret names** that clearly indicate their purpose
2. **Group related configuration** in the same secret when possible
3. **Regularly rotate secrets** and update references accordingly
4. **Use appropriate RBAC** to restrict access to sensitive secrets
5. **Document secret dependencies** in your deployment procedures
6. **Test in development** before applying to production environments
7. **Monitor secret expiration** and renewal processes

## See Also

- [Environment Variable Configuration](./environment-configuration.md) - Complete reference guide
- [Upstream Authentication](./upstream-auth.mdx) - Provider-specific authentication
- [GenAI Tracing](../observability/tracing.md) - OpenTelemetry configuration
- [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/) - Official documentation