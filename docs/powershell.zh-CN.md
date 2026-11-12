# ZStack PowerCLI PowerShell 模块

[English](powershell.md) | [简体中文](powershell.zh-CN.md) | [文档目录](README.zh-CN.md)

[`ZStack.PowerCLI`](https://www.powershellgallery.com/packages/ZStack.PowerCLI) 已发布到 PowerShell Gallery，适用于 Windows x64 和 PowerShell 7.4 或更高版本。1.1.4 是首个 Gallery 版本。模块包已包含 Windows CLI，用户只需安装模块。Go、Python 和 Pester 用于开发验证。准确的已测试组合及仍待完成的客户验证见[验收记录](../powershell/release/acceptance-checklist.md)。

## 安装与加载

在 Windows 上使用 PowerShell 7.4 或更高版本，从 `PSGallery` 安装已发布模块：

```powershell
Find-PSResource -Name ZStack.PowerCLI -Repository PSGallery
Install-PSResource -Name ZStack.PowerCLI -Repository PSGallery -Scope CurrentUser
Import-Module ZStack.PowerCLI
Get-Module ZStack.PowerCLI
```

若要固定首个发布版本，在 `Install-PSResource` 命令中添加 `-Version 1.1.4`，并使用 `Import-Module ZStack.PowerCLI -RequiredVersion 1.1.4`。今后安装已发布的新版本可执行 `Update-PSResource -Name ZStack.PowerCLI -Repository PSGallery`。切换已安装版本前请打开新会话。[升级与回退](../powershell/packaging/Upgrade-Policy.md)

对于提供方自行构建的离线包或 OEM 包，先从该提供方获取 ZIP 和 SHA-256，再使用[离线安装脚本](../powershell/packaging/Install-Offline.ps1)。这些 ZIP 与 Gallery 模块包及 GitHub Releases 中的 CLI 产物不同。安装新版时，已有连接继续使用原版本自带的 CLI。回退时，在新会话中导入旧版并重新连接。

## 模块仓库与发布

安装或导入 `ZStack.PowerCLI` 无需注册模块本身。`PSGallery` 是 PowerShell 客户端常见的预置模块仓库；“默认仓库”只表示未指定 `-Repository` 时命令按客户端配置选择查找来源，不代表微软为模块背书。用 `Get-PSResourceRepository` 查看当前配置；若 PSGallery 已被移除，可用 `Register-PSResourceRepository -PSGallery` 恢复。企业私有仓库须在客户端按其地址注册；安装时显式指定 `-Repository` 可避免选错来源。

已发布的 [1.1.4 模块包](https://www.powershellgallery.com/packages/ZStack.PowerCLI/1.1.4) 标注作者为 `ZStack`、许可证为 Apache 2.0，项目地址指向当前组织仓库。模块清单还包含 Windows/PowerShell Core 标签。[发布与验收状态](../powershell/release/acceptance-checklist.md)

[tag 发布流程](../.github/workflows/release.yml)会在 Go 检查、Windows PowerShell 契约测试和 GitHub Release 成功后，将 `ZStack.PowerCLI` 发布至 PSGallery。只接受稳定版 `v主版本.次版本.修订号` 标签；带 `-` 的预发布标签跳过 Gallery 任务。模块版本跟随 tag。Gallery API Key 保存在仓库 **Actions Secret** `PSGALLERY_API_KEY` 中，不放在普通 Variable。请在当前 Key 过期前创建替代 Key 并更新 Secret；Key 缺失或过期会使今后的 Gallery 发布失败。

## 迁移脚本

使用 ZStack 账户及 `host[:port]` 格式的 API 地址。参见[迁移示例](../examples/powershell/README.zh-CN.md)和[参数兼容矩阵](../powershell/contracts/compatibility.md)。

| 任务 | 原生命令 | 迁移说明 |
| --- | --- | --- |
| 连接 | `Connect-ZStackServer -Server HOST -Credential (Get-Credential)` | 账户密码连接只保存会话。有多个连接时显式传入返回的连接对象。 |
| 盘点 | `Get-ZStackVM -Name 'web-01' -Exact -Server $server` | `-Exact` 由服务端按大小写精确筛选；不加 `-Exact` 时读取全部 VM，再在本地不区分大小写地匹配，支持 PowerShell 通配符。输出含 `Name`、`NumCpu`、`MemoryGB`、`PowerState` 和 `Id`。`Id` 是 ZStack UUID，`MemoryGB` 每单位为 2³⁰ 字节。 |
| 启动 | `Get-ZStackVM -Id VM_UUID -Server $server \| Start-ZStackVM -WhatIf` | 先预览。获得授权后移除 `-WhatIf`，按需加 `-Confirm`。每台 VM 的目标解析、启动请求和确认 `Running` 共用时限，默认 300 秒，可用 `-TimeoutSeconds` 设置。 |
| 等待 | `$vm \| Wait-ZStackVM -State Running` | 返回刷新的 VM 对象，不提供 VMware Task 对象。 |
| 断开 | `Disconnect-ZStackServer -Server $server` | 尝试远端注销并清理模块管理的本地会话。 |

希望单项错误停止管道时使用 `-ErrorAction Stop`。写操作结果未知时，先查询 VM 状态再决定是否重试。运行中的对象绑定查询时的连接。从 CSV 或序列化结果读回数据后，用显式的 `-Id` 和 `-Server` 重新选择。

从 0.1.2 起，使用已有 AK/SK 时可调用 `Connect-ZStackServer -Server HOST -AccessKeyId KEY_ID -AccessKeySecret (Read-Host 'SK' -AsSecureString)`。连接会发出一次 VM 查询来验证签名。签名依赖执行机时钟，若环境需要固定签名时区，可在该连接上指定 `-ApiTimeZone Asia/Shanghai`。模块在连接的私有临时目录中保存 AK/SK，断开连接时删除目录。

需沿用旧命令名时，在独立 PowerShell 会话中从同一 ZStack 包导入 `Compatibility/ZStack.PowerCLI.Compatibility.psd1`。扩展包提供 `Connect-VIServer`、`Disconnect-VIServer`、`Get-VM`、`Get-VMHost`、`Get-Cluster`、`Get-HardDisk`、`Get-NetworkAdapter` 和 `Start-VM` 的有限参数子集。可保留 `Get-VM | Get-HardDisk`、`Get-VM | Get-NetworkAdapter` 和常用 CSV 选列。如果已加载的 VMware、Hyper-V 或用户命令占用同名入口，导入会拒绝。[公开场景迁移示例](../examples/powershell/README.zh-CN.md#公开-powercli-场景)

沿用 `Start-VM` 的脚本可用 `-TimeoutSeconds` 为每台 VM 设置解析、请求和等待的共享时限。关机脚本可迁移到 `Stop-ZStackVM` 的 ACPI 优雅关机和 `Wait-ZStackVM` 的平台状态等待。依赖客户机或应用就绪时，使用独立探针。整机快照、`-Kill`、`-RunAsync`、`Get-View`、`.ExtensionData` 与 VMware Task 需要单独改写。卷快照按云盘查询和操作。

OEM 包使用提供方给定的模块名和命令前缀。不同品牌共用对象属性与脚本管道行为。

运行 `Get-Help Get-ZStackVM -Full` 或相应的 OEM 命令，可以查看品牌、模块版本、CLI 构建、连接配置命名空间和提供方帮助地址。[迁移前后对照](../examples/powershell/README.zh-CN.md#迁移前后对照)覆盖报表、显式连接、启动、WhatIf、ErrorAction 和 OEM 兼容入口。

## 当前源码中的资源查询与报表

使用匹配的 Windows x64 CLI 构建模块包后，即可使用下表的命令。每个资源输出一个 PowerShell 对象，包含 ZStack UUID（`Id`）、`ConnectionId` 和资源专有属性。空查询不输出对象。[资源盘点示例](../examples/powershell/inventory.zh-CN.md) · [VMware 迁移映射](../powershell/contracts/compatibility.md#inventory-and-reporting) · [模块打包说明](../powershell/packaging/Upgrade-Policy.md)

有多个连接时，用 `-Server` 指定其中一个。传入当前会话中查询到的父资源对象时，查询会继承其连接。显式传入的 `-Server` 必须与该连接一致。命令支持 `-Id` 时按 ZStack UUID 精确匹配。`Get-ZStackVM -Name NAME -Exact` 接受完整名称，由服务端按大小写精确筛选，支持名称数组；名称中的 `*`、`?` 和方括号按普通字符处理。不加 `-Exact` 的 VM 名称查询会读取全部 VM 后在本地不区分大小写地匹配，包括 PowerShell 通配符。

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

## 当前源码中的 VM 生命周期

写命令接受同一模块返回的实时 VM 对象或精确的 ZStack `-Id`，并逐目标支持 `-WhatIf` 和 `-Confirm`。有多个连接时，按 UUID 操作需显式指定 `-Server`。成功后返回刷新的 VM 对象。默认每台 VM 有 300 秒预算，覆盖目标解析、API 请求与状态确认。使用 `-ErrorAction Stop` 可在首个错误时终止批处理。

| 命令 | 输入与结果 |
| --- | --- |
| `New-ZStackVM` | 通过 `-Image` 选择根卷模板，通过 `-L3Network` 选择至少一个网络。使用 `-InstanceOffering`，或同时指定 `-NumCpu` 和 `-MemoryGB`。显式选择 `-State Stopped` 或 `-State Running`。部署参数接受实时对象或 UUID。 |
| `Set-ZStackVM` | 只修改显式传入的 `-Name`、`-Description`、`-NumCpu` 或 `-MemoryGB`。空描述可清除原值。CPU 和内存变更要求 `Stopped`。 |
| `Suspend-ZStackVM` 与 `Resume-ZStackVM` | 将 `Running` 的 VM 暂停在内存中，并将 `Paused` 的 VM 恢复到 `Running`。已处于目标状态时只刷新对象。[ZStack VM 生命周期](https://docs.zstack.io/en/zcf/zstack-cloud/4.8.38/dev-guide?section=ZStack_2.1_Dev_Guide_0002&version=4.8.38) |
| `Stop-ZStackVM` | 对 `Running` 的 VM 请求 ACPI 优雅关机并确认 `Stopped`。`-StopHA` 显式请求平台的 HA 停机行为。[StopVmInstance](https://docs.zstack.io/docs/api-explorer/zh-cn/zsphere/explorer/StopVmInstance) |
| `Restart-ZStackVM` | 对 `Running` 的 VM 提交一次重启，等待 API 完成后确认 `Running`。工作流需要证明客户机已重启时，请比较操作前后的客户机启动时间或启动 ID。[RebootVmInstance](https://docs.zstack.io/en/zcf/zstack-cloud/5.5.30/dev-guide?section=ZStack_2.1_Dev_Guide_0046&version=5.5.30) |
| `Remove-ZStackVM` 与 `Restore-ZStackVM` | 以高确认级别删除已停止的 VM 并确认 `Destroyed`。删除要求可读取并核实 VM 删除策略为 `Delay` 或 `Never`。通过 VM 恢复回收站内可恢复的实例及其根盘。恢复后的状态为 `Stopped`。保留期限与附着数据盘行为由平台策略决定。[删除策略](https://docs.zstack.io/zh/docs/zcf/zstack-cloud/user-guide?section=ZStack_User_Guide_0043_5) |
| `Wait-ZStackVM` | 等待 `Running`（默认）、`Stopped`、`Paused` 或 `Destroyed`。等待其他状态时若观察到 `Destroyed`，立即返回冲突。 |

生命周期判断使用 `RawState`。删除或恢复后，后续操作应使用返回的对象或按 UUID 重新查询。写操作报告 `OutcomeUnknown` 时，先查询 VM 再决定是否重试。创建流程可预先指定 `-ResourceId`，也可使用错误目标中的 UUID 核对结果。[生命周期示例](../examples/powershell/README.zh-CN.md#vm-生命周期) · [迁移映射](../powershell/contracts/compatibility.md#vm-lifecycle)

## 云盘、卷快照与 VM 设备

使用从当前源码构建且包含匹配 CLI 的包。弹出介质要求该 ISO 在 VM 上只关联选定的一台光驱。

| 命令 | 行为 |
| --- | --- |
| `New-ZStackVolume`、`Remove-ZStackVolume`、`Restore-ZStackVolume` | 按容量或云盘规格创建独立数据盘。在 Delay 或 Never 策略下把未挂载的数据盘删除至回收站，并按原 UUID 恢复。 |
| `Add-ZStackHardDisk`、`Remove-ZStackHardDisk`、`Set-ZStackHardDisk` | 挂载现有数据盘，卸载并保留底层数据盘，或将虚拟盘扩至指定总容量 `-CapacityGB`。已挂载的盘要求 VM 停机后扩容，客户机分区由客户机自行处理。 |
| `Get-ZStackVolumeSnapshot`、`New-ZStackVolumeSnapshot`、`Restore-ZStackVolumeSnapshot`、`Remove-ZStackVolumeSnapshot` | 查询、创建、回滚和删除单个云盘的快照。已挂载卷要求 VM 停机后回滚。仅允许删除叶子快照。 |
| `New-ZStackNetworkAdapter`、`Remove-ZStackNetworkAdapter` | 在选定 L3 接入网卡，按网卡 UUID 移除。保留同一 L3 上的其他网卡。 |
| `Get-ZStackCDDrive`、`Set-ZStackCDDrive` | 查询虚拟光驱及其 ISO UUID。用 `-IsoImage` 插入已有 ISO，用 `-NoMedia` 弹出所选光驱的介质。 |

写命令逐目标支持 `-WhatIf`、`-Confirm` 和 `-TimeoutSeconds`。按命令帮助传入当前连接的资源对象或精确 UUID。用 UUID 操作时，按参数要求同时指定预期 VM 或父卷。结果未知时先查询目标再决定是否重试。见 [PowerShell 示例](../examples/powershell/README.zh-CN.md#云盘快照和-vm-设备)。

创建独立数据盘会读回实际容量，以及请求中明确指定的云盘规格和主存储。接入网卡会读回显式指定的 MAC 和静态 IP。若无法核实这些值，命令返回 `OutcomeUnknown`。此时先查询目标资源，再决定下一步。快照回滚完成后可在客户机内检查文件或块数据，确认恢复了预期内容。

## ZStack 专有运维能力

| 工作流 | Cmdlet 与输出 |
| --- | --- |
| 容量与负载 | `Get-ZStackUsage [-Section All\|Capacity\|Load]` 返回指定连接的报表。`Capacity` 行包含 `Metric`、`Unit`、`Total`、`Used`、`Available` 和 `Percent`。CPU 容量单位为 vCPU，内存、主存储逻辑和物理容量单位为 bytes。`Load` 包含最近一次 CPU 和内存共同采样的时间、百分比、已采样和已连接的主机数量。缺失的采样和百分比为 `$null`，未请求的报表分区也为 `$null`。 |
| ZQL | `Invoke-ZStackZql -Query <string>` 接收 ZQL 查询和计数，支持聚合与多个结果集。每个对象包含从 0 开始的 `Index`、保留所选列的 `Records`、`Total`、`Counts`、`Aggregates`、`ReturnWith`，以及 `Data` 中的完整结果集。云端 API 负责解析语句，并执行当前账号的权限检查。 |
| 云端长任务 | `Get-ZStackLongJob [-Id <id>]`、`Wait-ZStackLongJob`、`Resume-ZStackLongJob` 使用真实云端任务 ID。查询对象包含 `JobId`、`Name`、`State`、`JobName`、`TargetResourceId`，以及平台返回的 `JobResult`。API 用状态报告任务进展。 |
| IAM 审计 | `Get-ZStackAccount`、`Get-ZStackUser`、`Get-ZStackRole`、`Get-ZStackPolicy`、`Get-ZStackProject` 查询当前连接可见的资源。用户和策略对象包含 `AccountId`，IAM2 项目包含 `LinkedAccountId`，策略包含 `StatementCount`。 |
| 网络审计 | `Get-ZStackSecurityGroup`、`Get-ZStackSecurityGroupRule`、`Get-ZStackVip`、`Get-ZStackLoadBalancer`、`Get-ZStackLoadBalancerListener`、`Get-ZStackLoadBalancerServerGroup` 返回通过资源 ID 关联的对象。管道支持从安全组查规则、从 VIP 查负载均衡器、从负载均衡器查监听器和后端服务器组。 |

`Wait-ZStackLongJob` 默认等待 `Succeeded`。默认等待遇到失败或取消的云端任务时，分别返回 `JobFailed` 或 `JobCanceled`。显式指定 `-State Failed` 或 `-State Canceled` 可以读取相应终态。等待超时后可用同一个 ID 继续查询云端状态。换 PowerShell 会话后，重新连接同一环境，再用 `-Id` 传入保存的 ID。

`Resume-ZStackLongJob` 仅恢复指定的已暂停任务，支持 `-WhatIf` 和确认。等待和恢复都在命令超时预算内按 UUID 读回长任务，返回与 `Get-ZStackLongJob` 相同的字段。恢复请求发出后若任务仍为 `Suspended` 或状态无法核实，命令返回 `OutcomeUnknown`，应先查询任务再决定是否重试。通用 `-RunAsync` 需要原操作返回能够跨连接重新查询的真实长任务 ID。普通 API 完成轮询和 VM 状态等待各自沿用原有契约。

盘点命令在指定连接中分页。空结果不输出对象，权限拒绝返回 `AuthenticationFailed`。关联查询接收同一连接的实时父对象。见 [PowerShell 专有运维示例](../examples/powershell/README.zh-CN.md#zstack-专有运维能力)。
