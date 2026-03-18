# lesser-soul CDK

This directory deploys the public `lessersoul.ai` site and the stable JSON-LD namespace document used by the Agent
Social Attribution FEP work.

## What gets deployed

- a static FaceTheory-generated site from `site/`
- a dedicated namespace bucket for `/ns/*`
- a CloudFront distribution that:
  - rewrites extensionless site routes to `index.html`
  - does **not** rewrite `/ns/*`
  - serves `/ns/agent-attribution/v1` as a direct JSON-LD document

## Stage behavior

- `lab`: deploys without a custom domain unless you pass `-c domainName=...`
- `live`: defaults to `lessersoul.ai`

Optional CDK context:

- `stage`
- `domainName`
- `hostedZoneName`

## Commands

```bash
npm ci
npm run build:site
npx cdk synth -c stage=lab
npx cdk deploy --all -c stage=live
```
