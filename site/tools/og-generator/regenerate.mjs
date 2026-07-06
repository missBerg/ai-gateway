/* Batch-regenerate the OG images for every blog post in posts.config.mjs,
   plus the site-wide default social card. Output: site/static/img/blog/og/. */

import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { startServer } from './server.mjs';
import { renderCards } from './lib/render.mjs';
import { POSTS, DEFAULT_CARD } from './posts.config.mjs';

const TOOL_DIR = path.dirname(fileURLToPath(import.meta.url));
const SITE_DIR = path.resolve(TOOL_DIR, '..', '..');

const jobs = POSTS.map(({ slug, ...config }) => ({
  config,
  out: path.join(SITE_DIR, 'static', 'img', 'blog', 'og', `${slug}.png`),
}));
jobs.push({ config: DEFAULT_CARD.config, out: path.join(SITE_DIR, DEFAULT_CARD.out) });

const { url, close } = await startServer(0);
try {
  const written = await renderCards(url, jobs);
  for (const file of written) console.log(`wrote ${path.relative(SITE_DIR, file)}`);
  console.log(`\n${written.length} images generated.`);
} finally {
  await close();
}
