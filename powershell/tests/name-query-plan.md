# Name query test plan

[English](name-query-plan.md) | [简体中文](name-query-plan.zh-CN.md) | [Documentation index](../../docs/README.md)

## Goal and scope

Name queries must assert both the **result** and the **requests sent**. A test must fail if an implementation reads the full inventory before issuing an exact-name query, even when it returns the right VM. Automated tests use only Go loopback HTTP servers and the PowerShell loopback API fixture. They do not contact a cloud or perform cloud writes.

The scope includes CLI `get instances NAME`, PowerShell `Get-ZStackVM -Name`, `Get-ZStackVM -Name -Exact`, compatibility `Get-VM -Name`, and every exported resource `Get-* -Name` cmdlet. `-Exact` sends a server-side name condition and keeps only case-sensitive matches locally; regular `-Name` keeps case-insensitive PowerShell wildcard behavior.

| ID | Scenario and input | Required assertion | Automated case |
| --- | --- | --- | --- |
| N01 | CLI full Chinese name with spaces, brackets and `*` | One request with server-side `q=name=...` and `limit=2`; only the named row returns | `TestGetInstancesNameFiltersOnServer` |
| N02 | CLI missing name and two VMs with one name | Empty list or both matches; one filtered request per query | `TestGetInstancesNameReturnsEmptyAndDuplicateMatches` |
| N03 | Regular VM `-Name` with changed case and wildcard | Correct result; request trace records current local filtering | `Run-Contracts.ps1` |
| N04 | VM `-Exact`: Chinese, missing, case variants returned by the server, name array, repeated input | Every name is filtered on the server and checked case-sensitively locally; no full read or missing-name fallback; UUIDs deduplicate | `Run-Contracts.ps1` |
| N05 | VM `-Exact` name contains `*`, `?`, and brackets | Literal name resolves; request preserves the name condition | `Run-Contracts.ps1` |
| N06 | 105 VMs share a name across the default 100-row page | Return 105 objects; both pages carry the same name condition | `Run-Contracts.ps1` |
| N07 | `-Exact` without a name, an empty name-array element, or a mismatched server row | Invalid input makes no query; wrong identity raises a protocol error | `Run-Contracts.ps1` |
| N08 | Every exported `Get-* -Name` cmdlet | Fixture table covers all exported name selectors; uppercase full names select the expected resource; account sample is row 105 | `NameQueryContracts.ps1` |
| N09 | Compatibility `Get-VM`, cross-connection objects, authorization and request failures | Preserve compatibility filtering, connection boundaries and existing error classification | `MigrationContracts.ps1`, `InventoryContracts.ps1`, `Run-Contracts.ps1` |

## Execution and pass criteria

1. Run `make verify` and `make docs-check` locally; neither requires cloud credentials.
2. PR Windows CI runs `powershell/tests/Invoke-ContractPester.ps1` on PowerShell 7.4 and 7.6. The contract suite must fail when a new exported `Get-* -Name` cmdlet has no N08 fixture case.
3. Compare both language guides with the actual command behavior. Passing CI establishes loopback behavior only; it does not prove a real cloud or Gallery installation.

## Remaining risks and live validation

Regular PowerShell `-Name` still reads the full inventory to preserve case-insensitive wildcard behavior. Multiple VM UUIDs passed to `-Id` also read the full inventory, and other resource name selectors use the same local filtering path. Correct N08 results do not establish acceptable query cost; these are separate scalability improvements.

Gallery validation uses a read-only sample on the designated Windows test VM. Record the package version, connection context, name, result count, state and query scope. Limit the target, avoid a full inventory query, and keep credentials and raw server responses out of reports. Merging, tagging and publishing require a separate explicit user instruction.

## Read-only live record, 2026-11-12

Against the user-designated isolated test cloud, the current source CLI used a temporary account session without saving the password. It counted 3 VMs, sampled at most 5, and queried one sampled VM by its full name with a maximum of 3 results. Exactly one VM matched. A missing name returned an empty list, and the temporary CLI session was logged out.

On the Windows test VM, a temporary copy of the Gallery 1.1.5 module was patched with the VM command from PR source commit `c4fadd3`. `Get-ZStackVM -Name NAME -Exact` returned the same Running VM (1 vCPU, 1 GiB). Repeated name input returned one VM; a missing name and a literal `*` suffix returned zero. The PowerShell session was disconnected and the temporary module copy removed. No resource writes or access keys were created.

This first pass sampled one existing VM name path. The PR source ran from a temporary module copy, so this does not imply the new code has been published to Gallery.

## Plan rerun and expanded live checks, 2026-11-12

Local `make verify` passed (Go vet, uncached race tests, 49 Python cases, and Go build); `make docs-check` checked 227 command paths. PR Windows contract tests passed on PowerShell 7.4 and 7.6. Their logs contain the N03–N09 assertions and all 26 exported `Get-* -Name` cases. The English and Chinese PowerShell guides describe `-Name`, `-Exact`, and wildcard behavior consistently with the commands.

The test cloud was first counted through read-only CLI calls: each resource type corresponding to the 26 commands held at most 100 items; roles had the most at 69, followed by images at 22. On the Windows test VM, a temporary copy of Gallery 1.1.5 loaded the PR VM command. All 16 `Get-* -Name` commands with an existing sample returned exactly the expected UUID from an uppercased name; the other 10 resource types were empty and returned zero for a missing name. `Get-ZStackHardDisk` used an existing VM parent object. Compatibility `Get-VM -Name`, VM `-Exact`, duplicate-name deduplication, and a literal `*` suffix also passed. The final PowerShell connection was logged out, and an audit found no remaining temporary config directories. No access key was created and no cloud resource was changed.

The live cloud lacks a VM with Chinese and wildcard characters in its name, 105 same-name VMs, and a deliberately wrong-identity server response. Those N01, N05, N06, and N07 edge cases remain verified through loopback fixtures. The 10 empty resource types could only exercise the no-result path in the live cloud; CI verifies their positive name matches. Early expanded-check attempts with a temporary module import error did not confirm remote logout; a later audit found no remaining local temporary config directories. The successful rerun logged out normally.

## Follow-up checks after the merge

On the same test cloud, a full VM name with different casing still made the server return the original-cased row; the CLI passed that row through. The old PowerShell `-Exact` treated it as `QueryVM.Protocol`. The new N04 fixture models this server behavior: case-only variants are filtered out, exact-case rows are kept, and an unrelated name remains a protocol error. Updated source passed read-only checks on the Windows test VM for regular `-Name` case and wildcard matching, `-Exact` case handling, a two-name array, absent literal `?` and bracket suffixes, and compatibility `Get-VM`. The session was logged out and the temporary module copy removed.

The live cloud has no VM with a literal `?` in its name, so the new N05 loopback case verifies its positive match. N07 also adds an assertion that an empty element in a name array must cause no query.
