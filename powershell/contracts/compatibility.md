# PowerCLI script migration baseline

The 0.1.2 preview supports the compatibility subset below. The inventory section describes native query commands available when building from the current source. The acceptance record separates fixture tests, live ZStack checks and customer script validation. [Acceptance evidence](../release/acceptance-checklist.md)

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

## Inventory and reporting

The following mappings apply to packages built from the current source. The native names preserve report shapes while exposing ZStack identities and relationships. Use the existing compatibility module for its four documented commands. [Inventory examples](../../examples/powershell/inventory.md) · [盘点示例](../../examples/powershell/inventory.zh-CN.md)

| VMware workflow | Native ZStack workflow | Migration detail |
| --- | --- | --- |
| `Get-VM \| Get-HardDisk` | `Get-ZStackVM \| Get-ZStackHardDisk` | Use live ZStack VM objects. `CapacityGB` supports the same style of reports and `CapacityBytes` retains exact integers. Query standalone disks with `Get-ZStackVolume`. |
| `Get-VM \| Get-NetworkAdapter` | `Get-ZStackVM \| Get-ZStackNetworkAdapter` | Use `L3NetworkId` for the ZStack relationship or add `-IncludeNetworkName` to resolve its display name. |
| `Get-VMHost`, `Get-Cluster` | `Get-ZStackVMHost`, `Get-ZStackCluster` | Use the returned cluster with `-Cluster`. `ConnectionState` is ZStack's reported host status. |
| `Get-Datastore` reports | `Get-ZStackPrimaryStorage` or `Get-ZStackImageStorage` | Choose the storage class and capacity fields that match the report. |
| `Get-Template` deployment inventory | `Get-ZStackImage`, `Get-ZStackInstanceOffering`, `Get-ZStackDiskOffering` | Use `MediaType` to distinguish install media and image kinds. Offerings are separate ZStack objects. |
| `Get-VirtualPortGroup` inventory | `Get-ZStackL2Network`, `Get-ZStackL3Network`, `Get-ZStackIPRange` | Traverse native network relationships. `VirtualNetworkId` is the reported L2 ID, with its meaning determined by the network type. |

VMware references: [Get-HardDisk](https://developer.broadcom.com/powercli/latest/vmware.vimautomation.core/commands/get-harddisk), [Get-VMHost](https://developer.broadcom.com/powercli/latest/vmware.vimautomation.core/commands/get-vmhost), [Get-VirtualPortGroup](https://developer.broadcom.com/powercli/latest/vmware.vimautomation.core/commands/get-virtualportgroup).
