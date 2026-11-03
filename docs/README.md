# Documentation

[English](README.md) | [简体中文](README.zh-CN.md)

| Task | English | 简体中文 |
| --- | --- | --- |
| Installation and quick start | [README](../README.md) | [README](../README.zh-CN.md) |
| Login, contexts, query, IAM and lifecycle | [User manual](usage-en.md) | [使用手册](usage-zh.md) |
| Agent execution and recovery | [Agent guide](agents.en.md) | [Agent 指南](agents.md) |
| Skill installation for each agent and OS | [Skill guide](../skills/README.en.md) | [Skill 安装](../skills/README.md) |
| Example workflows and input files | [Examples](../examples/README.en.md) | [操作示例](../examples/README.md) |

## Generated reference

[Command reference](reference/commands.md) and [JSON index](reference/commands.json) preserve the English output of CLI help. Run `make docs` to regenerate and `make docs-check` to check consistency. Human guides provide bilingual explanations and workflows.

## Development history

The migration plans and completion checklists in this directory record implementation decisions and dated validation evidence. Start with the user manual for current operation. The latest accumulated evidence is in [next-stage-checklist.md](next-stage-checklist.md). Historical documents and raw validation summaries retain their original language.

## Maintaining translations

Update both language versions in the same change. Keep executable commands, default values and safety conditions consistent. Each maintained guide has a link to its counterpart. Run local link checks and compare examples against CLI help after changing documents. Preserve existing Chinese guide paths so current links remain valid.
