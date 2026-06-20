---
name: implement-milestone
description: Use to execute a scoped soul change — feature branch off main, commits (usually one or two per milestone; soul work is typically small), PR review, merge to main. Deploy itself is handled by deploy-namespace-site.
---

# Implement a milestone

This skill moves soul work through code, review, and merge to `main`. soul changes are typically small — a namespace-addition at `/v2` with its docs, a FEP draft update, a dependency bump, a site-content refresh, a prospective-pinning alignment. The minimal pipeline reflects this: no separate `enumerate-changes` or `plan-roadmap` — scope-need / specialist walks produce enough structure.

## Hard preconditions

Do not start without all of the following:

- **A specific change named**, from `scope-need` or a specialist-walk output (`evolve-namespace`, `manage-fep-submission`, `deploy-namespace-site`)
- **Clean working tree on `main`** at a known-green commit
- **MCP tools healthy.** Call `memory_recent` first.
- **Dependencies install cleanly** — `npm install` or `pnpm install` in `cdk/` succeeds
- **`cdk synth --context stage=lab`** succeeds for the current tree state
- **FaceTheory SSG build** (`npm run build` or equivalent in `cdk/`) succeeds for the current tree state
- **The change is in-mission** — not scope growth; not implementation work that belongs in host
- **For namespace-content changes**: the `evolve-namespace` walk is complete
- **For FEP changes**: the `manage-fep-submission` walk is complete
- **For CDK / deploy changes**: the `deploy-namespace-site` walk is complete
- **Advisor-dispatched milestones** have the principal's authorization from `review-advisor-brief` recorded

If any precondition fails, stop.

## Branch and PR setup

One feature branch per milestone. One PR per milestone. Usually one or two commits.

- **Branch name**: descriptive, scoped. Observed patterns: `codex/<topic>`, `feat/<topic>`, `issue/<N>-<topic>`, `chore/<maintenance>`.
- **Branched from**: `main` at a known-green commit.
- **PR target**: `main`.
- **PR title**: clear. Present-tense lowercase welcomed (`update theory pins and patch cdk advisories`) or Conventional Commits (`docs(fep): incorporate editorial feedback round 2`).
- **Open PR as draft** with the milestone goal and an unchecked task list.

PR description template:

```markdown
## Milestone
<short-name> — <goal from scope-need or specialist walk>

## Classification
<namespace-addition / FEP-submission / static-site / CDK-deploy / dependency-maintenance / prospective-pinning / docs>

## Surfaces affected
<enumerated — cdk/site/static/ns/... / cdk/site/faces.ts / cdk/lib/... / docs/ / roadmaps/ / app-theory/>

## Specialist walks referenced
- Namespace: <...>
- FEP: <...>
- Deploy / CDK: <...>
- Framework consumption: <idiomatic / reported upstream>

## Tasks
- [ ] <task>

## Validation
- `npm install` / `pnpm install` (cdk/)
- `cdk synth --context stage=lab`
- SSG build (`npm run build` or equivalent)
- For namespace changes: direct curl test of the `/ns/*` URL in lab and live after deploy
- For FEP changes: Codeberg editorial review conformance check

## Stage rollout plan (handoff to deploy-namespace-site)
- [ ] Merged to main
- [ ] Deployed to lab
- [ ] Lab verification: namespace URL resolves correctly, site pages render
- [ ] Deployed to live
- [ ] Live verification: namespace URL resolves correctly, CloudFront cache behaves as expected

## Cross-repo coordination
<required / none; host if namespace semantics change, lesser if namespace URL serialization affected>

## Advisor-brief authorization (if applicable)
<summary from review-advisor-brief>
```

## The per-task loop

For each task in the milestone, usually in this simple order:

1. **Read the scoped change.** Confirm acceptance and planned commit shape.
2. **`memory_recent`** — refresh context.
3. **Make the change.** Only files in the enumerated paths.
4. **For namespace content**: confirm the new document lives at `cdk/site/static/ns/agent-attribution/v<N>` (new version) — **never modifies existing `/v1`** unless the change is an unambiguous error correction scoped through `evolve-namespace` with documented reasoning.
5. **For site HTML / FaceTheory SSG changes**: confirm the build produces expected output; CSP-compliant; no inline scripts.
6. **For CDK changes**: `cdk synth --context stage=lab` succeeds; CloudFront behaviors preserve the `/ns/*` direct-pass-through contract.
7. **For FEP changes**: content aligns with the editorial process's current state; authorship + CC0 scope decisions documented.
8. **For prospective-pinning bumps**: `app-theory/app.json` aligns with the broader Theory Cloud stack's current versions; no Go code added.
9. **For dependency bumps**: run `npm install` / `pnpm install` fresh; `cdk synth` and SSG build still succeed; no AGPL-incompatible dependencies pulled in.
10. **Commit.** Clear subject. Explain *why* in the body for namespace, FEP, CDK / CloudFront, or prospective-pinning changes. Never `--no-verify`. Never `--amend` a pushed commit.
11. **Push.** Never force-push.
12. **Check task off** in PR description.
13. **`memory_append`** only when something worth remembering — namespace-versioning decision, FEP editorial subtlety, CloudFront-behavior pattern, prospective-pinning drift resolution, advisor-brief pattern. Routine commits aren't memory material. Five meaningful entries beat fifty log-shaped ones.

## The mission rule enforced at commit time

- **Every commit must be in-mission.** Scope growth → `scope-need`.
- **Namespace content at `/v1` never mutates silently.** The `evolve-namespace` walk gates all namespace changes.
- **No runtime code** (Lambda, backend service) added to the repo.
- **No Go code** against prospective AppTheory pinning.
- **No proprietary blobs or AGPL-incompatible dependencies.**
- **No weakening of `RemovalPolicy.RETAIN`** on the namespace bucket.
- **No weakening of `/ns/*` CloudFront behavior contract** (direct pass-through, long immutable cache, CORS open, correct content-type).
- **No framework patches locally** (FaceTheory, CDK, or prospective AppTheory).
- **No silent changes to `docs/archive/`** — historical record is preserved.

## If tests / build go red mid-milestone

- **Do not** add a "fix build" commit touching unrelated code.
- **Do** stop, investigate, surface.
- If failure caused by your most recent commit, revert with a new revert commit and re-plan.

## Finishing the milestone (PR side)

When all tasks committed and pushed:

1. Re-verify `cdk synth --context stage=lab` succeeds.
2. Re-verify SSG build succeeds.
3. Promote PR out of draft.
4. Update PR description: check task boxes (stage-rollout boxes check as `deploy-namespace-site` runs).
5. Request required review.
6. **Leave merging to a reviewer.**

## Hand off to deploy-namespace-site

Once merged to `main`, `deploy-namespace-site` owns:

- `theory app up --stage lab` deploy (operator-run)
- Lab verification: namespace URL curl, site page render, CloudFront cache-behavior check
- `theory app up --stage live` deploy (operator-authorized)
- Live verification
- Documentation updates to `docs/spec-lessersoul-ai-inventory.md` if infrastructure state changed

`implement-milestone` does not run deploy commands. Its output is a merged PR + handoff.

## What this skill will not do

- Will not implement more than one milestone per run.
- Will not accept scope growth as a task.
- Will not merge PRs — required review.
- Will not skip required review for "small" changes.
- Will not run deploy commands — that's `deploy-namespace-site`.
- Will not skip specialist walks for namespace / FEP / CDK work.
- Will not mutate `/v1` namespace content silently.
- Will not add Go code against prospective AppTheory pinning.
- Will not add runtime state, Lambda, or backend service to this repo.
- Will not force-push, amend pushed commits, or rewrite history.
- Will not introduce AGPL-incompatible dependencies or proprietary blobs.
- Will not weaken `RemovalPolicy.RETAIN` on the namespace bucket.
- Will not weaken `/ns/*` CloudFront behavior contract.
- Will not set timeouts on CDK commands.
- Will not delete `docs/archive/` content.
- Will not act on advisor briefs without `review-advisor-brief` authorization.