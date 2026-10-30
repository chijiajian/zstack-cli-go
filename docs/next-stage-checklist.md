# 下一阶段 CLI 交付清单

范围由用户在主干完成后确认：错误契约、只读 doctor、全查询回归、IAM 管理、
有价值的任务/项目生命周期操作，以及仅使用新建测试资源的真实写 E2E。
本轮允许创建和清理自有测试资源，不修改已有业务资源，不推送或发布。

- [x] 稳定退出码和 `--error-format json`；错误与正常输出分离，结构化诊断不泄露凭据。
- [x] `doctor` 在线/离线检查；不登录、不续期、不写配置；支持 context 和超时。
- [x] 覆盖 42 类规范查询命令；global-configs 暴露稳定 id；内置告警不误判 UUID。
- [x] account/user/role/policy/project/organization 管理：SDK 支持的创建、更新、删除；密码文件、脱敏预览、确认和可取消请求。
- [x] 项目 recover/expunge 与 job resume/clean，状态及确认检查。
- [x] 写 E2E：持久化预先分配的 UUID、所有权校验、endpoint 绑定、反序清理、失败恢复。
- [x] 真实元数据生命周期：volume、account、user、policy、role、organization、project。
- [x] 真实计算/存储生命周期：新建停止状态实例、云盘挂载/解绑、快照、恢复/扩容及清理。
- [x] 全量回归、竞态检查、工作流检查、Windows 交叉编译与最终真实回归。

SDK v0.0.10 无 policy 更新方法，不虚构 `update policy`。任务操作仅操作显式选定的任务；
真实环境不恢复或清理已有任务，异常/危险状态用 mock 回归。
缺少样本的只读资源检查明确报告 skip，不视为行为验证通过。

真实验证环境已按用户要求从 172.31.29.11 切到 172.34.1.10；
旧环境缺少 project-management 许可，新环境已通过组织/项目写测试。
旧环境的本轮测试资源均已清理。


## 本轮真实验收证据（2026-10-30）

- 新环境只读回归：42 类规范查询命令，240 passed / 0 failed / 4 skipped。
  Skip 为三类空集合的后续检查和实例可选关联资源；不算作这些行为实测通过。
- doctor 在线与离线检查均通过。
- 元数据 CUD：7 类资源，创建后查询、更新后查询和反序清理全部通过。
- 计算链路：停止实例、云盘挂载/解绑、快照创建/回滚/删除、1 GiB → 2 GiB 扩容、云盘/实例恢复全部通过。
- 项目单独完成删除 → 恢复 → 删除/彻底清理验证。
- [元数据清单](validation/2026-10-30/metadata-journal.json)、[计算清单](validation/2026-10-30/compute-journal.json)、[项目恢复清单](validation/2026-10-30/project-recovery-journal.json)中资源均为 cleaned。
- [只读摘要](validation/2026-10-30/readonly-summary.json)保留具体跳过项。
- 真实环境未恢复/清理已有 long job；job 的请求、状态限制、确认和取消使用本地回归验证。
- 账号更新绕过 SDK 生成方法的错误 `/actions` 后缀，使用[官方 API 路径](https://www.zstack.io/help/dev_manual/dev_guide/v4/10.1.4.html)，并由新环境实测确认。

## 最终本地验证

- `make verify`：`go vet`、完整 Go race 回归、41 项 Python 回归、全包构建通过。
- Windows amd64 CLI 构建及配置包测试交叉编译通过；未在 Windows 上执行测试。
- 本轮 actionlint 工作流检查通过；`git diff --check` 通过。
- 未推送、发布或触发远程 CI。

## QA 复审（2026-10-30）

本轮重点审查成功判定、异常响应与失败清理；使用本地故障注入，不重复真实环境写操作。

- P1：实例创建已成功但输出失败时，原清理路径可能未记录根盘，遗漏根盘残留检查。现在先从所属实例查询并持久化根盘 UUID，再删除；根盘缺失、身份冲突或清理后仍存在均不报告成功。
- P2：恢复测试仅排除 Deleted/Destroyed，删除无效果或返回缺少状态也可能通过。现在必须观察初始状态 → 删除状态 → 初始状态的完整转换。
- P2：项目恢复预填 UUID 掩盖响应缺失字段；IAM 更新及指定 UUID 创建未核对返回身份。现在缺失或不匹配均返回 api_failed，不输出成功或重试写入。
- P2：E2E 错误 JSON 的 error/apiCode 类型异常可能抛出未分类异常、中断后续清理。现在统一转为不含原始响应的 Failure。

新增回归覆盖上述失败路径，包括无效果的删除/恢复、缺失状态、遗留根盘、错误身份和畸形错误诊断。之前的真实环境验收记录保留为历史证据，不能替代本轮新增故障测试。

QA 修复后 `make verify` 全部通过：go vet、完整 Go race 回归、44 项 Python 测试、全包构建；`git diff --check` 通过。本轮未重新执行真实写 E2E。

## Skill 真实环境只读验收（2026-10-30，172.24.254.167）

按用户提供的新环境，使用 API endpoint `172.24.254.167:8080`。本轮从独立 Skill 安装包执行本地能力检查，再按 Skill 流程通过 CLI 登录、诊断、查询和等待；没有执行资源写操作或独立 agent 对照评估。

- 独立临时配置，账户登录不保存密码；doctor 在线/离线、Stopped 实例筛选与字段投影、已满足状态的 wait 均通过。
- 未知 context 返回 not_found（退出码 4），没有回退到其他环境。
- 42 类查询回归：169 passed / 0 failed / 18 skipped；17 项缺少样本，1 项云盘可选关联不可用，主身份另行验证。跳过项不视为行为已覆盖。
- 查询前后源配置字节一致；本次临时会话注销成功，临时配置目录已删除。
- 未发现需要修改 CLI/Skill 的问题；此结果不覆盖真实创建、更新、删除、扩容或任务恢复。
- [Skill 验证记录](validation/2026-10-30/172.24.254.167/skill-report.json)含二进制与安装包摘要；[完整只读报告](validation/2026-10-30/172.24.254.167/readonly-report.json)包含逐项结果。记录不含凭据和业务清单。
