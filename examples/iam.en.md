# IAM

[English](iam.en.md) | [简体中文](iam.md) | [Examples](README.en.md)

Prepare an account-password file containing only the new password, preferably with mode 0600, and keep it out of version control. The policy file defines one account-query permission. Creating a role does not assign it to users.

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

Organizations and projects need server licenses. Policy update is unsupported. The following cleanup applies to resources created for this example. Resolve UUIDs first if names are ambiguous.

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
