# Epics — PowerShell / OEM / PowerCLI migration

Track: **BMad Method**。来源：[PRD](prd.md)、[Architecture](architecture.md)、[Project Context](project-context.md)。

24个故事均已完成规划契约。done 表示 CLI 前置契约完成本地验证。review 表示已有实现或验收材料，具体原生回归结果与剩余评审项见各 story。每个故事目标为单会话2–8小时；发现过大先拆分。首版用户价值是连接、查询、启动/等待和OEM单包交付；兼容入口仅覆盖约定子集。

## Epic 1: CLI 前置契约

**Goal:** 交付CLI 前置契约的可验收切片。

**Requirements:** FR-001/002/003；NFR-001/006。[Source: prd.md#functional-requirements]

**Architecture:** [Source: architecture.md#component-design]；遵循全部ADR。

**Out of scope:** 其余Epic职责与addendum后续能力，不扩展完整PowerCLI。

| ID | User Story / intent | Blocked by | Status |
| --- | --- | --- | --- |
| 1.1 | [建立命令与脚本兼容基线](stories/1.1.compatibility-baseline.story.md) | none | done |
| 1.2 | [提供安全的 CLI 非交互密码输入](stories/1.2.password-stdin.story.md) | 1.1 | done |
| 1.3 | [让启动命令提供可靠机器输出](stories/1.3.start-json-contract.story.md) | 1.2 | done |
| 1.4 | [为VM查询补充原始内存字节](stories/1.4.vm-memory-bytes.story.md) | 1.3 | done |

**Cross-epic dependencies:** none。反向依赖在各story的Blocks字段中。

## Epic 2: 原生模块基础

**Goal:** 交付原生模块基础的可验收切片。

**Requirements:** FR-002/003；NFR-001/002/005。[Source: prd.md#functional-requirements]

**Architecture:** [Source: architecture.md#component-design]；遵循全部ADR。

**Out of scope:** 其余Epic职责与addendum后续能力，不扩展完整PowerCLI。

| ID | User Story / intent | Blocked by | Status |
| --- | --- | --- | --- |
| 2.1 | [装载版本化原生模块骨架](stories/2.1.module-loader.story.md) | 1.1 | review |
| 2.2 | [建立结构化CLI子进程适配器](stories/2.2.process-adapter.story.md) | 1.2, 2.1 | review |
| 2.3 | [映射CLI错误到PowerShell错误流](stories/2.3.error-records.story.md) | 2.2 | review |
| 2.4 | [维护独立且绑定身份的连接](stories/2.4.connection-registry.story.md) | 2.3 | review |

**Cross-epic dependencies:** Epic 1。反向依赖在各story的Blocks字段中。

## Epic 3: VM 查询与对象

**Goal:** 交付VM 查询与对象的可验收切片。

**Requirements:** FR-004；NFR-002/003/006。[Source: prd.md#functional-requirements]

**Architecture:** [Source: architecture.md#component-design]；遵循全部ADR。

**Out of scope:** 其余Epic职责与addendum后续能力，不扩展完整PowerCLI。

| ID | User Story / intent | Blocked by | Status |
| --- | --- | --- | --- |
| 3.1 | [生成兼容报表习惯的 VM 对象](stories/3.1.vm-projection.story.md) | 1.4, 2.3, 2.4 | review |
| 3.2 | [提供分页与名称筛选的原生查询](stories/3.2.vm-query.story.md) | 3.1 | review |
| 3.3 | [为写操作解析唯一资源与原连接](stories/3.3.vm-selection.story.md) | 3.2 | review |
| 3.4 | [验证并完成查询对象的 PowerShell 管道体验](stories/3.4.query-pipeline.story.md) | 3.3 | review |

**Cross-epic dependencies:** Epic 1, Epic 2。反向依赖在各story的Blocks字段中。

## Epic 4: 启动与兼容入口

**Goal:** 交付启动与兼容入口的可验收切片。

**Requirements:** FR-005/006；NFR-002/006。[Source: prd.md#functional-requirements]

**Architecture:** [Source: architecture.md#component-design]；遵循全部ADR。

**Out of scope:** 其余Epic职责与addendum后续能力，不扩展完整PowerCLI。

| ID | User Story / intent | Blocked by | Status |
| --- | --- | --- | --- |
| 4.1 | [统一逐目标确认与 WhatIf 边界](stories/4.1.should-process.story.md) | 3.3 | review |
| 4.3 | [提供状态等待与可诊断超时](stories/4.3.wait-vm.story.md) | 2.4, 3.1 | review |
| 4.2 | [用原生对象管道启动并等待 VM](stories/4.2.start-vm.story.md) | 1.3, 4.1, 4.3 | review |
| 4.4 | [提供显式加载的 PowerCLI 风格子集](stories/4.4.compatibility-module.story.md) | 1.1, 3.4, 4.2 | review |

**Cross-epic dependencies:** Epic 1, Epic 2, Epic 3。反向依赖在各story的Blocks字段中。

## Epic 5: OEM 与安装

**Goal:** 交付OEM 与安装的可验收切片。

**Requirements:** FR-007/008；NFR-004/005。[Source: prd.md#functional-requirements]

**Architecture:** [Source: architecture.md#component-design]；遵循全部ADR。

**Out of scope:** 其余Epic职责与addendum后续能力，不扩展完整PowerCLI。

| ID | User Story / intent | Blocked by | Status |
| --- | --- | --- | --- |
| 5.1 | [定义OEM品牌配置与校验](stories/5.1.brand-schema.story.md) | 2.1 | review |
| 5.2 | [生成内置CLI的品牌发行包](stories/5.2.bundled-package.story.md) | 5.1, 2.1, 1.4 | review |
| 5.3 | [统一OEM帮助和错误展示](stories/5.3.branded-diagnostics.story.md) | 5.1, 2.3 | review |
| 5.4 | [完成离线安装与整包升级回退](stories/5.4.offline-upgrade.story.md) | 5.2, 2.4 | review |

**Cross-epic dependencies:** Epic 1, Epic 2。反向依赖在各story的Blocks字段中。

## Epic 6: 验证与迁移交付

**Goal:** 交付验证与迁移交付的可验收切片。

**Requirements:** FR-001/009/010；NFR-001至006。[Source: prd.md#functional-requirements]

**Architecture:** [Source: architecture.md#component-design]；遵循全部ADR。

**Out of scope:** 其余Epic职责与addendum后续能力，不扩展完整PowerCLI。

| ID | User Story / intent | Blocked by | Status |
| --- | --- | --- | --- |
| 6.1 | [交付双语安装与迁移说明](stories/6.1.bilingual-guides.story.md) | 4.4, 5.4, 5.3 | review |
| 6.2 | [建立Windows原生模块契约回归](stories/6.2.windows-contract-ci.story.md) | 3.4, 4.4, 5.3, 5.4 | review |
| 6.3 | [建立脚本迁移样本与改造量报告](stories/6.3.migration-corpus.story.md) | 1.1, 4.4, 6.1 | review |
| 6.4 | [形成可审查的试点与发布证据包](stories/6.4.release-evidence.story.md) | 6.1, 6.2, 6.3 | review |

**Cross-epic dependencies:** Epic 1, Epic 3, Epic 4, Epic 5。反向依赖在各story的Blocks字段中。

## Execution order and scope constraints

保守串行顺序（不是已排期承诺；无需并行工具即可执行）：

1.1 → 1.2 → 1.3 → 1.4 → 2.1 → 2.2 → 2.3 → 2.4 → 3.1 → 3.2 → 3.3 → 3.4 → 4.1 → 4.3 → 4.2 → 4.4 → 5.1 → 5.2 → 5.3 → 5.4 → 6.1 → 6.2 → 6.3 → 6.4。

共享文件必须串行：1.2 → 1.3 → 1.4 共用双语usage与生成参考；3.2 → 3.3 → 3.4保护共享查询文件。全局scope checker会报告这些有效重叠；依赖已明确序列化，不宣称所有故事可并行。

4.3等待先于4.2启动，避免启动自行复制轮询。6.4生成验收模板与no-go判定，不代表真实环境验收完成。进一步调度时由bmad-sprint-planning生成sprint-status.yaml；本轮未进行开发调度或handoff发布。

## Delivery Tracking

- Total: 24
- review: 20
- done: 4
- Remaining review: 20
- 4 个 CLI 前置故事完成，20 个故事进入评审。Story 6.2 的双版本 Windows 原生模拟套件和 0.1.2 的双认证真实 API 验收均已通过。客户迁移脚本验收待办。
