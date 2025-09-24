# OpenTelemetry Metrics Export Guide

This guide explains how to configure and use OpenTelemetry (OTEL) metrics export in AI Gateway, which allows you to send metrics directly to OTEL-compatible systems without requiring a Prometheus endpoint.

## Overview

AI Gateway now supports exporting metrics via OpenTelemetry protocol (OTLP) in addition to the traditional Prometheus format. This enables integration with OTEL-native observability stacks like Elastic Stack, OTEL TUI, and other OTLP-compatible systems.

## Configuration

### Environment Variables

Configure OTEL metrics export using these environment variables:

```bash
# Enable OTEL metrics exporter (replaces Prometheus listener)
OTEL_METRICS_EXPORTER=otlp

# Set OTLP endpoint for metrics (can be different from traces)
OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=http://localhost:4318/v1/metrics

# For console output during development/debugging
OTEL_METRICS_EXPORTER=console
```

### Separate Endpoints for Traces and Metrics

You can configure different OTLP endpoints for traces and metrics:

```bash
# Traces endpoint
OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://traces-collector:4318/v1/traces

# Metrics endpoint  
OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=http://metrics-collector:4318/v1/metrics
```

### Exporter Types

The following exporter types are supported:

- `otlp` - Export to OTLP-compatible collector
- `console` - Output metrics to console (useful for debugging)
- `none` - Disable metrics export
- `prometheus` - Traditional Prometheus format (default)

## Usage Examples

### Console Output for Development

For immediate debugging without external dependencies:

```bash
export OTEL_METRICS_EXPORTER=console
./aigw run
```

This will output metrics directly to the console, making it easy to verify metrics are being generated during development.

### OTLP Export to Collector

For production use with an OTEL collector:

```bash
export OTEL_METRICS_EXPORTER=otlp
export OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=http://otel-collector:4318/v1/metrics
./aigw run
```

### Using with OTEL TUI

[OTEL TUI](https://github.com/ymtdzzz/otel-tui) provides a terminal-based interface for viewing OTEL data:

```bash
# Start OTEL TUI
otel-tui

# Configure AI Gateway to send to OTEL TUI
export OTEL_METRICS_EXPORTER=otlp
export OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=http://localhost:4318/v1/metrics
export OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://localhost:4318/v1/traces
./aigw run
```

### Phoenix Integration

When using Phoenix (which only supports traces), disable metrics export:

```bash
export OTEL_METRICS_EXPORTER=none
export OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://phoenix:4318/v1/traces
./aigw run
```

## Docker Compose Configuration

The AI Gateway repository includes updated Docker Compose examples with profile-based environment file selection:

```bash
# Use console profile (default)
docker-compose up

# Use OTEL TUI profile
COMPOSE_PROFILES=otel-tui docker-compose up

# Use Phoenix profile  
COMPOSE_PROFILES=phoenix docker-compose up
```

This automatically loads the appropriate `.env.otel.${COMPOSE_PROFILES}` file with the correct configuration.

## Migration from Prometheus

If you're currently using Prometheus metrics and want to switch to OTLP:

1. **Before**: Metrics available at `http://localhost:9090/metrics`
2. **After**: Set `OTEL_METRICS_EXPORTER=otlp` and configure your OTLP endpoint

Note that enabling OTLP metrics export replaces the Prometheus listener entirely. You cannot have both active simultaneously.

## Troubleshooting

### Metrics Not Appearing

1. Verify the exporter is set correctly:
   ```bash
   echo $OTEL_METRICS_EXPORTER
   ```

2. Check the endpoint is reachable:
   ```bash
   curl -v $OTEL_EXPORTER_OTLP_METRICS_ENDPOINT
   ```

3. Use console exporter to verify metrics generation:
   ```bash
   OTEL_METRICS_EXPORTER=console ./aigw run
   ```

### Different Endpoints for Traces and Metrics

If you need separate endpoints, ensure both are configured:

```bash
export OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://trace-collector:4318/v1/traces
export OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=http://metrics-collector:4318/v1/metrics
```

The generic `OTEL_EXPORTER_OTLP_ENDPOINT` will be used as a fallback if specific endpoints aren't set.