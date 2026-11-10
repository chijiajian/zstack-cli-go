# 文档目录

[English](README.md) | [简体中文](README.zh-CN.md)

| 使用场景 | English | 简体中文 |
| --- | --- | --- |
| 安装与快速开始 | [README](../README.md) | [README](../README.zh-CN.md) |
| 登录、context、查询、IAM 与生命周期 | [User manual](usage-en.md) | [使用手册](usage-zh.md) |
| Agent 执行与失败恢复 | [Agent guide](agents.en.md) | [Agent 指南](agents.md) |
| 各操作系统与 agent 的 Skill 安装 | [Skill guide](../skills/README.en.md) | [Skill 安装](../skills/README.md) |
| 操作示例与输入文件 | [Examples](../examples/README.en.md) | [操作示例](../examples/README.md) |

## 自动生成的命令参考

[命令参考](reference/commands.md)和 [JSON 索引](reference/commands.json)保留 CLI 英文帮助输出。运行 `make docs` 生成，用 `make docs-check` 核对一致性。双语手册负责解释用法和组合流程。

## 开发历史与验收记录

本目录的迁移计划和完成清单记录实现决策及特定日期的验收结果。日常操作请从使用手册进入。累计验收证据见 [next-stage-checklist.md](next-stage-checklist.md)。历史文档与原始验证摘要保留原语言。

## 双语维护规则

同一改动同时更新中英文版本，保持命令、默认值和操作条件一致。每篇维护中的指南都链接到另一语言。文档改动后检查本地链接，并与 CLI 帮助核对示例。现有中文指南路径继续保留，保证旧链接可用。

## PowerShell 模块

[PowerShell 模块预览](powershell.zh-CN.md)

## 功能规划

[PowerShell / OEM / PowerCLI迁移规划与故事状态](../bmad-output/README.zh-CN.md)
