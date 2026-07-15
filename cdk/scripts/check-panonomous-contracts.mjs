#!/usr/bin/env node
import { readFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(__dirname, '../..');
const siteDist = path.resolve(repoRoot, 'cdk/dist/site');

const expected = [
  {
    source: 'contracts/panonomous/soul-document/v1/schema.json',
    dist: 'contracts/panonomous/soul-document/v1/schema.json',
    id: 'https://spec.lessersoul.ai/contracts/panonomous/soul-document/v1/schema.json',
    required: ['agent_id', 'body'],
  },
  {
    source: 'contracts/panonomous/agent-naming/v1/vocabulary.json',
    dist: 'contracts/panonomous/agent-naming/v1/vocabulary.json',
    id: 'https://spec.lessersoul.ai/contracts/panonomous/agent-naming/v1/vocabulary.json',
    required: [],
  },
];

async function readJSON(relativePath) {
  const absolutePath = path.resolve(repoRoot, relativePath);
  const text = await readFile(absolutePath, 'utf8');
  return JSON.parse(text);
}

for (const contract of expected) {
  const source = await readJSON(contract.source);
  if (source.$id !== contract.id) {
    throw new Error(`${contract.source} $id = ${source.$id}; want ${contract.id}`);
  }
  for (const requiredField of contract.required) {
    if (!source.required?.includes(requiredField)) {
      throw new Error(`${contract.source} missing required field ${requiredField}`);
    }
  }

  const built = await readJSON(path.relative(repoRoot, path.join(siteDist, contract.dist)));
  if (built.$id !== source.$id) {
    throw new Error(`${contract.dist} was not copied to the site build output with the expected $id`);
  }
}

const namespaceV1 = await readFile(path.resolve(repoRoot, 'cdk/site/static/ns/agent-attribution/v1'), 'utf8');
const namespace = JSON.parse(namespaceV1);
if (namespace?.['@context']?.agentAttribution?.['@id'] !== 'lessersoul:agentAttribution') {
  throw new Error('agent-attribution /v1 namespace invariant changed unexpectedly');
}

console.log('Panonomous contract artifacts validated.');
