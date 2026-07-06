import React, {useState, useCallback} from 'react';
import Link from '@docusaurus/Link';
import Heading from '@theme/Heading';
import CodeBlock from '@theme/CodeBlock';
import versions from '@site/versions.json';
import {MCP_URL, CLAUDE_COMMAND, CURSOR_URL, VSCODE_URL} from '@site/src/data/mcp';
import styles from './styles.module.css';

// Versions come from the LATEST RELEASED docs, so the landing commands always
// show a pinned, installable release — never main-branch 0.0.0-latest builds.
// The webpack context require follows versions.json[0], so cutting a new docs
// version updates these commands with no extra edits.
type Vars = {aigwVersion: string; egVersion: string; k8sMinVersion: string};
// eslint-disable-next-line @typescript-eslint/no-var-requires
const vars = require(`@site/versioned_docs/version-${versions[0]}/_vars.json`) as Vars;

const AIGW = `v${vars.aigwVersion}`;
const EG = `v${vars.egVersion}`;
// Pin the values file to the release tag — the versioned docs leave
// aigwGitRef as 'main', which is not a stable ref for install commands.
const REF = `v${vars.aigwVersion}`;

const STEP_EG = `helm upgrade -i eg oci://docker.io/envoyproxy/gateway-helm \\
    --version ${EG} \\
    --namespace envoy-gateway-system --create-namespace \\
    -f https://raw.githubusercontent.com/envoyproxy/ai-gateway/${REF}/manifests/envoy-gateway-values.yaml

kubectl wait --timeout=2m -n envoy-gateway-system \\
    deployment/envoy-gateway --for=condition=Available`;

const STEP_CRDS = `helm upgrade -i aieg-crd oci://docker.io/envoyproxy/ai-gateway-crds-helm \\
    --version ${AIGW} \\
    --namespace envoy-ai-gateway-system --create-namespace`;

const STEP_CTRL = `helm upgrade -i aieg oci://docker.io/envoyproxy/ai-gateway-helm \\
    --version ${AIGW} \\
    --namespace envoy-ai-gateway-system --create-namespace

kubectl wait --timeout=2m -n envoy-ai-gateway-system \\
    deployment/ai-gateway-controller --for=condition=Available`;

const STEPS = [
  {
    n: 1,
    tab: 'Envoy Gateway',
    title: 'Install Envoy Gateway',
    desc: 'With the AI Gateway values file applied.',
    code: STEP_EG,
  },
  {
    n: 2,
    tab: 'AI Gateway CRDs',
    title: 'Install the AI Gateway CRDs',
    desc: 'The custom resource definitions.',
    code: STEP_CRDS,
  },
  {
    n: 3,
    tab: 'Controller',
    title: 'Install the AI Gateway controller',
    desc: 'Then wait for it to become available.',
    code: STEP_CTRL,
  },
];

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
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
      <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
    </svg>
  );
}

type CopyState = null | 'claude' | 'url';

/* "Skip the copy-paste" strip — the docs MCP install, kept inside Quick start
   so the agent-driven on-ramp sits right next to the manual one. */
function McpStrip(): React.ReactElement {
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
    <div id="mcp-install" className={styles.mcpRow}>
      <div className={styles.mcpIntro}>
        <Heading as="h3" className={styles.mcpTitle}>
          Or let your coding agent do it
        </Heading>
        <p className={styles.mcpText}>
          Add the docs MCP server — powered by{' '}
          <a href="https://www.kapa.ai" target="_blank" rel="noreferrer">Kapa.ai</a> — and ask
          your agent to install, troubleshoot, and generate config for you.{' '}
          <Link to="/docs/ai-assistant/install-mcp">Full guide →</Link>
        </p>
      </div>
      <div className={styles.mcpActions}>
        <a href={CURSOR_URL} className={styles.mcpBtn}>
          <img src="/img/logos/cursor.svg" alt="" width={18} height={18} />
          Add to Cursor
          <ArrowRightIcon />
        </a>
        <a href={VSCODE_URL} className={styles.mcpBtn}>
          <img src="/img/logos/vscode.svg" alt="" width={18} height={18} />
          Add to VS Code
          <ArrowRightIcon />
        </a>
        <button
          type="button"
          className={styles.mcpBtn}
          onClick={() => handleCopy('claude', CLAUDE_COMMAND)}>
          <img src="/img/logos/claude.svg" alt="" width={18} height={18} />
          {copied === 'claude' ? 'Command copied' : 'Add to Claude Code'}
          {copied === 'claude' ? <CheckIcon /> : <CopyIcon />}
        </button>
        <button
          type="button"
          className={styles.mcpBtn}
          onClick={() => handleCopy('url', MCP_URL)}>
          <LinkIcon />
          {copied === 'url' ? 'URL copied' : 'Copy MCP URL'}
          {copied === 'url' ? <CheckIcon /> : <CopyIcon />}
        </button>
      </div>
    </div>
  );
}

export default function QuickStart(): React.ReactElement {
  const [active, setActive] = useState(0);
  const step = STEPS[active];

  return (
    <section className={styles.section} aria-labelledby="quickstart-heading">
      <div className="container">
        <div className={styles.split}>
          {/* Left column: heading + blurb */}
          <div className={styles.intro}>
            <span className="sectionEyebrow">Quick start</span>
            <Heading as="h2" id="quickstart-heading" className={styles.introTitle}>
              Up and running in three steps
            </Heading>
            <p className={styles.introText}>
              Install onto any Kubernetes {vars.k8sMinVersion}+ cluster with Helm. Need a cluster or
              more detail? <Link to="/docs/getting-started/">Read the full getting-started guide →</Link>
            </p>
            <p className={styles.note}>
              These commands install the latest release, {AIGW}. For main-branch builds or older
              versions, see the <Link to="/docs/getting-started/installation">installation guide</Link>.
            </p>
          </div>

          {/* Right column: tabbed install panel */}
          <div className={styles.panel}>
            <div className={styles.tabs} role="tablist" aria-label="Installation steps">
              {STEPS.map((s, i) => (
                <button
                  key={s.n}
                  type="button"
                  role="tab"
                  aria-selected={i === active}
                  className={styles.tab}
                  data-active={i === active}
                  onClick={() => setActive(i)}>
                  <span className={styles.tabNum} aria-hidden="true">{s.n}</span>
                  <span className={styles.tabLabel}>{s.tab}</span>
                </button>
              ))}
            </div>

            <div className={styles.panelBody} role="tabpanel">
              <div className={styles.panelMeta}>
                <Heading as="h3" className={styles.stepTitle}>{step.title}</Heading>
                <p className={styles.stepDesc}>{step.desc}</p>
              </div>
              <CodeBlock language="bash" className={styles.code}>
                {step.code}
              </CodeBlock>
            </div>

            <div className={styles.panelFoot}>
              <span className={styles.prereqLabel}>Prerequisites</span>
              <span className={styles.chip}>Kubernetes {vars.k8sMinVersion}+</span>
              <span className={styles.chip}>kubectl</span>
              <span className={styles.chip}>helm</span>
            </div>
          </div>
        </div>

        <McpStrip />
      </div>
    </section>
  );
}
