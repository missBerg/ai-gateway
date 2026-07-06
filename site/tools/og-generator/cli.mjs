/* Headless CLI — generate an OG image without opening the preview UI.

   node cli.mjs --title "My post" --eyebrow "Feature" --icon waypoints \
     --out ../../static/img/blog/og/my-post.png

   Options:
     --title     (required) card headline
     --eyebrow   small uppercase label above the title
     --subtitle  one-liner under the title
     --icon      lucide icon name (see https://lucide.dev/icons) — eyebrow badge
                 and, with --visual icon, the corner texture
     --visual    icon | fade | frame | none            (default: icon)
     --image     side image for fade/frame — /site-img/... URL path or a
                 filesystem path inside the repo
     --accent    brand | signature                     (default: brand)
     --footer    footer text          (default: aigateway.envoyproxy.io)
     --out       (required) output PNG path, relative to this directory */

import path from 'node:path';
import { parseArgs } from 'node:util';
import { fileURLToPath } from 'node:url';
import { startServer } from './server.mjs';
import { renderCards } from './lib/render.mjs';

const TOOL_DIR = path.dirname(fileURLToPath(import.meta.url));
const REPO_DIR = path.resolve(TOOL_DIR, '..', '..', '..');

const { values: args } = parseArgs({
  options: {
    title: { type: 'string' },
    eyebrow: { type: 'string' },
    subtitle: { type: 'string' },
    icon: { type: 'string' },
    visual: { type: 'string', default: 'icon' },
    image: { type: 'string' },
    accent: { type: 'string', default: 'brand' },
    footer: { type: 'string' },
    out: { type: 'string' },
  },
});

if (!args.title || !args.out) {
  console.error('usage: node cli.mjs --title "..." --out path/to/image.png [options] (see file header)');
  process.exit(1);
}

/* Filesystem image paths are served through the dev server's /file route. */
let image = args.image;
if (image && !image.startsWith('/') && !image.startsWith('http')) {
  const abs = path.resolve(process.cwd(), image);
  const rel = path.relative(REPO_DIR, abs);
  if (rel.startsWith('..')) {
    console.error(`--image must be inside the repo (${REPO_DIR}); got ${abs}`);
    process.exit(1);
  }
  image = `/file?p=${encodeURIComponent(rel)}`;
}

const { url, close } = await startServer(0);
try {
  const out = path.resolve(process.cwd(), args.out);
  await renderCards(url, [{ config: { ...args, image, out: undefined }, out }]);
  console.log(`wrote ${out}`);
} finally {
  await close();
}
