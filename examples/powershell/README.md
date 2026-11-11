# PowerShell migration examples

[English](README.md) | [简体中文](README.zh-CN.md) | [PowerShell guide](../../docs/powershell.md) | [Inventory and reporting examples](inventory.md)

The following example assumes the preview package is installed and the endpoint and account have been authorized. Substitute the module name and command prefix for an OEM package.

```powershell
Import-Module ZStack.PowerCLI
$server = Connect-ZStackServer -Server '192.0.2.10:8080' -Credential (Get-Credential)
try {
    Get-ZStackVM -Name 'web*' -Server $server |
        Where-Object PowerState -eq 'PoweredOn' |
        Select-Object Name, NumCpu, MemoryGB |
        Export-Csv .\inventory.csv -NoTypeInformation -Encoding utf8

    Get-ZStackVM -Id '<VM_UUID>' -Server $server | Start-ZStackVM -WhatIf -ErrorAction Stop
} finally {
    Disconnect-ZStackServer -Server $server
}
```

When porting a VMware report, the filter and CSV stages remain the same. Replace the connection and VM lookup. Change the VM identifier to the UUID of the authorized ZStack resource before running the preview. [Supported parameter subset](../../powershell/contracts/compatibility.md)

## Before and after

In the VMware column, `$vi` is the connection returned by `Connect-VIServer`. In the ZStack column, `$server` is the connection from the example above. Replace `<VM_UUID>` with the UUID of a VM you are authorized to manage. Start examples use `-WhatIf` so you can review the target. Remove it only when you have authorization to start the VM.

| Scenario | VMware PowerCLI | ZStack PowerCLI preview |
| --- | --- | --- |
| Inventory report | `Get-VM -Server $vi \| Select-Object Name, NumCpu, MemoryGB \| Export-Csv .\inventory.csv -NoTypeInformation` | `Get-ZStackVM -Server $server \| Select-Object Name, NumCpu, MemoryGB \| Export-Csv .\inventory.csv -NoTypeInformation` |
| Explicit connection | `Get-VM -Name 'web*' -Server $vi` | `Get-ZStackVM -Name 'web*' -Server $server` |
| Start from a VM pipeline | `Get-VM -Id $vSphereId -Server $vi \| Start-VM -WhatIf` | `Get-ZStackVM -Id '<VM_UUID>' -Server $server \| Start-ZStackVM -WhatIf` |
| Review targets with WhatIf | `Start-VM -VM $viVm -WhatIf` | `Start-ZStackVM -Id '<VM_UUID>' -Server $server -WhatIf` |
| Stop a script when a command fails | `Get-VM -Name 'web-prod' -Server $vi -ErrorAction Stop` | `Get-ZStackVM -Id '<VM_UUID>' -Server $server -ErrorAction Stop` |
| Opt into legacy names in an OEM session | `Import-Module VCF.PowerCLI` followed by `Get-VM -Name 'web*' -Server $vi` | In a separate session, `Import-Module $oemCompatibilityManifest` followed by `Get-VM -Name 'web*' -Server $oemConnection` |

Set `$oemCompatibilityManifest` to the manifest in the installed OEM package, such as `Compatibility/Acme.Cloud.Compatibility.psd1`. Create `$oemConnection` using the same package's `Connect-AcmeServer`. When using the native OEM commands, call `Get-AcmeVM -Server $oemConnection` and run `Get-Help Get-AcmeVM -Full` for package details. Requery by UUID when a VM object came from CSV or an earlier session.

## Public PowerCLI scenarios

In a separate session, import the installed native module and its compatibility manifest. Use a ZStack connection and adapt the VM names and UUIDs to your environment.

```powershell
$module = Import-Module ZStack.PowerCLI -PassThru
Import-Module (Join-Path $module.ModuleBase 'Compatibility/ZStack.PowerCLI.Compatibility.psd1')
$server = Connect-VIServer -Server '192.0.2.10:8080' -Credential (Get-Credential)
$vms = Get-VM -Name 'app-*' -Server $server
$vms | Select-Object Name, NumCpu, MemoryGB, PowerState, Id |
    Export-Csv ./vm-inventory.csv -NoTypeInformation
$vms | Get-HardDisk | Select-Object VMId, Name, CapacityGB
$vms | Get-NetworkAdapter | Select-Object VMId, MacAddress, NetworkName
$clusters = Get-Cluster -Server $server
Get-VMHost -Location $clusters | Select-Object Name, ClusterId, ConnectionState, NumCpu, MemoryTotalGB
```

For an ordered startup, put approved ZStack UUIDs in `vm-order.csv` with a header of `Id`. This loop previews each target. Replace `-WhatIf` with `-Confirm` for an authorized run.

```powershell
foreach ($entry in (Import-Csv ./vm-order.csv)) {
    $vm = Get-VM -Id $entry.Id -Server $server -ErrorAction Stop
    if (-not $vm) { throw "VM not found: $($entry.Id)" }
    $vm | Start-VM -TimeoutSeconds 180 -WhatIf -ErrorAction Stop
}
```

The synchronous start returns a refreshed VM after it reaches `Running`. For planned graceful shutdown, use `Stop-ZStackVM` and verify `Stopped`. Check guest or application readiness separately when the next VM depends on it.

For snapshot age reports, select snapshots through their parent disks. The date and threshold logic can be retained.

```powershell
$cutoff = [datetimeoffset]::Now.AddDays(-7)
$vms | Get-HardDisk | Get-ZStackVolumeSnapshot |
    Where-Object { $_.CreateDate -and [datetimeoffset]::Parse($_.CreateDate) -lt $cutoff } |
    Select-Object Name, VolumeId, CreateDate
Disconnect-VIServer -Server $server
```

[Storage and network reports](inventory.md) show the primary storage and L2/L3 mappings. [Compatibility details](../../powershell/contracts/compatibility.md) list parameter and object differences.

## VM lifecycle

Use a package built from current source and UUIDs for an authorized test environment. The pipeline uses live ZStack objects and applies `-WhatIf` to each write. The create preview emits no VM, so use the UUID of a separately approved VM for the remaining previews.

```powershell
$image = Get-ZStackImage -Id '<IMAGE_UUID>' -Server $server
$network = Get-ZStackL3Network -Id '<L3_UUID>' -Server $server
$offering = Get-ZStackInstanceOffering -Id '<OFFERING_UUID>' -Server $server
New-ZStackVM -Name 'migration-qa' -Image $image -L3Network $network `
    -InstanceOffering $offering -State Stopped -Server $server -WhatIf

$vm = Get-ZStackVM -Id '<VM_UUID>' -Server $server
$vm | Set-ZStackVM -NumCpu 4 -MemoryGB 8 -WhatIf
$vm | Suspend-ZStackVM -WhatIf
$vm | Resume-ZStackVM -WhatIf
$vm | Restart-ZStackVM -WhatIf
$vm | Stop-ZStackVM -WhatIf
$vm | Remove-ZStackVM -WhatIf
$vm | Restore-ZStackVM -WhatIf
Wait-ZStackVM -Id $vm.Id -State Stopped -TimeoutSeconds 30 -Server $server
```

## Disks, snapshots, and VM devices

Use a disposable stopped VM and an existing ISO in the selected connection. Inspect the disk and the snapshot before approving a restore or deletion. Wait until the data volume is Ready before resizing it or creating a snapshot.

```powershell
$vm = Get-ZStackVM -Id $vmId -Server $server
$volume = New-ZStackVolume -Name 'test-data' -CapacityGB 1 -Server $server
$disk = $volume | Add-ZStackHardDisk -VM $vm
$disk = $disk | Set-ZStackHardDisk -CapacityGB 2
$snapshot = $disk | New-ZStackVolumeSnapshot -Name 'before-change'
$disk | Get-ZStackVolumeSnapshot
$snapshot | Restore-ZStackVolumeSnapshot -WhatIf
$snapshot | Remove-ZStackVolumeSnapshot -WhatIf
$disk | Remove-ZStackHardDisk -WhatIf

$nic = $vm | New-ZStackNetworkAdapter -L3Network $l3NetworkId
$nic | Remove-ZStackNetworkAdapter -WhatIf
$drive = $vm | Get-ZStackCDDrive -Id $driveId
$drive = $drive | Set-ZStackCDDrive -IsoImage $isoImageId
$drive | Set-ZStackCDDrive -NoMedia
```

## ZStack platform operations

Run these examples after connecting to the authorized cloud with `Connect-ZStackServer`. Save that connection in `$server`.

```powershell
$report = Get-ZStackUsage -Server $server
$report.Capacity | Select-Object Metric, Unit, Total, Used, Available, Percent |
    Export-Csv ./zstack-capacity.csv -NoTypeInformation
$report.Load | Select-Object SampleTime, ConnectedHosts, SampledHosts, CpuPercent, MemoryPercent

$sets = @(Invoke-ZStackZql -Server $server -Query 'query vmInstance.uuid,name limit 5; count vmInstance')
$sets[0].Records | ForEach-Object { [pscustomobject] $_ } | Select-Object uuid, name
$sets[1].Total

Get-ZStackSecurityGroup -Server $server -Name 'app*' | Get-ZStackSecurityGroupRule |
    Select-Object SecurityGroupId, Direction, Protocol, StartPort, EndPort, AllowedCidr
Get-ZStackVip -Server $server -Name 'app*' | Get-ZStackLoadBalancer | Get-ZStackLoadBalancerListener |
    Select-Object Name, LoadBalancerId, LoadBalancerPort, ServerGroupId
Get-ZStackProject -Server $server | Select-Object Id, Name, LinkedAccountId, State
```

Set `$jobId` to the UUID of a suspended long job you are authorized to resume. Review the preview before submitting the action.

```powershell
$job = Get-ZStackLongJob -Server $server -Id $jobId
$job | Resume-ZStackLongJob -WhatIf
$job | Resume-ZStackLongJob -Confirm | Wait-ZStackLongJob -TimeoutSeconds 600
```
