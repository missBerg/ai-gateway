---
id: observability
title: Observability
---
Envoy AI Gateway extends the capabilities of Envoy Gateway, and as you run Envoy Gateway you have access to the foundational observability in the Envoy Gateway system.
We recommend you familiarize yourself with the [Envoy Gateway Observability Documentation](https://gateway.envoyproxy.io/docs/tasks/observability/).

## AI/LLM Observability Features

The Envoy AI Gateway provides specialized observability capabilities for AI and LLM workloads:

- **[GenAI Metrics](./metrics.md)** - Prometheus metrics following OpenTelemetry Gen AI semantic conventions for monitoring token usage, latency, and model performance.
- **[GenAI Tracing](./tracing.md)** - OpenTelemetry integration with OpenInference semantic conventions for LLM request tracing and evaluation, including secure configuration using Kubernetes secrets.
- **[Access Logs with AI/LLM metadata](./accesslogs.md)** - AI metadata produced by the AI gateway (model name, token usage, etc.) can be included in the Envoy Access Logs.
