# Working on zstack-cli-go

## Repository workflow

- Read `docs/usage-zh.md` for user-facing behavior and `docs/agents.md` for operating the CLI against a cloud.
- Preserve existing uncommitted work. Do not commit, push, release or change a real environment unless the user authorizes it.
- Keep command naming, output and confirmation behavior consistent with adjacent commands. Use shared client, resolver, configuration and output abstractions; avoid duplicating transport or locking logic.
- For code changes, add regression coverage for changed behavior and run `make verify` (vet, uncached race tests, Python runner tests, build). These checks require no cloud credentials.
- For command/flag changes, run `make build`, then `python3 scripts/generate_cli_docs.py --cli ./zstack-cli`; review updates to `docs/reference/commands.md` and `commands.json`. Update the usage guide and relevant examples.
- Documentation-only changes need link, command-help and generation checks, not a real cloud run.
- Never read or use `.env` as implicit authorization to access a cloud. Never put credentials, session files or raw sensitive server responses in docs, examples or test fixtures.
- Real environment scripts are opt-in and separate from CI. `make verify` passing does not prove a real environment was tested; report exactly what ran and what was skipped.

## Documentation sources

- `docs/reference/commands.json`: generated full command help for machine discovery; not an authorization policy.
- `skills/zstack-operations/`: independently installable CLI operations skill; keep references inside its directory and update capability checks when relevant flags change. Validate with `make skill-check` and package with `make skill-package`.
- `examples/`: workflow examples and input data; write examples must be adapted to user-authorized test resources.
- `docs/next-stage-checklist.md`: historical validation evidence and limitations, not instructions to repeat those writes.
