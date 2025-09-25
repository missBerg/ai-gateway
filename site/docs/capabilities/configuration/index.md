---
id: configuration
title: Configuration
---

# Configuration

This section covers advanced configuration options for the Envoy AI Gateway beyond basic setup and provider connections.

## Available Configuration Topics

- **[External Processor Configuration](./external-processor.md)** - Configure the external processor component using environment variables, including secure secret references for sensitive data like API keys and authentication tokens.

## Overview

The Envoy AI Gateway provides flexible configuration options to customize its behavior for your specific use case. Configuration typically involves:

1. **Helm Chart Values** - Primary configuration method using Helm values
2. **Environment Variables** - Runtime configuration for components like the external processor
3. **Kubernetes Secrets** - Secure storage and reference of sensitive configuration data
4. **Custom Resource Definitions** - AI Gateway specific resources like AIGatewayRoute and AIServiceBackend

## Configuration Best Practices

### Security
- Use Kubernetes secrets for sensitive configuration values
- Avoid storing credentials directly in Helm values or environment variables
- Implement proper RBAC for secret access
- Regularly rotate sensitive credentials

### Maintainability
- Document custom configuration changes
- Use version control for Helm values files
- Test configuration changes in development environments
- Monitor configuration-related metrics and logs

### Performance
- Configure appropriate resource limits and requests
- Set reasonable timeout values for external integrations
- Monitor resource usage and adjust as needed

## Getting Help

If you need assistance with configuration:

1. Check the specific configuration documentation for your use case
2. Review the [API Reference](/docs/api/api) for detailed field descriptions
3. Look at the [examples directory](https://github.com/envoyproxy/ai-gateway/tree/main/examples) in the repository
4. Join the [Envoy Slack community](https://envoyproxy.slack.com/) for support