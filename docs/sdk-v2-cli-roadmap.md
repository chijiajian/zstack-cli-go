# zstack-sdk-go-v2 CLI 能力建设计划

## 当前开发交付状态

已按当前主干范围完成代码开发，收尾清单见 [completion-checklist.md](completion-checklist.md)。
`cmd/roadmap_test.go` 对主干命令注册和危险命令确认参数做持续检查。

| 能力组 | 开发状态 | 验证边界 |
| --- | --- | --- |
| 查询、分页、输出、关联详情、wait/job | 已实现 | mock、竞态回归、真实只读 smoke |
| update、计算/存储、网络、基础设施、镜像模板、运维主干 | 已实现 | 写操作仅本地 mock/预览；真实变更 E2E 未执行 |
| 本地主存储 add-local | 已补齐 | SDK 类型化请求 + Post 适配，mock 校验请求 |
| 多 context、跨进程锁、续期、logout | 已实现 | 并发回归、取消/冲突测试；logout 真实验证通过，远端会话失效且本地凭据清空 |
| 验证入口、只读脚本、CI 与发布检查门槛 | 已配置 | 本地验收；远程 CI 尚待推送触发 |
| IAM 资源写操作 | 下一阶段已实现 SDK 支持的创建/更新/删除及项目恢复/清理 | 172.34.1.10 元数据生命周期实测通过，详见 next-stage-checklist.md |

下文保留最初路线图与各轮验证记录，历史进展描述不是当前未完成清单。

## 2026-10-29 优先批次进展

已实现前三项：网络资源查询、实例/云盘关联详情、统一分页和机器输出。
网络查询包含安全组及规则、端口转发、负载均衡及监听器和后端组；
`get volumes` 兼容原有 `get disks`。具体用法与输出契约见 README。

实现分层：`pkg/query` 管理分页，`pkg/client` 管理可取消读取与认证，
`pkg/inspect` 通过窄 Reader 接口组装详情，`pkg/common` 连接命令参数与输出。
命令只声明资源路径、过滤条件及展示字段，避免复制分页或认证逻辑。

SDK 已从 v0.0.8 升级到 v0.0.10，并适配两处云盘解绑调用。
本批使用本地 HTTP mock 回归验证，尚未进行真实 ZStack 环境联调。
wait/job 增强已完成：job 默认等待成功、终态提前结束、任务结果、最后状态、
可选进度与 JSON/YAML 输出、详情请求超时；独立轮询策略及回归测试已补齐。
recover/resize 已补齐：instance recover、volume recover、volume resize（数据盘/根盘自动分流），
包含状态校验、防缩容、容量溢出校验和离线 dry-run；多 context 工作流也已补齐：命名登录、列表/当前/切换/重命名/删除、
单次 --context 覆盖、定向会话续期及配置脱敏；使用本地 mock 验证环境隔离。
以下为原能力建设清单，已完成项以本节和当前命令帮助为准。

## 目标

基于 `github.com/zstackio/zstack-sdk-go-v2 v0.0.10`，补齐 `zstack-cli` 的主干资源管理能力。

CLI 不追求一开始覆盖 SDK 中所有行业模块和长尾 action，而是优先覆盖日常运维闭环：

- 查询资源
- 创建资源
- 更新资源
- 删除或销毁资源
- 可恢复资源的 expunge
- 资源 attach/detach
- 异步任务和状态等待
- 真实环境 E2E 后自动清理

## 最初版本基线（历史记录）

### 查询

- `get zones`
- `get clusters`
- `get hosts`
- `get images`
- `get image-storages`
- `get primary-storages`
- `get disk-offerings`
- `get instance-offerings`
- `get instances`
- `get disks`
- `get snapshots`
- `get cdroms`
- `get nics`
- `get l2-networks`
- `get l3-networks`
- `get ip-ranges`
- `get vips`
- `get eips`
- `get virtual-router-offerings`
- `get virtual-routers`
- `get vm-scripts`
- `get global-configs`
- `get long-jobs`
- `get management-nodes`
- `get tags`

### 创建

- `create image`
- `create instance`
- `create disk-offering`
- `create instance-offering`

### 删除和清理

- `delete image`
- `delete instance`
- `expunge image`
- `expunge instance`

### 实例生命周期

- `instance start`
- `instance stop`
- `instance restart`
- `instance pause`
- `instance resume`

## 设计原则

1. 先补主干资源，不把 SDK 的所有长尾模块直接暴露成一级命令。
2. 每个写操作必须支持 `--dry-run` 或等价预览能力。
3. 每个危险操作必须支持确认提示和 `-y/--yes`。
4. 每个真实环境 E2E 必须使用唯一资源名，并在测试末尾 delete/expunge 或 detach/delete 清理。
5. 查询类命令必须保留通用 `-q`、`--limit`、`--start`、`--sort`、`--fields`、`-o` 能力。
6. 输出表格必须使用 CLI 自己的扁平格式化结构，避免直接打印 SDK 嵌入结构。
7. name/uuid 解析逻辑应统一复用，避免每个命令重复实现并产生不同匹配语义。

## P0：CLI 基础能力

### `update <resource>`

优先覆盖：

- `update instance`
- `update image`
- `update disk-offering`
- `update instance-offering`
- `update volume`
- `update l3-network`
- `update vip`
- `update eip`
- `update security-group`

验收：

```bash
go test ./...
go vet ./...
go build ./...
zstack-cli update instance <vm> --description ...
zstack-cli update image <image> --description ...
```

### `wait <resource>`

目标命令：

```bash
zstack-cli wait instance <name-or-uuid> --for Running --timeout 5m
zstack-cli wait image <name-or-uuid> --for Ready --timeout 10m
zstack-cli wait volume <name-or-uuid> --for Ready --timeout 5m
zstack-cli wait job <uuid> --for Succeeded --timeout 10m
```

验收：

- 成功状态返回 0
- 超时返回非 0
- 输出最后一次观测到的状态

### `jobs` / `long-jobs` 增强

目标命令：

```bash
zstack-cli get long-jobs
zstack-cli wait long-job <uuid> --timeout 10m
zstack-cli describe long-job <uuid>
```

验收：

- 能查看异步任务状态
- 失败任务能输出错误信息

### 统一资源解析

目标：

- name/uuid 解析 helper 统一放在 `pkg/client`
- 支持 exact name、fuzzy name、uuid 三类匹配策略
- 多匹配时默认报错，使用 `--all` 或交互确认才批量操作

## P1：计算和存储

### `volume`

目标命令：

```bash
zstack-cli create volume <name> --size 10G
zstack-cli volume attach <volume> --instance <vm>
zstack-cli volume detach <volume>
zstack-cli delete volume <volume>
zstack-cli expunge volume <volume>
```

SDK 重点：

- `CreateDataVolume`
- `AttachDataVolumeToVm`
- `DetachDataVolumeFromVm`
- `DeleteDataVolume`
- `ExpungeDataVolume`
- `QueryVolume`

E2E：

1. 创建临时 data volume
2. attach 到临时 VM
3. detach
4. delete
5. expunge
6. 确认查询不可见

### `snapshot`

目标命令：

```bash
zstack-cli snapshot create <volume> --name <name>
zstack-cli delete snapshot <snapshot>
zstack-cli snapshot revert <snapshot>
```

SDK 重点：

- `CreateVolumeSnapshot`
- `CreateVolumesSnapshot`
- `DeleteVolumeSnapshot`
- `QueryVolumeSnapshot`

E2E：

1. 创建临时 VM
2. 对 root volume 创建 snapshot
3. 查询 snapshot
4. 删除 snapshot

### `instance` 扩展

目标命令：

```bash
zstack-cli instance update <vm> --name ... --description ...
zstack-cli instance attach-l3 <vm> --l3-network <l3>
zstack-cli instance detach-l3 <vm> --l3-network <l3>
zstack-cli instance attach-iso <vm> --iso <image>
zstack-cli instance detach-iso <vm> --iso <image>
zstack-cli instance attach-volume <vm> --volume <volume>
zstack-cli instance detach-volume <vm> --volume <volume>
```

SDK 重点：

- `UpdateVmInstance`
- `AttachL3NetworkToVm`
- `DetachL3NetworkFromVm`
- `AttachIsoToVmInstance`
- `DetachIsoFromVmInstance`
- `AttachDataVolumeToVm`
- `DetachDataVolumeFromVm`

### `vm-script`

目标命令：

```bash
zstack-cli create vm-script <name> --file script.sh --platform Linux
zstack-cli update vm-script <script> ...
zstack-cli delete vm-script <script>
zstack-cli vm-script execute <script> --instance <vm>
zstack-cli get vm-script-records
```

SDK 重点：

- `CreateGuestVmScript`
- `UpdateGuestVmScript`
- `DeleteGuestVmScript`
- `ExecuteGuestVmScript`
- `QueryGuestVmScriptExecutedRecord`

## P1：网络

### `vip`

目标命令：

```bash
zstack-cli create vip <name> --l3-network <l3>
zstack-cli update vip <vip> --name ...
zstack-cli delete vip <vip>
```

SDK 重点：

- `CreateVip`
- `UpdateVip`
- `DeleteVip`
- `QueryVip`

### `eip`

目标命令：

```bash
zstack-cli create eip <name> --vip <vip>
zstack-cli eip attach <eip> --nic <nic>
zstack-cli eip detach <eip>
zstack-cli update eip <eip> --name ...
zstack-cli delete eip <eip>
```

SDK 重点：

- `CreateEip`
- `AttachEip`
- `DetachEip`
- `UpdateEip`
- `DeleteEip`
- `QueryEip`

### `security-group`

目标命令：

```bash
zstack-cli create security-group <name>
zstack-cli security-group attach-l3 <sg> --l3-network <l3>
zstack-cli security-group detach-l3 <sg> --l3-network <l3>
zstack-cli security-group add-rule <sg> --type Ingress --protocol TCP --start-port 22 --end-port 22 --cidr 0.0.0.0/0
zstack-cli security-group delete-rule <rule>
zstack-cli security-group add-nic <sg> --nic <nic>
zstack-cli security-group remove-nic <sg> --nic <nic>
zstack-cli delete security-group <sg>
```

SDK 重点：

- `CreateSecurityGroup`
- `UpdateSecurityGroup`
- `DeleteSecurityGroup`
- `AttachSecurityGroupToL3Network`
- `DetachSecurityGroupFromL3Network`
- `AddSecurityGroupRule`
- `DeleteSecurityGroupRule`
- `AddVmNicToSecurityGroup`
- `DeleteVmNicFromSecurityGroup`

### `port-forwarding`

目标命令：

```bash
zstack-cli create port-forwarding <name> --vip <vip> --protocol TCP --vip-port 22 --private-port 22
zstack-cli port-forwarding attach <rule> --nic <nic>
zstack-cli port-forwarding detach <rule>
zstack-cli delete port-forwarding <rule>
```

SDK 重点：

- `CreatePortForwardingRule`
- `AttachPortForwardingRule`
- `DetachPortForwardingRule`
- `DeletePortForwardingRule`
- `QueryPortForwardingRule`

### `load-balancer`

目标命令：

```bash
zstack-cli create load-balancer <name> --vip <vip>
zstack-cli create load-balancer-listener <name> --load-balancer <lb> --protocol TCP --port 80
zstack-cli create load-balancer-server-group <name> --load-balancer <lb>
zstack-cli load-balancer add-nic <listener> --nic <nic>
zstack-cli load-balancer remove-nic <listener> --nic <nic>
zstack-cli delete load-balancer <lb>
```

SDK 重点：

- `CreateLoadBalancer`
- `UpdateLoadBalancer`
- `DeleteLoadBalancer`
- `CreateLoadBalancerListener`
- `DeleteLoadBalancerListener`
- `CreateLoadBalancerServerGroup`
- `DeleteLoadBalancerServerGroup`
- `AddVmNicToLoadBalancer`
- `RemoveVmNicFromLoadBalancer`

## P2：基础设施管理

### Zone / Cluster / Host

目标命令：

```bash
zstack-cli create zone <name>
zstack-cli update zone <zone> ...
zstack-cli delete zone <zone>

zstack-cli create cluster <name> --zone <zone> --hypervisor KVM
zstack-cli update cluster <cluster> ...
zstack-cli delete cluster <cluster>

zstack-cli host add-kvm <name> --management-ip <ip> --username root --password ...
zstack-cli update host <host> ...
zstack-cli delete host <host>
```

SDK 重点：

- `CreateZone`
- `UpdateZone`
- `DeleteZone`
- `CreateCluster`
- `UpdateCluster`
- `DeleteCluster`
- `AddKVMHost`
- `UpdateHost`
- `DeleteHost`

### Primary Storage / Backup Storage

目标命令：

```bash
zstack-cli primary-storage add-nfs <name> --url ...
zstack-cli primary-storage attach-cluster <ps> --cluster <cluster>
zstack-cli primary-storage detach-cluster <ps> --cluster <cluster>
zstack-cli update primary-storage <ps> ...
zstack-cli delete primary-storage <ps>

zstack-cli backup-storage add-image-store <name> --url ...
zstack-cli backup-storage attach-zone <bs> --zone <zone>
zstack-cli backup-storage detach-zone <bs> --zone <zone>
zstack-cli update backup-storage <bs> ...
zstack-cli delete backup-storage <bs>
```

SDK 重点：

- `AddNfsPrimaryStorage`
- `AddLocalPrimaryStorage`
- `UpdatePrimaryStorage`
- `DeletePrimaryStorage`
- `AttachPrimaryStorageToCluster`
- `DetachPrimaryStorageFromCluster`
- `AddImageStoreBackupStorage`
- `UpdateBackupStorage`
- `DeleteBackupStorage`
- `AttachBackupStorageToZone`
- `DetachBackupStorageFromZone`

### L2 / L3 / IP Range / DNS / Host Route

目标命令：

```bash
zstack-cli create l2-network <name> ...
zstack-cli l2-network attach-cluster <l2> --cluster <cluster>
zstack-cli l2-network detach-cluster <l2> --cluster <cluster>

zstack-cli create l3-network <name> --l2-network <l2>
zstack-cli update l3-network <l3> ...
zstack-cli delete l3-network <l3>
zstack-cli l3-network add-ip-range <l3> ...
zstack-cli l3-network add-dns <l3> --dns 223.5.5.5
zstack-cli l3-network remove-dns <l3> --dns 223.5.5.5
zstack-cli l3-network add-host-route <l3> --prefix ... --nexthop ...
zstack-cli l3-network remove-host-route <l3> --prefix ...
```

SDK 重点：

- `CreateL2VlanNetwork`
- `AttachL2NetworkToCluster`
- `DetachL2NetworkFromCluster`
- `CreateL3Network`
- `UpdateL3Network`
- `DeleteL3Network`
- `AddIpRange`
- `AddIpRangeByNetworkCidr`
- `AddDnsToL3Network`
- `RemoveDnsFromL3Network`
- `AddHostRouteToL3Network`
- `RemoveHostRouteFromL3Network`

## P2：镜像和模板

目标命令：

```bash
zstack-cli update image <image> ...
zstack-cli image recover <image>
zstack-cli image sync <image>
zstack-cli image create-from-volume <volume> --name <name>
zstack-cli image create-from-snapshot <snapshot> --name <name>
zstack-cli get image-groups
zstack-cli image-group create-from-image <image> --name <name>
zstack-cli expunge image-group <group>
```

SDK 重点：

- `UpdateImage`
- `RecoverImage`
- `SyncImage`
- `CreateDataVolumeTemplateFromVolume`
- `CreateDataVolumeTemplateFromVolumeSnapshot`
- `CreateImageGroupFromImage`
- `ExpungeImageGroup`

## P3：运维和平台

### Global Config

目标命令：

```bash
zstack-cli get global-configs
zstack-cli global-config set <category.name> --value ...
```

### Alarm / Event / Audit

目标命令：

```bash
zstack-cli get alarms
zstack-cli create alarm ...
zstack-cli delete alarm <alarm>
zstack-cli get events
zstack-cli get audits
```

### IAM / Account

历史待定范围（现已按用户确认实现适用管理操作，详见 `next-stage-checklist.md`）：

- account
- user
- role
- policy
- project
- organization

这部分 SDK 覆盖较广，但命令语义和权限模型更复杂，建议放在主干资源稳定后再做。

## 验收策略

每个新增命令至少满足：

```bash
go test ./...
go vet ./...
go build ./...
zstack-cli <command> --help
```

写操作必须增加真实环境 E2E：

1. 使用唯一资源名，例如 `codex-<feature>-e2e-<timestamp>`。
2. 创建资源。
3. 查询确认资源存在。
4. 执行目标操作。
5. 查询确认状态或字段变化。
6. 执行 cleanup。
7. 查询确认资源已清理。

真实环境 E2E 推荐复用 `.env`：

```bash
set -a
source .env
set +a
```

`.env` 必须保持 git ignored。

## 建议执行顺序

1. P0：`update`、`wait`、资源解析统一、long job watch。
2. P1：`volume`、`snapshot`、`instance attach/detach`、`vm-script execute`。
3. P1：`vip`、`eip`、`security-group`、`port-forwarding`。
4. P1/P2：`load-balancer`。
5. P2：zone、cluster、host、storage、L2/L3 网络管理。
6. P2：image update/recover/sync/template。
7. P3：global-config、alarm、event、audit、IAM。

### wait/job 实现约定

轮询在 `pkg/waiter` 中，通过 Observe/Evaluate 函数依赖读取与完成策略，
不依赖 Cobra、HTTP 或 SDK。CLI 只负责参数、读取适配和输出。
任务终态依据 [ZStack LongJobState](https://github.com/zstackio/zstack/blob/master/header/src/main/java/org/zstack/header/longjob/LongJobState.java)：
Succeeded、Failed、Canceled；Suspended 不是终态。显式目标先于终态错误判断。
本批为读取和等待增强，不增加任务恢复、取消或清理动作；验证使用本地 mock。

### 配置并发写入

登录保存、context 管理、会话续期和登出统一通过 `UpdateConfig` 事务：
跨进程 OS 文件锁覆盖读取、修改和原子保存，网络请求在锁外执行。
包含多 CLI 写入、双续期与切换竞争、删除/替换凭据冲突、锁等待取消和
持锁进程被终止后的恢复测试。文件锁实现分别适配 Linux/macOS 与 Windows。

### 真实环境只读联调

使用隔离临时配置，通过 CLI 账号登录测试环境 API（8080；5000 为 Web 控制台）。
验证 20 类资源的有限条数查询，以及全量分页、count、JSON/YAML 字段投影、
父资源过滤、实例/云盘/任务详情、wait 当前状态和显式 context 查询。
未执行云资源创建、删除、恢复、扩容、解绑等操作，也未创建 AK。

联调发现并修复：
- 读取层缺少 SDK 的 ZStack 时间格式支持。仅对 time.Time 字段使用 SDK 时间解析，
  保留普通字符串、未类型化详情和大整数；覆盖 get/describe/wait 回归。
- 本环境不支持主存储按 UUID 路径 GET。云盘详情的主存储关联改用列表接口的
  UUID 精确过滤，保留返回数量校验和关联失败告警。

此验证不代表写操作或真实多环境并发工作流已完成 E2E；这些仍以 mock 回归为主。

### 剩余查询与边界只读联调

第二轮覆盖剩余 22 个查询主命令：accounts、alarms、audits、cdroms、eips、
events、global-configs、image-groups、ip-ranges、l2-networks、management-nodes、
organizations、policies、projects、roles、tags、users、vips、virtual-router-offerings、
virtual-routers、vm-script-records、vm-scripts。累计覆盖 42 个主命令。

验证名称过滤、count、空列表及空分页封装、字段投影、分页非零起点和越界起点，
以及非法参数的非零退出与 stdout/stderr 分离。审计和事件仅限量或精确过滤，
未全量拉取日志。CD-ROM 无位置参数，脚本记录的位置参数对应 recordName。

发现并修复：global-configs/events/audits 的库存无 uuid，默认分页排序应为 id；
通过 query.Option 由资源命令声明默认排序，分页核心无需感知资源路径。
表格/文本输出同样校验字段，避免不存在的字段产生成功空输出。
对应多页、用户排序覆盖、空/非空结果和各输出格式回归已加入。

### 两轮只读联调改动复审

新增可复现回归后修复：最后一页/count 返回时取消仍可能报告成功、单页错误
携带部分结果、文本字段筛选保留未选字段零值、响应 JSON 尾随内容被忽略，
以及 SDK 时间在大小写字段名和类型化 map 中解码失败。
分页取消规则、JSON 文档边界校验、类型化时间兼容与字段投影保持独立职责。
本轮以本地 mock 和单元测试验证，不扩展真实环境操作范围。

### 2026-10-29 真实环境只读复测

本轮使用当前源码构建 CLI、隔离配置及不保存密码的临时会话，完成 101 次
CLI 检查；除登录及退出临时会话外仅查询，未修改云资源或默认 context。

- 覆盖 zones、clusters、hosts、instances、volumes、images、primary-storages、
  long-jobs、global-configs、events、audits 共 11 类查询。
- 限定资源样本，对照单次查询与每页 1 条的 all-pages 结果，检查 start、count、
  total 和字段投影；global-configs/events/audits 另验证 3 条样本的倒序分页。
- 验证 JSON/YAML/text/table 字段输出、空结果、未知字段、负分页参数和极短超时。
- 验证 instance/volume 详情；将 job 原始 API 的 createDate/lastOpDate 与 CLI
  规范化输出比较，确认时间点一致。
- 未发现新的功能缺陷。加强自动回归的分页结果顺序断言，以及 text 字段投影在
  空/非空结果下将 total 输出到 stderr 的契约。

异常 JSON、类型化 map/指针时间、最后一页返回瞬间取消等无法稳定从真实服务
触发的边界继续由本地回归覆盖。真实数据验证采用有限样本，不代表并发变动期间
的快照一致性。测试会话已退出（HTTP 200），临时配置、二进制和脚本已删除。

### 2026-10-29 边界复审

- P1：资源解析和批量实例/镜像操作辅助函数原先接受缺失 UUID 的库存记录，
  后续操作可能获得空标识。现在遇到此类响应立即失败；回归覆盖 null/空对象/
  空白 UUID，以及缺失 UUID 时恢复命令不发送变更请求。
- P2：instance/volume/job 详情原先只检查记录数，可能把 null、空对象或其他
  UUID 的详情作为成功结果。现在验证返回 UUID 与请求标识一致。
- P2：公共 JSON/YAML/text/table 输出及查询 total 的写入失败未完整上报。
  现在传递写入错误；表格写入适配器记录依赖渲染器吞掉的底层错误，保持
  现有格式。回归使用已关闭的输出文件及失败的 stderr writer。

标识校验留在解析/详情边界，写入错误处理留在输出层；未调整命令参数和默认
输出样式。本轮使用本地单元与 CLI 回归，不涉及真实环境变更。

### 2026-10-29 标识与输出修复的真实环境复测

使用隔离配置和不保存密码的临时会话，通过 CLI 完成 46 项正常读取检查：
实例/卷/任务各取 2 条样本校验详情 UUID；核对关联资源标识；实例、卷、任务、
镜像等待当前已观察到的状态；检查四种列表投影与详情输出格式。

本地异常回归发现详情 warning 写入 stderr 失败仍返回成功，已在详情输出边界
补上传递错误。新增 CLI 子进程回归覆盖四种格式的不可写 stdout，以及详情
warning 的不可写 stderr。缺失 UUID、空详情、错误 UUID、禁止发送恢复请求等
已有回归继续覆盖，未在真实环境制造异常或执行资源变更。

修复后重建 CLI，又执行 10 项实例/卷查询和四种详情格式检查，全部通过；
累计 56 次 CLI 检查。临时会话已退出（HTTP 200），配置、二进制与脚本已删除，
默认 context 保持不变。SDK 批量操作辅助函数由本地测试覆盖，不为验证它们
在真实环境执行变更命令。

### 2026-10-29 错误传递一致性复审

- 资源解析在每次查询返回后检查取消状态，丢弃失败或取消后的结果。
- wait 的进度通知现在返回错误；写入失败立即终止，避免继续轮询或报告成功。
  通知期间取消也优先于“已达到目标”结果。
- 11 类资源结果辅助函数复用共享输出实现，文本/name 输出失败返回错误。
- dry-run 复用公共 JSON/YAML 格式化器并返回序列化和写入错误，所有调用方
  传递错误；保持原有文本、默认格式及无资源变更语义。

补充 CLI 不可写进度/预览输出、各资源结果写入失败、JSON 编码失败，以及
查询返回和通知期间取消的回归。进度测试给认证留出 5 秒预算，通过请求次数
断言输出失败后不再轮询，避免依赖短超时造成竞态测试误报。

### 可重复的只读回归入口

新增 `scripts/readonly_smoke.py`（Python 3 标准库）及 `make test-smoke`。
脚本要求显式提供已认证配置，复制到私有临时目录；只允许 get/describe，
分页使用最多 3 个样本的 UUID/ID 过滤。报告只包含检查名称、状态与简短失败
原因，区分失败和缺少样本导致的跳过。global-configs 的 CLI 输出没有稳定 ID，
自动分页检查明确跳过，避免按名称范围扫描。

本地 8 项测试覆盖正常结果、空环境、畸形响应/标识、分页不一致、变更命令
拒绝、错误信息不泄露原始输出、配置隔离与清理。2026-10-29 在真实环境运行：
68 项通过，0 项失败，1 项明确跳过（global-configs 分页）。临时认证会话已退出
（HTTP 200），临时文件随验证结束删除。现有 CLI 命令及默认输出未改变。

### 只读回归脚本边界复审

修复分页比较的共同漏报：以初始样本为独立基准，校验正倒序、重复标识、
完整性、offset 和严格整数 count；字段投影固定样本，验证 JSON/YAML/text/table
的实际值与形状，不再将任意非空文本视为通过。补上 Unicode 表格分隔线兼容，
以及 Python bool 与 int 相等可能掩盖错误 ID 的回归。

配置复制后先执行本地脱敏 config view 预检，拒绝选中账号保存密码的配置，
防止自动续期产生未归属的新会话。保留无密码会话和 AK 支持，清除继承的
ZSTACK_* 变量后显式指定临时配置。增加整轮 deadline 和 SIGINT/SIGTERM 清理；
实际子进程测试确认超时/信号退出后子进程被回收，临时配置删除，源配置未改动。

可选关联资源的 403/404 或空结果单独标记为覆盖跳过，其他错误（如身份不符）
仍失败；非详情命令成功退出但写出诊断信息同样失败，报告不泄露原始错误正文。
本地 23 项回归通过。真实环境复测为 68 通过、0 失败、1 项已知跳过；临时会话
退出 HTTP 200，临时验证文件已清理。本轮未更改 CLI 产品代码。

### 日常验证与 CI

新增统一 `make verify` 入口（vet、完整竞态回归、Python 脚本本地测试、构建），
Go 测试禁用缓存并设置 5 分钟超时；可单独运行 `make test` / `make test-smoke`。
CI 在 Linux/macOS 执行同一入口，Linux 追加 Windows 程序和配置锁测试交叉编译。
发布工作流调用可复用 CI，检查完成后才能进入签名和发布任务。

此工作流只运行本地 mock/子进程测试，不含真实环境凭据或真实 smoke 调用。
Windows 运行时测试、GitHub 远程 runner 的实际结果须在推送后确认；本地验证
不能替代这些平台的执行结果。

CI 接入时发现超时子进程测试的 0.5 秒预算不足以覆盖高负载下解释器启动，
已调整为 5 秒整轮预算，仍以休眠 30 秒的子进程确认超时终止及回收。
工作流经 actionlint 检查，Windows 程序和配置锁测试通过本地交叉编译。
