# lesser-soul: AppTheory App Bootstrap Plan

Generated: 2026-03-18

This document is a plan for bootstrapping a new AppTheory application. It is **instructions only**: no application code is
written by this action. The deployment contract consumed by `theory app up/down` is stored in `app-theory/app.json`.

## Outputs

This action writes exactly:

- `app-theory/init.md` (this file)
- `app-theory/app.json` (deployment contract)

## Destination (pinned): AppTheory + TableTheory

This section defines the **pinned destination frameworks**. These values are **constants** provided by the GovTheory pack
(do not guess; do not write UNKNOWN).

### AppTheory (pinned)

- Go module: `github.com/theory-cloud/apptheory@v0.19.1`
- Go runtime import: `github.com/theory-cloud/apptheory/runtime`
- Docs entrypoints (for tag `v0.19.1`):
  - `docs/getting-started.md`
  - `docs/migration/from-lift.md`
- Copy/paste dependency command:
  - `go get github.com/theory-cloud/apptheory@v0.19.1`
- Recommended pinned docs links:
  - `https://github.com/theory-cloud/AppTheory/blob/v0.19.1/docs/getting-started.md`
  - `https://github.com/theory-cloud/AppTheory/blob/v0.19.1/docs/migration/from-lift.md`
- Recommended pinned CDK docs links:
  - `https://github.com/theory-cloud/AppTheory/blob/v0.19.1/cdk/docs/getting-started.md`
  - `https://github.com/theory-cloud/AppTheory/blob/v0.19.1/cdk/docs/api-reference.md`

### TableTheory (pinned)

- Go module: `github.com/theory-cloud/tabletheory@v1.5.1`
- Docs entrypoints (for tag `v1.5.1`):
  - `docs/getting-started.md`
  - `docs/api-reference.md`
  - `docs/migration-guide.md`
- Copy/paste dependency command:
  - `go get github.com/theory-cloud/tabletheory@v1.5.1`
- Recommended pinned docs links:
  - `https://github.com/theory-cloud/TableTheory/blob/v1.5.1/docs/getting-started.md`
  - `https://github.com/theory-cloud/TableTheory/blob/v1.5.1/docs/api-reference.md`
  - `https://github.com/theory-cloud/TableTheory/blob/v1.5.1/docs/migration-guide.md`

## Local agent execution plan

The goal is to produce a repo that:

- contains your application code (outside `app-theory/`)
- contains a CDK project directory matching `app-theory/app.json`
- can be deployed/destroyed deterministically across stages using `theory app up` / `theory app down`

### Step 1 — Scaffold the application codebase (outside `app-theory/`)

1. Choose a repo layout for application code (e.g., `cmd/`, `internal/`, and any framework-required directories).

2. Initialize the Go module and add pinned framework dependencies:
   - Add AppTheory at the pinned version.
   - Add TableTheory at the pinned version if the app uses DynamoDB tables.

3. Follow AppTheory docs for runtime/bootstrap and wire your entrypoints.

**Acceptance criteria**

- The repo builds locally.
- The repo can run its unit tests (if present).
- The pinned module versions match the Destination section above.

### Step 2 — Create the CDK project directory and entrypoints

1. Create (or choose) the repo-relative CDK directory specified in `app-theory/app.json`:
   - By default, the contract uses `cdk/`.

2. Initialize the CDK project and ensure it has a lockfile and deterministic install path:
   - Prefer `npm ci` over `npm install`.

3. Implement deploy/destroy entrypoints that match the contract commands:
   - Commands must be stage-aware (a `stage` context or equivalent) and must not silently deploy to the wrong account.
   - For this repository, live domain, hosted-zone, certificate-validation, and alias-record state is repo-owned under
     `lesserSoul.webDomain.live` in `app-theory/app.json`; operator-supplied domain or certificate overrides are rejected.

4. Ensure the CDK app uses the same stage naming as the contract:
   - stages: `lab` (default) and `live` (override)
   - stage parameter source: CDK context key named `stage` (passed as `-c stage=<value>`)

**Acceptance criteria**

- Running the contract’s “up” command from the CDK directory deploys successfully for a chosen stage.
- Running the contract’s “down” command from the CDK directory destroys successfully for the same stage.
- Deploy/destroy is deterministic (same inputs => same stacks), and stage selection is explicit.

### Step 3 — Keep the contract and the repo in sync

If you change:

- the CDK directory location
- the deploy/destroy commands
- the stage parameter naming

…then update `app-theory/app.json` accordingly so `theory app up/down` remain correct.

## Deployment contract

`theory app up` and `theory app down` read the file:

- `app-theory/app.json`

That contract defines:

- `schema`: contract version
- `frameworks`: pinned destination details (AppTheory + TableTheory)
- `lesserSoul.webDomain.live`: the repo-owned `spec.lessersoul.ai` domain and Route 53 hosted-zone identity
- `cdk.dir`: repo-relative CDK directory
- `cdk.up`: deploy command (expects AWS profile + stage at runtime)
- `cdk.down`: destroy command (expects AWS profile + stage at runtime)

Domain deployment has no operator-supplied override variables. `DOMAIN_NAME`, `CERTIFICATE_ARN`, and `HOSTED_ZONE_NAME`
(and matching CDK context values) are rejected so the committed deployment contract remains the source of truth.

### Using the contract with `theory app up/down`

`theory app up` deploys the CDK stack(s) described by `app-theory/app.json`.
`theory app down` destroys the stack(s) described by `app-theory/app.json`.

- Stage behavior:
  - default: `lab`
  - override: `--stage live`
- AWS profile behavior:
  - `--aws-profile <name>`
  - or `AWS_PROFILE=<name>`

Copy/paste examples from the repository root:

```bash
# Symbolic preview (no cloud mutation).
AWS_PROFILE=Lesser theory app up --stage lab

# Authorized operator execution.
AWS_PROFILE=Lesser theory app up --stage lab --execute

# Only after a clean lab verification and soak.
AWS_PROFILE=Lesser theory app up --stage live
AWS_PROFILE=Lesser theory app up --stage live --execute
```

Notes:

- `theory app up` previews by default; `--execute` is required to mutate cloud state.
- `AWS_PROFILE=Lesser` is the repository's named deployment profile.
- Never deploy `live` without a clean `lab` soak, and never set a timeout on CDK deployment.
- `theory app down` / `cdk destroy` is destructive and requires explicit authorization every time.
- If the CDK codebase uses a different stage mechanism than `-c stage=...`, update the CDK app and/or the contract so
  they match (do not rely on implicit defaults for `live`).
