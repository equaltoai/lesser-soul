# lesser-soul CDK

This directory deploys the public `spec.lessersoul.ai` site and the stable JSON-LD namespace document used by the Agent
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
- `live`: defaults to `spec.lessersoul.ai`
- custom domains require either:
  - `-c certificateArn=...` for external DNS / non-Route-53 setups
  - `-c hostedZoneName=...` if CDK should manage Route 53 records and DNS validation

Optional CDK context:

- `stage`
- `domainName`
- `certificateArn`
- `hostedZoneName`

## Commands

```bash
npm ci
npm run build:site
npx cdk synth -c stage=lab
npx cdk deploy --all -c stage=live
npx cdk deploy --all -c stage=live -c certificateArn=arn:aws:acm:us-east-1:123456789012:certificate/abcd...
```

## License

This project follows the repository's GNU Affero General Public License v3.0. See [`../LICENSE`](../LICENSE) for details.
