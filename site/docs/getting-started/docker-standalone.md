---
id: docker-standalone
title: Docker / Standalone Quick Start
sidebar_position: 3
---

# Docker / Standalone Quick Start

This guide walks you through running Envoy AI Gateway as a standalone process using Docker. No Kubernetes cluster is required.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) installed and running
- An API key for your AI provider (e.g. OpenAI), or a locally running model server such as [Ollama](https://ollama.com/)
- `curl` for testing (installed by default on most systems)

## Quick Start with Docker

Run the AI Gateway with a single command:

```shell
docker run --rm -p 1975:1975 \
  -e OPENAI_API_KEY=$OPENAI_API_KEY \
  envoyproxy/ai-gateway-cli run
```

The gateway is now listening on `http://localhost:1975` and will proxy requests to the OpenAI API.

## Environment Variable Auto-Configuration

The gateway automatically generates a configuration when it detects OpenAI-compatible environment variables. This means you can point it at any OpenAI-compatible backend without writing a config file.

### OpenAI

```shell
docker run --rm -p 1975:1975 \
  -e OPENAI_API_KEY=$OPENAI_API_KEY \
  envoyproxy/ai-gateway-cli run
```

### Azure OpenAI

```shell
docker run --rm -p 1975:1975 \
  -e AZURE_OPENAI_ENDPOINT=https://example.openai.azure.com \
  -e AZURE_OPENAI_API_KEY=$AZURE_OPENAI_API_KEY \
  -e OPENAI_API_VERSION=2024-12-01-preview \
  envoyproxy/ai-gateway-cli run
```

### Ollama (local models)

```shell
docker run --rm -p 1975:1975 \
  -e OPENAI_BASE_URL=http://host.docker.internal:11434/v1 \
  -e OPENAI_API_KEY=unused \
  --add-host=host.docker.internal:host-gateway \
  envoyproxy/ai-gateway-cli run
```

:::tip
`host.docker.internal` allows the container to reach services on your host machine, such as an Ollama server running on `localhost:11434`.
:::

## Docker Compose

For a more reproducible setup, use Docker Compose. Create a `docker-compose.yaml` file:

```yaml
services:
  aigw:
    image: envoyproxy/ai-gateway-cli:latest
    container_name: aigw
    environment:
      - OPENAI_BASE_URL=http://host.docker.internal:11434/v1
      - OPENAI_API_KEY=unused
    ports:
      - "1975:1975" # OpenAI-compatible API
      - "1064:1064" # Admin: /metrics and /health
    extra_hosts:
      - "host.docker.internal:host-gateway"
    command: ["run"]
```

Start the stack:

```shell
docker compose up -d
```

:::note
The example above targets Ollama running on the host. Replace the `OPENAI_BASE_URL` and `OPENAI_API_KEY` variables to point at a different backend such as OpenAI or Azure OpenAI.
:::

## Custom Configuration

To use a custom YAML configuration file, mount it into the container:

```shell
docker run --rm -p 1975:1975 \
  -v $(pwd)/config.yaml:/etc/aigw/config.yaml \
  envoyproxy/ai-gateway-cli run /etc/aigw/config.yaml
```

The configuration uses the same API as the Envoy AI Gateway custom resource definitions. See the [API Reference](/docs/api/) and the [example configurations](https://github.com/envoyproxy/ai-gateway/tree/main/examples) for a starting point.

## Test the Gateway

Open a new terminal and send a request:

```shell
curl -H "Content-Type: application/json" \
  -d '{"model": "gpt-4o-mini", "messages": [{"role": "user", "content": "Say this is a test!"}]}' \
  http://localhost:1975/v1/chat/completions
```

:::tip
If you are using Ollama, replace the model name with one you have pulled locally (e.g. `qwen2.5:0.5b`).
:::

## Next Steps

- See the [CLI reference](/docs/cli/aigwrun) for the full set of `aigw run` options, including MCP gateway support, OpenTelemetry tracing, and advanced configuration
- Explore [Docker Compose with OpenTelemetry](https://github.com/envoyproxy/ai-gateway/blob/main/cmd/aigw/docker-compose-otel.yaml) for production-style observability
- Browse the [example configurations](https://github.com/envoyproxy/ai-gateway/tree/main/examples) to customize routing, rate limiting, and more
