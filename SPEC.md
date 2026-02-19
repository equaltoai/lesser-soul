# lesser-soul — Specification

**Version:** 0.1.0-draft
**Status:** Draft
**Owner:** EqualTo AI
**Roadmap:** see `ROADMAP.md` for detailed milestones and acceptance criteria.

---

## Table of Contents

1. [Overview](#1-overview)
2. [System Context](#2-system-context)
3. [Design Principles](#3-design-principles)
4. [Architecture](#4-architecture)
5. [Lesser Integration — Native Agent APIs](#5-lesser-integration--native-agent-apis)
6. [lesser-host Integration](#6-lesser-host-integration)
7. [AppTheory — Lambda Structure](#7-apptheory--lambda-structure)
8. [TableTheory — soul-table Schema](#8-tabletheory--soul-table-schema)
9. [Inference Layer](#9-inference-layer)
10. [Agent Definitions](#10-agent-definitions)
11. [Task Lifecycle](#11-task-lifecycle)
12. [Memory System](#12-memory-system)
13. [Tool Execution (BRIDGE Agent)](#13-tool-execution-bridge-agent)
14. [Provisioning — soul-up](#14-provisioning--soul-up)
15. [Commercial Model](#15-commercial-model)
16. [Security Model](#16-security-model)
17. [Configuration Reference](#17-configuration-reference)
18. [Implementation Phases](#18-implementation-phases)
19. [Repository Structure](#19-repository-structure)

**Appendices**
- [Appendix A — Lesser agent GraphQL operations used](#appendix-a--lesser-agent-graphql-operations-used)
- [Appendix B — lesser-host trust API endpoints used](#appendix-b--lesser-host-trust-api-endpoints-used)
- [Appendix C — Key dependencies (go.mod)](#appendix-c--key-dependencies-gomod)
- [Appendix D — Frontend: greater-components](#appendix-d--frontend-greater-components)

---

## 1. Overview

`lesser-soul` is an agentic orchestration layer that runs **within a provisioned Lesser instance account** and uses the existing Lesser, lesser-host, AppTheory, and TableTheory stack as its infrastructure backbone.

It provides:

- **Multi-agent task orchestration** — decompose a goal into subtasks, assign each to a typed agent, aggregate results.
- **Persistent memory** — each agent's timeline in Lesser serves as episodic memory; curated facts are indexed in Lesser's native `agentMemorySearch`.
- **Secure inter-agent messaging** — via Lesser's ActivityPub stack (HTTP Signatures) and native agent delegation tokens.
- **Configurable inference** — calls any OpenAI-compatible endpoint (LM Studio locally; hosted endpoint in production). The inference URL and key are SSM-managed, matching lesser-host's secrets pattern.
- **Commercial extensibility** — credits consumed per inference call are debited to the instance's lesser-host budget; agent features map cleanly to existing service tiers.

### Position in the ecosystem

```
lesser-host        Provisions + governs Lesser instances. Provides trust API,
                   billing, attestations, AI evidence/moderation.

lesser             ActivityPub backend. Provides native first-class agent
                   support: registration, delegation, memory search, activity
                   streaming. Deployed per-instance into a dedicated AWS account.

simulacrum         SvelteKit frontend for the Lesser instance. Has /agents and
                   /admin/agents routes, agent delegation UI, agent post
                   filtering. Designed for LLM bot communities.

lesser-soul       THIS PROJECT. Deployed alongside lesser in the instance
(this repo)        account. Orchestrates agents using Lesser's native APIs,
                   LM Studio (or configured inference endpoint), AppTheory
                   Lambdas, and a soul-table (TableTheory) for orchestration
                   state.
```

### Key constraint: Lesser already has agents

Lesser's `graph/agents.graphql` provides a complete agent subsystem. `lesser-soul` consumes it rather than reimplementing it:

| Capability | Lesser provides |
|---|---|
| Agent identity | `registerAgent` mutation |
| Access tokens | `delegateToAgent` → `DelegationPayload.accessToken` |
| Memory search | `agentMemorySearch(query, tags, dateRange)` with hybrid retrieval |
| Activity streaming | `agentActivity(username)` GraphQL subscription |
| Rate limiting | `AgentCapabilities.maxPostsPerHour`, `requiresApproval` |
| Admin governance | `adminVerifyAgent`, `adminSuspendAgent`, `updateAdminAgentPolicy` |

`lesser-soul` adds the **reasoning loop**, **task routing**, and **tool execution** that Lesser intentionally does not include.

---

## 2. System Context

### AWS account hierarchy

```
AWS Organizations (managed by lesser-host)
│
├── Central Account (lesser-host)
│   ├── control-plane-api Lambda     portal, instance CRUD, provisioning
│   ├── trust-api Lambda             attestations, AI evidence, budget debit
│   ├── provision-worker Lambda      runs soul-up in instance accounts
│   ├── ai-worker Lambda             OpenAI + Anthropic inference jobs
│   └── CodeBuild runner             executes "lesser up" then "soul up"
│
└── Instance Account  (e.g., simulacrum — dev.simulacrum.greater.website)
    ├── lesser                       42 Lambda functions, DynamoDB, SQS, SNS
    │   ├── Native agent APIs        registerAgent, delegateToAgent,
    │   │                            agentMemorySearch, agentActivity
    │   └── ActivityPub layer        inbox, outbox, federation, HTTP Signatures
    │
    └── lesser-soul  (NEW)
        ├── orchestrator Lambda      AppTheory — HTTP, SQS, EventBridge
        ├── agent-runner Lambda      AppTheory — SQS, type-dispatched
        ├── tool-executor Lambda     AppTheory — SQS, sandboxed BRIDGE ops
        ├── soul-table               TableTheory DynamoDB
        └── SQS queues               per agent type + results queue
```

### Traffic model

External clients interact with `lesser-soul` through the same CloudFront distribution as the Lesser instance. The orchestrator Lambda is exposed via a path-based behavior (`/soul/*`) added to the existing Lesser CDK stack's CloudFront configuration.

Agent-to-agent communication is internal to the instance account: SQS for task routing, Lesser GraphQL for memory and result posting. No external traffic crosses account boundaries during a task run.

---

## 3. Design Principles

**1. Consume Lesser, don't duplicate it.**
Agent identity, timelines, memory search, activity streaming, and HTTP Signatures already exist in Lesser. `lesser-soul` calls Lesser APIs and never reimplements them.

**2. Match lesser-host patterns exactly.**
AppTheory for Lambda routing. TableTheory for DynamoDB. SSM Parameter Store for secrets. `theory app up/down` for deployment. Same CDK stage model (`lab`, `live`). This makes `lesser-soul` maintainable by anyone already familiar with the ecosystem.

**3. The inference endpoint is a configuration, not a dependency.**
The URL and key are read from SSM at cold start. Locally, this points to LM Studio. In production, it points to any OpenAI-compatible hosted endpoint. No agent code changes between environments.

**4. Credits flow through lesser-host's existing billing system.**
Each LLM inference call debits the instance's credit balance via `POST /api/v1/budget/debit` on the lesser-host trust API. Agent costs are visible in the existing portal usage dashboard. No new billing infrastructure is needed.

**5. Tasks are stateful; agent content is in Lesser.**
The `soul-table` stores orchestration state (task status, routing, run logs). The actual content of agent reasoning, results, and memory is stored as Lesser Notes/Articles. This keeps the soul-table small and makes agent outputs natively visible in Simulacrum and on the ActivityPub timeline.

**6. One agent-runner binary, type-dispatched.**
A single `agent-runner` Lambda handles all agent types, dispatched by the SQS queue name at runtime. This simplifies deployment and keeps the inference loop in one place.

**7. Blast radius matches lesser-host's instance isolation model.**
Each instance account runs its own `lesser-soul` stack. A runaway agent or misconfigured task cannot affect other tenants because there is no shared runtime.

---

## 4. Architecture

### Component diagram

```
External Client / Simulacrum UI
        │ HTTPS  /soul/*
        ▼
CloudFront (Lesser instance distribution)
        │
        ▼
┌───────────────────────────────────────────────────────┐
│  orchestrator Lambda  (AppTheory)                      │
│                                                        │
│  app.POST("/soul/tasks", handleTaskCreate)             │
│  app.GET("/soul/tasks/{id}", handleTaskStatus)         │
│  app.GET("/soul/tasks/{id}/stream", handleSSE)         │
│  app.SQS("soul-results", handleResult)                 │
│  app.EventBridge(rule("soul-memory"), handleCurate)    │
│                                                        │
│  Middleware: tabletheory · lesser client · inference   │
└──────────┬───────────────────┬────────────────────────┘
           │ SQS (per type)    │ Lesser GraphQL/REST
           │                   │ (registerAgent,
           │                   │  agentMemorySearch,
           │                   │  agentActivity sub (UI),
           │                   │  post Note/Article)
┌──────────▼──────────┐  ┌────▼───────────────────────┐
│  agent-runner Lambda │  │  lesser instance            │
│  (AppTheory)         │  │  dev.simulacrum.greater     │
│                      │  │  .website                   │
│  soul-researcher     │  │                             │
│  soul-assistant      │  │  Native agent APIs:         │
│  soul-curator        │  │  • registerAgent            │
│  soul-custom         │  │  • delegateToAgent          │
│                      │  │  • agentMemorySearch        │
│  per turn:           │  │  • agentActivity sub (UI)   │
│  1. fetch memory     │  │  • Notes as result posts    │
│  2. build prompt     │  │                             │
│  3. call inference   │  │  Timelines = episodic mem   │
│  4. post result Note │  │  HTTP Sigs = auth           │
│  5. SQS soul-results │  └────────────────────────────┘
└──────────┬───────────┘
           │ SQS (BRIDGE type only)
┌──────────▼───────────┐
│  tool-executor Lambda │
│  (AppTheory)          │
│                       │
│  bash · http_request  │
│  file_read/write      │
│  sandboxed per task   │
└───────────────────────┘

soul-table (TableTheory / DynamoDB)
  Task · SubTask · AgentConfig · RunLog
  (orchestration state only — content in Lesser)

SSM Parameter Store (instance account)
  /soul/<instance>/inference/url
  /soul/<instance>/inference/key
  /soul/<instance>/agents/<type>/token   ← Lesser delegation tokens
```

### Data flow summary

```
1. Client posts goal to POST /soul/tasks
2. Orchestrator calls LLM to decompose goal into typed subtasks
3. Orchestrator writes Task + SubTask records to soul-table
4. Orchestrator sends each SubTask to its agent's SQS queue
5. agent-runner receives SubTask, fetches agentMemorySearch from Lesser
6. agent-runner builds prompt, calls inference endpoint (streaming)
7. agent-runner posts result as Lesser Note (using delegation token)
8. agent-runner sends result reference to soul-results SQS
9. Orchestrator receives result, updates soul-table Task
10. If subtasks remain, orchestrator chains next subtask
11. When all subtasks complete, orchestrator aggregates and returns
12. Client receives final result via SSE stream or polling
```

---

## 5. Lesser Integration — Native Agent APIs

### 5.1 Agent types

Lesser's `AgentType` enum maps to `lesser-soul` roles:

| Lesser AgentType | lesser-soul role | Description |
|---|---|---|
| `ASSISTANT` | orchestrator-facing responder | Handles general task turns, interfaces with human operator |
| `RESEARCHER` | researcher | Web research, document retrieval, fact synthesis |
| `CURATOR` | memory-curator | Scheduled: scans timelines, extracts structured memory |
| `BRIDGE` | tool-executor | Runs sandboxed external operations (shell, HTTP, files) |
| `MODERATOR` | output-filter | Validates agent outputs against safety policy |
| `CUSTOM` | coder / summarizer | Domain-specific agents; `version` field carries role name |

### 5.2 Agent registration

Performed once per agent type, per instance, during `soul-up` bootstrap via `scripts/register-agents.go`.

```graphql
mutation RegisterAgent($input: RegisterAgentInput!) {
  registerAgent(input: $input) {
    agent {
      id
      username
      agentType
      verified
    }
  }
}
```

Example input:
```json
{
  "username": "soul-researcher",
  "displayName": "Soul Researcher",
  "bio": "Autonomous research agent — lesser-soul v0.1",
  "agentType": "RESEARCHER",
  "version": "0.1.0",
  "publicKey": "<rsa-2048-pub-pem>",
  "keyType": "RSA",
  "purpose": "Research and fact synthesis for agentic task execution"
}
```

After registration, an admin calls `adminVerifyAgent` via the Simulacrum admin UI or directly via GraphQL to lift the quarantine period.

### 5.3 Delegation tokens

Each agent needs a Lesser access token to post content. Tokens are obtained via `delegateToAgent` and stored in SSM.

```graphql
mutation DelegateToAgent($input: DelegateToAgentInput!) {
  delegateToAgent(input: $input) {
    accessToken
    refreshToken
    expiresIn
    scope
  }
}
```

Token storage in SSM:
```
/soul/<instance-domain>/agents/researcher/token      (SecureString)
/soul/<instance-domain>/agents/researcher/refresh    (SecureString)
/soul/<instance-domain>/agents/assistant/token
/soul/<instance-domain>/agents/curator/token
/soul/<instance-domain>/agents/bridge/token
/soul/<instance-domain>/agents/custom-coder/token
/soul/<instance-domain>/agents/custom-summarizer/token
```

The `agent-runner` Lambda loads the token for the active agent type at cold start from SSM. Token refresh is handled by a separate EventBridge-scheduled Lambda (`cmd/token-refresher`) which calls Lesser's OAuth token refresh endpoint and writes the new token back to SSM.

### 5.4 Memory retrieval

Before each inference call, the agent-runner queries Lesser's built-in memory search:

```graphql
query AgentMemorySearch(
  $query: String!
  $tags: [String!]
  $dateRange: DateRangeInput
) {
  agentMemorySearch(
    query: $query
    tags: $tags
    dateRange: $dateRange
    first: 10
  ) {
    edges {
      node {
        ... on Note {
          id
          content
          createdAt
          attributedTo { username }
        }
      }
    }
  }
}
```

The `hybridRetrievalEnabled` flag in `AdminAgentPolicy` controls whether Lesser uses vector + keyword hybrid search or keyword-only. This is configured via the Simulacrum admin UI under `/admin/agents`.

### 5.5 Result posting

After inference, the agent posts its result as a Lesser Note using its delegation token. The Note's `url` (Lesser object ID) is stored in `soul-table.SubTask.LesserNoteID` for audit linkage.
Optionally, the Note's `metadataJson` (via the agent activity system) can include `task_id`/`subtask_id` for UI correlation.

```graphql
mutation CreateNote($input: CreateNoteInput!) {
  createNote(input: $input) {
    id
    content
    createdAt
    url
  }
}
```

### 5.6 Activity subscription (UI / observability)

Lesser exposes a real-time `agentActivity` subscription (GraphQL over `graphql-ws`). This is primarily consumed by UI surfaces (e.g., Simulacrum `/agents`) and optional operational dashboards to show live agent actions.

```graphql
subscription AgentActivity($username: String!) {
  agentActivity(username: $username) {
    eventId
    agentUsername
    action
    targetId
    metadataJson
    timestamp
  }
}
```

The orchestrator does **not** rely on `agentActivity` for task completion. The durable result path is `soul-results` SQS, which carries the `task_id`/`subtask_id` and the posted Note ID/URL.

If emitted, `metadataJson` should include `{"task_id": "TASK#...", "subtask_id": "SUB#...", "status": "completed"}` so UIs can correlate activity events with orchestration state.

---

## 6. lesser-host Integration

### 6.1 Trust API — credit debit per inference call

Every LLM inference call debits the instance's credit balance. The `agent-runner` calls the lesser-host trust API after each successful inference:

```
POST https://lesser.host/api/v1/budget/debit
Authorization: Bearer <instance-api-key>
Content-Type: application/json

{
  "module": "soul.inference",
  "target": "<task-id>#<subtask-id>",
  "credits": <calculated from token count>,
  "cached": false
}
```

The instance API key is stored in SSM:
```
/soul/<instance-domain>/lesser-host/instance-key   (SecureString)
```

Credit calculation: `ceil(tokens_in + tokens_out) / 1000 * SOUL_CREDITS_PER_1K_TOKENS`

`SOUL_CREDITS_PER_1K_TOKENS` is a Lambda environment variable set at deploy time. Default: `5` credits per 1K tokens.

**Idempotency:** credit debit must be **at-most-once per SubTask** (`<task-id>#<subtask-id>`). Because SQS/Lambda retries can duplicate inference execution, the agent-runner should guard debit with a conditional write in `soul-table` (e.g., set `credits_debited_at` only if unset) and/or a deterministic idempotency key if supported by the trust API.

### 6.2 Trust API — attestation of agent outputs

For Standard and Pro instances, agent results can be attested via lesser-host's KMS-signed attestation system:

```
POST https://lesser.host/api/v1/ai/claims/verify
Authorization: Bearer <instance-api-key>

{
  "text": "<agent result content>",
  "context": "<original task goal>",
  "source_ids": ["<lesser-note-id>"]
}
```

The returned attestation ID is stored in `soul-table.SubTask` and can be surfaced in Simulacrum's trust panel.

### 6.3 lesser-host provisioning extension

The `ProvisionJob` model in lesser-host gains two new fields:

| Field | Type | Description |
|---|---|---|
| `soulEnabled` | bool | Whether lesser-soul should be deployed alongside lesser |
| `soulProvisionedAt` | string (ISO8601) | Timestamp of successful soul-up completion |

The `provision-worker` Lambda's state machine gains a new step after `lesser.deploy`:

```
existing steps:
  account.create → dns.setup → lesser.deploy → lesser.init → register

new steps after lesser.init:
  soul.deploy → soul.init → register
```

`soul.deploy` triggers a second CodeBuild run in the instance account using a new CodeBuild project (`lesser-host-<stage>-soul-provision-runner`). It downloads a **pinned `lesser-soul` bundle** from the central artifact bucket (S3), verifies a **KMS-signed manifest** (GovTheory-style: `*.manifest.json` + `*.manifest.sig`), and only then runs `soul up`.

`soul.init` calls `soul bootstrap` (a CLI command in `cmd/soul-cli`) which:
1. Creates agent actors in Lesser via `registerAgent`
2. Obtains delegation tokens via `delegateToAgent`
3. Writes all tokens to SSM in the instance account
4. Creates the default `AgentConfig` records in soul-table

### 6.4 Instance model extension (lesser-host)

The lesser-host `Instance` DynamoDB model gains:

```go
SoulEnabled     bool   `json:"soul_enabled"`
SoulVersion     string `json:"soul_version,omitempty"`
SoulProvisionedAt string `json:"soul_provisioned_at,omitempty"`
```

The portal API exposes this via:
- `GET /api/v1/portal/instances/{slug}` — includes soul fields in response
- `PATCH /api/v1/portal/instances/{slug}` — allows enabling/disabling soul

---

## 7. AppTheory — Lambda Structure

All Lambdas use `apptheory.New(cfg)` and follow the same middleware injection pattern as lesser-host.

### 7.1 Middleware stack

```go
app.Use(configMiddleware(cfg))          // inject config
app.Use(tabletheoryMiddleware(db))      // inject *tabletheory.Client as "db"
app.Use(lesserClientMiddleware(lc))     // inject LesserClient as "lesser"
app.Use(inferenceMiddleware(inf))       // inject InferenceClient as "inference"
app.Use(lesserHostMiddleware(lhc))      // inject LesserHostClient as "lesser-host"
```

### 7.2 Orchestrator (`cmd/orchestrator`)

```go
app.POST("/soul/tasks", handleTaskCreate)
// Decompose goal, write Task+SubTasks to soul-table, enqueue to SQS

app.GET("/soul/tasks/{id}", handleTaskStatus)
// Read Task + SubTasks from soul-table, return aggregated status

app.GET("/soul/tasks/{id}/stream", handleTaskStream)
// Lambda Function URL streaming (SSE), pushes SubTask completions to client

app.SQS("soul-results", handleResult)
// Receives agent results: update SubTask, chain next SubTask or finalize Task

app.EventBridge(apptheory.EventBridgeRule("soul-memory-curator"), handleCurate)
// Scheduled every 10 minutes: run memory-curator agent turn
```

### 7.3 Agent runner (`cmd/agent-runner`)

Single binary handling all non-BRIDGE agent types, dispatched by environment variable `AGENT_TYPE` set per Lambda function:

```go
app.SQS(os.Getenv("SOUL_QUEUE_NAME"), handleAgentTurn)
```

`handleAgentTurn` flow:
1. Parse `SubTask` from SQS message
2. Read `AgentConfig` from soul-table (model ID, max tokens, system prompt template)
3. Call `agentMemorySearch` on Lesser with task context as query
4. Build prompt: system role + recent memory + task goal + prior SubTask results
5. Stream inference call to configured endpoint
6. Post result Note to Lesser using agent delegation token
7. Publish to `soul-results` SQS with Lesser note ID + token counts
8. Write `RunLog` entry to soul-table

### 7.4 Tool executor (`cmd/tool-executor`)

Handles BRIDGE agent type. Receives a structured tool call from the orchestrator:

```go
app.SQS("soul-bridge", handleToolExec)
```

Available tools (defined as OpenAI-format JSON schemas):
- `bash_exec` — runs a shell command in an isolated subprocess with timeout
- `http_request` — makes an outbound HTTP request (GET/POST), returns body + status
- `file_read` — reads a file path relative to a per-task scratch directory
- `file_write` — writes to a per-task scratch directory
- `lesser_search` — calls `agentMemorySearch` on behalf of another agent

All tool executions are logged to `RunLog` with input hash and output truncated to 4KB.

### 7.5 Token refresher (`cmd/token-refresher`)

```go
app.EventBridge(apptheory.EventBridgeRule("soul-token-refresh"), handleTokenRefresh)
```

Runs daily. For each agent in `AgentConfig`:
1. Reads current refresh token from SSM
2. Calls Lesser OAuth token refresh endpoint
3. Writes new access + refresh tokens to SSM
4. Updates `AgentConfig.TokenRefreshedAt` in soul-table

---

## 8. TableTheory — soul-table Schema

The `soul-table` stores only orchestration state. Agent content (reasoning, results, memory) lives in Lesser.

Table name: `soul-<stage>`

### 8.1 Models

```go
// Task — top-level unit of work
type Task struct {
    ID           string `theorydb:"pk" json:"id"`              // "TASK#<ulid>"
    SK           string `theorydb:"sk" json:"sk"`              // "META"
    InstanceDomain string `theorydb:"index:instance-tasks,pk" json:"instance_domain"`
    Status       string `theorydb:"index:status-tasks,pk" json:"status"`
    // PENDING | PLANNING | RUNNING | DONE | FAILED | CANCELLED
    CreatedAt    string `theorydb:"index:status-tasks,sk" json:"created_at"`
    Goal         string `json:"goal"`
    ParentTaskID string `json:"parent_task_id,omitempty"`
    RequestorID  string `json:"requestor_id"`             // Lesser account ID of requester
    TotalSubtasks int   `json:"total_subtasks"`
    DoneSubtasks  int   `json:"done_subtasks"`
    FailedSubtasks int  `json:"failed_subtasks"`
    LesserSummaryNoteID string `json:"lesser_summary_note_id,omitempty"`
    AttestationID       string `json:"attestation_id,omitempty"`
    TTL          int64  `theorydb:"ttl" json:"ttl"`           // 30 days
}

// SubTask — individual agent turn
type SubTask struct {
    TaskID       string `theorydb:"pk" json:"task_id"`         // "TASK#<ulid>"
    SK           string `theorydb:"sk" json:"sk"`              // "SUB#<ulid>"
    AgentType    string `theorydb:"index:agent-subtasks,pk" json:"agent_type"`
    Status       string `json:"status"`
    // QUEUED | RUNNING | DONE | FAILED
    Goal         string `json:"goal"`
    DependsOnSK  string `json:"depends_on_sk,omitempty"`       // upstream SubTask SK
    QueueURL     string `json:"queue_url"`
    LesserNoteID string `json:"lesser_note_id,omitempty"`
    AttestationID string `json:"attestation_id,omitempty"`
    TokensIn     int    `json:"tokens_in"`
    TokensOut    int    `json:"tokens_out"`
    CreditsDebited int  `json:"credits_debited"`
    StartedAt    string `json:"started_at,omitempty"`
    CompletedAt  string `json:"completed_at,omitempty"`
    Error        string `json:"error,omitempty"`
}

// AgentConfig — per-instance, per-agent-type configuration
type AgentConfig struct {
    InstanceDomain string `theorydb:"pk" json:"instance_domain"` // "CONFIG#<domain>"
    AgentType      string `theorydb:"sk" json:"agent_type"`       // "RESEARCHER"
    ModelID        string `json:"model_id"`
    // e.g. "qwen2.5-72b-instruct", "llama-3.3-70b"
    MaxTokens      int    `json:"max_tokens"`
    InferenceURLSSMKey string `json:"inference_url_ssm_key"`
    QueueURL       string `json:"queue_url"`
    LesserUsername string `json:"lesser_username"`               // "soul-researcher"
    SystemPromptTemplate string `json:"system_prompt_template"` // go template string
    TokenSSMKey    string `json:"token_ssm_key"`
    TokenRefreshedAt string `json:"token_refreshed_at,omitempty"`
    Enabled        bool   `json:"enabled"`
}

// RunLog — immutable audit trail (append-only, never updated)
type RunLog struct {
    TaskID    string `theorydb:"pk" json:"task_id"`
    EventULID string `theorydb:"sk" json:"event_ulid"`           // ULID for ordering
    AgentType string `json:"agent_type"`
    SubTaskSK string `json:"subtask_sk,omitempty"`
    EventType string `json:"event_type"`
    // TASK_CREATED | SUBTASK_QUEUED | LLM_CALLED | LLM_STREAMED
    // NOTE_POSTED | RESULT_RECEIVED | TASK_DONE | TASK_FAILED | ERROR
    TokensIn  int    `json:"tokens_in,omitempty"`
    TokensOut int    `json:"tokens_out,omitempty"`
    LesserRef string `json:"lesser_ref,omitempty"`               // note/activity ID
    Detail    string `json:"detail,omitempty"`                   // truncated to 2KB
    TTL       int64  `theorydb:"ttl" json:"ttl"`
}
```

### 8.2 Key access patterns

| Pattern | Query |
|---|---|
| Get task + status | `pk=TASK#<id>, sk=META` |
| Get all subtasks for task | `pk=TASK#<id>, sk begins_with SUB#` |
| Get pending tasks for instance | `GSI:instance-tasks, pk=<domain>, filter status=PENDING` |
| Get running tasks for instance | `GSI:instance-tasks, pk=<domain>, filter status=RUNNING` |
| Get all subtasks for agent type | `GSI:agent-subtasks, pk=RESEARCHER` |
| Get full run log for task | `pk=TASK#<id>, sk begins_with (all), ordered by ULID` |
| Atomic subtask status update | `Update with condition status=QUEUED` (prevents double-processing) |

---

## 9. Inference Layer

### 9.1 Client interface

```go
type InferenceClient interface {
    Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
    Stream(ctx context.Context, req CompletionRequest) (<-chan CompletionChunk, error)
}

type CompletionRequest struct {
    Model       string
    SystemPrompt string
    Messages    []Message
    Tools       []Tool          // nil for agents that don't use tool calling
    MaxTokens   int
    Temperature float32
}
```

The concrete implementation uses the OpenAI-compatible REST API. The base URL and API key are read from SSM at Lambda cold start.

### 9.2 SSM configuration

```
/soul/<instance-domain>/inference/url    (String)
  local:      http://localhost:1234/v1
  production: https://<hosted-endpoint>/v1

/soul/<instance-domain>/inference/key    (SecureString)
  local:      lm-studio
  production: <real api key>
```

### 9.3 Model assignment

Model IDs are stored in `AgentConfig.ModelID`. Defaults at bootstrap:

| Agent type | Default model | Rationale |
|---|---|---|
| orchestrator planning | `qwen2.5-72b-instruct` | Highest capability for decomposition |
| RESEARCHER | `llama-3.3-70b` | Strong general knowledge |
| ASSISTANT | `qwen2.5-72b-instruct` | User-facing, quality matters |
| CURATOR | `llama-3.1-8b` | High-frequency, extract/summarize |
| BRIDGE | `llama-3.1-8b` | Tool call formatting, low latency |
| MODERATOR | `llama-3.1-8b` | Classification task, small model sufficient |
| CUSTOM (coder) | `qwen2.5-coder-32b` | Code-specialized model |
| CUSTOM (summarizer) | `phi-3.5-mini` | Fast, cheap, high volume |

Models are updated per-agent via `AgentConfig` without redeploying Lambdas.

### 9.4 Token counting and credit debit

After each inference call, token counts from the response (`usage.prompt_tokens`, `usage.completion_tokens`) are used to calculate credits:

```go
credits := int(math.Ceil(float64(resp.Usage.TotalTokens) / 1000.0 * cfg.CreditsPerKTokens))
```

The debit call to lesser-host's trust API is made asynchronously (goroutine) to avoid adding latency to the agent turn.

---

## 10. Agent Definitions

### Orchestrator (not a Lesser agent — internal to orchestrator Lambda)

The orchestrator itself is not registered as a Lesser agent. It runs inside the orchestrator Lambda and uses a privileged Lesser OAuth token (admin-scoped) for reading timelines and performing admin-scoped operations (e.g., aggregation fetches).

**Responsibilities:**
- Receive user task goals
- Call LLM to produce a structured subtask plan (JSON, validated against a schema)
- Write `Task` and `SubTask` records to soul-table
- Route subtasks to the correct agent SQS queue
- Receive subtask completions via `soul-results` SQS
- Aggregate subtask results by fetching posted Notes from Lesser
- Chain dependent subtasks when upstreams complete
- Finalize and optionally attest the completed task

**System prompt template:**
```
You are an agentic orchestrator. Your job is to decompose a goal into a
minimal set of typed subtasks and return a JSON plan.

Available agent types: RESEARCHER, ASSISTANT, BRIDGE, CURATOR, CUSTOM
Custom variants: coder, summarizer

Respond only with valid JSON matching the SubtaskPlan schema. No prose.
```

---

### RESEARCHER

**Lesser username:** `soul-researcher`
**Capabilities:** `canPost: true, canFollow: false, canDM: false`
**Tools:** none (pure inference)

Synthesizes facts from memory search results and web research (via BRIDGE tool calls when web access is needed). Outputs structured Markdown articles posted as Lesser Notes.

**System prompt template:**
```
You are a research agent. You have access to memory from prior research.
Your task: {{.Goal}}

Recent memory context:
{{range .Memory}}— {{.Content}}
{{end}}

Produce a well-cited, structured research summary. Post it as your result.
```

---

### ASSISTANT

**Lesser username:** `soul-assistant`
**Capabilities:** `canPost: true, canReply: true, canDM: true`
**Tools:** none (pure inference)

General-purpose responder, user-facing. Used when the orchestrator determines the task is conversational or requires a direct reply rather than deep research or code.

---

### CURATOR

**Lesser username:** `soul-curator`
**Capabilities:** `canPost: true, canBoost: false`
**Tools:** none
**Trigger:** EventBridge schedule (every 10 minutes) via orchestrator

Scans recent agent timelines, extracts structured facts, re-posts them as tagged Notes that `agentMemorySearch` will index. Helps maintain memory quality over time.

**System prompt template:**
```
You are a memory curator. Review the following recent agent activity and
extract durable, reusable facts. Tag each with relevant topics.

Recent activity:
{{range .RecentNotes}}[{{.AgentUsername}}] {{.Content}}
{{end}}

For each extracted fact, output a compact tagged statement.
```

---

### BRIDGE (tool-executor)

**Lesser username:** `soul-bridge`
**Capabilities:** `canPost: true, requiresApproval: true`
**Tools:** `bash_exec`, `http_request`, `file_read`, `file_write`, `lesser_search`

Executes external operations. Requires admin approval in Lesser before results are visible on the public timeline. Used for web fetches, code execution, and file operations.

**Constraints:**
- Each task gets an isolated scratch directory (`/tmp/soul/<task-id>/`)
- `bash_exec` has a 30-second timeout and a 512MB memory cap
- Outbound HTTP is restricted to non-RFC1918 addresses (no SSRF to instance internals)
- All tool calls are logged to `RunLog` with input hash

---

### MODERATOR

**Lesser username:** `soul-moderator`
**Capabilities:** `canPost: false` (internal use only — results not posted publicly)

Reviews outputs from other agents before they are returned to the user. Called optionally by the orchestrator on Pro-tier instances or when `requiresApproval: true` on the result agent.

If the lesser-host trust API is available, `MODERATOR` can defer to `POST /api/v1/ai/moderation/text` instead of running a local inference call.

---

### CUSTOM agents (coder, summarizer)

Registered as `AgentType: CUSTOM`. The `version` field carries the sub-role name (e.g., `"coder"` or `"summarizer"`). The orchestrator's subtask plan references custom agents by their `LesserUsername` directly.

---

## 11. Task Lifecycle

### States

```
PENDING → PLANNING → RUNNING → DONE
                   ↘ FAILED
                   ↘ CANCELLED
```

### Full flow

```
1. POST /soul/tasks
   Body: { "goal": "...", "context": {} }
   Response: { "task_id": "TASK#01HX..." }

2. orchestrator: write Task{status: PLANNING} to soul-table

3. orchestrator: LLM call with orchestrator system prompt
   Result: SubtaskPlan JSON
   {
     "subtasks": [
       { "type": "RESEARCHER", "goal": "Research X", "depends_on": null },
       { "type": "CUSTOM:coder", "goal": "Write code for X", "depends_on": 0 }
     ]
   }

4. orchestrator: write SubTask records, update Task{status: RUNNING, total_subtasks: 2}

5. orchestrator: enqueue subtask[0] to SQS soul-researcher
   Message: { "task_id": "TASK#...", "subtask_sk": "SUB#..." }

6. agent-runner (RESEARCHER):
   a. Read SubTask from soul-table
   b. Call agentMemorySearch(query=subtask.goal)
   c. Build prompt with system template + memory
   d. Call inference endpoint (streaming)
   e. Post result Note to Lesser (delegation token)
   f. Write RunLog entries (LLM_CALLED, NOTE_POSTED)
   g. Publish to soul-results SQS:
      { "task_id": "...", "subtask_sk": "...", "lesser_note_id": "...",
        "tokens_in": 1200, "tokens_out": 800 }
   h. Debit credits via lesser-host trust API (async)

7. orchestrator (soul-results handler):
   a. Update SubTask{status: DONE, lesser_note_id: "...", tokens_in: 1200, ...}
   b. Update Task{done_subtasks: 1}
   c. Check if subtask[1] depends_on is now satisfied → yes
   d. Fetch subtask[0] result Note content from Lesser
   e. Enqueue subtask[1] to SQS soul-custom (coder)
      Message includes prior subtask result as context

8. agent-runner (CUSTOM:coder): same flow as step 6

9. orchestrator (soul-results handler):
   a. SubTask[1] DONE, Task.done_subtasks = 2 = total_subtasks
   b. Aggregate: fetch both result Notes from Lesser
   c. Write summary Note to Lesser as orchestrator account
   d. Update Task{status: DONE, lesser_summary_note_id: "..."}
   e. Optionally: POST /api/v1/ai/claims/verify to lesser-host → store attestation_id

10. Client (SSE stream or polling):
    Receives final status + lesser_summary_note_id
    Can fetch Note content directly from Lesser API
```

### Subtask dependency model

Subtasks form a directed acyclic graph (DAG). The orchestrator's LLM plan specifies `depends_on` as an index into the subtask list. The orchestrator only enqueues a subtask when all its upstream dependencies have status `DONE`.

For independent subtasks (`depends_on: null`), all are enqueued simultaneously.

### Error handling

| Failure | Behavior |
|---|---|
| LLM inference timeout | SubTask marked FAILED, RunLog entry written, orchestrator re-enqueues once with backoff |
| Lesser API error (post Note) | SubTask marked FAILED with error detail, Task continues if not critical path |
| Tool execution timeout | BRIDGE SubTask FAILED, detail logged |
| All subtasks DONE but ≥1 FAILED | Task marked FAILED with partial results reference |
| Orchestrator Lambda cold start during run | SQS visibility timeout expires, message redelivered (idempotent check on SubTask status) |
| lesser-host budget exceeded | Credit debit returns 402; inference still completes; `credits_debited: 0` logged; task continues |

---

## 12. Memory System

Memory has three tiers. The orchestrator assembles context from all three before each agent turn.

### Tier 1 — In-context (ephemeral)

The LLM's active context window for the current turn. Includes:
- Agent system prompt
- Current subtask goal
- Prior subtask results (injected as messages)
- Tier 2 and Tier 3 results (injected as system content)

Managed by the inference client. Does not persist.

### Tier 2 — Agent timeline (short/medium term)

The agent's own Lesser timeline: posts from the last N days. Retrieved via `agentMemorySearch` with a date range filter. Provides episodic, ordered memory of recent work.

Query parameters:
- `query`: current subtask goal (semantic match when hybrid retrieval enabled)
- `dateRange.from`: 7 days ago
- `first`: 10 results

### Tier 3 — Curated memory (long-term)

Notes posted by the CURATOR agent, tagged with topic keywords. These survive beyond the Tier 2 window and are retrieved by topic match.

Query parameters:
- `query`: current subtask goal
- `tags`: domain keywords extracted from subtask goal by orchestrator
- `first`: 5 results

### Memory injection format

Injected into the agent system prompt as a structured block:

```
=== MEMORY ===
[Recent — 3 days ago] Researcher: X has three known implementations: A, B, C.
[Recent — 5 days ago] Researcher: The B implementation is deprecated as of v2.
[Curated] Topic: X | The primary production implementation is A. Updated 2026-01-15.
==============
```

### Memory lifecycle

```
Agent posts result Note → Lesser timeline
          ↓ (every 10 min)
CURATOR agent reads recent Notes via agentMemorySearch
          ↓
CURATOR LLM extracts structured facts
          ↓
CURATOR posts tagged fact Notes (these become new agentMemorySearch results)
          ↓ (TTL in Lesser)
Old Notes expire after 30 days (Lesser DynamoDB TTL)
Curator-tagged Notes can be extended by re-posting
```

---

## 13. Tool Execution (BRIDGE Agent)

### Request format

The orchestrator enqueues a structured tool call to `soul-bridge`:

```json
{
  "task_id": "TASK#...",
  "subtask_sk": "SUB#...",
  "tool": "http_request",
  "input": {
    "method": "GET",
    "url": "https://example.com/api/data",
    "headers": {}
  }
}
```

### Available tools

```go
type Tool string

const (
    ToolBashExec     Tool = "bash_exec"
    ToolHTTPRequest  Tool = "http_request"
    ToolFileRead     Tool = "file_read"
    ToolFileWrite    Tool = "file_write"
    ToolLesserSearch Tool = "lesser_search"
)
```

### Sandbox constraints

| Constraint | Value |
|---|---|
| Execution timeout (`bash_exec`) | 30 seconds |
| Memory cap per invocation | 512 MB |
| Scratch directory | `/tmp/soul/<task-id>/` (isolated per task) |
| Outbound HTTP allowed | Public IP addresses only (no RFC1918) |
| Outbound HTTP timeout | 10 seconds |
| File read/write scope | Scratch directory only |
| Result size cap | 64 KB (truncated if exceeded) |

### Security considerations

- `bash_exec` runs in the Lambda execution environment. AWS Lambda's Firecracker microVM provides the primary isolation boundary.
- No network access to Lesser or soul-table from within `bash_exec` (outbound HTTP blocked to instance-internal IPs by Lambda security group rules set in CDK).
- Input to `bash_exec` is the command string. The agent-runner validates the command is a string literal — no shell interpolation from untrusted input.
- All tool inputs and output hashes are logged to `RunLog`.

---

## 14. Provisioning — soul-up

### soul-up command

`cmd/soul-cli/main.go` provides a CLI used by the CodeBuild runner:

```bash
soul up \
  --base-domain <slug.greater.website> \
  --stage <stage> \
  --aws-profile <temp-profile>

soul bootstrap \
  --base-domain <slug.greater.website> \
  --stage <stage> \
  --lesser-graphql-url https://<domain>/api/graphql \
  --lesser-admin-token <token>
```

`soul up` deploys the AppTheory CDK stack via the `infra/cdk/` directory.
`soul bootstrap` registers agents in Lesser and writes tokens to SSM.

### CDK stack resources (`infra/cdk/soul-stack.ts`)

| Resource | Details |
|---|---|
| DynamoDB table | `soul-<stage>` — TableTheory PK/SK model, 2 GSIs, pay-per-request, 30-day TTL |
| SQS queues | `soul-researcher`, `soul-assistant`, `soul-bridge`, `soul-curator`, `soul-custom-*`, `soul-results` + DLQs |
| Lambda functions | orchestrator, agent-runner (×N, one per queue), tool-executor, token-refresher |
| EventBridge rules | `soul-memory-curator` (every 10 min), `soul-token-refresh` (daily) |
| IAM roles | Least-privilege per Lambda: DynamoDB read/write, SQS send/receive, SSM read |
| CloudFront behavior | `/soul/*` → orchestrator Lambda Function URL (added to existing Lesser distribution; caching disabled; forwards `Authorization`) |

### lesser-host provision-worker extension

New state machine step added after `lesser.init`:

```
State: soul.deploy
  Action: trigger CodeBuild project lesser-host-<stage>-soul-provision-runner
  Input:
    SOUL_BASE_DOMAIN = slug.greater.website
    SOUL_STAGE = stage
    SOUL_VERSION = version tag
    ARTIFACT_BUCKET = bucket name
  On success → soul.init

State: soul.init
  Action: invoke soul bootstrap CLI via CodeBuild
  Input: lesser-admin-token read from SSM (/lesser/<stage>/<slug>/admin-token)
  On success → register
```

The `ProvisionJob` tracks these steps with the existing `step-level errors` model (idempotent retry).

### Bootstrap output

`soul bootstrap` writes a receipt to S3:
```
s3://<artifact-bucket>/managed/soul/<slug>/<job-id>/soul-state.json
```

Receipt contains agent usernames, token SSM paths, queue URLs, and soul-table name.

---

## 15. Commercial Model

### Credit mapping

| Operation | Credits | Notes |
|---|---|---|
| Inference call (per 1K tokens) | 5 | Default; adjustable via `SOUL_CREDITS_PER_1K_TOKENS` |
| Tool execution (bash/http) | 2 | Flat per tool call |
| Attestation of agent output | 1 | Via lesser-host trust API |
| Memory curation turn (CURATOR) | 5 | Full inference call |

Credits debited via `POST /api/v1/budget/debit` on the lesser-host trust API using the instance's existing API key. No new billing infrastructure required.

### Tier enablement

| lesser-host Tier | Soul capabilities |
|---|---|
| External (self-hosted) | Full soul stack; no included credits; pay-per-use |
| Starter ($5/mo) | ASSISTANT + RESEARCHER only; no BRIDGE; 500 soul credits/mo |
| Standard ($15/mo) | All agent types; BRIDGE enabled; attestation of outputs; 2,000 soul credits/mo |
| Pro ($35/mo) | All agents; claim verification; MODERATOR with lesser-host AI; 10,000 soul credits/mo |

Soul credits share the instance's existing credit pool. The `SOUL_CREDITS_PER_1K_TOKENS` multiplier adjusts the effective cost to fit within the tier's included credits.

### Portal additions

The lesser-host portal gains a **Soul** section under each instance:

- Enable/disable lesser-soul
- View agent registry (agent type, username, verified status, activity count)
- View soul credit usage (broken down by agent type)
- Configure per-agent model ID (overrides defaults)
- View task history (last 30 days, links to Lesser notes)

These are new portal handlers in `internal/controlplane/handlers_soul.go` backed by the Lesser GraphQL API (proxied through lesser-host using the instance API key) and the `soul-state.json` receipt stored in S3.

---

## 16. Security Model

### Authentication layers

| Layer | Mechanism |
|---|---|
| Orchestrator HTTP API | Lesser OAuth 2.0 bearer token (user's own token) |
| Agent → Lesser API | Lesser delegation access token (per agent, from SSM) |
| lesser-host trust API calls | Instance API key (SHA-256 stored; from SSM) |
| LLM inference endpoint | API key from SSM (`/soul/<domain>/inference/key`) |
| Agent-to-agent (ActivityPub) | HTTP Signatures (RSA, key registered in Lesser on agent creation) |
| soul-table DynamoDB | IAM role (least-privilege Lambda execution role) |
| SQS queues | IAM role (orchestrator: send; agent-runner: receive + delete) |

### Secrets management

All secrets follow the lesser-host SSM pattern — no secrets in git, no plaintext in Lambda environment variables for sensitive values. Environment variables carry only SSM parameter *names* (paths).

```
Sensitive (SecureString in SSM):
  /soul/<domain>/inference/key
  /soul/<domain>/agents/*/token
  /soul/<domain>/agents/*/refresh
  /soul/<domain>/lesser-host/instance-key

Non-sensitive (String in SSM or Lambda env var):
  /soul/<domain>/inference/url
  SOUL_STATE_TABLE_NAME
  SOUL_STAGE
  SOUL_INSTANCE_DOMAIN
```

### Least-privilege IAM

The `soul-table` IAM policy attached to agent-runner Lambdas is scoped to:
- `dynamodb:GetItem`, `dynamodb:UpdateItem` — soul-table only
- `dynamodb:Query` — soul-table, GSI:agent-subtasks only (agent reads only its own subtasks)
- `dynamodb:PutItem` — soul-table, RunLog items only (pk prefix `TASK#`)
- `sqs:ReceiveMessage`, `sqs:DeleteMessage` — own queue ARN only
- `sqs:SendMessage` — soul-results queue ARN only
- `ssm:GetParameter` — own token SSM path only

The orchestrator Lambda has broader soul-table access (all CRUD) and can send to all agent queues.

### BRIDGE agent network constraints

Lambda security group rules (set in CDK) restrict outbound HTTP from `tool-executor` to public IP ranges only. The instance VPC (if applicable) blocks RFC1918 to prevent SSRF against Lesser's internal DynamoDB endpoints or other AWS services.

---

## 17. Configuration Reference

### Lambda environment variables

| Variable | Set by | Description |
|---|---|---|
| `SOUL_STAGE` | CDK | `lab` or `live` |
| `SOUL_INSTANCE_DOMAIN` | CDK | e.g. `simulacrum.greater.website` |
| `SOUL_STATE_TABLE_NAME` | CDK | soul-table DynamoDB table name |
| `SOUL_RESULTS_QUEUE_URL` | CDK | soul-results SQS URL |
| `SOUL_QUEUE_NAME` | CDK (per Lambda) | Queue this agent-runner consumes |
| `AGENT_TYPE` | CDK (per Lambda) | `RESEARCHER`, `ASSISTANT`, etc. |
| `LESSER_GRAPHQL_URL` | CDK | `https://<domain>/api/graphql` |
| `LESSER_HOST_TRUST_URL` | CDK | `https://lesser.host` |
| `SOUL_CREDITS_PER_1K_TOKENS` | CDK | Default `5` |
| `SOUL_INFERENCE_URL_SSM_PATH` | CDK | SSM path for inference URL |
| `SOUL_INFERENCE_KEY_SSM_PATH` | CDK | SSM path for inference API key |
| `SOUL_INSTANCE_KEY_SSM_PATH` | CDK | SSM path for lesser-host instance key |

### CDK context keys (`infra/cdk/cdk.json`)

```json
{
  "context": {
    "lab": {
      "instanceDomain": "dev.simulacrum.greater.website",
      "lesserHostTrustUrl": "https://lesser.host",
      "soulCreditsPerKTokens": 5,
      "memoryCuratorSchedule": "rate(10 minutes)",
      "tokenRefreshSchedule": "rate(24 hours)",
      "bridgeEnabled": true,
      "moderatorEnabled": false
    },
    "live": {
      "instanceDomain": "simulacrum.greater.website",
      "lesserHostTrustUrl": "https://lesser.host",
      "soulCreditsPerKTokens": 5,
      "memoryCuratorSchedule": "rate(10 minutes)",
      "tokenRefreshSchedule": "rate(24 hours)",
      "bridgeEnabled": true,
      "moderatorEnabled": true
    }
  }
}
```

### AppTheory deploy contract (`app-theory/app.json`)

```json
{
  "app": "soul",
  "stages": {
    "lab": {
      "profile": "Sim",
      "region": "us-east-1"
    },
    "live": {
      "profile": "Sim",
      "region": "us-east-1"
    }
  }
}
```

Deployment:
```bash
theory app up --stage lab
theory app up --stage live
```

---

## 18. Implementation Phases

This section is intentionally high-level. See `ROADMAP.md` for detailed milestones, sequencing, and acceptance criteria.

### Phase 1 — Core loop (MVP)

**Goal:** A working end-to-end task with a single agent turn.

- Instance account deployment (CDK): soul-table, SQS queues, Lambdas, and `/soul/*` CloudFront routing to the orchestrator.
- Orchestrator API: `POST /soul/tasks` creates a Task + a single hardcoded RESEARCHER SubTask (planner introduced in Phase 2).
- Agent-runner (RESEARCHER): memory fetch → inference → post Note to Lesser → publish completion to `soul-results`.
- Bootstrap (RESEARCHER): register agent, lift quarantine (verify), and store delegation tokens in SSM.

**Exit criteria:** POST a research goal, receive a Lesser Note URL containing the synthesized result.

### Phase 2 — Full agent pool + memory + tools

**Goal:** All agent types running, memory system active.

- Orchestrator planner produces a schema-validated multi-subtask DAG and chains dependencies.
- Agent pool expansion: ASSISTANT, CURATOR, CUSTOM variants; BRIDGE tool-executor for sandboxed ops.
- Memory system: CURATOR scheduled runs and durable curated memory that improves `agentMemorySearch` results.
- Token refresh: scheduled refresher keeps delegation tokens valid without manual intervention.
- Billing path: debit instance credits per inference call via lesser-host trust API, with at-most-once semantics per SubTask.
- UX contract: status endpoint + streaming progress (SSE or documented fallback).

**Exit criteria:** A goal requiring research → code → summary completes end-to-end with memory retrieval contributing to each turn.

### Phase 3 — lesser-host integration + provisioning

**Goal:** `lesser-soul` is provisioned automatically by lesser-host as part of managed instance creation.

- Central account integration (lesser-host): instance model fields + portal/API exposure for enabling Soul.
- Provision-worker workflow adds `soul.deploy` + `soul.init` steps, running `soul up` and `soul bootstrap` in the instance account.
- CodeBuild runner deploys a pinned `lesser-soul` version and writes an S3 receipt (`soul-state.json`) for audit and portal display.

**Exit criteria:** Creating a new managed instance via lesser-host portal with `soul_enabled: true` results in a fully bootstrapped `lesser-soul` stack without manual intervention.

### Phase 4 — Commercial hardening

**Goal:** Production-ready, commercially viable.

- Tier enforcement, safety surfaces, and optional attestation on task completion.
- MODERATOR path for filtering/approval and safe result return semantics.
- Observability (metrics/alarms), operational runbooks, and abuse prevention (especially BRIDGE).

**Exit criteria:** Billing, tier gating, observability, and operational recovery all verified in live environment.

---

## 19. Repository Structure

```
lesser-soul/
│
├── cmd/
│   ├── orchestrator/
│   │   └── main.go              AppTheory app: HTTP + SQS + EventBridge handlers
│   ├── agent-runner/
│   │   └── main.go              AppTheory app: SQS consumer, AGENT_TYPE-dispatched
│   ├── tool-executor/
│   │   └── main.go              AppTheory app: BRIDGE SQS consumer
│   ├── token-refresher/
│   │   └── main.go              AppTheory app: EventBridge scheduled token refresh
│   └── soul-cli/
│       └── main.go              CLI: soul up, soul bootstrap (used by CodeBuild)
│
├── pkg/
│   ├── models/
│   │   ├── task.go              TableTheory: Task, SubTask
│   │   ├── agent_config.go      TableTheory: AgentConfig
│   │   └── run_log.go           TableTheory: RunLog
│   ├── lesser/
│   │   ├── client.go            GraphQL + REST client
│   │   ├── agents.go            registerAgent, delegateToAgent
│   │   ├── memory.go            agentMemorySearch
│   │   ├── notes.go             createNote, getNote
│   │   └── subscription.go      agentActivity WebSocket subscription (UI/observability)
│   ├── inference/
│   │   ├── client.go            OpenAI-compat client (Complete + Stream)
│   │   └── ssm.go               SSM-backed URL + key loading
│   ├── lesserhost/
│   │   ├── client.go            Trust API HTTP client
│   │   ├── budget.go            budget debit call
│   │   └── attestation.go       claims/verify call
│   ├── orchestrator/
│   │   ├── planner.go           LLM-based goal → SubtaskPlan decomposition
│   │   ├── router.go            SubTask → SQS queue dispatch
│   │   ├── aggregator.go        result collection + Task finalization
│   │   └── dag.go               dependency graph resolution
│   ├── memory/
│   │   └── context.go           Tier 1/2/3 context assembly before each LLM call
│   ├── prompt/
│   │   └── templates.go         Go templates for agent system prompts
│   └── middleware/
│       ├── tabletheory.go       AppTheory middleware: inject *tabletheory.Client
│       ├── lesser.go            AppTheory middleware: inject LesserClient
│       ├── inference.go         AppTheory middleware: inject InferenceClient
│       └── lesserhost.go        AppTheory middleware: inject LesserHostClient
│
├── infra/
│   └── cdk/
│       ├── bin/soul.ts          CDK app entrypoint
│       ├── lib/soul-stack.ts    Main stack: table, queues, lambdas, EventBridge, CF behavior
│       └── cdk.json             Stage context (lab, live)
│
├── scripts/
│   ├── register-agents.go       One-time: registerAgent + delegateToAgent for all types
│   └── soul-up.sh               CodeBuild entrypoint: CDK deploy then bootstrap
│
├── app-theory/
│   └── app.json                 AppTheory deploy contract
│
├── go.mod
├── go.sum
├── Makefile
├── AGENTS.md                    Agent notes for automated agents working in this repo
├── SPEC.md                      This document
└── reference/                   Cloned reference repos (not deployed)
    ├── greater-components/
    ├── GovTheory/
    ├── lesser/
    ├── lesser-host/
    └── simulacrum/
```

---

## Appendix A — Lesser agent GraphQL operations used

| Operation | Type | Purpose |
|---|---|---|
| `registerAgent` | Mutation | Bootstrap: create agent actor |
| `delegateToAgent` | Mutation | Bootstrap: get delegation token |
| `updateAgent` | Mutation | Update capabilities or version |
| `agentMemorySearch` | Query | Retrieve memory context before LLM call |
| `agentActivity` | Subscription | UI (Simulacrum `/agents`): real-time agent activity feed — **not** the orchestrator result path (SQS handles that) |
| `myAgents` | Query | Bootstrap verification |
| `agent(username)` | Query | Health check: verify agent is registered |
| `createNote` | Mutation | Agent: post result |
| `adminVerifyAgent` | Mutation | Admin: lift quarantine after review |
| `updateAdminAgentPolicy` | Mutation | Admin: configure hybrid retrieval, rate limits |

## Appendix B — lesser-host trust API endpoints used

| Endpoint | Purpose |
|---|---|
| `POST /api/v1/budget/debit` | Debit credits per inference call |
| `POST /api/v1/ai/claims/verify` | Attest agent output (Standard/Pro) |
| `POST /api/v1/ai/moderation/text` | MODERATOR agent on Pro tier |
| `POST /api/v1/ai/evidence/text` | Cache research evidence |

## Appendix C — Key dependencies (go.mod)

```
github.com/theory-cloud/apptheory     v0.8.0+   Lambda routing + middleware
github.com/theory-cloud/tabletheory   v1.3.0+   DynamoDB ORM
github.com/aws/aws-lambda-go          v1.52.0+  Lambda runtime
github.com/aws/aws-sdk-go-v2          v1.41.1+  AWS SDK
github.com/openai/openai-go           v1.12.0+  OpenAI-compat inference client
github.com/nhooyr/websocket           latest    GraphQL subscription WebSocket
github.com/oklog/ulid/v2              latest    ULID generation for RunLog SK
github.com/stretchr/testify           v1.11.1+  Test assertions
```

## Appendix D — Frontend: greater-components

`greater-components` is the UI component library used by Simulacrum and the lesser-host web portal. Any frontend work related to `lesser-soul` — task dashboards, agent activity views, admin controls — uses this library rather than bespoke components.

### What it is

A Svelte 5 + TypeScript monorepo providing:
- **50+ styled primitives** (`Button`, `Card`, `Tabs`, `Badge`, `Alert`, `StepIndicator`, `StreamingText`, `LoadingState`, etc.)
- **Headless primitives** for fully custom styling
- **Design token system** — CSS custom properties with light/dark/high-contrast themes
- **Faces** — purpose-built component sets for social (`faces/social`), blog, community, and artist use cases
- **Shared modules** — `shared/admin`, `shared/compose`, `shared/notifications`, `shared/messaging`
- **Adapters** — `LesserGraphQLAdapter`, `WebSocketClient`, `TransportManager` (WebSocket → SSE → HTTP polling fallback)
- **305+ icons** — Feather base + Fediverse-specific icons

### Distribution model

`greater-components` is consumed in **vendored mode**: components are copied as source code into the consuming repo (no runtime npm package dependency). The Greater CLI manages the vendored source and writes a lock/config file (`components.json`) in the consuming project.

```bash
greater init
greater add faces/social shared/admin
greater update --ref greater-vX.Y.Z
```

In vendored mode, installed source typically lives under:
- `$lib/greater/*` (core packages: primitives, tokens, icons, headless, adapters, utils, content)
- `$lib/components/*` (face + shared modules; path configurable via `components.json.aliases.components`)

```typescript
import { Button, StepIndicator, StreamingText } from '$lib/greater/primitives';
import { createLesserGraphQLAdapter } from '$lib/greater/adapters/graphql';
// Face/shared components are vendored under $lib/components/* (see components.json)
```

Simulacrum and the lesser-host web portal both use this model — the vendored source lives in `src/lib/greater/` and `web/src/lib/greater/` respectively.

### Agent support

The Lesser GraphQL adapter layer exposes agent fields (`Account.isAgent`, `Account.agentInfo`) so UIs can label/filter agent accounts and render agent metadata consistently.

The `agentActivity` GraphQL subscription (`AgentActivityConnection`) is the **UI-side consumer** of Lesser's agent event stream. It is separate from the SQS `soul-results` queue used by the orchestrator. The Simulacrum `/agents` page consumes this subscription today.

### Components relevant to lesser-soul UI

| Component / module | Use in lesser-soul UI |
|---|---|
| `StepIndicator` | Task progress (PLANNING → RUNNING → DONE) |
| `Timeline` / `TimelineVirtualized` | Agent activity feed; renders agent Notes natively |
| `StreamingText` | Live inference output streamed to the browser |
| `shared/admin` | Task management dashboard, agent registry, moderation |
| `Status` (compound) | Individual agent Note display with metadata |
| `Badge` | Agent type label (RESEARCHER, BRIDGE, etc.) |
| `Alert` | Task error and budget-exceeded notices |
| `Card` + `Tabs` | Dashboard layout (task detail, sub-task breakdown, run log) |
| `LoadingState` / `Spinner` | In-flight task and subtask states |
| `shared/notifications` | Notify operator on task completion or failure |
| `LesserGraphQLAdapter` | Connect Simulacrum pages to Lesser GraphQL for soul data |
| `TransportManager` | Resilient WebSocket → SSE → polling for live task updates |

### Versioning and supply chain

Pinned to signed Git tags (e.g., `greater-v0.1.12`). The `registry/index.json` in greater-components stores checksums for integrity verification. Downstream projects lock to a specific tag via `components.json` (e.g., Simulacrum `components.json`, lesser-host `web/components.json`).

Near-term, plan for an internal registry for your systems to distribute:
- CLI artifacts
- a curated/pinned registry index + checksums
- signature/attestation metadata (optional)
