---
id: capabilities
title: Capabilities
sidebar_position: 3
---

# Envoy AI Gateway Capabilities

Welcome to the Envoy AI Gateway capabilities documentation! This section provides detailed information about the various features and capabilities that Envoy AI Gateway offers to help you manage and optimize your AI/LLM traffic.

## Configuration

Core configuration options for Envoy AI Gateway:

- **[Environment Variables](./configuration/environment-variables.md)**: Configure AI Gateway components through environment variables and Kubernetes secrets

## LLM Providers Integrations

Support for various Large Language Model providers:

- **[Connecting to AI Providers](./llm-integrations/connect-providers.md)**: Learn how to establish connectivity with any supported AI provider
- **[Supported Providers](./llm-integrations/supported-providers.md)**: Compatible AI/LLM service providers
- **[Supported Endpoints](./llm-integrations/supported-endpoints.md)**: Available API endpoints and operations
- **[Vendor-Specific Fields](./llm-integrations/vendor-specific-fields.md)**: Use backend-specific parameters and access provider-unique capabilities in your OpenAI-compatible requests

## Inference Optimization

Advanced inference optimization capabilities for AI/LLM workloads:

- **[InferencePool Support](./inference/inferencepool-support.md)**: Intelligent routing and load balancing for inference endpoints
- **[HTTPRoute + InferencePool](./inference/httproute-inferencepool.md)**: Basic inference routing with standard Gateway API
- **[AIGatewayRoute + InferencePool](./inference/aigatewayroute-inferencepool.md)**: Advanced AI-specific routing with enhanced features

## Traffic Management

Comprehensive traffic handling and routing capabilities:

- **[Model Virtualization](./traffic/model-virtualization.md)**: Abstract and virtualize AI models
- **[Provider Fallback](./traffic/provider-fallback.md)**: Automatic failover between AI providers
- **[Usage-based Rate Limiting](./traffic/usage-based-ratelimiting.md)**: Token-aware rate limiting for AI workloads

## Security

Robust security features for AI gateway deployments:

- **[Upstream Authentication](./security/upstream-auth.mdx)**: Secure authentication to upstream AI services

## Observability

Monitoring and observability tools for AI workloads:

- **[Metrics](./observability/metrics.md)**: Comprehensive metrics collection and monitoring
