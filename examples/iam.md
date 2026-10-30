# IAM

准备权限 0600 的 `account-password` 文件（只含密码，不纳入版本控制）。下列策略文件仅允许一项账户查询，创建角色不会自动将角色分配给用户。

```sh
zstack-cli create account demo-account --context lab --password-file ./account-password --dry-run -o json
zstack-cli create account demo-account --context lab --password-file ./account-password -o json
zstack-cli update account demo-account --context lab --description 'Temporary test account'
zstack-cli create policy demo-query --context lab --statements-file examples/policy-statements.json
zstack-cli create role demo-role --context lab --policy demo-query
zstack-cli create organization demo-department --context lab --type Department
zstack-cli create project demo-project --context lab --organization demo-department
zstack-cli update project demo-project --context lab --description ''
```

组织/项目需要相应许可。更新不支持 policy。以下仅清理本例自行创建的资源，重名时先查 UUID：

```sh
zstack-cli delete project demo-project --context lab --yes
zstack-cli project recover demo-project --context lab --yes
zstack-cli delete project demo-project --context lab --yes
zstack-cli expunge project demo-project --context lab --yes
zstack-cli delete organization demo-department --context lab --yes
zstack-cli delete role demo-role --context lab --yes
zstack-cli delete policy demo-query --context lab --yes
zstack-cli delete account demo-account --context lab --yes
```
