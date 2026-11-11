# PowerShell 迁移示例

[English](README.md) | [简体中文](README.zh-CN.md) | [PowerShell 指南](../../docs/powershell.zh-CN.md) | [资源查询与报表示例](inventory.zh-CN.md)

示例假设已安装预览包，并获得对目标地址及账户的授权。OEM 包需替换模块名称与命令前缀。

```powershell
Import-Module ZStack.PowerCLI
$server = Connect-ZStackServer -Server '10.255.255.21:8080' -Credential (Get-Credential)
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
