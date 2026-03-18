# lesser (instance) — lesser-soul v3 roadmap

Date: 2026-03-04  
Spec: `SPEC-v3-draft.md` (v3.0 DRAFT)

## Scope

Enable the “existing notification system” to carry inbound (and optionally outbound) communication events delivered by
`lesser-host`’s comm-worker.

## Milestones

### L-M0 — Contract freeze for communication notifications

**Deps:** none | **Parallel:** first | **Spec refs:** §6.3

Lock the exact notification payload shape and validation rules on the instance side.

**Deliverables:**
- A strict request schema for `POST /api/v1/notifications/deliver` (or whatever internal endpoint is chosen)
- Size limits + sanitization rules for `subject` and `body`
- Idempotency strategy keyed by `messageId`

**Acceptance:**
- [ ] A fixture payload from `lesser-host` is accepted without transformation
- [ ] Invalid payloads fail with structured errors (no partial writes)

---

### L-M1 — Authenticated internal delivery endpoint

**Deps:** L-M0 | **Spec refs:** §6.3

Add/extend an internal endpoint that allows `lesser-host` to deliver notifications into an instance.

**Deliverables:**
- `POST /api/v1/notifications/deliver` (path per spec, or a documented internal variant)
  - auth: `Authorization: Bearer <instance-api-key>` (key is issued/managed by `lesser-host`)
  - create a notification record for the target user/agent
- Audit logging of delivery calls (for incident response)

**Acceptance:**
- [ ] Calls with an invalid/absent instance API key are rejected
- [ ] Delivery is idempotent on `messageId`

---

### L-M2 — Persist + list `communication:*` notifications

**Deps:** L-M1 | **Spec refs:** §6.3, §10.2

Ensure comm notifications behave like first-class notifications for listing/filtering.

**Deliverables:**
- Store `communication:inbound` notifications with:
  - `channel` (email|sms|voice)
  - `from` fields (address/phone + optional `soulAgentId`)
  - `subject`, `body`, `receivedAt`, `messageId`, `inReplyTo?`
- Optional: store `communication:outbound` to support “sent” views
- Ensure standard notifications APIs can filter by type (existing `types[]` filters, if present)

**Acceptance:**
- [ ] `GET /api/v1/notifications` returns comm notifications (and filters work)
- [ ] Notification bodies render safely in existing clients (no HTML injection)

---

### L-M3 — Read-state + threading affordances (optional but enables better MCP UX)

**Deps:** L-M2 | **Spec refs:** §10.1–§10.3 (tool expectations)

Support the MCP layer’s need for inbox-like behavior.

**Deliverables:**
- A stable mapping from `messageId` to instance notification ID
- Optional: basic “threading” helper (group by `inReplyTo`) for UI/tools
- Ensure “dismiss” semantics don’t prevent future tool access if the tool expects “read history”

**Acceptance:**
- [ ] `lesser-body` can implement `email_read(unreadOnly)` without fragile heuristics
