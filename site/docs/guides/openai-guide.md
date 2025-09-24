Based on the code analysis, I can see that this change involves refactoring OpenAI schema types to use tagged union patterns, specifically for `ChatCompletionMessageParamUnion` and `ChatCompletionContentPartUserUnionParam`. Let me create comprehensive guide documentation for this change:

# OpenAI Schema Tagged Union Implementation Guide

This guide documents the refactored OpenAI schema implementation that uses tagged union patterns for better type safety and explicit definitions in the Envoy AI Gateway.

## Overview

The AI Gateway's OpenAI schema implementation has been refactored to use tagged union patterns for union types, providing more explicit and type-safe handling of different message and content types. This change improves code maintainability and ensures consistent handling of OpenAI API structures.

## What Changed

### Tagged Union Implementation

The refactoring introduces proper tagged union patterns for two key OpenAI schema types:

1. **`ChatCompletionMessageParamUnion`** - Handles different message role types (user, assistant, system, developer, tool)
2. **`ChatCompletionContentPartUserUnionParam`** - Handles different content part types (text, image, audio, file)

### Key Improvements

- **Explicit Type Definitions**: Each union type now uses explicit field names with `Of` prefixes (e.g., `OfText`, `OfUser`, `OfAssistant`)
- **Type-Safe Unmarshaling**: JSON unmarshaling is based on discriminator fields (`role` for messages, `type` for content parts)
- **Consistent Naming**: Field names are now consistent across all union types
- **Enhanced File Support**: Added `ChatCompletionContentPartFileParam` for file content handling

## Implementation Details

### Message Union Structure

```go
type ChatCompletionMessageParamUnion struct {
    OfDeveloper *ChatCompletionDeveloperMessageParam `json:",omitzero,inline"`
    OfSystem    *ChatCompletionSystemMessageParam    `json:",omitzero,inline"`
    OfUser      *ChatCompletionUserMessageParam      `json:",omitzero,inline"`
    OfAssistant *ChatCompletionAssistantMessageParam `json:",omitzero,inline"`
    OfTool      *ChatCompletionToolMessageParam      `json:",omitzero,inline"`
}
```

### Content Part Union Structure

```go
type ChatCompletionContentPartUserUnionParam struct {
    OfText       *ChatCompletionContentPartTextParam       `json:",omitzero,inline"`
    OfInputAudio *ChatCompletionContentPartInputAudioParam `json:",omitzero,inline"`
    OfImageURL   *ChatCompletionContentPartImageParam      `json:",omitzero,inline"`
    OfFile       *ChatCompletionContentPartFileParam       `json:",omitzero,inline"`
}
```

## Usage Examples

### Working with Message Unions

```go
// Processing different message types
func processMessage(msgUnion openai.ChatCompletionMessageParamUnion) error {
    switch {
    case msgUnion.OfUser != nil:
        // Handle user message
        return processUserMessage(*msgUnion.OfUser)
    case msgUnion.OfAssistant != nil:
        // Handle assistant message
        return processAssistantMessage(*msgUnion.OfAssistant)
    case msgUnion.OfSystem != nil:
        // Handle system message
        return processSystemMessage(*msgUnion.OfSystem)
    case msgUnion.OfTool != nil:
        // Handle tool message
        return processToolMessage(*msgUnion.OfTool)
    default:
        return fmt.Errorf("unknown message type")
    }
}
```

### Working with Content Part Unions

```go
// Processing different content part types
func processContentPart(content openai.ChatCompletionContentPartUserUnionParam) error {
    switch {
    case content.OfText != nil:
        // Handle text content
        return processTextContent(*content.OfText)
    case content.OfImageURL != nil:
        // Handle image content
        return processImageContent(*content.OfImageURL)
    case content.OfInputAudio != nil:
        // Handle audio content
        return processAudioContent(*content.OfInputAudio)
    case content.OfFile != nil:
        // Handle file content
        return processFileContent(*content.OfFile)
    default:
        return fmt.Errorf("unknown content type")
    }
}
```

## JSON Marshaling/Unmarshaling

The tagged union types implement custom JSON marshaling and unmarshaling logic:

### Unmarshaling Process

1. **Discriminator Field Detection**: The unmarshaler checks for a discriminator field (`role` or `type`)
2. **Type-Specific Unmarshaling**: Based on the discriminator value, the appropriate struct is unmarshaled
3. **Union Field Assignment**: The unmarshaled struct is assigned to the corresponding union field

### Example JSON Structures

#### User Message
```json
{
    "role": "user",
    "content": "Hello, how can you help me?"
}
```

#### Text Content Part
```json
{
    "type": "text",
    "text": "This is a text message"
}
```

#### Image Content Part
```json
{
    "type": "image_url",
    "image_url": {
        "url": "https://example.com/image.jpg"
    }
}
```

## Migration Guide

### For Developers Using the Schema

If you're working with the OpenAI schema types, update your code to use the new union field names:

**Before:**
```go
// Old approach - direct field access (no longer valid)
message.Content
```

**After:**
```go
// New approach - union field access
switch {
case msgUnion.OfUser != nil:
    content := msgUnion.OfUser.Content
    // Process user content
case msgUnion.OfAssistant != nil:
    content := msgUnion.OfAssistant.Content
    // Process assistant content
}
```

### For Schema Extension

When extending the schema with new union types, follow the established pattern:

1. Use `Of` prefix for union fields
2. Implement custom JSON marshaling/unmarshaling
3. Include discriminator field validation
4. Use `omitzero,inline` JSON tags

## Benefits

### Type Safety
- Compile-time type checking ensures only one union field is populated
- Eliminates runtime type assertion errors
- Clear interface contracts for different message/content types

### Code Clarity
- Explicit field names make code more readable
- Consistent naming conventions across all union types
- Self-documenting code structure

### Maintainability
- Centralized unmarshaling logic for each union type
- Easy to extend with new message or content types
- Consistent error handling patterns

## Error Handling

The tagged union implementation includes comprehensive error handling:

```go
// Example error scenarios
func (c *ChatCompletionMessageParamUnion) UnmarshalJSON(data []byte) error {
    roleResult := gjson.GetBytes(data, "role")
    if !roleResult.Exists() {
        return errors.New("chat message does not have role")
    }
    
    role := roleResult.String()
    switch role {
    case ChatMessageRoleUser:
        // Unmarshal user message
    default:
        return fmt.Errorf("unknown ChatCompletionMessageParam type: %v", role)
    }
    return nil
}
```

## Best Practices

1. **Always Check Union Fields**: Use type switches to handle different union variants
2. **Error Handling**: Implement proper error handling for unknown types
3. **Validation**: Validate discriminator fields before processing
4. **Documentation**: Document expected JSON structures for new union types
5. **Testing**: Include comprehensive tests for all union variants

## Conclusion

The tagged union refactoring provides a robust, type-safe foundation for handling OpenAI API structures in the Envoy AI Gateway. This implementation pattern should be followed for any future schema extensions involving union types.