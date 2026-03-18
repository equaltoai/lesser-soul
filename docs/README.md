# Docs and Deployment Scope

Maintainer documentation for the soul registry still lives primarily in `lesser-host`, but `lesser-soul` now also hosts
the public `spec.lessersoul.ai` site and the FEP namespace/deployment work for agent social attribution.

Registry implementation references:

- ADRs: `lesser-host/docs/adr/`
- Agent ID test vectors: `lesser-host/docs/spec/agent-id-test-vectors.md`
- Soul surface + instance integration (incl. lesser-body MCP): `lesser-host/docs/soul-surface.md`

This repo now contains:

- specifications and roadmaps for Lesser Soul and related FEP work
- CDK deployment for `spec.lessersoul.ai`
- static namespace assets such as `/ns/agent-attribution/v1`

Roadmaps and plans:

- v2 implementation roadmap: `../ROADMAP.md`
- v3 implementation roadmap (stack-wide): `../ROADMAP-v3.md`
- issue 3 FEP implementation plan: `../roadmaps/issue-3-fep-agent-attribution.md`

Deployment:

- site and namespace infrastructure: `../cdk/`
