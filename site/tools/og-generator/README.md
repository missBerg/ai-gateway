# OG Image Generator

Generates 1200×630 Open Graph images for blog posts by rendering an HTML card
(styled with the site's design tokens — navy surface, Inter, indigo accent,
lucide icon textures) and screenshotting it to PNG at 2× resolution.

No browser download needed: it drives your installed Chrome/Edge via
`playwright-core`. (Fallback: `npx playwright install chromium`.)

## Setup

```bash
cd site/tools/og-generator
npm install
```

## Interactive preview UI

```bash
npm start
# → http://localhost:4630
```

Live-edit the title, eyebrow, lucide icon, and visual treatment; hit
**Export PNG** to write the file into the repo. The UI also prints the
equivalent CLI command for reproducibility.

## CLI (for humans and coding agents)

```bash
node cli.mjs \
  --title "My New Post Title" \
  --eyebrow "Feature" \
  --icon waypoints \
  --out ../../static/img/blog/og/my-new-post.png
```

Options:

| Flag         | Values                                      | Notes                                                        |
| ------------ | ------------------------------------------- | ------------------------------------------------------------ |
| `--title`    | text (required)                             | auto-shrinks to fit                                          |
| `--eyebrow`  | text                                        | small uppercase label, indigo accent                         |
| `--subtitle` | text                                        | optional one-liner under the title                           |
| `--icon`     | any [lucide](https://lucide.dev/icons) name | eyebrow badge + corner texture                               |
| `--visual`   | `icon` \| `fade` \| `frame` \| `none`       | right-side treatment (default `icon`)                        |
| `--image`    | `/site-img/...` or repo file path           | required for `fade`/`frame`                                  |
| `--accent`   | `brand` \| `signature`                      | top keyline: violet→indigo, or magenta→cyan for big releases |
| `--footer`   | text                                        | defaults to `aigateway.envoyproxy.io`                        |
| `--out`      | path ending in `.png` (required)            |                                                              |

### Visual treatments

- **`icon`** — an oversized thin-line lucide icon bleeding out of the top-right
  corner, masked to fade into the navy (echoes the homepage capability cards).
  The default; works for every post.
- **`fade`** — a ghosted, brand-tinted tease: the image is grayscaled and
  multiplied over the violet→indigo gradient, cropped to bleed off the right
  edge, and masked to dissolve toward the text. Good for architecture
  diagrams — hints at the content without reproducing the figure.
- **`frame`** — the image sits in a flat browser-style frame on the right.
  Good for UI screenshots.
- **`none`** — text only.

## Regenerating all post images

Per-post configs (title, eyebrow, icon) live in [`posts.config.mjs`](posts.config.mjs).
Add an entry when you write a new post, then:

```bash
npm run regenerate
```

writes `site/static/img/blog/og/<slug>.png` for every entry plus the site-wide
default social card. Point the post's frontmatter at it:

```yaml
image: /img/blog/og/<slug>.png
```

## For coding agents

To create an OG image for a new blog post: read the post's frontmatter, pick a
lucide icon matching the concept, add an entry to `posts.config.mjs`, run
`npm run regenerate` (or the one-off CLI), and set the post's `image:` field to
`/img/blog/og/<slug>.png`.
