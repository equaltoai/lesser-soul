# FEP Codeberg Submission Notes

Date: 2026-03-18
Issue: `equaltoai/lesser-soul#6`

## Prepared artifacts

- FEP title: `Agent Social Attribution for ActivityPub`
- Computed slug: `c81b`
- Local submission repo: `../codeberg-fep`
- Local submission branch: `fep-c81b-agent-social-attribution`
- Proposal path in the Codeberg repo layout: `fep/c81b/fep-c81b.md`

## Local source of truth

- Maintained draft in `lesser`: `../../lesser/docs/specs/fep-agent-attribution.md`
- Codeberg-ready copy in workspace clone: `../../codeberg-fep/fep/c81b/fep-c81b.md`

## Notes

- The proposal text is CC0 as required by the FEP process.
- Making `lesser-soul` public under AGPL-3.0 later does not change the license on the FEP text submitted to Codeberg.
- The current `discussionsTo` value assumes a fork at `https://codeberg.org/equaltoai/fep/issues`.

## Push sequence

Once the fork exists on Codeberg and SSH auth is working, use:

```bash
cd /home/aron/ai-workspace/codebases/equaltoai/codeberg-fep
git remote rename origin upstream
git remote add origin git@codeberg.org:equaltoai/fep.git
git push -u origin fep-c81b-agent-social-attribution
```

Then open a PR from `equaltoai/fep:fep-c81b-agent-social-attribution` to `fediverse/fep:main`.

## Remaining external actions

- Create the Codeberg fork `equaltoai/fep` if it does not already exist.
- Confirm the discussion location you want to use for `discussionsTo`.
- Open the PR and capture the resulting PR URL in the issue thread.
- After facilitators register the proposal, record any assigned tracking issue URL back into the local docs if desired.
