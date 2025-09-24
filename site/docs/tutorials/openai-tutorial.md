# Working with Response Format in OpenAI Chat Completions

This tutorial covers how to use the `response_format` parameter in OpenAI Chat Completions API to control the format of responses from the model.

## Overview

The `response_format` parameter allows you to specify the format that the model should use when generating responses. This is particularly useful when you need structured output or want to ensure compatibility with specific response types.

## Supported Response Formats

The AI Gateway supports three types of response formats:

### 1. Text Format (Default)

The basic text response format:

```json
{
  "response_format": {
    "type": "text"
  }
}
```

### 2. JSON Object Format

Forces the model to return valid JSON. Note that you must instruct the model to generate JSON in your system or user messages:

```json
{
  "response_format": {
    "type": "json_object"
  }
}
```

### 3. JSON Schema Format (Structured Outputs)

The most powerful format that ensures the model follows a specific JSON schema:

```json
{
  "response_format": {
    "type": "json_schema",
    "json_schema": {
      "name": "user_info",
      "description": "User information schema",
      "schema": {
        "type": "object",
        "properties": {
          "name": {
            "type": "string"
          },
          "age": {
            "type": "integer"
          },
          "email": {
            "type": "string",
            "format": "email"
          }
        },
        "required": ["name", "age"],
        "additionalProperties": false
      },
      "strict": true
    }
  }
}
```

## Complete Request Example

Here's a complete chat completion request using JSON schema format:

```json
{
  "model": "gpt-4o",
  "messages": [
    {
      "role": "system",
      "content": "You are a helpful assistant that extracts user information."
    },
    {
      "role": "user",
      "content": "Extract the information: John Doe is 30 years old and his email is john@example.com"
    }
  ],
  "response_format": {
    "type": "json_schema",
    "json_schema": {
      "name": "user_info",
      "description": "Extracted user information",
      "schema": {
        "type": "object",
        "properties": {
          "name": {
            "type": "string"
          },
          "age": {
            "type": "integer"
          },
          "email": {
            "type": "string"
          }
        },
        "required": ["name", "age", "email"]
      },
      "strict": true
    }
  }
}
```

## Key Features

### Type Safety

The response format is implemented with proper type safety using union types. The system automatically handles different format types:

- `ChatCompletionResponseFormatTextParam` for text responses
- `ChatCompletionResponseFormatJSONObjectParam` for JSON object responses
- `ChatCompletionResponseFormatJSONSchema` for structured JSON schema responses

### Schema Validation

When using `json_schema` format:
- Set `strict: true` to enforce exact schema adherence
- The model will always follow the defined JSON schema
- Only a subset of JSON Schema is supported in strict mode

### Error Handling

The system includes proper error handling for:
- Missing `type` field in response format
- Unsupported response format types
- Invalid JSON schema definitions

## Best Practices

1. **Use JSON Schema for Structured Data**: When you need consistent, structured output, prefer `json_schema` over `json_object`

2. **Include Clear Descriptions**: Provide descriptive names and descriptions in your JSON schema to help the model understand the expected format

3. **Set Appropriate Constraints**: Use `required` fields and `additionalProperties: false` to ensure the response matches your expectations exactly

4. **Handle All Response Types**: Your application should be prepared to handle different response format types depending on your configuration

## Provider Compatibility

This response format feature works across different AI providers supported by the gateway, with automatic translation between OpenAI format and provider-specific formats like Google Vertex AI's Gemini models.