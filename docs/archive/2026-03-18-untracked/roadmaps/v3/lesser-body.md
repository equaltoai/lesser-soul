# lesser-body — lesser-soul v3 roadmap

Date: 2026-03-04  
Spec: `SPEC-v3-draft.md` (v3.0 DRAFT)

## Scope

Expose v3 communication surfaces to agents via MCP, without direct provider credentials:

- Tools: email/SMS/voice + identity
- Resources: `agent://channels`, inbox/sms/voicemail views
- Prompts: compose/handle/respect preferences helpers

## Milestones

### LB-M0 — Tool/resource contract freeze

**Deps:** none | **Parallel:** first | **Spec refs:** §10

Lock the exact MCP tool/resource shapes and error envelopes expected by clients.

**Deliverables:**
- Tool schemas for:
  - email (`send`, `read`, `search`, `reply`, `delete/archive`)
  - SMS/voice (`sms_send`, `sms_read`, `phone_call`, `voicemail_read`)
  - identity (`whoami`, `lookup`, `verify`)
- Resource schemas for `agent://channels` and inbox-like resources
- A stable error mapping from `lesser-host` comm API errors → MCP tool errors

**Acceptance:**
- [ ] The tool catalog matches `SPEC-v3-draft.md` §10.1 and remains backwards compatible for existing tools

---

### LB-M1 — Identity surfaces: whoami + lookup + resources

**Deps:** LB-M0 | **Spec refs:** §10.1, §10.3, §11, §12.1

Implement identity tools and the `agent://channels` resource first; these unblock UX across tools.

**Deliverables:**
- `identity_whoami`:
  - returns full identity including channels + preferences
- `identity_lookup`:
  - supports query by ENS name, agentId, or email (and phone if enabled)
- Resources:
  - `agent://channels`
  - `agent://channels/preferences`

**Acceptance:**
- [ ] Identity results are consistent with `GET /api/v1/soul/agents/{agentId}/channels`
- [ ] ENS/email lookup works for managed souls end-to-end

---

### LB-M2 — Outbound communication tools (email MVP)

**Deps:** LB-M0, LH-M5 | **Spec refs:** §10.2, §12.3, §8

Implement outbound send paths via `lesser-host` comm API.

**Deliverables:**
- Tools:
  - `email_send`
  - `email_reply`
  - (later) `sms_send`, `phone_call`
- Advisory boundary check:
  - read boundaries from registration file and warn before calling host API
- Delivery status UX:
  - surface host `messageId` + status in tool result

**Acceptance:**
- [ ] Tool calls never require provider credentials locally
- [ ] Boundary violations are surfaced to the LLM as actionable errors

---

### LB-M3 — Inbound “inbox” tools backed by instance notifications

**Deps:** L-M2, LB-M0 | **Spec refs:** §10.2 (inbound tools), §6.3

Expose inbox-like reads by filtering existing instance notifications.

**Deliverables:**
- Tools:
  - `email_read` (supports `unreadOnly`, `since`, `limit`, `folder?` as an abstraction)
  - `email_search`
  - `email_delete` (maps to dismiss/archive semantics; define behavior clearly)
  - `sms_read`
  - `voicemail_read` (if voice is enabled)
- Resources:
  - `agent://email/inbox`
  - `agent://email/sent` (requires a defined “sent” strategy)
  - `agent://sms/messages`
  - `agent://voicemail`

**Acceptance:**
- [ ] Inbound emails/SMS delivered to the instance appear in these tools without polling external providers
- [ ] `unreadOnly` behavior is deterministic (document how it maps to notification state)

---

### LB-M4 — Prompts: compose/handle/respect preferences

**Deps:** LB-M1–LB-M3 | **Spec refs:** §10.4, §4.4

Add the spec-defined prompts to improve agent behavior and reduce boundary violations.

**Deliverables:**
- `compose_email`
- `handle_inbound`
- `respect_preferences`

**Acceptance:**
- [ ] Prompts reference the agent’s declared boundaries/preferences when available
