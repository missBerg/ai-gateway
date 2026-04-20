import React, {useState, useCallback} from 'react';
import Link from '@docusaurus/Link';
import styles from './styles.module.css';

const MCP_URL = 'https://envoy-gateway.mcp.kapa.ai';
const MCP_NAME = 'envoy-ai-gateway';

const CLAUDE_COMMAND = `claude mcp add --transport http ${MCP_NAME} ${MCP_URL}`;

// Cursor deep-link: base64-encoded JSON config
const CURSOR_CONFIG = btoa(JSON.stringify({url: MCP_URL}));
const CURSOR_URL = `cursor://anysphere.cursor-deeplink/mcp/install?name=${MCP_NAME}&config=${CURSOR_CONFIG}`;

// VS Code deep-link: URL-encoded JSON with name + type + url
const VSCODE_CONFIG = encodeURIComponent(
  JSON.stringify({name: MCP_NAME, type: 'http', url: MCP_URL}),
);
const VSCODE_URL = `vscode:mcp/install?${VSCODE_CONFIG}`;

const EXAMPLE_PROMPTS = [
  'How do I configure token rate limits?',
  'Generate a BackendSecurityPolicy for AWS Bedrock',
  'Why is my AIGatewayRoute returning 503?',
];

function CopyIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
      <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
    </svg>
  );
}

function CheckIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.4" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <polyline points="20 6 9 17 4 12" />
    </svg>
  );
}

function ArrowRightIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <line x1="5" y1="12" x2="19" y2="12" />
      <polyline points="12 5 19 12 12 19" />
    </svg>
  );
}

function CursorLogo() {
  return <img src="/img/logos/cursor.svg" alt="" width={24} height={24} />;
}

function VSCodeLogo() {
  return <img src="/img/logos/vscode.svg" alt="" width={24} height={24} />;
}

function ClaudeLogo() {
  return <img src="/img/logos/claude.svg" alt="" width={24} height={24} />;
}

function LinkIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
      <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
    </svg>
  );
}

type CopyState = null | 'claude' | 'url';

export default function McpInstall(): React.ReactElement {
  const [copied, setCopied] = useState<CopyState>(null);

  const handleCopy = useCallback(async (key: CopyState, value: string) => {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(key);
      setTimeout(() => setCopied(null), 1800);
    } catch {
      // Clipboard API unavailable.
    }
  }, []);

  return (
    <section id="mcp-install" className={styles.section}>
      <div className="container">
        <div className="sectionHeader">
          <span className="sectionEyebrow">AI-native Docs</span>
          <h2 className="sectionTitle">
            Bring the Envoy AI Gateway docs to where you do your work
          </h2>
          <p className="sectionSubtitle">
            Install the Envoy AI Gateway docs MCP server — powered by{' '}
            <a href="https://www.kapa.ai" target="_blank" rel="noreferrer">
              Kapa.ai
            </a>{' '}
            — into your coding agent and ask questions, troubleshoot configs,
            and generate YAML without leaving your editor.
          </p>
        </div>

        <div className={styles.grid}>
          {/* One-click: Cursor */}
          <a href={CURSOR_URL} className={`${styles.installCard} ${styles.oneClick}`}>
            <div className={styles.cardIcon}>
              <CursorLogo />
            </div>
            <div className={styles.cardBody}>
              <h3 className={styles.cardTitle}>Add to Cursor</h3>
              <p className={styles.cardSubtitle}>Requires Cursor · One-click install</p>
            </div>
            <ArrowRightIcon />
          </a>

          {/* One-click: VS Code */}
          <a href={VSCODE_URL} className={`${styles.installCard} ${styles.oneClick}`}>
            <div className={styles.cardIcon}>
              <VSCodeLogo />
            </div>
            <div className={styles.cardBody}>
              <h3 className={styles.cardTitle}>Add to VS Code</h3>
              <p className={styles.cardSubtitle}>Requires VS Code · One-click install</p>
            </div>
            <ArrowRightIcon />
          </a>

          {/* Claude Code: CLI copy */}
          <button
            type="button"
            onClick={() => handleCopy('claude', CLAUDE_COMMAND)}
            className={`${styles.installCard} ${styles.clickable}`}
          >
            <div className={styles.cardIcon}>
              <ClaudeLogo />
            </div>
            <div className={styles.cardBody}>
              <h3 className={styles.cardTitle}>Add to Claude Code</h3>
              <p className={styles.cardSubtitle}>
                {copied === 'claude'
                  ? 'CLI command copied'
                  : 'Requires Claude Code · Copy CLI command'}
              </p>
            </div>
            {copied === 'claude' ? <CheckIcon /> : <CopyIcon />}
          </button>

          {/* Copy URL */}
          <button
            type="button"
            onClick={() => handleCopy('url', MCP_URL)}
            className={`${styles.installCard} ${styles.clickable}`}
          >
            <div className={styles.cardIcon}>
              <LinkIcon />
            </div>
            <div className={styles.cardBody}>
              <h3 className={styles.cardTitle}>Copy MCP URL</h3>
              <p className={styles.cardSubtitle}>
                {copied === 'url'
                  ? 'URL copied'
                  : 'Paste into Claude Desktop, ChatGPT, etc.'}
              </p>
            </div>
            {copied === 'url' ? <CheckIcon /> : <CopyIcon />}
          </button>
        </div>

        <p className={styles.prereqNote}>
          These install the <strong>Envoy AI Gateway docs MCP server</strong>{' '}
          into an existing coding agent — you&apos;ll need Cursor, VS Code, or
          Claude Code installed first. New to MCP?{' '}
          <Link to="/docs/ai-assistant/install-mcp">See the full install guide →</Link>
        </p>

        <div className={styles.urlDisplay}>
          <span className={styles.urlLabel}>MCP endpoint</span>
          <code className={styles.urlValue}>{MCP_URL}</code>
        </div>

        <div className={styles.prompts}>
          <span className={styles.promptsLabel}>Try asking</span>
          <div className={styles.promptList}>
            {EXAMPLE_PROMPTS.map((prompt) => (
              <span key={prompt} className={styles.prompt}>
                {prompt}
              </span>
            ))}
          </div>
        </div>

        <div className={styles.footer}>
          <Link to="/docs/ai-assistant/install-mcp" className={styles.docsLink}>
            Full install guide →
          </Link>
        </div>
      </div>
    </section>
  );
}
