import React from 'react';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

/* The signature visual: many clients consolidate onto one OpenAI-compatible
   endpoint; the gateway adds routing / token governance / observability, then
   fans out to any provider. A symmetric fan-in → gateway → fan-out tells the
   whole product story in one glance.

   Desktop renders a single responsive SVG so connectors always line up with
   nodes; below 768px we swap to a clean stacked HTML layout (see styles). */

type Node = {key: string; name: string; logo?: string; icon?: string};

const CLIENTS: Node[] = [
  {key: 'app', name: 'Your app', icon: 'M9 8l-3 4 3 4 M15 8l3 4-3 4'},
  {key: 'agents', name: 'AI agents', icon: 'M12 3l1.7 3.9L18 8.5l-3.6 1.4L12 14l-2.4-4.1L6 8.5l4.3-1.6z'},
  {key: 'services', name: 'Services', icon: 'M4 6h16v4.5H4z M4 13.5h16V18H4z'},
];

const PROVIDERS: Node[] = [
  {key: 'openai', name: 'OpenAI', logo: '/img/providers/openai.svg'},
  {key: 'anthropic', name: 'Anthropic', logo: '/img/providers/anthropic.svg'},
  {key: 'bedrock', name: 'AWS Bedrock', logo: '/img/providers/aws-bedrock.svg'},
  {key: 'gemini', name: 'Google Gemini', logo: '/img/providers/google-gemini.svg'},
  {key: 'mistral', name: 'Mistral', logo: '/img/providers/mistral.svg'},
];

// Gateway capability pills — 24x24 icon paths drawn inline.
const CAPS = [
  {label: 'Smart routing', icon: 'M6 3v6a3 3 0 0 0 3 3h9 M18 9l3 3-3 3 M6 21v-6'},
  {label: 'Token governance', icon: 'M12 3l8 4v5c0 4.5-3.2 7.3-8 9-4.8-1.7-8-4.5-8-9V7z'},
  {label: 'Observability', icon: 'M3 3v18h18 M7 14l4-4 4 4 5-6'},
];

// --- SVG geometry (viewBox 0 0 1120 440) ---
const GW_IN_X = 430;
const GW_OUT_X = 690;
const GW_MID_Y = 220;
// Distinct entry points on the gateway's left edge so each client's line stays
// visually separate (otherwise the curves collapse onto the middle line).
const GW_IN_YS = [170, 220, 270];

const CL_X = 44;
const CL_W = 188;
const CL_H = 46;
const clTop = (i: number) => 119 + i * 78; // 3 clients centred on the gateway
const clMid = (i: number) => clTop(i) + CL_H / 2;

const CHIP_X = 890;
const CHIP_W = 196;
const CHIP_H = 46;
const chipTop = (i: number) => 34 + i * 64; // 6 rows (5 providers + "more")
const chipMid = (i: number) => chipTop(i) + CHIP_H / 2;

export default function HowItWorks(): React.ReactElement {
  return (
    <section className={styles.section} aria-labelledby="how-it-works-heading">
      <div className="container">
        <div className="sectionHeader">
          <span className="sectionEyebrow">How it works</span>
          <Heading as="h2" id="how-it-works-heading" className="sectionTitle">
            One API in. Every provider out.
          </Heading>
          <p className="sectionSubtitle">
            Point every app, agent, and service at one OpenAI-compatible endpoint. The gateway
            handles routing, token-aware governance, and observability — then forwards to
            whichever provider you choose, no SDK changes required.
          </p>
        </div>

        {/* ---------- Desktop: single responsive SVG ---------- */}
        <svg
          className={styles.diagram}
          viewBox="0 0 1120 440"
          role="img"
          aria-label="Apps, agents, and services send requests to a single OpenAI-compatible endpoint. Envoy AI Gateway applies routing, token governance, and observability, then fans out to providers including OpenAI, Anthropic, AWS Bedrock, Google Gemini, and Mistral.">
          <defs>
            {/* userSpaceOnUse (not the objectBoundingBox default): a perfectly
                horizontal path has a zero-height bbox, which degenerates the
                gradient and renders the stroke invisible — exactly what happened
                to the middle client line. Explicit coords also give all lines one
                coherent left→right color sweep. */}
            <linearGradient id="hiwFlow" x1={CL_X + CL_W} y1="0" x2={CHIP_X} y2="0" gradientUnits="userSpaceOnUse">
              <stop offset="0%" className={styles.flowStop0} />
              <stop offset="100%" className={styles.flowStop1} />
            </linearGradient>
            <linearGradient id="hiwPanel" x1="0" y1="0" x2="1" y2="1">
              <stop offset="0%" className={styles.panelStop0} />
              <stop offset="100%" className={styles.panelStop1} />
            </linearGradient>
            <linearGradient id="hiwBorder" x1="0" y1="0" x2="1" y2="1">
              <stop offset="0%" className={styles.borderStop0} />
              <stop offset="55%" className={styles.borderStop1} />
              <stop offset="100%" className={styles.borderStop2} />
            </linearGradient>
            <radialGradient id="hiwGlow" cx="50%" cy="50%" r="50%">
              <stop offset="0%" className={styles.glowStop0} />
              <stop offset="100%" className={styles.glowStop1} />
            </radialGradient>
          </defs>

          {/* glow behind the gateway */}
          <ellipse cx="560" cy="220" rx="250" ry="205" fill="url(#hiwGlow)" />

          {/* ---- connectors (behind nodes) ---- */}
          {/* clients -> gateway (fan in) — distinct entry points so each line reads */}
          {CLIENTS.map((c, i) => (
            <path
              key={c.key}
              className={styles.flow}
              style={{animationDelay: `${i * -0.45}s`}}
              d={`M${CL_X + CL_W} ${clMid(i)} C 330 ${clMid(i)} 366 ${GW_IN_YS[i]} ${GW_IN_X} ${GW_IN_YS[i]}`}
            />
          ))}
          {/* gateway -> providers + "more" (fan out) */}
          {[...PROVIDERS, {key: 'more'}].map((p, i) => (
            <path
              key={p.key}
              className={styles.flow}
              style={{animationDelay: `${i * -0.4}s`}}
              d={`M${GW_OUT_X} ${GW_MID_Y} C 800 ${GW_MID_Y} 790 ${chipMid(i)} ${CHIP_X} ${chipMid(i)}`}
            />
          ))}

          {/* ---- client chips ---- */}
          {CLIENTS.map((c, i) => (
            <g key={c.key}>
              <rect className={styles.clientChip} x={CL_X} y={clTop(i)} width={CL_W} height={CL_H} rx="12" />
              <svg x={CL_X + 14} y={clTop(i) + 11} width="24" height="24" viewBox="0 0 24 24">
                <path className={styles.clientIcon} d={c.icon} fill="none" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round" />
              </svg>
              <text className={styles.clientText} x={CL_X + 48} y={clTop(i) + 29}>{c.name}</text>
            </g>
          ))}

          {/* ---- gateway (always-dark centerpiece) ---- */}
          <g>
            <rect x="430" y="64" width="260" height="312" rx="22" fill="url(#hiwPanel)" />
            <rect x="430.5" y="64.5" width="259" height="311" rx="21.5" fill="none" stroke="url(#hiwBorder)" strokeWidth="1.5" />
            <image href="/img/logo-white.svg" x="526" y="90" width="68" height="68" />
            <text className={styles.gwTitle} x="560" y="186" textAnchor="middle">Envoy AI Gateway</text>

            {CAPS.map((c, i) => {
              const y = 206 + i * 46;
              return (
                <g key={c.label}>
                  <rect className={styles.pill} x="452" y={y} width="216" height="34" rx="17" />
                  <svg x="466" y={y + 9} width="16" height="16" viewBox="0 0 24 24">
                    <path className={styles.pillIcon} d={c.icon} fill="none" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                  </svg>
                  <text className={styles.pillText} x="492" y={y + 22}>{c.label}</text>
                </g>
              );
            })}
          </g>

          {/* ---- provider chips ---- */}
          {PROVIDERS.map((p, i) => (
            <g key={p.key}>
              <rect className={styles.chip} x={CHIP_X} y={chipTop(i)} width={CHIP_W} height={CHIP_H} rx="12" />
              <image href={p.logo} x={CHIP_X + 14} y={chipTop(i) + 11} width="24" height="24" />
              <text className={styles.chipText} x={CHIP_X + 50} y={chipTop(i) + 29}>{p.name}</text>
            </g>
          ))}
          {/* "+ more" chip */}
          <g>
            <rect className={styles.moreChip} x={CHIP_X} y={chipTop(5)} width={CHIP_W} height={CHIP_H} rx="12" />
            <text className={styles.moreText} x={CHIP_X + CHIP_W / 2} y={chipTop(5) + 29} textAnchor="middle">
              + 15 more providers
            </text>
          </g>
        </svg>

        {/* ---------- Mobile: clean stacked layout ---------- */}
        <div className={styles.mobile} aria-hidden="true">
          <div className={styles.mClients}>
            {CLIENTS.map((c) => (
              <span key={c.key} className={styles.mClientChip}>
                <svg width="16" height="16" viewBox="0 0 24 24">
                  <path className={styles.clientIcon} d={c.icon} fill="none" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round" />
                </svg>
                {c.name}
              </span>
            ))}
          </div>
          <span className={styles.mArrow} />
          <div className={styles.mGateway}>
            <img src="/img/logo-white.svg" alt="" width={40} height={40} />
            <strong>Envoy AI Gateway</strong>
            <ul>
              {CAPS.map((c) => (
                <li key={c.label}>{c.label}</li>
              ))}
            </ul>
          </div>
          <span className={styles.mArrow} />
          <div className={styles.mProviders}>
            {PROVIDERS.map((p) => (
              <span key={p.key} className={styles.mChip}>
                <img src={p.logo} alt="" width={20} height={20} />
                {p.name}
              </span>
            ))}
            <span className={`${styles.mChip} ${styles.mChipMore}`}>+ 15 more</span>
          </div>
        </div>
      </div>
    </section>
  );
}
