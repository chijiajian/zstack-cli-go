# PowerShell module preview

[English](powershell.md) | [简体中文](powershell.zh-CN.md) | [Documentation](README.md)

The preview targets Windows x64 and PowerShell 7.4 or later. See the [evidence record](../powershell/release/acceptance-checklist.md) for the exact tested combinations. `ZStack.PowerCLI` is the example brand. The package contains the CLI executable, so module users install only the module package. Go, Python and Pester are development dependencies.

## Install and load

Obtain the complete ZIP and SHA-256 from your provider. On Windows, run the [offline installer](../powershell/packaging/Install-Offline.ps1) against that ZIP. The installer defaults to your user module directory. Alternatively, extract the ZIP into a trusted directory and import the manifest by full path.

```powershell
# Replace the path, version and checksum with the supplied package details.
$manifest = .\Install-Offline.ps1 -Archive .\ZStack.PowerCLI-0.1.2-windows-amd64.zip -SHA256 '<provider checksum>'
Import-Module $manifest
```

After installation into a directory on `$env:PSModulePath`, you can use `Import-Module ZStack.PowerCLI`. Import loads local files. Installation from a configured module repository is a separate step once the provider publishes there. Keep each version in its own directory and open a fresh session when switching versions. [Upgrade and rollback](../powershell/packaging/Upgrade-Policy.md)

Installing an upgrade preserves the CLI used by existing connections. If the upgrade package fails validation, the installed versions remain available. To roll back, start a new session, import the previous version and reconnect.

## Migrate a script

Use a ZStack account and the API endpoint in `host[:port]` form. See [migration examples](../examples/powershell/README.md) and the [parameter compatibility matrix](../powershell/contracts/compatibility.md).

| Task | Native preview | Migration notes |
| --- | --- | --- |
| Connect | `Connect-ZStackServer -Server HOST -Credential (Get-Credential)` | Account login saves only the session. Pass the returned connection explicitly when more than one is active. |
| Inventory | `Get-ZStackVM -Name 'web*' -Server $server` | `Name`, `NumCpu`, `MemoryGB`, `PowerState` and `Id` are available. `Id` is a ZStack UUID. `MemoryGB` uses 2³⁰ bytes per unit. |
| Start | `Get-ZStackVM -Id VM_UUID -Server $server \| Start-ZStackVM -WhatIf` | Preview first. When authorized, remove `-WhatIf` and use `-Confirm` as appropriate. Each VM has a shared 300-second deadline for target resolution, the start request and verification of `Running`. |
| Wait | `$vm \| Wait-ZStackVM -State Running` | Returns a refreshed VM. This does not expose a VMware task object. |
| Disconnect | `Disconnect-ZStackServer -Server $server` | Attempts remote logout and clears the module-owned local session. |

Use `-ErrorAction Stop` when a failure should terminate the pipeline. After an unknown write result, query the VM state before retrying. Live objects carry their connection. Pass an explicit `-Id` and `-Server` when using data read back from a CSV or serialized session.

Starting with 0.1.2, run `Connect-ZStackServer -Server HOST -AccessKeyId KEY_ID -AccessKeySecret (Read-Host 'SK' -AsSecureString)` to use an existing key pair. The module checks the signature with a VM query. Signing depends on an accurate system clock. Use `-ApiTimeZone Asia/Shanghai` if the endpoint needs a specific signing time zone. The module saves the key pair in the connection's private temporary directory and deletes that directory on disconnect.

For the limited legacy command names, import the `Compatibility/ZStack.PowerCLI.Compatibility.psd1` manifest from the same package in a separate PowerShell session. Loaded VMware, Hyper-V or user commands with those names cause import to fail. Unsupported VMware parameters and objects need script changes. Stop, guest shutdown, snapshots, `-RunAsync`, `Get-View` and `.ExtensionData` are outside the preview.

For an OEM package, use the module name and command prefix supplied by the provider. Object properties and script pipeline behavior are shared across brands.

Use `Get-Help Get-ZStackVM -Full` or the corresponding OEM command to see the brand, module version, CLI build, connection namespace and provider help link. The [paired migration examples](../examples/powershell/README.md#before-and-after) cover reporting, explicit connections, startup, WhatIf, ErrorAction and OEM compatibility.

## Inventory and reporting in the current source

Use a module package built with the matching Windows x64 CLI. The following commands are available in packages built from this source. Each resource produces a PowerShell object with a ZStack UUID in `Id`, a `ConnectionId`, and resource-specific properties. An empty query emits no objects. [Inventory examples](../examples/powershell/inventory.md) · [VMware migration mappings](../powershell/contracts/compatibility.md#inventory-and-reporting) · [Package build policy](../powershell/packaging/Upgrade-Policy.md)

Use `-Server` to select a connection when multiple connections are active. Queries scoped by a live parent object inherit its connection. An explicit `-Server` must match that connection. Where supported, `-Id` matches ZStack UUIDs exactly and `-Name` accepts case-insensitive PowerShell wildcards.

| Cmdlet | Resource filters | Report properties |
| --- | --- | --- |
| `Get-ZStackHardDisk` | `-VM` or VM pipeline, `-Id`, `-Name` | `VMId`, `Type`, `CapacityBytes`, `ActualSizeBytes`, `CapacityGB` |
| `Get-ZStackVolume` | `-Id`, `-Name`, `-Attached` or `-Unattached` | `VMId`, `Attached`, `PrimaryStorageId`, `CapacityBytes`, `ActualSizeBytes` |
| `Get-ZStackNetworkAdapter` | `-VM` or VM pipeline, `-Id`, `-IncludeNetworkName` | `VMId`, `MacAddress`, `IPAddress`, `L3NetworkId`, `NetworkName` |
| `Get-ZStackVMHost` | `-Id`, `-Name`, `-Cluster` | `ClusterId`, `ConnectionState`, `NumCpu`, `CpuTotalHz`, `MemoryTotalBytes` |
| `Get-ZStackCluster` | `-Id`, `-Name`, `-Zone` | `ZoneId`, `State`, `HypervisorType` |
| `Get-ZStackZone` | `-Id`, `-Name` | `State`, `Type`, `Description` |
| `Get-ZStackPrimaryStorage` | `-Id`, `-Name`, `-Zone` | `ZoneId`, `CapacityBytes`, `FreeSpaceBytes`, `PhysicalCapacityBytes`, `PhysicalFreeSpaceBytes`, `SystemUsedCapacityBytes` |
| `Get-ZStackImageStorage` | `-Id`, `-Name` | `Type`, `CapacityBytes`, `FreeSpaceBytes` |
| `Get-ZStackImage` | `-Id`, `-Name`, `-MediaType` | `MediaType`, `Format`, `CapacityBytes`, `ActualSizeBytes` |
| `Get-ZStackInstanceOffering` | `-Id`, `-Name` | `NumCpu`, `MemoryBytes`, `MemoryGB` |
| `Get-ZStackDiskOffering` | `-Id`, `-Name` | `CapacityBytes`, `CapacityGB`, `Type` |
| `Get-ZStackL2Network` | `-Id`, `-Name`, `-Zone` | `ZoneId`, `VirtualNetworkId`, `PhysicalInterface` |
| `Get-ZStackL3Network` | `-Id`, `-Name`, `-L2Network` | `L2NetworkId`, `IPVersion`, `DnsDomain` |
| `Get-ZStackIPRange` | `-Id`, `-Name`, `-L3Network` | `L3NetworkId`, `CIDR`, `StartIP`, `EndIP`, `Gateway` |

Use `Bytes` fields for exact arithmetic and `GB` fields for display in units of 2³⁰ bytes. Reported zero stays zero. An unreported capacity is `$null`. `SystemUsedCapacityBytes` retains the API's signed value. `-IncludeNetworkName` resolves the adapter's L3 name within its connection when the network is available.

`Get-ZStackVM -Id VM_UUID -Detail` accepts one exact ZStack UUID. It keeps the ordinary VM properties and adds host, image and instance-offering IDs and names, `L3NetworkIds`, and `Warnings`. A warning identifies the affected relation and gives a fixed failure category such as `not_found` or `authentication_failed`.
