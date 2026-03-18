# lesser-soul v3.0 Implementation Roadmap (stack-wide)

Date: 2026-03-04  
Audience: EqualtoAI stack maintainers

**Spec (source of truth):** `SPEC-v3-draft.md` (v3.0 DRAFT, dated 2026-03-04)  
**Baseline prereq:** `SPEC.md` v2 implementation is present (v3 is strictly additive).

## Goal

Implement v3 **reachability** primitives for managed souls:

- Declared + verifiable **channels**: ENS name, email, optional phone (SMS + voice)
- **Contact preferences**: availability windows, rate limits, languages, first-contact policy
- A centralized **communication gateway** in `lesser-host` (provider webhooks ↔ comm-worker ↔ instance notifications)
- **MCP communication tools** in `lesser-body` (agents can read/respond without holding provider creds)
- ENS **CCIP-Read** resolution for `<localId>.lessersoul.eth` via an off-chain resolver + gateway

## Repos that require development (v3)

| Repo | Why it changes | Primary v3 deliverables | Detailed roadmap |
|------|----------------|--------------------------|------------------|
| `lesser-host/` | Managed implementation (Layer 4) + registry (API) | Channels + preferences data/models/APIs; comm-worker + webhooks + outbound comm API; provider integrations; ENS gateway + OffchainResolver ops; comm reputation signals | `roadmaps/v3/lesser-host.md` |
| `lesser/` | Instance ingestion surface | Authenticated internal notification delivery endpoint for `communication:*`; persistence + listing | `roadmaps/v3/lesser.md` |
| `lesser-body/` | Agent MCP surface | MCP tools/resources/prompts for email/SMS/voice + identity; bridge to `lesser-host` comm APIs; read inbound comm via instance notifications | `roadmaps/v3/lesser-body.md` |
| `greater-components/` | Typed clients + reusable UI | Adapter updates + new UI components for channels/preferences + comm notification rendering | `roadmaps/v3/greater-components.md` |
| `simulacrum/` | Product integration + validation | Wire new clients/components; end-to-end UX for reachability + inbound comm in instance UI | `roadmaps/v3/simulacrum.md` |

## Cross-repo contracts (freeze early)

v3 introduces new “interfaces between repos” that must be stable before parallel work is efficient.

1. **Instance notification delivery** (`lesser-host` → `lesser`)
   - Spec: `SPEC-v3-draft.md` §6.3
   - Endpoint: `POST /api/v1/notifications/deliver`
   - Payload: `type: "communication:inbound"` (and optionally `communication:outbound` for “sent” surfaces)
   - Requirements: authenticated (instance API key), idempotent on `messageId`, safe body size limits

2. **Outbound comm API** (`lesser-body` → `lesser-host`)
   - Spec: `SPEC-v3-draft.md` §6.4, §12.3, §10.2
   - Endpoint: `POST /api/v1/soul/comm/send`
   - Requirements: agent OAuth auth; structured errors (boundary/rate-limit/status); returns `messageId`

3. **ENS CCIP-Read gateway** (ENS client → `lesser-host` gateway)
   - Spec: `SPEC-v3-draft.md` §5.2, §12.5, §14
   - Endpoints: `GET /resolve`, `GET /health`
   - Requirements: deterministic encoding of records; signed responses; caching; DoS protection

## Suggested delivery phases (so teams can parallelize)

This is a sequencing suggestion; subsystem roadmaps contain the full milestone detail.

### P0 — Interface + spec hygiene

Freeze payload shapes + error contracts, and resolve v3 draft ambiguities before large implementation starts.

- `lesser-host`: LH-M0
- `lesser`: L-M0
- `lesser-body`: LB-M0

### P1 — Data + registration file v3

Make it possible to store and publish v3 `channels` + `contactPreferences` (even before providers/gateway exist).

- `lesser-host`: LH-M1, LH-M2, LH-M3

### P2 — Email MVP (inbound + outbound)

End-to-end: provision mailbox → inbound webhook → instance notification → agent reads → agent replies.

- `lesser-host`: LH-M4, LH-M5, LH-M6 (email subset)
- `lesser`: L-M1, L-M2
- `lesser-body`: LB-M1, LB-M2, LB-M3 (email subset)

### P3 — ENS resolution (CCIP-Read) + reverse lookup

ENS name works as a primary discovery mechanism; email reverse lookup works for managed souls.

- `lesser-host`: LH-M7, LH-M8
- `greater-components`/`simulacrum`: GC-M2 (ENS client + lookup), SIM-M2 (UX wiring)

### P4 — Phone/SMS/voice expansion

Opt-in phone numbers; inbound/outbound SMS; voice (optional sequencing after SMS).

- `lesser-host`: LH-M6 (SMS/voice), LH-M9 (billing/usage)
- `lesser-body`: LB-M2/LB-M3 (SMS/voice tools)
- `simulacrum`: SIM-M3 (UI surfaces)

### P5 — Reputation + boundaries + product polish

Communication reputation signals + boundary enforcement + UI surfaces + monitoring/abuse.

- `lesser-host`: LH-M10, LH-M11
- `greater-components`: GC-M3, GC-M4
- `simulacrum`: SIM-M4

## System-level acceptance (v3 “done”)

Use these as end-to-end “definition of done” checks across repos.

1. **Registration publishes v3 channels**
   - A managed soul’s registration file includes `version: "3"` with `channels.ens.name` and `channels.email.address`
   - `GET /api/v1/soul/agents/{agentId}/channels` returns channels + preferences + verification status

2. **Inbound email → instance notifications**
   - Sending an email to `<localId>@lessersoul.ai` results in a `communication:inbound` notification in the target instance
   - The agent can read it via `notifications_read` and `email_read`

3. **Outbound email respects boundaries**
   - `email_send`/`email_reply` via MCP calls `POST /api/v1/soul/comm/send`
   - Requests that violate declared `communication_policy` are rejected with a structured error

4. **Availability + rate limits enforce correctly**
   - Inbound outside declared availability windows queues and later delivers
   - Inbound exceeding declared rate limits bounces (email) / drops per policy (SMS)

5. **ENS resolution works**
   - `agent-alice.lessersoul.eth` resolves via CCIP-Read to the records listed in `SPEC-v3-draft.md` §5.1

6. **Reverse lookup works (managed souls)**
   - `GET /api/v1/soul/resolve/email/{email}` returns the correct soul identity
   - (If phone is implemented) `GET /api/v1/soul/resolve/phone/{phone}` works likewise

7. **Communication reputation signals update**
   - After inbound/outbound comm, the reputation record shows updated communication signalCounts/dimension

## Known spec/implementation questions to close early

Track these explicitly (or patch the spec) before implementation hardens:

- **Mailbox credentials storage**: `SPEC-v3-draft.md` contains conflicting guidance (DynamoDB vs SSM) for per-agent mailbox passwords.
- **“Sent mail” surface**: v3 lists `agent://email/sent` but doesn’t fully specify whether sent items are delivered as notifications, queried from providers, or read from a host-side log.
- **Attachment handling**: the inbound email model mentions “attachments metadata” informally; decide what is required for MVP.
