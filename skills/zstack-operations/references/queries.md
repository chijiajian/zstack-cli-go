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
