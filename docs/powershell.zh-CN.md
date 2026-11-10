# PowerShell 模块预览版

[English](powershell.md) | [简体中文](powershell.zh-CN.md) | [文档目录](README.zh-CN.md)

预览版面向 Windows x64 和 PowerShell 7.4 或更高版本。准确的已测试组合见[验收记录](../powershell/release/acceptance-checklist.md)。`ZStack.PowerCLI` 为示例品牌。模块包已包含 CLI，用户只需安装模块包。Go、Python 和 Pester 用于开发验证。

## 安装与加载

从提供方获取完整 ZIP 和 SHA-256。Windows 上可用[离线安装脚本](../powershell/packaging/Install-Offline.ps1)安装，默认放到当前用户的模块目录。也可以将 ZIP 解压至可信目录，通过完整路径导入模块清单。

```powershell
# 请替换为提供方给出的路径、版本和校验值。
$manifest = .\Install-Offline.ps1 -Archive .\ZStack.PowerCLI-0.1.2-windows-amd64.zip -SHA256 '<提供方校验值>'
Import-Module $manifest
```

安装至 `$env:PSModulePath` 后，可用 `Import-Module ZStack.PowerCLI`。导入读取本地文件。提供方发布到模块仓库后，从仓库获取模块是单独的安装步骤。各版本使用独立目录，切换版本时打开新会话。[升级与回退](../powershell/packaging/Upgrade-Policy.md)

安装新版时，已有连接继续使用原版本自带的 CLI。如果新版包校验失败，已安装的版本仍可使用。需要回退时，在新会话中导入旧版并重新连接。

## 迁移脚本

使用 ZStack 账户及 `host[:port]` 格式的 API 地址。参见[迁移示例](../examples/powershell/README.zh-CN.md)和[参数兼容矩阵](../powershell/contracts/compatibility.md)。

| 任务 | 原生命令 | 迁移说明 |
| --- | --- | --- |
| 连接 | `Connect-ZStackServer -Server HOST -Credential (Get-Credential)` | 账户密码连接只保存会话。有多个连接时显式传入返回的连接对象。 |
| 盘点 | `Get-ZStackVM -Name 'web*' -Server $server` | 支持 `Name`、`NumCpu`、`MemoryGB`、`PowerState` 和 `Id`。`Id` 是 ZStack UUID，`MemoryGB` 每单位为 2³⁰ 字节。 |
| 启动 | `Get-ZStackVM -Id VM_UUID -Server $server \| Start-ZStackVM -WhatIf` | 先预览。获得授权后移除 `-WhatIf`，按需加 `-Confirm`。每台 VM 的目标解析、启动请求和确认 `Running` 共用 300 秒时限。 |
| 等待 | `$vm \| Wait-ZStackVM -State Running` | 返回刷新的 VM 对象，不提供 VMware Task 对象。 |
| 断开 | `Disconnect-ZStackServer -Server $server` | 尝试远端注销并清理模块管理的本地会话。 |

希望单项错误停止管道时使用 `-ErrorAction Stop`。写操作结果未知时，先查询 VM 状态再决定是否重试。运行中的对象绑定查询时的连接。从 CSV 或序列化结果读回数据后，用显式的 `-Id` 和 `-Server` 重新选择。

从 0.1.2 起，使用已有 AK/SK 时可调用 `Connect-ZStackServer -Server HOST -AccessKeyId KEY_ID -AccessKeySecret (Read-Host 'SK' -AsSecureString)`。连接会发出一次 VM 查询来验证签名。签名依赖执行机时钟，若环境需要固定签名时区，可在该连接上指定 `-ApiTimeZone Asia/Shanghai`。模块在连接的私有临时目录中保存 AK/SK，断开连接时删除目录。

需沿用有限的旧命令名时，在独立 PowerShell 会话中从同一包导入 `Compatibility/ZStack.PowerCLI.Compatibility.psd1`。如果已加载的 VMware、Hyper-V 或用户命令占用同名入口，导入会拒绝。未支持的 VMware 参数和对象需要改写。停止、客户机正常关机、快照、`-RunAsync`、`Get-View` 与 `.ExtensionData` 不在预览范围内。

OEM 包使用提供方给定的模块名和命令前缀。不同品牌共用对象属性与脚本管道行为。

运行 `Get-Help Get-ZStackVM -Full` 或相应的 OEM 命令，可以查看品牌、模块版本、CLI 构建、连接配置命名空间和提供方帮助地址。[迁移前后对照](../examples/powershell/README.zh-CN.md#迁移前后对照)覆盖报表、显式连接、启动、WhatIf、ErrorAction 和 OEM 兼容入口。

## 当前源码中的资源查询与报表

使用匹配的 Windows x64 CLI 构建模块包后，即可使用下表的命令。每个资源输出一个 PowerShell 对象，包含 ZStack UUID（`Id`）、`ConnectionId` 和资源专有属性。空查询不输出对象。[资源盘点示例](../examples/powershell/inventory.zh-CN.md) · [VMware 迁移映射](../powershell/contracts/compatibility.md#inventory-and-reporting) · [模块打包说明](../powershell/packaging/Upgrade-Policy.md)

有多个连接时，用 `-Server` 指定其中一个。传入当前会话中查询到的父资源对象时，查询会继承其连接。显式传入的 `-Server` 必须与该连接一致。命令支持 `-Id` 时按 ZStack UUID 精确匹配，支持 `-Name` 时可使用不区分大小写的 PowerShell 通配符。

| 命令 | 资源筛选 | 常用报表字段 |
| --- | --- | --- |
| `Get-ZStackHardDisk` | `-VM` 或 VM 管道、`-Id`、`-Name` | `VMId`、`Type`、`CapacityBytes`、`ActualSizeBytes`、`CapacityGB` |
| `Get-ZStackVolume` | `-Id`、`-Name`、`-Attached` 或 `-Unattached` | `VMId`、`Attached`、`PrimaryStorageId`、`CapacityBytes`、`ActualSizeBytes` |
| `Get-ZStackNetworkAdapter` | `-VM` 或 VM 管道、`-Id`、`-IncludeNetworkName` | `VMId`、`MacAddress`、`IPAddress`、`L3NetworkId`、`NetworkName` |
| `Get-ZStackVMHost` | `-Id`、`-Name`、`-Cluster` | `ClusterId`、`ConnectionState`、`NumCpu`、`CpuTotalHz`、`MemoryTotalBytes` |
| `Get-ZStackCluster` | `-Id`、`-Name`、`-Zone` | `ZoneId`、`State`、`HypervisorType` |
| `Get-ZStackZone` | `-Id`、`-Name` | `State`、`Type`、`Description` |
| `Get-ZStackPrimaryStorage` | `-Id`、`-Name`、`-Zone` | `ZoneId`、`CapacityBytes`、`FreeSpaceBytes`、`PhysicalCapacityBytes`、`PhysicalFreeSpaceBytes`、`SystemUsedCapacityBytes` |
| `Get-ZStackImageStorage` | `-Id`、`-Name` | `Type`、`CapacityBytes`、`FreeSpaceBytes` |
| `Get-ZStackImage` | `-Id`、`-Name`、`-MediaType` | `MediaType`、`Format`、`CapacityBytes`、`ActualSizeBytes` |
| `Get-ZStackInstanceOffering` | `-Id`、`-Name` | `NumCpu`、`MemoryBytes`、`MemoryGB` |
| `Get-ZStackDiskOffering` | `-Id`、`-Name` | `CapacityBytes`、`CapacityGB`、`Type` |
| `Get-ZStackL2Network` | `-Id`、`-Name`、`-Zone` | `ZoneId`、`VirtualNetworkId`、`PhysicalInterface` |
| `Get-ZStackL3Network` | `-Id`、`-Name`、`-L2Network` | `L2NetworkId`、`IPVersion`、`DnsDomain` |
| `Get-ZStackIPRange` | `-Id`、`-Name`、`-L3Network` | `L3NetworkId`、`CIDR`、`StartIP`、`EndIP`、`Gateway` |

用 `Bytes` 字段做精确计算，`GB` 字段按 2³⁰ 字节换算，适用于展示。服务端报告的零会保留，未报告的容量为 `$null`。`SystemUsedCapacityBytes` 保留 API 返回的有符号字节值。L3 网络可查询时，`-IncludeNetworkName` 会在网卡所属连接中补充其名称。

`Get-ZStackVM -Id VM_UUID -Detail` 接受一个精确的 ZStack UUID。详情保留普通 VM 属性，并补充主机、镜像和计算规格的 ID 与名称，以及 `L3NetworkIds` 和 `Warnings`。警告会指出受影响的关联资源，并提供 `not_found` 或 `authentication_failed` 等固定失败类别。
