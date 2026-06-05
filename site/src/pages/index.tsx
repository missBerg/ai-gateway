import React from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import Hero from '@site/src/components/Hero';
import Reveal from '@site/src/components/Reveal';
import McpInstall from '@site/src/components/McpInstall';
import Capabilities from '@site/src/components/Capabilities';
import QuickStart from '@site/src/components/QuickStart';
import LLMProviders from '@site/src/components/LLMProviders';
import LatestBlogs from '@site/src/components/LatestBlogs';
import Adopters from '@site/src/components/Adopters';
import CommunityCta from '@site/src/components/CommunityCta';

export default function Home(): React.ReactElement {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout title={siteConfig.title} description={siteConfig.tagline}>
      <Hero />
      <main>
        {/* Each section floats up as it scrolls into view (reduced-motion safe). */}
        {/* MCP install first — the fastest on-ramp — then lead with value. */}
        <Reveal><McpInstall /></Reveal>
        <Reveal><Capabilities /></Reveal>
        <Reveal><LLMProviders /></Reveal>
        <Reveal><QuickStart /></Reveal>
        <Reveal><LatestBlogs /></Reveal>
        <Reveal><Adopters /></Reveal>
        <Reveal><CommunityCta /></Reveal>
      </main>
    </Layout>
  );
}
