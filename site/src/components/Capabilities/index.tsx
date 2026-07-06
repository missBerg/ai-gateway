import React from 'react';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

type Capability = {
  title: string;
  description: string;
  icon: React.ReactElement;
};

/* Lucide icons (route, gauge, activity) — inlined so each renders twice per
   card: small beside the title, and oversized bleeding out of the top-right
   corner as a thin-line background texture. */
const CAPABILITIES: Capability[] = [
  {
    title: 'Any LLM, one API',
    description:
      'Route to OpenAI, Anthropic, AWS Bedrock, Azure OpenAI, Google Gemini, Mistral, Cohere, and more — all behind a single OpenAI-compatible interface.',
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
        <circle cx="6" cy="19" r="3" />
        <path d="M9 19h8.5a3.5 3.5 0 0 0 0-7h-11a3.5 3.5 0 0 1 0-7H15" />
        <circle cx="18" cy="5" r="3" />
      </svg>
    ),
  },
  {
    title: 'Token-aware governance',
    description:
      'Enforce token-based rate limits, per-team quotas, and cost controls. Keep spend predictable across providers and stop noisy tenants from starving production traffic.',
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
        <path d="m12 14 4-4" />
        <path d="M3.34 19a10 10 0 1 1 17.32 0" />
      </svg>
    ),
  },
  {
    title: 'Production observability',
    description:
      'OpenTelemetry metrics, access logs, and distributed traces for every LLM call — out of the box. Debug request flows and attribute cost without bolting on extra tooling.',
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
        <path d="M22 12h-4l-3 9L9 3l-3 9H2" />
      </svg>
    ),
  },
];

export default function Capabilities(): React.ReactElement {
  return (
    <section className={styles.section} aria-labelledby="capabilities-heading">
      <div className="container">
        <div className="sectionHeader">
          <span className="sectionEyebrow">Capabilities</span>
          <Heading as="h2" id="capabilities-heading" className="sectionTitle">
            Built for production AI traffic
          </Heading>
          <p className="sectionSubtitle">
            Envoy AI Gateway brings the reliability of Envoy to the GenAI stack — routing,
            governance, and observability, purpose-built for LLM workloads.
          </p>
        </div>

        <div className={styles.grid}>
          {CAPABILITIES.map((cap) => (
            <div key={cap.title} className={styles.card}>
              <div className={styles.bgIcon} aria-hidden="true">
                {cap.icon}
              </div>
              <Heading as="h3" className={styles.cardTitle}>
                <span className={styles.titleIcon}>{cap.icon}</span>
                {cap.title}
              </Heading>
              <p className={styles.cardDescription}>{cap.description}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
