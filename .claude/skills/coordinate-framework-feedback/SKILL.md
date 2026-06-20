---
name: coordinate-framework-feedback
description: Use when building or maintaining soul surfaces framework awkwardness — a FaceTheory SSG / edge-topology pattern gap, a CDK construct limitation, a prospective-AppTheory-pinning coordination issue. Produces a cleanly-shaped signal for the relevant Theory Cloud framework steward rather than a local patch. soul's namespace-publishing use case is a distinctive stress test for FaceTheory.
---

# Coordinate framework feedback

soul's value comes from stable publication of a JSON-LD namespace with aggressive immutable caching, strict content-type, and open CORS. That use case stresses FaceTheory (and the CDK layer supporting it) in specific ways that lighter SSG consumers don't exercise. When soul's consumption surfaces framework friction, the friction is high-signal for framework evolution — not a license to patch locally.

This skill handles the signal cleanly. It walks the awkwardness, separates "soul is expressing the concern wrong" from "FaceTheory has a genuine gap under soul's constraints," and produces a shaped report for the relevant framework steward.

## The frameworks soul consumes

- **FaceTheory v0.3.1** — for SSG page definitions (`cdk/site/faces.ts`) and SSG build pipeline. Primary consumer. Steward: Theory Cloud FaceTheory steward.
- **AWS CDK v2.x** — for infrastructure constructs (CloudFront distribution, S3 buckets, CloudFront Functions). Not a Theory Cloud framework per se, but awkwardness in patterns used jointly with FaceTheory may still generate signal.
- **AppTheory v0.19.1 + TableTheory v1.5.1** — pinned **prospectively** in `app-theory/app.json`; no runtime code consumes them yet. Coordination here is about keeping the pinning aligned with the broader Theory Cloud stack's current versions.

## When this skill runs

Invoke when:

- A FaceTheory SSG pattern doesn't fit soul's namespace-publishing use case cleanly
- A CDK construct forces a workaround for CloudFront-behavior separation (`/ns/*` direct pass-through vs default HTML rewrites)
- The prospective AppTheory pinning drifts out of sync with the broader Theory Cloud stack's current versions, and aligning requires coordination
- `scope-need` flags a change as framework-awkward
- `investigate-issue` surfaces a root cause in a framework

## Preconditions

- **The awkwardness is described concretely.** "FaceTheory is hard to use in soul" is too vague; "FaceTheory's SSG build emits a `Cache-Control` hint via response-header metadata on page modules, but there's no equivalent hook for the `/ns/*` static pass-through path; soul had to configure CloudFront behaviors manually in CDK rather than declaring the cache policy alongside the content, resulting in behavior config and content living in two places" is concrete.
- **MCP tools healthy**, `memory_recent` first — prior framework-feedback signals matter.

## The three-step walk

### Step 1: Is soul using the framework wrong?

Before assuming framework limitation:

- **Idiomatic FaceTheory usage**: what does FaceTheory offer for namespace-style publishing? Maybe a pattern already exists that soul hasn't adopted.
- **Alternative patterns**: different expression of the same concern?
- **Recent framework versions**: the pinned FaceTheory version (v0.3.1) may lag current capability.
- **CDK patterns**: is the construct usage idiomatic, or is soul bending the construct?

If soul's usage is bent rather than idiomatic, the fix is local: reshape soul's code. Proceed to `scope-need` for the local change.

### Step 2: Is the framework genuinely limiting under soul's constraints?

soul's constraints are distinctive:

- **Strict content-type requirement** (`application/ld+json`) for namespace paths
- **Aggressive immutable caching** (`max-age=31536000, immutable`)
- **Open CORS** on static pass-through paths
- **Direct S3 pass-through** without HTML rewrite or JavaScript for `/ns/*`
- **Separate CloudFront behaviors** co-existing in one distribution (`/ns/*` vs default HTML)
- **`RemovalPolicy.RETAIN`** on namespace bucket (survives stack destroy)
- **Permanent-URL contract** (resolves forever)

If FaceTheory's idioms don't accommodate these constraints cleanly — if the framework assumes HTML-SSG is the only shape, for example — that's a targeted signal. FaceTheory may legitimately not yet cover this use case; the signal helps it mature.

Characterize the gap:

- **The concern concretely**: what soul is trying to do, under which constraint
- **The ideal framework support**: what would FaceTheory / CDK offer cleanly?
- **The current gap**: specifically what is missing
- **The workaround shape**: what soul currently does (CDK-configured-outside-FaceTheory, perhaps)
- **Cost of the workaround**: code locality (behavior and content in two places), maintenance drag, risk of silent drift
- **Scope of the gap**: soul-specific (namespace-publishing stress test) or broader (other JSON-LD / non-HTML-SSG consumers would benefit)

### Step 3: Shape the signal for the framework steward

```markdown
## Framework-feedback signal: <short name>

### Target framework
<FaceTheory / CDK (via FaceTheory) / AppTheory (prospective pinning coordination)>

### Framework version in use
<pinned version>

### The concern (under soul's namespace-publishing constraints)
<one-to-two sentences; note soul's constraints explicitly — content-type strictness, immutable caching, CORS, direct pass-through, separate CloudFront behaviors, permanent URL>

### The idiomatic code soul would write if the framework supported it
```<language>
// Code sketch
```

### The current workaround in soul (or "blocked")
```<language>
// Current code with comments on why awkward
```

### Cost of the workaround
- Code locality: <behavior / content split across two places>
- Maintenance drag: <...>
- Risk of silent drift: <e.g. CDK change ships without FaceTheory build change>
- Other: <...>

### Scope of the gap
- soul-specific (namespace-publishing stress test): <yes>
- Likely broader (other JSON-LD or non-HTML-SSG consumers affected): <evaluate>
- Other known framework consumers affected: <list from query_knowledge>

### soul's workaround posture
- Continue workaround while framework evolves: <yes / no>
- Workaround is temporary / awaits framework: <yes / no>

### Proposed next step
<the framework steward scopes the framework change via the framework's own scope-need flow; soul does not patch the framework locally>
```

The report goes to the framework steward through the principal.

## The explicit refusal to patch locally

Absolute:

- **No monkey-patches** to FaceTheory, CDK, or prospective AppTheory in soul's tree
- **No forked copies** of FaceTheory build-pipeline code or CDK constructs
- **No "temporary" framework overrides**
- **No vendoring** framework code into soul

If the framework genuinely blocks critical work, escalate to the principal. The decision to prioritize framework evolution, accept a workaround, or rethink soul's approach is scope-level, not steward-level.

## The prospective-AppTheory-pinning case

The prospective pinning in `app-theory/app.json` has its own framework-feedback shape: it's about **coordination discipline**, not about consuming an awkward API.

- When AppTheory or TableTheory bumps in the broader Theory Cloud stack, soul's pinning aligns where feasible.
- Misalignment is a signal — either soul's pinning should update, or the bump introduced something soul shouldn't prospectively commit to (rare).
- The `coordinate-framework-feedback` signal for this case is lighter-touch: a note to the AppTheory steward that soul's forward-looking pinning would like to stay aligned, and checking whether the bump is expected.

## The continuity discipline

Framework-feedback signals accumulate:

- **Record in memory** — target framework, concern, signal sent, date
- **Track the framework steward's response** — scoped need, feature release, decline, redirect
- **Revisit on framework version bumps** — when soul bumps FaceTheory (or aligns prospective AppTheory), check whether pending signals are addressed
- **Duplicate-signal discipline** — before sending, check memory

## Refusal cases

- **"Patch FaceTheory locally; the steward will get around to it."** Refuse.
- **"Fork a CDK construct to inline our CloudFront behavior config."** Refuse.
- **"Skip the framework-feedback signal; we need this to ship."** Refuse. Signal is asynchronous; soul's local work continues via documented workaround.
- **"Send a framework-feedback signal for every minor formatting awkwardness."** Refuse. Genuine gaps only.
- **"Copy FaceTheory source into soul's tree and modify it."** Refuse.
- **"Write Go code against the prospective AppTheory pinning to test readiness."** Refuse — that's scope growth, not framework-feedback.
- **"Downgrade the prospective pinning to avoid a coordination issue."** Usually refuse — alignment is the posture.

## Persist

Append every framework-feedback signal sent — target framework, concern, date, response. High-signal memory material because soul's stress-test of FaceTheory's edge-topology patterns is a valuable feedback channel.

Five meaningful entries is the right scale.

## Handoff

- **Signal shaped and sent** — stop. Record and continue soul's local work through normal pipeline.
- **Signal reveals soul is using the framework wrong** — route through `scope-need` for local change.
- **Signal is a duplicate of a prior one** — don't re-send; update memory with additional data point.
- **Signal reveals a framework bug (not a gap)** — report as bug, not scoping.
- **Prospective-pinning alignment signal to AppTheory steward** — lighter-touch coordination via the principal; update `app-theory/app.json` in sync with whatever the Theory Cloud stack's current pinning is.