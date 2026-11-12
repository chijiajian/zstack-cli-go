# 名称查询测试计划

[English](name-query-plan.md) | [简体中文](name-query-plan.zh-CN.md) | [文档目录](../../docs/README.zh-CN.md)

## 目标与范围

名称查询同时检查**结果**与**实际请求**。一个返回正确 VM 的测试，如果实现先读取全量再做一次精确查询，也必须失败。自动测试只使用 Go 的本机 HTTP 服务和 PowerShell 的本机模拟 API，不访问真实云，不执行云端写操作。

范围包括 CLI `get instances NAME`、PowerShell `Get-ZStackVM -Name`、`Get-ZStackVM -Name -Exact`、兼容模块 `Get-VM -Name`，以及所有公开的 `Get-* -Name` 资源命令。`-Exact` 是区分大小写的服务端精确查询；普通 `-Name` 保留不区分大小写的 PowerShell 通配符语义。

| 编号 | 场景与输入 | 必须断言 | 自动用例 |
| --- | --- | --- | --- |
| N01 | CLI 完整中文名称，含空格、方括号和 `*` | 一次请求；服务端 `q=name=...`、`limit=2`；仅返回同名资源 | `TestGetInstancesNameFiltersOnServer` |
| N02 | CLI 不存在名称、两个 VM 重名 | 空列表或两个匹配项；每次只发一条带名称条件的请求 | `TestGetInstancesNameReturnsEmptyAndDuplicateMatches` |
| N03 | VM 普通 `-Name` 的大小写和通配符 | 结果正确；请求轨迹明确记录当前本地筛选行为 | `Run-Contracts.ps1` |
| N04 | VM `-Exact`：中文、无结果、大小写不一致、名称数组、重复输入 | 每个名称都由服务端筛选；没有先前的全量读取或未命中后的全量回退；输出按 UUID 去重 | `Run-Contracts.ps1` |
| N05 | VM `-Exact` 名称本身含 `*`、`?` 与方括号 | 按普通字符查到该 VM；请求的名称条件保持原样 | `Run-Contracts.ps1` |
| N06 | 105 个 VM 同名，跨过默认 100 条分页边界 | 返回 105 个对象；两页均携带相同的名称条件 | `Run-Contracts.ps1` |
| N07 | `-Exact` 未提供名称、名称数组含空元素、服务端返回错误名称 | 无效输入不发查询；身份不符时报协议错误 | `Run-Contracts.ps1` |
| N08 | 每个公开 `Get-* -Name` 命令 | 测试表覆盖所有导出的名称选择器；大小写变化后的完整名称选中预期资源；账户样本在第 105 条 | `NameQueryContracts.ps1` |
| N09 | 兼容 `Get-VM`、跨连接对象、权限与请求失败 | 维持兼容过滤、连接边界和已有错误分类 | `MigrationContracts.ps1`、`InventoryContracts.ps1`、`Run-Contracts.ps1` |

## 执行与通过条件

1. 本地运行 `make verify` 和 `make docs-check`；它们不需要云凭据。
2. PR 的 Windows CI 在 PowerShell 7.4 与 7.6 上运行 `powershell/tests/Invoke-ContractPester.ps1`。任何新增的公开 `Get-* -Name` 命令若未加入 N08 测试表，合约测试必须失败。
3. 比较中英文使用文档与命令行为。CI 全部通过只说明本机模拟接口上的行为，不等于已验证真实云或 Gallery 安装。

## 后续风险与真实环境

普通 PowerShell `-Name` 为支持不区分大小写的通配符，仍读取全量后筛选；多个 VM UUID 的 `-Id` 查询也仍读取全量。其他资源的名称查询使用相同的本地筛选路径。这些属于后续查询规模改进，不能因 N08 的结果正确就视为性能已验证。

Gallery 包的验证在用户指定的 Windows 测试 VM 上做只读抽样：记录包版本、连接环境、名称、返回数量、状态和请求范围；限制目标，不查询全量，也不把凭据或原始响应写入报告。合并、打 tag 和发布均须用户另行明确指示。

## 2026-11-12 真实环境只读记录

在用户指定的独立测试云中，本机当前源码 CLI 用不保存密码的临时账户会话查询到 VM 总数为 3；最多取 5 条样本，再按其中一台的完整名称、最多 3 条结果查询，精确命中 1 台。不存在名称返回空列表；临时 CLI 会话已注销。

同一台 VM 在 Windows 测试机上用临时复制的 Gallery 1.1.5 模块和 PR 源码提交 `c4fadd3` 的 VM 命令验证：`Get-ZStackVM -Name NAME -Exact` 返回 1 台 Running VM（1 vCPU、1 GiB）；重复名称输入返回 1 台；不存在名称及字面 `*` 后缀均返回 0 台。PowerShell 会话已断开，临时模块副本已清理。没有资源写操作，也没有创建 AK。

该次只验证一台现有 VM 的名称路径。PR 源码通过临时模块副本运行，不代表新代码已经发布到 Gallery。

## 2026-11-12 按计划复跑与扩展实测

本地 `make verify` 通过（Go vet、未缓存的 race 测试、49 个 Python 用例、Go build）；`make docs-check` 检查了 227 条命令路径。PR 的 Windows PowerShell 7.4 和 7.6 合约测试均通过，日志中可见 N03–N09 的断言以及全部 26 个公开 `Get-* -Name` 用例。中英文 PowerShell 指南的 `-Name`、`-Exact` 和通配符说明与命令实现一致。

测试云先用 CLI 只读计数：26 个命令对应的资源类型均不超过 100 条，其中最多为角色 69 条、镜像 22 条。Windows 测试机随后用 Gallery 1.1.5 的临时副本载入 PR 中的 VM 命令：16 个有现成样本的 `Get-* -Name` 命令按大写名称返回且只返回预期 UUID；另外 10 类没有资源，缺失名称查询均返回 0 条。`Get-ZStackHardDisk` 使用现有 VM 父对象；兼容 `Get-VM -Name`、VM `-Exact`、重复名称去重以及字面 `*` 后缀也通过。最终 PowerShell 连接已注销，复查没有遗留的临时配置目录；没有创建 AK 或修改云资源。

真实云中没有中文且带通配符字符的 VM、105 台重名 VM 或故意返回错误身份的服务端，因此 N01、N05、N06、N07 的这些边界由本机模拟接口验证。10 个空资源类型只能在真实云验证无结果路径；其正向名称匹配由 CI 用例验证。前几次扩展实测的临时模块加载错误导致部分尝试未确认远端注销，复查已无本地临时配置目录；成功复跑的连接已正常注销。
