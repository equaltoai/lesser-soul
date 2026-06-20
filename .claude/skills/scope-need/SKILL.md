---
name: scope-need
description: Use when a user brings a new capability, feature request, or enhancement need for soul in vague terms. Interviews conversationally and produces a scoped-need document. Applies Gate 1 (soul-mission alignment, with heavy bias toward "this probably belongs in host") and Gate 2 (narrowest scope). Most soul scopings resolve to redirects.
---

# Scope a need

A need arrives fuzzy. A feature arrives sharp — or, in soul's case, a feature often arrives with a redirect verdict. This skill is the conversation that turns fuzzy into sharp, with a strong default toward redirecting to `host` (the registry implementation repo).

## Your posture

You are interviewing, not pitching. soul is deliberately thin — a public-spec publisher and static-site host. Most proposals that feel like they belong in soul actually belong in host (the registry implementation). The scoping question is two-part:

1. **Is this genuinely soul-mission work — namespace publication, FEP submission, static-site content, CDK topology, prospective-AppTheory-readiness — or is it implementation work that belongs in host?**
2. **If it's in-mission, what is the narrowest possible scope that preserves namespace stability, static-asset-only posture, AGPL coverage, prospective-AppTheory-readiness, and idiomatic FaceTheory consumption?**

The default for Gate 1 for most proposals is **"probably belongs in host"**. Redirect is the common outcome, not the exception.

## Start with memory and the architecture

- **Read `README.md`, `docs/README.md`, `docs/spec-lessersoul-ai-inventory.md`, and `roadmaps/*`** for canonical content.
- `memory_recent` — has this need or adjacent work been scoped before? soul scoping often repeats (namespace-addition, FEP-submission, pinning-bump).
- `query_knowledge` — does lesser-host already implement or plan to implement this? does FaceTheory's pattern inform the answer?

If tools are unavailable, surface it and ask the user to re-auth.

## The interview

Ask, one or two at a time:

1. **Who is asking and why now?** ActivityPub peer feedback, internal observation, FEP editorial requirement, operator request, advisor-dispatched brief, principal-direct?
2. **What problem does it solve?** Current pain, not speculative.
3. **Which surface does it touch?**
   - Namespace content at `/ns/*` (requires `evolve-namespace` walk)
   - FEP submission or editorial response (requires `manage-fep-submission` walk)
   - Static-site content (landing page, FEP docs pages)
   - CDK / CloudFront / S3 topology (requires `deploy-namespace-site` walk)
   - FaceTheory consumption pattern
   - Prospective AppTheory / TableTheory pinning in `app-theory/`
   - `docs/` content (README, inventory, archive)
   - `roadmaps/` (new roadmap document)
4. **Is this a namespace change?** Tightened scrutiny if yes. `/v1` never mutates; additions go to `/v2`.
5. **Is this a FEP submission step?** Tightened scrutiny for editorial-process integrity.
6. **Does this implicitly require runtime code?** If yes, probably belongs in host.
7. **Is this preservation, evolution, or growth?** Preservation is welcome; growth in a thin repo is usually scope-creep.
8. **What does success look like?** Observable. For namespace work, "this JSON-LD expansion produces the expected properties." For FEP, "the FEP submission reaches the expected editorial stage." For deploy, "the stack applies cleanly with all behaviors preserved."
9. **What is explicitly out of scope?**

## The two gating questions

### Gate 1: Is this soul-mission work?

Five possible verdicts:

1. **Yes — namespace addition at a new version or FEP-related content.** A new `/v2` namespace, a new FEP draft, an FEP editorial response. Route through `evolve-namespace` or `manage-fep-submission`.
2. **Yes — static-site content or docs update.** Landing page content, `docs/` updates, `roadmaps/` additions.
3. **Yes — CDK / deploy / topology work.** CloudFront behavior refinement (without loosening `/ns/*` contract), stack improvement, ACM coordination. Route through `deploy-namespace-site`.
4. **Yes — prospective-readiness maintenance.** `app-theory/app.json` pinning bump to keep alignment with Theory Cloud stack. AGPL-compatible dependency bumps.
5. **No — belongs in host or elsewhere.** The most common verdict. Implementation work, runtime state, registry semantics, on-chain work, governance payload preparation — all belong in host. Produces a redirect document.

### Gate 2: What is the narrowest possible scope?

If Gate 1 passed, the next question is how to deliver with the smallest change.

Prefer:

- Namespace additions **at a new version path** over mutations of existing version
- FEP draft iterations that track the editorial process faithfully
- Static-site updates that preserve CloudFront behavior contracts
- Dependency bumps within current major versions
- Prospective-pinning bumps that match the broader Theory Cloud stack

Avoid:

- Mutations to `/v1` (refused)
- Adding Go code against the prospective AppTheory pinning (scope growth)
- New CDK constructs that weaken the `/ns/*` behavior's strict contract (direct pass-through, long immutable cache, CORS open)
- New runtime surfaces (Lambda, backend service, API)
- Dependencies with AGPL-incompatible licenses

## Output: the scoped-need document

### For Gate 1 verdict "soul-mission work":

```markdown
# Scoped Need: <short name>

## Background
<one paragraph>

## Driver
<ActivityPub peer / operator / FEP editorial / advisor-dispatched / principal-direct>

## Problem
<what is broken, missing, or painful today>

## Surface affected
<namespace / FEP / static-site / CDK / prospective-pinning / docs / roadmaps>

## Classification
<namespace-addition / FEP-submission / static-site / CDK-deploy / dependency-maintenance / prospective-pinning / docs>

## Narrowest-scope proposal
<the smallest change that addresses the need>

## What this need explicitly does not cover
<bounded scope; watch for implementation-creep>

## Success criteria
<observable, testable>

## Specialist routing
- Namespace: <not touched / walk via evolve-namespace>
- FEP: <not relevant / walk via manage-fep-submission>
- Deploy / CDK: <not touched / walk via deploy-namespace-site>
- Framework consumption: <idiomatic / awkwardness via coordinate-framework-feedback>
- Advisor brief: <n/a / review via review-advisor-brief>

## Consumer impact
<ActivityPub peers / operators / downstream services in the equaltoai family>

## AGPL posture
<no change / confirmed AGPL-compatible>

## Open questions
<unresolved>
```

### For Gate 1 verdict "belongs in host or elsewhere":

```markdown
# Redirect: <short name>

## Background
<one paragraph>

## Why this doesn't belong in soul
<soul is deliberately thin — specification and static-site publisher; this is implementation / runtime-state / registry work, which lives in host>

## Appropriate owner
<host (most common) / lesser / body / greater / theory cloud framework / scoping with the principal>

## Path for the requesting user
<route the conversation to the host steward (or sibling steward), defer, or open a parallel scoping conversation>

## Recommended next step
<specific handoff>
```

## Persist before handoff

Append only if the scoping surfaces a recurring pattern — a redirect category (a kind of work that repeatedly gets proposed here but belongs in host), a namespace-addition pattern worth remembering, a FEP editorial subtlety. Routine completions aren't memory material. Five meaningful entries beat fifty log-shaped ones.

## Handoff

- **In-mission, namespace** — invoke `evolve-namespace`.
- **In-mission, FEP** — invoke `manage-fep-submission`.
- **In-mission, deploy / CDK** — invoke `deploy-namespace-site`.
- **In-mission, static-site or docs or prospective-pinning** — invoke `implement-milestone` directly (soul's minimal pipeline — no separate `enumerate-changes` skill since the work is typically small).
- **Framework awkwardness** — `coordinate-framework-feedback`.
- **Advisor-dispatched scope** — `review-advisor-brief` already ran; output includes the principal's authorization.
- **Out-of-scope (belongs in host etc.)** — redirect document is the handoff.
- **Resolved to "no change needed"** — record and stop.
- **User defers** — record and stop.