# AI agent 执行指南

这份文档面向使用 CLI 的 agent；仓库开发约定见 [AGENTS.md](../AGENTS.md)。它不授予环境访问或写操作权限，执行范围以用户指令为准。

## 命令发现

1. 用 `zstack-cli version` 确认二进制。源码中的功能可能尚未发布。
2. 读取 [commands.json](reference/commands.json) 获取全部命令路径、类别及帮助，再以当前二进制的 `<command> --help` 核对参数。不要猜测 SDK 方法是否已暴露为 CLI 命令。
3. 从[使用指南](usage-zh.md)和 [examples](../examples/README.md)选择工作流。示例名称和 UUID 均是占位数据。

JSON 索引的 category 仅用于发现：`read`、`mutation`、`authentication`、`local-config`、`discovery`。不能仅凭 category 自动授权执行。比如 config view 是本地读取，而 config use-context 会修改配置；get 在保存密码时可能触发会话续期。

## 执行协议

- 确认用户给定的 endpoint、context 和操作范围。每条云命令显式传 `--context`，或使用任务专用 `ZSTACK_CONFIG`。不要借用仓库内发现的凭据，也不要将生产环境作为默认测试环境。
- 只读任务使用不保存密码的已有会话或 AK/SK；`doctor --offline` 不联网，在线 doctor 不登录或续期。过期会话应报告认证失败，不能为“只读验证”自动创建其他账号或密钥。
- 查询使用 `-o json --error-format json`，将 stdout 和 stderr 分开，检查退出码、JSON 结构、资源身份以及详情的 warnings。分页使用稳定排序；分页结果不是快照，缺少样本应报告未覆盖。
- 单资源操作优先使用确认过的 UUID。重名、未知 context、缺少许可、意外的关联资源都应作为执行障碍，不擅自扩大范围。
- 写操作仅在用户授权后执行。有 dry-run 时先检查预览，但不要假定它都离线：IAM/job/recover/resize 的预览离线，部分旧 delete/expunge 预览会查询服务器。dry-run 不证明真实执行一定成功。
- `--yes` 仅用于已经明确授权和确认目标的操作。expunge 不可恢复；resume job、snapshot revert、全局配置修改、基础设施变更均是写操作。
- 给请求/等待设置合适 deadline，使用命令实际支持的 `--request-timeout` 或 `--timeout`，不要把两者混用。不同旧 SDK 写路径的取消能力不完全一致。
- 写操作失败、超时或输出异常后，先查询已知 UUID。非零退出码不表示服务器没有执行。不要无条件重试创建、扩容、恢复或删除。

## E2E 与恢复

优先使用仓库的 `scripts/readonly_smoke.py` 和 `scripts/write_e2e.py`，不要自行拼接批量删除来清理测试环境。写 E2E 在每次请求前记录本次生成的 UUID/name；使用不同 journal，不能并发复用。

出现失败时检查报告、保留 journal，等异步请求稳定后用 `--cleanup-only --execute` 恢复清理。只有通过所有权核对的本次资源才能被删除。根盘残留、错误身份、未知创建结果都应保持未完成；不修改 journal 的状态来制造通过。报告不得包含密码、AK/SK、会话或原始敏感响应。

交付时分开报告：实现/修复内容、本地检查、实际云端验证、跳过/未验证项、清理结果、剩余风险。历史通过记录不能当作本次执行证据。对于新增功能，只有 help 存在不能算回归通过。

## 可安装 Skill 与 MCP

仓库现已提供 [zstack-operations Skill](../skills/README.md)，作为与 CLI 同仓维护、可独立安装的操作入口。它直接指导 agent 调用 CLI，支持查询、生命周期、IAM 和失败恢复，不依赖本仓库文档路径。

仓库内开发使用 AGENTS.md；已安装 CLI 的运维 agent 使用 Skill。MCP 是按客户端接入、远程执行或集中管控需求增加的可选适配层，不是 Skill 调用 CLI 的前置条件。
