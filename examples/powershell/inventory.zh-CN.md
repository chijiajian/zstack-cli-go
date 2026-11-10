# PowerShell 资源盘点与报表

[English](inventory.md) | [简体中文](inventory.zh-CN.md) | [PowerShell 指南](../../docs/powershell.zh-CN.md)

使用当前源码构建的模块。`$server` 是 `Connect-ZStackServer` 返回的连接。以下命令查询资源信息。运行前请将名称与 UUID 换成所选环境中的资源。

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
$cluster = Get-ZStackCluster -Zone $zone
Get-ZStackVMHost -Cluster $cluster |
    Select-Object Name, ConnectionState, NumCpu, CpuTotalHz, MemoryTotalGB
Get-ZStackPrimaryStorage -Zone $zone | Select-Object Name, CapacityGB, FreeSpaceGB
Get-ZStackImageStorage -Server $server | Select-Object Name, CapacityGB, FreeSpaceGB

Get-ZStackImage -MediaType ISO -Server $server
Get-ZStackInstanceOffering -Server $server | Select-Object Name, NumCpu, MemoryGB
Get-ZStackDiskOffering -Server $server | Select-Object Name, CapacityGB
```

```powershell
$l2 = Get-ZStackL2Network -Id '<L2_UUID>' -Server $server
$l3 = Get-ZStackL3Network -L2Network $l2
Get-ZStackIPRange -L3Network $l3
Get-ZStackVM -Id '<VM_UUID>' -Detail -Server $server |
    Select-Object Id, HostName, ImageName, L3NetworkIds, Warnings
```

关联查询使用同一模块返回的实时 VM、Zone、集群或网络对象。从 CSV 或保存的对象恢复报表时，请按 ZStack UUID 重新查询。`Bytes` 结尾的字段为精确字节整数，`GB` 结尾的派生值按 GiB 换算。未报告的容量为 `$null`。主存储区分逻辑和物理容量，镜像存储、L2/L3 网络与 IP 段沿用 ZStack 的资源模型。[迁移映射](../../powershell/contracts/compatibility.md#inventory-and-reporting)
