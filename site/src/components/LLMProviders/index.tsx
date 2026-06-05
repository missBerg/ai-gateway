import React from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import Heading from '@theme/Heading';
import {LANDING_PROVIDERS, type Provider} from '@site/src/data/providers';
import styles from './styles.module.css';

const PLACEHOLDER = '/img/providers/placeholder.svg';

function ProviderTile({provider}: {provider: Provider}): React.ReactElement {
  const label = provider.gridLabel ?? provider.name;
  const comingSoon = provider.status !== 'supported';
  return (
    <div className={clsx(styles.tile, comingSoon && styles.tileMuted)}>
      <div className={styles.logoWrap}>
        <img
          src={provider.logoUrl ?? PLACEHOLDER}
          alt={`${label} logo`}
          className={styles.logo}
          width={44}
          height={44}
          loading="lazy"
          onError={(e) => {
            (e.currentTarget as HTMLImageElement).src = PLACEHOLDER;
          }}
        />
      </div>
      <div className={styles.name}>{label}</div>
      {comingSoon && <span className={styles.badge}>Coming soon</span>}
    </div>
  );
}

export default function LLMProviders(): React.ReactElement {
  return (
    <section
      id="llm-providers"
      className={clsx('sectionWrap', styles.section)}
      aria-labelledby="providers-heading">
      <div className="container">
        <div className="sectionHeader">
          <span className="sectionEyebrow sectionEyebrow--purple">Providers</span>
          <Heading as="h2" id="providers-heading" className="sectionTitle">
            One API, every major LLM provider
          </Heading>
          <p className="sectionSubtitle">
            Route to any of these out of the box — no SDK changes, just a model name.{' '}
            <Link to="/docs/capabilities/llm-integrations/supported-providers">
              See the full compatibility matrix →
            </Link>
          </p>
        </div>
        <div className={styles.grid}>
          {LANDING_PROVIDERS.map((provider) => (
            <ProviderTile key={provider.name} provider={provider} />
          ))}
        </div>
      </div>
    </section>
  );
}
