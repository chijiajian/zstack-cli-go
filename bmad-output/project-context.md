<a id="project-context"></a>
# Project Context

日期：2026-11-09。Track：BMad Method，用户已明确选择。本轮只规划，不实施、不发布、不连接云。

<a id="用户授权与目标"></a>
## 用户授权与目标

用户要求把本次 PowerCLI 迁移、PowerShell 原生体验、OEM 和内置 CLI 交付讨论形成 BMAD 文档及 user stories。客户大量来自 VMware；优先保留 Verb-Noun、参数、对象属性、管道、确认和错误处理习惯，以减少脚本改造。用户未承诺全量 PowerCLI 无修改兼容，未确认中性品牌名称或授权上线。

<a id="repository-constitution"></a>
## Repository constitution

遵守 [AGENTS.md](../AGENTS.md)、[使用指南](../docs/usage-zh.md)、[agent 指南](../docs/agents.md)。保留未提交工作；不读取 .env，不使用历史凭据或历史测试作为本次证据。开发阶段代码变更必须补回归并执行 make verify；命令变更再生成双语文档、命令参考并更新 Skill 能力检查。真实云验证另行授权。

<a id="当前证据"></a>
## 当前证据

- Go 1.25；Go SDK github.com/zstackio/zstack-sdk-go-v2 v0.0.10，见 [go.mod](../go.mod)。Go 包不能直接当作 PowerShell/.NET 模块加载。
- CLI 有配置锁、context、分页、错误退出码、等待、JSON 输出；并非每条旧命令都满足纯 JSON 契约。
- [登录](../cmd/login.go) 使用终端 ReadPassword 或参数，尚无安全非交互 stdin 密码入口；有自动续期与保存密码行为，需要适配层显式限制。
- [start](../cmd/resources/start_instance.go) 混合输出文本与结果；[stop](../cmd/resources/stop_instance.go) 写死 grace，不可直接宣称等价 VMware Stop-VM。
- [VM 查询](../cmd/get/instances.go) 的 memorySize 是格式化字符串，需要增量添加原始字节字段，不能逆解析格式化值。
- [.goreleaser.yml](../.goreleaser.yml) 支持 Windows amd64；[CI](../.github/workflows/ci.yml) 有部分 Windows pwsh 安装检查，非完整模块认证。

<a id="规划约束"></a>
## 规划约束

一套核心、不为 OEM 分叉业务；品牌与兼容层独立。首个交付范围为连接、VM 查询、启动与状态等待；关机/重启/快照/创建/删除和 VMware 专属 API 单列后续范围。首版 Windows x64 + PowerShell 7.4 作为技术最低线，发布时选择仍受微软支持的 7.x 版本验证；5.1 不在首版支持承诺。其他 OS 不因 Go 可交叉编译而自动视为已认证。

<a id="状态解释"></a>
## 状态解释

文档中所有新模块、命令、flag 均为设计，尚未实现或发布。ready-for-dev 仅表示故事契约完整，仍需满足依赖。客户脚本与发布身份缺失是客户验收/正式发布门槛，不阻断本地核心开发。
