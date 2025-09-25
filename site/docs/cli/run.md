---
id: aigwrun
title: aigw run
sidebar_position: 2
---

# `aigw run`

## Overview

This command runs the Envoy AI Gateway locally as a standalone proxy with a given configuration file without any dependencies such as docker or Kubernetes.
Since the project is primarily focused on the Kubernetes environment, this command is useful for testing the configuration locally before deploying it to a Kubernetes cluster.
Not only does it help in testing the configuration, but it is also useful in a local development environment of the provider-agnostic AI applications.

:::warning
Currently, `aigw run` supports Linux and macOS.
:::

## Default Proxy

By default, `aigw run` runs the AI Gateway with a default configuration that includes a proxy that listens on port `1975`.

It also includes a few default backend services:
* [Tetrate Agent Router Service (TARS)](https://router.tetrate.ai/): The API key is expected to be provided via the `TARS_API_KEY` environment variable.
* [OpenAI](https://platform.openai.com/docs/api-reference): The API key is expected to be provided via the `OPENAI_API_KEY` environment variable.
* [AWS Bedrock](https://aws.amazon.com/bedrock/)(us-east-1): The AWS credentials are expected to be provided via the conventional `~/.aws/credentials` file.
* [Ollama](https://ollama.com/): The Ollama service is expected to be running on `localhost:11434` which is the default as per [the official documentation](https://github.com/Ollama/Ollama/blob/main/docs/faq.md#how-can-i-expose-Ollama-on-my-network).

The routing configuration is as follows:
* The special header `aigw-backend-selector` can be used to select the backend regardless of the "model" field in the request body.
  * `aigw-backend-selector: tars` header will route to Tetrate Agent Router Service,
  * `aigw-backend-selector: openai` header will route to OpenAI,
  * `aigw-backend-selector: aws` will route to AWS Bedrock, and
  * `aigw-backend-selector: ollama` will route to Ollama.
* If `aigw-backend-selector` is not present, the "model" field in the request body will be used to select the backend.
  * If the "model" field is `gpt-4o-mini`, it will be routed to OpenAI,
  * If the "model" field is `us.meta.llama3-2-1b-instruct-v1:0`, it will be routed to AWS Bedrock, and
  * If the "model" field is `mistral:latest`, it will be routed to Ollama.

You can view the full configuration by running `aigw run --show-default` to see exactly what the default configuration looks like.

:::note
We welcome contributions to enhance the default configuration, for example, by adding more well-known AI providers and models.
:::

### Try it out

To run the AI Gateway with the default configuration, run the following command:
```shell
aigw run
```

Now, the AI Gateway is running locally with the default configuration serving at `localhost:1975`.

Then, open a new terminal and run the following curl commands to test the AI Gateway.
For example, use `mistral:latest` model to route to Ollama, assuming Ollama is running locally and `ollama pull mistral:latest` is executed to pull the model.

```shell
curl -H "Content-Type: application/json" -XPOST http://localhost:1975/v1/chat/completions \
    -d '{"model": "mistral:latest","messages": [{"role": "user", "content": "Say this is a test!"}]}'
```

Changing the `model` field value to `gpt-4o-mini` or `us.meta.llama3-2-1b-instruct-v1:0` will route to OpenAI or AWS Bedrock respectively.
They require the respective credentials to be available as mentioned above.

## Customize Configuration

To run the AI Gateway with a custom configuration, provide the path to the configuration file as an argument to the `aigw run` command.
For example, to run the AI Gateway with a custom configuration file named `config.yaml`, run the following command:

```shell
aigw run config.yaml
```

The configuration uses the same API as the Envoy AI Gateway custom resources definitions. See [API Reference](../api/) for more information.

The best way to start customizing the configuration is to start with the default configuration and modify it as needed.

### Modify the Default Configuration

First, save the default configuration to a file named `config.yaml` by running the following command:

```shell
aigw run --show-default > default.yaml
```

Next, let's say change the `mistral:latest` model to `deepseek-r1:1.5b` and save the configuration to `custom.yaml`:

```diff
--- default.yaml        2025-03-26 11:27:30
+++ custom.yaml 2025-03-26 11:28:55
@@ -115,9 +115,9 @@
     - matches:
         - headers:
             - type: Exact
               name: x-ai-eg-model
-              value: mistral:latest
+              value: deepseek-r1:1.5b
```

You can also use environment variable substitution (`envsubst`) to allow small
changes without needing to copy a file. For example, you could use this syntax
instead, to default the model to the `CHAT_MODEL` variable.

```diff
--- default.yaml        2025-03-26 11:27:30
+++ custom.yaml 2025-03-26 11:28:55
@@ -115,9 +115,9 @@
     - matches:
         - headers:
             - type: Exact
               name: x-ai-eg-model
-              value: mistral:latest
+              value: ${CHAT_MODEL:=deepseek-r1:1.5b}
```

### Run with the Custom Configuration

Now, run the AI Gateway with the custom configuration by running the following command:

```shell
aigw run custom.yaml
```

Now, the AI Gateway is running locally with the custom configuration serving at `localhost:1975`.

```shell
curl -H "Content-Type: application/json" -XPOST http://localhost:1975/v1/chat/completions \
    -d '{"model": "deepseek-r1:1.5b","messages": [{"role": "user", "content": "Say this is a test!"}]}'
```

### Note

* The ExtProc will serve the prometheus metrics at `localhost:1064/metrics` by default where you can scrape the [LLM/AI metrics](../capabilities/observability/metrics.md).
* For production deployments using Kubernetes, see [ExtProc Configuration](../concepts/extproc-configuration.md) to learn about configuring environment variables and secrets for the ExtProc container.

## OpenTelemetry

Envoy AI Gateway's router joins and records distributed traces when supplied
with an [OpenTelemetry](https://opentelemetry.io/) collector endpoint.

Requests to the OpenAI Chat Completions endpoint are recorded as Spans which
include typical timing and request details. In addition, there are GenAI
attributes representing the LLM call including full request and response
details, defined by the [OpenInference semantic conventions][openinference].

OpenInference attributes default to include full chat completion request and
response data. This can be toggled with configuration, but when enabled allows
systems like [Arize Phoenix][phoenix] to perform LLM evaluations of production
requests captured in OpenTelemetry spans.

### OpenTelemetry configuration

`aigw run` supports OpenTelemetry tracing via environment variables:

- **[OTEL SDK][otel-env]**: OTLP exporter configuration that controls span
  and metrics export such as:
    - `OTEL_EXPORTER_OTLP_ENDPOINT`: Collector endpoint (e.g., `http://phoenix:6006`)
    - `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`: Override traces endpoint separately
    - `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT`: Override metrics endpoint separately
    - `OTEL_TRACES_EXPORTER`: Traces exporter type (e.g., `console`, `otlp`, `none`)
    - `OTEL_METRICS_EXPORTER`: Metrics exporter type (e.g., `console`, `otlp`, `none`)
    - `OTEL_BSP_SCHEDULE_DELAY`: Batch span processor delay (default: 5000ms)
    - `OTEL_METRIC_EXPORT_INTERVAL`: Metrics export interval (default: 60000ms)

- **[OpenInference][openinference-config]**: Control sensitive data redaction,
  such as:
    - `OPENINFERENCE_HIDE_INPUTS`: Hide input messages/prompts (default: `false`)
    - `OPENINFERENCE_HIDE_OUTPUTS`: Hide output messages/completions (default: `false`)

See [docker-compose-otel.yaml][docker-compose-otel.yaml] for a complete example
configuration.

---
[openinference]: https://github.com/Arize-ai/openinference/tree/main/spec
[phoenix]: https://docs.arize.com/phoenix
[otel-env]: https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/
[openinference-config]: https://github.com/Arize-ai/openinference/blob/main/spec/configuration.md
[docker-compose-otel.yaml]: https://github.com/envoyproxy/ai-gateway/blob/main/cmd/aigw/docker-compose-otel.yaml
