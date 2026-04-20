import React from 'react';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import Capabilities from '@site/src/components/Capabilities';
import LLMProviders from '@site/src/components/LLMProviders';
import Adopters from '@site/src/components/Adopters';
import LatestBlogs from '@site/src/components/LatestBlogs';
import McpInstall from '@site/src/components/McpInstall';
import QuickStart from '@site/src/components/QuickStart';
import CommunityCta from '@site/src/components/CommunityCta';
import versions from '@site/versions.json';

const LATEST_VERSION: string = versions[0] ?? '0.5';

function GithubIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
      <path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.4 3-.405 1.02.005 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12" />
    </svg>
  );
}

function TerminalIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <polyline points="4 17 10 11 4 5" />
      <line x1="12" y1="19" x2="20" y2="19" />
    </svg>
  );
}

function ArrowRightIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <line x1="5" y1="12" x2="19" y2="12" />
      <polyline points="12 5 19 12 12 19" />
    </svg>
  );
}

function HomepageHeader() {
  return (
    <header className="heroBanner">
      <div className="container">
        <img className="heroImage" src="/img/ai-gw-logo.svg" alt="Envoy AI Gateway" />
        <h1 className="heroTitle">The open-source AI Gateway for Envoy.</h1>
        <p className="heroSubtitle">
          Route, govern, and observe traffic to any LLM — from one unified
          control plane, built on Envoy Proxy and Envoy Gateway.
        </p>

        <div className="heroCtas">
          <Link className="btn btn--accent btn--lg" href="#mcp-install">
            <TerminalIcon />
            Bring the docs to your coding agent
          </Link>
          <Link className="btn btn--primary btn--lg" href="/docs/getting-started/">
            Get started
            <ArrowRightIcon />
          </Link>
          <Link
            className="btn btn--ghost btn--lg"
            href="https://github.com/envoyproxy/ai-gateway"
          >
            <GithubIcon />
            View on GitHub
          </Link>
        </div>

        <Link className="heroVersionPill" to={`/release-notes/v${LATEST_VERSION}`}>
          v{LATEST_VERSION} · See release notes →
        </Link>
      </div>
    </header>
  );
}

export default function Home(): React.ReactElement {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout title={siteConfig.title} description={siteConfig.tagline}>
      <HomepageHeader />
      <main>
        <McpInstall />
        <Capabilities />
        <QuickStart />
        <LLMProviders />
        <LatestBlogs />
        <Adopters />
        <CommunityCta />
      </main>
    </Layout>
  );
}
