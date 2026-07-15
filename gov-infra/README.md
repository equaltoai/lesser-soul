# lesser-soul governance infrastructure (`gov-infra/`)

Repo-local governance materialization for the **`software_repo_gov_infra`** profile.
It makes lesser-soul's quality / consistency / completeness / security / compliance /
maintainability / docs posture explicit, versioned, deterministic, and fail-closed.

This surface is **CI-core and never retired** for this profile. MCP changes how
governance guidance is *managed*, not *whether* repo-local `gov-infra/` exists.

## Quick start

From the repository root:

1. Run the deterministic rubric verifier:
   - `bash gov-infra/verifiers/gov-verify-rubric.sh`
2. Read the machine report (schema `gov_rubric_report.v1`):
   - `gov-infra/evidence/gov-rubric-report.json`
3. Inspect per-check evidence:
   - `gov-infra/evidence/*-output.log`

Verifier scripts are safe to commit **without** execute permission; always run via `bash …`.

## What the verifier checks

All checks are lesser-soul's **real** repo checks (no simulated gates). Missing checks are
recorded `BLOCKED`, never "fixed" by weakening a gate.

| ID | Category | Check |
|----|----------|-------|
| QUA-1 | Quality | `npm run typecheck` (tsc `--noEmit`) |
| CON-1 | Consistency | every static JSON / JSON-LD object under `cdk/site/static/` is well-formed |
| CON-2 | Consistency | frozen-forever invariant for `/ns/agent-attribution/v1` (`@context.lessersoul`, `agentAttribution` `@id`/`@type`) |
| COM-1 | Completeness | `npm run build:site` (FaceTheory SSG build) |
| COM-2 | Completeness | `cdk synth -c stage=lab` (offline, no deploy) |
| COM-3 | Completeness | exact toolchain pins (`aws-cdk-lib`, `aws-cdk`, `constructs`, FaceTheory, node engine) |
| SEC-1 | Security | supply-chain: npm lockfile present + GitHub Actions pinned by commit SHA |
| SEC-2 | Security | secrets hygiene: no tracked credential material |
| CMP-1 | Compliance | profile resolution from `gov-infra/pack.json` (`software_repo_gov_infra`) |
| CMP-2 | Compliance | genome checksum verification (child-side sha256 vs authoritative govern genome index) |
| CMP-3 | Compliance | exact head/ref attestation for the commit under decision |
| MAI-1 | Maintainability | CI hook: `.github/workflows/ci.yml` invokes the verifier |
| MAI-2 | Maintainability | change scope constrained to `gov-infra/**` + `.github/workflows/ci.yml` |
| DOC-1 | Docs | this `gov-infra/README.md` is present |

## Namespace-genome provenance

The authoritative governance material was resolved from the **equaltoai namespace govern
lifecycle-pack genome** (`genome_commit bc41187efb6f5b3c3bfb4d9295836d4e071941d7`) via
`namespace_governance_profile_get`, `govern_lifecycle_turn`, the scaffold-inventory, the
genome index, and the `gov_rubric_report.v1` schema descriptor. Each namespace-provided
artifact was **checksum-verified child-side** against the genome index (the honesty gate;
the genome's own `pins.json` is never trusted). See `gov-infra/genome-provenance.json`.

The canonical maximal genome verifier template (`templates/gov-verify-rubric.template.sh`,
sha256 `033c696f…`) was inspected and checksum-recorded. It is **not** used verbatim: it
runs unit / integration / coverage / Go-module / SAST / vuln checks that a thin static-site
publisher does not have, so an empty-token fill would emit `BLOCKED` and a fabricated fill is
forbidden. A lean, repo-appropriate verifier is materialized instead — the accepted pattern
for the sibling `software_repo_gov_infra` stewards (`lesser-body`, `lesser-host`). The one
genome artifact committed byte-for-byte is the report schema at
`gov-infra/schemas/gov-rubric-report.schema.json` (sha256 `97ab13ba…`, byte-verified),
against which the emitted report validates.

## Authority (denied)

This governance surface grants **no** deploy, merge, branch-delete, signing, cloud-mutation,
or repository-mutation authority. Signing is retired. MCP guidance does not replace repo CI.
The verifier never changes `/ns/*` namespace semantics, touches deploy config, or mutates
cloud/on-chain state.
