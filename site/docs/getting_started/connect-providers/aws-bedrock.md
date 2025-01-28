---
id: aws-bedrock
title: Connect AWS Bedrock
sidebar_position: 3
---

# Connect AWS Bedrock

This guide will help you configure Envoy AI Gateway to work with AWS Bedrock's foundation models.

## Prerequisites

Before you begin, you'll need:
- AWS credentials with access to Bedrock
- Basic setup completed from the [Basic Usage](../basic-usage.md) guide
- Basic configuration removed as described in the [Advanced Configuration](./index.md) overview

## AWS Credentials Setup

Ensure you have:
1. An AWS account with Bedrock access enabled
2. AWS credentials with permissions to:
   - `bedrock:InvokeModel`
   - `bedrock:ListFoundationModels`
3. Your AWS access key ID and secret access key

:::tip AWS Best Practices
Consider using AWS IAM roles and limited-scope credentials for production environments.
:::

## Configuration Steps

### 1. Download Configuration Template

First, download the basic configuration template:

```shell
curl -O https://raw.githubusercontent.com/envoyproxy/ai-gateway/main/examples/basic/basic.yaml
```

### 2. Configure AWS Credentials

Edit the `basic.yaml` file to replace these placeholder values:
- `AWS_ACCESS_KEY_ID`: Your AWS access key ID
- `AWS_SECRET_ACCESS_KEY`: Your AWS secret access key

:::caution Security Note
Make sure to keep your AWS credentials secure and never commit them to version control.
The credentials will be stored in Kubernetes secrets.
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

In a new terminal, send a test request to AWS Bedrock:

```shell
curl --fail \
  -H "Content-Type: application/json" \
  -d '{
    "model": "us.meta.llama3-2-1b-instruct-v1:0",
    "messages": [
      {
        "role": "user",
        "content": "Hi."
      }
    ]
  }' \
  http://localhost:8080/v1/chat/completions
```

## AWS Region Configuration

By default, the gateway uses the `us-east-1` region. To use a different region:
1. Add the `AWS_REGION` environment variable in your configuration
2. Update the model name prefix accordingly (e.g., `eu-west-1.meta.llama...`)

## Troubleshooting

If you encounter issues:

1. Verify your AWS credentials are correct and active
2. Check pod status:
   ```shell
   kubectl get pods -n envoy-gateway-system
   ```
3. View controller logs:
   ```shell
   kubectl logs -n envoy-gateway-system deployment/ai-gateway-controller
   ```
4. Common errors:
   - 401/403: Invalid credentials or insufficient permissions
   - 404: Model not found or not available in region
   - 429: Rate limit exceeded

## Next Steps

After configuring AWS Bedrock:
- [Connect OpenAI](./openai.md) to add another provider
- Learn about [rate limiting](../../configuration/rate-limiting.md)
- Configure [model routing](../../configuration/routing.md) 
