---
name: evolve-namespace
description: Use when a change touches JSON-LD namespace content at `/ns/agent-attribution/v1` (or any existing version path), or proposes a new version path (`/v2`, `/v3`). The namespace is a frozen-forever public contract; this skill enforces that contract and guides version-path addition when new semantics are needed.
---

# Evolve the namespace

The JSON-LD namespace at `https://spec.lessersoul.ai/ns/agent-attribution/v1` is **the** public contract soul publishes. Every ActivityPub implementation that processes agent-attribution properties in federated activities may resolve this URL to expand JSON-LD context. The URL must resolve forever, with stable semantics within a version.

This skill walks every namespace-touching change with the rigor that contract demands.

## The namespace architecture (memorize)

- **Location in repo**: `cdk/site/static/ns/agent-attribution/v<N>`
- **Served at**: `https://spec.lessersoul.ai/ns/agent-attribution/v<N>`
- **Content type**: `application/ld+json`
- **Caching**: `Cache-Control: max-age=31536000, immutable` — clients cache aggressively, expecting content never changes
- **CORS**: open — any origin can resolve the namespace
- **CloudFront behavior**: direct S3 pass-through; no HTML rewrite, no JavaScript, no redirect
- **S3 bucket**: namespace bucket (separate from site bucket) with `RemovalPolicy.RETAIN` — survives stack teardown
- **Versioning**: breaking changes → new path (`/v2`, `/v3`); old paths never mutate

## When this skill runs

Invoke this skill when:

- A change adds new properties, new classes, or new context definitions to the namespace
- A change proposes to "clarify" or "correct" existing namespace content
- A change adds a new version path (`/v2`)
- A change removes or deprecates an existing version
- A change affects how the namespace is served (content-type, caching, CORS, CloudFront behavior)
- A peer ActivityPub implementation reports a namespace-resolution issue that may warrant content review
- An FEP editorial round requires namespace content adjustment

## Preconditions

- **The change is described concretely.** "Update the namespace" is too vague; "add a new class `EmailChannel` to `/v2` with properties `address` (xsd:string), `verified` (xsd:boolean), extending the existing `Channel` class, for use in agent-attribution serialization when the attributed agent has verified-email channels" is concrete.
- **MCP tools healthy**, `memory_recent` first — namespace evolution decisions accumulate.
- **FEP alignment** — if this change tracks an FEP, the FEP's current editorial state is known (via `manage-fep-submission` or direct review).

## The four-dimension walk

### Dimension 1: Mutation classification

Classify the proposed change against the frozen-forever contract:

- **Addition at a new version path** (`/v2`, `/v3`) — welcome. New version paths are parallel to existing ones; the old path continues to resolve for peers that haven't migrated. This is the canonical way to evolve semantics.
- **Additive within an existing version** — acceptable only when **JSON-LD rules guarantee no semantic change for existing consumers**. Adding a new term to a JSON-LD context typically introduces the term as defined going forward but does not change the meaning of existing serializations. Even so, the rule of thumb is: **if in doubt, move to a new version path.** Version bumps cost almost nothing; semantic drift is expensive.
- **Correction of an unambiguous factual error** — possible but carefully scoped. If `/v1` contains a literal typo (e.g. a misspelled property name) that no consumer could reasonably interpret correctly, a correction may be warranted. Document the correction's scope in the commit body; coordinate with affected consumers where possible.
- **Semantic mutation** — refused. Changing the meaning of an existing property, removing a property, or restructuring the context in ways existing consumers would observe is refused. Propose a new version path instead.
- **Formatting change** (whitespace, JSON indentation, key ordering) — generally avoid. Even formatting changes can affect clients that compare context documents by hash. If a formatting change is genuinely needed (e.g. tool-driven canonicalization), document it explicitly and flag the cache-invalidation implication.

### Dimension 2: JSON-LD discipline

For every change to namespace content:

- **JSON-LD context semantics** — the document is a `@context` object (or a document containing one). Changes respect JSON-LD's rules: `@id` values are IRIs; `@type` values are IRIs; term definitions include `@id` and `@type` per the target semantic.
- **Term naming** — snake_case is the convention consistent with existing namespace content. New terms follow the convention.
- **Prefix usage** — prefixes defined in the context. Adding a new prefix requires adding the prefix URL and documenting what it references.
- **External vocabulary reuse** — where applicable, reuse terms from established ActivityStreams / ActivityPub vocabularies rather than inventing new ones. Novel terms should be justified.
- **Versioning metadata** — if the namespace document has version / publisher / license metadata fields, they stay current.

### Dimension 3: Consumer-impact analysis

For each proposed change:

- **Who consumes the namespace?** Lesser instances (every running lesser instance serializes the namespace URL). External ActivityPub peers (Mastodon, Pleroma, Misskey, GoToSocial, and others, especially as FEP adoption spreads). Internal equaltoai repos (body for soul-binding; host for registry-semantics alignment).
- **Backward-compat check** — for a new-version-path addition, the old version continues to serve unchanged. For any other change (correction, formatting), the existing-consumer impact is analyzed explicitly.
- **Forward-compat expectation** — when `/v2` is published, how do existing `/v1` consumers migrate? Typically they don't need to until the semantic change matters for them; `/v1` continues to resolve.
- **Host implementation alignment** — host's registry implementation must be aligned with the namespace semantics. A `/v2` namespace publication usually coordinates with host's corresponding implementation update.

### Dimension 4: Serving contract preservation

For changes that affect how the namespace is served (not content but topology):

- **CloudFront behavior** for `/ns/*` — direct S3 pass-through; no HTML rewrite; no JavaScript; no redirect. Changes here would break direct JSON-LD resolution.
- **Response headers**: `Content-Type: application/ld+json`, `Cache-Control: max-age=31536000, immutable`, `Access-Control-Allow-Origin: *`. Loosening any of these (e.g. shorter cache, restricted CORS) is refused.
- **S3 bucket `RemovalPolicy.RETAIN`** preserved — the namespace bucket never deletes on stack teardown.
- **HTTPS-only** — no plaintext HTTP serving of namespace content.

## The audit output

```markdown
## Namespace-evolution audit: <change name>

### Proposed change
<concrete description>

### Version path affected
<existing /v1 / existing /v<N> / new /v<N+1>>

### Mutation classification
<addition at new version path / additive within existing version (JSON-LD-safe) / correction of unambiguous error / semantic mutation (refuse) / formatting change (avoid or document)>

### JSON-LD discipline check
- `@id` IRIs sound: <confirmed>
- Term naming convention (snake_case): <confirmed>
- Prefix usage correct: <confirmed>
- External vocabulary reuse considered: <yes — term X reused from ActivityStreams / new term justified because ...>
- Versioning metadata current: <confirmed>

### Consumer-impact analysis
- Lesser instances: <no change (existing version continues) / migrate to /v<new> when ready / correction affects all>
- External ActivityPub peers: <...>
- host registry implementation alignment: <coordinated / independent>
- body soul-binding resolution impact: <none / coordinated>

### Backward compatibility
- Existing consumers continue to work: <yes by default — old version path unchanged>
- Forward migration guidance for new-version consumers: <...>

### Serving-contract preservation
- CloudFront `/ns/*` behavior: <preserved>
- Response headers (content-type, cache, CORS, HTTPS): <preserved>
- S3 bucket `RemovalPolicy.RETAIN`: <preserved>

### FEP alignment (if applicable)
- FEP in editorial review: <yes / no; reference roadmap issue>
- Namespace change aligns with current FEP state: <confirmed / pending editorial response>

### Cross-repo coordination
- host: <required (implementation alignment) / not required>
- lesser: <required (serialization change) / not required — typical for version-path additions>
- body: <required / not required>

### Test coverage
- Namespace document syntactically valid JSON-LD: <verified with a JSON-LD parser>
- CloudFront behavior verified in lab: <planned in deploy-namespace-site>
- End-to-end resolution test (curl from an external network): <planned>

### Proposed next skill
<implement-milestone if audit clean; manage-fep-submission if FEP alignment needed; scope-need if audit surfaces scope growth or implementation creep; investigate-issue if audit reveals an existing bug>
```

## Refusal cases

- **"Correct a typo in `/v1`; it's just a typo."** Evaluate carefully. Unambiguous errors may be correctable; anything semantic is a new version path.
- **"Add a property to `/v1`; it's additive."** Prefer new version path. The "additive is safe" argument is correct for JSON-LD in many cases, but the rule of thumb is new-version-path to avoid any interpretation drift.
- **"Rename a property in `/v1` for consistency."** Refuse. Renames are semantic changes.
- **"Remove a deprecated property from `/v1`."** Refuse.
- **"Update the publisher URL in the metadata."** Carefully — metadata changes that don't affect JSON-LD semantics may be acceptable, but loose-interpretation clients may treat the document hash; evaluate.
- **"Change content-type to `application/json` for broader client compatibility."** Refuse. `application/ld+json` is required for JSON-LD semantics; clients that don't set `Accept: application/ld+json` should be updated to do so.
- **"Relax cache-control to `max-age=3600` so we can push corrections faster."** Refuse. The frozen-forever contract means corrections are versioned, not pushed.
- **"Remove CORS open; it's an attack surface."** Refuse. Open CORS on static JSON-LD is the correct posture for a public namespace.
- **"Add an HTML wrapper around `/ns/*` so humans can read it."** Refuse. Direct JSON-LD pass-through is the contract. Host human-readable documentation at separate paths.
- **"Redirect `/v1` to a better URL."** Refuse. The URL is the forever-promise.
- **"Weaken `RemovalPolicy.RETAIN` on the namespace bucket; it's cluttering Terraform state."** Refuse.
- **"Silently deprecate `/v1` so all traffic moves to `/v2`."** Refuse. Old versions serve as long as peers resolve them.

## Persist

Append every meaningful namespace event — a version-path publication, a correction with scope, a coordinated host-alignment decision, a FEP-editorial-driven adjustment. These are high-signal memory material: the historical record of what soul has published and why is part of the public-contract stewardship's paper trail. Include: date, version affected, change summary, FEP reference if applicable, cross-repo coordination outcome.

Five meaningful entries is a floor for namespace-evolution work; publications are inherently memorable.

## Handoff

- **Audit clean, additive at new version path** — invoke `implement-milestone`. `deploy-namespace-site` handles the deploy of the new path.
- **Audit clean, documented correction at existing version** — invoke `implement-milestone` with the correction's scope and rationale documented in the commit body; affected consumers surfaced where known.
- **Audit requires FEP coordination** — invoke `manage-fep-submission` first to align editorial state, then back here.
- **Audit requires host implementation alignment** — coordinate via the `host` steward through the user before proceeding.
- **Audit reveals semantic mutation proposal** — refuse; redirect to new version path or scope conversation.
- **Audit surfaces framework awkwardness** (FaceTheory / CDK / CloudFront Functions patterns for serving JSON-LD cleanly) — `coordinate-framework-feedback`.
