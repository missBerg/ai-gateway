---
id: operations
title: Deployment and Operations
sidebar_position: 6
---

# Deployment and Operations

This section provides guidance for deploying and operating Envoy AI Gateway in production environments.


### [High Availability and Scaling](./high-availability.md)

Learn considersations for configuring the AI Gateway controller for high availability and horizontal scaling.

**Essential reading for production deployments** - addresses the dual-responsibility architecture of the AI Gateway controller and why multiple replicas are recommended.

### [Deployment Architectures](./deployment-architectures.md)

Explore different architectural patterns for deploying Envoy AI Gateway.


## Deployment Considerations

When planning your AI Gateway deployment, consider:

1. **Scale Requirements**: How many requests per second? How many models and providers?
2. **Availability Needs**: What are your SLA requirements? Do you need multi-region deployment?
3. **Integration Complexity**: Are you using external providers only, or do you have self-hosted models?
4. **Security Posture**: What are your authentication, authorization, and compliance requirements?
5. **Cost Management**: How will you track and control AI/LLM costs across your organization?

## Getting Help

If you need assistance with deployment or operations:

- Review the [Architecture Documentation](../concepts/architecture/) to understand the system design
- Check the [Observability Guides](../capabilities/observability/) for monitoring and troubleshooting
- Visit the [GitHub repository](https://github.com/envoyproxy/ai-gateway) for community support
- Review [GitHub Issues](https://github.com/envoyproxy/ai-gateway/issues) for known operational considerations

## Next Steps

Start with the [High Availability and Scaling](./high-availability.md) guide to ensure your controller is properly configured for production workloads, then explore the architectural patterns that best fit your use case.
