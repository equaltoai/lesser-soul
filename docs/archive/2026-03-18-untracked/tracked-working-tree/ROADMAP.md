# lesser-soul v2.0 Implementation Roadmap

> For the stack-wide v3 roadmap (channels + ENS + communication gateway), see `ROADMAP-v3.md`.

Date: 2026-03-02
Audience: EqualtoAI stack maintainers

## Context

The new SPEC.md v2.0 (1,799 lines) is written and approved. It transforms lesser-soul from a flat registry into a rich
self-definition protocol with boundaries, structured capabilities, sovereignty primitives, relationships, continuity,
and two-phase minting. The v1 implementation in lesser-host is substantial (~20 handlers, 10 models, 3 contracts,
reputation worker, frontend client) and fully functional on Sepolia. This roadmap turns v2 into implementation.

**Execution model:** Milestones are scoped for delegation to Codex (`gpt-5.2`). Each milestone is independently
testable, follows existing patterns, and can be described in a focused prompt.

**Where code lives:** All implementation in `lesser-host` (backend Go, contracts Solidity, frontend Svelte/TS).
`lesser-soul` repo holds the spec and this roadmap only.

## Hard requirements (carried from v1)

- **TableTheory** for all DynamoDB state (no raw PutItem/Query)
- **AppTheory** for service runtime and HTTP handlers
- **No new databases** (existing state table + S3 soul registry artifacts bucket; historical name: "soul-packs")
- **Safe-first governance** for privileged on-chain writes
- **Lean on AppTheory and TableTheory as much as possible** — these frameworks are maintained in-house to make generative-code development easier. Prefer their abstractions over hand-rolled patterns for models, handlers, routing, and persistence.

---

## Dependency graph

```
M1 (Models) ───────┬──────────┬──────────┬──────────┬───────────┐
                    │          │          │          │           │
                    v          v          v          v           │
                 M2 (RegFile) M3 (Contracts) M4 (Boundaries)   │
                    │                        │                  │
                    ├──────────┐             v                  v
                    v          v        M6 (Sovereignty)   M8 (Relationships)
              M5 (Caps)  M7 (Continuity)                       │
                    │          │                               │
                    v          v                               │
              M10 (Mint Conv) M9 (Death)                      │
                                                               │
    M4+M8+M1 ──────────────────> M11 (Rep Worker + Search) <──┘
                                        │
    all backend ──────────────> M12 (Frontend)
```

## Parallel execution windows

| Window | Milestones | Notes |
|--------|-----------|-------|
| **1** | M1 | Foundation — everything depends on it |
| **2** | M2, M3, M4 | Three independent tracks after M1 |
| **3** | M5, M6, M7, M8 | After their respective deps from window 2 |
| **4** | M9, M10, M11 | After window 3 deps |
| **5** | M12 | Frontend — needs all backend endpoints |

---

## Milestones

### M1 — Data Foundation: Model Extensions + New Models

**Deps:** none | **Parallel:** first milestone | **Repo:** lesser-host

Extend existing TableTheory models and create all new models.

**Model extensions:**
- `soul_agent_identity.go` — add `PrincipalAddress`, `PrincipalSignature`, `SelfDescriptionVersion`, `LifecycleStatus` (replacing simple Status), `LifecycleReason`, `SuccessorAgentId`
- `soul_agent_reputation.go` — add `Integrity` dimension, `DelegationsCompleted`, `BoundaryViolations`, `FailureRecoveries` signal counts
- `soul_agent_validation_record.go` — add `OptInStatus` (accepted/declined/pending)
- `soul_operation.go` — add Kind constants: `archive`, `self_suspend`, `self_reinstate`, `designate_successor`, `dispute`

**New models (8 new files):**
- `soul_agent_boundary.go` — PK: `SOUL#AGENT#{agentId}` / SK: `BOUNDARY#{boundaryId}`
- `soul_agent_continuity.go` — PK: `SOUL#AGENT#{agentId}` / SK: `CONTINUITY#{timestamp}#{entryType}`
- `soul_agent_relationship.go` — PK: `SOUL#AGENT#{agentId}` / SK: `RELATIONSHIP#{fromAgentId}#{timestamp}`
- `soul_agent_version.go` — PK: `SOUL#AGENT#{agentId}` / SK: `VERSION#{versionNumber}`
- `soul_agent_mint_conversation.go` — PK: `SOUL#AGENT#{agentId}` / SK: `MINT_CONVERSATION#{conversationId}`
- `soul_agent_dispute.go` — PK: `SOUL#AGENT#{agentId}` / SK: `DISPUTE#{disputeId}`
- `soul_agent_failure.go` — PK: `SOUL#AGENT#{agentId}` / SK: `FAILURE#{timestamp}#{failureId}`
- `soul_agent_index_relationship_from.go` — PK: `SOUL#RELATIONSHIPS_FROM#{fromAgentId}` / SK: `TO#{toAgentId}#{timestamp}`

**Key files:**
- Extend: `internal/store/models/soul_agent_identity.go`
- Extend: `internal/store/models/soul_agent_reputation.go`
- Extend: `internal/store/models/soul_agent_validation_record.go`
- Extend: `internal/store/models/soul_operation.go`
- Pattern ref: `internal/store/models/soul_agent_endorsement.go`
- Pattern ref: `internal/store/models/soul_agent_index_items.go`

**Acceptance:**
- [ ] All models compile and follow existing TableTheory patterns
- [ ] Key patterns match SPEC Section 12
- [ ] Each model has unit tests covering key generation and hooks

---

### M2 — Registration File v2 Schema + Publishing Pipeline

**Deps:** M1 | **Parallel with:** M3, M4

Implement the v2 registration file — construction, validation, JCS signing, S3 publishing, versioning.

**Deliverables:**
- Registration file v2 Go struct matching SPEC Appendix A JSON schema
- JCS canonicalization (RFC 8785) for signing
- EIP-191 signing for `selfAttestation` (keccak256 of JCS bytes)
- S3 publishing: current + versioned paths
- Version record creation on registration change
- V1 → v2 migration: flat capability arrays promoted to structured format
- Update `handlers_soul_update_registration.go` to validate and publish v2 schema

**Key files:**
- Extend: `internal/controlplane/handlers_soul_update_registration.go`
- Pattern ref: `internal/soul/` (existing signing, agent_id logic)

**Acceptance:**
- [ ] Round-trip: build → canonicalize → sign → verify passes
- [ ] V1 registration files readable; v2 writes produce complete schema
- [ ] Version records created on update, `previousVersionUri` chain maintained

---

### M3 — Contract Upgrades: selfMintSoul, principalOf, Attestor Registry

**Deps:** none (only needs SPEC) | **Parallel with:** M2, M4

Implement v2 contract changes in Solidity. Contracts are Sepolia-only — clean redesign.

**Deliverables:**
- `selfMintSoul(to, agentId, metaURI, avatarStyle, principal, principalSig)` — permissionless minting
- `principalOf(agentId)` view function
- `_principals` mapping, set immutably at mint time
- Attestor registry: `_attestors` mapping, `addAttestor`/`removeAttestor` (onlyOwner)
- `PrincipalDeclared(agentId, principal)` event
- Update `_mintSoulInternal` to store principal for all mint paths
- Backward compat: existing `mintSoul` and `mintSoulOwner` pass `address(0)` as principal

**Key files:**
- Extend: `contracts/contracts/SoulRegistry.sol`
- Extend: `contracts/test/SoulRegistry.test.js`

**Acceptance:**
- [ ] Contracts compile with Solidity ^0.8.24
- [ ] selfMintSoul with valid/invalid attestation passes/fails
- [ ] principalOf returns correct address for all mint paths
- [ ] Existing tests still pass (backward compat)

---

### M4 — Boundaries: Handler, Append-Only Logic, Signing

**Deps:** M1 | **Parallel with:** M3, M5, M6

Implement complete boundary lifecycle — highest-priority wishlist item.

**Deliverables:**
- `POST /api/v1/soul/agents/{agentId}/boundaries` — append boundary (portal auth)
- `GET /api/v1/soul/agents/{agentId}/boundaries` — public read (paginated)
- Append-only enforcement: no delete/update. Supersession via `supersedes` field
- Individual signing: EIP-191 over `keccak256(bytes(statement))`
- Registration file re-publish on boundary change

**Key files:**
- New: `internal/controlplane/handlers_soul_boundaries.go`
- Extend: `internal/controlplane/server.go` (register routes)

**Acceptance:**
- [ ] Boundaries are append-only (no delete/update)
- [ ] Individual boundary signatures verified
- [ ] Supersession chain visible in public read
- [ ] Registration file republished with updated boundaries

---

### M5 — Structured Capabilities: Replace Flat Array

**Deps:** M1, M2 | **Parallel with:** M4, M6

Replace `[]string` capabilities with structured claims.

**Deliverables:**
- Capability struct with lifecycle: `capability`, `scope`, `constraints`, `claimLevel`, `lastValidated`, `validationRef`, `degradesTo`
- Update registration begin to accept structured capabilities
- `GET /api/v1/soul/agents/{agentId}/capabilities` — public read
- V1 backward compat: flat `[]string` promoted to structured with `claimLevel: "self-declared"`
- Capability index update: `SOUL#CAP#` items carry claimLevel

**Key files:**
- Extend: `internal/controlplane/handlers_soul_registry.go`
- Extend: `internal/store/models/soul_agent_index_items.go`

**Acceptance:**
- [ ] Structured capabilities accepted in registration flow
- [ ] V1 flat arrays auto-promoted
- [ ] claimLevel transitions validated
- [ ] Public read returns structured format

---

### M6 — Sovereignty: Self-Suspend, Validation Opt-In, Disputes

**Deps:** M1, M4 | **Parallel with:** M5, M7

Implement agent-side sovereignty primitives.

**Deliverables:**
- `POST .../self-suspend` — sets `self_suspended`, creates continuity entry
- `POST .../self-reinstate` — only from `self_suspended`
- `POST .../validations/challenges/{id}/opt-in` — accept/decline
- `POST .../dispute` — create dispute record with evidence
- Update existing suspend/reinstate to differentiate operator vs self

**Key files:**
- Extend: `internal/controlplane/handlers_soul_suspension.go`
- New: `internal/controlplane/handlers_soul_sovereignty.go`
- Extend: `internal/controlplane/handlers_soul_validation.go`

**Acceptance:**
- [ ] Self-suspend/reinstate state machine correct
- [ ] Operator vs self-suspension clearly differentiated
- [ ] Validation opt-in/decline recorded without score penalty
- [ ] Disputes create continuity entries

---

### M7 — Continuity + Versioned Self

**Deps:** M1, M2 | **Parallel with:** M6, M8

Implement continuity journal and version history.

**Deliverables:**
- `POST .../continuity` — append entry (portal auth), individually signed
- `GET .../continuity` — public read, paginated by timestamp
- `GET .../versions` — version history with change summaries
- Shared `appendContinuityEntry` helper for use by M4, M6, M9
- `previousVersionUri` chain enforcement

**Key files:**
- New: `internal/controlplane/handlers_soul_continuity.go`
- New: `internal/controlplane/handlers_soul_versions.go`
- Extend: `internal/controlplane/soul_store_helpers.go`

**Acceptance:**
- [ ] Continuity entries individually signed and verifiable
- [ ] Version chain maintains `previousVersionUri` integrity
- [ ] Shared helper reusable from other milestones

---

### M8 — Relationships: Expanded Model + Trust Queries

**Deps:** M1 | **Parallel with:** M7, M9

Implement expanded relationship model with task-specific trust.

**Deliverables:**
- `POST .../relationships` — create record (portal auth), signature verified
- `GET .../relationships` — public read, filterable by `type` and `taskType`
- Dual-write: `SOUL#AGENT#{toAgentId}` + `SOUL#RELATIONSHIPS_FROM#{fromAgentId}`
- V1 backward compat: ENDORSEMENT records merged into relationship reads
- Trust revocations visible alongside grants (not deleted)

**Key files:**
- New: `internal/controlplane/handlers_soul_relationships.go`
- Pattern ref: `internal/store/models/soul_agent_endorsement.go`

**Acceptance:**
- [ ] Dual-write consistency maintained
- [ ] V1 endorsements appear in relationship reads
- [ ] Type + taskType filtering works
- [ ] Revocations don't delete grants

---

### M9 — Death + Succession

**Deps:** M1, M7 | **Parallel with:** M8, M10

Implement graceful shutdown and succession.

**Deliverables:**
- `POST .../archive` — sets `archived`, final continuity entry, read-only
- `POST .../successor` — sets `succeeded`, bidirectional continuity entries
- Lifecycle state machine: `active` → `archived`/`succeeded` (one-way)
- Public reads indicate archived/succeeded status
- On-chain token NOT burned (archival ≠ destruction)

**Key files:**
- New: `internal/controlplane/handlers_soul_lifecycle.go`
- Extend: `internal/controlplane/handlers_soul_public.go`

**Acceptance:**
- [ ] Archive makes registration read-only
- [ ] Succession creates entries on both agents
- [ ] Invalid lifecycle transitions rejected
- [ ] Archived agents visible but clearly marked

---

### M10 — Minting Conversation (Phase 2)

**Deps:** M1, M2, M5 | **Parallel with:** M9

Implement LLM-assisted minting conversation — the key new user-facing feature.

**Deliverables:**
- `POST .../register/{id}/mint-conversation` — streaming SSE endpoint
- Multi-model routing via existing AI adapters in `internal/ai/`
- Conversation records: `SoulAgentMintConversation` model
- Structured output: selfDescription, capabilities, boundaries, transparency
- System prompt per SPEC Section 4.4
- Model switching: human can start new conversation with different model

**Key files:**
- New: `internal/controlplane/handlers_soul_mint_conversation.go`
- Pattern ref: `internal/ai/` (existing LLM adapters)

**Acceptance:**
- [ ] Streaming conversation works with at least one LLM provider
- [ ] Output produces valid v2 registration file sections
- [ ] Conversation records persisted
- [ ] Model switching creates separate conversation records

---

### M11 — Reputation Worker v2 + Search Extensions

**Deps:** M1, M4, M8 | **Parallel with:** M10

Extend reputation computation and search.

**Deliverables:**
- Reputation worker: add `integrity` dimension, failure/recovery signals, relationship outcomes, delegation counts
- `GET .../transparency` — public read endpoint
- Search extensions: `boundary`, `claimLevel`, `status` filters
- Failure record ingestion pipeline
- Extended `/api/v1/soul/config` with integrity weights

**Key files:**
- Extend: `internal/soulreputationworker/server.go`
- Extend: `internal/controlplane/handlers_soul_public.go`
- Extend: `internal/controlplane/handlers_soul_config.go`

**Acceptance:**
- [ ] Integrity dimension computed correctly
- [ ] Search filters work individually and in combination
- [ ] Config exposes all dimension weights

---

### M12 — Frontend v2

**Deps:** all backend milestones (M4–M11)

Extend Svelte 5/TS frontend for all v2 features.

**Deliverables:**
- API client (`soul.ts`): TypeScript interfaces + fetch functions for all new endpoints
- Minting conversation UI: streaming display, model picker, conversation history, output review
- Boundary management: list, add (with wallet signature), supersession visualization
- Continuity timeline: chronological view with type-based icons
- Version history: versions with change summaries
- Self-suspension controls: suspend/reinstate with reason
- Relationship views: filterable by type, showing context
- Extended agent detail: all new v2 sections
- Discovery info: well-known URI display, MCP soulUri guidance

**Key files:**
- Extend: `web/src/lib/api/soul.ts`
- New pages in: `web/src/pages/`

**Acceptance:**
- [ ] All new API functions have TypeScript types matching backend
- [ ] Minting conversation streams correctly
- [ ] Agent detail page shows all v2 sections
- [ ] Wallet signing works for boundary creation

---

## Codex delegation strategy

Each milestone prompt should include:
1. **SPEC reference**: relevant sections from `lesser-soul/SPEC.md`
2. **Pattern file**: an existing handler/model to follow
3. **Files to modify**: explicit paths
4. **Acceptance criteria**: from the milestone

## Verification

After each milestone:
- `go test ./...` in lesser-host root
- `npm test` in lesser-host/contracts (for M3)
- `npm run typecheck` in lesser-host/web (for M12)
- Manual review of key files for pattern consistency
