# Console Metrics Configuration Guide

This guide covers the enhanced console metrics functionality in AI Gateway, including configuration options and usage examples for debugging and monitoring.

## Overview

The console metrics feature provides real-time metric output to the console for debugging and development purposes. This feature has been improved to properly handle metric temporality and only emit updates when there are actual data points to report.

## Key Features

- **Smart Temporality Handling**: Console metrics now properly respect metric temporality settings
- **Efficient Output**: Only emits metric updates when there are actual data points, reducing noise
- **Real-time Debugging**: Provides immediate feedback on AI Gateway operations
- **Structured Output**: Metrics are formatted as structured JSON for easy parsing

## Configuration

### Docker Compose Setup

The console metrics can be enabled through Docker configuration. Here's an example setup:

```yaml
version: '3.8'
services:
  ai-gateway:
    image: envoyproxy/ai-gateway:latest
    environment:
      - OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317
      - METRICS_CONSOLE_ENABLED=true
    ports:
      - "8080:8080"
    volumes:
      - ./config:/config
```

### Environment Variables

Configure console metrics using these environment variables:

- `METRICS_CONSOLE_ENABLED`: Enable console metrics output (default: false)
- `METRICS_CONSOLE_TEMPORALITY`: Set metric temporality behavior (default: delta)

## Usage Examples

### Basic Console Output

When console metrics are enabled, you'll see structured output like this:

```json
{
  "Resource": [
    {
      "Key": "service.name",
      "Value": {
        "Type": "STRING",
        "Value": "ai-gateway"
      }
    }
  ],
  "ScopeMetrics": [
    {
      "Scope": {
        "Name": "envoyproxy/ai-gateway",
        "Version": ""
      },
      "Metrics": [
        {
          "Name": "gen_ai.client.token.usage",
          "Description": "Number of tokens processed.",
          "Unit": "token",
          "Data": {
            "DataPoints": [
              {
                "Attributes": [
                  {
                    "Key": "gen_ai.operation.name",
                    "Value": {
                      "Type": "STRING",
                      "Value": "chat"
                    }
                  },
                  {
                    "Key": "gen_ai.request.model",
                    "Value": {
                      "Type": "STRING",
                      "Value": "qwen2.5:0.5b"
                    }
                  }
                ],
                "Count": 1,
                "Sum": 44
              }
            ]
          }
        }
      ]
    }
  ]
}
```

### Monitoring Token Usage

Console metrics provide detailed token usage information:

```bash
docker logs ai-gateway -f | grep "gen_ai.client.token.usage"
```

This will show real-time token consumption metrics including:
- Input tokens
- Output tokens
- Total tokens per request
- Model information

### Request Duration Metrics

Monitor request performance with duration metrics:

```json
{
  "Name": "gen_ai.server.request.duration",
  "Description": "Generative AI server request duration such as time-to-last byte or last output token.",
  "Unit": "s",
  "Data": {
    "DataPoints": [
      {
        "Attributes": [
          {
            "Key": "gen_ai.operation.name",
            "Value": {
              "Type": "STRING",
              "Value": "chat"
            }
          }
        ],
        "Count": 1,
        "Min": 0.106552042,
        "Max": 0.106552042,
        "Sum": 0.106552042
      }
    ]
  }
}
```

## Debugging Workflow

### Step 1: Enable Console Metrics

Add console metrics to your development environment:

```bash
export METRICS_CONSOLE_ENABLED=true
docker-compose up --build
```

### Step 2: Monitor Real-time Output

Watch the console output while making requests:

```bash
docker logs aigw -f
```

### Step 3: Analyze Metrics

Look for key metrics in the output:
- Token usage patterns
- Request duration trends
- Error rates and patterns
- Model performance characteristics

## Integration with Official Images

The console metrics feature works seamlessly with official AI Gateway Docker images. When using official images, ensure you:

1. Use the `--build` flag carefully to ensure fresh images
2. Configure environment variables appropriately
3. Mount configuration files as needed

Example with official image:

```bash
docker run -e METRICS_CONSOLE_ENABLED=true \
  -p 8080:8080 \
  envoyproxy/ai-gateway:latest
```

## Best Practices

### Development Environment

- Enable console metrics only in development/debugging scenarios
- Use structured log parsing tools to analyze output
- Combine with access logs for comprehensive debugging

### Production Considerations

- Disable console metrics in production environments
- Use proper observability tools (Prometheus, OTEL collectors) instead
- Consider log volume impact on system resources

### Filtering and Processing

Process console metrics output using standard tools:

```bash
# Filter for token metrics only
docker logs aigw -f | jq 'select(.ScopeMetrics[].Metrics[].Name == "gen_ai.client.token.usage")'

# Monitor request durations
docker logs aigw -f | jq 'select(.ScopeMetrics[].Metrics[].Name == "gen_ai.server.request.duration")'
```

## Troubleshooting

### No Metrics Output

If you're not seeing console metrics:

1. Verify `METRICS_CONSOLE_ENABLED=true` is set
2. Check that requests are actually being processed
3. Ensure the container has proper logging configuration

### Excessive Output

If console output is too verbose:

1. Consider using metric filtering
2. Adjust sampling rates if available
3. Use external log aggregation tools

### Performance Impact

Console metrics can impact performance:

1. Use only for debugging and development
2. Monitor container resource usage
3. Consider disabling in high-throughput scenarios

## Related Documentation

- [Observability Metrics](../capabilities/observability/metrics.md)
- [Tracing Configuration](../capabilities/observability/tracing.md)
- [Access Logs](../capabilities/observability/accesslogs.md)