---
id: environment-variables
title: Environment Variables
sidebar_position: 1
---

# Environment Variables Configuration

AI Gateway uses environment variables to configure various components. This guide explains how to set environment variables and use Kubernetes secrets for sensitive configuration.

## Setting Environment Variables

There are several ways to set environment variables for AI Gateway components:

### Using Helm Values

The most common way to set environment variables is through Helm values when installing or upgrading the AI Gateway:

```yaml
# values.yaml
extProc:
  extraEnvVars:
    - name: ENVIRONMENT_VARIABLE_NAME
      value: "environment_variable_value"
    - name: ANOTHER_VARIABLE
      value: "another_value"
```

Apply these values during installation or upgrade:

```shell
helm upgrade ai-eg oci://docker.io/envoyproxy/ai-gateway-helm \
  --version v0.0.0-latest \
  --namespace envoy-ai-gateway-system \
  -f values.yaml
```

### Using Command-Line Arguments

You can also set environment variables directly in the Helm command:

```shell
helm upgrade ai-eg oci://docker.io/envoyproxy/ai-gateway-helm \
  --version v0.0.0-latest \
  --namespace envoy-ai-gateway-system \
  --set "extProc.extraEnvVars[0].name=ENVIRONMENT_VARIABLE_NAME" \
  --set "extProc.extraEnvVars[0].value=environment_variable_value"
```

## Using Kubernetes Secrets for Sensitive Data

For sensitive configuration like API keys, authentication tokens, and credentials, AI Gateway supports referencing Kubernetes secrets in environment variables.

### Creating a Kubernetes Secret

First, create a Kubernetes secret to store your sensitive data:

```shell
kubectl create secret generic my-secret \
  --from-literal=api-key="your-api-key-here" \
  --from-literal=auth-token="your-auth-token-here"
```

### Referencing Secrets in Helm Values

You can reference these secrets in your Helm values file:

```yaml
extProc:
  extraEnvVars:
    - name: API_KEY
      valueFrom:
        secretKeyRef:
          name: my-secret
          key: api-key
    - name: AUTH_TOKEN
      valueFrom:
        secretKeyRef:
          name: my-secret
          key: auth-token
```

### Using the `secretKeyRef` Syntax

Alternatively, you can use the `secretKeyRef` syntax directly in command-line arguments:

```shell
helm upgrade ai-eg oci://docker.io/envoyproxy/ai-gateway-helm \
  --version v0.0.0-latest \
  --namespace envoy-ai-gateway-system \
  --set "extProc.extraEnvVars[0].name=API_KEY" \
  --set "extProc.extraEnvVars[0].value=secretKeyRef {\"name\":\"my-secret\",\"key\":\"api-key\"}"
```

The `secretKeyRef` syntax requires valid JSON with both `name` and `key` properties specified. The format is:
```
secretKeyRef {"name":"SECRET_NAME","key":"KEY"}
```

## Common Environment Variables

### OpenTelemetry Configuration

For tracing and observability, you can configure OpenTelemetry using these environment variables:

```yaml
extProc:
  extraEnvVars:
    - name: OTEL_SERVICE_NAME
      value: "ai-gateway"
    - name: OTEL_EXPORTER_OTLP_ENDPOINT
      value: "http://collector:4317"
    - name: OTEL_EXPORTER_OTLP_HEADERS
      valueFrom:
        secretKeyRef:
          name: otel-headers
          key: auth-header
    - name: OTEL_METRICS_EXPORTER
      value: "none"
```

### OpenInference Configuration

For controlling sensitive data in traces:

```yaml
extProc:
  extraEnvVars:
    - name: OPENINFERENCE_HIDE_INPUTS
      value: "true"  # Hide input messages to the LLM
    - name: OPENINFERENCE_HIDE_OUTPUTS
      value: "true"  # Hide output messages from the LLM
```

## Multiple Environment Variables

You can set multiple environment variables at once using semicolons (`;`) as separators. This is useful for command-line setting of multiple variables:

```shell
helm upgrade ai-eg oci://docker.io/envoyproxy/ai-gateway-helm \
  --version v0.0.0-latest \
  --namespace envoy-ai-gateway-system \
  --set "extProc.extraEnvVars[0].name=ENV_VARS" \
  --set "extProc.extraEnvVars[0].value=VAR1=value1;VAR2=value2;VAR3=secretKeyRef {\"name\":\"my-secret\",\"key\":\"secret-value\"}"
```

This sets three environment variables: `VAR1`, `VAR2`, and `VAR3`, with the third one referencing a Kubernetes secret.

## See Also

- [Tracing Configuration](../observability/tracing.md) - OpenTelemetry and tracing setup
- [Upstream Authentication](../security/upstream-auth.mdx) - Security and authentication configuration