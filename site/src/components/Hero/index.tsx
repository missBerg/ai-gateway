import React, {useEffect, useState} from 'react';
import Link from '@docusaurus/Link';
import Heading from '@theme/Heading';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import versions from '@site/versions.json';
import styles from './styles.module.css';

const LATEST_VERSION = versions[0];

function TerminalIcon(): React.ReactElement {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <polyline points="4 17 10 11 4 5" />
      <line x1="12" y1="19" x2="20" y2="19" />
    </svg>
  );
}

function ArrowRightIcon(): React.ReactElement {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <line x1="5" y1="12" x2="19" y2="12" />
      <polyline points="12 5 19 12 12 19" />
    </svg>
  );
}

function GithubIcon(): React.ReactElement {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
      <path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.4 3-.405 1.02.005 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12" />
    </svg>
  );
}

/** Compact "12.3k" style formatter. */
function fmtCount(n: number): string {
  if (n >= 1000) return (n / 1000).toFixed(n >= 10000 ? 0 : 1).replace(/\.0$/, '') + 'k';
  return String(n);
}

/** Live GitHub star count, gracefully degrading to a static fallback. */
function useGitHubStars(): string | null {
  const [stars, setStars] = useState<string | null>(null);
  useEffect(() => {
    let cancelled = false;
    fetch('https://api.github.com/repos/envoyproxy/ai-gateway')
      .then((r) => (r.ok ? r.json() : null))
      .then((d) => {
        if (!cancelled && d && typeof d.stargazers_count === 'number') {
          setStars(fmtCount(d.stargazers_count));
        }
      })
      .catch(() => {/* keep null → row hides the star stat */});
    return () => {
      cancelled = true;
    };
  }, []);
  return stars;
}

export default function Hero(): React.ReactElement {
  const {siteConfig} = useDocusaurusContext();
  const stars = useGitHubStars();

  return (
    <header className={styles.hero}>
      <div className={styles.bg} aria-hidden="true">
        <div className={styles.grid} />
        <div className={styles.hexagonGlow} />
        <div className={styles.hexagon} />
        <div className={styles.glow} />
      </div>

      <div className={styles.inner}>
        {/* Left: large brand lockup — hexagon + "envoy AI gateway" wordmark */}
        <img
          className={styles.logo}
          src="/img/ai-gw-logo.svg"
          alt="Envoy AI Gateway"
          width={900}
          height={506}
          loading="eager"
        />

        {/* Right: headline, subhead, CTAs, factual stats */}
        <div className={styles.content}>
        <Heading as="h1" className={styles.title}>
          One gateway for <span className={styles.titleAccent}>all your LLM traffic</span>
        </Heading>

        <p className={styles.subtitle}>{siteConfig.tagline}</p>

        <div className={styles.ctas}>
          <Link className="btn btn--accent btn--lg" to="#mcp-install">
            <TerminalIcon />
            Install Docs MCP
          </Link>
          <Link className="btn btn--primary btn--lg" to="/docs/getting-started/">
            Get started
            <ArrowRightIcon />
          </Link>
          <Link
            className={styles.githubCta}
            href="https://github.com/envoyproxy/ai-gateway">
            <GithubIcon />
            <span>Star on GitHub</span>
            {stars && <span className={styles.starCount}>{stars}</span>}
          </Link>
        </div>

        {/* Concrete, factual stats — numbers beat adjectives for credibility */}
        <ul className={styles.stats}>
          <li className={styles.stat}>
            <span className={styles.statValue}>20+</span>
            <span className={styles.statLabel}>LLM providers</span>
          </li>
          <li className={styles.statSep} aria-hidden="true" />
          <li className={styles.stat}>
            <span className={styles.statValue}>Apache 2.0</span>
            <span className={styles.statLabel}>fully open source</span>
          </li>
          <li className={styles.statSep} aria-hidden="true" />
          <li className={styles.stat}>
            <span className={styles.statValue}>OpenAI-compatible</span>
            <span className={styles.statLabel}>drop-in API</span>
          </li>
          <li className={styles.statSep} aria-hidden="true" />
          <li className={styles.stat}>
            <Link className={styles.statVersion} to={`/release-notes/v${LATEST_VERSION}`}>
              <span className={styles.statDot} aria-hidden="true" />
              v{LATEST_VERSION}
            </Link>
            <span className={styles.statLabel}>latest release</span>
          </li>
        </ul>
        </div>
      </div>
    </header>
  );
}
