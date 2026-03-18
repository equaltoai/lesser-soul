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

  p,
  li {
    font-size: 1.05rem;
    line-height: 1.7;
    color: var(--muted);
  }

  ul {
    padding-left: 1.25rem;
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
