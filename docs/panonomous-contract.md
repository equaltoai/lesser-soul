# Panonomous contract

Status: public static contract for Project 48 M3 issues `#16` and `#17`.

Stable public artifacts after site build/deploy:

- `https://spec.lessersoul.ai/contracts/panonomous/soul-document/`
- `https://spec.lessersoul.ai/contracts/panonomous/soul-document/v1/schema.json`
- `https://spec.lessersoul.ai/contracts/panonomous/agent-naming/`
- `https://spec.lessersoul.ai/contracts/panonomous/agent-naming/v1/vocabulary.json`

These paths do not alter `https://spec.lessersoul.ai/ns/agent-attribution/v1`.

## Source evidence inspected

- lesser-soul static site and namespace behavior: `cdk/site/build.ts`, `cdk/site/faces.ts`,
  `cdk/lib/lesser-soul-site-stack.ts`, and `cdk/site/static/ns/agent-attribution/v1`.
- AgentSoul contract evidence in local `theory-mcp-server` source:
  - `internal/mcp/tools/agent_soul_tools.go` exposes `agent_soul_get` and `agent_soul_upsert` with `agent_id`, required
    `body`, optional `summary`, draft-only behavior, and size-bounded schemas.
  - `internal/models/agent_soul.go` defines the stored draft record: `client_namespace`, `agent_id`, `tenant_id`,
    `soul_version`, `lifecycle_state`, `body`, optional `summary`, `updated_by_subject_id`, timestamps, and `version`.
  - `internal/content/soul_service.go` increments `soul_version` and storage `version` on each save, preserves creation
    time, requires active child agents, and returns draft records only through `GetDraft`.
  - `SPEC.md` describes `agent_soul_get`/`agent_soul_upsert` as draft-only, non-memory, non-publishing tools.
- Lesser naming evidence in local `lesser` source:
  - `pkg/common/validation.go` defines `UsernamePattern = ^[a-zA-Z0-9_-]{1,30}$` and `ValidateUsername`.
  - `ValidateUsernameParamID` delegates to `ValidateUsername`.
  - `cmd/api/handlers/agents.go` uses `ValidateUsernameParamID` for `agent_username` in the delegate path.
  - PR `#1235` / commit `29f70f30854fb4c5df7ee3b9bc18a187a93b4b85` documents the M2/L2 delegate contract in
    `docs/contracts/agent-delegate-api.md`.
- lesser-body read-only inspection found no exact `agent_soul_*`, `AgentSoul`, `Panonomous`, or `soul-document`
  provisional docs/usages in the local checkout at the inspected commit. Later body milestones can cite this contract
  without this milestone editing body.

## Contract notes

The soul-document schema follows the minimal proven AgentSoul shape: a required Markdown-friendly `body`, optional
non-blank `summary`, server-assigned monotonic `soul_version`, draft lifecycle, attribution to the authenticated subject,
and timestamps. V1 intentionally does not define machine-readable body sections; headings inside `body` are plain text.

The naming vocabulary intentionally mirrors Lesser's current username validator for Ptah-created agents that will become
or address Lesser actors. It distinguishes `agent_id`/`agent_username` from `soul_agent_id`, which is a separate Lesser
Soul registry identifier and must not be validated with username rules.
