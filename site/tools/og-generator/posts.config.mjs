/* Per-post OG card configs for `npm run regenerate`.
   `slug` doubles as the output filename: site/static/img/blog/og/<slug>.png.
   Icon names are lucide (https://lucide.dev/icons); pick one that matches the
   post's concept, echoing the homepage capability cards (route/gauge/activity). */

export const POSTS = [
  {
    slug: 'introducing-envoy-ai-gateway',
    eyebrow: 'Announcement',
    icon: 'route',
    title: 'Introducing Envoy AI Gateway',
    subtitle: 'Open collaboration to bring AI gateway features to the Envoy community',
  },
  {
    slug: 'kubecon-end-user-keynote-2024',
    eyebrow: 'KubeCon NA 2024',
    icon: 'presentation',
    title: 'End User Keynote: Centralizing Enterprise AI Workflows',
    subtitle: 'How Bloomberg and Tetrate unify model access at scale',
  },
  {
    slug: '01-release-announcement',
    eyebrow: 'Release v0.1',
    icon: 'rocket',
    title: 'The First Envoy AI Gateway Release — A Community Milestone',
  },
  {
    slug: 'envoy-ai-gateway-reference-architecture',
    eyebrow: 'Reference architecture',
    icon: 'layers',
    title: 'A Reference Architecture for Adopters of Envoy AI Gateway',
    visual: 'fade',
    image: '/site-img/blog/aigw-ref.drawio.png',
  },
  {
    slug: 'endpoint-picker-for-inference-routing',
    eyebrow: 'Feature',
    icon: 'waypoints',
    title: 'Endpoint Picker Support for Intelligent Inference Routing',
    visual: 'fade',
    image: '/site-img/blog/epp-blog-overview.png',
  },
  {
    slug: 'v0.3-release-announcement',
    eyebrow: 'Release v0.3',
    icon: 'rocket',
    title: 'Announcing the Envoy AI Gateway v0.3 Release',
    subtitle: 'Intelligent inference routing, Vertex AI support, OpenInference tracing',
  },
  {
    slug: 'openinference-for-ai-observability',
    eyebrow: 'Observability',
    icon: 'activity',
    title: 'OpenTelemetry Tracing Arrives in Envoy AI Gateway',
    visual: 'fade',
    image: '/site-img/blog/envoy-ai-gateway-phoenix.drawio.png',
  },
  {
    slug: 'mcp-implementation',
    eyebrow: 'Feature — MCP',
    icon: 'plug',
    title: 'Announcing Model Context Protocol Support',
    subtitle: 'Enterprise-grade security, routing, and observability for AI agent tools',
  },
  {
    slug: 'mcp-in-envoy-ai-gateway',
    eyebrow: 'Deep dive — MCP',
    icon: 'network',
    title: 'The Reality and Performance of MCP Traffic Routing',
    visual: 'frame',
    image: '/site-img/blog/mcp-routing-benchmark-chart.png',
  },
  {
    slug: 'benchmarking-control-plane-scaling',
    eyebrow: 'Benchmarks',
    icon: 'gauge',
    title: 'Benchmarking Control Plane Scaling to 2,000 Routes',
    visual: 'frame',
    image: '/site-img/blog/benchmarking-route-readiness-latency.png',
  },
  {
    slug: 'v1.0-release-announcement',
    eyebrow: 'Release v1.0',
    icon: 'shield-check',
    title: 'Envoy AI Gateway 1.0 — Stable and Production-Ready',
    accent: 'signature',
  },
];

/* The site-wide fallback social card (docusaurus.config.ts → image:). */
export const DEFAULT_CARD = {
  out: 'static/img/social-card-envoy-ai-gw.png',
  config: {
    title: 'Route, govern, and observe LLM traffic',
    subtitle: 'The open source AI gateway built on Envoy',
    eyebrow: 'Envoy AI Gateway',
    icon: 'route',
  },
};
