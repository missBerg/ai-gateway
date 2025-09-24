# Metrics and Observability Guide

This guide covers the metrics and observability features in AI Gateway, including console metrics output, tracing, and debugging capabilities.

## Overview

AI Gateway provides comprehensive observability through:
- OpenTelemetry-based metrics collection
- Console metrics output for debugging
- Distributed tracing with span data
- Access logs with GenAI-specific fields

## Console Metrics

### Configuration

Console metrics allow you to debug and monitor AI Gateway operations in real-time through standard output. This is particularly useful during development and troubleshooting.

### Example Output

When console metrics are enabled, you'll see detailed output including:

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

### Key Metrics

#### Token Usage Metrics
- **Name**: `gen_ai.client.token.usage`
- **Type**: Histogram
- **Description**: Tracks input and output token consumption
- **Attributes**:
  - `gen_ai.operation.name`: Operation type (e.g., "chat")
  - `gen_ai.provider.name`: LLM provider (e.g., "openai")
  - `gen_ai.request.model`: Model name
  - `gen_ai.token.type`: "input" or "output"

#### Request Duration Metrics
- **Name**: `gen_ai.server.request.duration`
- **Type**: Histogram
- **Description**: Measures request processing time
- **Unit**: seconds
- **Attributes**:
  - `gen_ai.operation.name`: Operation type
  - `gen_ai.provider.name`: LLM provider
  - `gen_ai.request.model`: Model name

## Tracing

### Span Data

AI Gateway generates detailed span information for each request:

```json
{
  "Name": "ChatCompletion",
  "SpanContext": {
    "TraceID": "5450d3297f9ee876e03ce0ebcc78c73b",
    "SpanID": "4be4d3f88ada3a95"
  },
  "Attributes": [
    {
      "Key": "openinference.span.kind",
      "Value": {
        "Type": "STRING",
        "Value": "LLM"
      }
    },
    {
      "Key": "llm.system",
      "Value": {
        "Type": "STRING",
        "Value": "openai"
      }
    },
    {
      "Key": "llm.model_name",
      "Value": {
        "Type": "STRING",
        "Value": "qwen2.5:0.5b"
      }
    },
    {
      "Key": "llm.token_count.prompt",
      "Value": {
        "Type": "INT64",
        "Value": 44
      }
    },
    {
      "Key": "llm.token_count.completion",
      "Value": {
        "Type": "INT64",
        "Value": 12
      }
    }
  ]
}
```

### Key Span Attributes

- `openinference.span.kind`: Identifies LLM operations
- `llm.system`: LLM provider system
- `llm.model_name`: Specific model used
- `input.value`: Request payload
- `output.value`: Response payload
- `llm.token_count.*`: Token usage statistics

## Access Logs

Access logs include GenAI-specific fields alongside standard HTTP metrics:

```json
{
  "bytes_received": 159,
  "bytes_sent": 334,
  "duration": 108,
  "genai_backend_name": "default/ollama/route/aigw-run/rule/0/ref/0",
  "genai_model_name": "qwen2.5:0.5b",
  "genai_tokens_input": 44,
  "genai_tokens_output": 12,
  "method": "POST",
  "response_code": 200,
  "upstream_cluster": "httproute/default/aigw-run/rule/0",
  "upstream_host": "192.168.5.2:11434"
}
```

### GenAI-Specific Fields

- `genai_backend_name`: Backend service identifier
- `genai_model_name`: Model used for the request
- `genai_model_name_override`: Override model if different
- `genai_tokens_input`: Input token count
- `genai_tokens_output`: Output token count

## Docker Integration

### Using Official Images

The examples now use official AI Gateway Docker images. When running with Docker:

```bash
# View real-time logs with metrics
docker logs aigw -f

# Build fresh image if needed
docker-compose build --no-cache
```

### Development Setup

For development and debugging:

1. Enable console metrics in your configuration
2. Run AI Gateway with Docker Compose
3. Monitor logs for real-time metrics output
4. Use the console output to debug request flows and performance

## Debugging Best Practices

1. **Console Metrics**: Enable for real-time debugging during development
2. **Log Analysis**: Use the structured JSON output for automated analysis
3. **Performance Monitoring**: Track token usage and request duration metrics
4. **Error Tracking**: Monitor response codes and failure reasons
5. **Resource Usage**: Track bytes sent/received for bandwidth analysis

## Configuration Examples

### Enabling Console Metrics

Configure your AI Gateway to output metrics to console for debugging:

```yaml
# Add to your AI Gateway configuration
observability:
  metrics:
    console:
      enabled: true
      temporality: delta
```

This configuration ensures metrics are output with proper temporality handling and only when data points are available, making the console output more useful for debugging purposes.