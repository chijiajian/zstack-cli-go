# 错误、未知结果与清理

`-o json` 控制 stdout，`--error-format json` 控制 stderr 诊断。保存真实进程退出码；不要用管道后一个命令的退出码替代 CLI 退出码。stderr 错误 envelope 为 error，含 code/message/exitCode，可能含 httpStatus/apiCode。交互提示和 progress 不适合混入 JSON 解析。

| 退出码 | 分类 | 处理 |
| --- | --- | --- |
| 1 | operation_failed | 检查是否部分成功，不整体盲重试 |
| 2 | invalid_arguments | 核对当前 help/参数/字段 |
| 3 | authentication_failed | 核对所选 context、会话与权限，不自动扩大凭据范围 |
| 4 | not_found | 核对环境/UUID；删除后的不存在可能是预期状态 |
| 5 | conflict | 查实时状态/重名，不能任选资源 |
| 6/7 | timeout/connection_failed | 查询判断写请求是否已执行 |
| 8 | api_failed | 检查许可、状态和响应契约，避免重复写 |
| 9 | output_failed | 操作可能成功，仅输出失败 |
| 130 | canceled | 本地已取消不表示云端回滚 |

退出 0 仍需根据任务核验结果。列表空值、详情 warning、字段缺失和不匹配身份均不是完整行为验证。

创建测试资源前记录 context、endpoint、本次名称和预分配 UUID。只有确认该资源属于本次任务才清理；不要按全局名称前缀批量删除。删除后检查资源是否需 expunge，实例还要检查根盘，盘挂在原有业务实例上时不能自行解绑。

若用户要求完整 E2E，可使用匹配 CLI 版本的源码发行包中 `scripts/readonly_smoke.py` / `scripts/write_e2e.py`。这些测试脚本不随本 Skill 携带，不能假定当前目录已有它们；确认用户允许取得源码测试包或使用现有指定 checkout。普通操作不依赖这些脚本。

已存在的写 E2E journal 发生中断时，使用原脚本 `--cleanup-only --execute`，相同 endpoint 和新的有效会话可以继续清理。保留 journal、不并发复用、不编辑 state 绕过校验。未知创建结果等待 API 作业稳定后再查；若仍无法确定则报告未完成，不重复创建。只读任务不调用写 E2E。
