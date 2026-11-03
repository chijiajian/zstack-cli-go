# 查询、诊断与环境

先核对选择：`zstack-cli config get-contexts -o json`、`zstack-cli config view -o json`。后者返回脱敏的 contexts、current-context、selected-context。不要把它当成可登录配置导出。

- `--context lab` 临时选择，不修改默认值；`config use-context lab` 会持久化改变默认值。
- 账户登录可用 `login HOST:8080 --username admin --context-name lab --save-password=false` 在终端提示密码。登录属于认证操作，先确认任务包含登录需求。endpoint 是 host 或 host:port，不是控制台 URL。
- `--save-password` 默认为 true；严格只读任务需要 false 创建的会话或已有 AK/SK。config view 的 password 字段非空（即使已脱敏）说明保存了密码。
- `doctor --context lab --offline -o json` 只检查本地；在线 doctor 验证已有会话及一条 zone 查询，不续期。通过不代表所有模块许可和权限可用。
- logout 可能撤销会话；不要在查询结束时注销用户提供的共享会话。只注销任务明确拥有的临时会话。AK/SK logout 只清本地，不撤销远端密钥。

```sh
zstack-cli get instances --context lab -q 'state=Stopped' --sort '+uuid' --limit 20 -o json --error-format json
zstack-cli get instances --context lab --start 20 --limit 20 --reply-with-count -o json --error-format json
zstack-cli get instances --context lab --all-pages --limit 100 --fields uuid,name,state -o json --error-format json
zstack-cli describe instance VM_UUID --context lab --request-timeout 15s -o json --error-format json
```

get 普通 JSON 返回列表，空为 []；reply-with-count 返回 items/total；count 只返回总数。start 是偏移，limit 是页大小，all-pages 可能产生大量结果，仅在任务需要完整结果时使用。分页不是快照；默认使用有界查询，并根据实际资源字段稳定排序。global-configs/events/audits 使用 id，不能统一当 uuid。

字段名以目标命令实际输出为准。describe 返回 resource 和可能的关联/warnings；不要把列表行当详情结构。主资源成功但关联缺失不等于关联已验证。`get --help` 发现资源名，`describe --help` 发现可用详情；不要猜测所有资源都有 describe。

查询通常使用 request-timeout；wait 使用 timeout。不要给所有命令机械添加相同参数。

### 查询环境使用率

用户询问 CPU、内存、主存储容量或多个环境的资源使用率时，对每个指定 context 执行：

```sh
zstack-cli monitor usage --context lab -o json --error-format json
```

仅查容量用 `--section capacity`，仅查主机负载用 `--section load`。容量百分比是（总量减可用量）除以总量，JSON 中内存及存储单位为 bytes。load 是最近 15 分钟、60 秒粒度内最新共同采样时刻的主机算术平均；报告 sampleTime、sampledHosts / connectedHosts。null 表示无可计算值。若帮助中缺少 monitor usage，报告当前二进制缺少该能力。

### 指定云主机与 ZQL

精确名称列表用 `get instances NAME`，按 UUID 查看详情用 `describe instance UUID`。先确认目标 UUID，再查询单机配置或负载。ZQL 入口通过 `zql --help` 发现，完整语句作为一个参数传入，返回 JSON。

```sh
zstack-cli zql "query vmInstance.uuid where name='web-01' limit 10" --context lab
zstack-cli zql "query vmInstance.uuid where uuid='VM_UUID' return with (zwatch{resultName='cpu',metricName='CPUAverageUsedUtilization',offsetAheadOfCurrentTime=300,period=60,labels='VMUuid=VM_UUID'})" --context lab
```

第二条命令的两个 VM_UUID 必须相同。读取 results[0].returnWith.cpu 并核对 labels 和 time；空序列表示缺少采样。环境级负载仍用 monitor usage。用资源名拼接 ZQL 有引号转义风险，优先用已确认的十六进制 UUID。
