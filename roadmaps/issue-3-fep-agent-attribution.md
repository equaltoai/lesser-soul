# Issue 3 Implementation Plan: FEP Agent Social Attribution

Date: 2026-03-18
Audience: EqualtoAI stack maintainers
Issue: `equaltoai/lesser-soul#3`

## Goal

Deliver the remaining work required to submit the Agent Social Attribution FEP without carrying known conformance gaps
into submission.

This repo now owns the namespace infrastructure for this work. Remaining implementation spans `lesser` and the deployed
`spec.lessersoul.ai` infrastructure in `lesser-soul`.

## Outcome Definition

The work is complete when all of the following are true:

- `delegated_by` is guaranteed to serialize as a full `https://` actor URI in federated ActivityPub payloads
- `https://spec.lessersoul.ai/ns/agent-attribution/v1` resolves directly to the JSON-LD context with the correct content
  type
- creator decisions on CC0 scope, named authorship, and namespace versioning are recorded
- the FEP draft no longer needs an implementation caveat for known non-conformance
- the draft is ready for Codeberg submission with only slug assignment and submission date remaining

## Current Status Snapshot

As of 2026-03-18:

- the namespace infrastructure is deployed from `lesser-soul`
- the live hostname for the namespace work is `spec.lessersoul.ai`, not the earlier `lessersoul.ai` placeholder
- issue `#4` inventory is captured in `docs/spec-lessersoul-ai-inventory.md`
- issue `#5` delivery shape is live at `https://spec.lessersoul.ai/ns/agent-attribution/v1`
- `lesser#214` has landed, so the serialization blocker is no longer the critical path here

## Planning Principles

This plan follows two Theory Cloud constraints that matter to delivery shape:

1. **AppTheory boundary discipline**
   - synchronous request handlers should stay narrow and explicit
   - long-running or retryable mutation work should be isolated as tracked background work rather than hidden in request
     paths
   - if a data migration is chosen later, it should use an AppTheory/TableTheory-style job ledger with idempotency,
     record status, and leases

2. **FaceTheory edge topology**
   - when one domain serves both application routes and static documents, prefer CloudFront path-based routing with
     static paths backed by S3 and dynamic paths backed by the application origin
   - for this issue, `/ns/*` should be treated as a static edge concern, not an application redirect or JS-rendered page

## Recommended Technical Decisions

### 1. Submission blocker fix in `lesser`

Use **serialization-time normalization** first to close the standards gap quickly and safely.

Why:

- it removes submission risk without mutating existing DynamoDB records
- it is easier to test at the exact federation boundary that matters for the FEP claim
- it establishes the canonical full-URI format needed by the related delegated-principal work

Follow-up:

- if storage cleanup is still desired after submission, run it as a separate backfill using AppTheory job-ledger
  patterns rather than coupling it to the submission path

### 2. Namespace delivery on `spec.lessersoul.ai`

Use a **static namespace origin** for `/ns/*`, fronted by the existing or newly managed CloudFront distribution.

Why:

- JSON-LD context resolution needs a direct document response, not HTML plus JavaScript redirect logic
- static S3-backed delivery matches the stable, versioned nature of namespace documents
- FaceTheory deployment guidance supports path-based separation between static assets and dynamic application routes

### 3. Repository ownership

- `lesser-soul`: planning, specification, namespace hosting, CloudFront behavior, and deployment automation
- `lesser`: ActivityPub serialization fix and tests
- creator: decisions A, B, and C from the issue

## Workstreams

### WS1. Governance and submission authority

Owner: creator

Deliverables:

- written confirmation that CC0 dedication covers the FEP text and not the reference implementation
- written confirmation of the named human author
- written namespace versioning policy for `/ns/agent-attribution/v1`

Exit criteria:

- all three decisions are recorded in issue comments, an ADR, or the final FEP metadata notes

### WS2. Ops inventory for `spec.lessersoul.ai`

Owner: ops / infra maintainer

Questions to answer:

- which CloudFront distribution currently serves `spec.lessersoul.ai`
- whether Route 53 is authoritative for the domain
- which ACM certificate covers the hostname
- what origin layout serves the static site and namespace path

Deliverables:

- a short inventory note naming the current distribution, origin layout, DNS authority, and certificate ownership
- confirmation that `lesser-soul` owns the deployed `/ns/*` behavior

Exit criteria:

- deployment target is known and no DNS or certificate ambiguity remains

### WS3. ActivityPub conformance fix in `lesser`

Owner: `lesser`

Preferred scope:

- normalize legacy `agent_attribution.delegated_by` values in the ActivityPub serialization path before federation output
- add tests covering both stored short handles and already-canonical full actor URIs
- remove or update any implementation-status caveat in downstream docs once the fix is live

Non-goals for the submission-critical change:

- no bulk DynamoDB mutation
- no opportunistic schema redesign

Acceptance:

- federated `Note` payloads never emit short-handle `delegated_by` values when the field is present
- existing canonical URIs round-trip unchanged
- tests cover the legacy-storage case that originally caused the conformance gap

### WS4. Namespace infrastructure

Owner: `lesser-soul` / infra maintainer

Deliverables:

- static JSON-LD context document stored at the canonical object path for `/ns/agent-attribution/v1`
- CloudFront path routing for `/ns/*`
- response headers enforcing `Content-Type: application/ld+json`
- permissive CORS header for cross-origin JSON-LD fetches
- redirect-free resolution for the namespace path

Preferred deployment shape:

- S3 origin for namespace assets
- CloudFront behavior matching `/ns/*`
- existing dynamic application behavior left untouched for non-namespace routes

Acceptance:

- `curl -H "Accept: application/ld+json" https://spec.lessersoul.ai/ns/agent-attribution/v1` returns the JSON-LD
  document
- no 301, 302, HTML shell, or JavaScript redirect is involved
- repeated fetches are cache-safe and deterministic

### WS5. FEP finalization and submission

Owner: creator with advisory support

Deliverables:

- final pass over the draft after code and namespace verification
- removal of the implementation caveat if WS3 is complete
- submission PR to Codeberg
- slug backfill and submission-date update after registry assignment

Acceptance:

- the submitted document reflects the live namespace URL and the implemented conformance behavior

## Delivery Phases

### P0. Decision and inventory freeze

Purpose:

- remove ambiguity before engineering work fans out

Tasks:

- complete WS1 creator decisions
- complete WS2 infra inventory
- choose namespace deployment target

Exit criteria:

- infra path and policy decisions are fixed

### P1. Submission-critical conformance

Purpose:

- close the known standards gap in the product implementation

Tasks:

- implement WS3 in `lesser`
- add regression tests
- verify canonical output in representative ActivityPub payloads

Exit criteria:

- submission can claim correct `delegated_by` serialization

### P2. Namespace publication

Purpose:

- make the JSON-LD context URL live and stable

Tasks:

- implement WS4 deployment
- upload the context document
- verify headers and redirect-free access

Exit criteria:

- namespace URL is live, stable, and machine-consumable

### P3. FEP finalization

Purpose:

- convert engineering completion into submission readiness

Tasks:

- remove caveat text from the draft if applicable
- confirm authoring and licensing metadata
- perform final proofread against deployed namespace and runtime behavior

Exit criteria:

- draft is ready to submit with only external slug assignment outstanding

### P4. Submission and post-submission cleanup

Purpose:

- complete registry submission and capture follow-up work

Tasks:

- submit to Codeberg
- update slug and submission date
- move the final document into its canonical tracked location
- create follow-up issue if a data backfill still makes sense after submission

Exit criteria:

- FEP is submitted and remaining work is explicitly reduced to non-blocking follow-up items

## Sequence and Dependencies

```
P0 Decision + Inventory Freeze
    |
    +--> P1 `lesser` AP normalization
    |
    +--> P2 namespace infrastructure
             |
             v
        P3 FEP finalization
             |
             v
        P4 submission
```

Notes:

- P1 and P2 may proceed in parallel after P0 completes
- P3 should not start until P1 and P2 both verify successfully
- any storage backfill is explicitly post-submission unless a new requirement emerges

## Detailed Acceptance Checklist

- [ ] creator decisions A, B, and C are recorded
- [x] ops inventory for `spec.lessersoul.ai` is complete
- [x] `lesser` serializes `delegated_by` as a full actor URI in all federated cases
- [x] regression tests cover legacy stored short handles
- [x] `https://spec.lessersoul.ai/ns/agent-attribution/v1` serves the JSON-LD context directly
- [x] namespace response uses `application/ld+json`
- [x] namespace response is redirect-free
- [x] namespace response includes cross-origin access needed for JSON-LD fetches
- [ ] FEP draft caveat is removed or consciously retained with documented reason
- [ ] Codeberg submission is prepared and tracked

## Risks and Mitigations

### Risk: infra ownership is unclear

Mitigation:

- do not start namespace implementation until WS2 names the owning stack and certificate path

### Risk: migration work expands the critical path

Mitigation:

- keep submission unblocking work at serialization time only
- if cleanup is still wanted later, implement it as a job-ledger-backed backfill with explicit retry semantics

### Risk: namespace semantics change after `/v1` goes live

Mitigation:

- require a recorded versioning policy in WS1 before deployment
- treat breaking changes as a new namespace path, not an in-place mutation

### Risk: documentation drifts from deployed behavior

Mitigation:

- make final verification part of P3 and require live URL checks before submission

## Immediate Next Actions

1. Record creator decisions A, B, and C in writing.
2. Update the FEP draft so it names `https://spec.lessersoul.ai/ns/agent-attribution/v1` as the live namespace URL.
3. Remove the implementation caveat if it still references the pre-`lesser#214` serialization gap.
4. Prepare the final Codeberg submission pass and capture slug/date after submission.
