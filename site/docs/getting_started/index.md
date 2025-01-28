---
id: getting-started
title: Getting Started
sidebar_position: 2
---

# Getting Started with Envoy AI Gateway

Welcome to the Envoy AI Gateway getting started guide! This guide will walk you through setting up and using Envoy AI Gateway, a powerful tool for managing AI model traffic using Envoy Gateway.

## What is Envoy AI Gateway?

Envoy AI Gateway is an open-source project that leverages Envoy Gateway to handle request traffic between application clients and AI/ML services. It provides:

- Unified interface for multiple AI model providers
- Traffic management and load balancing
- Security and access control
- Monitoring and observability
- Cost management and rate limiting

## Guide Structure

This getting started guide is organized into several sections:

1. [Prerequisites](./prerequisites.md)
   - Setting up your Kubernetes cluster
   - Installing required tools
   - Setting up Envoy Gateway

2. [Installation](./installation.md)
   - Installing Envoy AI Gateway
   - Configuring the gateway
   - Verifying the installation

3. [Basic Usage](./basic-usage.md)
   - Deploying a basic configuration
   - Making your first request
   - Understanding the response format

4. [Connect Providers](./connect-providers.md)
   - Setting up OpenAI integration
   - Configuring AWS Bedrock
   - Managing credentials securely

## Quick Start

If you're familiar with Kubernetes and want to get started quickly, run these commands to install Envoy Gateway, Envoy AI Gateway, and deploy a basic configuration:

```shell
helm upgrade -i eg oci://docker.io/envoyproxy/gateway-helm \
  --version v0.0.0-latest \
  --namespace envoy-gateway-system \
  --create-namespace

helm upgrade -i aieg oci://ghcr.io/envoyproxy/ai-gateway/ai-gateway-helm \
  --version v0.0.0-latest \
  --namespace envoy-ai-gateway-system \
  --create-namespace

kubectl apply -f https://raw.githubusercontent.com/envoyproxy/ai-gateway/main/examples/basic/basic.yaml

kubectl wait --timeout=2m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
kubectl wait --timeout=2m -n envoy-ai-gateway-system deployment/ai-gateway-controller --for=condition=Available
```

:::tip
For detailed instructions and explanations, start with the [Prerequisites](./prerequisites.md) section.
:::

## Need Help?

If you run into any issues:
- Check our [Troubleshooting Guide](../operations/troubleshooting.md)
- Join our [Community Slack](https://envoyproxy.slack.com/archives/C07Q4N24VAA)
- File an issue on [GitHub](https://github.com/envoyproxy/ai-gateway/issues) 
