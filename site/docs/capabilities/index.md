---
id: capabilities
title: Capabilities
sidebar_position: 3
---

# Envoy AI Gateway Capabilities

Envoy AI Gateway ships a broad set of features for routing, securing, and observing GenAI traffic. Browse by area below.

## LLM integrations

<DocCardGrid columns={3} cards={[
  {
    title: 'Connect providers',
    href: './llm-integrations/connect-providers',
    description: 'Establish connectivity with any supported AI provider.',
    icon: 'plug'
  },
  {
    title: 'Supported providers',
    href: './llm-integrations/supported-providers',
    description: 'Compatible AI/LLM service providers and endpoint coverage.',
    icon: 'compass'
  },
  {
    title: 'Supported endpoints',
    href: './llm-integrations/supported-endpoints',
    description: 'Available API endpoints and operations.',
    icon: 'radio'
  },
  {
    title: 'Vendor-specific fields',
    href: './llm-integrations/vendor-specific-fields',
    description: 'Pass backend-specific parameters through an OpenAI-compatible request.',
    icon: 'puzzle'
  },
  {
    title: 'Prompt caching',
    href: './llm-integrations/prompt-caching',
    description: 'Provider-agnostic prompt caching via a unified cache_control API.',
    icon: 'zap'
  }
]} />

## Inference optimization

<DocCardGrid columns={3} cards={[
  {
    title: 'InferencePool support',
    href: './inference/inferencepool-support',
    description: 'Intelligent routing and load balancing for inference endpoints.',
    icon: 'brain'
  },
  {
    title: 'HTTPRoute + InferencePool',
    href: './inference/httproute-inferencepool',
    description: 'Basic inference routing with standard Gateway API.',
    icon: 'route'
  },
  {
    title: 'AIGatewayRoute + InferencePool',
    href: './inference/aigatewayroute-inferencepool',
    description: 'Advanced AI-specific routing with enhanced features.',
    icon: 'target'
  }
]} />

## Gateway configuration

<DocCardGrid columns={2} cards={[
  {
    title: 'GatewayConfig',
    href: './gateway-config',
    description: 'Gateway-scoped configuration for the external processor — env vars, resources, shared settings.',
    icon: 'settings'
  },
  {
    title: 'Scaling',
    href: './scaling',
    description: 'Multiple controller replicas and horizontal pod autoscaling for production.',
    icon: 'trending-up'
  }
]} />

## Traffic management

<DocCardGrid columns={2} cards={[
  {
    title: 'Model virtualization',
    href: './traffic/model-name-virtualization',
    description: 'Abstract and virtualize AI models behind a stable identifier.',
    icon: 'copy'
  },
  {
    title: 'Provider fallback',
    href: './traffic/provider-fallback',
    description: 'Automatic failover between AI providers.',
    icon: 'repeat'
  },
  {
    title: 'Usage-based rate limiting',
    href: './traffic/usage-based-ratelimiting',
    description: 'Token-aware rate limiting for AI workloads.',
    icon: 'traffic-cone'
  },
  {
    title: 'Header + body mutations',
    href: './traffic/header-body-mutations',
    description: 'Customize HTTP headers and JSON body fields per backend or route.',
    icon: 'scissors'
  }
]} />

## Security

<DocCardGrid columns={2} cards={[
  {
    title: 'Upstream authentication',
    href: './security/upstream-auth',
    description: 'Secure authentication to upstream AI services.',
    icon: 'lock'
  }
]} />

## Model Context Protocol (MCP)

<DocCardGrid columns={2} cards={[
  {
    title: 'MCP gateway',
    href: './mcp/',
    description: 'Server multiplexing, tool routing, OAuth authentication, and observability for MCP.',
    icon: 'wrench'
  }
]} />

## Observability

<DocCardGrid columns={2} cards={[
  {
    title: 'Metrics',
    href: './observability/metrics',
    description: 'Metrics collection and monitoring for AI workloads.',
    icon: 'bar-chart'
  }
]} />
