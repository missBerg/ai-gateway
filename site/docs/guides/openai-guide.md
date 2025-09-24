# OpenAI Schema Union Types Guide

This guide explains the refactored union type implementation in the OpenAI schema, specifically focusing on the tagged union approach for better type safety and consistency.

## Overview

The AI Gateway has refactored its OpenAI schema implementation to use tagged unions for union types, providing more explicit type definitions and improved consistency across the codebase. This change primarily affects how chat completion message parameters are handled internally.

## Key Changes

### ChatCompletionMessageParamUnion Refactoring

The `ChatCompletionMessageParamUnion` type has been refactored to use the same tagged union implementation as other union types in the system. This change provides:

- **More explicit type definitions**: Each union variant is clearly tagged
- **Improved type safety**: Better compile-time type checking
- **Consistent implementation**: Aligns with other union types in the codebase

### Content Part Parameters

The following updates have been made to content part parameters:

1. **New File Parameter Support**: Added `ChatCompletionContentPartFileParam` inside `ChatCompletionContentPartUserUnionParam`
2. **Consistent Naming**: Updated field names across union types for better consistency

## Implementation Details

### Tagged Union Structure

Tagged unions in the OpenAI schema now follow a consistent pattern:

```go
type ChatCompletionMessageParamUnion struct {
    Type string `json:"type"`
    // Variant-specific fields based on Type value
}
```

### Content Part Union Parameters

The content part union now supports file parameters alongside existing text and image parameters:

- `ChatCompletionContentPartTextParam`
- `ChatCompletionContentPartImageParam` 
- `ChatCompletionContentPartFileParam` (newly added)

## Benefits

### For Developers

- **Better IDE Support**: Tagged unions provide improved autocomplete and type inference
- **Clearer Code**: Type variants are more explicit and self-documenting
- **Reduced Errors**: Compile-time type checking helps catch type-related issues early

### For API Consistency

- **Unified Approach**: All union types now follow the same implementation pattern
- **Extensibility**: New union variants can be added more easily
- **Maintainability**: Consistent naming and structure across the codebase

## Migration Impact

This is an internal refactoring that maintains API compatibility. Existing configurations and API calls will continue to work without changes. The improvements are primarily in the internal type system and code organization.

## Related Components

The refactoring affects several internal components:

- **OpenAI Schema Definitions**: Core type definitions in `internal/apischema/openai/`
- **Message Translators**: Components that convert between different AI provider formats
- **Request Processing**: Internal request attribute handling
- **Tracing System**: OpenInference tracing integration

## Best Practices

When working with the updated union types:

1. **Use Type Guards**: Always check the `Type` field before accessing variant-specific fields
2. **Follow Naming Conventions**: Use the established naming patterns for consistency
3. **Leverage Type Safety**: Take advantage of the improved compile-time checking

This refactoring strengthens the foundation for future OpenAI schema enhancements while maintaining backward compatibility and improving developer experience.