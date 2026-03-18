# spec.lessersoul.ai inventory

Date: 2026-03-18
Scope: lesser-soul issue `#4` and issue `#5`

This document records the deployed ownership and delivery shape for the Agent Social Attribution namespace endpoint.
It supersedes earlier planning text that referred to `lessersoul.ai` or to `lesser-host` as the likely deployment home.

## Current ownership

- Repository owner: `equaltoai/lesser-soul`
- Infrastructure shape: CDK-managed static site + namespace delivery in this repo
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

- ACM certificate ARN:
  `arn:aws:acm:us-east-1:693925625407:certificate/3ba5692d-fe41-43d1-8809-282e5c07754b`
- Certificate coverage: `spec.lessersoul.ai`
- Validation model: external DNS, not Route 53-managed validation

### DNS authority

- Route 53 is not authoritative for the public host
- required public DNS record:
  - type: `CNAME`
  - host: `spec`
  - value: `d1quktmmrrqb1.cloudfront.net`

## Delivery guarantees

The namespace endpoint is deployed as a static JSON-LD object with:

- direct `200 OK`
- `Content-Type: application/ld+json`
- `Access-Control-Allow-Origin: *`
- no HTML shell
- no JavaScript redirect
- long-lived immutable caching on the versioned `/v1` object path

Because the namespace object is versioned and cached immutably, a CloudFront invalidation was issued on
`/ns/agent-attribution/v1` after the hostname migration to roll the cached document body forward.

## Verification performed

- `npm run build:site`
- `npm run typecheck`
- `npx cdk synth -c stage=live -c domainName=spec.lessersoul.ai -c certificateArn=...`
- `AWS_PROFILE=Lesser npx cdk deploy --all -c stage=live -c domainName=spec.lessersoul.ai -c certificateArn=...`
- AWS CLI verification of:
  - CloudFront alias
  - CloudFront ACM certificate ARN
  - CloudFormation stack outputs
  - namespace bucket object content
- HTTP verification through the CloudFront host path:
  - `200 OK`
  - `content-type: application/ld+json`
  - body identifies `https://spec.lessersoul.ai/ns/agent-attribution/v1#`

## Tracker implications

- Issue `#4` is satisfied by this inventory and can be closed once linked from the issue thread
- Issue `#5` is implemented in merged PR `#8`; the remaining tracker cleanup is to reflect the final hostname
  `spec.lessersoul.ai` instead of the earlier `lessersoul.ai` placeholder
