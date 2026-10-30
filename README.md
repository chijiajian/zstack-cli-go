
# zstack-cli

`zstack-cli` is a command-line interface for managing ZStack resources go version. It allows you to create, delete, expunge, and query ZStack resources efficiently.

---

## Features

- Query and describe compute, storage, networking, IAM and operational resources.
- Create/update/delete resources, recover instances/volumes/projects, and grow volumes.
- Wait for resource/job states; explicitly resume or clean eligible long jobs.
- Manage multiple contexts, diagnose access with `doctor`, and use structured errors.
- Run isolated read-only checks and opt-in write E2E with ownership-aware cleanup.
- Shell completion for Bash, Zsh, Fish and PowerShell.

## Documentation

- [中文使用指南](docs/usage-zh.md): login, contexts, queries, IAM, lifecycle actions, errors and regression cleanup.
- [Complete command reference](docs/reference/commands.md) and [machine-readable index](docs/reference/commands.json)
- [Workflow examples](examples/README.md)
- [Installable ZStack operations Skill](skills/README.md) — agent-to-CLI workflows, no MCP required.
- [AI agent execution guide](docs/agents.md) and [repository instructions](AGENTS.md)
- [Query and pagination reference](#query-output-and-pagination)
- [Error codes and doctor](#automation-errors-and-read-only-diagnostics)
- [Validation evidence and known coverage limits](docs/next-stage-checklist.md)

These features describe the current source tree. Check `zstack-cli version` and
subcommand `--help` when using an older release. Build the current CLI with `make build`.

---

## Installation

### Linux / macOS

```bash
# Download precompiled binary from GitHub releases
curl -LO https://github.com/zstackio/zstack-cli/releases/download/v<version>/zstack-cli_<version>_linux_amd64.tar.gz
tar -xzf zstack-cli_<version>_linux_amd64.tar.gz
sudo mv zstack-cli /usr/local/bin/

# Enable bash completion
sudo mkdir -p /etc/bash_completion.d/
sudo zstack-cli completion bash > /etc/bash_completion.d/zstack-cli
```

or execute install.sh
```bash
# Make the script executable
chmod +x install.sh
# Run the installation
sudo ./install.sh
```

### Windows

- Download the .zip from GitHub releases

- Extract and add the executable to your PATH

- Optionally, enable PowerShell completion:
```
zstack-cli completion powershell > zstack-cli.ps1
. .\zstack-cli.ps1
```

## Usage
### Authenticate with username/password
```bash
zstack-cli login 192.168.1.100 --username admin --password password
```

### Authenticate with AK/SK
```bash
zstack-cli login 192.168.1.100 --access-key-id <ak> --access-key-secret <sk>
```

### Show CLI version
`zstack-cli version`

### List all images
`zstack-cli get images`

### Create a disk offering
`zstack-cli create disk-offering my-disk-offering --size 100G`

### Delete an instance
`zstack-cli delete instance <instance-name-or-uuid>`

### Expunge an image
`zstack-cli expunge image <image-name-or-uuid>`

## Resource selection and command failures

Resource identifiers use UUID first, then an exact name. Partial names are not
matched automatically. Single-resource operations reject duplicate names; use a
UUID to disambiguate. Existing batch VM/image commands may select resources with
the same exact name and show the targets before confirmation.

`delete instance --dry-run` and `expunge instance --dry-run` query and display the
targets without prompting or sending write requests. API errors and partial batch
failures return a nonzero exit code; diagnostics are written to stderr.

Login endpoints accept a hostname/IP or `host:port` (default port: 8080).
`login --save-password=false` saves the session without the password. Subsequent
commands reuse a valid session; an expired session requires another login if no
password was saved.

`wait instance <name-or-uuid> --for Running --timeout 2m` applies its deadline to
authentication, name resolution, HTTP requests, and polling intervals.

## Output Formats

Query commands support the --output (-o) flag; other commands may have their own output contract. Check subcommand --help:
```
zstack-cli get instances -o table
zstack-cli get instances -o json
zstack-cli get instances -o yaml
zstack-cli get instances -o text
```

## Command Completion

### Bash
`source <(zstack-cli completion bash)`

### Zsh
`source <(zstack-cli completion zsh)`

### Fish
`zstack-cli completion fish | source`

### PowerShell
`zstack-cli completion powershell | Out-String | Invoke-Expression`

## Environment Variables
- ZSTACK_CONFIG: Path to the CLI configuration file. Defaults to ```~/.zstack-cli/config.yaml.```
- ZSTACK_ACCESS_KEY_ID: Access key ID used by `zstack-cli login <endpoint>` when AK/SK flags are omitted.
- ZSTACK_ACCESS_KEY_SECRET: Access key secret used by `zstack-cli login <endpoint>` when AK/SK flags are omitted.

## Contributing
1.Fork the repository

2.Create a feature branch: git checkout -b feature/your-feature

3.Make your changes

4.Commit your changes: git commit -m "feat: description"

5.Push to the branch: git push origin feature/your-feature

6.Open a Pull Request

7.Please follow the coding style and add tests where applicable.

## License
```
Apache License 2.0
Copyright © 2025 zstack.io
```

## Network queries and resource details

The CLI uses `zstack-sdk-go-v2 v0.0.10`. Network queries follow the existing
`get <plural-resource> [exact-name]` convention:

```bash
zstack-cli get security-groups
zstack-cli get security-group-rules --security-group web
zstack-cli get port-forwarding-rules
zstack-cli get load-balancers
zstack-cli get load-balancer-listeners --load-balancer public
zstack-cli get load-balancer-server-groups --load-balancer public
zstack-cli get volumes  # alias for get disks
zstack-cli describe instance web-01 -o json
zstack-cli describe volume data-01 -o yaml
```

`describe instance` includes the full inventory and related host, image, instance
offering, and L3 networks. `describe volume` includes the full inventory and
related instance, primary storage, and disk offering. A failed related-resource
lookup preserves the main detail and adds a warning to stderr and the structured
`warnings` field. Failure to read the main resource exits nonzero. The default
detail output is structured YAML; JSON is available with `-o json`.

## Query output and pagination

```bash
zstack-cli get instances --all-pages --limit 100 -o json
zstack-cli get volumes --count
zstack-cli get hosts --reply-with-count --fields uuid,name -o json
zstack-cli get security-groups --request-timeout 10s -o yaml
```

- `--all-pages` fetches every page from `--start`; `--limit` is the page size.
  It preserves filters and sorting, defaults to UUID order (ID order for global configs, events, and audits), and emits no partial
  result if a later page fails. Concurrent server-side changes can still affect
  pagination; this is not a snapshot.
- `--count` prints only the total number, including in JSON/YAML mode.
- Normal JSON/YAML output is a list; an empty result is `[]`.
- `--reply-with-count` (or legacy `--pagination`) wraps JSON/YAML results as
  `{"items": [...], "total": N}`. In table mode the total goes to stderr.
- `--fields` projects the displayed fields, including JSON/YAML item fields.
- `--request-timeout` defaults to 30 seconds and bounds authentication plus the
  query's page retrieval. Resource filters are resolved under separate bounded
  requests. `describe` shares one deadline for resolution, detail, and relations.

The SDK still drops HTTP request contexts in v0.0.10. Queries, detail reads, and
wait operations therefore share a cancellable HTTP read adapter; SDK types and
resource paths remain the API contract.

## Waiting for resources and jobs

```bash
zstack-cli wait instance web-01 --for Running --timeout 5m --interval 2s
zstack-cli wait job <name-or-uuid> --timeout 10m --progress -o json
zstack-cli wait long-job <uuid> --for Canceled --interval 1s -o yaml
zstack-cli get jobs -o json --fields uuid,state,jobResult
zstack-cli describe job <uuid> --request-timeout 10s -o json
```

`wait job` / `wait long-job` defaults to `--for Succeeded`; instance, image, and
volume waits still require `--for`. States are compared case-insensitively.
After resolving a name, polling uses the resolved UUID. An already satisfied
state returns immediately. The timeout includes authentication, resolution,
HTTP reads, and intervals between polls.

A job reaching `Failed` or `Canceled` while waiting for success exits nonzero
immediately and includes its `jobResult` in the error. Any terminal state that
differs from the requested target is an error. Explicitly waiting for
`--for Failed` succeeds when that state is reached: exit status describes the
wait condition. `Suspended` and `Canceling` continue waiting until the target,
a terminal state, or timeout.

`--progress` writes observations to stderr only when they change. Text remains
the default success output; `-o json` and `-o yaml` emit one final observation
with resource identity, state/status, and jobResult when available. Errors leave
stdout empty. Timeout and polling errors include the last observed state; errors
before initial resolution cannot report a state. Job results remain the raw
SDK string, preserving nested server error details.

`get jobs` includes jobResult; `describe job` retains the complete SDK inventory
and bounds authentication and detail retrieval with `--request-timeout` (30s by
default). Task failure is data for get/describe and does not make those reads fail.

## Recovering resources and growing volumes

```bash
zstack-cli instance recover <name-or-uuid>
zstack-cli volume recover <name-or-uuid> -o json
zstack-cli volume resize <name-or-uuid> --size 40G --dry-run
zstack-cli volume resize <name-or-uuid> --size 40G -o json
```

Recovery requires a `Destroyed` instance or a `Deleted` data volume. Root volumes
are recovered through their instance. Expunged resources are no longer available
for recovery. The existing `image recover` command remains available.

`--size` is the new total capacity, not an increment. It accepts integer bytes or
K/M/G/T (optionally followed by B), using powers of 1024 as in `create volume`.
Zero, negative, fractional, and overflowing sizes are rejected. A volume must be
`Ready`, and the requested capacity must exceed its current size. The CLI reads
the volume type and selects the SDK data-volume or root-volume resize action;
unknown types and shrinking are rejected before mutation. Backend constraints
still apply. Growing a virtual disk does not grow the guest partition/filesystem.

Both actions use exact UUID/name resolution and reject ambiguous names. Dry-run
is offline: it prints the action, supplied identifier, and byte size without
checking live resource state or choosing a volume-type-specific endpoint. Normal
output follows the existing resource commands; `-o name` prints `kind/uuid`, and
JSON/YAML output contains the returned inventory. API failures exit nonzero.

## Multiple contexts

```bash
zstack-cli login 192.168.1.100 --ak AK --sk SK --context-name production
zstack-cli login 192.168.1.101 --username admin --password PASSWORD --context-name staging --no-switch
zstack-cli config get-contexts
zstack-cli config current-context
zstack-cli config use-context staging
zstack-cli get instances --context production -o json
zstack-cli config rename-context staging test
zstack-cli config use-context production
zstack-cli config delete-context test
```

Login defaults to naming the context after the endpoint and making it current,
so existing login commands continue to work. `--context-name` permits multiple
accounts on the same endpoint. Logging in to an existing name replaces its
credentials. `--no-switch` saves credentials while preserving the default.

Global `--context` selects an existing context for one command and takes precedence
over the saved default. It does not change `current-context`, including when a
session is refreshed. Unknown names fail instead of falling back to another
context. For login use `--context-name`; combining login with `--context` is an
error. Configuration mutations still explicitly change the saved configuration.

`config get-contexts` (alias `contexts`) lists names in sorted order, with saved
`current` and effective `selected` indicators; JSON/YAML output is supported.
`config current-context` prints the effective name. Renaming the saved current
context updates the default reference. Rename collisions and unknown names fail.
To delete the saved current context, switch to another one first; deletion only
removes local credentials, without calling a server logout API. In the example,
switch away from `test` before deleting it if it is still current.

`config view` redacts passwords, access keys, and session tokens in every output
format. Configuration files are replaced atomically with mode 0600. The existing
`ZSTACK_CONFIG` file override remains supported. All CLI read-modify-write operations acquire a cross-process file lock before
loading the latest configuration, then atomically replace the file. Concurrent
CLI writers therefore preserve unrelated changes.

Configuration mutation commands also support JSON/YAML output, and invalid output
formats are rejected before saving. Session refresh re-reads the configuration
before saving to preserve changes made during the network login; if the selected
credentials were replaced or deleted meanwhile, the command fails and asks for a
retry. The credential comparison and save execute in the same locked transaction.

`describe job --fields` and recover/resize structured output project SDK fields
(including embedded fields such as UUID). Recover/resize validate requested fields
before contacting the server. Query responses missing inventories or containing
negative totals are treated as errors rather than successful empty results.

Configuration transactions use an OS-managed sidecar lock (`config.yaml.lock`)
on Linux/macOS and Windows. Lock waits honor the caller's deadline and are capped
at 10 seconds; failure to acquire the lock exits nonzero without saving changes.
The sidecar stays in place and must not be removed while commands are running.
The OS releases a held lock if a process exits or is killed.

Network authentication runs outside the lock. When two commands refresh the same
unchanged credentials, the later transaction reuses the session already committed.
A concurrent rename, deletion, or credential replacement prevents stale refresh
results from being written back. Context switching and unrelated context edits
remain intact. These guarantees require cooperating CLI writers on a filesystem
that supports OS file locking; external editors do not participate in the lock.

A command pins its selected context name and environment identity at startup.
Switching the saved default in another process affects the next command, not
later stages of an in-flight command. If the selected context's endpoint or
account credentials are replaced between stages, the command fails and requires
a retry; rotating its session token remains allowed.

Existing configuration symlinks use the target's lock and preserve the link on
save. A dangling configuration symlink is rejected instead of being overwritten.
An already-canceled transaction does not create configuration directories or lock
files. Session validity responses must contain an explicit boolean `valid`;
malformed responses fail without triggering a new login or modifying credentials.

Query field names are validated for table/text as well as JSON/YAML, including
empty result sets. Unknown `--fields` values exit nonzero without stdout output.
An explicit `--sort` overrides the resource's default pagination order.
`vm-script-records [record-name]` filters by `recordName`; CD-ROM queries accept
conditions through `-q` rather than a positional name.

Text output with `--fields` prints only the selected named fields, rather than
zero-filled structs. Queries discard results if cancellation is observed after
a page/count response, including the final page. The read adapter rejects trailing
JSON documents or garbage and supports SDK timestamps in typed maps, pointers,
and case-insensitive field names while preserving untyped values and integers.

### Repeatable read-only smoke checks

After building the CLI, use an already authenticated configuration to run the
bounded regression suite (Python 3 standard library only):

```sh
python3 scripts/readonly_smoke.py --cli ./zstack-cli \
  --config /path/to/dedicated-config.yaml > smoke-report.json
make test-smoke
```

The suite covers 42 canonical query commands, field projection in four formats, filtered
forward/reverse pagination and offsets/counts, and instance/volume/job identity
checks. It samples at most three records per resource; empty collections and unavailable
optional relations are reported as skipped dependent checks. Global configurations
expose a numeric id and participate in pagination checks.
`--sample-size 1..3` and `--timeout 1..120` control sample size and each CLI deadline.
`--suite-timeout 1..3600` bounds the complete CLI run (default: 600 seconds).
Command timeouts terminate and reap the child; SIGINT/SIGTERM also clean the
temporary config and return nonzero. SIGKILL cannot run cleanup handlers.
Concurrent resource deletion can cause comparison failures; rerun against stable
samples to distinguish those from regressions.

An explicit config is required and copied into a private temporary directory,
which is removed when the suite finishes. The source config and default context
are not changed. A local `config view -o json` preflight requires a session created
with `--save-password=false` or complete access keys. Saved account passwords are
rejected before network queries, preventing implicit renewal from creating sessions
that the script does not own. The
suite does not log out an existing user session or create access keys. Use a
separate session and log it out after verification when session isolation is needed.
After the local configuration preflight, only `get` and `describe` are allowed; resource mutations and lifecycle
operations are excluded. Reports contain check names/statuses, not inventories,
credentials or raw server error bodies. Failed checks exit nonzero; skipped checks
are counted separately and do not establish coverage for unavailable resources.

Pagination is checked against the original sampled identities, including exact
ordering, duplicates, offsets and integer counts. Field checks pin a sample and
validate the JSON/YAML/text/table projection, rather than accepting any nonempty
output. Optional relation permission/not-found warnings are reported as skipped relation
coverage; they do not fail an otherwise valid main-resource identity check. Other
relation failures (including malformed identities) still fail the check.

### Local and CI verification

```sh
make verify      # vet, uncached race tests, Python smoke-runner tests, build
make test        # uncached Go tests without the race detector
make test-smoke  # Python runner tests only; no cloud connection
```

`GO`, `PYTHON`, and `TEST_TIMEOUT` may be overridden, for example
`make verify PYTHON=python3 TEST_TIMEOUT=10m`. Race tests require a working C
compiler. The default Go test timeout is five minutes per test binary.

The CI workflow runs the same `make verify` target on Linux and macOS for pushes,
pull requests and manual dispatch. Linux also cross-compiles the Windows CLI and
configuration-lock tests; this is a compilation check, not Windows runtime coverage.
Tag releases invoke the reusable CI workflow before importing signing credentials
or running GoReleaser. The regression jobs use read-only repository permissions
and do not receive cloud credentials or invoke the real-environment smoke suite.

### Session logout and final operation consistency

```sh
zstack-cli logout --context lab --request-timeout 30s
zstack-cli logout --context lab --local
zstack-cli primary-storage add-local local-ps --zone zone-1 --url /data/zstack \
  --dry-run -o json
zstack-cli snapshot revert snapshot-1 --dry-run -o json
```

`logout` revokes the selected saved account session without validating or renewing
it first, then clears that context's local credentials. Network errors retain the
credentials; a concurrent context/session replacement is never overwritten. A
missing remote session is already logged out. `--local` only clears credentials;
access keys are always cleared locally and are never revoked by this command.
The selected context remains available with its endpoint and username, and the
persisted default context is unchanged.

`primary-storage add-local` accepts an absolute host directory, zone, description,
resource UUID and tags. SDK v0.0.10's empty-body convenience method is bypassed using
its typed request and shared `Post` transport. Creation is verified by mock tests;
no real storage is created by the regression suite.

Snapshot revert now requires confirmation or `--yes`. Legacy image/instance
`delete` and `expunge` also accept `-y/--yes`; previews still bypass confirmation.
Confirmation I/O failure never authorizes a mutation. Creation result formatters
propagate encoding and output failures, and unsupported output formats are rejected
before API requests. If an API operation succeeds but output fails, the CLI returns
nonzero; inspect resource state before retrying a write.

Explicit account login uses the cancellable HTTP adapter to avoid SDK logs that
include authentication request bodies. Login failures do not print password hashes
or server echoes of credentials; successful responses must contain a session UUID.

Account login and automatic renewal omit unset SDK optional authentication fields.
Explicit login reads session identity fields only, so localized server display dates
do not invalidate a successful login.

### Automation errors and read-only diagnostics

Use `--error-format json` to receive one structured diagnostic on **stderr**;
resource output stays on stdout. Diagnostics contain `error.code`, `error.message`
and `error.exitCode`, with `httpStatus` and a validated `apiCode` when available.
Raw SDK request bodies and server echoes are omitted from structured diagnostics.
Unclassified failures retain exit code 1; do not assume a failed write was rolled back.

| Exit | Code | Meaning |
| --- | --- | --- |
| 0 | — | Success |
| 1 | operation_failed | Other operation or partial batch failure |
| 2 | invalid_arguments | Invalid arguments, flags, query options, or validated fields |
| 3 | authentication_failed | Missing/invalid credentials or permission denied |
| 4 | not_found | Resource/context missing or HTTP 404 |
| 5 | conflict | Conflicting state or ambiguous resource selection |
| 6 | timeout | Request/deadline timeout |
| 7 | connection_failed | Network connection failure |
| 8 | api_failed | API failure or invalid mutation response |
| 9 | output_failed | Output/encoding failure |
| 130 | canceled | Canceled operation |

```sh
zstack-cli get instances -o json --error-format json
zstack-cli doctor --context lab -o json --request-timeout 15s
zstack-cli doctor --context lab --offline -o json
```

`doctor` validates endpoint syntax and saved credentials. Online checks validate
an existing account session and perform a one-record zone query (or sign that query
with saved access keys). It never logs in, renews sessions, or writes configuration.
An expired saved session fails even if a password is saved. Offline mode sends no
requests. A successful zone query checks basic response compatibility and that
permission only; it does not certify every API or licensed module.

### IAM and task lifecycle

The CLI now supports `create`, `update`, and `delete` for `account`, `user`, `role`,
`project`, and `organization`; `policy` supports create/delete because SDK v0.0.10
does not expose policy updates. Users belong to the account selected by the context.
IAM commands accept `--dry-run` and `--request-timeout`; deletes require confirmation
or `--yes`. Names and UUIDs resolve through the shared resolver.

```sh
# The password file contains plaintext; previews redact its value and digest.
zstack-cli create account test-account --password-file ./account-password --dry-run -o json
zstack-cli create user test-user --password-file ./user-password
zstack-cli create policy query-only --statements-file ./statements.json
zstack-cli create role operator --policy query-only
zstack-cli create organization test-department --type Department
zstack-cli create project test-project --organization test-department
zstack-cli update project test-project --description 'Test project'
zstack-cli delete project test-project --yes
zstack-cli project recover test-project --yes
zstack-cli delete project test-project --yes
zstack-cli expunge project test-project --yes
zstack-cli job resume suspended-job --yes
zstack-cli job clean finished-job --yes
```

`--statements-file` is a JSON array of SDK policy statements with explicit `Allow`
or `Deny` effects and nonempty actions. IAM creation uses flags rather than the
parent `create --file` option. Explicit empty descriptions clear descriptions.
Passwords are SHA-512 hashed before sending and never included in previews/results.
Role policy references, project organizations/accounts, and parent organizations
are resolved before writes. IAM2 features require the corresponding server license.

Job resume requires `Suspended`; clean accepts `Succeeded`, `Failed`, or `Canceled`.
Both actions require confirmation or `--yes`. IAM/job writes share a cancellable
transport that submits each mutation once and polls only its returned API job;
callbacks must stay on the configured origin and API-job path. Account update uses
the [documented account endpoint](https://www.zstack.io/help/dev_manual/dev_guide/v4/10.1.4.html),
which differs from the generated SDK v0.0.10 convenience method.

### Complete query coverage and opt-in write E2E

The read-only smoke suite covers 42 canonical query commands (aliases share coverage).
Global configurations now expose numeric `id` for stable sorting/projection/pagination.
Built-in alarms and tags may have opaque alphanumeric identifiers, which the suite
checks without treating them as hexadecimal UUIDs. Empty resource collections and
unavailable optional relations are explicitly reported as skipped dependent checks.

Real write tests are **never** part of `make verify` or CI. They require `--execute`
and a session-only/AK config. Use an isolated test environment with adequate quota:

```sh
python3 scripts/write_e2e.py --cli ./zstack-cli --config ./session.yaml \
  --journal ./metadata-run.json --execute

python3 scripts/write_e2e.py --cli ./zstack-cli --config ./session.yaml \
  --journal ./compute-run.json --execute --profile compute \
  --image ROOT_IMAGE_UUID --l3-network L3_UUID --host HOST_UUID

# Recover cleanup after interruption; never reuse this journal to create resources.
python3 scripts/write_e2e.py --cli ./zstack-cli --config ./session.yaml \
  --journal ./compute-run.json --execute --cleanup-only
```

Metadata tests cover volume/account/user/policy/role/organization/project lifecycle,
including project recovery. `--resources` explicitly selects a metadata subset;
the report records that selection. The compute profile creates a stopped 1-CPU,
1-GiB instance and a 1-GiB data volume, verifies attachment, snapshot creation/revert/
deletion, growth to 2 GiB, detachment, and volume/instance recovery. Cleanup checks
that the VM's root volume is also gone. Existing images, networks, hosts and storage
are references only; existing workloads are not mutated.

The journal records endpoint, generated names and UUIDs **before** creation. Cleanup
checks ownership, runs in reverse order, detaches volumes only from this run's own
instance, and has independent per-command deadlines. Unknown outcomes remain in the
journal instead of being reported as clean. The default suite deadline is 1200s and
command deadline 180s; cleanup still runs after the suite deadline. Preserve journals
until every resource is marked `cleaned`. The runner copies configuration with mode
0600, rejects saved account passwords, removes temporary credentials/files, and never
prints raw command errors or credentials. It does not revoke the supplied session.
