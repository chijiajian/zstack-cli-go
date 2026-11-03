# AI agent execution guide

[English](agents.en.md) | [简体中文](agents.md) | [Documentation index](README.md)

Use this guide when operating ZStack through the CLI. Repository development conventions are in [AGENTS.md](../AGENTS.md). The user's request determines the authorized environment and operations.

## Discover commands

1. Run `zstack-cli version` to identify the installed binary and `COMMAND --help` to confirm its capabilities.
2. Use [commands.json](reference/commands.json) for command discovery. Its paths, categories and help are generated from the current source. Confirm flags against the installed binary before execution.
3. Choose a workflow from the [user manual](usage-en.md) or [examples](../examples/README.en.md). Replace their placeholder names, UUIDs and paths.

Index categories are discovery hints. Inspect actual behavior before acting. For example, `config view` reads local configuration and `config use-context` changes it. Queries with a saved account password may renew a session. `monitor usage` queries capacity and load, while `zql` exposes server-parsed queries and per-VM metrics.

## Execute within the requested scope

- Confirm endpoint and context from the user's request. Pass `--context` on cloud commands or select a task-specific `ZSTACK_CONFIG`.
- Strict read-only work uses session-only credentials or existing AK/SK. Doctor checks configuration and existing access without login or renewal. An expired session requires an authorized authentication step.
- Use `-o json --error-format json`, separate stdout/stderr, and check exit status, response shape, resource identity and detail warnings. Stable sorting helps pagination, while concurrent resource changes can affect comparisons. Record unavailable samples as uncovered.
- Prefer verified UUIDs for single-resource actions. Ambiguous names, missing contexts, unavailable licenses and unexpected dependencies require investigation within the user's scope.
- Execute writes when authorized. Preview when available. IAM/job/recover/resize dry-run is offline, while some older delete/expunge previews query the server. Validate real state separately.
- Use `--yes` for already-authorized targets. Expunge is irreversible. Snapshot revert, task resume, global configuration and infrastructure changes are mutations.
- Use the deadline flag supported by the command. Wait commonly uses `--timeout` and other read commands use `--request-timeout`. Cancellation behavior of older SDK write paths varies.
- After write timeout, cancellation or output failure, query the known UUID to establish the server outcome before considering a retry.

## E2E and recovery

When using a matching source checkout, prefer `scripts/readonly_smoke.py` and `scripts/write_e2e.py` for their supported regression workflows. The write runner journals generated UUIDs/names before creation. Use a new journal for each run and one process per journal.

On failure preserve the journal, wait for asynchronous API work to settle and run `--cleanup-only --execute` against the same endpoint. Clean only resources whose ownership matches the journal. Root-volume leftovers, identity mismatches and unknown creation outcomes remain unresolved. Keep credentials and sensitive raw responses out of reports.

Report completed operations, local checks, actual cloud validation, skips, cleanup results and remaining gaps separately. A historical passing record or successful help lookup cannot establish current behavior.

## Installed Skill and MCP

The [zstack-operations Skill](../skills/README.en.md) is maintained alongside the CLI and installs independently. It guides the agent to invoke the CLI for queries, monitoring, lifecycle, IAM and recovery. All Skill references are contained inside its installed directory.

Use AGENTS.md for repository development and the installed Skill for operations. Add an MCP adapter when a client's tool interface, remote execution or centralized control requires one.
