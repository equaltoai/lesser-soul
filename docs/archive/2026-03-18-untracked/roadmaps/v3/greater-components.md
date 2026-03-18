# greater-components — lesser-soul v3 roadmap

Date: 2026-03-04  
Spec: `SPEC-v3-draft.md` (v3.0 DRAFT)

## Scope

Provide typed clients + reusable UI for v3 reachability features (channels, preferences, comm notifications),
so `simulacrum` can validate the experience without local hacks.

## Milestones

### GC-M1 — Contract pin + adapter generation for v3 endpoints

**Deps:** `lesser-host` implements discovery endpoints | **Spec refs:** §12.1, §12.2, §11.4

**Deliverables:**
- Update pinned contracts/specs used for codegen to include:
  - channels endpoints
  - resolve endpoints
  - search extensions (channel/ens filters)
- Regenerate adapters + registry index

**Acceptance:**
- [ ] Generated adapters compile and are published through the normal package workflow

---

### GC-M2 — Typed client utilities (soul channels + ENS resolution)

**Deps:** GC-M1 | **Spec refs:** §5, §11

**Deliverables:**
- A typed client wrapper for:
  - `GET /api/v1/soul/agents/{agentId}/channels`
  - `GET /api/v1/soul/resolve/*`
- ENS resolution helpers for `*.lessersoul.eth`:
  - resolve text records via standard ENS libraries (browser-safe approach)
  - fall back to `lesser-host` resolve endpoints where appropriate

**Acceptance:**
- [ ] `simulacrum` can resolve an ENS name and display channels without custom fetch logic

---

### GC-M3 — UI components: channels + preferences

**Deps:** GC-M2 | **Spec refs:** §3, §4

**Deliverables:**
- Components:
  - Channels display (ENS/email/phone + verification status)
  - Contact preferences editor/viewer (availability windows, languages, rate limits)
  - “Best way to contact” helper UI driven by preferences

**Acceptance:**
- [ ] Components are usable in `simulacrum` without forking vendored code

---

### GC-M4 — UI components: communication notification rendering

**Deps:** `lesser` supports `communication:*` notifications | **Spec refs:** §6.3

**Deliverables:**
- Notification renderers/cards for:
  - inbound email
  - inbound SMS
  - voice events/voicemail (if enabled)

**Acceptance:**
- [ ] Notifications are legible, safe, and consistent with existing notification UX patterns
