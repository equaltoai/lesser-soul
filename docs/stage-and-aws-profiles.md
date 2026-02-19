# Stage + domain configuration contract (M0.2)

This repo uses a **two-stage model**:
- `lab`
- `live`

The stage is selected at deploy time and is surfaced to runtime code via `SOUL_STAGE`.

## Single source-of-truth: stage → instance domain

Stage configuration lives in `infra/cdk/cdk.json` under `context`, keyed by stage name:
- `context.lab.instanceDomain`
- `context.live.instanceDomain`

The CDK app reads:
1) `-c stage=<lab|live>`
2) the matching stage config from `cdk.json` (context key `lab` / `live`)

## SSM path conventions

All instance-scoped SSM parameters are namespaced by instance domain:

`/soul/<instance-domain>/...`

Common keys (see `SPEC.md` for the full list):
- `/soul/<instance-domain>/inference/url` (String)
- `/soul/<instance-domain>/inference/key` (SecureString)
- `/soul/<instance-domain>/lesser-host/instance-key` (SecureString)
- `/soul/<instance-domain>/agents/<type>/token` (SecureString)
- `/soul/<instance-domain>/agents/<type>/refresh` (SecureString)

Go code should use `pkg/config` helpers (e.g., `config.SSMPath`, `config.InferenceURLSSMPath`) to avoid path drift.

## AWS profiles: when to use `Sim` vs `Lesser`

This milestone is **config only**, but the profile split matters:

- **Instance account (Simulacrum / Lesser instance)**: `AWS_PROFILE=Sim`
  - `make cdk-deploy STAGE=lab`
  - `make cdk-destroy STAGE=lab`
  - `theory app up --stage lab --aws-profile Sim`
- **Central account (`lesser-host`)**: `AWS_PROFILE=Lesser`
  - Use this profile only for work that touches central `lesser-host` resources (generally in the `lesser-host` repo, not here).

Notes:
- `make cdk-synth STAGE=lab` does **not** require AWS credentials (synth only).
- CDK deploy/destroy commands will fail (or deploy to the wrong account) if the profile is wrong — always set `AWS_PROFILE` explicitly.

