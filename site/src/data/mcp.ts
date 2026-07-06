// Single source of truth for the docs MCP server install targets.
// Consumed by the QuickStart section's "let your coding agent do it" strip
// and the docs ai-assistant pages.

export const MCP_URL = 'https://envoy-gateway.mcp.kapa.ai';
export const MCP_NAME = 'envoy-ai-gateway';

export const CLAUDE_COMMAND = `claude mcp add --transport http ${MCP_NAME} ${MCP_URL}`;

// Cursor deep-link: base64-encoded JSON config (Cursor infers HTTP from a bare url).
// btoa is a global in both Node >=16 (SSR) and the browser, so this is stable
// across SSR + hydration.
export const CURSOR_URL = `cursor://anysphere.cursor-deeplink/mcp/install?name=${MCP_NAME}&config=${btoa(
  JSON.stringify({url: MCP_URL}),
)}`;

// VS Code deep-link: URL-encoded JSON with name + type + url ("http" required for remote).
export const VSCODE_URL = `vscode:mcp/install?${encodeURIComponent(
  JSON.stringify({name: MCP_NAME, type: 'http', url: MCP_URL}),
)}`;
