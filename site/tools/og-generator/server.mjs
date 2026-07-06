/* Dev server for the OG image generator.
   Serves the preview UI, the card page, lucide icons, Inter fonts, and the
   site's static images; POST /export screenshots the card to a PNG.

   Run from this directory:  npm start   →  http://localhost:4630 */

import http from 'node:http';
import { readFile, readdir } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { renderCards } from './lib/render.mjs';

const TOOL_DIR = path.dirname(fileURLToPath(import.meta.url));
const SITE_DIR = path.resolve(TOOL_DIR, '..', '..');
const REPO_DIR = path.resolve(SITE_DIR, '..');
const LUCIDE_DIR = path.join(TOOL_DIR, 'node_modules', 'lucide-static', 'icons');
const FONT_DIR = path.join(TOOL_DIR, 'node_modules', '@fontsource', 'inter', 'files');

const MIME = {
  '.html': 'text/html; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.js': 'text/javascript; charset=utf-8',
  '.mjs': 'text/javascript; charset=utf-8',
  '.json': 'application/json',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.gif': 'image/gif',
  '.webp': 'image/webp',
  '.woff2': 'font/woff2',
  '.ico': 'image/x-icon',
};

function send(res, status, body, type = 'text/plain; charset=utf-8') {
  res.writeHead(status, { 'Content-Type': type });
  res.end(body);
}

/* Resolve `rel` inside `root`; null if it escapes (path traversal). */
function safeJoin(root, rel) {
  const resolved = path.resolve(root, rel);
  return resolved === root || resolved.startsWith(root + path.sep) ? resolved : null;
}

async function serveFile(res, filePath) {
  try {
    const body = await readFile(filePath);
    send(res, 200, body, MIME[path.extname(filePath).toLowerCase()] || 'application/octet-stream');
  } catch {
    send(res, 404, 'not found');
  }
}

function readBody(req) {
  return new Promise((resolve, reject) => {
    const chunks = [];
    req.on('data', (c) => chunks.push(c));
    req.on('end', () => resolve(Buffer.concat(chunks).toString('utf8')));
    req.on('error', reject);
  });
}

export function startServer(port = 4630) {
  return new Promise((resolve) => {
    const server = http.createServer(async (req, res) => {
      const url = new URL(req.url, 'http://localhost');
      const route = url.pathname;

      if (req.method === 'POST' && route === '/export') {
        try {
          const { config, out } = JSON.parse(await readBody(req));
          const outPath = safeJoin(REPO_DIR, out || '');
          if (!outPath || !outPath.endsWith('.png')) {
            return send(res, 400, JSON.stringify({ error: 'out must be a .png path inside the repo' }), MIME['.json']);
          }
          const baseUrl = `http://localhost:${server.address().port}`;
          await renderCards(baseUrl, [{ config, out: outPath }]);
          return send(res, 200, JSON.stringify({ path: outPath }), MIME['.json']);
        } catch (err) {
          return send(res, 500, JSON.stringify({ error: String(err.message || err) }), MIME['.json']);
        }
      }

      if (route === '/' || route === '/index.html') {
        return serveFile(res, path.join(TOOL_DIR, 'preview', 'index.html'));
      }
      if (route === '/card' || route === '/card/') {
        return serveFile(res, path.join(TOOL_DIR, 'card', 'card.html'));
      }
      if (route.startsWith('/card/')) {
        const file = safeJoin(path.join(TOOL_DIR, 'card'), route.slice('/card/'.length));
        return file ? serveFile(res, file) : send(res, 404, 'not found');
      }
      if (route.startsWith('/lucide/')) {
        const file = safeJoin(LUCIDE_DIR, route.slice('/lucide/'.length));
        return file ? serveFile(res, file) : send(res, 404, 'not found');
      }
      if (route === '/lucide-index.json') {
        const names = (await readdir(LUCIDE_DIR))
          .filter((f) => f.endsWith('.svg'))
          .map((f) => f.replace(/\.svg$/, ''));
        return send(res, 200, JSON.stringify(names), MIME['.json']);
      }
      if (route.startsWith('/fonts/')) {
        const file = safeJoin(FONT_DIR, route.slice('/fonts/'.length));
        return file ? serveFile(res, file) : send(res, 404, 'not found');
      }
      /* The site's static images, so cards can reference existing assets
         (e.g. /site-img/logo-white.svg or /site-img/blog/foo.png). */
      if (route.startsWith('/site-img/')) {
        const file = safeJoin(path.join(SITE_DIR, 'static', 'img'), route.slice('/site-img/'.length));
        return file ? serveFile(res, file) : send(res, 404, 'not found');
      }
      /* Any file inside the repo, for ad-hoc side images: /file?p=<repo-relative path> */
      if (route === '/file') {
        const file = safeJoin(REPO_DIR, url.searchParams.get('p') || '');
        return file ? serveFile(res, file) : send(res, 404, 'not found');
      }

      send(res, 404, 'not found');
    });

    server.listen(port, () => {
      resolve({
        server,
        port: server.address().port,
        url: `http://localhost:${server.address().port}`,
        close: () => new Promise((r) => server.close(r)),
      });
    });
  });
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  const { url } = await startServer(Number(process.env.PORT) || 4630);
  console.log(`OG image generator running at ${url}`);
}
