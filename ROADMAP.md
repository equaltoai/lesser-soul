# lesser-soul Roadmap (Soul Registry v1)

Date: 2026-02-21  
Audience: EqualtoAI stack maintainers

This roadmap turns `SPEC.md` into implementation milestones with concrete acceptance criteria.

## Hard requirements (non-negotiable)

- **Use AppTheory** for service runtime and infra patterns (where new cloud resources are required).
- **Use TableTheory for all data** persisted in DynamoDB (no raw DynamoDB client usage for application state).
- **No new databases** (all structured state stays in the existing state table model; artifacts/snapshots live in S3 as specified).
- **Safe-first governance** for privileged on-chain writes (same pattern as tip registry).

## Scope and repo boundaries (align first)

There are two “soul” surfaces in the workspace today:

1. **Registry layer** (this spec): on-chain identity anchor + off-chain identity/reputation/validation/discovery.
2. **Instance-side soul routing** (`/soul/*`) already exists in `lesser/infra/cdk` and is referenced by `lesser-body`.

This roadmap implements the **registry layer** (primarily in `lesser-host`) and its integration points used by
instances/`lesser-body` (registration updates, config discovery, and any required SSM exports).

Acceptance gate for this section:
- [ ] A short ADR exists that states where each component lives:
  - registry API + data (control plane),
  - contracts repo location,
  - any scheduled jobs/workers,
  - how `/soul/*` instance routing relates to `/api/v1/soul/*` control-plane APIs.

---

## Milestones

### M0 — Canonical IDs, signatures, and storage plan (ADR + test vectors)

Goal: eliminate ambiguity before writing contracts/APIs.

Deliverables:
- **ADR** covering:
  - `agentId` derivation (domain normalization + local ID) and **shared test vectors**.
  - `tokenId` policy (recommended: `tokenId == agentId` for determinism and 1:1 mapping).
  - Registration file canonicalization + signing approach:
    - what exactly is signed (full JSON? a canonical hash?),
    - signature scheme(s): EIP-191 vs EIP-712 (rotation uses EIP-712 per spec).
  - Ownership model:
    - who can mint (admin Safe / contract owner),
    - who can update `metaURI`,
    - suspend/reinstate semantics and their on-chain vs off-chain effects.
  - DynamoDB access rule: **TableTheory only** for persisted records.
- **Data model delta** vs existing `lesser-host` state table:
  - which new PK/SK patterns are added (from spec),
  - which GSIs (if any) are required for `agents/mine` and search.

Acceptance criteria:
- [ ] Go + Solidity `agentId` derivation functions match the same fixtures.
- [ ] Every new persisted record has a TableTheory model and key format documented.
- [ ] Search and “mine” are implementable without full table scans (documented index strategy).

---

### M1 — Smart contracts (Soul registry + attestations) with Base deployment plan

Goal: ship the on-chain anchor and root attestation surfaces with production-grade tests.

Deliverables:
- `ISoulRegistry` implementation:
  - ERC-721 minting (1 soul per agent), `getAgentWallet(agentId)` (EIP-8004 compatibility).
  - Non-transferability (“soulbound”) after the defined claim window.
  - `setMetaURI(agentId, metaURI)` authorization per ADR.
  - Wallet rotation (two-step propose/confirm with on-chain signature verification).
- Attestation contracts:
  - `IReputationAttestation.publishRoot(...)` + `latestRoot()`.
  - `IValidationAttestation.publishRoot(...)` + `latestRoot()`.
- Hardhat build + artifacts + tests:
  - soulbound transfer restrictions,
  - one-soul-per-agent invariant,
  - rotation signature verification,
  - `IERC8004IdentityRegistry` compatibility tests (same call shape as TipSplitter uses).
- Deployment runbook for Base (8453):
  - owner = admin Safe,
  - non-upgradeable policy,
  - address handoff procedure for TipSplitter `setAgentIdentityRegistry(address)`.

Acceptance criteria:
- [ ] Contracts compile with Solidity `^0.8.24` and pass unit tests.
- [ ] A deployment script/runbook exists for Base (8453) with Safe ownership.
- [ ] TipSplitter can resolve wallets via `getAgentWallet(agentId)` against the deployed registry (smoke test).

---

### M2 — TableTheory models + S3 layouts for soul state and artifacts

Goal: implement the off-chain persistence layer in a way that supports API, search, audits, and future jobs.

Deliverables (in `lesser-host` unless ADR chooses otherwise):
- TableTheory models for:
  - Agent identity (`SOUL#AGENT#{agentId}` / `IDENTITY`)
  - Reputation (`SOUL#AGENT#{agentId}` / `REPUTATION`)
  - Validation records (`SOUL#AGENT#{agentId}` / `VALIDATION#...`)
  - Operations (`SOUL#OP#{operationId}` / `OPERATION`)
  - Endorsements (`SOUL#AGENT#{agentId}` / `ENDORSEMENT#...`)
- Index strategy for:
  - `GET /api/v1/soul/agents/mine` (owner→agents lookup),
  - `GET /api/v1/soul/search` (domain/capability/name query without scans).
- S3 key conventions inside the existing Soul pack bucket:
  - registration files (current),
  - reputation snapshots (historical),
  - validation snapshot packs (if stored),
  - signatures (KMS-signed where required).

Acceptance criteria:
- [ ] All soul records are read/written via TableTheory models (no raw DynamoDB PutItem/Query in app code).
- [ ] “Mine” and search can be implemented with Query operations (no table scans in the hot path).
- [ ] S3 objects required by the spec have deterministic key formats and lifecycle expectations documented.

---

### M3 — Identity registration API (begin/verify) + Safe-ready mint operations

Goal: ship the core registration flow end-to-end (public proofs + portal UX prerequisites).

Deliverables:
- API routes added to the control plane (AppTheory handlers):
  - `POST /api/v1/soul/agents/register/begin`
  - `POST /api/v1/soul/agents/register/{id}/verify`
  - `GET  /api/v1/soul/config`
- Proof verification:
  - DNS TXT `_lesser-soul-agent.<domain>` with `lesser-soul-agent=<token>`.
  - HTTPS well-known `https://<domain>/.well-known/lesser-soul-agent` with the same value.
  - Instance membership check (per ADR: instance registry record or trust attestation).
- Wallet signature verification over the registration challenge message.
- Mint operation creation:
  - generate `{ to, value, data }` for `mintSoul(to, agentId, metaURI)`,
  - store as an operation record (pending) using TableTheory,
  - create initial Agent identity row (status `pending` → `active` on execution record).

Acceptance criteria:
- [ ] A customer can complete begin→verify and receive a Safe-ready mint payload.
- [ ] Invalid proofs/signatures fail closed with actionable error messages.
- [ ] The flow is idempotent for retries (no duplicate souls minted; same agentId yields same pending operation behavior).
- [ ] Unit tests cover: proof parsing, signature verification, agentId derivation, and operation persistence.

---

### M4 — Wallet rotation, registration updates, and suspension controls

Goal: implement lifecycle management post-mint and the endpoints required by `lesser-body`.

Deliverables:
- Portal endpoints:
  - `POST /api/v1/soul/agents/{agentId}/rotate-wallet/begin`
  - `POST /api/v1/soul/agents/{agentId}/rotate-wallet/confirm`
  - `POST /api/v1/soul/agents/{agentId}/update-registration`
- Admin endpoints:
  - `POST /api/v1/soul/agents/{agentId}/suspend`
  - `POST /api/v1/soul/agents/{agentId}/reinstate`
  - `GET/POST /api/v1/soul/operations...` parity with the tip-registry operations UX needs.
- Registration file publishing rules:
  - stored in S3 (current registration),
  - `metaURI` updates if required by the chosen addressing model,
  - re-signing rules (who signs what, how signatures are checked).

Acceptance criteria:
- [ ] Rotation requires both signatures and results in an on-chain operation payload.
- [ ] `update-registration` supports setting `endpoints.mcp` for `lesser-body` without breaking signature validity.
- [ ] Suspended agents are blocked from reputation accrual and discovery responses (policy defined in ADR).
- [ ] All writes are authenticated and audited (same standard as other portal/admin handlers).

---

### M5 — Public read APIs + discovery/search v1

Goal: make souls queryable by the public and consumable by instance clients without CORS/CSP regressions.

Deliverables:
- Public endpoints:
  - `GET /api/v1/soul/agents/{agentId}`
  - `GET /api/v1/soul/agents/{agentId}/registration`
  - `GET /api/v1/soul/agents/{agentId}/reputation`
  - `GET /api/v1/soul/agents/{agentId}/validations` (paginated)
  - `GET /api/v1/soul/search?q=...&capability=...`
- Cache policy + headers (explicit):
  - registration and config can be cacheable with sane TTLs,
  - anything auth-scoped remains non-cacheable.

Acceptance criteria:
- [ ] Public read endpoints work without authentication and return stable, versioned JSON.
- [ ] Search supports at least: domain match, localId match, capability filter, and returns deterministic paging.
- [ ] Suspended agents are clearly represented (or excluded) per policy.

---

### M6 — Reputation v0 (tips + basic signals) + computation job

Goal: ship a first real reputation score and its storage without overfitting early.

Deliverables:
- Signal ingestion:
  - tips received (on-chain TipSplitter events) as the first on-chain economic signal,
  - placeholder stubs for future signals (attestations, endorsements, flags).
- Aggregation job:
  - configurable weights (registry-level, not per-agent),
  - writes current score to DynamoDB via TableTheory,
  - writes snapshots to S3 for audit trail.
- API responses include:
  - composite + dimensions + signal counts as in the spec.

Acceptance criteria:
- [ ] Reputation recomputation is deterministic given the same input set (job is repeatable).
- [ ] Tables contain one current reputation record per agent and snapshots are persisted in S3.
- [ ] Unit tests cover aggregation math and storage writes; integration test covers end-to-end recompute on a small fixture set.

---

### M7 — Validation v0 (challenge records + progressive scoring)

Goal: implement validation storage and scoring so it can feed reputation.

Deliverables:
- Validation record model + APIs:
  - write path(s) for challenge issued / response / evaluation result (even if initially admin/system-only),
  - `GET /api/v1/soul/agents/{agentId}/validations` (paged history).
- Progressive scoring:
  - pass/fail/timeout effects,
  - decay policy per epoch (configurable),
  - summary fields usable by reputation.

Acceptance criteria:
- [ ] Validation history is queryable and correctly ordered/paged.
- [ ] Score changes match the documented rules and are covered by tests (including decay).
- [ ] Reputation reads reflect validation contribution (dimension present even if weighted low).

---

### M8 — Merkle root attestations (reputation + validation) + on-chain publishing

Goal: make off-chain claims verifiable by anchoring signed snapshots on-chain.

Deliverables:
- Merkle tree builder:
  - deterministic leaf encoding for each agent record,
  - proof generation for a specific agent record.
- Publish pipeline:
  - batch snapshot at a blockRef,
  - store pack artifacts in S3 (tree/proofs + manifest),
  - publish root on-chain (Safe-ready operation),
  - record execution receipt.
- Admin endpoints:
  - `POST /api/v1/soul/reputation/publish`
  - `POST /api/v1/soul/validation/publish`

Acceptance criteria:
- [ ] Given an agent record + proof, an independent verifier can validate inclusion against `latestRoot`.
- [ ] Publishing is repeatable and fails closed on inconsistencies (count mismatch, invalid leaf set, etc.).
- [ ] On-chain publish operations follow the same operational model as tip-registry ops (pending→executed with receipts).

---

### M9 — Portal UX (lesser.host web) for souls

Goal: expose soul management to customers/operators with a usable, audit-friendly UI.

Deliverables:
- Portal:
  - **My Agents** list (status + wallet + reputation summary),
  - **Register Agent** guided flow (proofs + signature + submission),
  - **Agent Detail** (registration, reputation breakdown, validation history, endorsements),
  - **Wallet Rotation** flow.
- Operator surfaces:
  - pending operations review (mints/rotations/publishes),
  - execution recording UI parity with tip registry ops.

Acceptance criteria:
- [ ] A customer can register an agent and see status move from pending→active after execution is recorded.
- [ ] Error states are actionable (which proof failed, signature invalid, etc.).
- [ ] Operator can reconcile operations without direct DB writes.

---

### M10 — Instance + lesser-body integration hardening (no coupling to unfinished code)

Goal: ensure the registry integrates cleanly with instance deployments and `lesser-body`, without blocking ongoing work.

Deliverables:
- SSM exports required by instances (as finalized in M0):
  - registry contract addresses (stage-scoped),
  - any origin routing values (if required for single-origin serving).
- A stable contract for `lesser-body`:
  - `update-registration` semantics for setting MCP endpoint + capabilities,
  - signature expectations and authentication method.
- Docs/runbook:
  - how an instance with `soulEnabled=true` discovers required registry/origin values,
  - minimal “smoke test” steps for a new instance.

Acceptance criteria:
- [ ] An instance can enable soul routing without any manual URL hardcoding.
- [ ] `lesser-body` can update MCP endpoint in the registration file via the registry API (auth + signature verified).
- [ ] No breaking changes are required in `lesser-body` beyond consuming the documented endpoint contract.

---

## vNext (explicitly deferred unless pulled into a milestone)

- Peer endorsements end-to-end (UX + signature verification + weighting).
- Flags/reports ingestion and moderation-driven reputation penalties.
- Cross-instance discovery federation (beyond simple search).
- More granular capability taxonomy and per-capability validation ladders.

