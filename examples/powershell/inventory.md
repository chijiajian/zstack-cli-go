# PowerShell inventory and reports

[English](inventory.md) | [简体中文](inventory.zh-CN.md) | [PowerShell guide](../../docs/powershell.md)

Use the module built from current source. `$server` is the connection returned by `Connect-ZStackServer`. These commands read inventory. Replace names and UUIDs for the selected cloud.

```powershell
Get-ZStackVM -Server $server -Name 'web*' |
    Get-ZStackHardDisk |
    Select-Object VMId, Name, Type, CapacityBytes, CapacityGB |
    Export-Csv ./vm-disks.csv -NoTypeInformation

Get-ZStackVM -Server $server -Name 'web*' |
    Get-ZStackNetworkAdapter -IncludeNetworkName |
    Select-Object VMId, MacAddress, IPAddress, NetworkName

Get-ZStackVolume -Unattached -Server $server |
    Select-Object Name, CapacityBytes, ActualSizeBytes
```

```powershell
$zone = Get-ZStackZone -Id '<ZONE_UUID>' -Server $server
if (-not $zone) { throw 'Zone not found in the selected connection.' }
$clusters = @(Get-ZStackCluster -Zone $zone)
if ($clusters.Count) {
    Get-ZStackVMHost -Cluster $clusters |
        Select-Object Name, ConnectionState, NumCpu, CpuTotalHz, MemoryTotalGB
}
Get-ZStackPrimaryStorage -Zone $zone |
    Select-Object Name, CapacityBytes, FreeSpaceBytes, PhysicalCapacityBytes, SystemUsedCapacityBytes
Get-ZStackImageStorage -Server $server | Select-Object Name, CapacityGB, FreeSpaceGB

Get-ZStackImage -MediaType ISO -Server $server
Get-ZStackInstanceOffering -Server $server | Select-Object Name, NumCpu, MemoryGB
Get-ZStackDiskOffering -Server $server | Select-Object Name, CapacityGB
```

```powershell
$l2 = Get-ZStackL2Network -Id '<L2_UUID>' -Server $server
if (-not $l2) { throw 'L2 network not found in the selected connection.' }
$l3Networks = @(Get-ZStackL3Network -L2Network $l2)
if ($l3Networks.Count) {
    Get-ZStackIPRange -L3Network $l3Networks | Select-Object L3NetworkId, CIDR, StartIP, EndIP
}
Get-ZStackVM -Id '<VM_UUID>' -Detail -Server $server |
    Select-Object Id, HostName, ImageName, L3NetworkIds, Warnings
```

Use live VM, zone, cluster and network objects from the same module for related queries. Requery by ZStack UUID when starting from a CSV row or a saved object. Capacity fields ending in `Bytes` are exact reported integers. Derived `GB` fields use GiB, and omitted capacity stays `$null`. Primary storage has separate logical and physical capacities. Image storage, L2/L3 networks and IP ranges use ZStack's resource model. [Migration mappings](../../powershell/contracts/compatibility.md#inventory-and-reporting)
