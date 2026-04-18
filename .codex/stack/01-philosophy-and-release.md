# Philosophy, release, and stage discipline

soul is **deliberately thin** — a public-spec publisher and static-site host, not an implementation. The philosophy follows from that: **namespace-stable, FEP-respectful, forward-ready, static-asset-disciplined, AGPL-true, framework-feedback-conscious.**

## Namespace stability is the mission

The single most load-bearing thing about soul is that `https://spec.lessersoul.ai/ns/agent-attribution/v1` resolves correctly, forever, with stable semantics.

- **`/v1` never mutates.** Once published, the namespace document at `cdk/site/static/ns/agent-attribution/v1` is effectively frozen. Corrections of outright typos or meta-information are evaluated cautiously; anything semantic moves to a new version.
- **Breaking changes move to `/v2`, `/v3`, etc.** Parallel paths. The old version continues to resolve for peers that haven't migrated.
- **Additive changes within a version** are acceptable only when they genuinely don't change the meaning of anything previously published. JSON-LD has well-defined rules about context evolution; follow them.
- **CloudFront caching is aggressive** for namespace documents (`Cache-Control: max-age=31536000, immutable`). This matches the "frozen forever" contract.
- **CORS is open** for namespace documents so ActivityPub implementations from any origin can resolve them.
- **Content-type is `application/ld+json`** — direct JSON-LD response, no HTML wrapper, no JavaScript redirect, no inline docs interleaved.

Every change that touches namespace content runs through the `evolve-namespace` skill. That skill refuses mutations of `/v1`; versioned additions are the accepted shape.

## The implementation lives elsewhere

The registry **implementation** — the code that mints soul tokens, stores off-chain state, serves registration and lookup APIs, coordinates Safe-ready governance — lives in `host` (lesser-host). soul publishes the namespace contract; host implements the semantics.

Historical note: earlier iterations of this repo held implementation code (see `docs/archive/2026-03-18/` which preserves the original SPEC.md and ROADMAP.md). That code moved to host in March 2026. soul is now deliberately thin.

When a change proposal feels like it belongs in the registry implementation — a new on-chain function, a new off-chain query, a new mutation endpoint — the steward's posture is: **this probably belongs in host, not soul**. Route through `scope-need` with that default; the `host` steward owns implementation scoping.

## FEP submission is a repo concern

The Fediverse Enhancement Proposal (FEP) process runs at Codeberg. soul is the submitter-of-record for FEPs related to agent attribution:

- **Active work** lives in `roadmaps/` (e.g., `roadmaps/issue-3-fep-agent-attribution.md`).
- **Governance decisions** — CC0 scope, named authorship, versioning policy — are recorded in issues and ADRs in this repo.
- **Submission path** — FEPs draft here, then submit to Codeberg's FEP repository through the FEP editorial process.
- **FEP numbers** are assigned by the FEP editorial process; soul's authoring work happens pre-assignment.
- **Post-submission shepherding** — once submitted, FEPs go through community discussion and editorial review. soul's role is to engage in that process and update the FEP as the editorial process requires.

The `manage-fep-submission` skill walks FEP-related changes.

## AppTheory / TableTheory readiness (forward-looking)

`app-theory/app.json` pins AppTheory v0.19.1 + TableTheory v1.5.1 **prospectively**:

- **No Go code exists yet.** No `cmd/`, no `internal/`, no handlers, no models.
- **The pinning is forward-looking** — if and when the registry implementation migrates from host to soul, the AppTheory contract is ready.
- **`app-theory/init.md`** records the initialization instructions for that future migration.
- **Do not bitrot the pinning.** When AppTheory or TableTheory bump in the broader Theory Cloud stack, soul's pinning bumps in sync where feasible. Keeping the readiness current is part of the stewardship.
- **Do not let the pinning overreach.** Don't add Go code here "while we're at it" — that's scope growth. The pinning is ready-state, not active-state.

The prospective-readiness posture is a deliberate forward-looking pattern the repo uses; preserve it.

## AGPL and static-asset discipline

soul is AGPL-3.0, consistent with the equaltoai family:

- **No proprietary blobs in the tree.**
- **Contributor-origin transparency** per repo convention.
- **AGPL-compatible dependencies only.** FaceTheory, AWS CDK, TypeScript toolchain — all license-compatible.
- **Public-release posture** — the repo is public; releases (via `main` tags where applicable) are on GitHub.

The **static-asset discipline**: soul ships static files + static-site-generated HTML + a CDK stack. No Lambda, no runtime state, no backend service. When a proposal would add runtime code, the default is: *this probably doesn't belong here*.

## Framework-feedback reciprocity with Theory Cloud

soul consumes FaceTheory canonically for its SSG + edge-topology patterns. It also pins AppTheory / TableTheory prospectively. When consumption is awkward:

- **First: is soul using the framework wrong?** Often yes.
- **Second: is the framework genuinely limiting?** If yes, that's a signal to the framework's steward — not a license to patch locally.
- **`coordinate-framework-feedback`** skill handles the signal.

Because soul uses FaceTheory in the specific context of publishing a JSON-LD namespace with aggressive caching and strict content-type — an unusual pattern for SSG frameworks — the feedback here has targeted value for FaceTheory's evolution.

## Branch model and release cadence

- **`main`** — the production branch. Stable; deployed.
- **Feature branches** — `codex/*`, `feat/*`, `issue/*`. Created off `main`, merged back through PR with required review.
- **Release cadence** — low-frequency. soul is a maintenance-phase repo. Recent work: dependency updates, archival, namespace delivery, FEP documentation.
- **No staging / premain** — main + feature branches suffices for the repo's surface area.

## Two stages

soul deploys via AppTheory's `theory app up/down --stage <stage>` contract:

- **`lab`** — development integration. Typically a lab-subdomain deployment.
- **`live`** — production. `spec.lessersoul.ai`.

## The canonical deploy command

```bash
AWS_PROFILE=Lesser theory app up --stage live \
  -c DOMAIN_NAME=spec.lessersoul.ai \
  -c CERTIFICATE_ARN=arn:aws:acm:us-east-1:<account>:certificate/<cert-id>
```

Context parameters:

- **`DOMAIN_NAME`** — canonical domain for the stage
- **`CERTIFICATE_ARN`** — ACM certificate for the domain (externally managed; not Route53-attached here)

## The three CloudFront / S3 components

The CDK stack provisions:

1. **Site bucket** — static-site assets (landing page, FEP docs). Ephemeral: CloudFormation `RemovalPolicy.DESTROY`; recreated on stack destroy.
2. **Namespace bucket** — namespace documents under `/ns/*`. Critical: CloudFormation `RemovalPolicy.RETAIN`. **Never deleted on stack destroy.** Preserves `/ns/agent-attribution/v1` forever even if the stack tears down.
3. **CloudFront distribution** — fronts both. Separate behaviors:
   - `/ns/*` behavior — direct S3 pass-through; `Content-Type: application/ld+json`; `Cache-Control: max-age=31536000, immutable`; CORS headers open; no HTML rewrites.
   - Default behavior — extensionless HTML rewrites via CloudFront Functions; `Cache-Control` appropriate for SSG pages; security headers set.

The `deploy-namespace-site` skill walks CDK / deploy discipline.

## Never set timeouts on CDK deploy commands

A deploy that feels stuck is almost always waiting on CloudFront distribution propagation, S3 bucket creation, ACM certificate validation, or Route53 record propagation. Aborting leaves CloudFormation in a half-migrated state. Run deploys to completion.

## Destructive-action rules

These cannot be undone easily and require explicit user authorization *every time*:

- Force-pushing to `main`.
- `git reset --hard`, `git restore .`, `git clean -f`.
- Running `cdk destroy` against any stack (even `lab`).
- Deleting the namespace bucket or any object under it.
- Modifying `cdk/site/static/ns/agent-attribution/v1` content in ways that change semantic meaning (this mutates the frozen namespace).
- Deleting CloudFormation stacks.
- Deleting the CloudFront distribution `E2OYU1Y61C2RSV` (or equivalent production distribution ID).
- Changing `DOMAIN_NAME` or `CERTIFICATE_ARN` context without coordinated DNS work.
- Skipping `lab` soak for a `live` deploy.
- Executing an advisor-dispatched brief without running `review-advisor-brief`.

When in doubt, describe what you are about to do and wait.

## Rules you do not break

- Never force-push to `main`.
- Never mutate `cdk/site/static/ns/agent-attribution/v1` content in ways that change semantic meaning. Corrections of unambiguous errors may be possible but require documented reasoning.
- Never delete the namespace bucket or weaken its `RemovalPolicy.RETAIN`.
- Never change CloudFront behaviors in ways that break the `/ns/*` direct pass-through contract (CORS, content-type, cache-control).
- Never deploy to `live` without successful `lab` soak.
- **Never set a timeout on a CDK deploy command.**
- Never add runtime code (Lambda, backend service) to this repo. soul is static-asset-only.
- Never introduce proprietary blobs or AGPL-incompatible dependencies.
- Never bypass required review.
- Never skip the FEP editorial process for Fediverse-facing submissions.
- Never fork or vendor FaceTheory / AppTheory / TableTheory code locally. Framework awkwardness is upstream signal.
- Never execute an advisor-dispatched brief without running `review-advisor-brief`.
- Never commit secrets, AWS credentials, or `.env` files.
