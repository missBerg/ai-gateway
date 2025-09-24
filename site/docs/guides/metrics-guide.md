# Console Metrics and Docker Setup Guide

This guide covers the recent improvements to console metrics output and Docker configuration in the AI Gateway project.

## Overview

The AI Gateway now provides enhanced console metrics functionality with proper temporality handling and improved Docker examples using official images. These changes make debugging and local development more efficient.

## Console Metrics Improvements

### What Changed

The console metrics reporter has been enhanced to:
- Properly handle metric temporality 
- Only emit console updates when there are actual data points
- Provide cleaner, more useful debugging output

### Key Benefits

- **Reduced noise**: Console output only shows metrics when there's actual data
- **Better debugging**: More sophisticated console reporting for troubleshooting
- **Improved accuracy**: Proper temporality handling ensures metrics are reported correctly

### Example Output

When running the AI Gateway with console metrics enabled, you'll see structured output like:

```bash
$ docker logs aigw -f
looking up the latest patch for Envoy version 1.35
1.35.3 is already downloaded
starting: /tmp/envoy-gateway/versions/1.35.3/bin/envoy in run directory /tmp/envoy-gateway/runs/1758094103988916388

# Trace data
{"Name":"ChatCompletion","SpanContext":{"TraceID":"5450d3297f9ee876e03ce0ebcc78c73b",...}

# Access logs
{"bytes_received":159,"bytes_sent":334,"genai_model_name":"qwen2.5:0.5b","genai_tokens_input":44,"genai_tokens_output":12,...}

# Metrics data
{"Resource":[{"Key":"service.name","Value":{"Type":"STRING","Value":"ai-gateway"}}],"ScopeMetrics":[...]}
```

## Docker Configuration Updates

### Official Images

The Docker examples now use official AI Gateway images instead of building from source. This provides:
- Faster startup times
- Consistent, tested images
- Easier deployment for testing

### Build Considerations

If you need to build fresh images, use the `--build` flag carefully:

```bash
# To ensure you get the latest image
docker-compose up --build
```

## Getting Started

### Enable Console Metrics

To use the improved console metrics in your setup:

1. Configure your AI Gateway with console metrics enabled
2. Run your Docker setup:
   ```bash
   docker-compose up
   ```
3. Monitor the logs for structured metrics output:
   ```bash
   docker logs aigw -f
   ```

### Understanding the Output

The console output includes three types of information:

1. **Trace Data**: OpenTelemetry spans showing request flow and timing
2. **Access Logs**: HTTP request/response details with AI-specific fields
3. **Metrics Data**: Aggregated metrics including token usage and request duration

## Debugging with Console Metrics

The enhanced console output is particularly useful for:

- **Token Usage Analysis**: Track input/output tokens across requests
- **Performance Monitoring**: View request durations and response times  
- **Model Usage**: Monitor which models are being called
- **Error Diagnosis**: Identify issues in request processing

## Configuration

### Metric Types Available

The console metrics include:

- `gen_ai.client.token.usage`: Number of tokens processed (input/output)
- `gen_ai.server.request.duration`: Request duration in seconds

### Attributes Tracked

Each metric includes relevant attributes such as:
- `gen_ai.operation.name`: Operation type (e.g., "chat")
- `gen_ai.provider.name`: AI provider (e.g., "openai")  
- `gen_ai.request.model`: Model name used
- `gen_ai.token.type`: Token type ("input" or "output")

## Best Practices

1. **Use for Development**: Console metrics are ideal for local development and debugging
2. **Monitor Resource Usage**: Keep an eye on token consumption patterns
3. **Performance Testing**: Use duration metrics to identify bottlenecks
4. **Model Comparison**: Compare performance across different AI models

This improved console metrics functionality makes the AI Gateway much more observable and debuggable for development workflows.