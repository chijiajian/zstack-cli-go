# PowerShell 迁移示例

[English](README.md) | [简体中文](README.zh-CN.md) | [PowerShell 指南](../../docs/powershell.zh-CN.md)

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
