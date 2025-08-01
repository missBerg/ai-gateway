import React from 'react';
import clsx from 'clsx';
import Heading from '@theme/Heading';
import Link from '@docusaurus/Link';
import { sortedAdopters, type Adopter } from '@site/src/data/adopters';
import styles from './styles.module.css';

function AdopterLogo({ name, logoUrl, url, description }: Adopter) {
  const content = (
    <div className={styles.adopterCard}>
      <div className={styles.adopterName}>{name}</div>
      <div className={styles.logoContainer}>
        <img
          src={logoUrl}
          alt={`${name} logo`}
          className={styles.adopterLogo}
          loading="lazy"
          onError={(e) => {
            const target = e.target as HTMLImageElement;
            target.src = '/img/adopters/placeholder-company.svg';
          }}
        />
      </div>
    </div>
  );

  if (url) {
    return (
      <div className={styles.adopterCol}>
        <a
          href={url}
          target="_blank"
          rel="noopener noreferrer"
          className={styles.adopterLink}
        >
          {content}
        </a>
      </div>
    );
  }

  return (
    <div className={styles.adopterCol}>
      {content}
    </div>
  );
}

export default function Adopters(): React.ReactElement {
  return (
    <section className={styles.adoptersSection}>
      <div className="container">
        <div className={styles.sectionHeader}>
          <Heading as="h2" className={styles.sectionTitle}>
            Adopters
          </Heading>
          <div className={styles.titleUnderline}></div>
          <p className={styles.sectionDescription}>
            See who's using Envoy AI Gateway.
            <br />
          </p>
        </div>
        <div className={styles.adoptersGrid}>
          {sortedAdopters.map((adopter, idx) => (
            <AdopterLogo key={idx} {...adopter} />
          ))}
        </div>
        <div className={styles.ctaSection}>
          <p className={styles.ctaText}>
            Using Envoy AI Gateway? We'd love to feature your organization!
          </p>
          <Link
            className="button button--primary button--lg"
            to="/support#add-your-logo">
            Add Your Logo
          </Link>
        </div>
      </div>
    </section>
  );
}
