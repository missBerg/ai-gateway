# AWS Bedrock Reasoning Stream Support

This guide covers the new reasoning stream support for AWS Bedrock, which enables Chain of Thought (CoT) reasoning capabilities in streaming responses.

## Overview

The AI Gateway now supports AWS Bedrock's reasoning stream functionality, allowing you to access the model's internal reasoning process while receiving streamed responses. This feature is particularly useful for complex problem-solving tasks where understanding the model's thought process is valuable.

## Supported Models

Reasoning stream support is available for:
- Anthropic Claude models with reasoning capabilities
- AWS Bedrock models that support the reasoning content block

## How It Works

### Request Format

To enable reasoning in your requests, include the `thinking` field in your OpenAI-compatible request:

```json
{
  "model": "anthropic.claude-3-5-sonnet-20241022-v2:0",
  "messages": [
    {
      "role": "user", 
      "content": "Solve this step by step: What is 15% of 240?"
    }
  ],
  "stream": true,
  "thinking": {
    "type": "enabled",
    "budget_tokens": 1000
  }
}
```

### Response Format

When streaming is enabled, you'll receive reasoning content in the response chunks:

```json
{
  "object": "chat.completion.chunk",
  "choices": [
    {
      "index": 0,
      "delta": {
        "role": "assistant",
        "reasoning_content": {
          "text": "I need to calculate 15% of 240. Let me break this down:",
          "signature": "abc123...",
          "redacted_content": null
        }
      }
    }
  ]
}
```

### Content Types

The reasoning stream support includes several content types:

#### 1. Reasoning Text
Contains the model's step-by-step reasoning:
- `text`: The reasoning content
- `signature`: Verification token for the reasoning text

#### 2. Redacted Content
For sensitive reasoning that has been filtered:
- `redacted_content`: Binary data of redacted reasoning

#### 3. Regular Response
Standard assistant response after reasoning is complete:
```json
{
  "delta": {
    "role": "assistant",
    "content": "15% of 240 is 36."
  }
}
```

## Message Format Support

The reasoning stream support also handles reasoning content in message history:

### Thinking Content
```json
{
  "role": "assistant",
  "content": [
    {
      "type": "thinking",
      "text": "Let me think about this problem...",
      "signature": "verification_token"
    },
    {
      "type": "text", 
      "text": "Here's my final answer..."
    }
  ]
}
```

### Redacted Thinking
```json
{
  "role": "assistant",
  "content": [
    {
      "type": "redacted_thinking",
      "redacted_content": "base64_encoded_redacted_data"
    }
  ]
}
```

## Configuration

### Enable Reasoning
Add the `thinking` field to your request with appropriate configuration:

```json
{
  "thinking": {
    "type": "enabled",
    "budget_tokens": 2000  // Optional: limit reasoning tokens
  }
}
```

### Stream Configuration
Ensure streaming is enabled to receive reasoning chunks:

```json
{
  "stream": true
}
```

## Usage Examples

### Basic Reasoning Request
```bash
curl -X POST https://your-ai-gateway/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "model": "anthropic.claude-3-5-sonnet-20241022-v2:0",
    "messages": [
      {
        "role": "user",
        "content": "Explain how photosynthesis works at the molecular level"
      }
    ],
    "stream": true,
    "thinking": {
      "type": "enabled",
      "budget_tokens": 1500
    }
  }'
```

### Processing Reasoning Streams
When processing the stream, handle different content types:

```python
import json
import requests

response = requests.post(
    "https://your-ai-gateway/v1/chat/completions",
    headers={"Authorization": "Bearer your-key"},
    json={
        "model": "anthropic.claude-3-5-sonnet-20241022-v2:0",
        "messages": [{"role": "user", "content": "Solve this math problem step by step: 2x + 5 = 15"}],
        "stream": True,
        "thinking": {"type": "enabled", "budget_tokens": 1000}
    },
    stream=True
)

for line in response.iter_lines():
    if line.startswith(b'data: '):
        data = json.loads(line[6:])
        for choice in data.get('choices', []):
            delta = choice.get('delta', {})
            
            # Handle reasoning content
            if 'reasoning_content' in delta:
                reasoning = delta['reasoning_content']
                if reasoning.get('text'):
                    print(f"Reasoning: {reasoning['text']}")
            
            # Handle regular content
            if delta.get('content'):
                print(f"Response: {delta['content']}")
```

## Benefits

1. **Transparency**: See the model's reasoning process
2. **Debugging**: Understand how the model arrived at its answer
3. **Quality**: Better responses through explicit reasoning
4. **Trust**: Verify the model's logic before using the final answer

## Limitations

- Reasoning tokens count toward your token usage
- Not all models support reasoning streams
- Reasoning content may be redacted for sensitive information
- Additional latency due to reasoning process

## Best Practices

1. **Set Appropriate Token Budgets**: Use `budget_tokens` to control reasoning length
2. **Handle Multiple Content Types**: Process both reasoning and regular content appropriately
3. **Verify Signatures**: Use the signature field to verify reasoning authenticity when needed
4. **Monitor Token Usage**: Reasoning can significantly increase token consumption