# OpenTelemetry Metrics Export Guide

This guide explains how to configure OpenTelemetry (OTEL) metrics export in AI Gateway, including support for native OTEL systems and console output for debugging.

## Overview

AI Gateway now supports exporting metrics directly to OpenTelemetry collectors, eliminating the need for Prometheus in OTEL-native environments. This feature enables integration with systems like Elastic Stack, OTEL TUI, and other OTEL-compatible monitoring solutions.

## Configuration

### Environment Variables

Configure metrics export using these environment variables:

```bash
# Enable OTEL metrics export (replaces Prometheus listener)
OTEL_METRICS_EXPORTER=otlp

# Console output for debugging (no external dependencies)
OTEL_METRICS_EXPORTER=console

# Disable metrics export (useful for Phoenix users who only need traces)
OTEL_METRICS_EXPORTER=none

# Separate OTLP endpoints for traces and metrics
OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://localhost:4318/v1/traces
OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=http://localhost:4318/v1/metrics
```

### Export Types

#### OTLP Export
Exports metrics to OpenTelemetry Protocol (OTLP) collectors:

```bash
OTEL_METRICS_EXPORTER=otlp
OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=http://your-otel-collector:4318/v1/metrics
```

#### Console Export
Outputs metrics to console for immediate debugging:

```bash
OTEL_METRICS_EXPORTER=console
```

#### Disabled Export
Disables metrics export entirely:

```bash
OTEL_METRICS_EXPORTER=none
```

## Usage Examples

### Docker Compose with OTEL TUI

Example configuration for using OTEL TUI for metrics visualization:

```yaml
# docker-compose.yaml
services:
  ai-gateway:
    environment:
      - OTEL_METRICS_EXPORTER=otlp
      - OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=http://otel-collector:4318/v1/metrics
      - OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://otel-collector:4318/v1/traces
  
  otel-collector:
    image: otel/opentelemetry-collector:latest
    ports:
      - "4317:4317"
      - "4318:4318"
  
  otel-tui:
    image: ymtdzzz/otel-tui:latest
    ports:
      - "8080:8080"
```

### Environment File Profiles

Use profile-based environment file selection for different configurations:

```bash
# .env.otel.console
OTEL_METRICS_EXPORTER=console
OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://localhost:4318/v1/traces

# .env.otel.otlp
OTEL_METRICS_EXPORTER=otlp
OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=http://localhost:4318/v1/metrics
OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://localhost:4318/v1/traces

# .env.otel.phoenix
OTEL_METRICS_EXPORTER=none
OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://phoenix:4317/v1/traces
```

Load the appropriate profile:
```bash
COMPOSE_PROFILES=console docker-compose up
COMPOSE_PROFILES=otlp docker-compose up
COMPOSE_PROFILES=phoenix docker-compose up
```

## Integration Examples

### Elastic Stack Integration

Configure for Elastic APM:

```bash
OTEL_METRICS_EXPORTER=otlp
OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=https://your-elastic-apm:8200/intake/v2/otlp/v1/metrics
OTEL_EXPORTER_OTLP_HEADERS="Authorization=Bearer your-secret-token"
```

### Phoenix Integration

For Phoenix users who only need tracing:

```bash
OTEL_METRICS_EXPORTER=none
OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://phoenix:4317/v1/traces
```

### Development and Debugging

For local development with immediate feedback:

```bash
OTEL_METRICS_EXPORTER=console
```

This outputs metrics directly to the console without requiring external services.

## Migration from Prometheus

When migrating from Prometheus-based metrics:

1. **Before**: Metrics were exposed via Prometheus endpoint
2. **After**: Set `OTEL_METRICS_EXPORTER=otlp` to export via OTEL
3. **Compatibility**: Both can coexist during migration period

## Troubleshooting

### Common Issues

**Metrics not appearing in collector:**
- Verify `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT` is correctly set
- Check collector configuration accepts OTLP metrics
- Ensure network connectivity between AI Gateway and collector

**Console output too verbose:**
- Use `OTEL_METRICS_EXPORTER=none` to disable metrics
- Consider using OTLP export with filtering in the collector

**Phoenix compatibility:**
- Always set `OTEL_METRICS_EXPORTER=none` for Phoenix deployments
- Phoenix only supports trace data, not metrics

### Verification

Verify metrics export is working:

1. **Console mode**: Check application logs for metric output
2. **OTLP mode**: Verify metrics appear in your OTEL collector/backend
3. **None mode**: Confirm no metrics-related errors in logs

## Best Practices

1. **Environment-specific configuration**: Use profile-based env files for different deployment environments
2. **Resource optimization**: Use `none` export type when metrics aren't needed
3. **Debugging**: Start with `console` export for initial setup and troubleshooting
4. **Production**: Use `otlp` export with appropriate collector configuration for production deployments