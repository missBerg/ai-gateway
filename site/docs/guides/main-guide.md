# OpenTelemetry Metrics Configuration Guide

This guide explains how to configure OpenTelemetry (OTEL) metrics export in AI Gateway, including native OTEL export capabilities and console debugging options.

## Overview

AI Gateway now supports native OpenTelemetry metrics export alongside the existing Prometheus metrics. This allows you to:

- Export metrics directly to OTEL-compatible systems (Elastic Stack, Jaeger, etc.)
- Use console output for immediate debugging
- Integrate with OTEL-native monitoring tools like `otel-tui`

## Configuration Options

### Environment Variables

Configure metrics export using these environment variables:

```bash
# Metrics exporter type (otlp, console, prometheus, none)
OTEL_METRICS_EXPORTER=otlp

# OTLP endpoint for metrics (can be different from traces)
OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=http://localhost:4318/v1/metrics

# OTLP endpoint for traces
OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://localhost:4318/v1/traces

# Generic OTLP endpoint (used if specific endpoints not set)
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
```

### Exporter Types

#### OTLP Export
Export metrics to any OTLP-compatible collector:

```bash
OTEL_METRICS_EXPORTER=otlp
OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=http://your-collector:4318/v1/metrics
```

#### Console Export
Output metrics to console for debugging:

```bash
OTEL_METRICS_EXPORTER=console
```

#### Prometheus (Default)
Continue using Prometheus metrics endpoint:

```bash
OTEL_METRICS_EXPORTER=prometheus
# Or leave unset for default behavior
```

#### Disable Metrics
Disable metrics export entirely:

```bash
OTEL_METRICS_EXPORTER=none
```

## Examples

### Using with otel-tui

Monitor metrics and traces in a terminal UI using [otel-tui](https://github.com/ymtdzzz/otel-tui):

1. Start otel-tui collector:
```bash
docker run -p 4317:4317 -p 4318:4318 ghcr.io/ymtdzzz/otel-tui:latest
```

2. Configure AI Gateway:
```bash
OTEL_METRICS_EXPORTER=otlp
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
```

### Phoenix Integration

When using Phoenix (traces only), disable metrics export:

```bash
OTEL_METRICS_EXPORTER=none
OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://localhost:4317
```

### Docker Compose Example

```yaml
services:
  ai-gateway:
    image: envoyproxy/ai-gateway
    environment:
      - OTEL_METRICS_EXPORTER=otlp
      - OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=http://otel-collector:4318/v1/metrics
      - OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://otel-collector:4318/v1/traces
    
  otel-collector:
    image: otel/opentelemetry-collector-contrib
    ports:
      - "4317:4317"
      - "4318:4318"
```

## Profile-based Configuration

The AI Gateway examples now support profile-based environment file selection:

```bash
# Use console profile (default)
docker-compose up

# Use otel-tui profile
COMPOSE_PROFILES=otel-tui docker-compose up

# Use phoenix profile
COMPOSE_PROFILES=phoenix docker-compose up
```

This automatically loads the appropriate `.env.otel.${COMPOSE_PROFILES}` file.

## Troubleshooting

### Metrics Not Appearing
- Verify the OTLP endpoint is accessible
- Check that `OTEL_METRICS_EXPORTER` is set correctly
- Ensure your collector supports the OTLP metrics format

### Console Output Too Verbose
- Use `OTEL_METRICS_EXPORTER=none` to disable metrics
- Filter console output or redirect to a file for analysis

### Mixed Metrics Systems
- You can run both Prometheus and OTLP export simultaneously by leaving `OTEL_METRICS_EXPORTER` unset (Prometheus) and configuring OTLP endpoints for traces only

## Related Documentation

- [Observability Overview](../capabilities/observability/index.md)
- [Metrics Configuration](../capabilities/observability/metrics.md)
- [Tracing Configuration](../capabilities/observability/tracing.md)