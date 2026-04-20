# Envoy AI Gateway Site — Design System

A short, contributor-facing reference for the visual language of envoyproxy.io/ai-gateway. The
source of truth lives in code — this document explains how to use it.

- **Tokens:** [`site/src/css/tokens.css`](./src/css/tokens.css) — every color, space, radius,
  shadow, font size, font weight, and motion constant.
- **Selectors & primitives:** [`site/src/css/custom.css`](./src/css/custom.css) — hero, buttons,
  navbar, section primitives, markdown styles, admonitions, `.doc-steps`.
- **Homepage components:** `site/src/components/*/styles.module.css` — consume tokens + primitives.

If you change design language, change the token. Component CSS should not contain raw hex/rgb/px
literals (see **"When to add a new token"** below for the one exception).

---

## Principles

1. **Consistency before cleverness.** Reach for an existing token or primitive before inventing
   something new. Drift is the #1 thing this system exists to prevent.
2. **Brand, applied with restraint.** Purple + cyan are the brand. Use them for actions,
   highlights, and section eyebrows — not for decoration.
3. **Legibility & accessibility.** Body text is `--neutral-800` on white; small text on brand
   navy must keep ≥4.5:1 contrast. Always verify with `preview_inspect`.
4. **Open-source lightness.** This is a CNCF gateway, not a marketing page for a paid SaaS. Prefer
   clear typography and breathing room over gradients and glass effects.

---

## How to use tokens

```css
/* ✅ correct */
.card {
  background: #fff; /* literal white is fine */
  border: 1px solid var(--neutral-200);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
  transition: box-shadow var(--motion-base) var(--ease-out);
}
.card:hover {
  border-color: var(--tint-purple-border);
}

/* ❌ do not */
.card {
  border: 1px solid #e4e4e7; /* use --neutral-200 */
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05); /* use --shadow-sm */
  transition: box-shadow 0.25s ease; /* use --motion-base + --ease-out */
}
```

**The rule:** raw color/spacing/shadow/motion literals outside `tokens.css` are bugs unless you've
added a comment explaining why (e.g., macOS terminal window-dot colors are semantic — they map to
macOS, not to the brand).

---

## Section primitives

Every homepage section uses the same header pattern. Don't reinvent it.

```tsx
<section className={styles.section}>
  <div className="container">
    <div className="sectionHeader">
      <span className="sectionEyebrow sectionEyebrow--purple">
        Capabilities
      </span>
      <Heading as="h2" className="sectionTitle">
        One gateway, every major LLM
      </Heading>
      <p className="sectionSubtitle">
        Route traffic through a single OpenAI-compatible API.
      </p>
    </div>
    {/* section body */}
  </div>
</section>
```

Available primitives (all defined in `custom.css`):

| Class                     | Purpose                                                  |
| ------------------------- | -------------------------------------------------------- |
| `.sectionWrap`            | Standard vertical padding + max-width container.         |
| `.sectionHeader`          | Centered header block (eyebrow / title / subtitle).      |
| `.sectionEyebrow`         | Pill label above the title. Add a color modifier.        |
| `.sectionEyebrow--purple` | Purple eyebrow on light backgrounds (most sections).     |
| `.sectionEyebrow--accent` | Cyan eyebrow for MCP / interactive sections.             |
| `.sectionTitle`           | Clamp-sized section title (`--title-section`).           |
| `.sectionSubtitle`        | Lead paragraph under the title.                          |
| `.sectionOverlay--brand`  | Radial brand-purple glow — use on `--brand-navy` cards.  |
| `.card-base`              | White card with border / radius / shadow hover baseline. |

Per-component CSS should not redeclare these — the LLMProviders / LatestBlogs / Adopters refactor
removed every local `.sectionTitle` and `.titleUnderline`.

---

## Component patterns

### Buttons

- **`.button--accent`** — cyan gradient pill; hero CTAs and the navbar "Get started" button.
- **`.button--primary`** (Infima default) — solid purple; standard actions.
- **`.button--ghost`** / `.githubBtn` — transparent with white border; use on brand-navy backgrounds.

### Status badges

- **New** — cyan pill (`--status-new-bg` / `--status-new-text`). See
  [`LLMProviders/styles.module.css`](./src/components/LLMProviders/styles.module.css) `.newBadge`.
- **Coming soon** — soft-gold pill (`--status-coming-soon-*`). Reserved for future use.

### Admonitions

Docusaurus admonitions (`:::tip`, `:::note`, `:::info`, `:::warning`, `:::danger`) are
tokenized in `custom.css`. Use them directly in markdown — no custom styling needed.

### Tables

Markdown tables render full-width with horizontal scroll on narrow viewports. Don't override.

### `<DocCardGrid>`

Use on docs landing pages instead of a bulleted list of sub-links.

```mdx
<DocCardGrid
  columns={3}
  cards={[
    {
      title: "Connect providers",
      href: "./llm-integrations/connect-providers",
      description: "Establish connectivity with any supported AI provider.",
      icon: "🔌",
    },
    {
      title: "Supported providers",
      href: "./llm-integrations/supported-providers",
      description: "Compatible AI/LLM service providers.",
      icon: "🧭",
    },
  ]}
/>
```

`columns` accepts `2 | 3 | 4`, defaults to `3`. Collapses to one column below 576px.

### `.doc-steps`

Numbered step sequences for multi-step tutorials. Apply to an `<ol>`:

```md
<ol class="doc-steps">
  <li>Install the controller.</li>
  <li>Deploy an `AIGatewayRoute`.</li>
  <li>Send your first request.</li>
</ol>
```

Not every numbered list needs this treatment — reserve it for stepwise tutorials where the order
is load-bearing.

---

## Responsive breakpoints

Mobile-first. CSS `@media` cannot consume custom properties, so treat these as shared constants:

| Breakpoint | Max width | Typical use                    |
| ---------- | --------- | ------------------------------ |
| Small      | 576px     | Phones — stack to one column.  |
| Mobile     | 768px     | Larger phones & small tablets. |
| Tablet     | 996px     | Tablets & small laptops.       |
| Desktop    | 1200px    | Full grid layouts.             |

Always test at 576 / 768 / 996 / 1200 / 1400 via `preview_resize`. No horizontal scroll on the
homepage at any size.

---

## Accessibility checklist

Before shipping a visual change:

- **Contrast.** Body text ≥4.5:1; large text and icons ≥3:1. Use `preview_inspect` or the
  browser dev tools color picker.
- **Focus rings.** Use `:focus-visible` with `outline: 2px solid var(--brand-purple); outline-offset: 2px;`.
  Every interactive element (link, button, card) gets one.
- **Alt text.** Decorative images get `alt=""`; meaningful images get a real description.
- **Headings.** One `<h1>` per page. Section titles are `<h2>`, subsections `<h3>`.
- **Motion.** Keep durations ≤400ms. Respect `prefers-reduced-motion` if you add anything animated
  beyond a simple hover state.

---

## When to add a new token

Decision tree:

1. **Is this value used in two or more places?** → Add a token.
2. **Is it used once but likely to recur?** → Add a token (with a descriptive name).
3. **Is it a semantic literal tied to something external** (macOS window-chrome colors, a
   third-party brand color in an icon)? → Inline with a comment explaining why.
4. **None of the above?** → Inline, but reconsider why you need it.

Add tokens to the appropriate section of `tokens.css`. Name them by role, not value:

- ✅ `--tint-purple-border` (role)
- ❌ `--purple-35` (value)

---

## Verifying changes

1. **Build:** `npm run build` from `site/`. Must not add broken-link warnings over the baseline.
2. **Dev server:** `npm run start`. Preview via the `preview_*` MCP tools.
3. **Token audit:** `grep -nE '#[0-9a-fA-F]{3,8}|rgba?\(' src/components src/css/custom.css`
   should return nothing outside `tokens.css` (and any commented semantic literals).
4. **Responsive:** `preview_resize` to 1400 / 1200 / 996 / 768 / 576 — screenshot each.

---

## File map

| File                                         | Contains                                      |
| -------------------------------------------- | --------------------------------------------- |
| `src/css/tokens.css`                         | All CSS custom properties. Source of truth.   |
| `src/css/custom.css`                         | Selectors, primitives, markdown, admonitions. |
| `src/components/<Section>/styles.module.css` | Per-section styles. Tokens only.              |
| `src/components/DocCardGrid/`                | `<DocCardGrid>` MDX component.                |
| `src/components/ApiField.tsx`                | `<ApiField>` MDX component for API reference. |
| `src/theme/MDXComponents.tsx`                | Registers MDX-visible components.             |
