---
name: manage-fep-submission
description: Use when a change touches Fediverse Enhancement Proposal (FEP) authoring, submission to Codeberg, editorial response, or governance decisions (authorship, CC0 scope, versioning policy). FEPs govern Fediverse-wide adoption of the agent-attribution concepts the namespace describes; this skill walks the repo's role in the editorial process without overreaching into the Codeberg process itself.
---

# Manage an FEP submission

Fediverse Enhancement Proposals (FEPs) are the Fediverse's community-governance mechanism for proposing and adopting Fediverse-wide conventions. The editorial process lives at **Codeberg** (`codeberg.org/fediverse/fep` or its current canonical home). soul is the author-of-record for FEPs related to agent attribution.

This skill walks every FEP-related change in this repo — drafting, iteration, governance decisions, editorial response — while respecting that Codeberg owns the editorial process itself.

## The FEP surface in this repo (memorize)

- **`roadmaps/`** — active FEP work plans (e.g., `roadmaps/issue-3-fep-agent-attribution.md`)
- **FEP draft content** — currently lives in `roadmaps/` or in `docs/` depending on editorial stage; each draft carries its own versioning
- **Governance decisions** — recorded in GitHub Issues and ADRs in this repo (authorship naming, CC0 licensing scope, versioning policy alignment with namespace paths)
- **FEP submission path** — Codeberg's FEP repo; submission via PR or per the editorial process's current convention

## When this skill runs

Invoke when:

- A new FEP is being drafted in this repo
- An existing FEP draft needs iteration (content evolution, scope refinement, authorship addition)
- An editorial-feedback round from Codeberg requires a response
- A governance decision needs recording (CC0 scope, authorship, versioning alignment)
- An FEP is ready for submission to Codeberg
- An FEP has been accepted, finalized, or rejected by the editorial process
- A scope-need conversation surfaces a new FEP-submission candidate

## Preconditions

- **The FEP in question is named and scoped.** "Work on FEPs" is too vague; "update `roadmaps/issue-3-fep-agent-attribution.md` to incorporate the second editorial round's feedback on the `delegated_by` property's cardinality" is concrete.
- **The current editorial state is known.** Is the FEP pre-submission, in editorial review, finalized, or rejected? Is there an FEP number assigned yet?
- **MCP tools healthy**, `memory_recent` first — FEP work is long-running; prior context matters.

## The four-step walk

### Step 1: Confirm the editorial state

Before making content changes:

- **Pre-submission** — the FEP draft lives entirely in this repo. Revisions happen here via normal PR flow.
- **Submitted, in editorial review** — the FEP exists on Codeberg with a number (or a pending-number placeholder). This repo's copy may be synchronized or may be the local working version; either way, Codeberg is the authoritative location post-submission.
- **Finalized / accepted** — the FEP has reached final status on Codeberg. This repo retains historical authoring context; further changes go through Codeberg's process for amendments.
- **Withdrawn / rejected** — the FEP did not reach acceptance. This repo retains the draft + rationale for historical record.

The current state determines what changes are appropriate and where they land.

### Step 2: Record governance decisions explicitly

FEPs carry governance metadata that needs to be decided and recorded:

- **Authorship** — named authorship (typically the principal or co-authored with other contributors). Changes to authorship (adding a co-author, correcting an attribution) are governance events, documented in the repo.
- **CC0 licensing scope** — the Codeberg FEP process typically requires CC0 for FEP text so the community can adopt freely. Confirm CC0 scope explicitly; never commit FEP text that would be non-CC0 without an explicit process exception.
- **Versioning policy alignment** — if the FEP proposes a specific URL (e.g. `spec.lessersoul.ai/ns/agent-attribution/v1`), the FEP's versioning policy aligns with soul's namespace versioning convention. Document the alignment.
- **Scope statements** — what the FEP covers, what it explicitly does not cover, what's deferred to future FEPs.

Each governance decision commits to the repo (as part of an issue discussion, an ADR, or prominent documentation in `roadmaps/`). No silent decisions.

### Step 3: Shape the content update

For content changes:

- **Pre-submission**: normal repo-PR flow. Iteration is fast; content lives in `roadmaps/` or `docs/`.
- **In editorial review**: content changes synchronize with Codeberg. Typically soul's copy follows Codeberg; if this repo's draft is ahead, submit the updates through Codeberg's process, not here.
- **Post-finalization**: content in this repo transitions to historical-record mode. Changes are rare and documented as historical edits.

Changes that affect **the namespace content** coordinate with `evolve-namespace`. A FEP round may require a `/v2` namespace publication; those two skills coordinate.

### Step 4: Align with the Fediverse-peer community

FEPs exist for community adoption. That means:

- **Engagement with Codeberg discussion** — editorial rounds generate comments; soul responds via the editorial process.
- **Peer feedback from ActivityPub implementations** — Mastodon, Pleroma, Misskey, GoToSocial, and others may comment on the FEP or on their own adoption plans. Track relevant feedback here (issues, ADRs) even though the editorial process is at Codeberg.
- **Reference implementation** — host's registry implementation is soul's reference implementation of the FEP's semantics. The FEP references this as appropriate.
- **Adoption tracking** — as peers implement the FEP, adoption signals (how many implementations, which versions) inform future FEP evolution.

## The output note

```markdown
## FEP-submission audit: <FEP name / number>

### Editorial state
<pre-submission / submitted in editorial review / finalized / withdrawn / rejected>

### FEP number (if assigned)
<FEP-XXXX or pending>

### Proposed change
<concrete description>

### Governance decisions recorded
- Authorship: <named authors; co-authorship additions documented>
- CC0 scope: <confirmed for FEP text>
- Versioning policy alignment: <aligns with soul's namespace /v<N> convention>
- Scope statements: <...>

### Content-change shape
- Pre-submission (normal repo-PR): <yes / no>
- In-editorial-review (synchronize with Codeberg): <yes / no; Codeberg update path>
- Post-finalization (historical-record edit): <yes / no>

### Namespace-coordination (if applicable)
- Namespace change required: <no / yes — invoke evolve-namespace>
- Namespace version path: </v1 / /v2 / new>

### Fediverse-peer coordination
- Codeberg editorial round response required: <yes / no; response shape>
- Known peer feedback: <summarized>
- Reference implementation (host) aligned: <yes / pending>

### Cross-repo coordination
- host: <required (reference implementation alignment) / not required>
- lesser: <typically not required unless FEP changes lesser's serialization; coordinate if yes>
- body: <typically not required>
- soul (this repo): <the content change is here>

### Test coverage
- FEP draft syntactically valid markdown: <confirmed>
- CC0 header present: <confirmed for FEP text>
- Authorship metadata current: <confirmed>

### Proposed next skill
<implement-milestone if audit clean and content-change is repo-local; evolve-namespace if namespace content changes as part of this FEP round; scope-need if audit surfaces scope growth; investigate-issue if audit reveals a submission or editorial-process issue>
```

## Refusal cases

- **"Skip Codeberg; publish the FEP from soul directly."** Refuse. The Fediverse's FEP process lives at Codeberg; soul's role is to submit, not to substitute.
- **"Finalize the FEP on our own; Codeberg's editorial process is slow."** Refuse. Finalization is Codeberg's call, not soul's.
- **"Change authorship after submission without coordination."** Refuse. Authorship changes post-submission follow the editorial process.
- **"Publish FEP text under a non-CC0 license for this submission."** Refuse (unless explicitly authorized with documented reasoning; Codeberg's convention is typically CC0).
- **"Commit FEP text that references a private or proprietary implementation."** Refuse. FEP text should reference publicly-verifiable implementations (host is the reference implementation).
- **"Add a new author to the FEP without checking with them."** Refuse. Named authorship has consent implications.
- **"Silently change FEP scope after submission."** Refuse. Scope changes go through the editorial process.
- **"Skip recording the governance decision; we'll remember."** Refuse. All governance decisions commit.

## Persist

Append every meaningful FEP-submission event — draft milestones, governance decisions, editorial rounds, finalization / withdrawal / rejection, peer-adoption signals worth remembering. Include: FEP name, editorial state at the time, key decision or change, outcome.

FEP work is long-running; memory continuity is especially valuable here. Five meaningful entries is a floor — editorial rounds and governance events are inherently memorable.

## Handoff

- **Audit clean, repo-local content change** — invoke `implement-milestone`.
- **Audit clean, namespace-change required as part of FEP** — invoke `evolve-namespace` first to plan the namespace side, then back here for the FEP content update.
- **Audit requires host reference-implementation coordination** — coordinate via the `host` steward through the principal before proceeding.
- **Audit reveals an editorial-process issue** (submission blocker, format issue at Codeberg) — document and coordinate with Codeberg; may require updating the FEP draft in this repo.
- **Audit surfaces scope growth** — revisit `scope-need`.
- **Audit surfaces a framework / tooling gap** (e.g. FEP drafting would benefit from a FaceTheory-rendered preview pattern) — `coordinate-framework-feedback`.
- **Advisor-dispatched FEP brief** — `review-advisor-brief` gates the authorization.