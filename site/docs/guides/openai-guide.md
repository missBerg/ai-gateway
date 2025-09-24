# OpenAI Schema Tagged Union Types Guide

This guide explains the refactored OpenAI schema implementation that uses tagged union types for better type safety and clarity in the AI Gateway.

## Overview

The AI Gateway's OpenAI schema has been refactored to use tagged union patterns for union types, providing more explicit type definitions and improved developer experience when working with chat completion messages.

## Key Changes

### ChatCompletionMessageParamUnion

The `ChatCompletionMessageParamUnion` has been restructured to use a tagged union implementation, making message type handling more explicit and consistent with other union types in the codebase.

#### Before
```go
// Less explicit union handling
type ChatCompletionMessageParamUnion interface{}
```

#### After
```go
// Tagged union with explicit type discrimination
type ChatCompletionMessageParamUnion struct {
    Type string `json:"type"`
    // Additional fields based on type
}
```

### ChatCompletionContentPartUserUnionParam

Enhanced content part handling with the addition of `ChatCompletionContentPartFileParam` and consistent naming conventions.

#### New Structure
```go
type ChatCompletionContentPartUserUnionParam struct {
    Type string `json:"type"`
    // Type-specific fields
}

type ChatCompletionContentPartFileParam struct {
    Type     string `json:"type"`
    FileID   string `json:"file_id"`
    // Additional file-related parameters
}
```

## Benefits

### Type Safety
- More explicit type checking at compile time
- Reduced runtime errors from incorrect type assumptions
- Better IDE support and autocompletion

### Code Clarity
- Clear distinction between different message types
- Consistent naming conventions across all union types
- Self-documenting type structures

### Maintainability
- Easier to extend with new message types
- Reduced complexity in type handling logic
- Better alignment with OpenAI API specifications

## Implementation Details

### Message Type Handling

When working with chat completion messages, the tagged union approach allows for explicit type checking:

```go
switch msg.Type {
case "system":
    // Handle system message
case "user":
    // Handle user message  
case "assistant":
    // Handle assistant message
case "tool":
    // Handle tool message
default:
    // Handle unknown type
}
```

### Content Part Processing

The enhanced content part union supports file attachments and maintains consistency:

```go
switch contentPart.Type {
case "text":
    // Process text content
case "image_url":
    // Process image content
case "file":
    // Process file attachment
default:
    // Handle unsupported content type
}
```

## Migration Guide

If you're working with custom extensions or modifications to the OpenAI schema handling:

1. **Update Type Checks**: Replace interface{} type assertions with tagged union type checks
2. **Use Type Field**: Access the `Type` field to determine the specific union variant
3. **Update Field Names**: Ensure field names align with the new consistent naming conventions

### Example Migration

```go
// Old approach
if userMsg, ok := msg.(ChatCompletionUserMessageParam); ok {
    // Process user message
}

// New approach
if msg.Type == "user" {
    userMsg := msg.AsUserMessage()
    // Process user message
}
```

## Related Components

This refactoring affects several internal components:

- **Translator modules**: Enhanced type safety in message translation between different AI providers
- **Request attribute handling**: More precise attribute extraction for observability
- **Schema validation**: Improved validation logic with explicit type definitions

## Best Practices

1. **Always check the Type field** before accessing type-specific fields
2. **Use type-safe accessors** when available rather than direct field access  
3. **Handle unknown types gracefully** to maintain forward compatibility
4. **Validate input early** to catch type mismatches at the API boundary

This refactoring provides a more robust foundation for handling OpenAI API messages while maintaining backward compatibility and improving the overall developer experience.