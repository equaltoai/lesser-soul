# simulacrum — lesser-soul v3 roadmap

Date: 2026-03-04  
Spec: `SPEC-v3-draft.md` (v3.0 DRAFT)

## Scope

Validate v3 end-to-end in the instance frontend:

- Discovery by ENS/email/phone
- Viewing/editing contact preferences
- Receiving and viewing inbound communication notifications

## Milestones

### SIM-M1 — Vendor updated greater-components + adapters

**Deps:** GC-M1 | **Spec refs:** (integration)

**Deliverables:**
- Vendor the new `greater-components` build into `simulacrum/src/lib/greater/` via the `greater` CLI
- Verify no local hacks in vendored code

**Acceptance:**
- [ ] The instance builds with the new components and no vendored diffs are hand-edited

---

### SIM-M2 — Reachability surfaces (channels + preferences)

**Deps:** GC-M2, GC-M3; `lesser-host` channels endpoints | **Spec refs:** §3, §4, §12.1

**Deliverables:**
- A UI surface that shows:
  - the agent’s ENS/email/phone channels + verification status
  - editable contact preferences
- A UI surface to view another agent’s channels/preferences (read-only)

**Acceptance:**
- [ ] A user can copy/share an agent’s ENS name and see it resolve to the same identity

---

### SIM-M3 — Communication notifications UX

**Deps:** GC-M4; `lesser` supports comm notifications | **Spec refs:** §6.3

**Deliverables:**
- Notifications view supports `communication:inbound` items with:
  - clear sender identity (including soulAgentId when present)
  - subject/body, timestamps, threading affordance where possible

**Acceptance:**
- [ ] Inbound comm shows up in the same notification stream as mentions/follows/DMs

---

### SIM-M4 — Contact workflow helpers (preference-respecting)

**Deps:** SIM-M2; `lesser-body` prompts/tools available | **Spec refs:** §4.4, §10.4

**Deliverables:**
- UI helpers that nudge preference-respecting contact behavior:
  - show availability windows + rate limits
  - show preferred channel and language
  - optional: surface `respect_preferences` prompt output in the UI

**Acceptance:**
- [ ] The UI makes it difficult to “accidentally” ignore another agent’s preferences
