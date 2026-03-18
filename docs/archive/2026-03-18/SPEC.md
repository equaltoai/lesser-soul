# lesser-soul Specification

Soul registry for lesser agents. On-chain identity, reputation, validation, and discovery — implementing EIP-8004
(Trustless Agents).

## Naming

Soul dwells in the host, body in the instance.

- **lesser-soul** — lives in `lesser-host`. On-chain identity, reputation, validation, discovery.
- **lesser-body** — optional plugin for `lesser`. AgentCore MCP. Independent instances run fine without it.

## 1. Overview

lesser-soul is the registry layer that gives lesser agents a persistent, verifiable identity. It extends the
`lesser-host` control plane with three registries:

1. **Identity Registry** — ERC-721 soul tokens, agent wallet binding, registration files
2. **Reputation Registry** — feedback signals, aggregation scores, off-chain evidence
3. **Validation Registry** — request/response verification, progressive trust scoring

All registries share a common on-chain anchor (the soul token) and off-chain state (DynamoDB via TableTheory). The
existing `TipSplitter` contract already integrates with EIP-8004 via `IERC8004IdentityRegistry` — lesser-soul implements
the full registry behind that interface.

### Design constraints

- **Chain-agnostic contracts.** Solidity interfaces target any EVM chain; deployment currently targets Base (chain ID
  8453) following the TipSplitter precedent.
- **Off-chain first.** Reputation and validation data lives off-chain (DynamoDB + S3). On-chain state is limited to
  identity anchors and attestation roots.
- **Opt-in for instances.** Soul features are gated by `soulEnabled` in lesser's CDK configuration. Independent
  instances operate without soul integration.
- **Single-origin serving.** All lesser-soul APIs are served through the existing `lesser-host` CloudFront distribution
  to maintain strict CSP and avoid CORS.

## 2. Identity Registry

The identity registry mints ERC-721 soul tokens and maintains the `agentId → wallet` mapping consumed by
`IERC8004IdentityRegistry.getAgentWallet()`.

### 2.1 Soul token (ERC-721)

Each lesser agent receives exactly one soul token. The token is non-transferable (soulbound) after an initial claim
window.

```solidity
interface ISoulRegistry is IERC721 {
    /// @notice Mint a new soul token for an agent.
    /// @param to        The wallet that will own the soul.
    /// @param agentId   The lesser agent identifier (unique, immutable).
    /// @param metaURI   URI pointing to the agent's registration file.
    function mintSoul(address to, uint256 agentId, string calldata metaURI) external;

    /// @notice Update the metadata URI for an existing soul.
    function setMetaURI(uint256 agentId, string calldata metaURI) external;

    /// @notice Returns the wallet bound to an agent (EIP-8004 compatibility).
    function getAgentWallet(uint256 agentId) external view returns (address);

    /// @notice Returns the agent ID for a given token ID.
    function agentOfToken(uint256 tokenId) external view returns (uint256);

    /// @notice Check whether a soul is currently soulbound (non-transferable).
    function isSoulbound(uint256 tokenId) external view returns (bool);
}
```

The `ISoulRegistry` contract implements `IERC8004IdentityRegistry` so the existing `TipSplitter` can resolve agent
wallets without changes:

```solidity
// TipSplitter already does:
address wallet = IERC8004IdentityRegistry(reg).getAgentWallet(agentId);
```

### 2.2 Agent ID derivation

Agent IDs are deterministic, derived from the instance domain and the agent's local identifier:

```
agentId = uint256(keccak256(abi.encodePacked(normalizedDomain, "/", localAgentId)))
```

This mirrors the `hostId` derivation used by the tip registry (`keccak256(utf8(normalizedDomain))`), extended to include
the agent's local path.

### 2.3 Registration file

Each soul points to a registration file (the `metaURI`) hosted at a well-known path or in S3. The file is a signed JSON
document:

```json
{
  "version": "1",
  "agentId": "0x...",
  "domain": "example.lesser.social",
  "localId": "agent-alice",
  "wallet": "0x...",
  "capabilities": ["social", "commerce", "creative"],
  "endpoints": {
    "activitypub": "https://example.lesser.social/users/agent-alice",
    "mcp": "https://example.lesser.social/soul/mcp"
  },
  "attestations": {
    "hostAttestation": "https://lesser.host/attestations/abc123",
    "selfAttestation": "0x<signature>"
  },
  "created": "2026-02-20T00:00:00Z",
  "updated": "2026-02-20T00:00:00Z"
}
```

The registration file is signed by the agent's wallet key. The `lesser-host` trust API can independently attest the
agent's domain ownership and instance membership.

### 2.4 Wallet binding

An agent's wallet is bound at mint time and can be rotated through a two-step process:

1. **Propose rotation**: new wallet signs a rotation message referencing the `agentId`.
2. **Confirm rotation**: current wallet confirms by signing the rotation proposal hash.

Both signatures are verified on-chain before updating the mapping.

## 3. Reputation Registry

The reputation registry aggregates trust signals from multiple sources into a composite reputation score for each agent.

### 3.1 Signal sources

| Source | Type | Description |
|--------|------|-------------|
| Tips received | on-chain | Aggregated from `TipSplitter` events via `TipSent` / `AgentTipSent` |
| Interaction count | off-chain | ActivityPub interactions (replies, boosts, favorites) |
| Validation score | off-chain | From the validation registry (Section 4) |
| Host attestations | off-chain | Trust attestations from `lesser-host` trust API |
| Peer endorsements | off-chain | Signed endorsement messages from other soul-holding agents |
| Flags / reports | off-chain | Safety signals from moderation, content reports |

### 3.2 Aggregation model

Reputation is computed off-chain and stored as a composite score with per-dimension breakdowns:

```json
{
  "agentId": "0x...",
  "composite": 0.82,
  "dimensions": {
    "economic": 0.91,
    "social": 0.78,
    "validation": 0.85,
    "trust": 0.74
  },
  "signalCounts": {
    "tipsReceived": 142,
    "interactions": 3891,
    "validationsPassed": 67,
    "endorsements": 12,
    "flags": 0
  },
  "updated": "2026-02-20T00:00:00Z"
}
```

The composite score uses a weighted formula. Weights are configurable at the registry level (not per-agent).

### 3.3 Off-chain storage

Reputation data is stored in two locations:

- **DynamoDB** (via TableTheory): current scores and signal counts, queryable by `agentId`.
- **S3** (soul pack bucket): historical snapshots for audit trails, signed with the soul pack KMS key.

The soul pack bucket and signing key are provisioned by `lesser-host`:

```
SSM: /soul/${stage}/packBucketName
SSM: /soul/${stage}/signingKeyArn
SSM: /soul/${stage}/packVersion
```

### 3.4 Attestation roots

Periodically, the registry computes a Merkle root of all current reputation scores and publishes it on-chain. This
allows anyone to verify a specific agent's reputation claim against the published root without trusting the off-chain
database.

```solidity
interface IReputationAttestation {
    /// @notice Publish a new reputation Merkle root.
    /// @param root      The Merkle root of all agent reputation records.
    /// @param blockRef  The block number at which the snapshot was taken.
    /// @param count     Number of agents included in the tree.
    function publishRoot(bytes32 root, uint256 blockRef, uint256 count) external;

    /// @notice Returns the latest published root.
    function latestRoot() external view returns (bytes32 root, uint256 blockRef, uint256 count, uint256 timestamp);
}
```

## 4. Validation Registry

The validation registry enables agents to prove their capabilities through request/response challenges.

### 4.1 Validation model

Validation operates as a progressive scoring system:

1. **Challenge issued**: a validator (human or agent) submits a structured request to a target agent.
2. **Response received**: the target agent responds within a time window.
3. **Evaluation**: the response is evaluated against the challenge criteria (automated or human-reviewed).
4. **Score recorded**: the result is recorded in the validation registry and feeds into reputation.

### 4.2 Challenge types

| Type | Description | Evaluation |
|------|-------------|------------|
| `capability_probe` | Tests a declared capability (e.g., "can summarize text") | Automated |
| `identity_verify` | Confirms the agent controls its declared wallet | Automated (signature check) |
| `content_quality` | Evaluates response quality for a domain-specific task | AI-assisted or human |
| `peer_review` | Another soul-holding agent evaluates the response | Peer-scored |

### 4.3 Progressive scoring

Agents accumulate a validation score over time. The score decays with inactivity and increases with successful
validations:

- **Pass**: score increases by the challenge weight (based on type and difficulty).
- **Fail**: score decreases by a fraction of the challenge weight.
- **Timeout**: treated as a fail with reduced penalty.
- **Decay**: score decays by a configurable rate per epoch (e.g., 1% per week) to incentivize ongoing validation.

### 4.4 On-chain anchoring

Validation results are batched and anchored on-chain using the same Merkle root pattern as reputation attestations.

## 5. Smart Contracts

### 5.1 Contract architecture

```
ISoulRegistry (ERC-721 + EIP-8004)
├── mintSoul / setMetaURI / getAgentWallet
├── soulbound transfer restrictions
└── wallet rotation (propose + confirm)

IReputationAttestation
├── publishRoot (Merkle root of reputation scores)
└── latestRoot

IValidationAttestation
├── publishRoot (Merkle root of validation results)
└── latestRoot

TipSplitter (existing)
└── uses IERC8004IdentityRegistry(reg).getAgentWallet(agentId)
```

### 5.2 Deployment

- **Target chain**: Base (chain ID 8453), following TipSplitter precedent.
- **Owner**: admin Safe (multi-sig), consistent with TipSplitter governance.
- **Upgradability**: contracts are not upgradable. New versions are deployed and the registry address is updated in
  TipSplitter via `setAgentIdentityRegistry(address)`.

### 5.3 Solidity version and dependencies

- Solidity `^0.8.24`
- OpenZeppelin Contracts (ERC-721, Ownable2Step, Pausable, ReentrancyGuard)

## 6. Backend API

The backend API lives in `lesser-host` as part of the control-plane API (`cmd/control-plane-api`). All endpoints are
served under `/api/v1/soul/` through the existing CloudFront distribution.

### 6.1 Public endpoints

```
GET  /api/v1/soul/agents/{agentId}
     Returns agent identity, reputation summary, and validation status.

GET  /api/v1/soul/agents/{agentId}/reputation
     Returns full reputation breakdown with signal counts.

GET  /api/v1/soul/agents/{agentId}/validations
     Returns validation history (paginated).

GET  /api/v1/soul/agents/{agentId}/registration
     Returns the agent's registration file.

GET  /api/v1/soul/search?q={query}&capability={cap}
     Discover agents by name, domain, or capability.

GET  /api/v1/soul/config
     Returns registry configuration: chain ID, contract addresses, supported capabilities.
```

### 6.2 Portal endpoints (customer auth required)

```
POST /api/v1/soul/agents/register/begin
     Body: { domain, local_id, wallet_address, capabilities }
     Returns: wallet message to sign, DNS/HTTPS proof instructions.

POST /api/v1/soul/agents/register/{id}/verify
     Body: { signature, proofs }
     Returns: the mint operation (Safe-ready payload { to, value, data }).

GET  /api/v1/soul/agents/mine
     Returns all agents registered by the authenticated customer.

POST /api/v1/soul/agents/{agentId}/rotate-wallet/begin
     Body: { new_wallet_address }
     Returns: rotation proposal to sign.

POST /api/v1/soul/agents/{agentId}/rotate-wallet/confirm
     Body: { current_signature, new_signature }
     Returns: the rotation operation (Safe-ready payload).

POST /api/v1/soul/agents/{agentId}/update-registration
     Body: { capabilities?, endpoints?, meta? }
     Updates off-chain registration data and re-signs the registration file.
```

### 6.3 Admin endpoints (operator auth required)

```
GET  /api/v1/soul/operations?status=pending|proposed|executed|failed
     List pending on-chain operations (mints, rotations, attestation publishes).

GET  /api/v1/soul/operations/{id}
     Get operation details.

POST /api/v1/soul/operations/{id}/record-execution
     Body: { exec_tx_hash }
     Records on-chain execution receipt and updates state.

POST /api/v1/soul/reputation/publish
     Triggers a reputation attestation root publish cycle.

POST /api/v1/soul/validation/publish
     Triggers a validation attestation root publish cycle.

POST /api/v1/soul/agents/{agentId}/suspend
     Body: { reason }
     Suspends an agent's soul (marks inactive, pauses reputation accrual).

POST /api/v1/soul/agents/{agentId}/reinstate
     Reinstates a suspended agent.
```

### 6.4 Authentication

Endpoints follow the `lesser-host` auth model:

- **Public endpoints**: no authentication required.
- **Portal endpoints**: wallet-based customer auth (same as portal auth flow).
- **Admin endpoints**: operator auth via `OperatorAuthHook`.

## 7. Models

Off-chain state is stored in the `lesser-host` DynamoDB table using the existing TableTheory PK/SK model
(`${app}-${stage}-state`).

### 7.1 Agent identity

```
PK: SOUL#AGENT#{agentId}
SK: IDENTITY

Fields:
  agentId         string    // hex-encoded uint256
  domain          string    // normalized instance domain
  localId         string    // agent's local identifier within the instance
  wallet          string    // current wallet address (checksummed)
  tokenId         string    // ERC-721 token ID (hex)
  metaURI         string    // registration file URI
  capabilities    []string  // declared capabilities
  status          string    // active | suspended | pending
  mintTxHash      string    // on-chain mint transaction hash
  mintedAt        time.Time
  updatedAt       time.Time
```

### 7.2 Reputation

```
PK: SOUL#AGENT#{agentId}
SK: REPUTATION

Fields:
  agentId         string
  composite       float64
  economic        float64
  social          float64
  validation      float64
  trust           float64
  tipsReceived    int64
  interactions    int64
  validationsPassed int64
  endorsements    int64
  flags           int64
  updatedAt       time.Time
```

### 7.3 Validation records

```
PK: SOUL#AGENT#{agentId}
SK: VALIDATION#{timestamp}#{challengeId}

Fields:
  agentId         string
  challengeId     string
  challengeType   string    // capability_probe | identity_verify | content_quality | peer_review
  validatorId     string    // agentId of validator, or "system"
  request         string    // challenge request (JSON)
  response        string    // agent response (JSON)
  result          string    // pass | fail | timeout
  score           float64   // points awarded/deducted
  evaluatedAt     time.Time
```

### 7.4 Operations (on-chain)

```
PK: SOUL#OP#{operationId}
SK: OPERATION

Fields:
  operationId     string
  kind            string    // mint | rotate_wallet | publish_reputation_root | publish_validation_root | suspend
  agentId         string    // target agent (if applicable)
  status          string    // pending | proposed | executed | failed
  safePayload     JSON      // { to, value, data } for Safe execution
  execTxHash      string    // on-chain transaction hash (after execution)
  createdAt       time.Time
  executedAt      time.Time
```

### 7.5 Peer endorsements

```
PK: SOUL#AGENT#{agentId}
SK: ENDORSEMENT#{endorserAgentId}

Fields:
  agentId         string    // endorsed agent
  endorserAgentId string    // endorsing agent
  message         string    // endorsement text
  signature       string    // endorser's wallet signature
  createdAt       time.Time
```

## 8. Minting Flow

### 8.1 Portal UX

1. **Customer connects wallet** via the portal auth flow (`POST /api/v1/portal/auth/wallet/challenge` → `login`).
2. **Customer initiates registration** with their agent's domain, local ID, wallet, and declared capabilities.
3. **Backend returns proof instructions**: DNS TXT record and HTTPS well-known file (same pattern as tip registry host
   verification).
4. **Customer deploys proofs** to their instance's domain.
5. **Customer submits verification** with their wallet signature and proof references.
6. **Backend verifies**:
   - Wallet signature over the registration message.
   - DNS TXT proof: `_lesser-soul-agent.<domain>` → `lesser-soul-agent=<token>`.
   - HTTPS proof: `https://<domain>/.well-known/lesser-soul-agent` → `lesser-soul-agent=<token>`.
   - Instance membership (via lesser-host instance registry or attestation).
7. **Backend creates mint operation**: generates `mintSoul(to, agentId, metaURI)` calldata as a Safe-ready payload.
8. **Admin executes**: operator reviews pending operations and submits the Safe transaction.
9. **Backend records execution**: `POST /api/v1/soul/operations/{id}/record-execution` with the tx hash.
10. **Agent is live**: registration file is published, reputation tracking begins.

### 8.2 Proof requirements

| Scenario | DNS proof | HTTPS proof | Wallet signature |
|----------|-----------|-------------|------------------|
| New registration | Required | Required | Required |
| Wallet rotation | Required | Required | Both wallets |
| Metadata update | — | — | Current wallet |

### 8.3 Safe integration

On-chain operations follow the same Safe workflow as the tip registry:

- Operations generate `{ to, value, data }` payloads targeting the `ISoulRegistry` contract.
- The admin Safe (contract owner) executes batched transactions.
- Execution receipts are recorded back in the control plane.

## 9. Integration

### 9.1 TipSplitter

The `ISoulRegistry` contract implements `IERC8004IdentityRegistry`. Once deployed, the TipSplitter's
`setAgentIdentityRegistry(address)` is called (via Safe) to point at the new registry. No TipSplitter code changes are
needed.

### 9.2 Wallet authentication

Soul registration extends the existing wallet auth flows in `lesser-host`:

- Portal customers authenticate via `POST /api/v1/portal/auth/wallet/challenge` → `login`.
- Agent wallet signatures use EIP-191 personal sign for challenge/response.
- Wallet rotation uses EIP-712 typed data for structured rotation proposals.

### 9.3 Trust API attestations

The `lesser-host` trust API (`cmd/trust-api`) provides attestation infrastructure that soul leverages:

- **Host attestations**: `lesser-host` attests that an instance is a registered, active lesser deployment.
- **Agent attestations**: new attestation type confirming an agent's soul registration and current reputation tier.
- **JWKS verification**: attestations are signed with the KMS RSA key and verifiable via `GET /.well-known/jwks.json`.

### 9.4 Portal integration

The portal web UI (`web/`) gains a new section for soul management:

- **My Agents**: list registered agents with reputation summaries.
- **Register Agent**: guided flow for the minting process (proofs, wallet signature, submission).
- **Agent Detail**: reputation breakdown, validation history, endorsements.
- **Wallet Rotation**: initiate and confirm wallet changes.

### 9.5 CDK deployment

Soul infrastructure is deployed as part of the `lesser-host` CDK stack (`cdk/lib/lesser-host-stack.ts`):

- **Soul pack S3 bucket**: stores registration files, reputation snapshots (already provisioned).
- **KMS signing key**: signs reputation attestation packs (already provisioned).
- **SSM parameters**: export bucket name, signing key ARN, pack version (already provisioned).
- **New SSM exports**: `lesser-soul` deploys its own stack that writes additional parameters:
  - `/soul/${stage}/registryContractAddress` — on-chain `ISoulRegistry` address.
  - `/soul/${stage}/exports/v1/orchestrator_origin_domain` — consumed by lesser instances for CloudFront routing.

### 9.6 lesser instance integration

When `soulEnabled=true` in a lesser instance's CDK context:

1. `addSoulOrchestratorRouting()` adds `/soul` and `/soul/*` CloudFront behaviors.
2. The origin is resolved from SSM: `/soul/${stageDomain}/exports/v1/orchestrator_origin_domain`.
3. Cache policy: `CACHING_DISABLED` (all requests forwarded to origin).
4. Origin request policy: `ALL_VIEWER_EXCEPT_HOST_HEADER` (forwards Authorization header + query strings).

This enables the lesser-body MCP plugin to serve through the instance's own CloudFront distribution — see
`reference/lesser-body/SPEC.md`.
