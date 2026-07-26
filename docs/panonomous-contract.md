# Panonomous contract

Status: public static contract lineage for Project 48 M3 issues `#16` and `#17`, extended by the operator-settled
soul-document v2 design in `#28`.

Stable public artifacts after site build and operator deployment:

- `https://spec.lessersoul.ai/contracts/panonomous/soul-document/`
- `https://spec.lessersoul.ai/contracts/panonomous/soul-document/v1/schema.json`
- `https://spec.lessersoul.ai/contracts/panonomous/soul-document/v2/`
- `https://spec.lessersoul.ai/contracts/panonomous/soul-document/v2/schema.json`
- `https://spec.lessersoul.ai/contracts/panonomous/soul-document/v2/examples/valid-published-five-bodies.json`
- `https://spec.lessersoul.ai/contracts/panonomous/agent-naming/`
- `https://spec.lessersoul.ai/contracts/panonomous/agent-naming/v1/vocabulary.json`

These paths do not alter `https://spec.lessersoul.ai/ns/agent-attribution/v1`. Soul-document v1 is also immutable:
v2 is a parallel version path, and the v1 schema remains byte-identical.

## Source evidence inspected

- `contracts/panonomous/soul-document/v1/schema.json`: the minimal AgentSoul-compatible core, required
  `agent_id`/`body`, optional `summary`, server fields, draft-only lifecycle, and UTF-8 byte-bound annotations.
- `contracts/panonomous/agent-naming/v1/vocabulary.json`: the route-local `agent_id` vocabulary, Lesser's exact
  username-lexical rule, and the distinction from `soul_agent_id`.
- `cdk/site/build.ts`: FaceTheory writes the SSG output, then recursively copies the repository `contracts/` tree to
  `cdk/dist/site/contracts/`.
- `cdk/site/faces.ts`: the existing Panonomous contract index, soul-document reference page, and agent-naming page
  establish the site documentation pattern mirrored by the v2 page.
- `docs/panonomous-contract.md` at the v1 baseline: Project 48 M3 `#16`/`#17` lineage and the original source evidence
  remain the historical basis for the core contract.
- `equaltoai/lesser-soul#28`: operator-settled decisions dated 2026-07-26 for layering, provenance, published
  lifecycle semantics, and the lesser-body/Ptah application boundary.

The M3 evidence underlying v1 remains:

- AgentSoul contract evidence in local `theory-mcp-server` source:
  - `internal/mcp/tools/agent_soul_tools.go` exposes `agent_soul_get` and `agent_soul_upsert` with `agent_id`, required
    `body`, optional `summary`, draft-only behavior, and size-bounded schemas.
  - `internal/models/agent_soul.go` defines the stored draft record: `client_namespace`, `agent_id`, `tenant_id`,
    `soul_version`, `lifecycle_state`, `body`, optional `summary`, `updated_by_subject_id`, timestamps, and `version`.
  - `internal/content/soul_service.go` increments `soul_version` and storage `version` on each save, preserves creation
    time, requires active child agents, and returns draft records only through `GetDraft`.
  - `SPEC.md` describes `agent_soul_get`/`agent_soul_upsert` as draft-only, non-memory, non-publishing tools.
- Lesser naming evidence in local `lesser` source:
  - `pkg/common/validation.go` defines `UsernamePattern = ^[a-zA-Z0-9_-]{1,30}$` and `ValidateUsername`.
  - `ValidateUsernameParamID` delegates to `ValidateUsername`.
  - `cmd/api/handlers/agents.go` uses `ValidateUsernameParamID` for `agent_username` in the delegate path.
  - PR `#1235` / commit `29f70f30854fb4c5df7ee3b9bc18a187a93b4b85` documents the M2/L2 delegate contract in
    `docs/contracts/agent-delegate-api.md`.
- The original lesser-body read-only inspection found no exact `agent_soul_*`, `AgentSoul`, `Panonomous`, or
  `soul-document` provisional docs/usages in that checkout. V2 still does not edit lesser-body; adoption is a sibling
  issue.

## V1: immutable minimal core

V1 follows the minimal proven AgentSoul shape: required Markdown-friendly `body`, optional non-blank `summary`,
server-assigned monotonic `soul_version`, draft lifecycle, attribution to the authenticated subject, timestamps, and
optimistic storage `version`. V1 intentionally defines no machine-readable body sections; headings inside `body` are
plain text.

V2 does not amend that schema. Consumers that implement only v1 continue to validate the same bytes at the same URL.

## V2: layered structured declaration

V2 keeps the v1 core field names and bounds, changes the schema marker to
`lessersoul.panonomous.soul-document.v2`, expands lifecycle values, and adds two optional overlays:

- `structure.five_bodies`
- `provenance`

The canonical Markdown `body` remains required and authoritative. A document with only the v1-compatible core can be a
valid v2 document; machine-readable sections are not a publication prerequisite. When `structure` is present in v2 it
contains a complete `five_bodies` block.

### Settled design decisions (2026-07-26)

1. **Layered v2:** v1 stays byte-identical and immutable; v2 adds optional machine-readable structure.
2. **Provenance block:** v2 defines mint provenance as a closed optional block. When present, all provenance fields are
   required so partial provenance is not mistaken for complete evidence.
3. **Published lifecycle now:** downstream materialization gates on `lifecycle_state = "published"`.
4. **Deterministic Ptah-side declaration application:** the transformation is implemented in lesser-body under a
   sibling issue. It is not a runtime or API concern in lesser-soul.

### Lifecycle semantics

`lifecycle_state` accepts:

- `draft`: mutable authoring state.
- `published`: reached only through an explicit owner act; the resulting document is an immutable snapshot.
- `archived`: retires a published snapshot from rendering without rewriting its published content.

JSON Schema validates the state value, not state-transition history. A materializer must require `published`; it must
not treat a missing state, `draft`, or `archived` as renderable.

### `structure.five_bodies`

The optional structured overlay faithfully carries the hosted-genesis declaration:

- `identity`, `philosophy`, `discipline`, and `boundaries` each require `summary: string` and optionally carry
  `notes: string[]`.
- `soul` requires `summary: string` and `refusals`, optionally carries `notes: string[]`, and requires at least one
  refusal.
- Every refusal is a closed triple of `bypass`, `invariant`, and `closestSafePath` strings.
- The five-body object and every nested object use `additionalProperties: false`.

### Provenance

The optional `provenance` object is closed and complete when present:

- `declaration_schema_version`
- `declaration_candidate_hash`, matching `^sha256:[0-9a-f]{64}$`
- `registration_id`
- `conversation_id`
- `model`
- `source`, one of `host_genesis_finalize`, `ptah_seed`, or `owner`

Optionality preserves the layered design for core-only and pre-provenance drafts. A downstream policy may require
provenance in addition to the schema's minimum, but it must not accept a partial provenance object.

### Extension convention

`structure` is the only extension container. Block names use lowercase `snake_case`, for example
`structure.five_bodies`. Every future `structure.<block_name>` must:

1. be declared explicitly with a closed nested schema;
2. preserve the canonical `body` as the authoritative human-readable representation;
3. land at a new soul-document version path before publication.

Because v2 uses `additionalProperties: false`, unknown structure blocks are invalid under v2. This is intentional:
future blocks do not accrete silently into a frozen schema. A new contract version can recognize the new block while
v1 and v2 remain stable.

## Identifier vocabulary

The soul document's `agent_id` retains v1's route-local AgentSoul selector shape: 1–128 characters and no `|`, `=`, or
`/`. Do not conflate it with:

- Lesser's username-lexical `local_id` / `agent_username`, which uses
  `^[A-Za-z0-9_-]{1,30}$`; or
- `soul_agent_id`, the separate Lesser Soul registry identifier.

When a Ptah-created Lesser actor intentionally maps the same lexical value into `agent_id` and
`local_id`/`agent_username`, apply the stricter agent-naming v1 vocabulary before Lesser's authoritative validation.

## Examples and validation

V2 publishes:

- `contracts/panonomous/soul-document/v2/examples/valid-published-five-bodies.json`
- `contracts/panonomous/soul-document/v2/examples/invalid-missing-safe-path.json`
- `contracts/panonomous/soul-document/v2/examples/invalid-provenance-hash.json`

The schema is compiled as JSON Schema 2020-12 and the fixtures are checked with AJV CLI. The valid fixture must pass;
each invalid fixture must be rejected for its named invariant.

```bash
npx --yes --package ajv-cli@5.0.0 --package ajv-formats@3.0.1 \
  ajv compile --spec=draft2020 --strict=false -c ajv-formats \
  -s contracts/panonomous/soul-document/v2/schema.json

npx --yes --package ajv-cli@5.0.0 --package ajv-formats@3.0.1 \
  ajv validate --spec=draft2020 --strict=false --all-errors -c ajv-formats \
  -s contracts/panonomous/soul-document/v2/schema.json \
  -d contracts/panonomous/soul-document/v2/examples/valid-published-five-bodies.json

for fixture in \
  contracts/panonomous/soul-document/v2/examples/invalid-missing-safe-path.json \
  contracts/panonomous/soul-document/v2/examples/invalid-provenance-hash.json; do
  if npx --yes --package ajv-cli@5.0.0 --package ajv-formats@3.0.1 \
    ajv validate --spec=draft2020 --strict=false --all-errors -c ajv-formats \
    -s contracts/panonomous/soul-document/v2/schema.json -d "$fixture"; then
    echo "invalid fixture unexpectedly passed: $fixture" >&2
    exit 1
  fi
done
```

The site-copy gate is:

```bash
cd cdk
npm run build:site
cmp ../contracts/panonomous/soul-document/v2/schema.json \
  dist/site/contracts/panonomous/soul-document/v2/schema.json
```

## Lesser-body adoption handoff

The sibling lesser-body issue must reference the exact v2 schema URL and the following settled requirements:

- validate incoming/upserted documents against soul-document v2;
- materialize or render only `lifecycle_state = "published"`;
- preserve published snapshots as immutable and exclude `archived` snapshots from rendering;
- deterministically transform hosted-genesis declarations into `structure.five_bodies` on the Ptah side, with no
  MicroVM or LLM in the application path;
- preserve and validate the complete provenance block when supplied;
- retire the provisional opaque marker only after validated v2 upserts are active; and
- preserve the identifier distinctions described above.

That implementation remains outside lesser-soul. This repository publishes static schema, examples, and documentation
only; it adds no runtime/API surface and does not deploy this change.
