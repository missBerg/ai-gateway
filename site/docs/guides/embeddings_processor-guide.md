# Per-Backend Header Mutations for Embeddings and Messages

This guide covers the header mutation capabilities for embeddings and messages processors in the AI Gateway, which allows you to modify HTTP headers when routing requests to different backend services.

## Overview

Header mutations enable you to:
- Remove sensitive headers (like authorization tokens) before forwarding requests
- Add custom headers required by specific backend services
- Transform headers to match backend-specific requirements

This functionality is now available for:
- **Embeddings processor** - for text embedding requests
- **Messages processor** - for message handling requests
- **Chat completion processor** - (existing functionality)

## Configuration

Header mutations are configured at the backend level in your AI Gateway route configuration. You can specify which headers to remove and which headers to add for each backend service.

### Basic Syntax

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGatewayRoute
metadata:
  name: example-route
spec:
  backends:
  - name: my-backend
    url: "https://api.example.com"
    # Header mutation configuration
    headerMutations:
      remove:
        - "authorization"
        - "x-api-key"
      set:
        - name: "x-custom-header"
          value: "custom-value"
        - name: "content-type"
          value: "application/json"
```

## Use Cases

### Removing Sensitive Headers

Remove authorization headers that shouldn't be forwarded to specific backends:

```yaml
backends:
- name: public-embedding-service
  url: "https://public-embeddings.example.com"
  headerMutations:
    remove:
      - "authorization"
      - "x-api-key"
      - "x-internal-token"
```

### Adding Backend-Specific Headers

Add headers required by specific backend services:

```yaml
backends:
- name: custom-embedding-provider
  url: "https://custom-provider.example.com"
  headerMutations:
    set:
      - name: "x-provider-key"
        value: "your-provider-key"
      - name: "x-model-version"
        value: "v2"
```

### Combined Mutations

Remove sensitive headers and add backend-specific ones:

```yaml
backends:
- name: secure-backend
  url: "https://secure-api.example.com"
  headerMutations:
    remove:
      - "authorization"
    set:
      - name: "x-backend-auth"
        value: "backend-specific-token"
      - name: "x-request-source"
        value: "ai-gateway"
```

## Examples by Processor Type

### Embeddings Processor

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGatewayRoute
metadata:
  name: embeddings-route
spec:
  processor: embeddings
  backends:
  - name: openai-embeddings
    url: "https://api.openai.com"
    headerMutations:
      set:
        - name: "authorization"
          value: "Bearer ${OPENAI_API_KEY}"
  - name: local-embeddings
    url: "http://local-embedding-service:8080"
    headerMutations:
      remove:
        - "authorization"
      set:
        - name: "x-local-service"
          value: "true"
```

### Messages Processor

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGatewayRoute
metadata:
  name: messages-route
spec:
  processor: messages
  backends:
  - name: anthropic-messages
    url: "https://api.anthropic.com"
    headerMutations:
      remove:
        - "x-openai-key"
      set:
        - name: "x-api-key"
          value: "${ANTHROPIC_API_KEY}"
        - name: "anthropic-version"
          value: "2023-06-01"
```

## Best Practices

### Security Considerations

1. **Remove sensitive headers** when routing to external services
2. **Use environment variables** for API keys and tokens
3. **Validate header values** before setting them

### Performance Optimization

1. **Minimize header mutations** to reduce processing overhead
2. **Use specific header names** rather than wildcards
3. **Cache header mutation rules** when possible

### Debugging

Monitor header mutations in your logs:

```yaml
# Enable debug logging for header mutations
logging:
  level: debug
  components:
    - header_mutator
```

## Troubleshooting

### Common Issues

**Headers not being removed:**
- Verify header names match exactly (case-sensitive)
- Check that the processor supports header mutations

**Custom headers not appearing:**
- Ensure proper YAML syntax in configuration
- Verify environment variable expansion if used

**Backend authentication failing:**
- Confirm header mutation order (remove then set)
- Check that required headers are being added correctly

### Validation

Test your header mutations by examining request logs or using debugging tools to verify that:
1. Specified headers are removed from outbound requests
2. Custom headers are added with correct values
3. Backend services receive expected headers

## Migration Notes

If you're upgrading from a version without header mutation support for embeddings and messages processors:

1. Review existing configurations for chat completion processors
2. Apply similar header mutation patterns to embeddings and messages routes
3. Test thoroughly in a staging environment before production deployment