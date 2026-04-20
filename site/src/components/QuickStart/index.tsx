import React, {useState, useCallback} from 'react';
import Link from '@docusaurus/Link';
import styles from './styles.module.css';

// Fast path — one curl|bash line (interactive, prompts for version).
// The script itself lives at site/static/scripts/install.sh and handles
// prereq checks, prompts, and the three helm releases.
const INSTALL_URL =
  'https://raw.githubusercontent.com/envoyproxy/ai-gateway/main/site/static/scripts/install.sh';

const FAST_SCRIPT = `# Interactive — prompts for AI Gateway version
curl -sSL ${INSTALL_URL} | bash

# Non-interactive — pin a version, skip prompts
curl -sSL ${INSTALL_URL} | bash -s -- --version v0.5.0 --yes

# Audit first (recommended for production)
curl -sSL ${INSTALL_URL} -o install.sh
less install.sh && bash install.sh`;

// Manual path — mirrors https://aigateway.envoyproxy.io/docs/getting-started/installation/
const MANUAL_SCRIPT = `# 1. Install Envoy Gateway (with AI Gateway values) + wait for ready
helm upgrade -i eg oci://docker.io/envoyproxy/gateway-helm \\
  --version v0.0.0-latest \\
  --namespace envoy-gateway-system --create-namespace \\
  -f https://raw.githubusercontent.com/envoyproxy/ai-gateway/main/manifests/envoy-gateway-values.yaml

kubectl wait --timeout=2m -n envoy-gateway-system \\
  deployment/envoy-gateway --for=condition=Available

# 2. Install AI Gateway CRDs
helm upgrade -i aieg-crd oci://docker.io/envoyproxy/ai-gateway-crds-helm \\
  --version v0.5.0 \\
  --namespace envoy-ai-gateway-system --create-namespace

# 3. Install the AI Gateway controller + wait for ready
helm upgrade -i aieg oci://docker.io/envoyproxy/ai-gateway-helm \\
  --version v0.5.0 \\
  --namespace envoy-ai-gateway-system --create-namespace

kubectl wait --timeout=2m -n envoy-ai-gateway-system \\
  deployment/ai-gateway-controller --for=condition=Available`;

type Mode = 'fast' | 'manual';

function CopyIcon() {
  return (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
      <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
    </svg>
  );
}

function CheckIcon() {
  return (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.4" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <polyline points="20 6 9 17 4 12" />
    </svg>
  );
}

const PREREQS = [
  {label: 'Kubernetes 1.32+'},
  {label: 'kubectl'},
  {label: 'helm'},
];

const MODE_COPY: Record<Mode, {title: string; subtitle: string}> = {
  fast: {
    title: 'Install with one command.',
    subtitle:
      'Pipe the installer into bash on a clean Kubernetes 1.32+ cluster. It prompts for the AI Gateway version, runs preflight checks, then installs Envoy Gateway, the CRDs, and the controller.',
  },
  manual: {
    title: 'Install with Helm.',
    subtitle:
      'Three commands get Envoy Gateway and AI Gateway running on any Kubernetes 1.32+ cluster. The full guide covers provider connections and your first route.',
  },
};

function renderScript(script: string, commentClass: string): React.ReactNode {
  return script.split('\n').map((line, i) => {
    const isComment = line.trimStart().startsWith('#');
    return (
      <span key={i} className={isComment ? commentClass : undefined}>
        {line}
        {'\n'}
      </span>
    );
  });
}

export default function QuickStart(): React.ReactElement {
  const [mode, setMode] = useState<Mode>('fast');
  const [copied, setCopied] = useState(false);

  const script = mode === 'fast' ? FAST_SCRIPT : MANUAL_SCRIPT;
  const copyText = mode === 'fast' ? `curl -sSL ${INSTALL_URL} | bash` : MANUAL_SCRIPT;
  const {title, subtitle} = MODE_COPY[mode];

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(copyText);
      setCopied(true);
      setTimeout(() => setCopied(false), 1800);
    } catch {
      /* noop */
    }
  }, [copyText]);

  return (
    <section className={styles.section}>
      <div className="container">
        <div className={styles.wrap}>
          <div className={styles.left}>
            <span className="sectionEyebrow sectionEyebrow--purple">Quick start</span>
            <h2 className={styles.title}>{title}</h2>
            <p className={styles.subtitle}>{subtitle}</p>

            <div className={styles.prereqs}>
              <span className={styles.prereqLabel}>You&apos;ll need</span>
              <ul className={styles.prereqList}>
                {PREREQS.map((p) => (
                  <li key={p.label} className={styles.prereqChip}>
                    {p.label}
                  </li>
                ))}
              </ul>
            </div>

            <Link to="/docs/getting-started/" className={styles.cta}>
              Read the full guide →
            </Link>
          </div>

          <div className={styles.codeWrap}>
            <div className={styles.codeTopbar}>
              <span className={styles.dots}>
                <i />
                <i />
                <i />
              </span>

              <div className={styles.tabs} role="tablist" aria-label="Install method">
                <button
                  type="button"
                  role="tab"
                  aria-selected={mode === 'fast'}
                  className={`${styles.tab} ${mode === 'fast' ? styles.tabActive : ''}`}
                  onClick={() => setMode('fast')}
                >
                  Fast
                </button>
                <button
                  type="button"
                  role="tab"
                  aria-selected={mode === 'manual'}
                  className={`${styles.tab} ${mode === 'manual' ? styles.tabActive : ''}`}
                  onClick={() => setMode('manual')}
                >
                  Manual
                </button>
              </div>

              <button
                type="button"
                className={styles.copyBtn}
                onClick={handleCopy}
                aria-label={copied ? 'Copied' : 'Copy install command'}
              >
                {copied ? <CheckIcon /> : <CopyIcon />}
                <span>{copied ? 'Copied' : 'Copy'}</span>
              </button>
            </div>
            <pre className={styles.pre}>
              <code className={styles.code}>
                {renderScript(script, styles.commentLine)}
              </code>
            </pre>
          </div>
        </div>
      </div>
    </section>
  );
}
