import React from 'react';
import Link from '@docusaurus/Link';
import Heading from '@theme/Heading';
import versions from '@site/versions.json';
import styles from './styles.module.css';

const LATEST_VERSION: string = versions[0] ?? '0.5';

type LLMProvider = {
  name: string;
  logoUrl: string;
  docsPath: string;
  addedIn?: string;
};

const SUPPORTED_PROVIDERS_DOC = '/docs/capabilities/llm-integrations/supported-providers';
const CONNECT_PROVIDERS_ROOT = '/docs/getting-started/connect-providers';

const PROVIDERS: LLMProvider[] = [
  {
    name: 'OpenAI',
    logoUrl: '/img/providers/openai.svg',
    docsPath: `${CONNECT_PROVIDERS_ROOT}/openai`,
  },
  {
    name: 'Anthropic',
    logoUrl: '/img/providers/anthropic.svg',
    docsPath: `${CONNECT_PROVIDERS_ROOT}/anthropic`,
  },
  {
    name: 'AWS Bedrock',
    logoUrl: '/img/providers/aws-bedrock.svg',
    docsPath: `${CONNECT_PROVIDERS_ROOT}/aws-bedrock`,
  },
  {
    name: 'Azure OpenAI',
    logoUrl: '/img/providers/azure-openai.svg',
    docsPath: `${CONNECT_PROVIDERS_ROOT}/azure-openai`,
  },
  {
    name: 'Google Gemini',
    logoUrl: '/img/providers/google-gemini.svg',
    docsPath: SUPPORTED_PROVIDERS_DOC,
  },
  {
    name: 'Vertex AI',
    logoUrl: '/img/providers/vertex-ai.svg',
    docsPath: `${CONNECT_PROVIDERS_ROOT}/gcp-vertexai`,
  },
  {
    name: 'Mistral',
    logoUrl: '/img/providers/mistral.svg',
    docsPath: SUPPORTED_PROVIDERS_DOC,
  },
  {
    name: 'Cohere',
    logoUrl: '/img/providers/cohere.svg',
    docsPath: SUPPORTED_PROVIDERS_DOC,
  },
  {
    name: 'Grok',
    logoUrl: '/img/providers/grok.svg',
    docsPath: SUPPORTED_PROVIDERS_DOC,
  },
  {
    name: 'Groq',
    logoUrl: '/img/providers/groq.svg',
    docsPath: SUPPORTED_PROVIDERS_DOC,
  },
  {
    name: 'Together AI',
    logoUrl: '/img/providers/together-ai.svg',
    docsPath: SUPPORTED_PROVIDERS_DOC,
  },
  {
    name: 'DeepInfra',
    logoUrl: '/img/providers/deepinfra.svg',
    docsPath: SUPPORTED_PROVIDERS_DOC,
  },
  {
    name: 'DeepSeek',
    logoUrl: '/img/providers/deepseek.svg',
    docsPath: SUPPORTED_PROVIDERS_DOC,
  },
  {
    name: 'Hunyuan',
    logoUrl: '/img/providers/hunyuan.svg',
    docsPath: SUPPORTED_PROVIDERS_DOC,
  },
  {
    name: 'SambaNova',
    logoUrl: '/img/providers/sambanova.svg',
    docsPath: SUPPORTED_PROVIDERS_DOC,
  },
  {
    name: 'Tetrate Agent Router Service',
    logoUrl: '/img/providers/tars.svg',
    docsPath: `${CONNECT_PROVIDERS_ROOT}/tars`,
  },
];

function NewBadge() {
  return <span className={styles.newBadge}>New</span>;
}

function ProviderCard({name, logoUrl, docsPath, addedIn}: LLMProvider) {
  const isNew = addedIn === LATEST_VERSION;
  return (
    <Link to={docsPath} className={styles.card} aria-label={`${name} integration`}>
      {isNew && <NewBadge />}
      <div className={styles.logoContainer}>
        <img
          src={logoUrl}
          alt={`${name} logo`}
          className={styles.logo}
          loading="lazy"
          onError={(e) => {
            const target = e.target as HTMLImageElement;
            target.src = '/img/providers/placeholder.svg';
          }}
        />
      </div>
      <div className={styles.name}>{name}</div>
    </Link>
  );
}

export default function LLMProviders(): React.ReactElement {
  const providers = [...PROVIDERS].sort((a, b) => a.name.localeCompare(b.name));

  return (
    <section id="llm-providers" className={styles.section}>
      <div className="container">
        <div className="sectionHeader">
          <span className="sectionEyebrow sectionEyebrow--purple">LLM Providers</span>
          <Heading as="h2" className="sectionTitle">
            One gateway, every major LLM
          </Heading>
          <p className="sectionSubtitle">
            Route traffic to any of the providers below through a single
            OpenAI-compatible API. Click any provider for configuration details.
          </p>
        </div>

        <div className={styles.grid}>
          {providers.map((p) => (
            <ProviderCard key={p.name} {...p} />
          ))}
        </div>

        <div className={styles.footer}>
          <Link to={SUPPORTED_PROVIDERS_DOC} className={styles.footerLink}>
            See the full compatibility matrix →
          </Link>
        </div>
      </div>
    </section>
  );
}
