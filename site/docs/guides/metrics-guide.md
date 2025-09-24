# Console Metrics and Docker Configuration Guide

This guide covers the enhanced console metrics functionality and updated Docker configuration in Envoy AI Gateway.

## Overview

The AI Gateway now provides improved console metrics output and standardized Docker image usage across examples. This enhancement makes debugging and monitoring much more effective for development and testing scenarios.

## Console Metrics Improvements

### What Changed

The console metrics exporter has been enhanced to:
- Properly handle metric temporality
- Only emit updates when there are actual data points
- Provide cleaner, more readable output for debugging

### Key Features

- **Filtered Output**: Console metrics now only display when there are actual data points, reducing noise
- **Proper Temporality**: Metrics correctly respect temporal settings for accurate reporting  
- **Enhanced Debugging**: Cleaner output format makes it easier to monitor AI Gateway behavior

### Example Output

When running with console metrics enabled, you'll see structured output like this:

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
        "Name": "envoyproxy/ai-gateway"
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

## Docker Configuration Updates

### Official Images

All Docker examples now use official AI Gateway images instead of local builds, ensuring consistency and reliability.

### Build Considerations

When using `docker-compose up --build`:
- The `--build` flag will rebuild containers
- To ensure fresh AI Gateway images, pull the latest version first
- This prevents stale local images from being used

### Example Usage

```bash
# Pull latest images and run
docker-compose pull
docker-compose up

# Force rebuild (ensure fresh images first)
docker-compose pull
docker-compose up --build
```

## Monitoring with Console Metrics

### Enable Console Output

To enable console metrics for debugging:

1. Configure your AI Gateway deployment with console metrics enabled
2. Monitor the logs for real-time metrics output
3. Use the structured JSON output to track:
   - Token usage (input/output)
   - Request duration
   - Model performance
   - Provider interactions

### Debugging Workflow

```bash
# Start AI Gateway with console metrics
docker-compose up

# In another terminal, monitor logs
docker logs aigw -f

# Make requests to see metrics in real-time
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qwen2.5:0.5b",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

## Benefits

### For Developers
- Real-time visibility into AI Gateway performance
- Cleaner debugging experience with reduced noise
- Consistent Docker image usage across environments

### For Operations
- Reliable metrics collection for monitoring
- Standardized deployment configurations
- Better observability into AI model interactions

## Best Practices

1. **Development**: Use console metrics for local debugging and testing
2. **Production**: Configure proper metrics exporters (Prometheus, OTLP) instead of console output
3. **Docker**: Always use official images in production deployments
4. **Monitoring**: Combine console metrics with access logs for comprehensive debugging

## Related Features

- [OpenTelemetry Integration](./tracing.md)
- [Access Logs](./accesslogs.md)
- [Metrics Overview](./metrics.md)