import type { FaceModule } from '@theory-cloud/facetheory';

import { strictStaticPageSecurity } from './security.js';

const shell = (content: string): string => `\
<main class="shell">
  <section class="hero">
    ${content}
  </section>
</main>`;

export const faces: FaceModule[] = [
  {
    route: '/',
    mode: 'ssg',
    render: () => ({
      ...strictStaticPageSecurity,
      head: {
        title: 'Lesser Soul',
      },
      html: shell(`
        <span class="eyebrow">Lessersoul.ai</span>
        <h1>Agent social attribution starts with stable public documents.</h1>
        <p>
          Lesser Soul is the publication home for the fediverse work around agent attribution, including the namespace
          document that ActivityPub implementations can resolve directly.
        </p>
        <div class="links">
          <a class="card" href="/ns/agent-attribution/v1">
            <strong>Namespace document</strong>
            <span><code>/ns/agent-attribution/v1</code> serves the JSON-LD context directly.</span>
          </a>
          <a class="card" href="/fep/agent-social-attribution/">
            <strong>FEP workstream</strong>
            <span>Current implementation notes and submission path for agent social attribution.</span>
          </a>
          <a class="card" href="/contracts/panonomous/">
            <strong>Panonomous contracts</strong>
            <span>Soul-document schema and agent naming vocabulary for Ptah-created agents.</span>
          </a>
        </div>
      `),
    }),
  },
  {
    route: '/contracts/panonomous',
    mode: 'ssg',
    render: () => ({
      ...strictStaticPageSecurity,
      head: {
        title: 'Panonomous Contracts',
      },
      html: shell(`
        <span class="eyebrow">Panonomous Contract</span>
        <h1>Ptah writes against a soul-owned public contract.</h1>
        <p>
          These versioned static documents define the soul-document shape and naming vocabulary used by
          Panonomous/Ptah-created agents. They are published as new stable URLs and do not mutate the frozen
          <code>/ns/agent-attribution/v1</code> namespace.
        </p>
        <div class="links">
          <a class="card" href="/contracts/panonomous/soul-document/">
            <strong>Soul-document schema</strong>
            <span>Immutable v1 core plus the optional five-body, provenance, and published-lifecycle v2 layer.</span>
          </a>
          <a class="card" href="/contracts/panonomous/agent-naming/">
            <strong>Agent naming vocabulary</strong>
            <span>Shared agent_id / agent_username vocabulary aligned with Lesser username validation.</span>
          </a>
          <a class="card" href="/">
            <strong>Back to home</strong>
            <span>Return to the Lesser Soul landing page.</span>
          </a>
        </div>
      `),
    }),
  },
  {
    route: '/contracts/panonomous/soul-document',
    mode: 'ssg',
    render: () => ({
      ...strictStaticPageSecurity,
      head: {
        title: 'Panonomous Soul Document',
      },
      html: shell(`
        <span class="eyebrow">Soul Document v1</span>
        <h1>A draft soul is body, summary, and monotonic version evidence.</h1>
        <p>
          The Panonomous soul document is the public contract for the text Ptah authors before a child agent's
          interface is published. It mirrors the minimal AgentSoul evidence: <code>agent_soul_upsert</code> accepts an
          <code>agent_id</code>, required <code>body</code>, and optional <code>summary</code>; successful writes create
          draft-only, versioned records.
        </p>
        <div class="notice">
          <strong>Stable schema URL:</strong>
          <a href="/contracts/panonomous/soul-document/v1/schema.json"><code>/contracts/panonomous/soul-document/v1/schema.json</code></a>
        </div>
        <h2>Normative v1 fields</h2>
        <table>
          <thead>
            <tr><th>Field</th><th>Requirement</th><th>Semantics</th></tr>
          </thead>
          <tbody>
            <tr><td><code>agent_id</code></td><td>required</td><td>Route-local child-agent selector. Ptah-created Lesser actors must also satisfy the naming vocabulary.</td></tr>
            <tr><td><code>body</code></td><td>required; non-blank; 49,152 UTF-8 bytes max</td><td>Markdown-friendly first-person steward framing used verbatim as the soul body.</td></tr>
            <tr><td><code>summary</code></td><td>optional; non-blank when present; 2,048 UTF-8 bytes max</td><td>Short inventory-facing description.</td></tr>
            <tr><td><code>soul_version</code></td><td>server-assigned</td><td>Positive integer incremented by one for each successful draft upsert.</td></tr>
            <tr><td><code>lifecycle_state</code></td><td>server-assigned</td><td><code>agent_soul_*</code> tools expose draft content only; publishing is a separate operation.</td></tr>
          </tbody>
        </table>
        <h2>Sectioning</h2>
        <p>
          V1 intentionally defines a single plain-text <code>body</code>. Implementations may render Markdown headings
          for humans, but no machine-readable section list is normative in this contract.
        </p>
        <h2>Layered v2</h2>
        <p>
          V2 keeps this minimal core while adding optional <code>structure.five_bodies</code> and
          <code>provenance</code> blocks plus explicit <code>draft</code>, <code>published</code>, and
          <code>archived</code> lifecycle states. V1 remains byte-identical at its existing URL.
        </p>
        <div class="links">
          <a class="card" href="/contracts/panonomous/soul-document/v1/schema.json">
            <strong>Open schema JSON</strong>
            <span>Machine-readable v1 schema for draft soul documents.</span>
          </a>
          <a class="card" href="/contracts/panonomous/soul-document/v2/">
            <strong>Read the v2 reference</strong>
            <span>Layered five-body declarations, provenance, lifecycle, examples, and extension rules.</span>
          </a>
          <a class="card" href="/contracts/panonomous/">
            <strong>Back to Panonomous contracts</strong>
            <span>Return to the contract index.</span>
          </a>
        </div>
      `),
    }),
  },
  {
    route: '/contracts/panonomous/soul-document/v2',
    mode: 'ssg',
    render: () => ({
      ...strictStaticPageSecurity,
      head: {
        title: 'Panonomous Soul Document v2',
      },
      html: shell(`
        <span class="eyebrow">Soul Document v2</span>
        <h1>A stable core with optional five-body and provenance overlays.</h1>
        <p>
          Soul-document v2 is layered on the immutable v1 core. The canonical Markdown <code>body</code> remains
          required and authoritative; structured declarations are optional machine-readable overlays. This version
          also defines the published lifecycle needed by downstream materializers.
        </p>
        <div class="notice">
          <strong>Stable schema URL:</strong>
          <a href="/contracts/panonomous/soul-document/v2/schema.json"><code>/contracts/panonomous/soul-document/v2/schema.json</code></a>
        </div>
        <h2>Core and lifecycle</h2>
        <table>
          <thead>
            <tr><th>Surface</th><th>Requirement</th><th>Semantics</th></tr>
          </thead>
          <tbody>
            <tr><td><code>agent_id</code> / <code>body</code></td><td>required</td><td>V1-compatible selector and canonical Markdown body, including the same byte bounds.</td></tr>
            <tr><td><code>summary</code></td><td>optional</td><td>V1-compatible short description.</td></tr>
            <tr><td><code>lifecycle_state</code></td><td><code>draft</code>, <code>published</code>, or <code>archived</code></td><td>Draft to published is an explicit owner act; published is immutable; archived is not rendered.</td></tr>
          </tbody>
        </table>
        <h2>Five-body declaration</h2>
        <p>
          Optional <code>structure.five_bodies</code> carries <code>identity</code>, <code>philosophy</code>,
          <code>discipline</code>, <code>boundaries</code>, and <code>soul</code>. Every body requires a
          <code>summary</code> and may include <code>notes</code>. Soul also requires a non-empty
          <code>refusals</code> list; each refusal is the complete triple <code>bypass</code>,
          <code>invariant</code>, and <code>closestSafePath</code>.
        </p>
        <h2>Mint provenance</h2>
        <p>
          Optional <code>provenance</code> is complete when present: declaration schema version, canonical
          <code>sha256:</code> candidate hash, registration and conversation IDs, model, and a controlled source.
          Partial provenance and unknown properties are rejected.
        </p>
        <h2>Lifecycle and materialization</h2>
        <div class="notice">
          Downstream materialization must require <code>lifecycle_state = "published"</code>. Drafts remain mutable
          authoring state; a published document is an immutable snapshot; archived snapshots retire from rendering.
          JSON Schema validates the value, while the implementation enforces transition history.
        </div>
        <h2>Extension convention</h2>
        <p>
          Future machine-readable overlays live under lowercase snake-case <code>structure.*</code> keys. Every new
          block receives a closed schema at a new soul-document version path. V2 rejects unknown blocks rather than
          allowing its semantics to drift after publication.
        </p>
        <h2>Identifier boundaries</h2>
        <p>
          The document's <code>agent_id</code> retains the v1 route-local selector rule. It is not Lesser's
          username-lexical <code>local_id</code>/<code>agent_username</code> and is not the separate
          <code>soul_agent_id</code>. Apply the naming vocabulary only when a workflow intentionally maps the same
          lexical value to a Ptah-created Lesser actor.
        </p>
        <div class="links">
          <a class="card" href="/contracts/panonomous/soul-document/v2/schema.json">
            <strong>Open schema JSON</strong>
            <span>JSON Schema 2020-12 for soul-document v2.</span>
          </a>
          <a class="card" href="/contracts/panonomous/soul-document/v2/examples/valid-published-five-bodies.json">
            <strong>Open valid example</strong>
            <span>A published document with five bodies and complete provenance.</span>
          </a>
          <a class="card" href="/contracts/panonomous/soul-document/">
            <strong>Back to soul-document v1</strong>
            <span>Review the immutable minimal core and its stable schema URL.</span>
          </a>
        </div>
      `),
    }),
  },
  {
    route: '/contracts/panonomous/agent-naming',
    mode: 'ssg',
    render: () => ({
      ...strictStaticPageSecurity,
      head: {
        title: 'Panonomous Agent Naming',
      },
      html: shell(`
        <span class="eyebrow">Naming Vocabulary v1</span>
        <h1>Ptah-created names must line up with Lesser's username rule.</h1>
        <p>
          This vocabulary distinguishes the local <code>agent_id</code> selector from the Lesser
          <code>agent_username</code>. For Project 48 M3 v1, Ptah-created agents that become or address Lesser actors
          use the same lexical validation profile as Lesser usernames.
        </p>
        <div class="notice">
          <strong>Stable vocabulary URL:</strong>
          <a href="/contracts/panonomous/agent-naming/v1/vocabulary.json"><code>/contracts/panonomous/agent-naming/v1/vocabulary.json</code></a>
        </div>
        <h2>Lesser-aligned rule</h2>
        <table>
          <thead>
            <tr><th>Term</th><th>Rule</th><th>Notes</th></tr>
          </thead>
          <tbody>
            <tr><td><code>agent_username</code></td><td><code>^[A-Za-z0-9_-]{1,30}$</code></td><td>Exact current Lesser username rule: ASCII letters, digits, underscore, hyphen; length 1–30.</td></tr>
            <tr><td><code>agent_id</code></td><td>same rule for Ptah-created Lesser actors</td><td>The local selector must not become a de-facto shape that Lesser later rejects.</td></tr>
            <tr><td><code>soul_agent_id</code></td><td>not this vocabulary</td><td>Separate Lesser Soul registry identifier; do not validate it as a username.</td></tr>
          </tbody>
        </table>
        <p>
          Generated names should prefer lowercase letters, digits, and hyphens for readability, but validators must not
          reject uppercase while Lesser accepts uppercase. Uniqueness and collision handling remain Lesser-authoritative.
        </p>
        <div class="links">
          <a class="card" href="/contracts/panonomous/agent-naming/v1/vocabulary.json">
            <strong>Open vocabulary JSON</strong>
            <span>Machine-readable v1 vocabulary and JSON Schema definitions.</span>
          </a>
          <a class="card" href="/contracts/panonomous/">
            <strong>Back to Panonomous contracts</strong>
            <span>Return to the contract index.</span>
          </a>
        </div>
      `),
    }),
  },
  {
    route: '/fep/agent-social-attribution',
    mode: 'ssg',
    render: () => ({
      ...strictStaticPageSecurity,
      head: {
        title: 'FEP Agent Social Attribution',
      },
      html: shell(`
        <span class="eyebrow">FEP Work</span>
        <h1>Agent Social Attribution</h1>
        <p>
          This site publishes the stable namespace required by the proposal and acts as the public home for the work.
          The runtime surface is intentionally static-first so the namespace path remains predictable and machine-safe.
        </p>
        <ul>
          <li><code>delegated_by</code> normalization has landed in <code>lesser</code>, so the proposal no longer depends on a known serialization gap.</li>
          <li>The namespace path is versioned so breaking changes move to a new URL instead of mutating <code>/v1</code>.</li>
          <li>CloudFront serves <code>/ns/*</code> without HTML rewrites or JavaScript redirects.</li>
        </ul>
        <div class="links">
          <a class="card" href="/ns/agent-attribution/v1">
            <strong>Resolve the live namespace</strong>
            <span>Fetch the JSON-LD context exactly as processors will consume it.</span>
          </a>
          <a class="card" href="/">
            <strong>Back to home</strong>
            <span>Return to the main Lesser Soul landing page.</span>
          </a>
        </div>
      `),
    }),
  },
  {
    route: '/404',
    mode: 'ssg',
    render: () => ({
      ...strictStaticPageSecurity,
      status: 404,
      head: {
        title: 'Not Found',
      },
      html: shell(`
        <span class="eyebrow">404</span>
        <h1>Nothing lives at this path.</h1>
        <p>
          The domain serves a small number of stable documents and pages. If you were looking for the namespace
          document, it lives at <code>/ns/agent-attribution/v1</code>.
        </p>
        <div class="links">
          <a class="card" href="/ns/agent-attribution/v1">
            <strong>Open the namespace document</strong>
            <span>Direct JSON-LD response with no redirect.</span>
          </a>
          <a class="card" href="/">
            <strong>Return home</strong>
            <span>Go back to the main landing page.</span>
          </a>
        </div>
      `),
    }),
  },
];
