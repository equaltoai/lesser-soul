# Soul Pack Cleanup Plan

**Date:** 2026-03-04
**Status:** Proposed
**Context:** The soul pack was designed as a signed deployment tarball containing a separate CDK stack and bootstrap CLI (`cmd/soul-cli`) for deploying lesser-soul infrastructure into managed instances. This artifact was never built — no tarball, no manifest, no CLI exist. All soul implementation lives in `lesser-host` and deploys through its existing pipeline. The pack concept is dead code.

**What stays:** The S3 bucket (`lesser-host-<stage>-<account>-<region>-soul-packs`) is still used for registry artifacts (registration files, reputation snapshots, validation snapshots) under `registry/v1/`. The KMS attestation signing key is still used by the reputation worker for snapshot signing. These are unrelated to the pack deployment concept and must be preserved.

**What goes:** Everything related to fetching, verifying, and deploying a signed tarball.

---

## 1. CDK Stack — Remove Pack-Specific Infrastructure

**File:** `lesser-host/cdk/lib/lesser-host-stack.ts`

### 1a. Remove SoulPackSigningKey (lines 167-173)

The `SoulPackSigningKey` KMS key and its alias (`alias/${namePrefix}-soul-pack-signing`) exist solely to verify pack tarball signatures. This is distinct from the `AttestationSigningKey` (line 150) which is used by the reputation worker and must stay.

Remove:
- `SoulPackSigningKey` KMS key construct
- `SoulPackSigningKeyArnParam` SSM parameter (`/soul/${stage}/signingKeyArn`)

### 1b. Remove SoulPackVersionParam (lines 187-192)

The `/soul/${stage}/packVersion` SSM parameter tracks which tarball version to deploy. No longer needed.

Remove:
- `SoulPackVersionParam` SSM parameter construct

### 1c. Rename SoulPackBucket (lines 158-165) — Optional

The bucket itself stores registry artifacts and is actively used. The name `soul-packs` is misleading now. Consider renaming to `soul-registry` in a future migration, but this is cosmetic and can be deferred. The `SoulPackBucketNameParam` (`/soul/${stage}/packBucketName`) SSM parameter is consumed by the reputation worker and controlplane for S3 reads/writes to `registry/v1/` — it stays, but consider renaming the parameter path to `/soul/${stage}/registryBucketName` alongside any bucket rename.

### 1d. Remove `fetch_and_verify_soul_pack()` function (lines 441-476)

The entire shell function in the CodeBuild runner script that downloads, signature-verifies, checksum-verifies, and unpacks the tarball. Remove the full function definition.

### 1e. Remove `soul-deploy` RUN_MODE branch (lines 540-547)

```
elif [ "$RUN_MODE" = "soul-deploy" ]; then
  echo "Deploying lesser-soul pack..."
  fetch_and_verify_soul_pack
  echo "Deploying lesser-soul version: $SOUL_PACK_VERSION"
  cd "$SOUL_PACK_DIR/infra/cdk"
  npm ci
  AWS_PROFILE=managed npx cdk deploy --all ...
  cd - >/dev/null
```

This deploys a CDK stack from inside the pack tarball. The tarball and its `infra/cdk/` directory don't exist.

### 1f. Remove `soul-init` RUN_MODE branch (lines 548-701)

```
elif [ "$RUN_MODE" = "soul-init" ]; then
  echo "Bootstrapping lesser-soul from signed pack..."
  fetch_and_verify_soul_pack
  ...
  (cd "$SOUL_PACK_DIR" && GOTOOLCHAIN=auto go run ./cmd/soul-cli bootstrap)
  ...
```

This is ~150 lines that:
1. Fetches and verifies the pack
2. Extracts wallet credentials from `bootstrap.json`
3. Runs wallet challenge/response authentication
4. Executes `cmd/soul-cli bootstrap` from the pack
5. Writes a soul receipt with queue URLs and agent token paths

The `cmd/soul-cli` binary does not exist. The bootstrap logic (agent registration, token provisioning) needs to be evaluated — if it's needed, it should be implemented directly in lesser-host rather than through a phantom CLI.

---

## 2. Provision Worker — Remove Soul Deploy/Init State Machine Steps

### 2a. Remove step constants

**File:** `lesser-host/internal/provisionworker/server.go` (lines 301-304)

Remove:
```go
provisionStepSoulDeployStart = "soul.deploy.start"
provisionStepSoulDeployWait  = "soul.deploy.wait"
provisionStepSoulInitStart   = "soul.init.start"
provisionStepSoulInitWait    = "soul.init.wait"
```

Keep:
```go
provisionStepSoulReceiptIngest = "soul.receipt.ingest"
provisionStepSoulRoutingStart  = "soul.routing.start"
// ... other soul routing steps
```

The receipt ingest and routing steps are for wiring `/soul/*` CloudFront paths and are part of the active deployment pipeline.

### 2b. Remove advance functions

**File:** `lesser-host/internal/provisionworker/server.go` (lines ~1625-1790)

Remove these four methods on `*Server`:
- `advanceProvisionSoulDeployStart` — starts CodeBuild with `RUN_MODE=soul-deploy`
- `advanceProvisionSoulDeployWait` — polls CodeBuild status
- `advanceProvisionSoulInitStart` — starts CodeBuild with `RUN_MODE=soul-init`
- `advanceProvisionSoulInitWait` — polls CodeBuild status

### 2c. Remove note constant

**File:** `lesser-host/internal/provisionworker/constants.go` (line 9)

Remove:
```go
noteStartingSoulDeployRunner = "starting soul deploy runner"
```

### 2d. Update state machine flow

**File:** `lesser-host/internal/provisionworker/advance_body_mcp.go` (line 30)

The MCP deploy completion currently checks if soul is enabled and not yet provisioned, then transitions to `soul.deploy.start`. This transition needs to be updated to skip directly to `soul.receipt.ingest` or `soul.routing.start` (or wherever the flow should go after MCP wiring when soul steps are removed).

### 2e. Evaluate `SoulVersion` field on instance model

**File:** `lesser-host/internal/store/models/instance.go`
**File:** `lesser-host/internal/store/models/provision_job.go`

The `SoulVersion` field was used to pin a specific pack version. With no pack, this field is vestigial. However, removing a persisted model field requires care — existing DynamoDB records may have this field set. Options:
- Leave the field but stop writing it (safe, no migration needed)
- Remove and ignore on read (requires confirming no code path depends on it)

### 2f. Evaluate receipt struct

**File:** `lesser-host/internal/provisionworker/receipts.go` (line 32)

`SoulVersion` field on the soul receipt struct. The receipt is written by the `soul-init` runner and ingested by `soul.receipt.ingest`. If the init runner is removed, the receipt format changes. The receipt ingest step needs to be updated to work without pack-related fields.

### 2g. Remove `SOUL_VERSION` env var injection

**File:** `lesser-host/internal/provisionworker/server_helpers.go` (line 84)

```go
cbtypes.EnvironmentVariable{Name: aws.String("SOUL_VERSION"), Value: aws.String(strings.TrimSpace(inst.SoulVersion))},
```

This passes the pack version to CodeBuild. Remove it.

---

## 3. Provision Worker Tests

### 3a. Update state machine flow test

**File:** `lesser-host/internal/provisionworker/state_machine_flow_internal_test.go`

- `TestProvisionStateMachine_SoulEnabled_SuccessPathAcrossSteps` (line 458) — remove or rewrite to skip `soul.deploy.*` and `soul.init.*` steps
- Remove `soul_version` from test receipt JSON (line 523)

### 3b. Update branch/optional step tests

**File:** `lesser-host/internal/provisionworker/state_machine_optional_steps_branches_internal_test.go`

Remove test cases for:
- `soul_deploy_start` (line 550)
- `soul_init_start` (line 564)
- `soul_deploy_wait` timeout/failure (lines 609, 633)
- `soul_init_wait` timeout/failure (lines 610, 634)
- Soul-enabled branch transition tests that route through deploy/init

---

## 4. Documentation

### 4a. Update soul-pack-bucket-layout.md

**File:** `lesser-host/docs/soul-pack-bucket-layout.md`

Remove the "Signed Soul packs (bucket root)" section (lines 12-18) describing the tarball, manifest, and signature files. Remove the "Stage pointers (SSM)" entries for `signingKeyArn` and `packVersion`. Keep the `registry/v1/` section intact — that's the active layout.

Consider renaming the file to `soul-registry-bucket-layout.md`.

### 4b. Update soul-surface.md

**File:** `lesser-host/docs/soul-surface.md`

Review and update references to `soul.deploy.*` and `soul.init.*` provisioning steps and `RUN_MODE` values.

### 4c. Update ROADMAP in lesser-soul

**File:** `lesser-soul/ROADMAP.md`

The roadmap's "Hard requirements" section (line 23) references "No new databases (existing state table + S3 soul pack bucket)". Update to clarify the bucket is a registry artifact store, not a pack store.

### 4d. Update SPEC.md references

**File:** `lesser-soul/SPEC.md`

Section 12.4 and 12.5 reference the soul pack bucket and SSM pointers. Update to remove pack-specific references while keeping registry artifact layout.

---

## 5. Remove Bootstrap Artifacts (No Replacement Needed)

**Decision (2026-03-04):** The bootstrap flow is dead. The built-in agent concept (researcher, assistant, curator, bridge, custom-coder, custom-summarizer) with SQS queue orchestration and `POST /soul/tasks` was a misguided pattern superseded by MCP. No replacement is needed — the entire flow is purely subtractive.

### 5a. Remove integration test runbook

**File:** `lesser-host/docs/soul-managed-integration-test.md`

Delete entirely. References `publish-pack.sh`, pack verification, `soul.deploy.*`/`soul.init.*` steps, built-in agent registration, `POST /soul/tasks` orchestrator, and SQS queue infrastructure — all dead.

### 5b. Remove soul receipt struct fields

**File:** `lesser-host/internal/provisionworker/receipts.go`

Remove the `soulReceipt` struct (or its dead fields: `SoulVersion`, queue URLs, agent token paths). If the struct is only consumed by `soul.receipt.ingest` for writing `SoulProvisionedAt` on the instance, simplify to only what the receipt ingest step actually needs.

### 5c. Clean up `SoulVersion` on instance and provision job models

**Files:**
- `lesser-host/internal/store/models/instance.go` — stop writing `SoulVersion`
- `lesser-host/internal/store/models/provision_job.go` — remove `SoulVersion` if present

Leave the field defined to avoid breaking reads of existing DynamoDB records, but stop setting it in any write path.

---

## Execution Order

1. **Documentation first** (steps 4a-4d) — low risk, clarifies intent
2. **CDK runner script** (steps 1d-1f) — remove dead shell code
3. **Provision worker** (steps 2a-2g, 3a-3b) — remove state machine steps + tests
4. **CDK infrastructure** (steps 1a-1b) — remove KMS key and SSM param constructs
5. **Bootstrap evaluation** (step 5) — decide and implement replacement if needed
6. **Bucket rename** (step 1c) — optional, deferred

Steps 1a-1b (CDK construct removal) must be done carefully — if the KMS key or SSM parameters are already deployed, removing them from CDK will trigger deletion on next deploy. Verify no other consumers reference these before removing.
