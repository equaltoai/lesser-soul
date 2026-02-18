# agentic-soul — Roadmap

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

---

## Phase 0 — Project setup & alignment

### Milestone 0.1 — Repo scaffolding + local dev ergonomics

**Account:** N/A (local only)

Deliverables:
- Initial Go module layout per `SPEC.md` (cmd/, pkg/, infra/cdk/, scripts/, app-theory/).
- `Makefile` targets for build/test/deploy helpers (no secrets).
- Minimal CI entrypoint (even if CI wiring lives elsewhere): `go test ./...` + basic lint.

Acceptance criteria:
- [ ] `go test ./...` runs clean on a fresh clone.
- [ ] `make build` produces binaries for `cmd/orchestrator`, `cmd/agent-runner` (stubs acceptable at this milestone).
- [ ] `infra/cdk/` can synthesize (`cdk synth`) without AWS credentials.

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
  - orchestrator Lambda (HTTP) + agent-runner Lambda (SQS)
- Basic IAM least-privilege policies for Lambda ↔ DynamoDB/SQS/SSM.

Acceptance criteria:
- [ ] `cdk deploy` (stage `lab`) completes successfully in the instance account.
- [ ] DynamoDB table exists with expected name and tags (stage + instance domain).
- [ ] SQS queues exist and Lambda event source mapping is enabled for agent-runner.

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

### Milestone 1.3 — Orchestrator: POST task → enqueue 1 RESEARCHER subtask

**Account:** Instance (profile `Sim`)

Deliverables:
- `POST /soul/tasks`:
  - accepts `{ "goal": "..." }`
  - creates a `Task` + a single `SubTask` (RESEARCHER)
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

### Milestone 1.5 — Bootstrap: register agent + delegate token → SSM

**Account:** Instance (profile `Sim`)

Deliverables:
- Bootstrap script/command that:
  - registers `soul-researcher` in Lesser
  - obtains delegation token (`delegateToAgent`)
  - stores access/refresh tokens in SSM under `/soul/<domain>/agents/researcher/{token,refresh}`

Acceptance criteria:
- [ ] After bootstrap, SSM contains `SecureString` parameters at the expected paths.
- [ ] A fresh Lambda deployment can read the token from SSM and successfully post a Note.

**Phase 1 exit criteria:**
- [ ] POST a research goal; receive a Lesser Note URL containing the synthesized result.

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
  - `http_request` (public IP allowlist / RFC1918 deny)
  - `file_read`/`file_write` (per-task scratch dir)
  - `lesser_search` (delegated `agentMemorySearch`)
- RunLog logging + truncation + input hashing.

Acceptance criteria:
- [ ] Tool calls cannot access RFC1918 addresses (validated with test cases).
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

**Phase 2 exit criteria:**
- [ ] A goal requiring research → code → summary completes end-to-end, with memory retrieval contributing to each turn.

---

## Phase 3 — lesser-host integration + automated provisioning

Goal: enabling Soul at instance creation time, provisioned by `lesser-host` without manual steps.

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
- CodeBuild project (or equivalent runner) that:
  - fetches a pinned `agentic-soul` artifact/version
  - deploys CDK to the instance account
  - runs bootstrap
  - writes `soul-state.json` receipt to S3
- Receipt schema includes: agent usernames, SSM token paths, queue URLs, soul-table name, deployed version.

Acceptance criteria:
- [ ] A new instance provision produces a valid `soul-state.json` in the expected S3 prefix.
- [ ] lesser-host can read and present receipt details in logs/portal without additional AWS permissions beyond least-privilege.

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
- UI integration in Simulacrum (`/agents` + task dashboard) beyond the minimal API contract.
