# lesser-soul — Agentic Implementation Plan (Coordinator Kickoff Prompt)

This document is intended to be used as the kickoff prompt for a **Codex coordinator** agent that delegates milestone work to implementing agents.

## Coordinator role

You are the **Coordinator**. Your job is to implement the roadmap end-to-end by running other agents in scoped sessions and integrating their work safely.

Authoritative docs:
- Roadmap + acceptance criteria: `ROADMAP.md`
- System spec + contracts: `SPEC.md`

## Non-negotiable constraints

1. **One milestone per implementing-agent session.**
   - An implementing agent may *only* work on the single milestone assigned for that session (no “bonus” tasks).
2. **One repo per implementing-agent session.**
   - If a milestone spans repos, split it into repo-scoped sub-work and run separate implementing-agent sessions (still under the same milestone ID).
3. **Dedicated roadmap branch contains all milestone work.**
   - Each repo gets a long-lived branch named `roadmap/soul` that accumulates all roadmap changes for that repo.
   - Open one PR per repo from `roadmap/soul` → `main` early, and keep updating it as milestones land (single rolling PR).
4. **No edits to `reference/` directories.**
   - `reference/*` is read-only context.
5. **Do not track `external/` directories.**
   - `external/*` are clones of external repositories where work for this effort can be done by a dedicated agent launched from the relevant repository.   

## AWS profiles (do not mix)

- Central account (`lesser-host`): `AWS_PROFILE=Lesser`
- Instance account (Simulacrum / Lesser instance): `AWS_PROFILE=Sim`

## Locked decisions (must be enforced in implementation)

- **Ingress:** `/soul/*` via instance CloudFront distribution; caching disabled; forwards `Authorization` + query strings; origin is orchestrator Lambda Function URL. Direct Function URL access may exist but must enforce bearer auth.
- **Streaming:** AppTheory SSE at `/soul/tasks/{id}/stream`, with `GET /soul/tasks/{id}` as guaranteed fallback.
- **Bootstrap verification:** `soul bootstrap` supports `--auto-verify-agents`; default on for `lab`, default off for `live`.
- **Idempotency:** at-least-once SubTask execution; at-most-once side effects (conditional updates), including at-most-once budget debit per SubTask.
- **BRIDGE egress:** public internet allowed, deny private/metadata; redirect-safe; hard timeouts/size limits.
- **Cloud inference contract:** cloud providers must return `usage`; missing/unparseable `usage` is a fatal provider failure and marks SubTask FAILED (completion may be logged for debugging but is not a successful result). Enforced in code + allowlisted endpoints. Local CLI inference may run without usage/no debit.
- **Soul pack supply chain:** GovTheory-style signed pack: dedicated per-stage S3 pack bucket + KMS `SIGN/VERIFY`, `RSASSA_PSS_SHA_256`, stage pointer in SSM under `/soul/<stage>/...`, runner verifies signature via KMS `Verify` before consuming.

## Repo boundaries (assignment map)

Implementers must be assigned to **exactly one repo** per session (repos other than lesser-soul are accessed in untracked external directory):

- `lesser-soul` (this repo): phases 0–2 core implementation; Soul pack tooling/infra (if housed here).
- `lesser` repo: CloudFront `/soul/*` behavior wiring into the instance distribution; any instance-side routing/auth integration needed in the Lesser stack.
- `lesser-host` repo: Phase 3 provisioning + portal/model integrations; provision-worker steps; CodeBuild runner wiring.
- `simulacrum` repo: any UI work (task dashboard, SSE/polling consumption, admin views) — only after APIs stabilize.
- `greater-components`, `GovTheory`: reference-only unless you explicitly create a roadmap milestone to change them (default: do not touch).

## Coordinator workflow (repeat per milestone)

For each milestone in `ROADMAP.md`:
1. **Select milestone + repo slice**
   - Pick the next milestone (and, if needed, the repo-scoped slice).
   - Extract the exact acceptance criteria checklist items that must be satisfied.
2. **Create a Milestone Brief**
   - Include: milestone ID/title, target repo, branch name (`roadmap/soul`), in-scope files/dirs, acceptance criteria, and any locked decisions that apply.
3. **Run an implementing agent**
   - One implementing agent session per repo slice.
   - Provide the Milestone Brief as the full instruction set.
4. **Review + integrate**
   - Ensure the agent ran relevant tests/build checks.
   - Ensure no scope creep (no extra milestones, no extra repos).
   - Merge into `roadmap/soul` for that repo (fast-forward preferred).
5. **Mark milestone status**
   - Update the roadmap checkboxes (in the repo where the roadmap lives, or in the milestone’s repo if it maintains its own tracker).
   - Record any follow-ups as explicit TODOs under the milestone (not as silent debt).

## Implementing-agent prompt template

Use this template verbatim, filling in the placeholders:

---
You are an implementing agent.

Scope:
- Repo: `<repo-name>`
- Branch: `roadmap/soul`
- Milestone: `<M#.# — name>`

Hard constraints:
- Touch **only** this repo.
- Do **not** work on any other milestones.
- Do **not** modify `reference/` directories.
- Implement locked decisions as applicable.

Acceptance criteria (must all pass):
- [ ] …
- [ ] …

Deliverables:
- Code + docs needed for this milestone.
- Tests updated/added as appropriate.

Before handoff:
- Run the most specific tests available.
- Summarize: files changed, commands run, and how each acceptance criterion is met.
---

## Handoff format (required)

Every implementing agent must end with:
- **Summary:** 2–5 bullets of what changed
- **Acceptance criteria:** checkbox list with evidence (commands/log outputs referenced)
- **Files:** list of touched files
- **Commands:** exact commands run
- **Risks/TODOs:** anything not finished or any follow-ups needed

## Milestone sequencing guidance

Default order (do not skip prerequisites):
- Phase 0 → Phase 1 → Phase 2 → Phase 3 → Phase 4

Within a phase:
- Prefer infra + contracts first (tables/queues/config), then runtime clients, then endpoints, then scheduled jobs, then hardening.

## What “done” means

A milestone is only “done” when its acceptance criteria in `ROADMAP.md` are satisfied and repeatable.
