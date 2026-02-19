# lesser-soul: AppTheory App Bootstrap Plan

Generated: 2026-02-19

This document is a plan for bootstrapping a new AppTheory application. It is **instructions only**: no application code is
written by this action. The deployment contract consumed by `theory app up/down` is stored in `app-theory/app.json`.

Project context (repo-specific):
- Repository: `equaltoai/lesser-soul`
- Scope summary: LLM agent management and coordination for Lesser and Lesser host

## Outputs

This action writes exactly:
- `app-theory/init.md` (this file)
- `app-theory/app.json` (deployment contract)

## Destination (pinned): AppTheory + TableTheory

This section defines the **pinned destination frameworks**. These values are **constants** provided by the GovTheory pack.

### AppTheory (pinned)
- Go module: `github.com/theory-cloud/apptheory@v0.9.1`
- Go runtime import: `github.com/theory-cloud/apptheory/runtime`
- Docs entrypoints (for tag `v0.9.1`):
  - `docs/getting-started.md`
  - `docs/migration/from-lift.md`
- Copy/paste dependency command:
  - `go get github.com/theory-cloud/apptheory@v0.9.1`
- Recommended pinned docs links:
  - `https://github.com/theory-cloud/AppTheory/blob/v0.9.1/docs/getting-started.md`
  - `https://github.com/theory-cloud/AppTheory/blob/v0.9.1/docs/migration/from-lift.md`
- Recommended pinned CDK docs links:
  - `https://github.com/theory-cloud/AppTheory/blob/v0.9.1/cdk/docs/getting-started.md`
  - `https://github.com/theory-cloud/AppTheory/blob/v0.9.1/cdk/docs/api-reference.md`

### TableTheory (pinned)
- Go module: `github.com/theory-cloud/tabletheory@v1.4.0`
- Docs entrypoints (for tag `v1.4.0`):
  - `docs/getting-started.md`
  - `docs/api-reference.md`
  - `docs/migration-guide.md`
- Copy/paste dependency command:
  - `go get github.com/theory-cloud/tabletheory@v1.4.0`
- Recommended pinned docs links:
  - `https://github.com/theory-cloud/TableTheory/blob/v1.4.0/docs/getting-started.md`
  - `https://github.com/theory-cloud/TableTheory/blob/v1.4.0/docs/api-reference.md`
  - `https://github.com/theory-cloud/TableTheory/blob/v1.4.0/docs/migration-guide.md`

## Local agent execution plan

The goal is to produce a repo that:
- contains your application code (outside `app-theory/`)
- contains a CDK project directory matching `app-theory/app.json`
- can be deployed/destroyed deterministically across stages using `theory app up` / `theory app down`

### Step 1 — Scaffold the application codebase (outside `app-theory/`)

1) Choose a repo layout for application code (typical Go layout examples):
   - `cmd/lesser-soul/` (Lambda/API entrypoints)
   - `internal/` (domain logic)
   - `pkg/` (optional shared packages)
2) Initialize (or confirm) the Go module for this repo.
3) Add pinned framework dependencies:
   - AppTheory at `v0.8.0`
   - TableTheory at `v1.3.0` if the app provisions/uses DynamoDB tables

Copy/paste commands (adjust as needed to match your module path):

```bash
go mod init <your-module-path>
go get github.com/theory-cloud/apptheory@v0.8.0
# If using DynamoDB tables via TableTheory:
go get github.com/theory-cloud/tabletheory@v1.3.0
```

4) Follow the AppTheory runtime/bootstrap docs to wire your app runtime entrypoints.

**Acceptance criteria**
- `go test ./...` runs successfully (or there are no tests yet, but `go test` completes).
- `go mod tidy` completes cleanly.
- `go.mod` pins AppTheory to `v0.8.0` (and TableTheory to `v1.3.0` if used).

### Step 2 — Create the CDK project directory and entrypoints

1) Create (or choose) the repo-relative CDK directory specified in `app-theory/app.json`:
   - This contract uses `infra/cdk/`.
2) Initialize the CDK project in that directory.
   - Keep installs deterministic by committing a lockfile.
   - Prefer `npm ci` in automation (it uses the lockfile strictly).
3) Implement deploy/destroy entrypoints that match the contract commands (see `app-theory/app.json`):
   - Deploy uses `npx cdk deploy --all -c stage=<stage>`
   - Destroy uses `npx cdk destroy --all -c stage=<stage>`
   - The CDK app must read the `stage` context and must not silently default to an unintended account/region.

**Acceptance criteria**
- From the `infra/cdk/` directory, `npm ci` succeeds.
- From the `infra/cdk/` directory, a stage-scoped deploy succeeds (for a chosen stage).
- From the `infra/cdk/` directory, the corresponding stage-scoped destroy succeeds.

### Step 3 — Keep infra commands deterministic and stage-aware

Local agent guidance:
- Ensure `stage` is always explicitly part of the CDK command (via `-c stage=...` as in the contract).
- Ensure your CDK code selects stack names / resource names using the stage value to avoid collisions.
- Ensure the deploy/destroy commands are safe to run repeatedly (idempotent where possible).

**Acceptance criteria**
- Re-running the same deploy command for the same stage results in no unintended changes (beyond expected drift).
- Deploying `lab` does not affect `live` resources.

## Deployment contract

`theory app up` and `theory app down` read the file:
- `app-theory/app.json`

That contract defines:
- `schema`: contract version
- `frameworks`: pinned destination details (AppTheory + TableTheory)
- `cdk.dir`: repo-relative CDK directory
- `cdk.up`: deploy command (expects AWS profile + stage at runtime)
- `cdk.down`: destroy command (expects AWS profile + stage at runtime)

### Using the contract with `theory app up/down`

`theory app up` deploys the CDK stack(s) described by `app-theory/app.json`.
`theory app down` destroys the stack(s) described by `app-theory/app.json`.

- Stage behavior:
  - default: `lab`
  - override: `--stage live`
- AWS profile behavior:
  - `--aws-profile <name>`
  - or `AWS_PROFILE=<name>`

Copy/paste examples:

```bash
# Default stage (lab) using an explicit AWS profile via env var:
AWS_PROFILE=my-profile theory app up

# Explicit stage + explicit profile via flags:
theory app up --aws-profile my-profile --stage lab

# Destroy live:
theory app down --aws-profile my-profile --stage live
```
