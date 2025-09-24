# Metrics Console Debugging Guide

This guide covers enhanced console metrics features in the Envoy AI Gateway, focusing on improved debugging capabilities and proper usage of metrics exporters.

## Overview

The Envoy AI Gateway now provides improved console metrics export functionality that properly handles temporality and only outputs metrics when data is available. This enhancement makes console metrics a valuable debugging tool for development and troubleshooting.

## Features

### Non-Empty Metrics Export

The console exporter now intelligently filters out empty metrics exports, reducing noise in the output and only displaying relevant data when metrics are actually generated.

### Temporality Support

The console exporter respects the `OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE` environment variable:

- **Cumulative** (default): Shows cumulative metric values over time
- **Delta**: Shows only the change in metric values since the last export

## Configuration

### Environment Variables

The following environment variables control metrics behavior:

```bash
# Disable OpenTelemetry SDK entirely
OTEL_SDK_DISABLED=true

# Set metrics exporter type
OTEL_METRICS_EXPORTER=console  # or "none", "prometheus", "otlp"

# Configure temporality preference
OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=delta  # or "cumulative"

# OTLP endpoint configuration
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=http://localhost:4318/v1/metrics
```

### Docker Compose Example

To use console metrics with the official AI Gateway images:

```yaml
services:
  aigw:
    image: envoyproxy/ai-gateway-cli:latest
    container_name: aigw
    environment:
      - OTEL_METRICS_EXPORTER=console
      - OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=delta
    ports:
      - "1975:1975"
    volumes:
      - ./ai-gateway-local.yaml:/config.yaml:ro
    command: ["run", "/config.yaml"]
```

## Usage

### Basic Console Metrics

To enable console metrics output:

```bash
# Set environment variable
export OTEL_METRICS_EXPORTER=console

# Run AI Gateway
./aigw run config.yaml
```

### Debugging with Console Output

When console metrics are enabled, you'll see detailed output including:

1. **Request traces** with OpenInference span information:
   ```json
   {
     "Name": "ChatCompletion",
     "SpanContext": {
       "TraceID": "5450d3297f9ee876e03ce0ebcc78c73b",
       "SpanID": "4be4d3f88ada3a95"
     },
     "Attributes": [
       {
         "Key": "llm.model_name",
         "Value": {"Type": "STRING", "Value": "qwen2.5:0.5b"}
       },
       {
         "Key": "llm.token_count.prompt",
         "Value": {"Type": "INT64", "Value": 44}
       }
     ]
   }
   ```

2. **Access logs** with GenAI-specific fields:
   ```json
   {
     "genai_backend_name": "default/ollama/route/aigw-run/rule/0/ref/0",
     "genai_model_name": "qwen2.5:0.5b",
     "genai_tokens_input": 44,
     "genai_tokens_output": 12,
     "response_code": 200
   }
   ```

3. **Metrics data** with proper temporality handling:
   ```json
   {
     "ScopeMetrics": [{
       "Metrics": [{
         "Name": "gen_ai.client.token.usage",
         "Data": {
           "DataPoints": [{
             "Attributes": [
               {"Key": "gen_ai.token.type", "Value": {"Type": "STRING", "Value": "input"}},
               {"Key": "gen_ai.request.model", "Value": {"Type": "STRING", "Value": "qwen2.5:0.5b"}}
             ],
             "Count": 1,
             "Sum": 44
           }]
         }
       }]
     }]
   }
   ```

### Filtering Console Output

To view specific types of telemetry data:

```bash
# View only traces
docker logs aigw -f | grep "SpanContext"

# View only metrics
docker logs aigw -f | grep "gen_ai"

# View only access logs with GenAI fields
docker logs aigw -f | grep "genai_model_name"
```

## Docker Integration

### Using Official Images

The enhanced metrics functionality works seamlessly with the official AI Gateway Docker images:

```bash
# Pull latest image
docker pull envoyproxy/ai-gateway-cli:latest

# Run with console metrics
docker run -e OTEL_METRICS_EXPORTER=console \
           -p 1975:1975 \
           -v $(pwd)/config.yaml:/config.yaml:ro \
           envoyproxy/ai-gateway-cli:latest run /config.yaml
```

### Building from Source

When building custom images, ensure you use the `--build` flag to get fresh builds:

```bash
# Build and run with fresh image
docker compose up --build --wait -d

# This ensures the image includes latest metrics improvements
```

## Troubleshooting

### No Metrics Output

If you're not seeing metrics in the console:

1. Verify the exporter is configured:
   ```bash
   echo $OTEL_METRICS_EXPORTER  # Should output "console"
   ```

2. Check that the SDK is not disabled:
   ```bash
   echo $OTEL_SDK_DISABLED      # Should be empty or "false"
   ```

3. Ensure requests are being processed by sending test traffic to the gateway.

### Too Much Output

If console output is too verbose:

1. Use delta temporality to see only changes:
   ```bash
   export OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=delta
   ```

2. Filter output using grep or similar tools to focus on specific metrics.

### Empty Metrics

The enhanced exporter automatically suppresses empty metrics exports. If you're not seeing any metrics output, it means no metrics data is being generated, which could indicate:

- No requests are being processed
- Metrics collection is not properly configured
- The application is not generating the expected metrics

## Best Practices

1. **Development**: Use console metrics with delta temporality for debugging
2. **Production**: Use Prometheus for metrics collection, OTLP for distributed tracing
3. **Container Logs**: Use `docker logs -f` to stream real-time metrics output
4. **Filtering**: Combine console output with grep/jq for focused debugging

## Related Documentation

- [Observability Overview](../capabilities/observability/index.md)
- [Metrics Configuration](../capabilities/observability/metrics.md)
- [Docker Compose Examples](../../examples/README.md)