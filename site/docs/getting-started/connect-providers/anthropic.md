---
id: anthropic
title: Connect Anthropic
sidebar_position: 6
---

import CodeBlock from '@theme/CodeBlock';
import vars from '../../\_vars.json';

# Connect Anthropic

This guide will help you configure Envoy AI Gateway to work with Anthropic's Claude models directly via the Anthropic API.

:::tip Other Ways to Access Claude
You can also access Claude models through cloud providers:
- [AWS Bedrock](./aws-bedrock.md) - Claude via AWS Bedrock
- [GCP VertexAI](./gcp-vertexai.md) - Claude via Google Cloud
:::

## Prerequisites

Before you begin, you'll need:

- An Anthropic API key from [Anthropic's Console](https://console.anthropic.com/)
- Basic setup completed from the [Basic Usage](../basic-usage.md) guide
- Basic configuration removed as described in the [Advanced Configuration](./index.md) overview

## Configuration Steps

:::info Ready to proceed?
Ensure you have followed the steps in [Connect Providers](../connect-providers/)
:::

### 1. Download configuration template

<CodeBlock language="shell">
{`curl -O https://raw.githubusercontent.com/envoyproxy/ai-gateway/${vars.aigwGitRef}/examples/basic/anthropic.yaml`}
</CodeBlock>

### 2. Configure Anthropic Credentials

Edit the `anthropic.yaml` file to replace the Anthropic placeholder value:

- Find the section containing `ANTHROPIC_API_KEY`
- Replace it with your actual Anthropic API key

:::caution Security Note
Make sure to keep your API key secure and never commit it to version control.
The key will be stored in a Kubernetes secret.
:::

### 3. Apply Configuration

Apply the updated configuration and wait for the Gateway pod to be ready. If you already have a Gateway running,
then the secret credential update will be picked up automatically in a few seconds.

```shell
kubectl apply -f anthropic.yaml

kubectl wait pods --timeout=2m \
  -l gateway.envoyproxy.io/owning-gateway-name=envoy-ai-gateway-basic \
  -n envoy-gateway-system \
  --for=condition=Ready
```

### 4. Test the Configuration

You should have set `$GATEWAY_URL` as part of the basic setup before connecting to providers.
See the [Basic Usage](../basic-usage.md) page for instructions.

#### Test with Native Anthropic Messages API

The native Anthropic API provides full access to Claude-specific features:

```shell
curl -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4-5",
    "max_tokens": 100,
    "messages": [
      {
        "role": "user",
        "content": "Hello, Claude!"
      }
    ]
  }' \
  $GATEWAY_URL/anthropic/v1/messages
```

#### Test with OpenAI-Compatible API

You can also use the OpenAI-compatible chat completions endpoint:

```shell
curl -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4-5",
    "messages": [
      {
        "role": "user",
        "content": "Hello, Claude!"
      }
    ]
  }' \
  $GATEWAY_URL/v1/chat/completions
```

#### Test Streaming Responses

Both endpoints support streaming:

```shell
curl -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4-5",
    "max_tokens": 100,
    "stream": true,
    "messages": [
      {
        "role": "user",
        "content": "Tell me a short story."
      }
    ]
  }' \
  $GATEWAY_URL/anthropic/v1/messages
```

## Configuration Details

The Anthropic configuration uses the following key components:

### AIServiceBackend Schema

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIServiceBackend
metadata:
  name: envoy-ai-gateway-basic-anthropic
spec:
  schema:
    name: Anthropic  # Uses native Anthropic API schema
  backendRef:
    name: envoy-ai-gateway-basic-anthropic
    kind: Backend
    group: gateway.envoyproxy.io
```

### BackendSecurityPolicy

Anthropic uses a dedicated `AnthropicAPIKey` type that injects the API key into the `x-api-key` header:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: BackendSecurityPolicy
metadata:
  name: envoy-ai-gateway-basic-anthropic-apikey
spec:
  targetRefs:
    - group: aigateway.envoyproxy.io
      kind: AIServiceBackend
      name: envoy-ai-gateway-basic-anthropic
  type: AnthropicAPIKey
  anthropicAPIKey:
    secretRef:
      name: envoy-ai-gateway-basic-anthropic-apikey
```

## Troubleshooting

If you encounter issues:

1. Verify your API key is correct and active

2. Check pod status:

   ```shell
   kubectl get pods -n envoy-gateway-system
   ```

3. View controller logs:

   ```shell
   kubectl logs -n envoy-ai-gateway-system deployment/ai-gateway-controller
   ```

4. View External Processor Logs

   ```shell
   kubectl logs -n envoy-gateway-system -l gateway.envoyproxy.io/owning-gateway-name=envoy-ai-gateway-basic -c ai-gateway-extproc
   ```

5. Common errors:
   - 401: Invalid API key - verify your `x-api-key` is correct
   - 429: Rate limit exceeded - check your Anthropic usage limits
   - 400: Bad request - ensure `max_tokens` is specified for Anthropic API
   - 503: Anthropic service unavailable

## Configuring More Models

To use more Claude models, add more rules to the AIGatewayRoute. For example, to add Claude Opus:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGatewayRoute
metadata:
  name: envoy-ai-gateway-basic-anthropic
  namespace: default
spec:
  parentRefs:
    - name: envoy-ai-gateway-basic
      kind: Gateway
      group: gateway.networking.k8s.io
  rules:
    - matches:
        - headers:
            - type: Exact
              name: x-ai-eg-model
              value: claude-sonnet-4-5
      backendRefs:
        - name: envoy-ai-gateway-basic-anthropic
    - matches:
        - headers:
            - type: Exact
              name: x-ai-eg-model
              value: claude-opus-4
      backendRefs:
        - name: envoy-ai-gateway-basic-anthropic
```

## Next Steps

After configuring Anthropic:

- [Configure Prompt Caching](../../capabilities/traffic/prompt-caching.md) to reduce costs with Claude's cache control feature
- [Connect AWS Bedrock](./aws-bedrock.md) to add another provider for redundancy
- [Set Up Provider Fallback](../../capabilities/traffic/provider-fallback.md) for high availability
