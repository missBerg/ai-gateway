/* Renders card configs to PNG by screenshotting the card page with
   playwright-core driving an already-installed browser (Chrome/Edge/Chromium),
   so nobody has to download the Playwright browser bundle. */

import { mkdir } from 'node:fs/promises';
import path from 'node:path';
import { chromium } from 'playwright-core';

async function launchBrowser() {
  const attempts = [
    { channel: 'chrome' },
    { channel: 'msedge' },
    {}, // bundled chromium, if the user ran `npx playwright install chromium`
  ];
  const errors = [];
  for (const options of attempts) {
    try {
      return await chromium.launch(options);
    } catch (err) {
      errors.push(err.message.split('\n')[0]);
    }
  }
  throw new Error(
    'No browser found. Install Google Chrome or Microsoft Edge, or run ' +
      `"npx playwright install chromium" in this directory.\n${errors.join('\n')}`,
  );
}

function cardUrl(baseUrl, config) {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(config)) {
    if (value !== undefined && value !== null && value !== '') params.set(key, value);
  }
  return `${baseUrl}/card?${params.toString()}`;
}

/** Render one or more {config, out} jobs, reusing a single browser. */
export async function renderCards(baseUrl, jobs) {
  const browser = await launchBrowser();
  try {
    const page = await browser.newPage({
      viewport: { width: 1280, height: 700 },
      deviceScaleFactor: 2, // 1200×630 card → crisp 2400×1260 PNG
    });
    const results = [];
    for (const job of jobs) {
      await page.goto(cardUrl(baseUrl, job.config));
      await page.waitForFunction(() => window.__ogReady || window.__ogError, null, {
        timeout: 15_000,
      });
      const renderError = await page.evaluate(() => window.__ogError);
      if (renderError) throw new Error(`card render failed: ${renderError}`);
      await mkdir(path.dirname(job.out), { recursive: true });
      await page.locator('#card').screenshot({ path: job.out });
      results.push(job.out);
    }
    return results;
  } finally {
    await browser.close();
  }
}
