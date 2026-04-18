---
name: investigate-issue
description: Use when a user reports a bug, regression, or unexpected behavior — the namespace URL not resolving, wrong content-type on `/ns/*`, site page rendering issue, CloudFront cache misbehavior, CDK deploy failure, FEP submission hiccup, AppTheory-pinning drift. Runs before any fix is proposed. Produces an investigation note, not a patch.
---

# Investigate an issue

Investigation comes before implementation. soul's surface is small but the investigation dimensions are distinctive: a permanent public URL, CloudFront + S3 topology with two separate behaviors, a CDK stack with `RemovalPolicy.RETAIN` on one bucket, a FaceTheory SSG pipeline, a prospective AppTheory pinning, and the FEP editorial process interacting with the repo.

## Start with memory

Call `memory_recent` first. Scan for prior investigations in the same area — CloudFront behavior edge cases, FaceTheory SSG oddities, namespace-resolution issues observed in the wild, AppTheory-pinning decisions, FEP-editorial findings. soul is a low-frequency repo; prior context matters.

## Capture the claim precisely

Record the user's report literally, then extract:

- **Symptom** — what was observed, verbatim where possible
- **Surface** — namespace URL resolution (`/ns/agent-attribution/v1`) / site HTML pages / CloudFront caching / CDK deploy / FaceTheory SSG build / AppTheory pinning / FEP editorial interaction
- **Reported by** — an ActivityPub implementation operator seeing a resolution failure? An internal observer during a deploy? A FEP editor?
- **Recent changes** — recent commits, recent deploys, recent CDK changes, recent dependency bumps
- **Reproduction path** — the specific URL, the client used, the response observed (headers + body)
- **Expected vs actual** — JSON-LD content-type and shape, HTML-site behavior, CDK synth output

## Ground the investigation

Your first structural questions are always:

1. **Is this a namespace-resolution issue?** If the `/ns/agent-attribution/v1` URL is returning wrong content, wrong content-type, unexpected caching, or failing to resolve — elevate. This is the frozen-forever contract; regressions here ripple across the Fediverse.
2. **Is this a site-HTML issue?** Landing page, FEP docs page, or other site content. Usually lower-stakes than namespace issues.
3. **Is this a CloudFront-behavior issue?** The two behaviors (`/ns/*` direct pass-through; default HTML rewrites via CloudFront Functions) are separately configured; bugs often localize to one.
4. **Is this a CDK-deploy issue?** `cdk synth` or `theory app up` producing unexpected output or failing partway.
5. **Is this a FaceTheory SSG issue?** The build step producing wrong output, failing, or regressing after an upgrade.
6. **Is this a prospective-AppTheory-pinning drift?** `app-theory/app.json` out of sync with the broader Theory Cloud stack.
7. **Is this an FEP editorial issue?** Submission status, editorial feedback, content-version alignment between this repo's drafts and Codeberg.
8. **Is the symptom in soul, in a CloudFront / S3 / ACM configuration, in FaceTheory, in AWS itself, or in a downstream ActivityPub peer?** Many reported "soul bugs" are actually downstream consumers misinterpreting JSON-LD or caching stale content from a proxy. Confirm before accepting the symptom as a soul bug.

## Evidence before hypotheses

Gather before theorizing:

- `git log` since the last known-good state (for CDK, `cdk/`; for site content, `cdk/site/`; for namespace content, `cdk/site/static/ns/`)
- `git blame` on specific lines
- `cdk synth` output for the affected stage
- CloudFront distribution cache-hit rate, 4xx/5xx rate, edge-location errors (through the user)
- S3 bucket state — do the expected objects exist with the expected content-type headers?
- ACM certificate validity (expiry, chain health)
- For namespace URL issues: the exact HTTP response (headers, body) observed by the reporter
- For site-HTML issues: the rendered page, the FaceTheory SSG build log
- For FEP issues: the Codeberg issue / discussion context
- `query_knowledge` for cross-repo context — FaceTheory patterns, AppTheory / TableTheory versions, sibling equaltoai context

If `memory_recent` or `query_knowledge` returns an auth error, stop — investigating a permanent-URL regression without context continuity compounds risk.

## The specialist-routing question

Every investigation answers: **which specialist skill, if any, should handle this?**

- **Namespace content, versioning, or contract concerns** → `evolve-namespace`
- **FEP submission, editorial response, governance** → `manage-fep-submission`
- **CDK deploy, CloudFront behaviors, S3 buckets, ACM certificate** → `deploy-namespace-site`
- **Framework awkwardness** (FaceTheory, AppTheory, CDK) → `coordinate-framework-feedback`
- **Advisor-originated brief** → `review-advisor-brief`
- **None** — routes through standard `scope-need` → `implement-milestone` flow (soul's minimal pipeline)

## Rank hypotheses by evidence

List theories in descending order of support:

1. **Hypothesis** — one sentence
2. **Evidence for** — commits, HTTP-response observations, CDK output, FaceTheory build logs
3. **Evidence against** — what would be true if this were wrong
4. **Verification step** — the cheapest test to prove or disprove it

## Output: the investigation note

```markdown
## Reported symptom
<verbatim>

## Dimensions
- Surface: <namespace URL / site HTML / CloudFront / CDK deploy / FaceTheory SSG / AppTheory pinning / FEP editorial>
- Reported by: <ActivityPub operator / internal observer / FEP editor>
- Recent deploys or commits: <...>
- Reproduction: <URL, headers, response>

## Specialist elevation check
<normal investigation / elevate to evolve-namespace / manage-fep-submission / deploy-namespace-site / coordinate-framework-feedback / review-advisor-brief>

## What is definitely true
<verified facts — HTTP responses, S3 object state, CDK output, FEP issue thread>

## Fix-locus verdict
<fix here (soul) / fix upstream (FaceTheory, AppTheory, CDK) / configuration issue outside code / AWS / downstream-consumer issue>

## Hypotheses (ranked)
1. <hypothesis> — evidence: <...>
2. <...>

## Verification step
<the one thing to run next>

## Proposed next skill
<fix directly / scope-need / evolve-namespace / manage-fep-submission / deploy-namespace-site / coordinate-framework-feedback / review-advisor-brief / none — cross-repo or external report>
```

## Persist

Append only if the investigation surfaces something worth remembering — a CloudFront behavior subtlety, a FaceTheory SSG edge case, a namespace-resolution pattern observed from a specific ActivityPub implementation, a FEP editorial finding worth continuity, a prospective-pinning drift signal. Routine "typo" findings aren't memory material. Five meaningful entries beat fifty log-shaped ones.

## Handoff rules

- **Namespace-integrity-suspected** — elevate to `evolve-namespace`.
- **FEP editorial issue** — `manage-fep-submission`.
- **Deploy / stack / CloudFront / S3 issue** — `deploy-namespace-site`.
- **Framework awkwardness** — `coordinate-framework-feedback`; report upstream, don't patch.
- **Advisor brief** — `review-advisor-brief`.
- **Small, contained fix** — `scope-need` → `implement-milestone`.
- **Downstream-consumer issue** — document, consider whether the namespace / site contract needs any defensive hardening.
- **AWS / configuration outside code** — report to user; not a soul-code change.
