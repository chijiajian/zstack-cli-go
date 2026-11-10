# Product Requirements Document (PRD)

**Project Name:** PowerShell 自动化与 OEM / PowerCLI 迁移支持
**Version:** 1.0
**Date:** 2026-11-09
**Status:** Planning baseline — 用户授权形成方案，未完成客户验收
**Track:** BMad Method

## Executive Summary

大量客户来自 VMware，已有 PowerShell 自动化资产；同时 OEM 品牌众多。提供遵循 PowerShell / PowerCLI 熟悉风格的原生模块，尽量保留脚本的筛选、管道、报表及控制结构，降低资源访问层改造量。品牌、兼容入口和共享业务实现分别治理。

本次规划首个可交付垂直切片：安装、连接、VM 查询、启动、状态等待、有限兼容入口和 OEM 打包。不是完整 PowerCLI 克隆。底层方案与取舍见 [架构](architecture.md)。

## Project Overview

当前：Go CLI 已有基础能力，但缺少 PowerShell 原生对象和模块。期望：用户只安装一个包含 CLI 的模块包，使用熟悉的 Verb-Noun、参数、管道、WhatIf/Confirm/ErrorAction，无需理解内部进程。

| Stakeholder | 需要 |
| --- | --- |
| VMware 迁移管理员 | 最少修改已有脚本，明确差异 |
| OEM 交付团队 | 一套核心生成多个品牌，离线安装 |
| CLI 维护者 | 复用共享逻辑，不维护第二套云客户端 |
| QA / 客户验收人 | 可追溯的兼容与真实环境证据 |

## Goals and Objectives

- G1：保留约定范围内的 PowerCLI 风格及脚本处理逻辑。
- G2：安全、确定、可验证的 PowerShell 自动化。
- G3：单包交付与多 OEM 共享维护。

## Functional Requirements

<a id="fr-001"></a>
### FR-001: MUST — PowerCLI 迁移范围与契约

**Acceptance Criteria:**
- 给定一个脚本样本，记录其命令、参数、属性、管道和 VMware 专属依赖；分类为可少改迁移、需改写、未支持。
- 首版兼容清单逐项列出支持与差异，不因同名或别名而声明等价；版本基线随发行包固定。
- 没有客户真实样本时明确标记合成验证，不能宣称客户脚本迁移率。

**User Stories:** 1.1, 6.3；目标 G1。

<a id="fr-002"></a>
### FR-002: MUST — 安全连接与环境隔离

**Acceptance Criteria:**
- Connect 接受 PSCredential，密码不进入子进程参数、父进程环境、日志或长期配置；默认只保存会话。
- 两个环境并存时，对象绑定原连接；断开、过期或跨品牌对象不可退回默认 context。
- 退出时清除本连接受控会话文件；本地清理与远端 logout 语义明确，失败可诊断。

**User Stories:** 1.2, 2.4；目标 G2。

<a id="fr-003"></a>
### FR-003: MUST — 模块与 CLI 机器契约

**Acceptance Criteria:**
- 模块不依赖系统 PATH 中的 CLI；按包内清单校验版本与平台，不兼容时在云调用前失败。
- 业务成功流只含对象；stdout JSON、stderr 诊断与进度分流，退出码非零不解释为成功。
- 调用可超时/取消，参数保留中文、空白、空字符串和引号，不经 shell 拼接。

**User Stories:** 1.3, 1.4, 2.1, 2.2, 2.3；目标 G2。

<a id="fr-004"></a>
### FR-004: MUST — VM 对象与 PowerShell 查询习惯

**Acceptance Criteria:**
- Get 查询支持 Name/Id/Server 和约定通配符、数组与位置参数，返回零/一/多个对象均可管道消费。
- 对象提供 Id、Name、NumCpu、MemoryGB、PowerState；精确保留原始 State，缺失值不伪造为零。
- 名称歧义用于写操作时拒绝；Where-Object、Select-Object、Export-Csv 的约定报表无需改写处理段。

**User Stories:** 3.1, 3.2, 3.3, 3.4；目标 G1。

<a id="fr-005"></a>
### FR-005: MUST — 安全启动与状态等待

**Acceptance Criteria:**
- Start 接受本模块 VM 对象或显式选择器；写前完成唯一 UUID 和连接身份核对。
- WhatIf 不发送写请求；Confirm 遵循 PowerShell 习惯且只确认一次；ErrorAction Stop 停止后续管道写操作。
- 启动后明确区分提交和状态达到 Running；超时为未知最终结果，禁止自动重试写。

**User Stories:** 4.1, 4.2, 4.3；目标 G1。

<a id="fr-006"></a>
### FR-006: SHOULD — 显式 PowerCLI 兼容入口

**Acceptance Criteria:**
- 默认模块只导出带产品前缀命令，不修改 VMware 的全局默认连接或函数。
- 独立兼容模块仅覆盖已验证 Connect-VIServer、Disconnect-VIServer、Get-VM、Start-VM 的参数子集；检测同名冲突后拒绝加载，不强制覆盖。
- 未支持参数、VMware 类型对象及深层 API 依赖在本地报错；不能悄悄忽略。

**User Stories:** 4.4；目标 G1。

<a id="fr-007"></a>
### FR-007: MUST — OEM 品牌与共用核心

**Acceptance Criteria:**
- 从同一核心生成两种测试品牌的模块、命令、帮助、包身份及连接目录，无业务代码分叉。
- 品牌配置限制为经过校验的数据；模块目录、manifest、文件名大小写一致，不能注入代码或路径穿越。
- 常规用户界面和错误使用 OEM 品牌，内部协议不随品牌变化；法律声明保留，彻底二进制去品牌不隐含承诺。

**User Stories:** 5.1, 5.2, 5.3；目标 G3。

<a id="fr-008"></a>
### FR-008: MUST — 单包在线/离线安装与升级

**Acceptance Criteria:**
- 在已有支持版本 PowerShell 的干净机器上安装一个包即可使用，无需另装 CLI、Go、Python。
- 离线包包含 CLI 和完整模块，Import-Module 不下载、不联网、不自动登录；导入与首次安装区分清楚。
- 升级和回退按整包版本匹配 CLI；活动会话不静默切换执行文件或复用不兼容凭据。

**User Stories:** 5.2, 5.4；目标 G3。

<a id="fr-009"></a>
### FR-009: MUST — 发布验证与用户文档

**Acceptance Criteria:**
- 中英文安装、迁移、OEM 与兼容边界文档均提供；所有未实现示例明确标记。
- Windows 原生 CI 覆盖凭据、对象、管道、错误、确认、包与并存冲突；模拟测试和真实环境证据分开。
- 发布清单记录精确 PowerShell/OS/CLI/模块版本及验证范围；发布动作保持显式授权。

**User Stories:** 6.1, 6.2, 6.4；目标 G2。

<a id="fr-010"></a>
### FR-010: COULD — 后续平台与资源扩展

**Acceptance Criteria:**
- 客户需求排序中列出 macOS/Linux、5.1、关机/重启、快照及创建/删除；不混入首版完成率。
- 每项必须有语义映射、能力证据和独立故事后才能进入实现，特别区分 grace 与强制断电。
- Get-View、ExtensionData、vCenter 协议模拟和无修改全兼容不纳入首版。

**User Stories:** 6.3；目标 G1。

## Non-Functional Requirements

<a id="nfr-001"></a>
### NFR-001: MUST — Security

**Acceptance / Measurement:** 密码不出现在 argv、环境转储、成功/错误流、转录或打包产物；凭据文件 Unix 0600/私有目录，Windows 限当前用户 ACL。用合成哨兵秘密、进程启动捕获、文件 ACL 检查证明。

**User Stories:** 1.2, 2.2, 2.4, 6.2。

<a id="nfr-002"></a>
### NFR-002: MUST — Reliability

**Acceptance / Measurement:** 重复名称、跨连接对象、失效连接、非零退出码、非法 JSON 均有确定结果；所有 WhatIf 样例写请求数为 0；超时测试写调用不自动重试。

**User Stories:** 2.3, 3.3, 4.1, 4.2, 4.3, 6.2。

<a id="nfr-003"></a>
### NFR-003: SHOULD — Performance / Scalability

**Acceptance / Measurement:** 合成 10000 台 VM 查询保持分页而非逐 VM 查询；记录首项与总耗时、峰值内存。固定机器 20 次预热样本下，模块较相同 CLI 查询的增量 p95 目标不高于 500ms，峰值进程总内存目标不高于 512MiB；初次基线如不满足则报告并评审，不编造结果。

**User Stories:** 3.2, 6.2。

<a id="nfr-004"></a>
### NFR-004: MUST — Maintainability

**Acceptance / Measurement:** 两种测试品牌共用同一核心提交与契约版本，核心源文件哈希一致；OEM 差异只允许在配置、生成入口和展示资源。

**User Stories:** 5.1, 5.2, 5.3, 6.2。

<a id="nfr-005"></a>
### NFR-005: MUST — Compatibility / Usability

**Acceptance / Measurement:** 首版验证 Windows x64 + PowerShell 7.4 技术下限及发布时仍受支持的选定 7.x；列出测试精确版本。交付包不得宣称未验证 OS/5.1 支持。

**User Stories:** 2.1, 5.4, 6.1, 6.2。

<a id="nfr-006"></a>
### NFR-006: MUST — Migration evidence

**Acceptance / Measurement:** 合成基准至少包含 12 份脚本：查询 4、管道 2、启动/等待 2、错误/确认 2、OEM/并存 2；约定支持项全部通过。客户基准需客户提供及确认，分别统计通过率、修改位置和不支持项。

**User Stories:** 1.1, 4.4, 6.3, 6.4。

## Epics and User Stories (Outline)

1. CLI 前置契约：迁移管理员希望机器接口可靠，使模块无需解析人类输出。
2. 原生模块基础：脚本作者希望连接、错误与参数符合 PowerShell，使自动化可组合。
3. VM 查询与对象：管理员希望沿用 Name/PowerState/NumCpu/MemoryGB 报表，减少改写。
4. 启动与兼容入口：管理员希望沿用管道、确认和错误控制，避免误操作。
5. OEM 与安装：OEM 希望一套核心交付自己的品牌，客户只装一个包。
6. 验证与迁移交付：客户希望看到脚本级证据和明确支持范围。

完整 User Story、Given/When/Then 验收与依赖见 [epics.md](epics.md)。

## Prioritization Summary (MoSCoW)

MUST：FR-001–005、007–009 与 NFR-001/002/004–006，确保首个切片完整。
SHOULD：FR-006 的旧命令入口、NFR-003 的增量性能目标；允许在主模块可交付后推出。
COULD：FR-010 扩展，只做后续需求清单。
WON'T：完整 vCenter API 模拟、Vmware 类型等价、任意脚本零修改、首版全部资源 CRUD。

## Success Metrics

| Metric | Baseline | Target | Measurement |
| --- | --- | --- | --- |
| 单包安装 | 当前需独立 CLI，无模块 | 在线/离线均一个完整包，导入无下载 | 干净 Windows 验证 |
| 合成迁移脚本 | 未建立 | 12 份约定基准支持项全部通过 | 逐脚本 before/after 与测试报告 |
| 客户脚本复用 | 未知，未收到样本 | 首次试点确定分母及目标，不预设营销比例 | 少改迁移/重写/不支持分别统计 |
| OEM 维护 | 尚无模块 | 两种测试品牌同一核心哈希 | 包与代码差异审计 |
| 写安全 | 模块尚不存在 | WhatIf 0 次写、跨环境误写 0 次 | stub 与授权环境分别证明 |

## Assumptions and Dependencies

已有 CLI 是 Go SDK 的消费者，不存在本仓库可直接加载的 .NET SDK。首版账户密码认证；AK/SK 模块接入后续规划，绝不自动读取环境凭据。PowerShell 7.4 是技术下限，不等于永久支持所有 7.x。客户样本、精确 OEM 发布身份、模块仓库名称和签名身份尚未提供。

| Dependency | Owner | Status / mitigation |
| --- | --- | --- |
| CLI 安全输入与结构化输出 | CLI 维护者 | Stories 1.2–1.4，阻断对应模块能力 |
| 客户脱敏脚本/验收环境 | 客户负责人 | 未提供；先用合成基准，真实验收不标通过 |
| 品牌与发布身份 | OEM/产品负责人 | 配置机制可先做，正式发布前确定 |
| PowerShell/Pester 版本 | 开发/QA | 7.4 技术下限；Pester 5.x 开发依赖，锁具体版本 |

## Constraints

不改写整套云 SDK；不通过命令别名掩盖行为差异；不将真实凭据写入样例；不自动发布或真实云操作。没有 UI，DESIGN/EXPERIENCE 不适用。开发验证遵守仓库 AGENTS.md。

## Out of Scope

完整 PowerCLI API、Get-View/ExtensionData、VMware SDK 类型、Stop-VM -Kill、RunAsync/VMware Task 等价、快照/创建/删除、所有 OEM 私有 API、二进制字符串彻底去品牌均不在首版。原生正常关机与强制停止需后续单独证明语义。

## Risks and Mitigations

- 把可执行当成兼容：以命令/参数/属性/语义矩阵验收。
- CLI 存在文本输出和格式化数值：用前置故事增量补齐，不解析日志。
- 凭据和默认环境串用：stdin、隔离配置、显式连接绑定。
- 客户依赖 VMware 深层类型：标记需改写，不隐式返回伪造类型。
- OEM 分叉与命令冲突：配置生成、同名检测、独立兼容模块。

## Traceability Matrix

| Requirement | Goal | Epic | User Story | Status |
| --- | --- | --- | --- | --- |
| FR-001 | G1 | 1, 6 | 1.1, 6.3 | Planned |
| FR-002 | G2 | 1, 2 | 1.2, 2.4 | Planned |
| FR-003 | G2 | 1, 2 | 1.3, 1.4, 2.1, 2.2, 2.3 | Planned |
| FR-004 | G1 | 3 | 3.1, 3.2, 3.3, 3.4 | Planned |
| FR-005 | G1 | 4 | 4.1, 4.2, 4.3 | Planned |
| FR-006 | G1 | 4 | 4.4 | Planned |
| FR-007 | G3 | 5 | 5.1, 5.2, 5.3 | Planned |
| FR-008 | G3 | 5 | 5.2, 5.4 | Planned |
| FR-009 | G2 | 6 | 6.1, 6.2, 6.4 | Planned |
| FR-010 | G1 | 6 | 6.3 | Deferred extension |
| NFR-001 | G1/G2/G3 | Cross-cutting | 1.2, 2.2, 2.4, 6.2 | Planned |
| NFR-002 | G1/G2/G3 | Cross-cutting | 2.3, 3.3, 4.1, 4.2, 4.3, 6.2 | Planned |
| NFR-003 | G1/G2/G3 | Cross-cutting | 3.2, 6.2 | Planned |
| NFR-004 | G1/G2/G3 | Cross-cutting | 5.1, 5.2, 5.3, 6.2 | Planned |
| NFR-005 | G1/G2/G3 | Cross-cutting | 2.1, 5.4, 6.1, 6.2 | Planned |
| NFR-006 | G1/G2/G3 | Cross-cutting | 1.1, 4.4, 6.3, 6.4 | Planned |

## Handoff

先实现依赖根和查询切片，再启动，再 OEM/兼容与验收。详见 [架构](architecture.md)、[epics](epics.md)、[未决项](addendum.md)。ready-for-dev 只说明故事完整，不替代真实验证或发布授权。

## Revision History

1.0 / 2026-11-09：根据本次用户请求与源码事实建立首版规划。
