# PowerCLI script migration baseline

The PowerShell preview implements the subset below. CLI and module behavior were checked against local fixtures. The evidence record identifies pending Windows and customer acceptance. [Acceptance evidence](../release/acceptance-checklist.md) See the [module API design](../../bmad-output/architecture.md#api-contract).

| VMware entry | Native entry (example brand) | Planned parameter subset | CLI dependency |
| --- | --- | --- | --- |
| `Connect-VIServer` | `Connect-ZStackServer` | `Server`, `Credential` | `login --password-stdin --save-password=false --no-switch` |
| `Disconnect-VIServer` | `Disconnect-ZStackServer` | `Server` connection objects | `logout --context` |
| `Get-VM` | `Get-ZStackVM` | `Name` (arrays, position 0, wildcard), `Id`, `Server` | `get instances --all-pages --sort +uuid -o json` |
| `Start-VM` | `Start-ZStackVM` | `VM` (objects or names in the compatibility module), `Server`, `WhatIf`, `Confirm`, common parameters | `instance start --uuid-only --yes -o json`, then `wait instance` |

The independent compatibility module will expose only this subset. Native names come from the brand configuration. `Get-VM` projects `Name`, `NumCpu`, `MemoryGB`, `PowerState`, and `Id` so that filtering and CSV export can retain their existing shape. `Id` remains a ZStack UUID. With multiple connections, callers select one explicitly. Start waits for `Running`; a timeout can leave the operation outcome unknown.

| Requires review or rewrite | Why |
| --- | --- |
| `Stop-VM`, `-Kill`, or `Shutdown-VMGuest` | Existing CLI stop requests graceful stop. The guest and force semantics need separate verification. |
| `-RunAsync` and VMware tasks | The planned module exposes VM state waiting and has no VMware task object. |
| `Get-View`, `.ExtensionData`, VMware types | These depend on the VMware object and API model. |
| Module-qualified VMware commands or conflicting commands | Compatibility requires a deliberate import in a separate session. Qualified calls still refer to the named module. |

[CLI capability contract](cli-capabilities.json) records source availability separately from package and customer validation. Official references: [Connect-VIServer](https://developer.broadcom.com/powercli/latest/vmware.vimautomation.core/commands/connect-viserver), [Get-VM](https://developer.broadcom.com/powercli/latest/vmware.vimautomation.core/commands/get-vm), [Start-VM](https://developer.broadcom.com/powercli/latest/vmware.vimautomation.core/commands/start-vm), [Stop-VM](https://developer.broadcom.com/powercli/latest/vmware.vimautomation.core/commands/stop-vm).
