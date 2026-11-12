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

## Module repositories and publication

Local installation and `Import-Module` do not require registering the module with PowerShell or PowerShell Gallery. `PSGallery` is a commonly preconfigured client repository. A "default repository" only affects which configured source an install or search command selects when none is specified; it does not mean this module has been published or reviewed by Microsoft. Use `Get-PSResourceRepository` to inspect the client configuration. If PSGallery was removed, `Register-PSResourceRepository -PSGallery` restores it. Register a private repository by its address on each client that needs it, and specify `-Repository` when installing to select the intended source.

Publishing to PowerShell Gallery requires a Gallery account and API key, an approved module name and package owner, and distribution rights for the bundled binary. Publish the complete module directory, not the offline ZIP shown above. The generated manifest contains version, author, description, license URI, and Windows/PowerShell Core tags; brand configuration can also supply `projectUri`. Validate the generated file with `Test-ModuleManifest`. The license URI must match the delivered license. The `ZStack.PowerCLI` brand points its help and project URIs to this source repository. Confirm the public brand, links, signing policy, and customer acceptance before publishing. [Publication status](../powershell/release/acceptance-checklist.md)

The tag release workflow publishes `ZStack.PowerCLI` to PSGallery after the Go checks, Windows PowerShell contracts, and GitHub Release succeed. It accepts stable `vMAJOR.MINOR.PATCH` tags; prerelease tags containing `-` skip the Gallery job. The module version follows the tag. Store the Gallery API key as a repository **Actions secret** named `PSGALLERY_API_KEY` in `terraform-zstack-modules/zstack-cli-go` (Settings → Secrets and variables → Actions), not as a plain variable. The key needs permission to push new packages and versions for `ZStack.PowerCLI`. A missing secret fails the Gallery job. Creating the secret or publishing a stable tag is a separate release action; the workflow file alone does not publish a package.

## Migrate a script

Use a ZStack account and the API endpoint in `host[:port]` form. See [migration examples](../examples/powershell/README.md) and the [parameter compatibility matrix](../powershell/contracts/compatibility.md).

| Task | Native preview | Migration notes |
| --- | --- | --- |
| Connect | `Connect-ZStackServer -Server HOST -Credential (Get-Credential)` | Account login saves only the session. Pass the returned connection explicitly when more than one is active. |
| Inventory | `Get-ZStackVM -Name 'web*' -Server $server` | `Name`, `NumCpu`, `MemoryGB`, `PowerState` and `Id` are available. `Id` is a ZStack UUID. `MemoryGB` uses 2³⁰ bytes per unit. |
| Start | `Get-ZStackVM -Id VM_UUID -Server $server \| Start-ZStackVM -WhatIf` | Preview first. When authorized, remove `-WhatIf` and use `-Confirm` as appropriate. Each VM has a shared deadline for target resolution, the start request and verification of `Running`. The default is 300 seconds and can be set with `-TimeoutSeconds`. |
| Wait | `$vm \| Wait-ZStackVM -State Running` | Returns a refreshed VM. This does not expose a VMware task object. |
| Disconnect | `Disconnect-ZStackServer -Server $server` | Attempts remote logout and clears the module-owned local session. |

Use `-ErrorAction Stop` when a failure should terminate the pipeline. After an unknown write result, query the VM state before retrying. Live objects carry their connection. Pass an explicit `-Id` and `-Server` when using data read back from a CSV or serialized session.

Starting with 0.1.2, run `Connect-ZStackServer -Server HOST -AccessKeyId KEY_ID -AccessKeySecret (Read-Host 'SK' -AsSecureString)` to use an existing key pair. The module checks the signature with a VM query. Signing depends on an accurate system clock. Use `-ApiTimeZone Asia/Shanghai` if the endpoint needs a specific signing time zone. The module saves the key pair in the connection's private temporary directory and deletes that directory on disconnect.

To reuse legacy names, import `Compatibility/ZStack.PowerCLI.Compatibility.psd1` from the same ZStack package in a separate PowerShell session. The expanded package supports a parameter subset of `Connect-VIServer`, `Disconnect-VIServer`, `Get-VM`, `Get-VMHost`, `Get-Cluster`, `Get-HardDisk`, `Get-NetworkAdapter`, and `Start-VM`. You can retain `Get-VM | Get-HardDisk`, `Get-VM | Get-NetworkAdapter`, and common CSV projections. Import is rejected if VMware, Hyper-V, or user commands already occupy those names. [Public PowerCLI scenarios](../examples/powershell/README.md#public-powercli-scenarios)

Use `Start-VM -TimeoutSeconds` to set the per-VM budget shared by target resolution, submission, and state verification. Migrate shutdown scripts to `Stop-ZStackVM` for graceful ACPI shutdown and `Wait-ZStackVM` for platform-state checks. Use a separate probe when the workflow depends on guest or application readiness. Whole-VM snapshots, `-Kill`, `-RunAsync`, `Get-View`, `.ExtensionData`, and VMware Task need separate migration work. Query and manage volume snapshots per volume.

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

## VM lifecycle in the current source

The native write cmdlets accept a live VM from the same module or an exact ZStack `-Id`. They support `-WhatIf` and `-Confirm` per target. Select `-Server` for UUIDs when multiple connections are active. A successful write returns a refreshed VM object. The default budget is 300 seconds per target for lookup, the API request and observation. Use `-ErrorAction Stop` to terminate a batch on the first failure.

| Cmdlet | Input and observed result |
| --- | --- |
| `New-ZStackVM` | Supply a root volume template with `-Image` and at least one `-L3Network`. Use `-InstanceOffering` or both `-NumCpu` and `-MemoryGB`. Choose `-State Stopped` or `-State Running` explicitly. Deployment references accept live objects or UUIDs. |
| `Set-ZStackVM` | Changes only the supplied `-Name`, `-Description`, `-NumCpu` or `-MemoryGB`. An empty description clears it. CPU and memory changes require `Stopped`. |
| `Suspend-ZStackVM` and `Resume-ZStackVM` | Pause `Running` in memory and resume `Paused` to `Running`. Repeating an operation in its target state returns a fresh object without a write. [ZStack VM lifecycle](https://docs.zstack.io/en/zcf/zstack-cloud/4.8.38/dev-guide?section=ZStack_2.1_Dev_Guide_0002&version=4.8.38) |
| `Stop-ZStackVM` | Request a graceful ACPI shutdown from `Running` and observe `Stopped`. `-StopHA` explicitly requests the platform's HA shutdown behavior. [StopVmInstance](https://docs.zstack.io/docs/api-explorer/zh-cn/zsphere/explorer/StopVmInstance) |
| `Restart-ZStackVM` | Submit one reboot from `Running`, await API completion, then observe `Running`. If the workflow needs guest restart proof, compare the guest boot time or boot ID before and after. [RebootVmInstance](https://docs.zstack.io/en/zcf/zstack-cloud/5.5.30/dev-guide?section=ZStack_2.1_Dev_Guide_0046&version=5.5.30) |
| `Remove-ZStackVM` and `Restore-ZStackVM` | Remove a stopped VM with high confirmation impact and observe `Destroyed`. Removal requires access to verify the `Delay` or `Never` VM deletion policy. Restore a recoverable VM and its root disk through the VM. The result is `Stopped`. The platform policy controls retention and attached data disks. [Deletion policy](https://docs.zstack.io/zh/docs/zcf/zstack-cloud/user-guide?section=ZStack_User_Guide_0043_5) |
| `Wait-ZStackVM` | Wait for `Running` (default), `Stopped`, `Paused` or `Destroyed`. For other requested states, observing `Destroyed` ends the wait with a conflict. |

Use `RawState` for lifecycle conditions. After a remove or restore, use the returned VM or query by UUID to obtain an object for subsequent operations. If a write reports `OutcomeUnknown`, query the VM before retrying. For creation, use the UUID from the error target or supply a known `-ResourceId` when starting a workflow that needs to reconcile by UUID. [Lifecycle examples](../examples/powershell/README.md#vm-lifecycle) · [Migration mappings](../powershell/contracts/compatibility.md#vm-lifecycle)

## Disks, volume snapshots, and VM devices

Use a package built from this source with its matching CLI. Ejection requires that the selected ISO be associated with exactly one drive on the VM.

| Cmdlets | Behavior |
| --- | --- |
| `New-ZStackVolume`, `Remove-ZStackVolume`, `Restore-ZStackVolume` | Create an independent data volume with a capacity or disk offering. Delete a detached data volume to the recycle bin under a Delay or Never policy, and recover it by the same ID. |
| `Add-ZStackHardDisk`, `Remove-ZStackHardDisk`, `Set-ZStackHardDisk` | Attach an existing data volume, detach it while retaining the volume, or grow a virtual disk to a total `-CapacityGB`. Resizing an attached disk requires a stopped VM and leaves guest partitions untouched. |
| `Get-ZStackVolumeSnapshot`, `New-ZStackVolumeSnapshot`, `Restore-ZStackVolumeSnapshot`, `Remove-ZStackVolumeSnapshot` | Query, create, revert to, or remove a snapshot of an individual volume. Revert requires a stopped VM when the volume is attached. Removal requires a leaf snapshot. |
| `New-ZStackNetworkAdapter`, `Remove-ZStackNetworkAdapter` | Attach an adapter to a selected L3 network and remove an exact adapter. A VM can retain other adapters on the same L3. |
| `Get-ZStackCDDrive`, `Set-ZStackCDDrive` | List drives with their ISO IDs. Insert an existing ISO with `-IsoImage` or eject it with `-NoMedia` on a selected drive. |

Write commands support per-target `-WhatIf`, `-Confirm`, and `-TimeoutSeconds`. Supply live module objects or exact UUIDs as documented. When operating by UUID, include the expected VM or parent volume where the command requires it. After an unknown result, query the target before retrying. See the [PowerShell examples](../examples/powershell/README.md#disks-snapshots-and-vm-devices).

Independent volume creation reads back capacity and any specified disk offering or primary storage. NIC attachment reads back any supplied MAC address and static IP. If the requested values cannot be verified, the cmdlet raises `OutcomeUnknown`. Query the target before deciding whether to retry. After a snapshot revert, verify the expected file or block data in the guest.

## ZStack platform operations

| Workflow | Cmdlets and output |
| --- | --- |
| Capacity and load | `Get-ZStackUsage [-Section All\|Capacity\|Load]` returns a report for one connection. Its `Capacity` rows have `Metric`, `Unit`, `Total`, `Used`, `Available`, and `Percent`. CPU capacity uses vCPU. Memory and logical and physical primary storage use bytes. `Load` contains the latest common CPU and memory sample time, the percentages and the sampled and connected host counts. Missing samples and percentages are `$null`. An unrequested section is also `$null`. |
| ZQL | `Invoke-ZStackZql -Query <string>` accepts ZQL query and count statements, including aggregates and multiple result sets. Each output object has a zero-based `Index`, `Records` with the selected columns, `Total`, `Counts`, `Aggregates`, `ReturnWith`, and the full result set in `Data`. The cloud API parses the statement and enforces the selected account's permissions. |
| Long jobs | `Get-ZStackLongJob [-Id <id>]`, `Wait-ZStackLongJob`, and `Resume-ZStackLongJob` use real cloud job IDs. Query objects include `JobId`, `Name`, `State`, `JobName`, `TargetResourceId`, and any reported `JobResult`. The API reports progress through states. |
| IAM audit | `Get-ZStackAccount`, `Get-ZStackUser`, `Get-ZStackRole`, `Get-ZStackPolicy`, and `Get-ZStackProject` list the resources visible to the connection. User and policy objects expose `AccountId`. IAM2 projects expose `LinkedAccountId`. Policy output includes `StatementCount`. |
| Network audit | `Get-ZStackSecurityGroup`, `Get-ZStackSecurityGroupRule`, `Get-ZStackVip`, `Get-ZStackLoadBalancer`, `Get-ZStackLoadBalancerListener`, and `Get-ZStackLoadBalancerServerGroup` return objects linked by resource IDs. A security group can feed its rules through the pipeline. A VIP can feed its load balancers, and a load balancer can feed its listeners and server groups. |

`Wait-ZStackLongJob` defaults to `Succeeded`. For a default wait, a failed or canceled cloud job raises `JobFailed` or `JobCanceled`. You can explicitly wait for `-State Failed` or `-State Canceled` to read that terminal result. A local timeout leaves the cloud job running according to its own state. To continue in a fresh session, reconnect to the same cloud and pass the saved ID with `-Id`.

`Resume-ZStackLongJob` targets a suspended job and supports `-WhatIf` and confirmation. Wait and resume read the job by UUID within the command deadline and return the same object fields as `Get-ZStackLongJob`. If the job is still `Suspended` or its state cannot be verified after the resume request, the cmdlet raises `OutcomeUnknown`. Check the job before retrying. General `-RunAsync` support requires the initiating operation to return a real long job ID that can be queried on a new connection. Ordinary API completion polling and VM state waiting have their own contracts.

The inventory cmdlets page within the selected connection. An empty result emits no objects, and a permission denial raises `AuthenticationFailed`. Use a live object from the same connection in a parent pipeline. See the [PowerShell platform examples](../examples/powershell/README.md#zstack-platform-operations).
