---
name: deploy-namespace-site
description: Use to walk a merged change through CDK deploy — `theory app up --stage lab` → `theory app up --stage live` — preserving the namespace bucket's `RemovalPolicy.RETAIN`, the `/ns/*` CloudFront behavior's direct pass-through contract, and never timing out CDK commands.
---

# Deploy the namespace site

After `implement-milestone` lands a PR to `main`, the change is ready to reach `spec.lessersoul.ai`. This skill walks the CDK deploy through `lab` → `live`, preserves the namespace bucket's retention contract, and verifies that the permanent public URL continues to resolve correctly after deploy.

## When this skill runs

Invoke when:

- A PR has merged to `main` and is ready for rollout
- A namespace-addition at a new version path needs deployment
- A site-HTML or FEP-docs content change needs deployment
- A CDK-topology change needs deployment
- A dependency bump or AppTheory-pinning alignment is ready to land
- A rollback is required

## Preconditions

- **The change is merged to `main`.**
- **MCP tools healthy**, `memory_recent` first.
- **The operator's AWS credentials** (`AWS_PROFILE=Lesser`) are configured.
- **The ACM certificate** for `spec.lessersoul.ai` (or the lab-stage domain) is valid and non-expiring.
- **`cdk synth --context stage=lab`** succeeds locally before deploy.
- **For rollback**: the target commit is identified.

## The canonical deploy command

```bash
AWS_PROFILE=Lesser theory app up --stage <lab|live> \
  -c DOMAIN_NAME=<domain> \
  -c CERTIFICATE_ARN=<arn>
```

For `live`:

```bash
AWS_PROFILE=Lesser theory app up --stage live \
  -c DOMAIN_NAME=spec.lessersoul.ai \
  -c CERTIFICATE_ARN=arn:aws:acm:us-east-1:693925625407:certificate/3ba5692d-fe41-43d1-8880-282e5c07754b
```

(ARNs and domains come from `docs/spec-lessersoul-ai-inventory.md`; verify that document before deploy if the values look unfamiliar.)

## The stack architecture (reminder)

`cdk/lib/lesser-soul-site-stack.ts` deploys:

1. **Site bucket** — `RemovalPolicy.DESTROY`; ephemeral. Holds landing page + FEP docs HTML.
2. **Namespace bucket** — **`RemovalPolicy.RETAIN`**; critical. Holds `/ns/agent-attribution/v<N>` objects. Survives stack teardown.
3. **CloudFront distribution** — single distribution with two behaviors:
   - `/ns/*` path: direct S3 pass-through from namespace bucket; CORS open; `Content-Type: application/ld+json`; `Cache-Control: max-age=31536000, immutable`; no HTML rewrite; no JavaScript.
   - Default path: SSG HTML from site bucket; extensionless-URL rewrites via CloudFront Functions; security headers; cache-control per SSG output.
4. **Route53 / DNS** — externally managed; `spec.lessersoul.ai` points at the CloudFront distribution.

**Production CloudFront distribution ID**: `E2OYU1Y61C2RSV` (per `docs/spec-lessersoul-ai-inventory.md`). Never delete this distribution.

## Never set timeouts on CDK deploy commands

A deploy that feels stuck is almost always waiting on CloudFront distribution propagation (which can take many minutes), S3 bucket creation, ACM certificate attachment, or Route53 record propagation. Aborting leaves CloudFormation in a half-migrated state; propagation in-flight continues even after an abort, potentially creating a partial configuration that takes longer to unblock than just waiting.

Run deploys to completion. Capture full output. If genuinely stuck, check CloudFormation console state through the user — don't abort.

## Lab deploy and verification

```bash
AWS_PROFILE=Lesser theory app up --stage lab \
  -c DOMAIN_NAME=<lab-domain> \
  -c CERTIFICATE_ARN=<lab-arn>
```

After the CDK deploy completes:

1. **Verify stack status.** CloudFormation stack reaches `UPDATE_COMPLETE` or `CREATE_COMPLETE`.
2. **Verify S3 objects.** Expected namespace version(s) exist under the namespace bucket. Expected SSG HTML exists in the site bucket.
3. **Verify CloudFront behaviors.** The distribution has the two expected behaviors (`/ns/*` and default) with the correct config.
4. **Curl test the namespace URL**:
   ```bash
   curl -sv https://<lab-domain>/ns/agent-attribution/v1 | head -30
   ```
   Expect: HTTP 200, `Content-Type: application/ld+json`, `Cache-Control: max-age=31536000, immutable`, `Access-Control-Allow-Origin: *`, body is valid JSON-LD matching the deployed content.
5. **Curl test any newly-added namespace version** (e.g. `/v2`) if this deploy adds one.
6. **Browse the site HTML**: landing page, FEP docs pages render correctly.
7. **Browse a non-existent path**: CloudFront Function rewrite handles extensionless URLs; 404 pages render as designed.
8. **Check response headers** on site HTML: CSP strict, security headers present.
9. **Cache-behavior spot check**: the `/ns/*` paths should have aggressive caching; the default paths should not have `max-age=31536000, immutable`.

Do not promote to `live` until lab verification is clean.

## Live deploy and verification

```bash
AWS_PROFILE=Lesser theory app up --stage live \
  -c DOMAIN_NAME=spec.lessersoul.ai \
  -c CERTIFICATE_ARN=arn:aws:acm:us-east-1:693925625407:certificate/3ba5692d-fe41-43d1-8880-282e5c07754b
```

**Live is production.** `spec.lessersoul.ai` is resolved by every ActivityPub implementation that consumes agent-attribution.

- **Operator authorizes live deploy explicitly.**
- **Post-deploy verification same as lab** but against `https://spec.lessersoul.ai/`.
- **Additional live checks**:
  - `curl -sv https://spec.lessersoul.ai/ns/agent-attribution/v1 | head -30` — expect unchanged response for existing content
  - If a new version was deployed: `curl -sv https://spec.lessersoul.ai/ns/agent-attribution/v<N> | head -30`
  - CloudFront cache-hit rate observable (through the user); no unusual spikes in 4xx/5xx
  - ACM certificate valid with months of runway before expiry
- **Update `docs/spec-lessersoul-ai-inventory.md`** if the infrastructure state changed (new bucket, new behavior, new distribution ID — rare).

## If lab or live surfaces a regression

- **Stop.** Do not promote (if in lab) or continue (if in live).
- **Diagnose quickly** — which behavior (`/ns/*` or default) is affected? A specific URL?
- **For namespace regressions** — extremely high severity. The permanent URL must resolve correctly.
- **Rollback options**:
  - **Revert the commit on `main`** and redeploy via `theory app up`. CloudFormation applies the prior state.
  - **Per-stage rollback** — roll back `live` while keeping `lab` on the new commit.
  - **Emergency recovery** — if `RemovalPolicy.RETAIN` on the namespace bucket protected against an accidental deletion, the preserved objects remain available.
- **Never delete the CloudFormation stack.**
- **Never weaken `RemovalPolicy.RETAIN` on the namespace bucket** as part of a rollback.
- **Coordinate through the user** if ActivityPub peers are affected by a live regression.
- **Record the regression.** Namespace-URL regressions are high-signal memory material.

## Output: the deploy record

```markdown
## Deploy record: <change name>

### Stage
<lab / live>

### Command
`AWS_PROFILE=Lesser theory app up --stage <stage> -c DOMAIN_NAME=<domain> -c CERTIFICATE_ARN=<arn>`

### Timestamp
<...>

### CloudFormation stack status
<UPDATE_COMPLETE / CREATE_COMPLETE>

### S3 objects verified
- Namespace bucket: <paths verified — /ns/agent-attribution/v1, /v2 (if new), etc.>
- Site bucket: <paths verified — index, FEP docs, etc.>

### CloudFront behaviors verified
- `/ns/*`: direct pass-through, CORS open, content-type application/ld+json, cache immutable — confirmed
- Default: extensionless rewrite, SSG HTML, security headers — confirmed

### Namespace URL verification
- `curl -sv https://<domain>/ns/agent-attribution/v1` → HTTP 200 + expected headers + body
- (If new version) `curl -sv https://<domain>/ns/agent-attribution/v<N>` → HTTP 200 + expected

### Site HTML verification
- Landing page renders: <confirmed>
- FEP docs pages render: <confirmed>
- Extensionless URL rewrites work: <confirmed>

### Infrastructure inventory update (if applicable)
- `docs/spec-lessersoul-ai-inventory.md` updated: <yes / no — no change needed>

### Issues observed
<none / described>

### Rollback (if any)
<trigger, mechanism, prior commit>
```

## Refusal cases

- **"Set a 10-minute timeout on the CDK deploy."** Never.
- **"Skip lab verification; the change is just a site-HTML update."** Refuse. Verification is the gate.
- **"Run `cdk destroy` against `lab` to rebuild cleanly."** Refuse unless explicitly authorized; the namespace bucket's `RemovalPolicy.RETAIN` saves the content but the rebuild is a disruption worth thinking about.
- **"Delete the CloudFront distribution `E2OYU1Y61C2RSV` and create a new one."** Refuse. Would take the live namespace URL offline during propagation.
- **"Weaken `RemovalPolicy.RETAIN` to simplify teardowns."** Refuse.
- **"Modify CloudFront behaviors manually outside CDK to patch a cache-control issue."** Refuse. CDK is the source of truth.
- **"Skip updating `docs/spec-lessersoul-ai-inventory.md` after an infrastructure change."** Refuse. The inventory document is the operator's reference for deployed state.
- **"Deploy to live without `DOMAIN_NAME` / `CERTIFICATE_ARN` context set explicitly."** Refuse. Context parameters are load-bearing.
- **"Delete an old namespace version (e.g. `/v1`) to simplify the bucket."** Refuse. Old versions serve forever.

## Persist

Append only when the deploy surfaces something worth remembering — a CDK quirk (CloudFront propagation timing, ACM attachment), an S3 object layout observation, a cache-behavior subtlety, a rollback scenario, a cross-repo coordination moment (host observed the namespace update). Routine clean deploys aren't memory material. Five meaningful entries beat fifty log-shaped ones.

## Handoff

- **All stages clean** — stop. Record deploy, append memory if warranted.
- **Regression surfaced** — rollback per the above, then route through `investigate-issue`.
- **ActivityPub peer reports a namespace-resolution issue post-deploy** — route through `investigate-issue` with peer context, then `evolve-namespace` if the issue implicates namespace content.
- **Deploy reveals a scoping question** — `scope-need` once the current deploy is stable.
- **Deploy surfaces framework awkwardness** (FaceTheory SSG pattern friction, CDK construct gap) — `coordinate-framework-feedback`.