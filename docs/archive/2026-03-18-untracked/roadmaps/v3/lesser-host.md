# lesser-host — lesser-soul v3 roadmap

Date: 2026-03-04  
Spec: `SPEC-v3-draft.md` (v3.0 DRAFT)

## Scope

Implement the managed v3 protocol surface in `lesser-host`:

- v3 registry data + registration publishing (`channels`, `contactPreferences`)
- Channel provisioning + verification (ENS/email/phone)
- Communication gateway: inbound provider webhooks → instance notifications; outbound agent requests → provider dispatch
- ENS CCIP-Read gateway + OffchainResolver contract ops
- Communication boundaries + reputation signals

## Non-negotiables (carried from v2 + v3)

- **TableTheory** for DynamoDB state (single-table patterns)
- **AppTheory** for service runtime + HTTP handlers
- **lesser-body holds no provider creds**; comm API is the single egress for providers
- **One gateway ingress** (webhooks terminate at `lesser-host`, not instances)

## Milestones

### LH-M0 — Interface contracts + v3 draft reconciliation

**Deps:** none | **Parallel:** first | **Spec refs:** §6.3, §6.4, §10, §12

Freeze cross-repo payloads/errors and resolve known v3 draft ambiguities before implementation forks.

**Deliverables:**
- JSON schemas + fixtures for:
  - `communication:inbound` (and `communication:outbound` if implemented) notification payloads
  - `/api/v1/soul/comm/send` request/response + error envelope
  - `/api/v1/soul/agents/{agentId}/channels` response shape
- Written decisions (or spec patch) for:
  - per-agent mailbox credential storage (SSM vs DynamoDB)
  - “sent mail” strategy (`agent://email/sent`)
  - attachment metadata MVP contract

**Acceptance:**
- [ ] Fixtures can be consumed by `lesser` and `lesser-body` without ad-hoc parsing
- [ ] Every cross-repo error case has a stable `code` + human message contract

---

### LH-M1 — Data foundation: channels + preferences + comm logs

**Deps:** v2 soul models present | **Spec refs:** §3, §4, §13

Add v3 persistence primitives to the existing soul single-table model.

**Deliverables:**
- New TableTheory models (names indicative; follow existing patterns under `internal/store/models/`):
  - `soul_agent_channel.go` (PK `SOUL#AGENT#{agentId}` / SK `CHANNEL#{type}`)
  - `soul_agent_contact_preferences.go` (PK `SOUL#AGENT#{agentId}` / SK `CONTACT_PREFERENCES`)
  - `soul_agent_comm_activity.go` (PK `SOUL#AGENT#{agentId}` / SK `COMM#{ts}#{id}`)
  - `soul_agent_comm_queue.go` (PK `COMM#QUEUE#{agentId}` / SK `MSG#{scheduled}#{messageId}`)
  - `soul_agent_ens_resolution.go` (or equivalent) for `ENS#NAME#{ensName}` records
- Reverse-lookup indexes (single-table “index items” or equivalent TableTheory patterns):
  - email → agentId
  - phone → agentId
  - ensName → agentId (if not stored directly as `ENS#NAME#...`)

**Acceptance:**
- [ ] Model keying matches `SPEC-v3-draft.md` §13 (or a documented, query-equivalent variation)
- [ ] Unit tests cover key generation + reverse lookup + TTL serialization

---

### LH-M2 — Registration file v3: build, validate, publish

**Deps:** LH-M1 | **Spec refs:** §3, §4, Appendix F

Extend the registration file pipeline to emit and validate v3 additions.

**Deliverables:**
- v3 schema support:
  - `version: "3"`
  - `channels` object (ens/email/phone)
  - `contactPreferences`
- Registration publishing updates:
  - republish on channel or preference changes
  - preserve v2 fields and invariants (append-only boundaries, etc.)
- Migration behavior:
  - existing v2 agents can be read as v2; v3 publishing is opt-in when channels exist

**Acceptance:**
- [ ] Registration file round-trip (build → canonicalize/sign → verify) passes
- [ ] API reads always return the latest published registration file URI

---

### LH-M3 — Public discovery APIs: channels + resolve + search extensions

**Deps:** LH-M1 | **Spec refs:** §11, §12.1, §11.4

Add public read surfaces for channel-based discovery.

**Deliverables:**
- New public endpoints (per spec):
  - `GET /api/v1/soul/agents/{agentId}/channels`
  - `GET /api/v1/soul/agents/{agentId}/channels/preferences`
  - `GET /api/v1/soul/resolve/ens/{ensName}`
  - `GET /api/v1/soul/resolve/email/{emailAddress}`
  - `GET /api/v1/soul/resolve/phone/{phoneNumber}`
- Search extensions:
  - `GET /api/v1/soul/search?channel=email&channel=phone`
  - `GET /api/v1/soul/search?ens=<ensName>`

**Acceptance:**
- [ ] Reverse lookups succeed for managed souls (and do not leak secrets)
- [ ] Search filters are indexed and do not require full-table scans

---

### LH-M4 — Channel provisioning: ENS + email (MVP) + optional phone

**Deps:** LH-M1, LH-M2 | **Spec refs:** §7

Implement provisioning flows and state transitions for channel records.

**Deliverables (MVP = ENS + email):**
- ENS provisioning writes/updates `ENS#NAME#...` records so names resolve via the gateway
- Email provisioning via Migadu:
  - create mailbox
  - store mailbox metadata + credential secret reference
  - update channel record + registration file channels object
- Verification:
  - email: token flow (`/channels/email/verify`)
  - phone: code flow (`/channels/phone/verify`) if phone is enabled

**Acceptance:**
- [ ] Provisioning is idempotent on `(agentId, channelType)`
- [ ] Provisioning updates are reflected in:
  - channel records
  - `GET /channels`
  - published registration file

---

### LH-M5 — Outbound comm API + provider dispatch (email MVP)

**Deps:** LH-M3, LH-M4 | **Spec refs:** §6.4, §12.3, §10.2, §8

Expose an agent-authenticated API for outbound communication, with authoritative enforcement.

**Deliverables:**
- `POST /api/v1/soul/comm/send`
  - auth: agent OAuth token
  - enforce: agent active, channel verified, rate limits, communication_policy boundaries
  - log: activity record + send status
  - dispatch: provider (SMTP/API) for email; Telnyx for SMS/voice when enabled
- `GET /api/v1/soul/comm/status/{messageId}`
  - returns delivery status + timestamps

**Acceptance:**
- [ ] Boundary violations block sends with stable error codes
- [ ] Delivery status is queryable and does not require provider polling for basic states

---

### LH-M6 — Inbound comm gateway: provider webhooks → instance delivery

**Deps:** LH-M3, L-M1 (instance endpoint), LH-M1 | **Spec refs:** §6.2–§6.9, §12.4

Implement inbound ingestion and routing through a dedicated worker.

**Deliverables:**
- `comm-worker` Lambda (new `cmd/comm-worker` entrypoint) with:
  - recipient resolution (email/phone → agentId → instance)
  - preference enforcement (availability windows; rate limits)
  - delivery: call instance notification delivery endpoint with instance API key
  - logging: activity log + counters
- Provider-facing webhooks:
  - `POST /webhooks/comm/email/inbound` (Migadu)
  - `POST /webhooks/comm/sms/inbound` (Telnyx)
  - `POST /webhooks/comm/voice/inbound` + `/webhooks/comm/voice/status` (Telnyx) (optional sequencing)

**Acceptance:**
- [ ] Inbound email/SMS results in a `communication:inbound` notification in the correct instance
- [ ] Messages outside availability windows are queued (not delivered immediately)
- [ ] Email rate-limit overflow bounces with a clear notice

---

### LH-M7 — ENS gateway service (CCIP-Read)

**Deps:** LH-M1, LH-M3 | **Spec refs:** §5.2, §12.5, §14

Implement the HTTP gateway used by the OffchainResolver contract.

**Deliverables:**
- `GET https://ens-gateway.lessersoul.ai/resolve`
  - parse EIP-3668 `name` + `data`
  - fetch resolution material from soul registry
  - build ENS record responses per §5.1/§5.3
  - sign response with authorized signer key
- `GET /health` for liveness
- Caching strategy + abuse protection (rate limits; minimal DB reads)

**Acceptance:**
- [ ] Gateway responses verify against the resolver’s signer
- [ ] Name lifecycle states (active/suspended/archived) resolve deterministically

---

### LH-M8 — OffchainResolver contract + deployment tooling

**Deps:** LH-M7 (gateway behavior defined) | **Spec refs:** §14

Ship the Solidity contract and an operator workflow for mainnet deployment and updates.

**Deliverables:**
- `OffchainResolver.sol` implementing `IOffchainResolver`
- Tests covering:
  - `OffchainLookup` revert shape
  - signature verification in `resolveWithProof`
  - owner-only setters
- Deployment scripts + runbook:
  - deploy
  - set gateway URL
  - rotate signer
  - transfer ownership to admin Safe

**Acceptance:**
- [ ] Contract tests pass and match the CCIP-Read client expectations
- [ ] Runbook covers signer rotation without downtime

---

### LH-M9 — Phone/SMS/voice provisioning + billing integration

**Deps:** LH-M4, LH-M5, LH-M6 | **Spec refs:** §7.4–§7.6

Enable opt-in phone channels and integrate usage with credits/billing.

**Deliverables:**
- Telnyx number search/order/release flow + per-agent mapping
- SMS inbound/outbound and voice event handling (sequenced: SMS first, voice later)
- Usage metering:
  - ledger entries for SMS/calls
  - enforce “insufficient credits” failure mode

**Acceptance:**
- [ ] Phone provisioning and deprovisioning are safe and auditable
- [ ] Usage is metered and blocks sends when credits are insufficient

---

### LH-M10 — Reputation: communication dimension + signals

**Deps:** LH-M1, LH-M5, LH-M6 | **Spec refs:** §9, §13.6

Add the communication dimension and compute the new signal counts.

**Deliverables:**
- Reputation model update:
  - new `communication` dimension
  - new `signalCounts` fields (emails/sms/calls, violations, response stats)
- Aggregation logic updates (worker/job) to compute:
  - response rate
  - avg response time
  - bounce rates / spam reports (if captured)
- Config surface:
  - publish weights including `w_communication` (spec mentions `/api/v1/soul/config`)

**Acceptance:**
- [ ] Communication signals update after comm activity
- [ ] Boundary violations impact integrity/communication per policy

---

### LH-M11 — Portal + ops polish (channels UI, monitoring, abuse)

**Deps:** LH-M3–LH-M10 | **Spec refs:** §6, §7, §8, §9

Make v3 operable for humans (principals/operators).

**Deliverables:**
- Portal UI:
  - view/provision/verify channels
  - edit contact preferences
  - view comm activity + delivery statuses
- Monitoring + abuse controls:
  - dashboards/alarms for webhook failures, bounce rates, queue depth
  - rate-limit tuning hooks
  - audit log entries for provisioning and sends

**Acceptance:**
- [ ] Operator can debug “why didn’t this message deliver?” end-to-end from the portal
- [ ] Basic abuse/spam protections are measurable and adjustable
