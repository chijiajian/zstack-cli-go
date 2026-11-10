# PowerShell / OEM / PowerCLI 迁移规划

[English](README.md) | [简体中文](README.zh-CN.md) | [文档目录](../docs/README.zh-CN.md)

首个切片已实现为 0.1.2 预览版。参见 [PowerShell 使用指南](../docs/powershell.zh-CN.md)和[验收记录](../powershell/release/acceptance-checklist.md)。 用户于2026-11-09选择BMad Method。首个交付切片覆盖连接、VM查询、启动/等待、OEM单包安装和有限PowerCLI风格兼容入口。

| 文档 | 内容 |
| --- | --- |
| [PRD](prd.md) | 10项功能需求、6项非功能需求及追溯 |
| [架构](architecture.md) | 7项ADR，CLI与SDK取舍、对象与进程契约 |
| [Epics与stories](epics.md) | 6个Epic、24个故事及依赖 |
| [项目约束](project-context.md) | 仓库事实、授权和边界 |
| [决策记录](decision-log.md) | 选择与理由 |
| [未决项与路线图](addendum.md) | OEM名称、客户脚本、后续能力 |
| [校验记录](validation-notes.md) | 已做规划检查与未验证项 |
| [Windows 契约证据](windows-contract-evidence-2026-11-10.md) | 原生预览包、精确构建标识和合成迁移指标 |

PowerShell 模块内置 CLI，CLI 继续复用 Go SDK。用户安装一个模块包即可使用，Import-Module 加载已安装的模块。预览包含对象、管道、错误映射、有限兼容入口和 OEM 打包。

ready-for-dev仅表示故事契约完整，不表示依赖已完成、客户验收通过或发布获授权。Windows 上的两种认证方式已通过授权的真实 API 验收。客户脚本和 OEM 正式发布身份仍需单独验收。
