# Decision Log

## [2026-11-09] 采用 BMad Method

**Decision:** 用户明确选择完整 PRD、architecture、epics 和 stories。
**Rationale:** 迁移语义、OEM、交付及 CLI 缺口需要统一契约。

## [2026-11-09] CLI-first 与分阶段兼容

**Decision:** 本轮方案选择复用内置 CLI，必要缺口在共享 Go 层补齐；不在 PowerShell 重建 SDK。优先发布查询、启动与等待垂直切片，不宣称全量 PowerCLI 兼容。
**Rationale:** 当前 SDK 是 Go 库，另接 .NET/REST 会引入第二套认证、锁、解析和错误模型。CLI 复用节省后端重复劳动，但不消除对象和语义适配。

## [2026-11-09] 品牌不绑定核心

**Decision:** ZStack.PowerCLI / ZStack 为自有品牌候选；OEM 从同一配置生成模块/命令/帮助。中性名称未定，不能当作已批准发行名。稳定内部对象类型为 Cloud.Automation.*。
**Rationale:** 大量 OEM 需要重复打包而非重复业务开发，兼容旧名称必须独立、显式加载。

## [2026-11-09] 将源码发现纳入前置故事

**Decision:** 新增安全 stdin 登录、start 机器输出、VM 原始内存字段前置故事。WhatIf 在模块层阻止写请求，不简单等价为 CLI dry-run。
**Rationale:** login、start、stop、get instances 的现有行为不支持直接薄包装后宣称行为等价。

## [2026-11-09] ADR 与评审边界

**Decision:** architecture.md 的 ADR-001 至 ADR-007 为本轮规划采用的设计基线；来源是用户授权形成方案，不代表用户逐条批准或完成客户认证。如实现发现需要更改 LOCKED 契约，更新规划并追加日志。
**Rationale:** 不把未决品牌、真实脚本覆盖率或发布资格伪装成已确认事实。

## [2026-11-09] 故事一致性复核

**Decision:** 在同一创建任务内补充对象身份校验、兼容参数白名单、剩余deadline及wait Observation刷新约定；24个故事完成后标记ready-for-dev。
**Rationale:** 保证实现不会信任可变对象字段、重置超时或把状态Observation当完整VM；这属于本次授权范围的规划完善，不是实现完成。

## [2026-11-09] ADR摘要索引

- ADR-001：固定版本内置CLI进程契约，避免第二套REST/SDK客户端。
- ADR-002：中性PowerShell对象投影，不伪造VMware类型。
- ADR-003：stdin密码与受限session-only配置。
- ADR-004：连接registry/代际绑定，禁止默认环境回退。
- ADR-005：ErrorRecord、ShouldProcess、未知写结果不重试。
- ADR-006：OEM品牌配置与兼容入口独立。
- ADR-007：完整单包分发、离线导入与整包回退。

**Rationale:** 每项详细Context/Decision/Consequences/Alternatives见architecture.md，均为当前规划基线。

## [2026-11-09] 启动按 UUID 严格选择

**Decision:** 模块启动适配使用 `instance start UUID --uuid-only --yes -o json`。CLI 在该模式拒绝零匹配、多个匹配或返回不同 UUID。现有名称入口仍支持按名称选择多个目标。
**Rationale:** 模块查询与写操作之间可能发生资源删除。只传 UUID 给旧 CLI 仍可回退到同名 VM，无法满足写前身份约束。

## [2026-11-09] 实现首版预览

**Decision:** 本地构建 0.1.0 自有品牌候选包和 Acme 测试包。主模块导出五个原生命令，兼容模块独立导入。Windows x64 原生测试和真实环境验收作为下一阶段门槛。
**Rationale:** 可先验证整个调用链及双品牌打包，避免将 macOS 模拟测试当作 Windows ACL、安装或客户兼容证明。
