---
id: prompt-caching
title: Prompt Caching
sidebar_position: 8
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

This guide explains how to use prompt caching with Claude models through Envoy AI Gateway using the `cache_control` API, as well as how to track cached tokens from providers with automatic caching.

## Overview

Prompt caching allows you to cache frequently used content (system prompts, large documents, tool definitions) to reduce token costs and improve response times.

### Explicit Cache Control (Claude Models)

The `cache_control` API allows you to explicitly mark content for caching. This is supported for **Claude models** across all providers:

- **Direct Anthropic** (via Anthropic Messages API)
- **GCP Vertex AI** (Anthropic Claude models)
- **AWS Bedrock** (Anthropic Claude models)

This unified approach ensures your caching implementation works consistently regardless of which provider hosts the Claude model.

### Automatic Caching (OpenAI)

OpenAI provides [automatic prompt caching](https://platform.openai.com/docs/guides/prompt-caching) that caches matching prompt prefixes without requiring explicit `cache_control` markers. The gateway automatically tracks cached tokens reported in OpenAI responses for cost tracking and rate limiting.

## Benefits

| Benefit | Description |
|---------|-------------|
| **Cost Optimization** | Reduce token costs by caching repeated content |
| **Performance** | Faster response times for cached content |
| **Provider Consistency** | Same `cache_control` API across all Claude providers |
| **Token Transparency** | Clear reporting of cache read/write tokens |
| **Easy Migration** | Switch Claude providers without code changes |

## Provider Compatibility

### Claude Models (Explicit `cache_control`)

| Feature | Anthropic Direct | GCP Vertex AI | AWS Bedrock |
|---------|------------------|---------------|-------------|
| Max Cache Points | 4 per request | 4 per request | 4 per request |
| Min Token Threshold | 1,024+ tokens | 1,024+ tokens | 1,024+ tokens |
| Billing Integration | Native | Native | Native |
| Cache Types | ephemeral | ephemeral | ephemeral |

### OpenAI (Automatic Caching)

| Feature | OpenAI |
|---------|--------|
| Cache Control | Automatic (no explicit markers needed) |
| Min Token Threshold | 1,024+ tokens |
| Cache Duration | 5-10 minutes of inactivity |
| Discount | 50% on cached input tokens |

:::note
OpenAI's automatic caching caches the **longest matching prefix** of your prompts. Reordering content or adding new content at the beginning will cause cache misses. See [OpenAI's prompt caching guide](https://platform.openai.com/docs/guides/prompt-caching) for details.
:::

## Using `cache_control` with Claude Models

Add `cache_control` to content blocks that you want to cache. The syntax is identical across all Claude providers - just change the model name:

<Tabs>
<TabItem value="anthropic" label="Direct Anthropic">

```bash
curl -X POST $GATEWAY_URL/anthropic/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4-5",
    "max_tokens": 1024,
    "messages": [
      {
        "role": "user",
        "content": [
          {
            "type": "text",
            "text": "You are an expert assistant with access to a comprehensive knowledge base covering technology, science, and business. Always provide detailed, accurate, and well-sourced responses. When analyzing documents, break down complex concepts into digestible parts. [Include additional context to reach 1024+ tokens for caching...]",
            "cache_control": {"type": "ephemeral"}
          }
        ]
      },
      {
        "role": "user",
        "content": "What are the key principles of microservices architecture?"
      }
    ]
  }'
```

</TabItem>
<TabItem value="gcp" label="GCP Vertex AI">

```bash
curl -X POST $GATEWAY_URL/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4@20250514",
    "messages": [
      {
        "role": "system",
        "content": [
          {
            "type": "text",
            "text": "You are an expert assistant with access to a comprehensive knowledge base covering technology, science, and business. Always provide detailed, accurate, and well-sourced responses. When analyzing documents, break down complex concepts into digestible parts. [Include additional context to reach 1024+ tokens for caching...]",
            "cache_control": {"type": "ephemeral"}
          }
        ]
      },
      {
        "role": "user",
        "content": "What are the key principles of microservices architecture?"
      }
    ]
  }'
```

</TabItem>
<TabItem value="aws" label="AWS Bedrock">

```bash
curl -X POST $GATEWAY_URL/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "anthropic.claude-sonnet-4-20250514-v1:0",
    "messages": [
      {
        "role": "system",
        "content": [
          {
            "type": "text",
            "text": "You are an expert assistant with access to a comprehensive knowledge base covering technology, science, and business. Always provide detailed, accurate, and well-sourced responses. When analyzing documents, break down complex concepts into digestible parts. [Include additional context to reach 1024+ tokens for caching...]",
            "cache_control": {"type": "ephemeral"}
          }
        ]
      },
      {
        "role": "user",
        "content": "What are the key principles of microservices architecture?"
      }
    ]
  }'
```

</TabItem>
</Tabs>

## Common Use Cases (Claude)

### Document Analysis with Caching

Cache a large document for analysis across multiple queries:

```json
{
  "model": "claude-sonnet-4-5",
  "max_tokens": 1024,
  "messages": [
    {
      "role": "user",
      "content": [
        {
          "type": "text",
          "text": "Please analyze this technical specification document:\n\n# Technical Specification: Distributed System Architecture\n\n## 1. Introduction\nThis document outlines the architecture for a distributed system designed to handle high-throughput data processing... [large document content continues]",
          "cache_control": {"type": "ephemeral"}
        },
        {
          "type": "text",
          "text": "What are the main components described in this document?"
        }
      ]
    }
  ]
}
```

### Tool Definition Caching

Cache complex tool schemas to save tokens on repeated function-calling requests:

```json
{
  "model": "claude-sonnet-4-5",
  "max_tokens": 1024,
  "messages": [
    {
      "role": "user",
      "content": "Help me search for information about cloud computing trends."
    }
  ],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "search_knowledge_base",
        "description": "Search through a comprehensive knowledge base containing technical articles, research papers, industry reports, and documentation covering cloud computing, distributed systems, microservices, containers, orchestration, serverless computing, machine learning platforms, data engineering tools, and related technologies...",
        "parameters": {
          "type": "object",
          "properties": {
            "query": {
              "type": "string",
              "description": "Natural language search query"
            },
            "category": {
              "type": "string",
              "enum": ["research_papers", "tutorials", "documentation"]
            }
          },
          "required": ["query"]
        },
        "cache_control": {"type": "ephemeral"}
      }
    }
  ]
}
```

### Multiple Cache Points

You can use up to 4 cache points in a single request to cache different content sections:

```json
{
  "model": "claude-sonnet-4-5",
  "max_tokens": 1024,
  "messages": [
    {
      "role": "system",
      "content": [
        {
          "type": "text",
          "text": "You are a technical documentation assistant specialized in API design and software architecture.",
          "cache_control": {"type": "ephemeral"}
        }
      ]
    },
    {
      "role": "user",
      "content": [
        {
          "type": "text",
          "text": "Here is our API specification: [Large API spec content...]",
          "cache_control": {"type": "ephemeral"}
        },
        {
          "type": "text",
          "text": "Here is our database schema: [Large schema content...]",
          "cache_control": {"type": "ephemeral"}
        },
        {
          "type": "text",
          "text": "Please analyze these components and suggest improvements."
        }
      ]
    }
  ]
}
```

## Tracking OpenAI Automatic Caching

For OpenAI models, no explicit `cache_control` markers are needed. OpenAI automatically caches prompt prefixes and reports the cached tokens in responses:

```bash
curl -X POST $GATEWAY_URL/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [
      {
        "role": "system",
        "content": "You are an expert assistant with comprehensive knowledge... [long system prompt that will be automatically cached on repeat requests]"
      },
      {
        "role": "user",
        "content": "What is the capital of France?"
      }
    ]
  }'
```

On subsequent requests with the same prompt prefix, OpenAI will automatically use cached tokens and report them in the response.

## Response Format

Cached token usage is reported consistently across all providers in the response:

```json
{
  "choices": [...],
  "usage": {
    "prompt_tokens": 2000,
    "completion_tokens": 150,
    "total_tokens": 2150,
    "prompt_tokens_details": {
      "cached_tokens": 1800
    }
  }
}
```

- `cached_tokens`: Number of tokens read from cache (reduced cost)
- Cache write tokens are tracked internally for billing

## Cost Tracking Configuration

To track cached tokens for rate limiting or cost analysis, configure your `AIGatewayRoute` with the appropriate cost types:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGatewayRoute
metadata:
  name: my-route
spec:
  llmRequestCosts:
    - metadataKey: llm_input_token
      type: InputToken
    - metadataKey: llm_cached_input_token
      type: CachedInputToken        # Tokens read from cache
    - metadataKey: llm_cache_creation_token
      type: CacheCreationInputToken # Tokens written to cache
    - metadataKey: llm_output_token
      type: OutputToken
```

You can also use CEL expressions for custom cost calculations that account for cached tokens:

```yaml
llmRequestCosts:
  - metadataKey: adjusted_cost
    type: CEL
    cel: "(input_tokens - cached_input_tokens) + cached_input_tokens * 0.1 + cache_creation_input_tokens * 1.25 + output_tokens"
```

This example:
- Charges full price for non-cached input tokens
- Charges 10% for cached input tokens (cache read)
- Charges 125% for cache creation tokens (cache write)
- Charges full price for output tokens

## Best Practices

### For All Providers

1. **Minimum Token Threshold**: All providers require 1,024+ tokens for effective caching. Content below this threshold won't be cached.

2. **Monitor Usage**: Track `cached_tokens` in responses to measure cost savings.

3. **Consistent Prompts**: Keep cacheable content identical between requests. Any changes (even whitespace) cause cache misses.

### For Claude Models (Explicit `cache_control`)

4. **Strategic Placement**: Place `cache_control` markers on content that will be reused:
   - System prompts
   - Reference documents
   - Tool definitions
   - Examples and few-shot prompts

5. **Order Matters**: Content after the last `cache_control` marker is not cached.

6. **Max 4 Cache Points**: You can use up to 4 `cache_control` markers per request.

### For OpenAI (Automatic Caching)

7. **Prefix Matching**: OpenAI caches the longest matching prefix. Put stable content at the beginning of your prompts.

8. **Avoid Reordering**: Moving content around breaks prefix matching and causes cache misses.

## Troubleshooting

### Content Not Being Cached (Claude)

- Verify the content block is at least 1,024 tokens
- Ensure `cache_control` is placed on the correct content block
- Check that you haven't exceeded 4 cache points per request

### Cache Misses (All Providers)

- Cached content must be identical between requests
- Any changes to cached content (even whitespace) will cause a cache miss
- Cache is ephemeral and may expire after periods of inactivity

### OpenAI Cache Misses

- Content must be at the beginning of the prompt (prefix matching)
- Reordering messages breaks cache matching
- Cache expires after 5-10 minutes of inactivity

### Provider-Specific Documentation

Refer to provider documentation for specific error messages and limitations:
- [Anthropic Prompt Caching](https://docs.anthropic.com/en/docs/build-with-claude/prompt-caching)
- [AWS Bedrock Prompt Caching](https://docs.aws.amazon.com/bedrock/latest/userguide/prompt-caching.html)
- [OpenAI Prompt Caching](https://platform.openai.com/docs/guides/prompt-caching)

## See Also

- [Usage-based Rate Limiting](./usage-based-ratelimiting.md) - Configure rate limits based on token usage
- [Connect Anthropic](../../getting-started/connect-providers/anthropic.md) - Set up direct Anthropic access
- [Connect AWS Bedrock](../../getting-started/connect-providers/aws-bedrock.md) - Set up AWS Bedrock access
