# lesser-soul — Roadmap

This roadmap translates `SPEC.md` into an actionable delivery plan with milestones and acceptance criteria.

## Environments & AWS profiles

### Terminology

- **Central account**: the `lesser-host` control-plane AWS account (org root / shared services).
- **Instance account**: a single Lesser instance AWS account (e.g., Simulacrum dev/live instance).

### Profile conventions (requested)

| Work happens in… | AWS profile to use |
|---|---|
| Central account (`lesser-host`) | `Lesser` |
| Instance account (Simulacrum / Lesser instance) | `Sim` |

Assumptions:
- Default region: `us-east-1`.
- Stages: `lab` (dev instance domain) and `live` (production instance domain).

## Roadmap overview

- **Phase 1 (MVP):** single-agent, end-to-end task execution in an instance account.
- **Phase 2:** full agent pool, multi-subtask DAG, memory curation, BRIDGE tools, token refresh.
- **Phase 3:** central-account provisioning support in `lesser-host` (automated `soul up` + `soul bootstrap`).
- **Phase 4:** commercial hardening (tiers, moderation, attestations, observability, runbooks).

## Decisions (locked)

- **Ingress:** `/soul/*` is served via the instance CloudFront distribution (not a direct Function URL). CloudFront caching is disabled for `/soul/*`, and it forwards `Authorization` + query strings to the orchestrator origin (Lambda Function URL). Direct Function URL access is allowed but must enforce Lesser OAuth bearer auth.
- **Progress streaming:** implement SSE via AppTheory (`/soul/tasks/{id}/stream`), with `GET /soul/tasks/{id}` as the guaranteed fallback path.
- **Agent verification:** `soul bootstrap` supports `--auto-verify-agents`; default enabled for `lab`, default disabled for `live` unless explicitly opted in.
- **Delegation tokens:** Phase 1 logs `expiresIn` and fails loud on expiry. Phase 2 ships scheduled token refresh + alarms; optional “refresh-on-401” can be added later.
- **Idempotency:** SubTask execution is at-least-once; side effects are at-most-once (DynamoDB conditional updates), including budget debit (at-most-once per SubTask).
- **BRIDGE egress:** Phase 2 allows public internet with deny rules for private/metadata ranges; blocks redirects to private ranges; enforces hard timeouts and response size limits. Tighten to allowlists later if needed.
- **Cloud inference contract:** cloud providers must return `usage`. Missing/unparseable `usage` is a fatal provider failure and marks the SubTask FAILED (completion may be logged for debugging but is not a successful result). Enforced in code plus an allowlist of approved endpoints. Local CLI inference may run without usage/no debit.
- **Soul pack supply chain:** distribute a pinned, signed pack via a dedicated per-stage S3 bucket + KMS `SIGN/VERIFY` (GovTheory pattern). The provision runner verifies the KMS signature (RSASSA_PSS_SHA_256) before consuming any artifact. Packs are discovered via SSM stage pointers under `/soul/<stage>/...`.

## Risks & mitigations

- **SSE through CloudFront:** streaming can buffer/time out depending on behaviors and origin; mitigate by validating early and relying on the polling fallback.
- **BRIDGE SSRF/DNS tricks:** “deny RFC1918” is necessary but not sufficient; mitigate with redirect checks, DNS resolution safeguards, and (where possible) network-layer egress constraints + tests.

---

## Phase 0 — Project setup & alignment

### Milestone 0.1 — Repo scaffolding + local dev ergonomics

**Account:** N/A (local only)

Deliverables:
- Initial Go module layout per `SPEC.md` (cmd/, pkg/, infra/cdk/, scripts/, app-theory/).
- `Makefile` targets for build/test/deploy helpers (no secrets).
- Minimal CI entrypoint (even if CI wiring lives elsewhere): `go test ./...` + basic lint.

Acceptance criteria:
- [x] `go test ./...` runs clean on a fresh clone.
- [x] `make build` produces binaries for `cmd/orchestrator`, `cmd/agent-runner` (stubs acceptable at this milestone).
- [x] `infra/cdk/` can synthesize (`cdk synth`) without AWS credentials.

### Milestone 0.2 — Stage + domain configuration contract

**Account:** Instance (profile `Sim`) and Central (profile `Lesser`) — config only

Deliverables:
- A single source-of-truth for:
  - `SOUL_STAGE` (`lab`/`live`)
  - instance domain per stage (e.g., `dev.simulacrum.greater.website`, `simulacrum.greater.website`)
  - SSM path conventions: `/soul/<instance-domain>/...`
- Explicit documentation of which commands require `AWS_PROFILE=Sim` vs `AWS_PROFILE=Lesser`.

Acceptance criteria:
- [ ] A developer can follow docs and correctly choose `Sim` for instance deployments and `Lesser` for central deployments without guessing.

---

## Phase 1 — Core loop (MVP)

Goal: a working end-to-end task with a single agent turn in an instance account.

### Milestone 1.1 — Instance infrastructure skeleton (table/queues/lambdas)

**Account:** Instance (profile `Sim`)

Deliverables:
- CDK stack creating:
  - `soul-<stage>` DynamoDB table (TableTheory schema backing)
  - SQS queues: at minimum `soul-researcher` + `soul-results`
  - orchestrator Lambda (HTTP, Lambda Function URL) + agent-runner Lambda (SQS)
  - `/soul/*` path behavior added to the existing Lesser CloudFront distribution, routing to the orchestrator Lambda Function URL
- Basic IAM least-privilege policies for Lambda ↔ DynamoDB/SQS/SSM.

Acceptance criteria:
- [ ] `cdk deploy` (stage `lab`) completes successfully in the instance account.
- [ ] DynamoDB table exists with expected name and tags (stage + instance domain).
- [ ] SQS queues exist and Lambda event source mapping is enabled for agent-runner.
- [ ] `POST https://<instance-domain>/soul/tasks` is reachable through CloudFront (not a direct Lambda URL).
- [ ] CloudFront `/soul/*` behavior forwards auth (`Authorization`) to the orchestrator origin (no auth header stripping).
- [ ] CloudFront `/soul/*` behavior forwards query strings (no loss of `?task_id=...` / pagination parameters, etc.).
- [ ] CloudFront does not cache `/soul/*` responses (use a “caching disabled” policy) to prevent cross-user leakage.

### Milestone 1.2 — Lesser client + inference client (instance runtime)

**Account:** Instance (profile `Sim`) — runtime uses instance resources

Deliverables:
- `pkg/lesser/` client supporting:
  - `agentMemorySearch`
  - `createNote`
  - (optional for MVP) `getNote` for aggregation
- `pkg/inference/` OpenAI-compatible client supporting `Complete` (streaming optional here).
- SSM-backed loading for:
  - `/soul/<domain>/inference/url`
  - `/soul/<domain>/inference/key` (SecureString)

Acceptance criteria:
- [ ] In a deployed Lambda, cold-start loads inference URL/key from SSM (no secrets in env vars).
- [ ] Unit tests cover request/response marshaling for Lesser GraphQL and inference requests.
- [ ] Cloud inference responses without `usage` fail closed with a clear error (provider considered non-compliant).

### Milestone 1.3 — Orchestrator: POST task → enqueue 1 RESEARCHER subtask

**Account:** Instance (profile `Sim`)

> **Simplification:** Phase 1 does not use the LLM planner. The orchestrator hardcodes a single RESEARCHER subtask for every task. The planner LLM call and multi-subtask DAG are introduced in Phase 2 (M2.2).

Deliverables:
- `POST /soul/tasks`:
  - accepts `{ "goal": "..." }`
  - creates a `Task` + a single hardcoded `SubTask` (RESEARCHER, no planning step)
  - enqueues SQS message to the RESEARCHER queue
- `soul-results` handler:
  - updates `SubTask` and `Task` status on completion
  - stores `lesser_note_id` (or URL) as the audit reference

Acceptance criteria:
- [ ] Calling `POST /soul/tasks` returns `200` with `task_id`.
- [ ] DynamoDB contains `Task` with status `RUNNING` then `DONE`.
- [ ] DynamoDB contains `SubTask` with status `DONE` and a non-empty `lesser_note_id`.

### Milestone 1.4 — Agent-runner (RESEARCHER): memory → inference → Note post → result publish

**Account:** Instance (profile `Sim`)

Deliverables:
- SQS-triggered agent-runner path for `AGENT_TYPE=RESEARCHER` implementing the MVP loop:
  1. read SubTask + AgentConfig
  2. `agentMemorySearch`
  3. inference call
  4. `createNote` as the agent (delegation token from SSM)
  5. publish to `soul-results`
  6. write `RunLog` entries
- Credit debit call to `lesser-host` trust API can be stubbed/disabled for MVP (but interface should exist).

Acceptance criteria:
- [ ] For a test task, the agent posts a Note to Lesser attributed to `soul-researcher`.
- [ ] `RunLog` entries exist for LLM call + Note post + result publish (with truncation rules).
- [ ] The orchestrator observes completion via `soul-results` and marks the Task `DONE`.

### Milestone 1.5 — Bootstrap: register agent + verify + delegate token → SSM

**Account:** Instance (profile `Sim`)

> **Quarantine note:** Lesser places newly registered agents in quarantine. An agent in quarantine cannot post Notes, which means the Phase 1 exit criterion cannot be met until quarantine is lifted. `adminVerifyAgent` must be called (via the Simulacrum `/admin/agents` UI or directly via GraphQL) as an explicit step in the bootstrap sequence.
>
> **Token TTL:** `delegateToAgent` returns an `expiresIn` value. Confirm the token lifetime is long enough to cover the Phase 1 test window before relying on it in `lab`. The token-refresher (M2.5) does not exist yet in Phase 1 — if `expiresIn` is shorter than the test window, manually re-run bootstrap to obtain a fresh token rather than leaving the Lambda broken silently.

Deliverables:
- Bootstrap script/command that:
  - registers `soul-researcher` in Lesser via `registerAgent`
  - calls `adminVerifyAgent` (or documents the required manual step with exact GraphQL mutation) to lift quarantine
  - obtains delegation token (`delegateToAgent`)
  - stores access/refresh tokens in SSM under `/soul/<domain>/agents/researcher/{token,refresh}`
  - logs `expiresIn` so the operator knows when re-bootstrap is needed before M2.5 is live

Acceptance criteria:
- [ ] After bootstrap, SSM contains `SecureString` parameters at the expected paths.
- [ ] `soul-researcher` is verified (not in quarantine) in Lesser — confirmed via `agent(username: "soul-researcher") { verified }` query.
- [ ] A fresh Lambda deployment can read the token from SSM and successfully post a Note attributed to `soul-researcher`.

**Phase 1 exit criteria:**
- [ ] POST a research goal; receive a Lesser Note URL containing the synthesized result, attributed to `soul-researcher`.

---

## Phase 2 — Full agent pool + memory + tool execution

Goal: multi-subtask DAG tasks with memory curation, token refresh, and BRIDGE tools.

### Milestone 2.1 — Agent pool expansion + AgentConfig model

**Account:** Instance (profile `Sim`)

Deliverables:
- TableTheory models for `AgentConfig` (enabled flags, model id, prompt templates, limits).
- agent-runner support for:
  - `ASSISTANT`
  - `CURATOR`
  - `CUSTOM` variants (`coder`, `summarizer`)
- Bootstrap updated to register/delegate tokens for all enabled agents.

Acceptance criteria:
- [ ] All agent types can post Notes in Lesser using their own delegation tokens.
- [ ] Disabling an agent via `AgentConfig.Enabled=false` prevents queue consumption or returns a clear failure state without retries.

### Milestone 2.2 — Orchestrator planning: multi-subtask DAG + dependency chaining

**Account:** Instance (profile `Sim`)

Deliverables:
- Planner that produces a validated JSON subtask plan (schema-validated, deterministic error handling).
- Router that:
  - enqueues independent subtasks in parallel
  - enqueues dependent subtasks only when upstream subtasks are `DONE`
- Aggregation that fetches upstream results (from Lesser Note content) to provide context downstream.

Acceptance criteria:
- [ ] A goal that requires `RESEARCHER → CUSTOM:coder → CUSTOM:summarizer` completes end-to-end.
- [ ] Dependent subtasks never run before prerequisites are `DONE` (verified by Task/SubTask timestamps).
- [ ] A malformed planner response results in `Task` status `FAILED` with a useful error record (no partial/undefined state).

### Milestone 2.3 — Status + streaming UX contract (API)

**Account:** Instance (profile `Sim`)

Deliverables:
- `GET /soul/tasks/{id}` returning Task + SubTask status summary.
- `GET /soul/tasks/{id}/stream` providing SSE progress events (or a clearly documented alternative if SSE via Lambda URL is not viable).

Acceptance criteria:
- [ ] A client can track progress without polling faster than once per ~5s.
- [ ] Stream disconnect/reconnect does not corrupt task state and does not duplicate finalization.

### Milestone 2.4 — BRIDGE tool executor (sandboxed) + tool schemas

**Account:** Instance (profile `Sim`)

Deliverables:
- `cmd/tool-executor` Lambda consuming `soul-bridge` queue.
- Implement tools from `SPEC.md` with constraints:
  - `bash_exec` (timeout, memory cap)
  - `http_request` (public internet; deny private/metadata ranges; redirect-safe)
  - `file_read`/`file_write` (per-task scratch dir)
  - `lesser_search` (delegated `agentMemorySearch`)
- RunLog logging + truncation + input hashing.

Acceptance criteria:
- [ ] Tool calls cannot access RFC1918 addresses (validated with test cases).
- [ ] Redirects cannot be used to reach private/metadata ranges (e.g., AWS metadata), even if the initial URL is public.
- [ ] Tool outputs are truncated to the configured maximum and logged with hashes.
- [ ] Per-task scratch directories are isolated and cleaned up on completion (or TTL’d with scheduled cleanup).

### Milestone 2.5 — Memory curation + token refresher schedulers

**Account:** Instance (profile `Sim`)

Deliverables:
- EventBridge rule for curator turns (e.g., `rate(10 minutes)`).
- `cmd/token-refresher` scheduled Lambda:
  - refreshes all agent tokens using refresh tokens
  - writes back to SSM
  - updates `AgentConfig.TokenRefreshedAt`

Acceptance criteria:
- [ ] Curator posts tagged fact Notes that subsequently appear in `agentMemorySearch` results for related queries.
- [ ] Token refresh runs end-to-end without manual intervention for at least 7 consecutive days in `lab`.

### Milestone 2.6 — lesser-host credit debit integration

**Account:** Instance (profile `Sim`) → calls Central (`lesser-host` trust API)

> **Why Phase 2, not Phase 4:** Basic credit debit (`POST /api/v1/budget/debit`) is a simple async call after each inference. Deferring it to Phase 4 means Phase 2 and 3 operate with unbilled usage. Wiring it here establishes the billing path early and keeps the commercial model honest from the start.
>
> Note: this is distinct from attestation (`POST /api/v1/ai/claims/verify`), which remains in Phase 4 (M4.1).

Deliverables:
- `pkg/lesserhost/` client with `DebitBudget(ctx, module, target string, credits int, cached bool) error`.
- `lesserhost` middleware injecting the client into all agent-runner Lambdas.
- After each inference call, fire `DebitBudget` asynchronously (goroutine) — failures are logged to `RunLog` but do not fail the task.
- Debit is **at-most-once per SubTask**: protect against SQS redelivery/retries by guarding with a DynamoDB conditional update (e.g., only debit when `SubTask.CreditsDebitedAt` is unset) and/or a deterministic idempotency key if supported by the trust API.
- Instance API key loaded from SSM: `/soul/<domain>/lesser-host/instance-key` (SecureString).
- `AgentConfig` stores `CreditsPerKTokens` (default: `5`); credit calculation: `ceil(tokens_total / 1000.0 * CreditsPerKTokens)`.

Acceptance criteria:
- [ ] After a completed task, a `POST /api/v1/budget/debit` call is visible in lesser-host trust API logs for the instance.
- [ ] A budget-exceeded response (HTTP 402) from lesser-host does not fail the task — the `SubTask` completes with `credits_debited: 0` and a `RunLog` entry records the budget event.
- [ ] Reprocessing the same SubTask (simulated retry / duplicate SQS delivery) does not result in multiple debits for the same `<task_id>#<subtask_id>`.
- [ ] No inference secrets or token values appear in debit request payloads or logs.

**Phase 2 exit criteria:**
- [ ] A goal requiring research → code → summary completes end-to-end, with memory retrieval contributing to each turn and inference credits debited to the instance budget.

---

## Phase 3 — lesser-host integration + automated provisioning

Goal: enabling Soul at instance creation time, provisioned by `lesser-host` without manual steps.

> **`agentActivity` subscription — clarification:** The SPEC describes a GraphQL `agentActivity` subscription. This is **not** the orchestrator's result notification path — SQS (`soul-results`) handles that reliably. The `agentActivity` subscription is a Lesser-native real-time stream intended for **UI consumers** (e.g., the Simulacrum `/agents` page showing live agent activity). It does not need to be implemented in any orchestrator Lambda. Phase 3 is the right time to expose it if Simulacrum UI integration is desired.
>
> **greater-components (vendored):** Any Simulacrum UI additions (agent activity feed, task dashboard) are built with `greater-components` — the same library Simulacrum already uses. The `agentActivity` subscription is consumed via the vendored Lesser GraphQL adapter + `TransportManager`. Lock the vendored version via the project’s `components.json` (`ref: greater-vX.Y.Z`). Near-term, plan for an internal registry to distribute CLI artifacts + a curated registry index for your systems.

### Milestone 3.1 — lesser-host data model + portal API surface (central)

**Account:** Central (profile `Lesser`)

Deliverables (in `lesser-host` repo):
- `ProvisionJob` fields: `soulEnabled`, `soulProvisionedAt`.
- `Instance` model fields: `SoulEnabled`, `SoulVersion`, `SoulProvisionedAt`.
- Portal API `GET/PATCH` exposes Soul fields.

Acceptance criteria:
- [ ] Existing instance CRUD continues to work with zero changes for non-soul instances.
- [ ] Toggling `soul_enabled` via portal API updates the instance model and is persisted.

### Milestone 3.2 — Provision-worker: soul.deploy + soul.init steps (central)

**Account:** Central (profile `Lesser`)

Deliverables (in `lesser-host` repo):
- State machine adds `soul.deploy` after `lesser.init`:
  - triggers CodeBuild runner that runs `soul up` in the instance account
- `soul.init` runs `soul bootstrap` with the Lesser admin token and writes a receipt.

Acceptance criteria:
- [ ] Provisioning is idempotent: retrying after a partial failure does not create duplicate stacks or agents.
- [ ] A provision job with `soulEnabled=true` reliably completes both `lesser` and `soul` steps in `lab`.

### Milestone 3.3 — CodeBuild runner + artifact contract (central + instance)

**Account:** Central (profile `Lesser`) for orchestration, Instance (profile `Sim`) for actual deploy

Deliverables:
- Soul pack publishing infra (GovTheory-style):
  - dedicated per-stage S3 pack bucket (immutable/versioned artifacts)
  - KMS asymmetric key (`SIGN/VERIFY`) for pack manifest signatures
  - SSM discovery under `/soul/<stage>/...`: `packBucketName`, `signingKeyArn`, `packVersion`, (optional) `readerPolicyArn`, `publisherPolicyArn`
- Pack publisher (CI job or scripted release) uploads an immutable pack version:
  - `soul-pack-<version>.tgz`
  - `soul-pack-<version>.manifest.json` (deterministic bytes; includes file list + sha256 digests; may embed pins)
  - `soul-pack-<version>.manifest.sig` (KMS signature over `sha256(manifest.json)` using `RSASSA_PSS_SHA_256`)
  - updates `/soul/<stage>/packVersion` stage pointer
- CodeBuild project (or equivalent runner) that:
  - resolves the pack version from `SOUL_VERSION` (override) or `/soul/<stage>/packVersion` (default)
  - fetches the pinned `lesser-soul` pack from the pack bucket
  - verifies the pack’s signed manifest via KMS `Verify` (fail closed on verify failure) before consuming
  - deploys CDK to the instance account
  - runs bootstrap
  - writes `soul-state.json` receipt to S3
- Receipt schema includes: agent usernames, SSM token paths, queue URLs, soul-table name, deployed version.

Acceptance criteria:
- [ ] A new instance provision produces a valid `soul-state.json` in the expected S3 prefix.
- [ ] lesser-host can read and present receipt details in logs/portal without additional AWS permissions beyond least-privilege.
- [ ] The runner fails closed if manifest verification fails (no deploy/bootstrapping occurs on unsigned or tampered packs).
- [ ] Changing any file in the pack without updating the signed manifest causes verification failure (integrity is enforced).

### Milestone 3.4 — End-to-end integration test (managed instance)

**Account:** Central (profile `Lesser`) + Instance (profile `Sim`)

Deliverables:
- A reproducible integration test plan (script or runbook) that:
  1. provisions an instance with Soul enabled
  2. verifies agents registered + verified
  3. runs a multi-subtask Task and verifies completion

Acceptance criteria:
- [ ] Provision → bootstrap → execute task works without manual AWS console steps (except any explicit “admin verify agent” requirement, which must be documented).

**Phase 3 exit criteria:**
- [ ] Creating a new managed instance with `soul_enabled: true` yields a fully bootstrapped Soul stack without manual intervention.

---

## Phase 4 — Commercial hardening

Goal: production-ready reliability, tier gating, safety, and operational excellence.

### Milestone 4.1 — Attestations + trust surfaces

**Account:** Central (profile `Lesser`) + Instance (profile `Sim`)

Deliverables:
- Optional task-completion attestation (`POST /api/v1/ai/claims/verify`) on eligible tiers.
- Store attestation IDs on SubTask/Task records and surface via APIs.

Acceptance criteria:
- [ ] Attestation is generated only when configured/enabled and failures do not block task completion (graceful degradation).

### Milestone 4.2 — Tier enforcement + MODERATOR path

**Account:** Central (profile `Lesser`) + Instance (profile `Sim`)

Deliverables:
- Tier gating rules (Starter/Standard/Pro) implemented via configuration:
  - BRIDGE disabled on Starter
  - MODERATOR enabled on Pro (or when `requiresApproval=true`)
- MODERATOR implemented either as an agent turn or via lesser-host moderation API.

Acceptance criteria:
- [ ] Attempting to invoke BRIDGE tools on a tier where it is disabled fails closed (no tool execution).
- [ ] Moderator decisions are logged, explainable, and do not leak sensitive content.

### Milestone 4.3 — Observability + operational runbooks

**Account:** Instance (profile `Sim`) and Central (profile `Lesser`) as applicable

Deliverables:
- CloudWatch metrics (EMF) for:
  - task latency, error rate, retries
  - token usage, credit spend, cache hit rate
  - DLQ depth / queue age
- Alarms for stuck tasks, token refresh failures, and abnormal cost spikes.
- Runbooks for: stuck tasks, token refresh, DLQ replay, BRIDGE abuse, rate limit events.

Acceptance criteria:
- [ ] On-call can detect and resolve a stuck task using only documented runbooks.
- [ ] A controlled chaos test (forced downstream failure) results in expected alarms and safe degradation.

### Milestone 4.4 — Security review + abuse prevention

**Account:** Instance (profile `Sim`)

Deliverables:
- Threat model review for:
  - delegation tokens in SSM
  - BRIDGE SSRF / command injection surface
  - least-privilege IAM validation
- Automated checks where possible (policy linting, static analysis).

Acceptance criteria:
- [ ] BRIDGE cannot reach AWS metadata endpoints or private networks (tested).
- [ ] No secrets are logged; log scanning confirms absence of token-like strings.

**Phase 4 exit criteria:**
- [ ] Billing/tier gating, observability, and recovery procedures verified in `live`.

---

## Backlog (post–Phase 4 / optional)

- Multi-tenant load testing (concurrent tasks across many instances).
- Advanced planner improvements (tool-use planning, self-critique, budget-aware routing).
- **Simulacrum task dashboard** (consider pulling into Phase 3 as a parallel track once M2 is stable):
  Simulacrum already ships `/agents`, `/admin/agents`, agent delegation UI, and agent post filtering — the foundation is there. The remaining delta is a task history view and a `POST /soul/tasks` submission form. Built with `greater-components` (same library Simulacrum uses throughout):
  - `StepIndicator` for task progress (PLANNING → RUNNING → DONE)
  - `Timeline` / `TimelineVirtualized` for the agent activity feed (agent Notes render natively — no custom card needed)
  - `StreamingText` for live inference output if SSE is piped to the browser
  - `shared/admin` components for the task management and run-log views
  - `Badge` for agent type labels; `Alert` for budget-exceeded and error states
  - `agentActivity` subscription consumed via the vendored Lesser GraphQL adapter — lock to the git tag in Simulacrum’s `components.json`
- **Internal Greater registry** (near-term): host CLI artifacts + a curated registry index/checksums for deterministic vendoring across your systems.
