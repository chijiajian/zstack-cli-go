# PowerCLI script migration

The table describes the ZStack compatibility subset in packages built from the current source. Import the compatibility manifest in a separate session. The acceptance record tracks fixture tests, live ZStack checks and customer script validation. [Acceptance evidence](../release/acceptance-checklist.md)

| VMware entry | Native entry (example brand) | Planned parameter subset | CLI dependency |
| --- | --- | --- | --- |
| `Connect-VIServer` | `Connect-ZStackServer` | `Server`, `Credential` | `login --password-stdin --save-password=false --no-switch` |
| `Disconnect-VIServer` | `Disconnect-ZStackServer` | `Server` connection objects | `logout --context` |
| `Get-VM` | `Get-ZStackVM` | `Name` (arrays, position 0, wildcard; native `-Exact` for literal names), `Id`, `Server` | Native `-Exact` adds a server-side name condition; other names use `get instances --all-pages --sort +uuid -o json` |
| `Get-HardDisk` | `Get-ZStackHardDisk` | `VM` (live objects or pipeline), `Name`, `Id`, `Server` | `get disks` |
| `Get-NetworkAdapter` | `Get-ZStackNetworkAdapter` | `VM` (live objects or pipeline), `Id`, `Server`. The wrapper looks up the ZStack L3 network name. | `get nics`, `get l3-networks` |
| `Get-VMHost` | `Get-ZStackVMHost` | `Name`, `Id`, `Server`, `Location` (live ZStack cluster objects) | `get hosts` |
| `Get-Cluster` | `Get-ZStackCluster` | `Name`, `Id`, `Server` | `get clusters` |
| `Start-VM` | `Start-ZStackVM` | `VM` (objects or names in the compatibility module), `Server`, `WhatIf`, `Confirm`, common parameters and the ZStack extension `TimeoutSeconds` | `instance start --uuid-only --yes -o json`, then `wait instance` |

The opt-in ZStack module exposes these eight names. `Get-VM` projects `Name`, `NumCpu`, `MemoryGB`, `PowerState`, and `Id` so that filtering and CSV export can retain their existing shape. Use ZStack UUIDs for `Id` and one explicit connection for independent queries when multiple connections are open. Use live objects for related queries and writes. Start waits for `Running` with a per-VM budget of 300 seconds by default. After an unknown outcome, read the VM state before retrying.

| Requires review or rewrite | Why |
| --- | --- |
| `Stop-VM`, `-Kill`, or `Shutdown-VMGuest` | Use the native graceful stop workflow below. Force and guest-tools behavior require separate migration work. |
| `-RunAsync` and VMware tasks | The planned module exposes VM state waiting and has no VMware task object. |
| `Get-View`, `.ExtensionData`, VMware types | These depend on the VMware object and API model. |
| Module-qualified VMware commands or conflicting commands | Compatibility requires a deliberate import in a separate session. Qualified calls still refer to the named module. |

[CLI capability contract](cli-capabilities.json) records source availability separately from package and customer validation. Official references: [Connect-VIServer](https://developer.broadcom.com/powercli/latest/vmware.vimautomation.core/commands/connect-viserver), [Get-VM](https://developer.broadcom.com/powercli/latest/vmware.vimautomation.core/commands/get-vm), [Start-VM](https://developer.broadcom.com/powercli/latest/vmware.vimautomation.core/commands/start-vm), [Stop-VM](https://developer.broadcom.com/powercli/latest/vmware.vimautomation.core/commands/stop-vm).

## Inventory and reporting

The following mappings apply to packages built from the current source. Both entry points return live ZStack objects. Disk `CapacityGB`, NIC `MacAddress` and `NetworkName`, and common VM columns can remain in reports. The compatibility NIC query also requires permission to read the related L3 inventory. Use ZStack `VMId`, `ClusterId`, and `L3NetworkId` for relationships, and the reported byte and hertz fields for capacity. VMware-specific selectors such as disk type, adapter name, host state, and datacenter or folder scope require a migration review. [Inventory examples](../../examples/powershell/inventory.md) · [盘点示例](../../examples/powershell/inventory.zh-CN.md)

| VMware workflow | Native ZStack workflow | Migration detail |
| --- | --- | --- |
| `Get-VM \| Get-HardDisk` | `Get-ZStackVM \| Get-ZStackHardDisk` | Use live ZStack VM objects. `CapacityGB` supports the same style of reports and `CapacityBytes` retains exact integers. Query standalone disks with `Get-ZStackVolume`. |
| `Get-VM \| Get-NetworkAdapter` | `Get-ZStackVM \| Get-ZStackNetworkAdapter` | Use `L3NetworkId` for the ZStack relationship or add `-IncludeNetworkName` to resolve its display name. |
| `Get-VMHost`, `Get-Cluster` | `Get-ZStackVMHost`, `Get-ZStackCluster` | Use the returned cluster with `-Cluster`. `ConnectionState` is ZStack's reported host status. |
| `Get-Datastore` reports | `Get-ZStackPrimaryStorage` or `Get-ZStackImageStorage` | Choose the storage class and capacity fields that match the report. |
| `Get-Template` deployment inventory | `Get-ZStackImage`, `Get-ZStackInstanceOffering`, `Get-ZStackDiskOffering` | Use `MediaType` to distinguish install media and image kinds. Offerings are separate ZStack objects. |
| `Get-VirtualPortGroup` inventory | `Get-ZStackL2Network`, `Get-ZStackL3Network`, `Get-ZStackIPRange` | Traverse native network relationships. `VirtualNetworkId` is the reported L2 ID, with its meaning determined by the network type. |

[VMware Get-HardDisk](https://developer.broadcom.com/powercli/latest/vmware.vimautomation.core/commands/get-harddisk) · [Get-NetworkAdapter](https://developer.broadcom.com/powercli/latest/vmware.vimautomation.core/commands/get-networkadapter) · [Get-VMHost](https://developer.broadcom.com/powercli/latest/vmware.vimautomation.core/commands/get-vmhost) · [Get-Cluster](https://developer.broadcom.com/powercli/latest/vmware.vimautomation.core/commands/get-cluster) · [Get-VirtualPortGroup](https://developer.broadcom.com/powercli/latest/vmware.vimautomation.core/commands/get-virtualportgroup).

## VM lifecycle

These native cmdlets are available in packages built from the current source. Replace VMware references and identifiers with live objects or UUIDs from the selected ZStack connection. [PowerShell lifecycle examples](../../examples/powershell/README.md#vm-lifecycle) · [生命周期示例](../../examples/powershell/README.zh-CN.md#vm-生命周期)

| VMware workflow | Native ZStack cmdlet | Migration detail |
| --- | --- | --- |
| `New-VM` | `New-ZStackVM` | Select a ZStack root volume template, L3 network, sizing and explicit initial state. |
| `Set-VM` | `Set-ZStackVM` | Update the supplied metadata or change CPU and memory on a stopped VM. |
| `Suspend-VM` | `Suspend-ZStackVM`, `Resume-ZStackVM` | Use ZStack's in-memory pause and resume states. |
| `Stop-VM` or `Stop-VMGuest` | `Stop-ZStackVM` | Review the workflow for graceful ACPI shutdown and explicit HA behavior. |
| `Restart-VM` | `Restart-ZStackVM` | Wait for the platform reboot request and observe VM state. Verify guest startup separately when required. |
| `Remove-VM` | `Remove-ZStackVM`, `Restore-ZStackVM` | Use a recoverable VM deletion policy and confirm deletion. Recovery is a separate operation. |

[VMware VM cmdlet reference](https://developer.broadcom.com/powercli/latest/products/vmwarevsphereandvsan/categories/vm)

## Disk and device operations

| VMware PowerCLI pattern | ZStack cmdlets and migration scope |
| --- | --- |
| `New-HardDisk`, `Remove-HardDisk`, `Set-HardDisk -CapacityGB` | Use `New-ZStackVolume` and `Add-ZStackHardDisk` for provisioning and attachment. `Remove-ZStackHardDisk` detaches the data volume. Use `Remove-ZStackVolume` separately for recoverable deletion. `Set-ZStackHardDisk` grows the virtual disk to its requested total capacity. |
| `Get-Snapshot`, `New-Snapshot`, `Remove-Snapshot` | Use the `*-ZStackVolumeSnapshot` commands on individual ZStack volumes. Each snapshot captures one ZStack volume. |
| `New-NetworkAdapter`, `Remove-NetworkAdapter` | Use `New-ZStackNetworkAdapter -L3Network` and `Remove-ZStackNetworkAdapter`. Use the ZStack L3 network reference and retain the returned adapter identity. |
| `Get-CDDrive`, `Set-CDDrive -IsoPath/-NoMedia` | Use `Get-ZStackCDDrive` and `Set-ZStackCDDrive -IsoImage/-NoMedia`. Select a registered ZStack ISO image and the exact virtual drive. |

## Platform cmdlets

Usage reports, ZQL, IAM, security group and load balancer inventories, and long jobs use the native module naming and connection conventions. A long job uses the UUID returned by ZStack's long-job inventory. Wait errors distinguish a backend Failed or Canceled state from a client deadline. Wait and resume return the same inventory projection as Get, read back by UUID within the caller's deadline. The module does not expose a percentage when the API only reports job state.

A command may add `-RunAsync` when its successful submission returns a typed long-job UUID, a fresh connection can read the same job by UUID, and real acceptance verifies target ownership, permission, terminal states and retry behavior. An API callback location or a resource UUID alone does not satisfy that contract. Platform write candidates continue to use their documented CLI paths while those contracts are assessed.
