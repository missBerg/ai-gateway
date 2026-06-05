// Single source of truth for supported AI providers.
//
// Consumed by:
//   - src/components/LLMProviders   → compact logo grid on the landing page
//                                      (entries with showOnLanding !== false)
//   - src/components/ProvidersMatrix → full sortable/filterable matrix rendered
//                                      in docs/.../supported-providers.mdx
//
// Adding a provider: append one entry here and both surfaces update. Keep the
// list ordered alphabetically by `name` — the matrix default-sorts ascending by
// name, so this ordering matches what users see first.

export type ProviderStatus = 'supported' | 'partial' | 'planned';

// Auth link targets the generated API reference at /docs/api/.
export type AuthKind =
  | 'api-key'
  | 'aws-bedrock'
  | 'gcp'
  | 'azure-credentials'
  | 'azure-api-key'
  | 'anthropic-api-key'
  | 'none';

export type SchemaVariant = {
  // Optional short label when a provider exposes more than one schema
  // (e.g. "Native" vs "OpenAI-compatible"). Omit if there's only one.
  label?: string;
  // The actual AIServiceBackend `schema` value, as a plain object so we can
  // pretty-print it. Keys mirror the CRD fields.
  config: {
    name: string;
    prefix?: string;
    version?: string;
  };
};

export type Provider = {
  name: string;
  url?: string;
  schemas: SchemaVariant[];
  auth: AuthKind[];
  status: ProviderStatus;
  note?: string;
  // ---- Landing-grid fields (ignored by the matrix) ----
  logoUrl?: string;        // /img/providers/*.svg — required when showOnLanding
  gridLabel?: string;      // shorter label for the logo grid; falls back to name
  showOnLanding?: boolean; // default true; set false for matrix-only entries
  addedIn?: string;        // e.g. '0.6' — drives a "New" badge vs the latest version
};

// Links into the auto-generated API reference.
export const AUTH_META: Record<AuthKind, { label: string; href?: string }> = {
  'api-key': { label: 'API Key', href: '/docs/api/#backendsecuritypolicyapikey' },
  'aws-bedrock': { label: 'AWS Bedrock Credentials', href: '/docs/api/#backendsecuritypolicyawscredentials' },
  gcp: { label: 'GCP Credentials', href: '/docs/api/#backendsecuritypolicygcpcredentials' },
  'azure-credentials': { label: 'Azure Credentials', href: '/docs/api/#backendsecuritypolicyazurecredentials' },
  'azure-api-key': { label: 'Azure API Key', href: '/docs/api/#backendsecuritypolicyazureapikey' },
  'anthropic-api-key': { label: 'Anthropic API Key', href: '/docs/api/#backendsecuritypolicyanthropicapikey' },
  none: { label: 'N/A' },
};

export const PROVIDERS: Provider[] = [
  {
    name: 'Anthropic',
    url: 'https://docs.claude.com/en/home',
    schemas: [{ config: { name: 'Anthropic' } }],
    auth: ['anthropic-api-key'],
    status: 'supported',
    note: 'Native Anthropic messages endpoint only.',
    logoUrl: '/img/providers/anthropic.svg',
  },
  {
    name: 'Anthropic on GCP Vertex AI',
    url: 'https://cloud.google.com/vertex-ai/generative-ai/docs/partner-models/claude',
    schemas: [{ config: { name: 'GCPAnthropic', version: 'vertex-2023-10-16' } }],
    auth: ['gcp'],
    status: 'supported',
    note: 'Native Anthropic and OpenAI-compatible endpoints. Supports ADC, Service Account Keys, and Workload Identity Federation.',
    showOnLanding: false,
  },
  {
    name: 'AWS Bedrock',
    url: 'https://docs.aws.amazon.com/bedrock/latest/APIReference/',
    schemas: [{ config: { name: 'AWSBedrock' } }],
    auth: ['aws-bedrock'],
    status: 'supported',
    logoUrl: '/img/providers/aws-bedrock.svg',
  },
  {
    name: 'Azure OpenAI',
    url: 'https://learn.microsoft.com/en-us/azure/ai-services/openai/reference',
    schemas: [
      { label: 'Native', config: { name: 'AzureOpenAI', version: '2025-01-01-preview' } },
      { label: 'OpenAI-compatible', config: { name: 'OpenAI', prefix: '/openai/v1' } },
    ],
    auth: ['azure-credentials', 'azure-api-key'],
    status: 'supported',
    logoUrl: '/img/providers/azure-openai.svg',
  },
  {
    name: 'Cohere',
    url: 'https://docs.cohere.com/v2/docs/compatibility-api',
    schemas: [
      { label: 'Native', config: { name: 'Cohere', version: 'v2' } },
      { label: 'OpenAI-compatible', config: { name: 'OpenAI', prefix: '/compatibility/v1' } },
    ],
    auth: ['api-key'],
    status: 'supported',
    note: 'Native Cohere v2 (e.g. /cohere/v2/rerank) and OpenAI-compatible endpoints.',
    logoUrl: '/img/providers/cohere.svg',
  },
  {
    name: 'DeepInfra',
    url: 'https://deepinfra.com/docs/inference',
    schemas: [{ config: { name: 'OpenAI', prefix: '/v1/openai' } }],
    auth: ['api-key'],
    status: 'supported',
    note: 'OpenAI-compatible endpoint only.',
    logoUrl: '/img/providers/deepinfra.svg',
  },
  {
    name: 'DeepSeek',
    url: 'https://api-docs.deepseek.com/',
    schemas: [{ config: { name: 'OpenAI', prefix: '/v1' } }],
    auth: ['api-key'],
    status: 'supported',
    logoUrl: '/img/providers/deepseek.svg',
  },
  {
    name: 'Google Gemini on AI Studio',
    url: 'https://ai.google.dev/gemini-api/docs/openai',
    schemas: [{ config: { name: 'OpenAI', prefix: '/v1beta/openai' } }],
    auth: ['api-key'],
    status: 'supported',
    note: 'OpenAI-compatible endpoint only.',
    logoUrl: '/img/providers/google-gemini.svg',
    gridLabel: 'Google Gemini',
  },
  {
    name: 'Google Vertex AI',
    url: 'https://cloud.google.com/vertex-ai/docs/reference/rest',
    schemas: [{ config: { name: 'GCPVertexAI' } }],
    auth: ['gcp'],
    status: 'supported',
    note: 'Supports ADC, Service Account Keys, and Workload Identity Federation.',
    logoUrl: '/img/providers/vertex-ai.svg',
    gridLabel: 'Vertex AI',
  },
  {
    name: 'Grok',
    url: 'https://docs.x.ai/docs/api-reference#chat-completions',
    schemas: [{ config: { name: 'OpenAI', prefix: '/v1' } }],
    auth: ['api-key'],
    status: 'supported',
    logoUrl: '/img/providers/grok.svg',
  },
  {
    name: 'Groq',
    url: 'https://console.groq.com/docs/openai',
    schemas: [{ config: { name: 'OpenAI', prefix: '/openai/v1' } }],
    auth: ['api-key'],
    status: 'supported',
    logoUrl: '/img/providers/groq.svg',
  },
  {
    name: 'Hunyuan',
    url: 'https://cloud.tencent.com/document/product/1729/111007',
    schemas: [{ config: { name: 'OpenAI', prefix: '/v1' } }],
    auth: ['api-key'],
    status: 'supported',
    logoUrl: '/img/providers/hunyuan.svg',
  },
  {
    name: 'Mistral',
    url: 'https://docs.mistral.ai/api/#tag/chat/operation/chat_completion_v1_chat_completions_post',
    schemas: [{ config: { name: 'OpenAI', prefix: '/v1' } }],
    auth: ['api-key'],
    status: 'supported',
    logoUrl: '/img/providers/mistral.svg',
  },
  {
    name: 'OpenAI',
    url: 'https://platform.openai.com/docs/api-reference',
    schemas: [{ config: { name: 'OpenAI', prefix: '/v1' } }],
    auth: ['api-key'],
    status: 'supported',
    logoUrl: '/img/providers/openai.svg',
  },
  {
    name: 'SambaNova',
    url: 'https://docs.sambanova.ai/sambastudio/latest/open-ai-api.html',
    schemas: [{ config: { name: 'OpenAI', prefix: '/v1' } }],
    auth: ['api-key'],
    status: 'supported',
    logoUrl: '/img/providers/sambanova.svg',
  },
  {
    name: 'Self-hosted models',
    schemas: [{ config: { name: 'OpenAI', prefix: '/v1' } }],
    auth: ['none'],
    status: 'partial',
    note: 'Depends on the API schema spoken by the self-hosted server (e.g. vLLM speaks the OpenAI format). API Key auth can still be configured.',
    showOnLanding: false,
  },
  {
    name: 'Tencent LLM Knowledge Engine',
    url: 'https://www.tencentcloud.com/document/product/1255/70381?lang=en',
    schemas: [{ config: { name: 'OpenAI', prefix: '/v1' } }],
    auth: ['api-key'],
    status: 'supported',
    showOnLanding: false,
  },
  {
    name: 'Tetrate Agent Router Service (TARS)',
    url: 'https://router.tetrate.ai/',
    schemas: [{ config: { name: 'OpenAI', prefix: '/v1' } }],
    auth: ['api-key'],
    status: 'supported',
    logoUrl: '/img/providers/tars.svg',
    gridLabel: 'Tetrate Agent Router',
  },
  {
    name: 'Together AI',
    url: 'https://docs.together.ai/docs/openai-api-compatibility',
    schemas: [{ config: { name: 'OpenAI', prefix: '/v1' } }],
    auth: ['api-key'],
    status: 'supported',
    logoUrl: '/img/providers/together-ai.svg',
  },
];

// Providers shown as logo tiles on the landing page, alphabetised by label.
export const LANDING_PROVIDERS: Provider[] = PROVIDERS.filter(
  (p) => p.showOnLanding !== false,
).sort((a, b) => (a.gridLabel ?? a.name).localeCompare(b.gridLabel ?? b.name));
