import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { cp, mkdir, readFile, rm } from 'node:fs/promises';

import { buildSsgSite, validateStrictCspDocument } from '@theory-cloud/facetheory';

import { faces } from './faces.js';
import { SITE_STYLESHEET_PATH, strictSiteCspPolicy } from './security.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(__dirname, '../..');
const outDir = path.resolve(__dirname, '../dist/site');
const staticDir = path.resolve(__dirname, './static');
const contractsDir = path.resolve(repoRoot, 'contracts');

await rm(outDir, { recursive: true, force: true });
await mkdir(outDir, { recursive: true });

const result = await buildSsgSite({
  faces,
  outDir,
  trailingSlash: 'always',
  emitHydrationData: false,
});

await cp(staticDir, outDir, { recursive: true });
await cp(contractsDir, path.join(outDir, 'contracts'), { recursive: true });

const stylesheet = await readFile(path.join(outDir, SITE_STYLESHEET_PATH.slice(1)), 'utf8');
if (stylesheet.trim().length === 0) {
  throw new Error(`Strict-CSP stylesheet is empty: ${SITE_STYLESHEET_PATH}`);
}

const generatedHtmlFiles = new Set([
  ...result.pages.map((page) => page.file),
  ...(result.notFoundFile ? [result.notFoundFile] : []),
]);

for (const relativeFile of generatedHtmlFiles) {
  const html = await readFile(path.join(outDir, relativeFile), 'utf8');
  validateStrictCspDocument(html, { policy: strictSiteCspPolicy });

  const headEnd = html.indexOf('</head>');
  const head = headEnd >= 0 ? html.slice(0, headEnd) : '';
  if (!head.includes(`href="${SITE_STYLESHEET_PATH}"`)) {
    throw new Error(`Strict-CSP stylesheet link is missing from ${relativeFile}`);
  }
}

console.log(`SSG build wrote ${result.pages.length} page(s) to ${result.outDir}`);
