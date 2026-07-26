# lesser-soul

Public specifications, namespace assets, and deployment infrastructure for Lesser Soul.

## Overview

`lesser-soul` contains the public-facing materials and infrastructure for the Lesser Soul project, including:

- specifications and implementation roadmaps
- the static `spec.lessersoul.ai` site
- namespace assets such as `/ns/agent-attribution/v1`
- versioned static contracts such as `/contracts/panonomous/soul-document/v1/schema.json`
- deployment inventory and supporting maintainer docs

Implementation details for the broader soul registry and control-plane work still primarily live in related repositories such
as `lesser-host`.

## Repository Layout

- `cdk/`: AWS CDK app for the public site and namespace hosting
  - operator deployment flow and commands: `cdk/README.md`
- `docs/`: maintainer docs and deployment notes
- `roadmaps/`: implementation plans for active work
- `contracts/`: contract and schema material, including the Panonomous soul-document schema and agent naming vocabulary
- `SPEC.md`: canonical product and protocol specification
- `ROADMAP.md`: milestone plan for the repo

## License

lesser-soul is licensed under the GNU Affero General Public License v3.0. See [LICENSE](LICENSE) for details.
