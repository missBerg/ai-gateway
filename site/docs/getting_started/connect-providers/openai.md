---
id: openai
title: Connect OpenAI
sidebar_position: 2
---

# Connect OpenAI

This guide will help you configure Envoy AI Gateway to work with OpenAI's models.

## Prerequisites

Before you begin, you'll need:
- An OpenAI API key from [OpenAI's platform](https://platform.openai.com)
- Basic setup completed from the [Basic Usage](../basic-usage.md) guide
- Basic configuration removed as described in the [Advanced Configuration](./index.md) overview

## Configuration Steps

### 1. Download Configuration Template

First, download the basic configuration template:

```shell
curl -O https://raw.githubusercontent.com/envoyproxy/ai-gateway/main/examples/basic/basic.yaml
```

### 2. Configure OpenAI Credentials

Edit the `basic.yaml` file to replace the OpenAI placeholder value:
- Find the section containing `OPENAI_API_KEY`
- Replace it with your actual OpenAI API key

:::caution Security Note
Make sure to keep your API key secure and never commit it to version control.
The key will be stored in a Kubernetes secret.
:::

### 3. Apply Configuration

Apply the updated configuration and wait for the Gateway pod to be ready:

```shell
kubectl apply -f basic.yaml

kubectl wait pods --timeout=2m \
  -l gateway.envoyproxy.io/owning-gateway-name=envoy-ai-gateway-basic \
  -n envoy-gateway-system \
  --for=condition=Ready
```

### 4. Test the Configuration

Set up port forwarding (this will block the terminal):

```shell
export ENVOY_SERVICE=$(kubectl get svc -n envoy-gateway-system \
  --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=envoy-ai-gateway-basic \
  -o jsonpath='{.items[0].metadata.name}')

kubectl port-forward -n envoy-gateway-system svc/$ENVOY_SERVICE 8080:80
```

In a new terminal, send a test request to OpenAI:

```shell
curl --fail \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [
      {
        "role": "user",
        "content": "Hi."
      }
    ]
  }' \
  http://localhost:8080/v1/chat/completions
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
   kubectl logs -n envoy-gateway-system deployment/ai-gateway-controller
   ```
4. Common errors:
   - 401: Invalid API key
   - 429: Rate limit exceeded
   - 503: OpenAI service unavailable

## Next Steps

After configuring OpenAI:
- [Connect AWS Bedrock](./aws-bedrock.md) to add another provider
- Learn about [rate limiting](../../configuration/rate-limiting.md)
- Configure [model routing](../../configuration/routing.md) 
