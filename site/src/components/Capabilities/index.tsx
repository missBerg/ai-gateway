import React from 'react';
import styles from './styles.module.css';

type Capability = {
  title: string;
  description: string;
  icon: React.ReactElement;
};

const CAPABILITIES: Capability[] = [
  {
    title: 'Any LLM. One API.',
    description:
      'Route traffic to OpenAI, Anthropic, AWS Bedrock, Azure OpenAI, Google Gemini, Mistral, Cohere, and more — all behind a single OpenAI-compatible interface.',
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
        <circle cx="12" cy="12" r="3" />
        <path d="M12 2v4" />
        <path d="M12 18v4" />
        <path d="M4.93 4.93l2.83 2.83" />
        <path d="M16.24 16.24l2.83 2.83" />
        <path d="M2 12h4" />
        <path d="M18 12h4" />
        <path d="M4.93 19.07l2.83-2.83" />
        <path d="M16.24 7.76l2.83-2.83" />
      </svg>
    ),
  },
  {
    title: 'Token-aware governance.',
    description:
      'Enforce token-based rate limits, per-team quotas, and cost controls. Keep spend predictable across providers and prevent noisy tenants from starving production traffic.',
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
        <path d="M12 2a10 10 0 0 1 10 10h-10z" />
        <path d="M12 2v10l-7.07 7.07A10 10 0 0 1 12 2z" />
        <path d="M22 12a10 10 0 0 1-17.07 7.07" />
      </svg>
    ),
  },
  {
    title: 'Production observability.',
    description:
      'OpenTelemetry metrics, access logs, and distributed traces for every LLM call — out of the box. Debug request flows and attribute cost without bolting on extra tooling.',
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
        <path d="M3 3v18h18" />
        <path d="M7 14l4-4 4 4 5-6" />
      </svg>
    ),
  },
];

export default function Capabilities(): React.ReactElement {
  return (
    <section className={styles.section}>
      <div className="container">
        <div className="sectionHeader">
          <span className="sectionEyebrow sectionEyebrow--purple">Capabilities</span>
          <h2 className="sectionTitle">Built for production AI traffic</h2>
          <p className="sectionSubtitle">
            Envoy AI Gateway brings the reliability of Envoy to the GenAI stack —
            routing, governance, and observability, purpose-built for LLM workloads.
          </p>
        </div>

        <div className={styles.grid}>
          {CAPABILITIES.map((cap) => (
            <div key={cap.title} className={styles.card}>
              <div className={styles.iconWrap}>{cap.icon}</div>
              <h3 className={styles.cardTitle}>{cap.title}</h3>
              <p className={styles.cardDescription}>{cap.description}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
