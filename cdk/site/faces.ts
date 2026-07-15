import type { FaceModule } from '@theory-cloud/facetheory';

const shell = (content: string): string => `\
<main class="shell">
  <section class="hero">
    ${content}
  </section>
</main>
<style>
  :root {
    color-scheme: light;
    --bg: #f4efe5;
    --surface: rgba(255, 252, 246, 0.9);
    --ink: #1e1a17;
    --muted: #63584d;
    --accent: #a4451e;
    --accent-soft: rgba(164, 69, 30, 0.14);
    --line: rgba(30, 26, 23, 0.12);
  }

  * {
    box-sizing: border-box;
  }

  body {
    margin: 0;
    min-height: 100vh;
    font-family: "Iowan Old Style", "Palatino Linotype", "Book Antiqua", serif;
    color: var(--ink);
    background:
      radial-gradient(circle at top left, rgba(198, 104, 44, 0.24), transparent 28rem),
      radial-gradient(circle at bottom right, rgba(43, 89, 74, 0.18), transparent 24rem),
      linear-gradient(180deg, #fbf8f1 0%, var(--bg) 100%);
  }

  a {
    color: var(--accent);
  }

  .shell {
    width: min(72rem, calc(100vw - 2rem));
    margin: 0 auto;
    padding: 3rem 0 4rem;
  }

  .hero {
    background: var(--surface);
    border: 1px solid var(--line);
    border-radius: 1.5rem;
    padding: 2rem;
    box-shadow: 0 1.5rem 3rem rgba(30, 26, 23, 0.08);
  }

  .eyebrow {
    display: inline-block;
    margin-bottom: 1rem;
    padding: 0.35rem 0.7rem;
    border-radius: 999px;
    background: var(--accent-soft);
    color: var(--accent);
    font-size: 0.8rem;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }

  h1 {
    margin: 0 0 1rem;
    font-size: clamp(2.4rem, 6vw, 4.6rem);
    line-height: 0.95;
  }

  h2 {
    margin: 2rem 0 0.75rem;
    font-size: clamp(1.45rem, 3vw, 2rem);
  }

  h3 {
    margin: 1.5rem 0 0.5rem;
    font-size: 1.15rem;
  }

  p,
  li {
    font-size: 1.05rem;
    line-height: 1.7;
    color: var(--muted);
  }

  ul {
    padding-left: 1.25rem;
  }

  table {
    width: 100%;
    margin: 1rem 0;
    border-collapse: collapse;
    font-size: 0.98rem;
  }

  th,
  td {
    padding: 0.75rem;
    text-align: left;
    vertical-align: top;
    border-bottom: 1px solid var(--line);
  }

  th {
    color: var(--ink);
  }

  .notice {
    margin: 1rem 0;
    padding: 1rem;
    border: 1px solid var(--line);
    border-left: 0.35rem solid var(--accent);
    border-radius: 0.85rem;
    background: rgba(255, 255, 255, 0.48);
  }

  .links {
    display: grid;
    gap: 0.85rem;
    margin-top: 1.5rem;
  }

  .card {
    display: block;
    text-decoration: none;
    color: inherit;
    border: 1px solid var(--line);
    border-radius: 1rem;
    padding: 1rem 1.1rem;
    background: rgba(255, 255, 255, 0.58);
  }

  .card strong {
    display: block;
    color: var(--ink);
    margin-bottom: 0.35rem;
  }

  code {
    font-family: "SFMono-Regular", Consolas, "Liberation Mono", monospace;
    font-size: 0.92em;
    background: rgba(30, 26, 23, 0.08);
    padding: 0.14rem 0.35rem;
    border-radius: 0.35rem;
  }

  @media (max-width: 640px) {
    .shell {
      width: min(72rem, calc(100vw - 1rem));
      padding-top: 1rem;
    }

    .hero {
      padding: 1.25rem;
      border-radius: 1rem;
    }
  }
</style>`;

export const faces: FaceModule[] = [
  {
    route: '/',
    mode: 'ssg',
    render: () => ({
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
            <span>Required body text, optional summary, and server-assigned draft version semantics.</span>
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
        <div class="links">
          <a class="card" href="/contracts/panonomous/soul-document/v1/schema.json">
            <strong>Open schema JSON</strong>
            <span>Machine-readable v1 schema for draft soul documents.</span>
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
    route: '/contracts/panonomous/agent-naming',
    mode: 'ssg',
    render: () => ({
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
