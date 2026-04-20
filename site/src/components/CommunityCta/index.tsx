import React from 'react';
import styles from './styles.module.css';

export default function CommunityCta(): React.ReactElement {
  return (
    <section className={styles.section}>
      <div className="container">
        <div className={`${styles.card} sectionOverlay--brand`}>
          <div className={styles.copy}>
            <h2 className={styles.title}>Build with the community.</h2>
            <p className={styles.subtitle}>
              Envoy AI Gateway is built in the open by the CNCF community and
              used in production across cloud providers, startups, and
              enterprises. Jump in on Slack or GitHub.
            </p>
          </div>
          <div className={styles.actions}>
            <a
              className={styles.slackBtn}
              href="https://envoyproxy.slack.com/archives/C07Q4N24VAA"
              target="_blank"
              rel="noreferrer"
            >
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                <rect x="9" y="2" width="6" height="9" rx="3" />
                <path d="M15 2v2a3 3 0 0 0 3 3h2" />
                <rect x="13" y="13" width="9" height="6" rx="3" />
                <path d="M22 15h-2a3 3 0 0 0-3 3v2" />
                <rect x="9" y="13" width="6" height="9" rx="3" transform="rotate(180 12 17.5)" />
                <path d="M9 22v-2a3 3 0 0 0-3-3H4" />
                <rect x="2" y="9" width="9" height="6" rx="3" transform="rotate(180 6.5 12)" />
                <path d="M2 9h2a3 3 0 0 0 3-3V4" />
              </svg>
              Join us on Slack
            </a>
            <a
              className={styles.githubBtn}
              href="https://github.com/envoyproxy/ai-gateway"
              target="_blank"
              rel="noreferrer"
            >
              <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                <path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.4 3-.405 1.02.005 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12" />
              </svg>
              Star on GitHub
            </a>
          </div>
        </div>
      </div>
    </section>
  );
}
