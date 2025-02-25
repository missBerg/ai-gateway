# Provider Integration

Envoy AI Gateway supports seamless integration with various LLM providers through a flexible configuration system that handles authentication, request/response transformation, and usage tracking.

## Currently Supported Providers

- OpenAI - Access OpenAI's GPT models through their API
- AWS Bedrock - Connect to AWS Bedrock's foundation models

## Upstream Provider Authentication Configuration

The gateway supports multiple authentication mechanisms through the `BackendSecurityPolicy` resource:

### API Key Authentication

For providers that use API key authentication (like OpenAI):

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: BackendSecurityPolicy
metadata:
  name: openai-auth
  namespace: default
spec:
  type: APIKey
  apiKey:
    secretRef:
      name: openai-api-key
      namespace: default
```

### AWS Credentials

For AWS Bedrock, you can authenticate using either AWS credentials file or OIDC token exchange:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: BackendSecurityPolicy
metadata:
  name: aws-auth
  namespace: default
spec:
  type: AWSCredentials
  awsCredentials:
    region: us-east-1
    # Option 1: Using credentials file
    credentialsFile:
      secretRef:
        name: aws-credentials
        namespace: default
      profile: default

    # Option 2: Using OIDC token exchange
    oidcExchangeToken:
      awsRoleArn: "arn:aws:iam::123456789012:role/AIGatewayRole"
      oidc:
        provider:
          issuer: "https://your-oidc-provider"
          tokenEndpoint: "https://your-oidc-provider/token"
        clientID: "your-client-id"
        clientSecret:
          name: oidc-client-secret
          namespace: default
        scopes:
          - "openid"
          - "profile"
```

## Request and Response Transformation

The gateway automatically handles request transformation between your standardized API and the provider-specific formats.

This is configured through the API schema specification:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGateway
metadata:
  name: my-gateway
spec:
  schema: OpenAI
```

The `schema` field determines both the expected input format for client requests and how responses are normalized. Currently supported schemas include:

- `OpenAI`: Uses the OpenAI API format as the standard interface. This is recommended for most use cases as it's widely adopted across the industry.
- `Raw`: Passes requests through without transformation. Useful when you need direct control over provider-specific formats.

When using the OpenAI schema:
- Requests must follow the OpenAI chat completion format
- Responses from all providers are transformed to match OpenAI's response structure
- Token usage information is normalized across providers
- Streaming responses use the same Server-Sent Events (SSE) format as OpenAI

Example OpenAI-compatible request format:
```json
{
  "model": "gpt-4",
  "messages": [
    {"role": "user", "content": "Hello, how are you?"}
  ],
  "temperature": 0.7
}
```

## Model Selection

### Basic Model Routing

Models are selected using the `x-ai-eg-model` header. The gateway extracts the model name from the request content and uses it for routing decisions. The routing process works as follows:

1. The model name is extracted from the request body (for example, from the `model` field in OpenAI-compatible requests)
2. The model name is set in the `x-ai-eg-model` header
3. The gateway uses this header to match against routing rules defined in your `AIGatewayRoute` configuration
4. Once a matching rule is found, the request is routed to the appropriate backend(s)

### Multiple Models to Same Provider

Example routing configuration:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGatewayRoute
metadata:
  name: my-route
spec:
  rules:
    # Route different OpenAI models to the same backend
    - matches:
        - headers:
            - type: Exact
              name: x-ai-eg-model
              value: gpt-4
      backendRefs:
        - name: openai-backend
          weight: 100
    - matches:
        - headers:
            - type: Exact
              name: x-ai-eg-model
              value: gpt-3.5-turbo
      backendRefs:
        - name: openai-backend  # Same backend as gpt-4
          weight: 100

    # Route to a different provider
    - matches:
        - headers:
            - type: Exact
              name: x-ai-eg-model
              value: claude-3-sonnet
      backendRefs:
        - name: anthropic-backend
          weight: 100
```

You can map multiple models to the same provider backend by creating separate routing rules that reference the same `backendRef`. This is useful when you want to:

1. Support multiple models from the same provider (e.g., different GPT models from OpenAI)
2. Implement A/B testing between models while using the same backend infrastructure
3. Manage different model versions through the same provider endpoint

The `weight` field in `backendRefs` allows you to implement traffic splitting between different backends when needed, though it's typically set to 100 when routing to a single backend.

The gateway also provides a `/v1/models` endpoint that returns a list of all available models configured in your routing rules.

## Response Handling

The gateway provides:

1. Response format normalization across providers
2. Token usage tracking and metadata extraction based on normalized response
