# IAM

account/user/role/project/organization 支持 create/update/delete；policy 只有 create/delete，不虚构 update policy。user 属于所选 context 的账户。组织/项目要求 IAM2 相应许可；许可缺失不是重试能修复的。

IAM 的 password-file 包含明文密码，CLI 发送前哈希，预览脱敏；不要在示例或日志写密码。login 不支持 password-file。创建使用参数和 statements-file，不支持 create 的通用 file。

```sh
zstack-cli create account demo-account --context lab --password-file /secure/password --dry-run -o json
zstack-cli create account demo-account --context lab --password-file /secure/password -o json --error-format json
zstack-cli create policy query-only --context lab --statements-file /path/statements.json -o json
zstack-cli create role demo-role --context lab --policy query-only -o json
zstack-cli create organization demo-department --context lab --type Department -o json
zstack-cli create project demo-project --context lab --organization demo-department -o json
zstack-cli update project PROJECT_UUID --context lab --description '' -o json
```

statements.json 为数组，例如：

```json
[{"name":"query-account","effect":"Allow","actions":["org.zstack.header.identity.APIQueryAccountMsg"]}]
```

创建 policy/role 不表示已分配给用户。空 description 显式清空字段。创建 resource-uuid 必须为 32 位十六进制；用它记录请求身份后，可在结果未知时核对云端资源。update/delete 使用核实的 UUID；引用组织、账户、策略名称时精确解析。

项目恢复与彻底清理是独立写动作：delete 后可以 recover；要 expunge 一个已恢复项目，需要先再次 delete。expunge 不可恢复。不要为演示恢复而删除现有项目。IAM 请求默认超时 2 分钟，可指定 request-timeout；修改返回 UUID 不匹配是 API 失败，不当成成功，也不重发写入。
