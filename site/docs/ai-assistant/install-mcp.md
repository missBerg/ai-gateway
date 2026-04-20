---
id: install-mcp
title: Install the AI Gateway MCP in your coding agent
sidebar_label: Install MCP in Claude Code
sidebar_position: 1
---

# Install the Envoy AI Gateway MCP in your coding agent

The Envoy AI Gateway project ships a hosted [Model Context Protocol](https://modelcontextprotocol.io) server powered by [Kapa](https://www.kapa.ai/). Point your coding agent at it and you can query the docs, troubleshoot configurations, and generate YAML — without ever leaving your editor.

The MCP server indexes:

- Envoy AI Gateway documentation
- Envoy Gateway documentation
- Envoy Proxy documentation

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

## Verify it works

Start a fresh Claude Code session and try one of these questions:

- _"What LLM providers does Envoy AI Gateway support?"_
- _"How do I configure token-based rate limiting?"_
- _"Generate an AIGatewayRoute that splits traffic 80/20 between OpenAI and Bedrock."_
- _"Why would an AIGatewayRoute return 503?"_

You should see Claude Code call the `envoy-ai-gateway` tool and return answers backed by the official docs.

## Example queries to try

Once installed, the MCP server is most useful for:

- **Config generation.** Ask for a starting-point CRD and iterate from there. _"Generate a BackendSecurityPolicy for AWS Bedrock using IRSA."_
- **Troubleshooting.** Paste a log line or status condition and ask what it means.
- **Capability lookup.** _"Does AI Gateway support prompt caching for Anthropic? Which version introduced it?"_
- **Upgrade planning.** _"What changed between v0.4 and v0.5? Anything breaking?"_

## Privacy

Queries sent through this MCP server are forwarded to Kapa's hosted infrastructure so the LLM can retrieve relevant documentation chunks. No Envoy AI Gateway traffic, telemetry, or cluster data passes through it — only the text of your questions and the docs context Kapa returns. See [Kapa's privacy policy](https://www.kapa.ai/privacy-policy) for details.

## Prefer to ask in the browser?

Every page on this site has a floating **Ask AI** button in the bottom-right corner, backed by the same Kapa-indexed documentation. Use whichever surface fits your workflow.

## Feedback

Something inaccurate? Missing? Open a [GitHub Discussion](https://github.com/envoyproxy/ai-gateway/discussions) and we'll improve the indexed content.
