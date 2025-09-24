# OpenAI Schema Refactoring Guide

This guide documents the refactoring of OpenAI message schema unions to use tagged union implementation for improved type safety and consistency across the AI Gateway codebase.

## Overview

The changes introduce a consistent tagged union approach for handling polymorphic OpenAI message types. This refactoring improves schema validation, type safety, and makes the API implementation more explicit and maintainable.

## Key Changes

### 1. ChatCompletionMessageParamUnion Refactoring

The `ChatCompletionMessageParamUnion` type has been refactored to use the same tagged union implementation pattern as other union types in the codebase.

**Before:**
```go
type ChatCompletionMessageParamUnion interface {
    // Various message implementations
}
```

**After:**
```go
type ChatCompletionMessageParamUnion struct {
    OfDeveloper *ChatCompletionDeveloperMessageParam `json:",omitzero,inline"`
    OfSystem    *ChatCompletionSystemMessageParam    `json:",omitzero,inline"`
    OfUser      *ChatCompletionUserMessageParam      `json:",omitzero,inline"`
    OfAssistant *ChatCompletionAssistantMessageParam `json:",omitzero,inline"`
    OfTool      *ChatCompletionToolMessageParam      `json:",omitzero,inline"`
}
```

### 2. Enhanced Content Part Support

Added support for file content parts and improved naming consistency:

**New File Content Support:**
```go
type ChatCompletionContentPartFileParam struct {
    File ChatCompletionContentPartFileFileParam `json:"file,omitzero"`
    Type ChatCompletionContentPartFileType      `json:"type"`
}
```

**Updated User Content Union:**
```go
type ChatCompletionContentPartUserUnionParam struct {
    OfText       *ChatCompletionContentPartTextParam       `json:",omitzero,inline"`
    OfInputAudio *ChatCompletionContentPartInputAudioParam `json:",omitzero,inline"`
    OfImageURL   *ChatCompletionContentPartImageParam      `json:",omitzero,inline"`
    OfFile       *ChatCompletionContentPartFileParam       `json:",omitzero,inline"`
}
```

### 3. Improved JSON Marshaling/Unmarshaling

The tagged union approach provides more robust JSON handling with type-safe unmarshaling based on discriminator fields.

**Message Union Unmarshaling:**
```go
func (c *ChatCompletionMessageParamUnion) UnmarshalJSON(data []byte) error {
    roleResult := gjson.GetBytes(data, "role")
    if !roleResult.Exists() {
        return errors.New("chat message does not have role")
    }

    role := roleResult.String()
    switch role {
    case ChatMessageRoleUser:
        var userMessage ChatCompletionUserMessageParam
        if err := json.Unmarshal(data, &userMessage); err != nil {
            return err
        }
        c.OfUser = &userMessage
    // ... other cases
    }
    return nil
}
```

## Benefits

### Type Safety
- Explicit typing reduces runtime errors
- Better IDE support with autocomplete
- Compile-time validation of message types

### Consistency
- All union types now follow the same implementation pattern  
- Uniform naming conventions across content part types
- Standardized JSON handling approach

### Extensibility
- Easy to add new message or content types
- Clean separation of concerns for different message roles
- Better support for vendor-specific extensions

## Migration Guide

### For Developers Using the Schema

If you were previously working with message unions, update your code to use the new tagged union fields:

**Before:**
```go
// Direct interface usage
message := someMessageUnion
```

**After:**
```go
// Access specific message type via tagged fields
switch {
case message.OfUser != nil:
    // Handle user message
    userMsg := message.OfUser
case message.OfAssistant != nil:
    // Handle assistant message  
    assistantMsg := message.OfAssistant
// ... other cases
}
```

### For Content Parts

Update content part handling to use the new unified structure:

**Before:**
```go
// Various content part handling approaches
```

**After:**
```go
switch {
case content.OfText != nil:
    // Handle text content
case content.OfImageURL != nil:  
    // Handle image content
case content.OfFile != nil:
    // Handle file content (new)
case content.OfInputAudio != nil:
    // Handle audio content
}
```

## Implementation Details

### Files Modified

1. **`internal/apischema/openai/openai.go`** - Core schema definitions and union types
2. **`internal/extproc/translator/gemini_helper.go`** - Translation logic updates  
3. **`internal/extproc/translator/openai_awsbedrock.go`** - AWS Bedrock translator
4. **`internal/extproc/translator/openai_gcpanthropic.go`** - GCP Anthropic translator
5. **`internal/tracing/openinference/openai/request_attrs.go`** - Tracing attributes

### Backward Compatibility

The changes maintain JSON API compatibility while improving the internal type system. External API consumers should see no breaking changes in the JSON interface.

### Error Handling

Enhanced error messages provide clearer feedback when message types are missing required fields:

```go
if !roleResult.Exists() {
    return errors.New("chat message does not have role")
}
```

## Testing Considerations

When working with these changes:

1. **Type Assertions**: Update any type assertions to use the new tagged union fields
2. **JSON Tests**: Verify JSON marshaling/unmarshaling works correctly with new structure  
3. **Translation Tests**: Ensure translator functions handle all message types properly
4. **Content Parts**: Test all supported content types (text, image, audio, file)

## Future Enhancements

This refactoring enables:

- Easy addition of new message types or content parts
- Better support for streaming message handling  
- Enhanced validation and error reporting
- Improved integration with code generation tools

The tagged union approach provides a solid foundation for future OpenAI API schema evolution while maintaining type safety and code clarity.