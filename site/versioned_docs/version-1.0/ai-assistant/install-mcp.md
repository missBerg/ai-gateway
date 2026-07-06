---
id: install-mcp
title: Install the AI Gateway MCP in your coding agent
sidebar_label: Install the MCP server
sidebar_position: 1
---

# Install the Envoy AI Gateway MCP in your coding agent

The Envoy AI Gateway project ships a hosted [Model Context Protocol](https://modelcontextprotocol.io)
server powered by [Kapa](https://www.kapa.ai/). Point your coding agent at it and you can query the
docs, troubleshoot configurations, and generate YAML — without ever leaving your editor.

The MCP server indexes:

- Envoy AI Gateway documentation
- Envoy Gateway documentation
- Envoy Proxy documentation

The endpoint is:

```
https://envoy-gateway.mcp.kapa.ai
```

## One-click install

If you use Cursor or VS Code, install it in one click:

- **Cursor** — [Add to Cursor](cursor://anysphere.cursor-deeplink/mcp/install?name=envoy-ai-gateway&config=eyJ1cmwiOiJodHRwczovL2Vudm95LWdhdGV3YXkubWNwLmthcGEuYWkifQ==)
- **VS Code** — [Add to VS Code](vscode:mcp/install?%7B%22name%22%3A%22envoy-ai-gateway%22%2C%22type%22%3A%22http%22%2C%22url%22%3A%22https%3A%2F%2Fenvoy-gateway.mcp.kapa.ai%22%7D)

The buttons on the [homepage](/) do the same thing.

## Install in Claude Code

Run the following command in your terminal:

```bash
claude mcp add --transport http envoy-ai-gateway https://envoy-gateway.mcp.kapa.ai
```

That's it. Claude Code will pick up the new MCP server on the next session.

### Or: add it manually

If you prefer to edit your Claude Code config by hand, open `~/.claude/mcp.json` and add:

```json
{
  "mcpServers": {
    "envoy-ai-gateway": {
      "transport": "http",
      "url": "https://envoy-gateway.mcp.kapa.ai"
    }
  }
}
```

## Install in Cursor or VS Code manually

Both editors accept a remote (HTTP) MCP server. Use this configuration:

```json
{
  "name": "envoy-ai-gateway",
  "type": "http",
  "url": "https://envoy-gateway.mcp.kapa.ai"
}
```

In VS Code, add it under `mcp.servers` in your settings; in Cursor, add it to your MCP servers list.

## Verify it works

Start a fresh session and try one of these questions:

- _"What LLM providers does Envoy AI Gateway support?"_
- _"How do I configure token-based rate limiting?"_
- _"Generate an AIGatewayRoute that splits traffic 80/20 between OpenAI and Bedrock."_
- _"Why would an AIGatewayRoute return 503?"_

You should see your agent call the `envoy-ai-gateway` tool and return answers backed by the official docs.

## What it's good for

- **Config generation.** Ask for a starting-point CRD and iterate. _"Generate a BackendSecurityPolicy for AWS Bedrock using IRSA."_
- **Troubleshooting.** Paste a log line or status condition and ask what it means.
- **Capability lookup.** _"Does AI Gateway support prompt caching for Anthropic? Which version introduced it?"_
- **Upgrade planning.** _"What changed between v0.5 and v0.6? Anything breaking?"_

## Privacy

Queries sent through this MCP server are forwarded to Kapa's hosted infrastructure so the model can
retrieve relevant documentation chunks. No Envoy AI Gateway traffic, telemetry, or cluster data passes
through it — only the text of your questions and the docs context Kapa returns. See
[Kapa's privacy policy](https://www.kapa.ai/privacy-policy) for details.

## Prefer to ask in the browser?

Every page on this site has a floating **Ask AI** button in the bottom-right corner, backed by the same
Kapa-indexed documentation. Use whichever surface fits your workflow.

## Feedback

Something inaccurate or missing? Open a
[GitHub Discussion](https://github.com/envoyproxy/ai-gateway/discussions) and we'll improve the indexed content.
