# CLI examples

这些文件是流程示例，不是可一键执行的云端测试。先按[使用指南](../docs/usage-zh.md)登录指定 context，逐条替换环境、名称、UUID 和路径；写操作仅用于已经授权的目标。

| 场景 | 文件 |
| --- | --- |
| 查询、分页、诊断、多 context | [查询与自动化](queries.md) |
| IAM 与输入文件 | [IAM](iam.md)、[策略 JSON](policy-statements.json) |
| 实例、盘、快照、任务生命周期 | [生命周期](lifecycle.md) |
| 网络与基础设施操作 | [网络和基础设施](infrastructure.md) |
| 只读/写回归与中断清理 | [回归运行](regression.md) |
| 任意命令的用法、别名、全部参数 | [完整命令参考](../docs/reference/commands.md) |

命令参考自动覆盖当前全部可见命令；这些 examples 覆盖主要组合流程，不代表每种服务器配置、许可、参数组合都有实测样例。优先使用命令支持的 dry-run 检查请求，再确认写入范围。

维护文档时运行 `make docs`，用 `make docs-check` 检查生成内容是否与当前命令树一致。该过程仅运行 `--help`，不读取云配置或调用 API。
