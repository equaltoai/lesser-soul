import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { cp, mkdir, rm } from 'node:fs/promises';

import { buildSsgSite } from '@theory-cloud/facetheory';

import { faces } from './faces.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const outDir = path.resolve(__dirname, '../dist/site');
const staticDir = path.resolve(__dirname, './static');

await rm(outDir, { recursive: true, force: true });
await mkdir(outDir, { recursive: true });

const result = await buildSsgSite({
  faces,
  outDir,
  trailingSlash: 'always',
  emitHydrationData: false,
});

await cp(staticDir, outDir, { recursive: true });

console.log(`SSG build wrote ${result.pages.length} page(s) to ${result.outDir}`);
