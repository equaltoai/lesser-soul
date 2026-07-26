# spec.lessersoul.ai inventory

Original inventory date: 2026-03-18
Last verified: 2026-07-26
Original scope: lesser-soul issues `#4` and `#5`

This document records the deployed ownership and delivery shape for the Agent Social Attribution namespace endpoint.
It supersedes earlier planning text that referred to `lessersoul.ai` or to `lesser-host` as the likely deployment home.

## Current ownership

- Repository owner: `equaltoai/lesser-soul`
- AWS account: `693925625407`
- Operator AWS profile: `Lesser`
- Infrastructure shape: CDK-managed static site + namespace delivery in this repo
- CloudFormation stack: `LesserSoulSite-live` (`UPDATE_COMPLETE` when last verified)
- Live public host target: `spec.lessersoul.ai`
- Namespace URL: `https://spec.lessersoul.ai/ns/agent-attribution/v1`

## Edge inventory

### CloudFront distribution

- Distribution ID: `E2OYU1Y61C2RSV`
- Distribution domain: `d1quktmmrrqb1.cloudfront.net`
- Alias: `spec.lessersoul.ai`

### Origins

- Site origin bucket: `lessersoulsite-live-sitebucket397a1860-kz6uqr6bvip5`
- Namespace origin bucket: `lessersoulsite-live-namespacebucket7d6583f5-sokdtnyhk4na`
- `/ns/*` is routed to the dedicated namespace bucket without HTML rewrites
- non-namespace site routes use the static site bucket with extensionless HTML rewrites

### Certificate

- Active CloudFront ACM certificate ARN when last verified:
  `arn:aws:acm:us-east-1:693925625407:certificate/e90ca462-9351-4b01-ad04-538031c2e423`
- Certificate coverage: `spec.lessersoul.ai`
- Certificate model: Amazon-issued, CDK-managed, DNS-validated, and renewal-eligible
- CloudFormation logical resource: `SiteCertificate38C247F6`
- CDK source of truth: `app-theory/app.json` defines
  `lesserSoul.webDomain.live.{domainName,hostedZoneId,hostedZoneName}`
- Operator deploys must not provide `CERTIFICATE_ARN`; the certificate is part of the CDK topology

The physical certificate ARN is inventory, not operator input. It can change if CloudFormation replaces the certificate;
the committed web-domain contract and stack resource are authoritative.

### DNS authority

- Route 53 hosted zone name: `spec.lessersoul.ai`
- Route 53 hosted zone ID: `Z08181472DNYJKVQ0IEFV`
- CDK-managed apex records:
  - `A spec.lessersoul.ai` → `d1quktmmrrqb1.cloudfront.net`
  - `AAAA spec.lessersoul.ai` → `d1quktmmrrqb1.cloudfront.net`
- CloudFormation logical resources: `AliasRecordAB84985CE`, `AliasRecordAaaaA6E84686`

## Delivery guarantees

The namespace endpoint is deployed as a static JSON-LD object with:

- direct `200 OK`
- `Content-Type: application/ld+json`
- `Access-Control-Allow-Origin: *`
- no HTML shell
- no JavaScript redirect
- long-lived immutable caching on the versioned `/v1` object path
- namespace-bucket `RemovalPolicy.RETAIN`

Because the namespace object is versioned and cached immutably, a CloudFront invalidation was issued on
`/ns/agent-attribution/v1` after the hostname migration to roll the cached document body forward.

## Canonical deployment entrypoint

From an updated, reviewed `main` checkout:

```bash
# Preview and execute lab first.
AWS_PROFILE=Lesser theory app up --stage lab
AWS_PROFILE=Lesser theory app up --stage lab --execute

# Only after successful lab verification and soak.
AWS_PROFILE=Lesser theory app up --stage live
AWS_PROFILE=Lesser theory app up --stage live --execute
```

`theory app up` is preview-only unless `--execute` is present. Do not pass domain/certificate overrides, do not skip the
lab soak, and do not set a timeout on deployment.

## Verification performed

Repository validation:

- `npm run build:site`
- `npm run typecheck`
- `npm run check:contracts`
- `npx cdk synth -c stage=lab`
- `npx cdk synth -c stage=live`

Read-only AWS verification with `AWS_PROFILE=Lesser`:

- STS account `693925625407`
- CloudFormation stack status and outputs
- CloudFront alias, distribution status, and active viewer certificate
- ACM certificate status and CloudFront attachment
- Route 53 hosted zone and apex `A` / `AAAA` aliases
- namespace and site bucket identities

HTTP verification through the live host checks:

- `200 OK`
- `content-type: application/ld+json`
- permissive CORS
- body identifies `https://spec.lessersoul.ai/ns/agent-attribution/v1#`

## Tracker history

- Issues `#4` and `#5` are closed.
- PR `#8` established the original deployed site/namespace shape and final `spec.lessersoul.ai` hostname.
- The 2026-07-16 stack update moved certificate validation and apex DNS aliases into the CDK-managed live topology.
