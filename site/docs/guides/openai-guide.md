# OpenAI Schema Tagged Union Implementation Guide

This guide covers the refactored OpenAI schema implementation that introduces tagged unions for better type safety and consistency across the AI Gateway codebase.

## Overview

The AI Gateway has refactored its OpenAI schema implementation to use tagged unions for union types, providing more explicit type definitions and improved consistency. This change affects how chat completion messages and content parts are structured internally.

## What Changed

### Tagged Union Implementation

The `ChatCompletionMessageParamUnion` has been refactored to use the same tagged union implementation as other union types in the codebase. This provides:

- More explicit type definitions
- Improved type safety
- Consistent naming conventions
- Better maintainability

### Content Part Enhancements

The schema now includes `ChatCompletionContentPartFileParam` within `ChatCompletionContentPartUserUnionParam`, with updated field names for consistency:

- Standardized naming conventions across content part types
- Enhanced file parameter support
- Improved union type handling

## Technical Implementation

### Message Parameter Union Structure

The refactored message parameter union follows a tagged approach where each message type is explicitly identified:

```go
// Tagged union structure for chat completion messages
type ChatCompletionMessageParamUnion struct {
    Type string `json:"type"`
    // Type-specific fields based on the message type
}
```

### Content Part Union Types

Content parts now use consistent naming and include file parameter support:

```go
type ChatCompletionContentPartUserUnionParam struct {
    // Text content
    Text *ChatCompletionContentPartTextParam `json:"text,omitempty"`
    
    // Image content
    ImageURL *ChatCompletionContentPartImageParam `json:"image_url,omitempty"`
    
    // File content (newly added)
    File *ChatCompletionContentPartFileParam `json:"file,omitempty"`
}
```

## Impact on Components

### Translation Layer

The changes affect several translator components:

- **Gemini Helper**: Updated to work with the new tagged union structure
- **AWS Bedrock Translator**: Modified to handle the refactored message types
- **GCP Anthropic Translator**: Adapted for the new schema format

### Tracing and Observability

The OpenInference tracing implementation has been updated to properly extract attributes from the refactored request structures, ensuring observability features continue to work correctly.

## Best Practices

### Working with Tagged Unions

When implementing or extending functionality that uses these tagged unions:

1. **Always check the type field** before accessing type-specific properties
2. **Use proper type assertions** when working with union types in Go
3. **Follow the established naming conventions** for new fields or types

### Schema Evolution

When adding new content types or message parameters:

1. **Maintain backward compatibility** with existing implementations
2. **Follow the tagged union pattern** established by this refactor
3. **Update all relevant translator components** to handle new types

## Migration Considerations

This refactor is primarily internal and should not affect:

- External API contracts
- User configuration
- Existing routing behavior
- Provider integrations

The changes maintain full backward compatibility while improving the internal type system's robustness and maintainability.

## Related Resources

- [OpenAI API Schema Documentation](../api/api.mdx)
- [Provider Integration Guide](../capabilities/llm-integrations/connect-providers.md)
- [Tracing and Observability](../capabilities/observability/tracing.md)