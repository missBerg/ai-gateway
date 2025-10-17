---
id: deployment-architectures
title: Deployment Architectures
sidebar_position: 2
---

# Deployment Architectures

Envoy AI Gateway supports flexible deployment patterns to match your organization's needs. This guide explores different architectural approaches, from simple single-gateway setups to enterprise-grade two-tier deployments.

:::tip Reference Architecture
This guide is based on the [Reference Architecture for Adopters of Envoy AI Gateway](https://gateway.envoyproxy.io/blog/envoy-ai-gateway-reference-architecture) blog post.
:::

## Choosing Your Architecture

The right architecture depends on several factors:

| Factor | Single-Tier | Two-Tier |
|--------|-------------|----------|
| **Use Case** | Simple provider aggregation | Complex multi-team platform |
| **Model Hosting** | External providers only | Mixed external + self-hosted |
| **Team Structure** | Single team | Multiple teams with autonomy |
| **Scale** | Low to medium | Medium to high |
| **Operational Complexity** | Lower | Higher (but more flexible) |

## Single-Tier Architecture

A single-tier deployment is the simplest approach, ideal for organizations starting with AI Gateway or those using only external model providers.

### Architecture Overview

```
┌─────────────┐
│   Clients   │
│  (Apps/SDKs)│
└──────┬──────┘
       │
       ▼
┌─────────────────────────┐
│   Envoy AI Gateway      │
│   (Single Gateway)      │
└────┬────────────┬───────┘
     │            │
     ▼            ▼
┌─────────┐  ┌─────────┐
│ OpenAI  │  │ Bedrock │
└─────────┘  └─────────┘
```

### When to Use

- **Getting Started**: Quick setup to explore AI Gateway capabilities
- **External Providers Only**: All models are hosted by third-party services (OpenAI, Anthropic, AWS Bedrock, etc.)
- **Single Team**: One platform team manages all AI infrastructure
- **Simpler Operations**: Fewer components to manage and monitor

### Configuration Example

A basic single-tier deployment with multiple providers:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: ai-gateway
  namespace: default
spec:
  gatewayClassName: envoy-ai-gateway
  listeners:
    - name: http
      protocol: HTTP
      port: 80
---
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGatewayRoute
metadata:
  name: multi-provider-route
  namespace: default
spec:
  parentRefs:
    - name: ai-gateway
  rules:
    # Route to OpenAI
    - matches:
        - headers:
            - name: x-ai-eg-model
              value: gpt-4
      backendRefs:
        - name: openai-backend
    # Route to AWS Bedrock
    - matches:
        - headers:
            - name: x-ai-eg-model
              value: claude-3-sonnet
      backendRefs:
        - name: bedrock-backend
```

### Benefits

- **Simplicity**: One gateway to manage
- **Fast Setup**: Quick to deploy and configure
- **Unified API**: Single endpoint for all AI models
- **Centralized Policies**: Authentication, rate limiting, and observability in one place

### Limitations

- Less flexibility for self-hosted model management
- Harder to delegate operational control to multiple teams
- Single point of policy enforcement (less granular control)

## Two-Tier Gateway Architecture

The two-tier architecture separates concerns between external API access (Tier 1) and internal model serving (Tier 2), providing enterprise-grade flexibility and scale.

### Architecture Overview

```
┌─────────────┐
│   Clients   │
│  (Apps/SDKs)│
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────┐
│      Tier 1: Gateway Cluster        │
│   (Centralized Entry Point)         │
└──┬──────────────┬──────────────┬────┘
   │              │              │
   ▼              ▼              ▼
┌──────┐   ┌──────────┐   ┌─────────────────┐
│OpenAI│   │Anthropic │   │  Tier 2 Gateway │
└──────┘   └──────────┘   │ (Model Serving) │
                           └────────┬────────┘
                                    │
                             ┌──────┴──────┐
                             ▼             ▼
                        ┌────────┐   ┌────────┐
                        │KServe  │   │KServe  │
                        │Model A │   │Model B │
                        └────────┘   └────────┘
```

### Tier 1: Gateway Cluster

The Tier 1 gateway serves as the centralized API entry point for all client applications.

#### Responsibilities

- **Unified API**: Single endpoint for all AI models (external and internal)
- **External Provider Routing**: Direct routing to OpenAI, Anthropic, AWS Bedrock, GCP Vertex, etc.
- **Internal Gateway Routing**: Routes to Tier 2 gateways for self-hosted models
- **Global Policies**: Top-level authentication, rate limiting, and cost tracking
- **Credential Management**: Centralized secret injection for external providers

#### Benefits

- Client applications use one API regardless of where models are hosted
- Platform team controls the main entry point and global policies
- Simplified developer experience (no need to know about model locations)
- Centralized observability and cost tracking

#### Configuration Example

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGatewayRoute
metadata:
  name: tier1-gateway
  namespace: gateway-cluster
spec:
  parentRefs:
    - name: tier1-gateway
  rules:
    # External provider: OpenAI
    - matches:
        - headers:
            - name: x-ai-eg-model
              value: gpt-4
      backendRefs:
        - name: openai-backend

    # External provider: AWS Bedrock
    - matches:
        - headers:
            - name: x-ai-eg-model
              value: claude-3-5-sonnet
      backendRefs:
        - name: bedrock-backend

    # Internal model: Route to Tier 2 gateway
    - matches:
        - headers:
            - name: x-ai-eg-model
              value: llama-3-70b
      backendRefs:
        - name: tier2-gateway
          namespace: model-serving-cluster
          group: gateway.networking.k8s.io
          kind: Gateway
```

### Tier 2: Model Serving Cluster

The Tier 2 gateway is deployed within a self-hosted model serving cluster, typically alongside KServe.

#### Responsibilities

- **Internal Traffic Routing**: Load balancing across self-hosted model instances
- **Model-Specific Policies**: Fine-grained rate limiting, quotas, and access control
- **Version Management**: A/B testing, canary deployments, model versioning
- **Model Serving Integration**: Works with KServe, vLLM, or other serving platforms

#### Benefits

- Platform/ML teams have autonomy over internal model operations
- Changes to self-hosted models don't affect Tier 1 gateway
- Fine-grained control over model serving policies
- Decoupled scaling and lifecycle management

#### Configuration Example

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGatewayRoute
metadata:
  name: tier2-model-routing
  namespace: model-serving-cluster
spec:
  parentRefs:
    - name: tier2-gateway
  rules:
    # Route to KServe-hosted Llama model
    - matches:
        - headers:
            - name: x-ai-eg-model
              value: llama-3-70b
      backendRefs:
        - name: llama-3-70b-kserve
          namespace: model-serving

    # Support for multiple versions (A/B testing)
    - matches:
        - headers:
            - name: x-ai-eg-model
              value: llama-3-70b-experimental
      backendRefs:
        - name: llama-3-70b-v2-kserve
          namespace: model-serving
```

### Integration with KServe

[KServe](https://kserve.github.io/website/latest/) is a model serving platform that automates the deployment of production-ready ML inference endpoints.

#### KServe Benefits

- **Autoscaling**: Token-based autoscaling for LLMs, scale-to-zero for cost savings
- **Multi-node Inference**: Distributed inference via vLLM
- **OpenAI-Compatible APIs**: Drop-in replacement for OpenAI SDK
- **Model Caching**: Built-in support for model and prompt caching
- **Advanced Runtimes**: Integration with vLLM, TorchServe, and more

#### Example Integration

Deploy a model with KServe:

```yaml
apiVersion: serving.kserve.io/v1beta1
kind: InferenceService
metadata:
  name: llama-3-70b
  namespace: model-serving
spec:
  predictor:
    model:
      modelFormat:
        name: vllm
      runtime: kserve-vllm-runtime
      storageUri: s3://my-models/llama-3-70b
      resources:
        requests:
          nvidia.com/gpu: 4
        limits:
          nvidia.com/gpu: 4
```

Route to it from Tier 2 gateway:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIServiceBackend
metadata:
  name: llama-3-70b-kserve
  namespace: model-serving
spec:
  schema:
    name: OpenAI
  backendRef:
    name: llama-3-70b-predictor
    kind: Service
```

### Two-Tier Architecture Benefits

The two-tier design provides several key advantages:

#### 1. Separation of Concerns

- **External Access** (Tier 1): Focused on client-facing API and global policies
- **Internal Operations** (Tier 2): Focused on model serving and internal routing

#### 2. Team Autonomy

- Platform team controls Tier 1 (stable, customer-facing)
- ML/Model teams control Tier 2 (rapid iteration, experimentation)
- Changes in one tier don't require changes in the other

#### 3. Security and Governance

- Centralized credential management at Tier 1
- Fine-grained access control at Tier 2
- Network segmentation between clusters

#### 4. Scalability

- Scale Tier 1 based on client request volume
- Scale Tier 2 based on model serving capacity
- Independent scaling decisions for each tier

#### 5. Cost Optimization

- Centralized tracking and rate limiting at Tier 1
- Model-specific optimizations at Tier 2
- Clear cost attribution per model/provider

## Hybrid Deployments

You can start with a single-tier architecture and evolve to two-tier as needs grow:

### Evolution Path

1. **Phase 1**: Single-tier with external providers only
2. **Phase 2**: Add self-hosted models behind the same gateway
3. **Phase 3**: Extract model serving to dedicated Tier 2 gateway
4. **Phase 4**: Scale Tier 2 across multiple clusters/regions

### Migration Strategy

- Keep client applications unchanged (same API)
- Add Tier 2 gateway in parallel
- Gradually migrate model routing from Tier 1 to Tier 1 → Tier 2
- No client code changes required

## Multi-Region and Multi-Cloud

Both architectures can be extended for multi-region or multi-cloud deployments:

### Multi-Region Pattern

Deploy gateways in multiple regions:

- **Primary Region**: Tier 1 + Tier 2 in primary data center
- **Secondary Regions**: Tier 2 gateways for locality
- **Global Load Balancing**: DNS-based routing to nearest Tier 1

### Multi-Cloud Pattern

- **Tier 1**: Deployed in primary cloud provider
- **Tier 2**: Distributed across multiple cloud providers
- **Hybrid Models**: Some models on AWS, others on GCP or on-prem

## Decision Guide

Use this guide to choose your architecture:

### Start with Single-Tier if:

- ✅ You're using only external model providers
- ✅ You have a single platform team managing AI infrastructure
- ✅ Operational simplicity is a priority
- ✅ You're in the exploration/proof-of-concept phase

### Adopt Two-Tier if:

- ✅ You have self-hosted models (KServe, custom endpoints)
- ✅ Multiple teams need autonomy over model operations
- ✅ You need fine-grained policy control per model
- ✅ Scale and complexity justify the additional infrastructure
- ✅ You want to separate customer-facing from internal operations

## Next Steps

- Review [High Availability and Scaling](./high-availability.md) to configure your gateways for production
- Explore [Production Best Practices](./production-best-practices.md) for operational excellence
- Check [Connect Providers](../getting-started/connect-providers/) to integrate with external AI services
- Learn about [Model Virtualization](../capabilities/traffic/model-virtualization.md) for advanced routing patterns
- See the [KServe Documentation](https://kserve.github.io/website/latest/) for self-hosted model serving
