# ZStack CLI user manual

[English](usage-en.md) | [简体中文](usage-zh.md) | [Documentation index](README.md)

This manual describes the current source tree, using SDK v0.0.10. Check `zstack-cli version` and subcommand `--help` for installed capabilities. Replace example names, UUIDs, addresses and paths with your environment's values.

See the [command reference](reference/commands.md) for flags and defaults, [examples](../examples/README.en.md) for workflow inputs, and the [agent guide](agents.en.md) for agent execution.

## 1. Build and authenticate

Release binaries run directly after extraction. See the [README installation steps](../README.md#run-directly-after-extraction) for exact platform filenames, PowerShell commands and PATH setup.

Run `make build` in the repository root to create `./zstack-cli`. Examples assume it is on PATH. Building requires Go 1.25 or compatible. On Windows use `go build -o zstack-cli.exe .` and `./zstack-cli.exe` in PowerShell when running from the current directory.

```sh
zstack-cli version
zstack-cli --help
zstack-cli login 192.0.2.10:8080 --username admin --context-name lab --save-password=false
```

Login prompts for a password. The endpoint is `host` or `host:port`, with 8080 as the default API port. `--save-password=false` stores the session and requires a new login when it expires. Saving passwords is enabled by default and allows later commands to renew expired sessions.

Authenticate with an existing AK/SK pair as follows.

```sh
zstack-cli login 192.0.2.10:8080 --ak "$LAB_AK" --sk "$LAB_SK" --context-name lab
```

Both AK/SK request paths use the same Date and signing logic, defaulting to the CLI host's local time zone. For servers with GMT/CST date parsing compatibility issues, set `ZSTACK_API_TIMEZONE` to the IANA zone required by that environment (for example `Asia/Shanghai`, `Asia/Singapore`, `Asia/Kuala_Lumpur`, `Europe/Berlin`, or `Asia/Dubai`). `UTC` produces the standard GMT header. The time zone database handles daylight saving time; offsets are not hardcoded and the endpoint is not used to guess the server's zone. Unknown zones fail before a request is sent.

For example, run `ZSTACK_API_TIMEZONE=Europe/Berlin zstack-cli get zones --context lab` in a shell. In PowerShell, set `$env:ZSTACK_API_TIMEZONE='Europe/Berlin'` before the command and clear it afterward with `Remove-Item Env:ZSTACK_API_TIMEZONE`. This variable applies to environments accessed by that process and is not saved in a context; check it when switching environments. Override the default only for a confirmed server requirement: a correct server should treat different time zone representations of the same instant equivalently. Acceptance of numeric zone abbreviations still depends on the server version; this setting cannot work around every date parsing defect. It does not change the system clock or add automatic retries. The host clock must remain accurate; account session authentication is unaffected.

The default configuration is `~/.zstack-cli/config.yaml`. Set `ZSTACK_CONFIG` to select another file. Keep credential files out of version control. `config view` displays redacted values and its output cannot replace an authenticated configuration file.

## 2. Contexts and diagnostics

```sh
zstack-cli config get-contexts
zstack-cli config use-context lab
zstack-cli config view -o json
zstack-cli get instances --context lab -o json
zstack-cli doctor --context lab --offline -o json
zstack-cli doctor --context lab --request-timeout 15s -o json
```

`--context` selects an environment for one command. To save another login while preserving the default context, use `login ... --context-name staging --no-switch`. Login accepts `--context-name`.

`doctor --offline` checks local configuration. Online doctor also validates an existing account session and queries one zone. It performs no login, renewal or configuration update. An expired session fails even when a password is saved. A successful zone query establishes that query's access and basic response compatibility. Other API permissions and module licenses need their own checks.

```sh
zstack-cli config rename-context staging test
zstack-cli config use-context lab
zstack-cli config delete-context test
zstack-cli logout --context lab
```

Deleting a context removes its local configuration. Ordinary account logout revokes the saved remote session and clears local credentials. `logout --local` clears local credentials only. For AK/SK, logout clears local credentials and leaves the remote key intact. Logout preserves the context name and endpoint.

Configuration updates use an OS-managed sidecar lock and atomic replacement. Preserve lock files while commands are running. A concurrent change to the saved default affects later commands. In-flight commands pin their selected context. Context deletion, rename or credential replacement during renewal can require a retry. Cooperating CLI writers preserve unrelated changes.

## 3. Query, pagination and output

```sh
zstack-cli get instances -q 'state=Stopped' --limit 20 --sort '+uuid' -o json
zstack-cli get instances --all-pages --limit 100 --fields uuid,name,state -o json
zstack-cli get instances --start 20 --limit 20 --reply-with-count -o json
zstack-cli get instances --count
zstack-cli describe instance web-01 -o json
zstack-cli describe volume data-01 --request-timeout 15s -o yaml
```

- Ordinary JSON/YAML queries return a list, with `[]` for empty results. `--reply-with-count` returns `items` and `total`. `--count` prints the total alone.
- `--limit` is the page size and `--start` is the offset. `--all-pages` retrieves subsequent pages. Concurrent resource changes can affect ordering and membership.
- `--fields` projects output. Unknown fields fail. List and detail responses have different structures, so inspect the command's JSON output before choosing fields.
- Queries commonly support table, text, JSON and YAML. Check each command's help for its output contract.
- Resource selection accepts UUIDs or exact names. Single-resource operations reject ambiguous names. Existing VM/image batch operations can select multiple resources sharing a name. Check their target list before execution.
- A failed main detail lookup exits nonzero. Failed optional relation lookups may preserve the main resource and return `warnings`. Automation should inspect those warnings.

Use `get --help` to discover resources. `get volumes` and `get disks` share an implementation. `get jobs` queries long jobs. Query deadlines generally use `--request-timeout`. A single describe deadline covers resolution, main inventory and relations. Read adapters reject malformed/trailing JSON and missing inventories rather than reporting empty success.

## 4. Wait, recover and resize

```sh
zstack-cli wait instance web-01 --for Running --timeout 5m --interval 2s
zstack-cli wait job JOB_UUID --timeout 10m --progress -o json
zstack-cli instance recover VM_UUID
zstack-cli volume recover VOLUME_UUID -o json
zstack-cli volume resize VOLUME_UUID --size 40G --dry-run
zstack-cli volume resize VOLUME_UUID --size 40G -o json
```

Wait polls the state of a resource. A job defaults to the Succeeded target. Failed/Canceled while waiting for success exits nonzero, while an explicit `--for Failed` succeeds when that condition is reached. Progress goes to stderr and the final result goes to stdout. Wait deadlines include authentication, resolution, reads and polling intervals.

VM recovery requires Destroyed and data-volume recovery requires Deleted. Recover root volumes through their VM. Expunged resources cannot be recovered. Resize uses the new total capacity and only accepts growth. G means 1024³ bytes. Guest partitions and filesystems need their own expansion. Recover/resize dry-run is offline and does not establish current resource state.

## 5. IAM management

| Resource | Create | Update | Delete | Additional behavior |
| --- | --- | --- | --- | --- |
| account, user, role | Yes | Yes | Yes | Users belong to the selected account |
| policy | Yes | No | Yes | Defined by a statements file |
| organization | Yes | Yes | Yes | Creation requires type |
| project | Yes | Yes | Yes | recover and expunge |

Prepare a local `account-password` file containing only the new password, preferably with mode 0600. IAM create/update accepts `--password-file`. Login has its own password input. IAM removes one trailing newline and hashes with SHA-512 before sending. Previews and results redact the password and digest.

```sh
zstack-cli create account test-account --password-file ./account-password --dry-run -o json
zstack-cli create account test-account --password-file ./account-password -o json
zstack-cli update account test-account --description 'CLI test account'
zstack-cli update account test-account --description ''
```

An explicitly empty description clears it. A `statements.json` file can contain the following rule.

```json
[{"name":"query-account","effect":"Allow","actions":["org.zstack.header.identity.APIQueryAccountMsg"]}]
```

```sh
zstack-cli create policy query-only --statements-file ./statements.json
zstack-cli create role operator --policy query-only
zstack-cli create organization test-department --type Department
zstack-cli create project test-project --organization test-department
zstack-cli update project test-project --description 'Temporary test project'
```

Creating a policy or role does not assign it to another user. IAM creation uses flags and `--statements-file`, rather than the general `create --file`. IAM2 organization/project operations need the corresponding server license. References use exact resolution. `--resource-uuid` assigns a fixed 32-hex-character creation identity. IAM requests default to a two-minute timeout and accept `--request-timeout`.

Project deletion, recovery and final cleanup follow this sequence. Delete the recovered project again before expunging it.

```sh
zstack-cli delete project test-project --yes
zstack-cli project recover test-project --yes
# Delete again before irreversible expunge.
zstack-cli delete project test-project --yes
zstack-cli expunge project test-project --yes
```

For other IAM resources, for example `delete account test-account`, the CLI prompts for confirmation unless `--yes` is given. Use `--yes` after confirming the intended target. Deletion mode defaults to Permissive. `--delete-mode Enforcing` applies the server's enforcing dependency semantics.

## 6. Long-job operations

```sh
zstack-cli get jobs -o json
zstack-cli describe job JOB_UUID -o json
zstack-cli job resume JOB_UUID --dry-run -o json
zstack-cli job resume JOB_UUID --yes --request-timeout 2m -o json
zstack-cli job clean JOB_UUID --yes -o json
```

Resume accepts Suspended. Clean accepts Succeeded, Failed or Canceled. Dry-run previews the request, while execution resolves and checks the current state. Continue with `wait job` when completion is needed after resume. Local tests cover state protection and request contracts. Existing live long jobs have not been resumed or cleaned during validation.

## 7. Automation errors

```sh
if zstack-cli get instances -o json --error-format json > result.json 2> error.json; then
  cat result.json
else
  status=$?
  printf 'CLI exit: %s\n' "$status"
  cat error.json
fi
```

`--error-format json` controls error diagnostics on stderr. `-o json` controls resource output on stdout. The error envelope contains `error.code`, `message` and `exitCode`, with optional `httpStatus` and `apiCode`. Keep progress and interactive prompts separate from machine-parsed output.

| Exit code | Meaning |
| --- | --- |
| 0 | Success |
| 1 | Unclassified operation failure or partial batch failure |
| 2 | Invalid arguments |
| 3 | Authentication or permission failure |
| 4 | Resource/context not found |
| 5 | State conflict or ambiguous selection |
| 6 | Timeout |
| 7 | Connection failure |
| 8 | API failure or invalid response contract |
| 9 | Output/encoding failure |
| 130 | Canceled |

After a write timeout, output failure or mismatched response identity, query the known UUID to establish whether the server executed the operation. Retrying creation immediately can duplicate resources. Structured diagnostic messages are sanitized, while successful resource output remains business data.

## 8. Regression runners and cleanup

These scripts run from the repository root and require Python 3. Start with a dedicated account session without a saved password.

```sh
export ZSTACK_CONFIG="$PWD/session.yaml"
zstack-cli login 192.0.2.10:8080 --username admin --context-name qa --save-password=false
python3 scripts/readonly_smoke.py --cli ./zstack-cli --config "$ZSTACK_CONFIG" > readonly-report.json
```

Both scripts copy configuration into a private temporary directory and remove inherited ZSTACK environment variables. They use the supplied file's current context. Select it before running, since the scripts have no `--context` option. Accepted credentials are session-only or complete AK/SK. Configurations containing saved account passwords are rejected.

Read-only regression covers 42 canonical query commands. Sample size is 1–3. The command timeout defaults to 30 seconds and accepts 1–120. The suite timeout defaults to 600 seconds and accepts 1–3600. Dependent checks for empty collections and unavailable optional relations are reported as skipped. Global configuration pagination uses numeric id. Concurrent resource changes can cause comparison failures.

The write runner creates and deletes test resources and needs quota, licenses and authorization in a test environment. The default metadata profile covers volume, account, user, policy, role, organization and project.

```sh
python3 scripts/write_e2e.py --cli ./zstack-cli --config "$ZSTACK_CONFIG" \
  --journal ./metadata-run.json --execute
python3 scripts/write_e2e.py --cli ./zstack-cli --config "$ZSTACK_CONFIG" \
  --journal ./compute-run.json --execute --profile compute \
  --image ROOT_IMAGE_UUID --l3-network L3_UUID --host HOST_UUID
```

Use a new journal path for every creation run. `--resources account,user,policy,role` narrows metadata coverage. `--profile all` also executes compute and requires image/network UUIDs. Compute creates a stopped 1-CPU/1-GiB VM and a 1-GiB data volume, validates attachment, snapshot operations, growth to 2 GiB, detachment and recovery, then cleans up and checks the root volume. Recovery checks must observe deletion and subsequent restoration.

The write command timeout defaults to 180 seconds, accepting 1–600. The suite timeout defaults to 1200 seconds, accepting 1–3600. Cleanup has independent per-command deadlines, so the process can run longer than suite-timeout.

After interruption or failure, preserve the journal. Once asynchronous work settles, resume cleanup.

```sh
python3 scripts/write_e2e.py --cli ./zstack-cli --config "$ZSTACK_CONFIG" \
  --journal ./compute-run.json --execute --cleanup-only
zstack-cli logout
```

Cleanup-only reads resource identities from the journal and needs the same endpoint. A new authenticated session for that endpoint can replace an expired one. Cleanup verifies generated names and UUIDs and only detaches volumes from the run's own VM. Unknown outcomes and root-volume leftovers remain unresolved. Preserve those records and use one process per journal.

Evaluate the report's `passed`, `failure`, `cleanup_errors` and process exit status together. A failed test remains failed even when cleanup succeeds. Scripts retain the supplied session, which the caller can log out afterward. SIGKILL and power loss require later journal recovery.

Local `make verify` runs vet, Go race tests, Python tests and builds without cloud credentials. Historical coverage and skips are recorded in [the delivery checklist](next-stage-checklist.md).

## 9. Operations monitoring

Run `monitor usage --context lab` for environment capacity and host load. Select `--section capacity` or `--section load` for one part. Automation can use `-o json --error-format json`.

Capacity includes vCPU, memory, logical primary storage and physical primary storage. Used percentage is (total − available) / total. Memory and storage are bytes in JSON. Load uses 60-second samples from the last 15 minutes, selecting the latest shared CPU/memory timestamp and averaging the hosts with both samples at that timestamp. Reports include sample time and sampled/connected host counts. Percentages without matching data, or with zero total capacity, are N/A in text and null in JSON.

## 10. Individual VMs and ZQL

Use `get instances web-01` for exact-name selection, `get instances -q 'uuid=VM_UUID'` for a UUID filter, and `describe instance VM_UUID` for details. For projections, counts, relations or VM-specific ZWatch metrics, use `zql` and pass the whole statement as one quoted argument.

```sh
zstack-cli zql "query vmInstance.name where uuid='VM_UUID' limit 1" --context lab
```

The selected server parses ZQL. Supported syntax and visible resources depend on server version and account permissions. JSON output preserves results, aggregate fields and returnWith. See [query examples](../examples/queries.en.md) for per-VM CPU time series.

## 11. Install the operations Skill

The CLI bundles the Skill. Run `zstack-cli skill install --agent codex` for a user installation, or choose claude-code, cursor or copilot. `--scope project --project PATH` selects a project. `--dest` overrides the parent directory. `--dry-run` previews the target and `--replace` updates with a backup. Installation works through the native binary on Windows, macOS and Linux. See [the installation guide](../skills/README.en.md) for directory mappings and PowerShell examples.
