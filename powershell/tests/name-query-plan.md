# Name query test plan

[English](name-query-plan.md) | [简体中文](name-query-plan.zh-CN.md) | [Documentation index](../../docs/README.md)

## Goal and scope

Name queries must assert both the **result** and the **requests sent**. A test must fail if an implementation reads the full inventory before issuing an exact-name query, even when it returns the right VM. Automated tests use only Go loopback HTTP servers and the PowerShell loopback API fixture. They do not contact a cloud or perform cloud writes.

The scope includes CLI `get instances NAME`, PowerShell `Get-ZStackVM -Name`, `Get-ZStackVM -Name -Exact`, compatibility `Get-VM -Name`, and every exported resource `Get-* -Name` cmdlet. `-Exact` is a case-sensitive server-side lookup; regular `-Name` keeps case-insensitive PowerShell wildcard behavior.

| ID | Scenario and input | Required assertion | Automated case |
| --- | --- | --- | --- |
| N01 | CLI full Chinese name with spaces, brackets and `*` | One request with server-side `q=name=...` and `limit=2`; only the named row returns | `TestGetInstancesNameFiltersOnServer` |
| N02 | CLI missing name and two VMs with one name | Empty list or both matches; one filtered request per query | `TestGetInstancesNameReturnsEmptyAndDuplicateMatches` |
| N03 | Regular VM `-Name` with changed case and wildcard | Correct result; request trace records current local filtering | `Run-Contracts.ps1` |
| N04 | VM `-Exact`: Chinese, missing, case mismatch, name array, repeated input | Every name is filtered on the server; no preceding full read or missing-name fallback; UUIDs deduplicate | `Run-Contracts.ps1` |
| N05 | VM `-Exact` name contains `*` and brackets | Literal name resolves; request preserves the name condition | `Run-Contracts.ps1` |
| N06 | 105 VMs share a name across the default 100-row page | Return 105 objects; both pages carry the same name condition | `Run-Contracts.ps1` |
| N07 | `-Exact` without a name and a mismatched server row | Invalid input makes no query; wrong identity raises a protocol error | `Run-Contracts.ps1` |
| N08 | Every exported `Get-* -Name` cmdlet | Fixture table covers all exported name selectors; uppercase full names select the expected resource; account sample is row 105 | `NameQueryContracts.ps1` |
| N09 | Compatibility `Get-VM`, cross-connection objects, authorization and request failures | Preserve compatibility filtering, connection boundaries and existing error classification | `MigrationContracts.ps1`, `InventoryContracts.ps1`, `Run-Contracts.ps1` |

## Execution and pass criteria

1. Run `make verify` and `make docs-check` locally; neither requires cloud credentials.
2. PR Windows CI runs `powershell/tests/Invoke-ContractPester.ps1` on PowerShell 7.4 and 7.6. The contract suite must fail when a new exported `Get-* -Name` cmdlet has no N08 fixture case.
3. Compare both language guides with the actual command behavior. Passing CI establishes loopback behavior only; it does not prove a real cloud or Gallery installation.

## Remaining risks and live validation

Regular PowerShell `-Name` still reads the full inventory to preserve case-insensitive wildcard behavior. Multiple VM UUIDs passed to `-Id` also read the full inventory, and other resource name selectors use the same local filtering path. Correct N08 results do not establish acceptable query cost; these are separate scalability improvements.

Gallery validation should be a separately authorized read-only sample on the designated Windows test VM. Record the package version, connection context, name, result count, state and query scope. Limit the target, avoid a full inventory query, and keep credentials and raw server responses out of reports. Merging, tagging and publishing require a separate explicit user instruction.
