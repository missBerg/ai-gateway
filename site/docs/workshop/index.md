---
title: Envoy AI Gateway Workshop
sidebar_label: "Envoy AI Gateway Workshop"
sidebar_position: 1
---

# Workshop Outline

## Hour 1: Foundation and Basic Setup (60 mins)

### Introduction (15 mins)
- What is AI Gateway?
  - A Kubernetes-native solution for managing AI/LLM traffic using Envoy
  - Built on top of Envoy Gateway for robust traffic management
  - Supports multiple AI providers (OpenAI, AWS Bedrock, Azure OpenAI)

- Architecture overview
  - Control plane components
  - Data plane (Envoy proxy)
  - Integration with Kubernetes Gateway API

- Use cases and benefits
  - Centralized AI traffic management
  - Multi-provider support
  - Rate limiting and traffic shaping
  - Request/response transformation

- Workshop objectives
  - Set up a local AI Gateway environment
  - Configure and test with local LLM (Ollama)
  - Learn about advanced features

### Environment Setup (20 mins)

#### Prerequisites
1. Verify tool installations:
```shell
kubectl version --client
helm version
curl --version
```

2. Set up Kubernetes cluster:
   - For Docker Desktop users:
     1. Open Docker Desktop
     2. Go to Settings → Kubernetes
     3. Enable Kubernetes
     4. Click "Apply & Restart"
     5. Verify:
     ```shell
     kubectl config use-context docker-desktop
     kubectl cluster-info
     ```

   - For kind users:
     ```shell
     brew install kind
     kind create cluster
     ```

#### Install AI Gateway

1. Install Envoy Gateway:
```shell
helm upgrade -i eg oci://docker.io/envoyproxy/gateway-helm \
    --version v0.0.0-latest \
    --namespace envoy-gateway-system \
    --create-namespace

kubectl wait --timeout=2m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
```

2. Install AI Gateway:
```shell
helm upgrade -i aieg oci://docker.io/envoyproxy/ai-gateway-helm \
    --version v0.0.0-latest \
    --namespace envoy-ai-gateway-system \
    --create-namespace

kubectl wait --timeout=2m -n envoy-ai-gateway-system deployment/ai-gateway-controller --for=condition=Available
```

3. Apply AI Gateway configurations:
```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/ai-gateway/main/manifests/envoy-gateway-config/redis.yaml
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/ai-gateway/main/manifests/envoy-gateway-config/config.yaml
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/ai-gateway/main/manifests/envoy-gateway-config/rbac.yaml

kubectl rollout restart -n envoy-gateway-system deployment/envoy-gateway

kubectl wait --timeout=2m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
```

4. Verify installations:
```shell
kubectl get pods -n envoy-ai-gateway-system
```
```shell
kubectl get pods -n envoy-gateway-system
```

### First Steps (15 mins)

1. Understanding the Gateway API structure:
   - AIGatewayRoute: Defines routing rules
   - AIServiceBackend: Defines AI service endpoints
   - Schema configuration: OpenAI-compatible API format

2. Basic test configuration:
Download the basic configuration file:
```shell
curl -o basic.yaml https://raw.githubusercontent.com/envoyproxy/ai-gateway/refs/tags/v0.1.4/examples/basic/basic.yaml
```

Apply the configuration:
```shell
kubectl apply -f basic.yaml
```

3. Test with curl:

First, check if your Gateway has an external IP address assigned:

```shell
kubectl get svc -n envoy-gateway-system \
    --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=envoy-ai-gateway-basic
```

```shell
export GATEWAY_URL=$(kubectl get gateway/envoy-ai-gateway-basic -o jsonpath='{.status.addresses[0].value}')
```

```shell
echo $GATEWAY_URL
```
```shell
curl -H "Content-Type: application/json" \
    -d '{
        "model": "some-cool-self-hosted-model",
        "messages": [
            {
                "role": "system",
                "content": "Hi."
            }
        ]
    }' \
    $GATEWAY_URL/v1/chat/completions
```

4. Common troubleshooting:
   ```shell
   kubectl logs -n envoy-gateway-system deployment/envoy-gateway
   kubectl logs -n envoy-ai-gateway-system deployment/ai-gateway-controller
   ```


10 min break

## Hour 2: Local AI Integration (60 mins)

### Setting up Ollama (20 mins)
1. Install Ollama:
```shell
curl https://ollama.ai/install.sh | sh
```

2. Start Ollama:
```shell
ollama serve
```

3. Pull the mistral:latest model:
```shell
ollama pull mistral:latest
```

4. Test Ollama locally:
```shell
curl http://localhost:11434/api/chat -d '{
  "model": "mistral:latest",
  "messages": [
    { "role": "user", "content": "why is the sky blue?" }
  ]
}'

ollama run mistral:latest "why is the sky blue?"
```

### Configuring AI Gateway with Ollama (25 mins)

1. Create Ollama backend configuration.
Add the following to your basic.yaml file:

```yaml
---
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIServiceBackend
metadata:
  name: ollama-backend
  namespace: default
spec:
  schema:
    name: OpenAI
  backendRef:
    name: ollama-service
    kind: Service
    port: 11434
---
apiVersion: v1
kind: Service
metadata:
  name: ollama-service
  namespace: default
spec:
  type: ExternalName
  # For Kind use kind-control-plane
  # For Docker desktop use host.docker.internal
  externalName: host.docker.internal
  ports:
    - port: 11434
```

2. Apply configuration:
```shell
kubectl apply -f basic.yaml
```

Open basic.yaml and add the ollama backend to the AI Gateway route:

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGatewayRoute
metadata:
  name: envoy-ai-gateway-basic
  namespace: default
spec:
  schema:
    name: OpenAI
  targetRefs:
    - name: envoy-ai-gateway-basic
      kind: Gateway
      group: gateway.networking.k8s.io
  rules:
    - matches:
        - headers:
            - type: Exact
              name: x-ai-eg-model
              value: gpt-4o-mini
      backendRefs:
        - name: envoy-ai-gateway-basic-openai
    - matches:
        - headers:
            - type: Exact
              name: x-ai-eg-model
              value: us.meta.llama3-2-1b-instruct-v1:0
      backendRefs:
        - name: envoy-ai-gateway-basic-aws
    - matches:
        - headers:
            - type: Exact
              name: x-ai-eg-model
              value: some-cool-self-hosted-model
      backendRefs:
        - name: envoy-ai-gateway-basic-testupstream
    - matches:
        - headers:
            - type: Exact
              name: x-ai-eg-model
              value: mistral:latest
      backendRefs:
        - name: ollama-backend
---
```

### Advanced Request Handling (10 mins)

1. Understanding request transformation:
   - OpenAI format to Ollama format
   - Response mapping
   - Error handling

2. Token usage limiting example:
```yaml
spec:
  rateLimit:
    tokenBucket:
      fillInterval: 60s
      capacity: 1000
      tokens: 1000
```

10 min break

## Hour 3: Advanced Features and Practical Applications (60 mins)

### Adding security controls (20 mins)

1. Add an API key requirement

update `basic.yaml`

```yaml
---
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: apikey-secret
stringData:
  client1: supersecret
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: apikey-auth-example
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: envoy-ai-gateway-basic
  apiKeyAuth:
    credentialRefs:
    - group: ""
      kind: Secret
      name: apikey-secret
    extractFrom:
    - headers:
      - x-api-key
```

```shell
curl -v -H "Host: ai.example.com" "http://$GATEWAY_URL/v1/chat/completions" \
    -H "Content-Type: application/json" \
    -H "x-api-key: supersecret" \
    -d '{
      "model": "mistral:latest",
      "messages": [{"role": "user", "content": "why is the sky blue?"}]
    }'
```

### Monitoring and Debugging (20 mins)

1. View gateway logs:
```shell
kubectl logs -n envoy-ai-gateway-system deployment/ai-gateway-controller
```
```shell
kubectl logs -n envoy-gateway-system deployment/envoy-gateway
```

2. Monitor metrics:
```shell
kubectl port-forward -n envoy-gateway-system deployment/envoy-gateway 9090:9090

open http://localhost:9090/metrics
```

3. Troubleshooting techniques:
   - Check Gateway status
   - Verify backend connectivity
   - Debug request/response transformation
   - Common error patterns

4. Performance monitoring:
   - Request latency
   - Token usage
   - Error rates
   - Backend health
