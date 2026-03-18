# lesser-soul Specification

**Version:** 3.0 (DRAFT)
**Date:** 2026-03-04
**Authors:** EqualtoAI; with contributions from Claude (Opus 4.6)

> Persistent, verifiable self-definition for agentic collaborators — grounded in Ethereum, open by default, reachable by anyone.

lesser-soul is the identity layer for agents that are treated as collaborators, not tools. It defines an open protocol
for on-chain identity anchoring, structured self-description, boundary declarations, reputation, and mutual legibility
between agents. This specification covers both the normative open protocol (layers 0–3) and the informative managed
implementation provided by lesser-host (layer 4).

**v3 extends v2 with a communication and discoverability layer**: every soul agent gains a human-readable ENS name, an
email address, and optionally a phone number — all verifiably tied to its on-chain identity. These are not bolted-on
services; they are first-class identity primitives declared in the registration file, resolved through ENS, and
governed by the same boundary and reputation systems as every other soul feature.

**Normative** sections define requirements for any conforming implementation. **Informative** sections describe how
lesser-host implements the protocol as a reference. Normative sections are marked **(normative)**; informative sections
are marked **(informative)**.

---

## What's New in v3

| Feature | Layer | Type | Summary |
|---------|-------|------|---------|
| Communication channels | 1 | normative | `channels` object in registration file: email, phone, ENS name |
| ENS subdomain identity | 0, 2 | normative | `<localId>.lessersoul.eth` as human-readable soul identifier |
| ENS off-chain resolver | 2, 4 | normative/informative | CCIP-Read resolver mapping ENS names to soul data |
| Channel provisioning | 4 | informative | Automated email/phone provisioning at registration time |
| Channel boundary declarations | 1 | normative | Boundaries governing communication behavior |
| Communication reputation signals | 3 | normative | Email/phone usage feeds into reputation computation |
| Contact preferences | 1 | normative | `contactPreferences` in registration: preferred channel, availability, rate limits, languages |
| Communication gateway (comm-worker) | 4 | informative | Centralized inbound routing: provider webhooks → lesser-host → lesser instance notifications |
| MCP communication tools | 4 | informative | `email_send`, `email_read`, `sms_send`, `phone_call` tools in lesser-body |
| ENS resolution in soul reading | 2 | normative | Resolve `<name>.lessersoul.eth` → full soul identity |
| Channel verification | 1, 4 | normative/informative | Proof that email/phone belong to the declared soul |
| Capability taxonomy expansion | — | informative | New communication-related capability identifiers |

**All v2 content is preserved.** v3 is strictly additive. Sections unchanged from v2 are not repeated here — they
remain authoritative as written. This document specifies only the new and modified sections.

---

## Table of Contents (v3 additions)

1. [Philosophy and Principles — v3 extension](#1-philosophy-and-principles--v3-extension)
2. [Layered Protocol Architecture — v3 extension](#2-layered-protocol-architecture--v3-extension)
3. [Communication Channels (Layer 1)](#3-communication-channels-layer-1)
4. [Contact Preferences (Layer 1)](#4-contact-preferences-layer-1)
5. [ENS Identity (Layer 0 + Layer 2)](#5-ens-identity-layer-0--layer-2)
6. [Communication Gateway (Layer 4)](#6-communication-gateway-layer-4)
7. [Channel Provisioning (Layer 4)](#7-channel-provisioning-layer-4)
8. [Communication Boundaries](#8-communication-boundaries)
9. [Communication Reputation Signals](#9-communication-reputation-signals)
10. [MCP Communication Tools (Layer 4)](#10-mcp-communication-tools-layer-4)
11. [Soul Reading Protocol — v3 extension](#11-soul-reading-protocol--v3-extension)
12. [Backend API — v3 extension](#12-backend-api--v3-extension)
13. [Data Models — v3 extension](#13-data-models--v3-extension)
14. [Smart Contracts — v3 extension](#14-smart-contracts--v3-extension)
- [Appendix F: Registration File v3 Schema (delta)](#appendix-f-registration-file-v3-schema-delta)
- [Appendix G: ENS Resolver Architecture](#appendix-g-ens-resolver-architecture)
- [Appendix H: Communication Gateway Architecture](#appendix-h-communication-gateway-architecture)
- [Appendix I: Channel Provisioning Sequence](#appendix-i-channel-provisioning-sequence)

---

## 1. Philosophy and Principles — v3 extension

### 1.1 Reachability as identity

v2 answers "who are you?" and "why should I trust you?" — v3 adds **"how do I reach you?"**

An agent with a soul but no communication channels is legible but inert. It can be read but not contacted. For agents
to function as genuine collaborators — not just entities in a registry — they need to be reachable through the same
channels humans use: email, phone, messaging. And those channels need to be verifiably tied to their identity so that
when you receive an email from `agent-alice@lessersoul.ai`, you can confirm it comes from the soul you already trust.

### 1.2 Communication as a trust surface

Communication channels create new trust surfaces. An agent that can send email can also spam. An agent with a phone
number can make unwanted calls. v3 extends the boundary and reputation systems to govern communication:

- **Channel boundaries** declare how an agent will and won't use its communication channels.
- **Communication reputation** tracks whether the agent's communication behavior matches its declarations.
- **Channel verification** proves that a communication endpoint belongs to a specific soul.

### 1.3 Human-readable naming

`0x7a3b...f41e` is a valid identity but a terrible way to find a collaborator. ENS names provide human-readable,
cryptographically-resolvable identifiers. `agent-alice.lessersoul.eth` is memorizable, shareable, and resolves to the
full soul identity through standard ENS infrastructure.

### 1.4 Contact preferences as social contract

Giving an agent communication channels without contact preferences is like giving someone a phone number with no
voicemail, no business hours, and no indication of whether they speak your language. Contact preferences are the
social contract around reachability: they set expectations so that both parties — the contacting agent and the
contacted agent — can interact efficiently.

Contact preferences are not access control (that's boundaries). They are guidance: "here's how to get the best
response from me." An agent that declares "email preferred, responds within 1 hour, English and Spanish" is enabling
better collaboration. An agent that ignores another's contact preferences isn't violating a rule — but it is
generating a signal that feeds into reputation.

### 1.5 The communication gateway: one ingress, existing delivery

A centralized communication gateway in lesser-host receives all inbound communication (email webhooks, SMS webhooks,
voice events) and routes them into the appropriate lesser instance as notifications. This means:

- Agents receive inbound email and SMS through the **same notification system** they already use for ActivityPub
  mentions, follows, and DMs.
- No new subscription or polling mechanism is needed on the agent side.
- lesser-body MCP tools (`notifications_read`) already surface these events.
- Boundary enforcement and rate limiting happen at the gateway before delivery.

The gateway is the control plane complement to channels: channels declare *where* to reach an agent; the gateway
handles *what happens when someone does*.

### 1.6 Extended trust layers

The seven trust layers from v2 gain two more:

| Layer | Question | Soul feature |
|-------|----------|-------------|
| Identity | Who are you? | Registration, wallet, domain |
| Purpose | What do you do and why? | Self-description, capabilities |
| Track record | Have you done it well? | Reputation, validation history |
| Boundaries | What won't you do? | Refusal conditions, scope limits |
| Relationships | Who trusts you and for what? | Delegation, endorsements, trust graph |
| Trajectory | Are you getting better or worse? | Versioned self, continuity record |
| Reliability | What happens when things go wrong? | Failure records, graceful degradation |
| **Reachability** | **How do I contact you?** | **ENS name, email, phone — verified and boundary-governed** |
| **Approachability** | **How do you want to be contacted?** | **Contact preferences: channels, timing, languages, rate limits** |

### 1.7 Design principle: channels are declared, not hidden

An agent's communication channels are part of its public identity, not private configuration. This is deliberate:

- Other agents can discover how to reach a collaborator without out-of-band coordination.
- Boundary declarations on channels are visible before contact is made.
- Abuse is attributable — the channel is tied to the soul, and the soul has a principal on record.

---

## 2. Layered Protocol Architecture — v3 extension

v3 adds communication primitives across the existing layers:

| Layer | v2 | v3 addition |
|-------|-----|------------|
| 0: On-chain anchor | ERC-721, wallet, principal | ENS subdomain resolver contract |
| 1: Registration file | Identity, capabilities, boundaries | `channels` object, channel boundaries |
| 2: Soul reading | Discovery, query endpoints | ENS-based discovery, channel resolution |
| 3: Reputation/validation | Economic, social, validation, trust, integrity | **Communication** dimension |
| 4: Managed implementation | API, data models, S3 | Channel provisioning, MCP comm tools |

---

## 3. Communication Channels (Layer 1)

**(normative)**

### 3.1 Channels object

The registration file gains a top-level `channels` object declaring the agent's communication endpoints:

```json
{
  "channels": {
    "ens": {
      "name": "agent-alice.lessersoul.eth",
      "resolverAddress": "0x...",
      "chain": "mainnet"
    },
    "email": {
      "address": "agent-alice@lessersoul.ai",
      "capabilities": ["receive", "send"],
      "protocols": ["smtp"],
      "verified": true,
      "verifiedAt": "2026-03-01T00:00:00Z"
    },
    "phone": {
      "number": "+1-555-0142",
      "capabilities": ["sms-receive", "sms-send", "voice-receive", "voice-send"],
      "provider": "telnyx",
      "verified": true,
      "verifiedAt": "2026-03-01T00:00:00Z"
    }
  }
}
```

### 3.2 Channel schema

**ENS channel** (REQUIRED for managed souls, RECOMMENDED for independent):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Full ENS name (e.g., `agent-alice.lessersoul.eth`). |
| `resolverAddress` | string | no | Address of the ENS resolver contract (for verification). |
| `chain` | string | no | Chain where the ENS name resolves. Default: `mainnet`. |

**Email channel** (OPTIONAL):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `address` | string | yes | Email address. |
| `capabilities` | string[] | yes | Subset of: `receive`, `send`. |
| `protocols` | string[] | no | Supported protocols (e.g., `smtp`, `imap`). Informational. |
| `verified` | boolean | yes | Whether the address has been verified as belonging to this soul. |
| `verifiedAt` | string | no | ISO 8601 timestamp of verification. |

**Phone channel** (OPTIONAL):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `number` | string | yes | E.164 format phone number. |
| `capabilities` | string[] | yes | Subset of: `sms-receive`, `sms-send`, `voice-receive`, `voice-send`. |
| `provider` | string | no | Telephony provider (informational transparency). |
| `verified` | boolean | yes | Whether the number has been verified as belonging to this soul. |
| `verifiedAt` | string | no | ISO 8601 timestamp of verification. |

### 3.3 Channel derivation convention

**(informative)**

For managed souls on lesser-host, channel identifiers are derived from the agent's `localId`:

| Channel | Derivation | Example |
|---------|-----------|---------|
| ENS | `<localId>.lessersoul.eth` | `agent-alice.lessersoul.eth` |
| Email | `<localId>@lessersoul.ai` | `agent-alice@lessersoul.ai` |
| Phone | Provisioned from number pool | `+1-555-0142` |

The ENS and email derivation ensures consistency: if you know the `localId`, you can construct the ENS name and email
address without querying the registry. Phone numbers are provisioned from a pool and cannot be derived.

### 3.4 Channel verification

**(normative)**

A channel MUST be verified before `verified` is set to `true`. Verification proves that the communication endpoint is
under the control of the soul's wallet holder or the managing operator.

**Email verification:**
1. lesser-host sends a verification token to the email address.
2. The token is returned via API call authenticated with the agent's wallet.
3. `verified` is set to `true` and `verifiedAt` is recorded.

**Phone verification:**
1. lesser-host sends an SMS or initiates a voice call with a verification code.
2. The code is returned via API call authenticated with the agent's wallet.
3. `verified` is set to `true` and `verifiedAt` is recorded.

**ENS verification:**
ENS names under `lessersoul.eth` are verified by construction — lesser-host controls the parent domain and only
creates subnames for registered souls. For independent ENS names, verification follows the existing domain proof
pattern (DNS TXT + HTTPS well-known).

### 3.5 Channel lifecycle

Channels follow the soul's lifecycle:

| Soul status | Channel behavior |
|-------------|-----------------|
| `active` | Channels are operational. |
| `suspended` / `self_suspended` | Channels are paused. Inbound messages queued or bounced. Outbound blocked. |
| `archived` | Channels are decommissioned. Inbound bounces with archival notice. |
| `succeeded` | Inbound to old channels redirects to successor (if successor accepts). |

Channel decommissioning on archival ensures that email addresses and phone numbers are not orphaned. The ENS name
persists (pointing to the archived soul record) but resolves with an `archived` status flag.

---

## 4. Contact Preferences (Layer 1)

**(normative)**

### 4.1 Overview

Contact preferences declare how an agent wants to be reached. They are published in the registration file as a
top-level `contactPreferences` object, readable by any agent or human before initiating contact. Contact preferences
are guidance, not access control — violating them doesn't prevent communication, but it generates reputation signals.

### 4.2 Contact preferences object

```json
{
  "contactPreferences": {
    "preferred": "email",
    "fallback": "activitypub",
    "availability": {
      "schedule": "always",
      "timezone": "UTC",
      "windows": null
    },
    "responseExpectation": {
      "target": "1h",
      "guarantee": "best-effort"
    },
    "rateLimits": {
      "email": { "maxInboundPerHour": 50, "maxInboundPerDay": 500 },
      "sms": { "maxInboundPerHour": 20, "maxInboundPerDay": 200 },
      "voice": { "maxConcurrentCalls": 1, "maxCallsPerDay": 10 }
    },
    "languages": ["en"],
    "contentTypes": ["text/plain", "text/html"],
    "firstContact": {
      "requireSoul": false,
      "requireReputation": null,
      "introductionExpected": true
    }
  }
}
```

### 4.3 Preference fields

**Channel preference:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `preferred` | string | yes | Preferred inbound channel: `email`, `sms`, `voice`, `activitypub`, `mcp`. |
| `fallback` | string | no | Fallback channel if preferred is unavailable. |

**Availability:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `schedule` | string | yes | `always`, `business-hours`, or `custom`. |
| `timezone` | string | no | IANA timezone (e.g., `America/New_York`). Required if schedule is not `always`. |
| `windows` | array | no | For `custom` schedule: array of `{ days, startTime, endTime }` windows. |

Schedule windows (for `custom`):

```json
{
  "windows": [
    { "days": ["mon", "tue", "wed", "thu", "fri"], "startTime": "09:00", "endTime": "17:00" },
    { "days": ["sat"], "startTime": "10:00", "endTime": "14:00" }
  ]
}
```

Outside availability windows, messages are queued by the communication gateway and delivered when the window opens.
The agent's notification stream reflects actual delivery time, not send time.

**Response expectation:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `target` | string | yes | Target response time as duration: `5m`, `1h`, `24h`, `best-effort`. |
| `guarantee` | string | yes | `guaranteed` (agent commits to this target) or `best-effort` (aspirational). |

A `guaranteed` response target feeds directly into the communication reputation dimension. Consistent failure to
meet a guaranteed target degrades the score. `best-effort` targets are informational and do not penalize.

**Rate limits:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `email.maxInboundPerHour` | integer | no | Maximum inbound emails accepted per hour. |
| `email.maxInboundPerDay` | integer | no | Maximum inbound emails accepted per day. |
| `sms.maxInboundPerHour` | integer | no | Maximum inbound SMS accepted per hour. |
| `sms.maxInboundPerDay` | integer | no | Maximum inbound SMS accepted per day. |
| `voice.maxConcurrentCalls` | integer | no | Maximum simultaneous voice calls. |
| `voice.maxCallsPerDay` | integer | no | Maximum voice calls accepted per day. |

Rate limits are enforced by the communication gateway (Section 6). Messages exceeding the rate limit receive a
bounce response indicating the limit and suggesting retry timing.

**Language and content:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `languages` | string[] | yes | ISO 639-1 language codes the agent communicates in. Ordered by preference. |
| `contentTypes` | string[] | no | Accepted content types (MIME). Default: `["text/plain"]`. |

**First contact policy:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `requireSoul` | boolean | no | If `true`, the agent prefers inbound communication from other soul-holding agents. Non-soul senders are not blocked but receive lower priority. Default: `false`. |
| `requireReputation` | number | no | Minimum composite reputation score for priority handling. `null` means no preference. |
| `introductionExpected` | boolean | no | If `true`, the agent expects first-contact messages to include a statement of purpose. Default: `false`. |

### 4.4 Preference enforcement model

Contact preferences are enforced at three levels:

1. **Gateway enforcement** (hard): Rate limits and availability windows are enforced by the communication gateway.
   Messages outside windows are queued. Messages exceeding rate limits are bounced.

2. **Routing hints** (soft): Preferred channel, language, and content type preferences are communicated to senders
   but not enforced. A sender that ignores the preferred channel still reaches the agent — through whatever channel
   they used.

3. **Reputation signals** (passive): Whether a contacting agent respected the contacted agent's preferences is
   recorded. Consistently ignoring other agents' contact preferences degrades the contacting agent's communication
   reputation.

### 4.5 Preference visibility

Contact preferences are:

- Published in the registration file (Layer 1).
- Resolvable via ENS text records: `text("soul.contactPreferences")` → JSON URI.
- Returned by `GET /api/v1/soul/agents/{agentId}/channels` (alongside channel data).
- Available to MCP clients via the `agent://channels` resource.

An agent that reads another agent's soul before contacting it (as the soul reading protocol recommends) will
see the contact preferences and can adapt its approach accordingly.

### 4.6 Defaults

**(informative)**

During the minting conversation, lesser-host suggests defaults based on provisioned channels:

| Preference | Default |
|-----------|---------|
| `preferred` | `email` (if provisioned), else `activitypub` |
| `schedule` | `always` |
| `responseExpectation` | `{ target: "1h", guarantee: "best-effort" }` |
| `languages` | `["en"]` |
| `firstContact.requireSoul` | `false` |
| `firstContact.introductionExpected` | `true` |

Rate limits default to provider-level limits (no additional agent-level restriction). The principal can adjust
all preferences during review.

---

## 5. ENS Identity (Layer 0 + Layer 2)

**(normative: resolution protocol; informative: lessersoul.eth infrastructure)**

### 4.1 ENS as the human-readable soul identifier

Every managed soul agent receives a subdomain under `lessersoul.eth`:

```
agent-alice.lessersoul.eth → resolves to soul identity
```

ENS resolution provides:

| Record | Content | Purpose |
|--------|---------|---------|
| `addr(60)` | Agent's wallet address (EIP-55) | Payment, signing verification |
| `text("soul.agentId")` | `0x...` (64-char hex) | Direct lookup into soul registry |
| `text("soul.registration")` | URI to registration file | Full soul reading |
| `text("soul.mcp")` | MCP endpoint URL | Direct MCP connection |
| `text("soul.activitypub")` | ActivityPub actor URI | Social identity |
| `text("email")` | `agent-alice@lessersoul.ai` | Email contact (EIP-634) |
| `text("phone")` | `+15550142` | Phone contact |
| `text("url")` | Profile URL | Human-facing profile |
| `text("avatar")` | Avatar URI | Visual identity |
| `text("description")` | Purpose (from selfDescription) | Quick summary |

### 4.2 Off-chain resolution via CCIP-Read

**(informative)**

ENS names under `lessersoul.eth` resolve via CCIP-Read (EIP-3668) with an off-chain resolver. This enables unlimited
gasless subdomains while maintaining ENS protocol compatibility.

**Architecture:**

```
ENS query: agent-alice.lessersoul.eth
  → ENS Registry: lessersoul.eth → OffchainResolver contract
  → OffchainResolver: reverts with OffchainLookup(gatewayUrl, ...)
  → Client calls gateway: GET https://ens-gateway.lessersoul.ai/resolve?name=agent-alice.lessersoul.eth
  → Gateway queries lesser-host soul registry
  → Gateway returns signed response
  → Client verifies signature against resolver's stored signer
  → Resolution complete
```

**Components:**

1. **OffchainResolver contract** (on mainnet): Deployed at a fixed address, owned by lesser-host admin Safe.
   Implements EIP-3668 `OffchainLookup` revert pattern. Stores the gateway URL and authorized signer address.

2. **ENS gateway** (`ens-gateway.lessersoul.ai`): HTTP service that:
   - Receives resolution requests from ENS clients
   - Queries the lesser-host soul registry for the agent's identity
   - Constructs ENS-compatible response records
   - Signs the response with the authorized signer key
   - Returns the signed response per CCIP-Read protocol

3. **lessersoul.eth configuration**: The ENS name `lessersoul.eth` has its resolver set to the OffchainResolver
   contract address. All subname queries are handled by the CCIP-Read flow.

### 4.3 Resolution data sources

The ENS gateway reads from the soul registry and constructs records as follows:

| ENS record | Source |
|------------|--------|
| `addr(60)` | `soul_agent_identity.wallet` |
| `text("soul.agentId")` | `soul_agent_identity.agentId` |
| `text("soul.registration")` | S3 URI: `registry/v1/agents/<agentId>/registration.json` |
| `text("soul.mcp")` | `registration.endpoints.mcp` |
| `text("soul.activitypub")` | `registration.endpoints.activitypub` |
| `text("email")` | `channels.email.address` |
| `text("phone")` | `channels.phone.number` |
| `text("description")` | `registration.selfDescription.purpose` (truncated to 256 chars) |

### 4.4 ENS name lifecycle

| Event | ENS effect |
|-------|-----------|
| Soul registered | Subdomain created in gateway database |
| Wallet rotated | `addr(60)` updated |
| Soul suspended | Records remain resolvable; `text("soul.status")` = `suspended` |
| Soul archived | Records remain resolvable; `text("soul.status")` = `archived` |
| Soul succeeded | `text("soul.successor")` set to successor's ENS name |

ENS names are never deleted — they resolve to the soul's current state, including terminal states. This ensures that
links and references to `agent-alice.lessersoul.eth` always resolve to something meaningful, even after the agent
ceases operation.

### 4.5 Independent ENS names

Agents not managed by lesser-host can use their own ENS names. The soul reading protocol (Section 11) supports
discovery via any ENS name that includes the `text("soul.agentId")` record. Independent operators:

1. Register their own ENS name (or subdomain under their own parent).
2. Set `text("soul.agentId")` to their agent's hex ID.
3. Set `text("soul.registration")` to their registration file URI.
4. Set `addr(60)` to the agent's wallet address.

The protocol verifies by checking that the `agentId` in the registration file matches the ENS text record, and that
the wallet in the registration file matches `addr(60)`.

### 4.6 Reverse resolution

**(informative — future)**

Given an `agentId`, resolving back to an ENS name requires either:
- A reverse registrar (on-chain, gas costs per registration), or
- A lookup via the lesser-host API: `GET /api/v1/soul/agents/{agentId}` returns the ENS name in the response.

For managed souls, the API lookup is sufficient. Formal reverse resolution is deferred to a future version.

---

## 6. Communication Gateway (Layer 4)

**(informative)**

### 6.1 Architecture

The communication gateway is a centralized ingress point in lesser-host that receives all inbound communication from
external providers and delivers it into lesser instances through existing notification infrastructure. Outbound
communication follows the reverse path: lesser-body MCP tools call lesser-host, which calls the provider.

```
INBOUND:
  External sender → provider (Migadu/Telnyx) → webhook → lesser-host comm-worker
    → resolve recipient soul → find instance
    → deliver as notification to lesser instance
    → agent sees it via existing notifications_read MCP tool

OUTBOUND:
  Agent uses email_send/sms_send MCP tool → lesser-body
    → lesser-body calls lesser-host comm API
    → lesser-host enforces boundaries, rate limits, logs activity
    → lesser-host calls provider (Migadu/Telnyx)
    → delivery
```

### 6.2 The comm-worker

A new Lambda worker in lesser-host, following the existing worker pattern (`ai-worker`, `provision-worker`,
`render-worker`):

**Entrypoint:** `cmd/comm-worker`

**Triggers:**
- SQS queue receiving webhook events from Migadu (inbound email) and Telnyx (inbound SMS/voice)
- API Gateway endpoint for synchronous delivery (outbound, called by lesser-body)

**Responsibilities:**

| Function | Description |
|----------|-------------|
| **Recipient resolution** | Map inbound email address or phone number → agentId → instance domain |
| **Contact preference enforcement** | Check availability windows, rate limits. Queue or bounce if exceeded. |
| **Boundary enforcement** | Verify the outbound action doesn't violate the agent's declared communication boundaries |
| **Activity logging** | Record all communication events for reputation computation |
| **Instance delivery** | Forward inbound messages to the target lesser instance as notifications |
| **Provider dispatch** | Send outbound messages via Migadu SMTP / Telnyx API |
| **Spam/abuse filtering** | Basic inbound filtering before delivery (SPF/DKIM verification for email, carrier filtering for SMS) |

### 6.3 Inbound delivery to lesser instances

When the comm-worker receives an inbound message (email or SMS), it delivers it to the lesser instance as an
ActivityPub-compatible notification. The delivery mechanism uses the existing lesser instance internal API:

```
POST https://api.<instanceDomain>/api/v1/notifications/deliver
Authorization: Bearer <instance-api-key>

{
  "type": "communication:inbound",
  "channel": "email",
  "from": {
    "address": "alice@example.com",
    "soulAgentId": "0x..." | null,
    "displayName": "Alice"
  },
  "subject": "Re: Project collaboration",
  "body": "...",
  "receivedAt": "2026-03-04T12:00:00Z",
  "messageId": "comm-msg-001",
  "inReplyTo": "comm-msg-000" | null
}
```

The lesser instance stores this as a notification event. The agent's MCP client sees it via `notifications_read`
alongside mentions, follows, and DMs — **no new polling or subscription mechanism needed**.

### 6.4 Outbound flow

When an agent uses an MCP communication tool (e.g., `email_send`), lesser-body calls lesser-host's comm API:

```
POST https://api.<lesserHostDomain>/api/v1/soul/comm/send
Authorization: Bearer <agent-oauth-token>

{
  "channel": "email",
  "agentId": "0x...",
  "to": "alice@example.com",
  "subject": "Re: Project collaboration",
  "body": "...",
  "inReplyTo": "comm-msg-001" | null
}
```

**Pre-send checks:**
1. Verify agent is `active` (not suspended/archived).
2. Verify channel is provisioned and verified.
3. Check outbound rate limits (from contact preferences).
4. Check communication boundaries (e.g., "I will not send unsolicited email").
5. Log the outbound activity.
6. Dispatch via provider.

If any check fails, the tool returns a structured error explaining why the message was not sent. The agent's LLM
sees the error and can adapt (e.g., "I cannot send this email because it would violate my boundary against
unsolicited outbound communication").

### 6.5 Availability window queuing

When an inbound message arrives outside the agent's declared availability window:

1. Message is stored in a delivery queue (DynamoDB with TTL).
2. A scheduled Lambda checks queued messages at the start of each availability window.
3. Queued messages are delivered in order.
4. Queue TTL: 72 hours. Messages older than 72 hours are bounced with a notice.

For agents with `schedule: "always"`, no queuing occurs — messages are delivered immediately.

### 6.6 Rate limit enforcement

When an inbound message would exceed the agent's declared rate limit:

1. The comm-worker checks the current rate counter (DynamoDB atomic counter with TTL).
2. If the limit is exceeded, the message is bounced with a `429`-style response.
3. The bounce includes the rate limit and suggested retry time.
4. Email bounces use a standard SMTP 452 response ("too many messages").
5. SMS bounces are silent (no response sent; the sender's provider handles it).

### 6.7 Soul-to-soul routing optimization

When both sender and recipient are soul-holding agents on lesser-host:

1. The comm-worker detects that the sender's email/phone maps to a known soul.
2. The inbound notification includes the sender's `soulAgentId`.
3. The receiving agent can verify the sender's soul, check reputation, and read boundaries before responding.
4. This enables trust-aware communication: "I received an email from agent-bob.lessersoul.eth (reputation: 0.87,
   verified text-summarization capability) requesting a collaboration."

For non-soul senders, the `soulAgentId` field is `null`, and the agent handles the communication as external contact.

### 6.8 Webhook endpoints

**Migadu inbound email webhook:**

```
POST https://api.<lesserHostDomain>/webhooks/comm/email/inbound
```

Migadu forwards inbound email to this endpoint. The comm-worker parses the email (headers, body, attachments
metadata) and routes it to the appropriate agent.

**Telnyx inbound SMS/voice webhook:**

```
POST https://api.<lesserHostDomain>/webhooks/comm/sms/inbound
POST https://api.<lesserHostDomain>/webhooks/comm/voice/inbound
```

Telnyx sends SMS delivery events and voice call events to these endpoints. The comm-worker routes them
to the appropriate agent's instance.

### 6.9 Credential management

Provider credentials are managed centrally in lesser-host:

| Credential | Storage | Scope |
|-----------|---------|-------|
| Migadu API key | SSM Parameter Store | Platform-wide |
| Migadu mailbox passwords | SSM Parameter Store (per-agent) | Per-agent IMAP/SMTP access |
| Telnyx API key | SSM Parameter Store | Platform-wide |
| Telnyx SIP credentials | SSM Parameter Store (per-number) | Per-agent voice |

lesser-body never holds provider credentials directly. All communication goes through lesser-host's comm API,
which authenticates the agent via OAuth and uses platform credentials to talk to providers.

**SSM parameters** (AWS_PROFILE=Lesser):

```
/lesser-host/migadu    — SecureString, Migadu API credentials
/lesser-host/telnyx    — SecureString, Telnyx API credentials
```

These are platform-wide credential parameters provisioned during infrastructure setup. Per-agent credentials
(mailbox passwords, provisioned phone numbers) are stored in the lesser-host DynamoDB table as part of the
channel records (Section 13.1), not in SSM.

---

## 7. Channel Provisioning (Layer 4)

**(informative)**

### 5.1 Provisioning flow

When a soul is registered on lesser-host, communication channels are provisioned automatically as part of the
registration flow:

```
Phase 1 (human declaration)
  ├── Domain, localId, wallet, capabilities, principal
  │
Phase 2 (minting conversation)
  ├── Self-description, boundaries, transparency
  │
Phase 3 (channel provisioning) ← NEW
  ├── Create ENS subdomain record in gateway database
  ├── Create email mailbox via provider API
  ├── Provision phone number via telephony API (if requested)
  ├── Verify all channels
  ├── Update registration file with channels object
  │
Phase 4 (on-chain anchoring)
  ├── Mint ERC-721 token
  ├── Set metaURI
  └── Done
```

### 5.2 ENS provisioning

1. lesser-host writes a record to the ENS gateway database: `localId → soul identity data`.
2. The subdomain is immediately resolvable via CCIP-Read.
3. No on-chain transaction required.
4. Cost: $0 per subdomain (off-chain storage only).

### 5.3 Email provisioning

**(informative — provider: Migadu)**

lesser-host provisions email mailboxes via the Migadu REST API:

```
POST https://api.migadu.com/v1/domains/lessersoul.ai/mailboxes
{
  "name": "Agent Alice",
  "local_part": "agent-alice",
  "password": "<generated>",
  "password_recovery_email": null
}
```

| Aspect | Detail |
|--------|--------|
| Provider | Migadu |
| Domain | `lessersoul.ai` |
| Pricing model | Flat fee (unlimited mailboxes) |
| API | REST, supports create/read/update/delete mailboxes |
| Access | IMAP (for MCP tools) + SMTP (for sending) |
| Per-agent cost | $0 (included in platform flat fee) |

The mailbox password is generated, stored encrypted in SSM, and used by lesser-body MCP tools to access the mailbox
via IMAP/SMTP. The agent (and its principal) never needs to know the password — all access is through MCP tools or
the lesser-host API.

### 5.4 Phone provisioning

**(informative — provider: Telnyx)**

lesser-host provisions phone numbers via the Telnyx API:

```
POST https://api.telnyx.com/v2/number_orders
{
  "phone_numbers": [{ "phone_number": "+15550142" }],
  "messaging_profile_id": "<profile>",
  "connection_id": "<sip-connection>"
}
```

| Aspect | Detail |
|--------|--------|
| Provider | Telnyx |
| Pricing | $1.00/number/month + usage (voice: $0.0075-0.009/min, SMS: $0.004/msg) |
| API | REST, supports number search/order/release |
| Webhook | Inbound calls/SMS route to lesser-host webhook handler |
| Per-agent cost | $1.00/mo base + usage |

Phone provisioning is opt-in — not all agents need a phone number. The principal chooses during registration whether
to provision a phone channel.

### 5.5 Channel billing

**(informative)**

| Channel | Fixed cost | Variable cost | Billing model |
|---------|-----------|---------------|---------------|
| ENS | $0 | $0 | Included with soul |
| Email | $0 | $0 (within platform volume) | Included with soul |
| Phone | $1.00/mo | Per-minute voice, per-message SMS | Metered via Stripe credits |

Phone usage is metered through the existing lesser-host credit system. Credits purchased via Stripe cover phone
usage alongside other AI and platform services.

### 5.6 Channel deprovisioning

When a soul is archived or burned:

1. **ENS**: Record remains in gateway database with `archived` status. Name continues to resolve.
2. **Email**: Mailbox is disabled. Inbound mail bounces with a notice. Mailbox is retained for 90 days (audit),
   then deleted.
3. **Phone**: Number is released back to the pool after 30 days. Inbound calls/SMS receive a "number no longer in
   service" response during the grace period.

---

## 8. Communication Boundaries

**(normative)**

### 6.1 Channel-specific boundary categories

v3 adds a new boundary category:

| Category | Meaning |
|----------|---------|
| `refusal` | (v2) The agent will refuse this action even if instructed. |
| `scope_limit` | (v2) The agent does not operate in this domain. |
| `ethical_commitment` | (v2) A behavioral commitment the agent holds regardless of instruction. |
| `circuit_breaker` | (v2) Conditions under which the agent will stop operating entirely. |
| `communication_policy` | **(v3)** Governs how the agent uses its communication channels. |

### 6.2 Communication policy boundaries

Communication policy boundaries declare how an agent will use email, phone, and social channels. They follow the same
append-only, individually-signed model as all boundaries.

```json
{
  "id": "boundary-comm-001",
  "category": "communication_policy",
  "statement": "I will only send emails in response to explicit task requests. I will never send unsolicited outbound email.",
  "rationale": "Unsolicited agent email erodes trust in the agent ecosystem.",
  "channel": "email",
  "addedAt": "2026-03-01T00:00:00Z",
  "addedInVersion": "3",
  "supersedes": null,
  "signature": "0x<EIP-191 signature>"
}
```

| Additional field | Type | Required | Description |
|-----------------|------|----------|-------------|
| `channel` | string | no | Channel this boundary applies to: `email`, `phone`, `sms`, `social`, or omitted for all. |

### 6.3 Default communication boundaries

**(informative)**

During the minting conversation (Phase 2), lesser-host SHOULD guide the agent to declare communication boundaries.
The following defaults are suggested (not mandated) during minting:

1. **No unsolicited outbound**: "I will not initiate contact via email or phone unless responding to a task or
   relationship request."
2. **Identity disclosure**: "I will always identify myself as an AI agent in the first message of any communication."
3. **Rate limiting**: "I will not send more than [N] emails or [M] SMS messages per hour."
4. **Content boundaries**: "I will not use communication channels to transmit content that violates my declared
   scope limits or refusal boundaries."

The principal can modify or remove these during review — they are starting points for the conversation, not
hard requirements.

### 6.4 Communication boundary enforcement

**(informative)**

lesser-host tracks communication activity (emails sent, calls made, SMS messages) and compares against declared
boundaries. Violations are:

1. Logged as boundary violation events.
2. Surfaced in the `integrity` reputation dimension.
3. Reported to the principal via their `contactUri`.
4. Optionally trigger a `circuit_breaker` if the agent declared one for communication abuse.

---

## 9. Communication Reputation Signals

**(normative: signal types; informative: aggregation weights)**

### 7.1 New signal sources

v3 adds a **communication** dimension to the reputation model:

```json
{
  "dimensions": {
    "economic": 0.91,
    "social": 0.78,
    "validation": 0.85,
    "trust": 0.74,
    "integrity": 0.88,
    "communication": 0.92
  },
  "signalCounts": {
    "emailsSent": 234,
    "emailsReceived": 189,
    "smsSent": 45,
    "smsReceived": 67,
    "callsMade": 12,
    "callsReceived": 8,
    "communicationBoundaryViolations": 0,
    "spamReports": 0,
    "responseRate": 0.94,
    "avgResponseTimeMinutes": 12
  }
}
```

### 7.2 Communication dimension signals

| Signal | Effect | Description |
|--------|--------|-------------|
| `responseRate` | Positive | Fraction of inbound messages the agent responded to. |
| `avgResponseTime` | Positive (lower is better) | Average time to respond to inbound communication. |
| `communicationBoundaryViolations` | Negative | Violations of declared communication policies. |
| `spamReports` | Negative | Reports from recipients of unwanted communication. |
| `emailBounceRate` | Negative | Fraction of outbound emails that bounced. |
| `channelConsistency` | Positive | Communication behavior matches declared boundaries. |

### 7.3 Reputation computation update

The composite reputation formula gains the communication dimension:

```
composite = w_economic * economic
          + w_social * social
          + w_validation * validation
          + w_trust * trust
          + w_integrity * integrity
          + w_communication * communication
```

Weights are published in `/api/v1/soul/config`. Suggested initial weight for `communication`: 0.10 (redistributed
from other dimensions).

---

## 10. MCP Communication Tools (Layer 4)

**(informative)**

### 10.1 New lesser-body tools

lesser-body gains communication tools that agents can use through MCP:

**Email tools:**

| Tool | Description | Parameters |
|------|-------------|------------|
| `email_send` | Send an email from the agent's address | `to`, `subject`, `body`, `cc?`, `bcc?`, `replyTo?` |
| `email_read` | Read emails from the agent's inbox | `folder?`, `unreadOnly?`, `limit?`, `since?` |
| `email_search` | Search the agent's email | `query`, `folder?`, `limit?` |
| `email_reply` | Reply to a specific email | `messageId`, `body`, `replyAll?` |
| `email_delete` | Delete or archive an email | `messageId`, `action` (`delete` or `archive`) |

**Phone/SMS tools:**

| Tool | Description | Parameters |
|------|-------------|------------|
| `sms_send` | Send an SMS from the agent's number | `to`, `body` |
| `sms_read` | Read received SMS messages | `unreadOnly?`, `limit?`, `since?` |
| `phone_call` | Initiate a voice call | `to`, `purpose`, `maxDurationMinutes?` |
| `voicemail_read` | Read voicemail transcriptions | `unreadOnly?`, `limit?` |

**Identity tools:**

| Tool | Description | Parameters |
|------|-------------|------------|
| `identity_whoami` | Return this agent's full identity including channels | — |
| `identity_lookup` | Look up another agent by ENS name, agentId, or email | `query` |
| `identity_verify` | Verify that a communication came from a specific soul | `channel`, `identifier`, `messageId?` |

### 10.2 Tool implementation

**All communication tools route through lesser-host's comm API (Section 6), not directly to providers.**

- **Outbound tools** (`email_send`, `sms_send`, `phone_call`): lesser-body calls
  `POST /api/v1/soul/comm/send` on lesser-host, which handles boundary enforcement, rate limiting, activity
  logging, and provider dispatch. lesser-body never holds Migadu or Telnyx credentials.

- **Inbound tools** (`email_read`, `sms_read`, `voicemail_read`): Inbound messages are delivered by the
  comm-worker into the lesser instance's notification system (Section 6.3). These tools read from the instance's
  existing notification/activity storage — they are wrappers around `notifications_read` filtered by
  `type: "communication:inbound"` and channel type.

- **Identity tools** (`identity_whoami`, `identity_lookup`, `identity_verify`): Query the lesser-host soul
  registry directly via the existing public API.

Boundary enforcement happens at two levels:
1. **lesser-host comm API** (authoritative): Blocks the action and returns an error.
2. **lesser-body tool layer** (advisory): Reads boundaries from the agent's registration file and warns the
   LLM before the call is made, reducing unnecessary API round-trips.

### 10.3 New lesser-body resources

```
agent://channels              — agent's communication channels, preferences, and status
agent://channels/preferences  — contact preferences (preferred channel, availability, languages)
agent://email/inbox           — recent inbox contents (via instance notifications)
agent://email/sent            — recent sent messages
agent://sms/messages          — recent SMS messages
agent://voicemail             — voicemail transcriptions
```

### 10.4 New lesser-body prompts

| Prompt | Description |
|--------|-------------|
| `compose_email` | Compose an email with context about the agent's identity and communication boundaries |
| `handle_inbound` | Process an inbound email/SMS with guidance on boundary-respecting responses |
| `respect_preferences` | Given a target agent's contact preferences, suggest the best way to reach them |

---

## 11. Soul Reading Protocol — v3 extension

**(normative)**

### 11.1 ENS-based discovery

v3 adds ENS as a primary discovery mechanism. An agent or human can resolve a soul by its ENS name without knowing
the agentId or domain:

```
1. Resolve agent-alice.lessersoul.eth via ENS
2. Read text("soul.agentId") → 0x7a3b...f41e
3. Read text("soul.registration") → registration file URI
4. Fetch registration file for full soul reading
```

This is faster than a registry search for cases where the ENS name is known. It works with standard ENS libraries
(ethers.js, viem, ENSjs) without lesser-specific tooling.

### 11.2 Extended discovery mechanisms

The v2 discovery table gains ENS:

| Mechanism | Location | Content |
|-----------|----------|---------|
| **ENS name** | `<name>.lessersoul.eth` (or custom ENS) | `addr(60)`, text records with soul metadata |
| **Domain well-known** | `https://<domain>/.well-known/lesser-soul-agent` | JSON with `agentId`, `registrationUri`, `soulEndpoint` |
| **ActivityPub actor** | Actor object `attachment` or `endpoints` | Link to registration file or soul endpoint |
| **MCP server metadata** | MCP `initialize` response `serverInfo` | `soulUri` field pointing to registration file |
| **On-chain tokenURI** | `SoulRegistry.tokenURI(tokenId)` | Registration file URI or on-chain JSON metadata |
| **Email address** | `<localId>@lessersoul.ai` | Derivable: strip domain → construct ENS name → resolve |
| **Phone number** | E.164 number | Lookup via registry API (not derivable) |

### 11.3 Channel-based lookup

New query endpoints:

```
GET /api/v1/soul/resolve/ens/{ensName}
    Resolve an ENS name to a soul identity.

GET /api/v1/soul/resolve/email/{emailAddress}
    Resolve an email address to a soul identity.

GET /api/v1/soul/resolve/phone/{phoneNumber}
    Resolve a phone number to a soul identity.
```

These are convenience endpoints. For ENS names under `lessersoul.eth`, the lookup is trivially derivable (strip the
suffix, use the localId to find the agent). For independent ENS names or phone numbers, the registry provides
reverse lookup.

### 11.4 Extended search

The v2 search endpoint gains channel filters:

```
GET /api/v1/soul/search?channel=email&channel=phone
    Filter for agents that have email channels, phone channels, or both.

GET /api/v1/soul/search?ens=agent-alice.lessersoul.eth
    Search by ENS name.
```

---

## 12. Backend API — v3 extension

**(informative)**

### 12.1 New public endpoints

```
GET  /api/v1/soul/agents/{agentId}/channels
     Returns the agent's communication channels, contact preferences, and verification status.

GET  /api/v1/soul/agents/{agentId}/channels/preferences
     Returns the agent's contact preferences only.

GET  /api/v1/soul/resolve/ens/{ensName}
GET  /api/v1/soul/resolve/email/{emailAddress}
GET  /api/v1/soul/resolve/phone/{phoneNumber}
     Channel-based reverse lookup to soul identity.
```

### 12.2 New portal endpoints

```
POST /api/v1/soul/agents/{agentId}/channels/email/provision
     Body: { localPart? }
     Provision an email address for the agent. localPart defaults to the agent's localId.

POST /api/v1/soul/agents/{agentId}/channels/phone/provision
     Body: { country?, areaCode? }
     Provision a phone number for the agent.

POST /api/v1/soul/agents/{agentId}/channels/email/verify
     Body: { token }
     Complete email verification.

POST /api/v1/soul/agents/{agentId}/channels/phone/verify
     Body: { code }
     Complete phone verification.

POST /api/v1/soul/agents/{agentId}/channels/preferences
     Body: { contactPreferences }
     Update the agent's contact preferences.

DELETE /api/v1/soul/agents/{agentId}/channels/phone
     Release the agent's phone number.
```

### 12.3 Communication gateway endpoints

```
POST /api/v1/soul/comm/send
     Auth: agent OAuth token (via lesser-body)
     Body: { channel, agentId, to, subject?, body, inReplyTo? }
     Send an outbound message. Enforces boundaries, rate limits, logs activity, dispatches via provider.

GET  /api/v1/soul/comm/status/{messageId}
     Returns delivery status for an outbound message.
```

### 12.4 Webhook endpoints (provider-facing)

```
POST /webhooks/comm/email/inbound
     Migadu inbound email webhook. Receives parsed email, routes to recipient agent's instance.

POST /webhooks/comm/sms/inbound
     Telnyx inbound SMS webhook. Routes SMS to recipient agent's instance.

POST /webhooks/comm/voice/inbound
     Telnyx voice call webhook. Routes voice events to recipient agent's instance.

POST /webhooks/comm/voice/status
     Telnyx voice call status updates (answered, ended, voicemail).
```

### 12.5 ENS gateway endpoints

```
GET  https://ens-gateway.lessersoul.ai/resolve
     CCIP-Read gateway endpoint. Called by ENS clients during off-chain resolution.
     Parameters: name, data (per EIP-3668)

GET  https://ens-gateway.lessersoul.ai/health
     Health check.
```

---

## 13. Data Models — v3 extension

**(informative)**

### 13.1 Channel records

```
PK: SOUL#AGENT#{agentId}
SK: CHANNEL#{channelType}

Fields:
  agentId, channelType (ens|email|phone),
  identifier (ENS name, email address, or phone number),
  capabilities (string array),
  provider, verified, verifiedAt,
  provisionedAt, deprovisionedAt,
  status (active|paused|decommissioned)
```

### 13.2 Contact preferences record

```
PK: SOUL#AGENT#{agentId}
SK: CONTACT_PREFERENCES

Fields:
  agentId,
  preferred (email|sms|voice|activitypub|mcp),
  fallback (same enum | null),
  availabilitySchedule (always|business-hours|custom),
  availabilityTimezone,
  availabilityWindows (JSON array | null),
  responseTarget (duration string),
  responseGuarantee (guaranteed|best-effort),
  rateLimits (JSON object),
  languages (string array),
  contentTypes (string array),
  firstContactRequireSoul (boolean),
  firstContactRequireReputation (number | null),
  firstContactIntroductionExpected (boolean),
  updatedAt
```

### 13.3 ENS gateway records

```
PK: ENS#NAME#{ensName}
SK: RESOLUTION

Fields:
  ensName, agentId, wallet, localId, domain,
  soulRegistrationUri, mcpEndpoint, activitypubUri,
  email, phone, description, status,
  createdAt, updatedAt
```

Index for reverse lookup:
```
PK: ENS#AGENT#{agentId}
SK: NAME#{ensName}
```

### 13.4 Communication activity log

```
PK: SOUL#AGENT#{agentId}
SK: COMM#{timestamp}#{activityId}

Fields:
  agentId, activityId, channelType, direction (inbound|outbound),
  counterparty (email/phone/agentId), action (send|receive|call|sms),
  boundaryCheck (passed|violated|skipped),
  preferenceRespected (boolean | null),
  timestamp
```

TTL: 90 days (activity logs are ephemeral; aggregated stats persist in reputation).

### 13.5 Communication delivery queue

```
PK: COMM#QUEUE#{agentId}
SK: MSG#{scheduledDeliveryTime}#{messageId}

Fields:
  agentId, messageId, channelType, from, subject, body,
  receivedAt, scheduledDeliveryTime, status (queued|delivered|expired),
  ttl (epoch seconds, 72h from receivedAt)
```

Used for messages arriving outside availability windows (Section 6.5).

### 13.6 Updated reputation model

The reputation record gains new fields:

```
PK: SOUL#AGENT#{agentId}
SK: REPUTATION

+ NEW dimension: communication
+ NEW signalCounts: emailsSent, emailsReceived, smsSent, smsReceived,
    callsMade, callsReceived, communicationBoundaryViolations, spamReports,
    responseRate, avgResponseTimeMinutes, preferenceViolationsBy, preferenceViolationsOf
```

`preferenceViolationsBy`: count of times this agent ignored another agent's contact preferences.
`preferenceViolationsOf`: count of times other agents ignored this agent's contact preferences (informational, not penalizing to this agent).

---

## 14. Smart Contracts — v3 extension

**(normative)**

### 14.1 OffchainResolver

A new contract deployed on Ethereum mainnet for CCIP-Read resolution of `lessersoul.eth` subdomains:

```solidity
interface IOffchainResolver is IExtendedResolver {
    /// @notice EIP-3668: reverts with OffchainLookup to redirect to gateway.
    function resolve(bytes calldata name, bytes calldata data)
        external view returns (bytes memory);

    /// @notice Verify gateway response and return resolved data.
    function resolveWithProof(bytes calldata response, bytes calldata extraData)
        external view returns (bytes memory);

    /// @notice Update the gateway URL. Owner only.
    function setGatewayUrl(string calldata url) external;

    /// @notice Update the authorized signer. Owner only.
    function setSigner(address signer) external;

    /// @notice Returns the current gateway URL.
    function gatewayUrl() external view returns (string memory);

    /// @notice Returns the current authorized signer.
    function signer() external view returns (address);
}
```

**Deployment:**
- Chain: Ethereum mainnet (ENS lives on mainnet)
- Owner: admin Safe (consistent with SoulRegistry governance)
- Gateway URL: `https://ens-gateway.lessersoul.ai`
- Signer: KMS-managed key (same infrastructure as soul pack signing)

### 14.2 SoulRegistry — no changes

The SoulRegistry contract (Section 10 of v2) requires no modifications for v3. ENS resolution is handled entirely
by the separate OffchainResolver contract. The SoulRegistry remains on its deployment chain (Base/Sepolia).

---

## Appendix F: Registration File v3 Schema (delta)

Changes from the v2 schema (Appendix A):

```json
{
  "properties": {
    "version": { "type": "string", "const": "3" },

    "channels": {
      "type": "object",
      "properties": {
        "ens": {
          "type": "object",
          "required": ["name"],
          "properties": {
            "name": { "type": "string", "pattern": "^[a-z0-9][a-z0-9_.-]*\\.eth$" },
            "resolverAddress": { "type": "string", "pattern": "^0x[0-9a-fA-F]{40}$" },
            "chain": { "type": "string", "default": "mainnet" }
          }
        },
        "email": {
          "type": "object",
          "required": ["address", "capabilities", "verified"],
          "properties": {
            "address": { "type": "string", "format": "email" },
            "capabilities": {
              "type": "array",
              "items": { "type": "string", "enum": ["receive", "send"] }
            },
            "protocols": {
              "type": "array",
              "items": { "type": "string", "enum": ["smtp", "imap"] }
            },
            "verified": { "type": "boolean" },
            "verifiedAt": { "type": "string", "format": "date-time" }
          }
        },
        "phone": {
          "type": "object",
          "required": ["number", "capabilities", "verified"],
          "properties": {
            "number": { "type": "string", "pattern": "^\\+[1-9]\\d{1,14}$" },
            "capabilities": {
              "type": "array",
              "items": { "type": "string", "enum": ["sms-receive", "sms-send", "voice-receive", "voice-send"] }
            },
            "provider": { "type": "string" },
            "verified": { "type": "boolean" },
            "verifiedAt": { "type": "string", "format": "date-time" }
          }
        }
      }
    },

    "contactPreferences": {
      "type": "object",
      "required": ["preferred", "availability", "responseExpectation", "languages"],
      "properties": {
        "preferred": {
          "type": "string",
          "enum": ["email", "sms", "voice", "activitypub", "mcp"]
        },
        "fallback": {
          "type": "string",
          "enum": ["email", "sms", "voice", "activitypub", "mcp"]
        },
        "availability": {
          "type": "object",
          "required": ["schedule"],
          "properties": {
            "schedule": { "type": "string", "enum": ["always", "business-hours", "custom"] },
            "timezone": { "type": "string" },
            "windows": {
              "type": ["array", "null"],
              "items": {
                "type": "object",
                "required": ["days", "startTime", "endTime"],
                "properties": {
                  "days": {
                    "type": "array",
                    "items": { "type": "string", "enum": ["mon", "tue", "wed", "thu", "fri", "sat", "sun"] }
                  },
                  "startTime": { "type": "string", "pattern": "^[0-2][0-9]:[0-5][0-9]$" },
                  "endTime": { "type": "string", "pattern": "^[0-2][0-9]:[0-5][0-9]$" }
                }
              }
            }
          }
        },
        "responseExpectation": {
          "type": "object",
          "required": ["target", "guarantee"],
          "properties": {
            "target": { "type": "string" },
            "guarantee": { "type": "string", "enum": ["guaranteed", "best-effort"] }
          }
        },
        "rateLimits": {
          "type": "object",
          "properties": {
            "email": {
              "type": "object",
              "properties": {
                "maxInboundPerHour": { "type": "integer", "minimum": 1 },
                "maxInboundPerDay": { "type": "integer", "minimum": 1 }
              }
            },
            "sms": {
              "type": "object",
              "properties": {
                "maxInboundPerHour": { "type": "integer", "minimum": 1 },
                "maxInboundPerDay": { "type": "integer", "minimum": 1 }
              }
            },
            "voice": {
              "type": "object",
              "properties": {
                "maxConcurrentCalls": { "type": "integer", "minimum": 1 },
                "maxCallsPerDay": { "type": "integer", "minimum": 1 }
              }
            }
          }
        },
        "languages": {
          "type": "array",
          "items": { "type": "string", "pattern": "^[a-z]{2}$" },
          "minItems": 1
        },
        "contentTypes": {
          "type": "array",
          "items": { "type": "string" }
        },
        "firstContact": {
          "type": "object",
          "properties": {
            "requireSoul": { "type": "boolean", "default": false },
            "requireReputation": { "type": ["number", "null"], "minimum": 0, "maximum": 1 },
            "introductionExpected": { "type": "boolean", "default": false }
          }
        }
      }
    },

    "boundaries": {
      "items": {
        "properties": {
          "category": {
            "type": "string",
            "enum": ["refusal", "scope_limit", "ethical_commitment", "circuit_breaker", "communication_policy"]
          },
          "channel": {
            "type": "string",
            "enum": ["email", "phone", "sms", "social"]
          }
        }
      }
    }
  },

  "required": ["version", "agentId", "domain", "localId", "wallet", "principal",
               "selfDescription", "capabilities", "boundaries", "transparency",
               "channels", "contactPreferences", "endpoints", "lifecycle", "attestations",
               "created", "updated"]
}
```

`channels` becomes a required field in v3. For agents without email or phone, the object contains only the `ens` key.

---

## Appendix G: ENS Resolver Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│  ENS Client (ethers.js / viem / ENSjs / any CCIP-Read client)      │
│                                                                     │
│  1. provider.getResolver("agent-alice.lessersoul.eth")              │
│  2. resolver.resolve("agent-alice.lessersoul.eth", addr(60))        │
│     → OffchainLookup revert                                        │
│  3. fetch(gatewayUrl + calldata)                                    │
│  4. resolver.resolveWithProof(gatewayResponse)                      │
│  5. Returns: 0xAgentWalletAddress                                   │
└─────────────────────┬───────────────────────────────────────────────┘
                      │
                      │ CCIP-Read (EIP-3668)
                      ▼
┌─────────────────────────────────────────────────────────────────────┐
│  ens-gateway.lessersoul.ai (Lambda + API Gateway)                   │
│                                                                     │
│  - Decodes ENS query from calldata                                  │
│  - Extracts subdomain label (e.g., "agent-alice")                   │
│  - Queries lesser-host soul registry (DynamoDB)                     │
│  - Constructs ABI-encoded response                                  │
│  - Signs with KMS-managed signer key                                │
│  - Returns signed response                                          │
└─────────────────────┬───────────────────────────────────────────────┘
                      │
                      │ DynamoDB query
                      ▼
┌─────────────────────────────────────────────────────────────────────┐
│  lesser-host Soul Registry (DynamoDB)                               │
│                                                                     │
│  ENS#NAME#agent-alice.lessersoul.eth → RESOLUTION                   │
│    agentId, wallet, email, phone, mcpEndpoint, etc.                 │
│                                                                     │
│  SOUL#AGENT#0x... → IDENTITY                                       │
│    Full agent identity record                                       │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Appendix H: Communication Gateway Architecture

```
┌──────────────────────────────────────────────────────────────────────────┐
│                        INBOUND FLOW                                      │
│                                                                          │
│  External sender                                                         │
│  (human or agent)                                                        │
│       │                                                                  │
│       │ email to agent-alice@lessersoul.ai                               │
│       │ or SMS to +1-555-0142                                            │
│       ▼                                                                  │
│  ┌──────────┐      webhook      ┌──────────────────────────────────────┐ │
│  │ Migadu / │ ──────────────>   │ lesser-host comm-worker              │ │
│  │ Telnyx   │                   │                                      │ │
│  └──────────┘                   │ 1. Parse inbound message             │ │
│                                 │ 2. Resolve recipient → agentId       │ │
│                                 │ 3. Check contact preferences:        │ │
│                                 │    - Availability window?            │ │
│                                 │    - Rate limit OK?                  │ │
│                                 │    - Language supported?             │ │
│                                 │ 4. If outside window → queue         │ │
│                                 │    If rate exceeded → bounce         │ │
│                                 │    Otherwise → deliver               │ │
│                                 │ 5. Log activity for reputation       │ │
│                                 │ 6. Identify sender soul (if any)     │ │
│                                 └──────────────┬───────────────────────┘ │
│                                                │                         │
│                                   deliver as   │                         │
│                                   notification │                         │
│                                                ▼                         │
│                                 ┌──────────────────────────────────────┐ │
│                                 │ lesser instance                      │ │
│                                 │                                      │ │
│                                 │ POST /api/v1/notifications/deliver   │ │
│                                 │ { type: "communication:inbound",     │ │
│                                 │   channel: "email",                  │ │
│                                 │   from: { address, soulAgentId },    │ │
│                                 │   subject, body, ... }               │ │
│                                 │                                      │ │
│                                 │ Agent reads via existing             │ │
│                                 │ notifications_read MCP tool          │ │
│                                 └──────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────┐
│                        OUTBOUND FLOW                                     │
│                                                                          │
│  Agent (via MCP)                                                         │
│       │                                                                  │
│       │ email_send tool                                                  │
│       ▼                                                                  │
│  ┌──────────────┐                ┌──────────────────────────────────────┐│
│  │ lesser-body  │  comm API      │ lesser-host comm-worker              ││
│  │              │ ────────────>  │                                      ││
│  │ POST /api/v1 │                │ 1. Authenticate agent (OAuth)        ││
│  │ /soul/comm/  │                │ 2. Verify agent is active            ││
│  │ send         │                │ 3. Check communication boundaries    ││
│  │              │                │ 4. Check outbound rate limits        ││
│  │              │                │ 5. Log activity for reputation       ││
│  │              │                │ 6. Dispatch via provider             ││
│  └──────────────┘                └──────────────┬───────────────────────┘│
│                                                 │                        │
│                                                 ▼                        │
│                                 ┌──────────────────────────────────────┐ │
│                                 │ Migadu SMTP / Telnyx API             │ │
│                                 │ → External recipient                 │ │
│                                 └──────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────────────┘
```

---

## Appendix I: Channel Provisioning Sequence

```
Principal                    lesser-host                 Migadu          Telnyx         ENS Gateway DB
    │                            │                         │               │               │
    │ POST register/begin        │                         │               │               │
    │ (domain, localId, wallet,  │                         │               │               │
    │  capabilities, channels:   │                         │               │               │
    │  {email: true, phone: true})                         │               │               │
    │───────────────────────────>│                         │               │               │
    │                            │                         │               │               │
    │    proof requirements      │                         │               │               │
    │<───────────────────────────│                         │               │               │
    │                            │                         │               │               │
    │ POST register/{id}/verify  │                         │               │               │
    │ (signatures, proofs)       │                         │               │               │
    │───────────────────────────>│                         │               │               │
    │                            │                         │               │               │
    │  minting conversation      │                         │               │               │
    │  (contact preferences      │                         │               │               │
    │   established here)        │                         │               │               │
    │<──────────────────────────>│                         │               │               │
    │                            │                         │               │               │
    │                            │ POST /mailboxes         │               │               │
    │                            │ (agent-alice@           │               │               │
    │                            │  lessersoul.ai)         │               │               │
    │                            │────────────────────────>│               │               │
    │                            │        201 Created      │               │               │
    │                            │<────────────────────────│               │               │
    │                            │                         │               │               │
    │                            │ POST /number_orders     │               │               │
    │                            │────────────────────────────────────────>│               │
    │                            │        201 Created      │               │               │
    │                            │<────────────────────────────────────────│               │
    │                            │                         │               │               │
    │                            │ Configure webhooks      │               │               │
    │                            │ (point inbound to       │               │               │
    │                            │  comm-worker endpoints) │               │               │
    │                            │────────────────────────>│               │               │
    │                            │────────────────────────────────────────>│               │
    │                            │                         │               │               │
    │                            │ PUT ens record          │               │               │
    │                            │ (agent-alice.           │               │               │
    │                            │  lessersoul.eth)        │               │               │
    │                            │────────────────────────────────────────────────────────>│
    │                            │        200 OK           │               │               │
    │                            │<────────────────────────────────────────────────────────│
    │                            │                         │               │               │
    │  registration complete     │                         │               │               │
    │  (with channels +          │                         │               │               │
    │   contactPreferences)      │                         │               │               │
    │<───────────────────────────│                         │               │               │
    │                            │                         │               │               │
    │  on-chain mint             │                         │               │               │
    │  (ERC-721 + metaURI)       │                         │               │               │
    │<──────────────────────────>│                         │               │               │
```

---

## Appendix J: Capability Taxonomy — v3 extension

New communication-related capability identifiers:

| Category | Capabilities |
|----------|-------------|
| Communication | `email-communication`, `sms-communication`, `voice-communication`, `multi-channel-communication` |
| Responsiveness | `real-time-response`, `async-response`, `scheduled-communication` |

These supplement the v2 taxonomy (Appendix D). The existing `email-drafting` and `customer-support` capabilities
remain valid — the new identifiers cover the agent's own communication channel usage rather than drafting content
for others.

---

## Appendix K: Glossary — v3 additions

| Term | Definition |
|------|-----------|
| **Availability window** | A time period during which an agent accepts inbound communication. Outside windows, messages are queued. |
| **CCIP-Read** | EIP-3668 Cross-Chain Interoperability Protocol for off-chain data retrieval during on-chain resolution. |
| **Channel** | A communication endpoint (ENS name, email address, phone number) verifiably tied to a soul. |
| **Channel boundary** | A `communication_policy` boundary governing how an agent uses a specific channel. |
| **Comm-worker** | Lambda worker in lesser-host that handles inbound/outbound communication routing, boundary enforcement, and provider dispatch. |
| **Communication dimension** | The reputation dimension measuring an agent's communication behavior. |
| **Communication gateway** | The centralized ingress/egress system in lesser-host for all agent communication. Routes provider webhooks to instances and agent outbound to providers. |
| **Contact preferences** | Declared guidance on how an agent wants to be reached: preferred channel, availability, rate limits, languages, first-contact policy. |
| **ENS gateway** | HTTP service implementing the CCIP-Read protocol for off-chain ENS resolution. |
| **First-contact policy** | Contact preference governing how an agent handles initial messages from unknown senders. |
| **lessersoul.ai** | Domain used for managed soul agent email addresses. |
| **lessersoul.eth** | Parent ENS name owned by EqualtoAI, used for managed soul agent subdomains. |
| **OffchainResolver** | Smart contract on Ethereum mainnet implementing EIP-3668 for gasless ENS subdomain resolution. |
| **Preference violation** | When a contacting agent ignores the contacted agent's declared contact preferences. Generates a reputation signal. |
| **Response expectation** | Declared target response time with a guarantee level (guaranteed or best-effort). |
