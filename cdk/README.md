# lesser-soul CDK

This directory deploys the public `spec.lessersoul.ai` site and the stable JSON-LD namespace document used by the Agent
Social Attribution FEP work.

## What gets deployed

- a static FaceTheory-generated site from `site/`
- a dedicated, retained namespace bucket for `/ns/*`
- a CloudFront distribution that:
  - rewrites extensionless site routes to `index.html`
  - does **not** rewrite `/ns/*`
  - serves `/ns/agent-attribution/v1` as a direct JSON-LD document
- for `live`, a CDK-managed ACM certificate plus Route 53 apex `A` and `AAAA` aliases

## Stage and domain contract

- `lab`: deploys without a custom domain
- `live`: deploys `spec.lessersoul.ai`
- live custom-domain state is defined in `../app-theory/app.json` at
  `lesserSoul.webDomain.live.{domainName,hostedZoneId,hostedZoneName}`
- the CloudFront viewer certificate is created by CDK with DNS validation in the configured Route 53 hosted zone
- legacy `DOMAIN_NAME`, `CERTIFICATE_ARN`, and `HOSTED_ZONE_NAME` environment variables and matching CDK context are rejected

The only supported CDK context is `stage`, and its only values are `lab` and `live`.

## Canonical operator flow

Run these commands from the repository root. `theory app up` is a symbolic preview unless `--execute` is present.

```bash
# Preview and then execute lab.
AWS_PROFILE=Lesser theory app up --stage lab
AWS_PROFILE=Lesser theory app up --stage lab --execute

# Only after lab verification and soak: preview and then execute live.
AWS_PROFILE=Lesser theory app up --stage live
AWS_PROFILE=Lesser theory app up --stage live --execute
```

Never deploy `live` before a clean `lab` soak. Never add a timeout to a deploy command; CloudFront and certificate
propagation can take many minutes. A reviewer merges to `main`, and an authorized operator runs deployment from updated
`main`.

## Local build and synth

```bash
cd cdk
npm ci
npm run build:site
npm run typecheck
npx cdk synth -c stage=lab
npx cdk synth -c stage=live
```

These commands synthesize locally and do not deploy.

## License

This project follows the repository's GNU Affero General Public License v3.0. See [`../LICENSE`](../LICENSE) for details.
