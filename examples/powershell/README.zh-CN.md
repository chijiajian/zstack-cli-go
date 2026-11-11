# PowerShell 迁移示例

[English](README.md) | [简体中文](README.zh-CN.md) | [PowerShell 指南](../../docs/powershell.zh-CN.md) | [资源查询与报表示例](inventory.zh-CN.md)

示例假设已安装预览包，并获得对目标地址及账户的授权。OEM 包需替换模块名称与命令前缀。

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

迁移 VMware 报表时，筛选与 CSV 导出可以保留。替换连接和 VM 查询入口，预览前将 VM 标识换成已授权的 ZStack UUID。[支持的参数子集](../../powershell/contracts/compatibility.md)

## 迁移前后对照

VMware 示例中的 `$vi` 来自 `Connect-VIServer`，ZStack 示例中的 `$server` 来自上面的连接步骤。请将 `<VM_UUID>` 替换为已获管理授权的 VM UUID。启动示例用 `-WhatIf` 预览目标。确认有启动授权后再移除此参数。

| 场景 | VMware PowerCLI | ZStack PowerCLI 预览 |
| --- | --- | --- |
| 导出清单 | `Get-VM -Server $vi \| Select-Object Name, NumCpu, MemoryGB \| Export-Csv .\inventory.csv -NoTypeInformation` | `Get-ZStackVM -Server $server \| Select-Object Name, NumCpu, MemoryGB \| Export-Csv .\inventory.csv -NoTypeInformation` |
| 显式选择连接 | `Get-VM -Name 'web*' -Server $vi` | `Get-ZStackVM -Name 'web*' -Server $server` |
| 用管道传入待启动 VM | `Get-VM -Id $vSphereId -Server $vi \| Start-VM -WhatIf` | `Get-ZStackVM -Id '<VM_UUID>' -Server $server \| Start-ZStackVM -WhatIf` |
| 用 WhatIf 预览目标 | `Start-VM -VM $viVm -WhatIf` | `Start-ZStackVM -Id '<VM_UUID>' -Server $server -WhatIf` |
| 命令失败时终止脚本 | `Get-VM -Name 'web-prod' -Server $vi -ErrorAction Stop` | `Get-ZStackVM -Id '<VM_UUID>' -Server $server -ErrorAction Stop` |
| 在 OEM 会话启用旧命令名 | `Import-Module VCF.PowerCLI`，再运行 `Get-VM -Name 'web*' -Server $vi` | 在独立会话中 `Import-Module $oemCompatibilityManifest`，再运行 `Get-VM -Name 'web*' -Server $oemConnection` |

`$oemCompatibilityManifest` 指向已安装 OEM 包中的兼容清单，例如 `Compatibility/Acme.Cloud.Compatibility.psd1`。用同一包中的 `Connect-AcmeServer` 创建 `$oemConnection`。使用 OEM 原生命令时调用 `Get-AcmeVM -Server $oemConnection`，并通过 `Get-Help Get-AcmeVM -Full` 查看包信息。VM 对象来自 CSV 或旧会话时，按 UUID 重新查询。

## 公开 PowerCLI 场景

在独立会话中导入已安装的原生模块和兼容模块。使用 ZStack 连接，按环境替换 VM 名称和 UUID。

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

按顺序启动时，把已授权的 ZStack UUID 放入 `vm-order.csv`，表头为 `Id`。此循环逐台预览。实际执行获得授权后，将 `-WhatIf` 改为 `-Confirm`。

```powershell
foreach ($entry in (Import-Csv ./vm-order.csv)) {
    $vm = Get-VM -Id $entry.Id -Server $server -ErrorAction Stop
    if (-not $vm) { throw "找不到 VM $($entry.Id)" }
    $vm | Start-VM -TimeoutSeconds 180 -WhatIf -ErrorAction Stop
}
```

同步启动会在 VM 达到 `Running` 后返回刷新对象。计划优雅关机时使用 `Stop-ZStackVM`，并确认达到 `Stopped`。下一台 VM 依赖客户机或应用就绪时，使用独立探针。

快照年龄巡检通过父云盘查询卷快照，可保留按日期筛选的报表逻辑。

```powershell
$cutoff = [datetimeoffset]::Now.AddDays(-7)
$vms | Get-HardDisk | Get-ZStackVolumeSnapshot |
    Where-Object { $_.CreateDate -and [datetimeoffset]::Parse($_.CreateDate) -lt $cutoff } |
    Select-Object Name, VolumeId, CreateDate
Disconnect-VIServer -Server $server
```

[存储和网络报表](inventory.zh-CN.md)展示主存储和 L2/L3 的改写方式。[兼容矩阵](../../powershell/contracts/compatibility.md)列出了参数与对象差异。

## VM 生命周期

使用当前源码构建的模块包，并把占位符换成获准测试环境中的 UUID。管道使用实时 ZStack 对象，写命令均带 `-WhatIf`。创建预览不返回 VM，因此后续预览请指定另一台已获授权的测试 VM。

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

## 云盘、快照和 VM 设备

使用当前连接中的临时停机 VM 和已有 ISO。确认回滚或删除前先核对卷和快照。尚未实例化的数据盘需等到状态为 Ready 后才能扩容或建快照。

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

## ZStack 专有运维能力

先通过 `Connect-ZStackServer` 连接已获授权的云环境，并将连接对象保存为 `$server`。

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

把 `$jobId` 设为已获授权恢复的暂停长任务 UUID。先查看预览，再确认操作。

```powershell
$job = Get-ZStackLongJob -Server $server -Id $jobId
$job | Resume-ZStackLongJob -WhatIf
$job | Resume-ZStackLongJob -Confirm | Wait-ZStackLongJob -TimeoutSeconds 600
```
