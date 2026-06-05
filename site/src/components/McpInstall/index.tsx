import React, {useState, useCallback} from 'react';
import Link from '@docusaurus/Link';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

const MCP_URL = 'https://envoy-gateway.mcp.kapa.ai';
const MCP_NAME = 'envoy-ai-gateway';

const CLAUDE_COMMAND = `claude mcp add --transport http ${MCP_NAME} ${MCP_URL}`;

// Cursor deep-link: base64-encoded JSON config (Cursor infers HTTP from a bare url).
// btoa is a global in both Node >=16 (SSR) and the browser, so this is stable
// across SSR + hydration.
const CURSOR_CONFIG = btoa(JSON.stringify({url: MCP_URL}));
const CURSOR_URL = `cursor://anysphere.cursor-deeplink/mcp/install?name=${MCP_NAME}&config=${CURSOR_CONFIG}`;

// VS Code deep-link: URL-encoded JSON with name + type + url ("http" required for remote).
const VSCODE_CONFIG = encodeURIComponent(JSON.stringify({name: MCP_NAME, type: 'http', url: MCP_URL}));
const VSCODE_URL = `vscode:mcp/install?${VSCODE_CONFIG}`;

function CopyIcon(): React.ReactElement {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
      <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
    </svg>
  );
}

function CheckIcon(): React.ReactElement {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.4" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <polyline points="20 6 9 17 4 12" />
    </svg>
  );
}

function ArrowRightIcon(): React.ReactElement {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <line x1="5" y1="12" x2="19" y2="12" />
      <polyline points="12 5 19 12 12 19" />
    </svg>
  );
}

function LinkIcon(): React.ReactElement {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
      <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
    </svg>
  );
}

const AgentLogo = ({src}: {src: string}): React.ReactElement => (
  <img src={src} alt="" width={24} height={24} />
);

type CopyState = null | 'claude' | 'url' | 'endpoint';

export default function McpInstall(): React.ReactElement {
  const [copied, setCopied] = useState<CopyState>(null);

  const handleCopy = useCallback(async (key: Exclude<CopyState, null>, value: string) => {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(key);
      setTimeout(() => setCopied(null), 1800);
    } catch {
      /* clipboard unavailable */
    }
  }, []);

  return (
    <section id="mcp-install" className={styles.section} aria-labelledby="mcp-install-heading">
      <div className="container">
        <div className={styles.split}>
          {/* Left column: heading + blurb */}
          <div className={styles.intro}>
            <span className="sectionEyebrow sectionEyebrow--accent">AI-native docs</span>
            <Heading as="h2" id="mcp-install-heading" className={styles.introTitle}>
              Bring the docs into your coding agent
            </Heading>
            <p className={styles.introText}>
              Install the Envoy AI Gateway docs MCP server — powered by{' '}
              <a href="https://www.kapa.ai" target="_blank" rel="noreferrer">Kapa.ai</a> — into your
              agent and ask questions, troubleshoot configs, and generate YAML without leaving your editor.
            </p>
            <p className={styles.prereqNote}>
              Works with an existing coding agent — you&apos;ll need Cursor, VS Code, or Claude Code
              first. New to MCP? <Link to="/docs/ai-assistant/install-mcp">See the full install guide →</Link>
            </p>
          </div>

          {/* Right column: install actions + endpoint */}
          <div className={styles.actions}>
            <div className={styles.grid}>
              <a href={CURSOR_URL} className={`${styles.installCard} ${styles.oneClick}`}>
                <span className={styles.cardIcon}><AgentLogo src="/img/logos/cursor.svg" /></span>
                <span className={styles.cardBody}>
                  <span className={styles.cardTitle}>Add to Cursor</span>
                </span>
                <ArrowRightIcon />
              </a>

              <a href={VSCODE_URL} className={`${styles.installCard} ${styles.oneClick}`}>
                <span className={styles.cardIcon}><AgentLogo src="/img/logos/vscode.svg" /></span>
                <span className={styles.cardBody}>
                  <span className={styles.cardTitle}>Add to VS Code</span>
                </span>
                <ArrowRightIcon />
              </a>

              <button
                type="button"
                onClick={() => handleCopy('claude', CLAUDE_COMMAND)}
                className={styles.installCard}>
                <span className={styles.cardIcon}><AgentLogo src="/img/logos/claude.svg" /></span>
                <span className={styles.cardBody}>
                  <span className={styles.cardTitle}>Add to Claude Code</span>
                  {copied === 'claude' && (
                    <span className={styles.cardSubtitle}>CLI command copied</span>
                  )}
                </span>
                {copied === 'claude' ? <CheckIcon /> : <CopyIcon />}
              </button>

              <button
                type="button"
                onClick={() => handleCopy('url', MCP_URL)}
                className={styles.installCard}>
                <span className={styles.cardIcon}><LinkIcon /></span>
                <span className={styles.cardBody}>
                  <span className={styles.cardTitle}>Copy MCP URL</span>
                  {copied === 'url' && (
                    <span className={styles.cardSubtitle}>URL copied</span>
                  )}
                </span>
                {copied === 'url' ? <CheckIcon /> : <CopyIcon />}
              </button>
            </div>

            <div className={styles.endpointWrap}>
              <span className={styles.endpointCaption}>MCP server endpoint</span>
              <button
                type="button"
                className={styles.endpointField}
                onClick={() => handleCopy('endpoint', MCP_URL)}
                aria-label={copied === 'endpoint' ? 'Copied MCP endpoint URL' : 'Copy MCP endpoint URL'}>
                <code className={styles.endpointUrl}>{MCP_URL}</code>
                <span className={styles.endpointAction}>
                  {copied === 'endpoint' ? <CheckIcon /> : <CopyIcon />}
                  {copied === 'endpoint' ? 'Copied' : 'Copy'}
                </span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
