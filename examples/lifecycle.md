# 资源生命周期

以下是独立操作示例，不是连续执行脚本；UUID 必须指向已授权的测试资源。状态不满足时先调查，不强行推进。

```sh
zstack-cli wait instance VM_UUID --context lab --for Stopped --timeout 5m
zstack-cli volume attach VOLUME_UUID --context lab --instance VM_UUID
zstack-cli snapshot create VOLUME_UUID --context lab --name demo-snapshot
zstack-cli snapshot revert SNAPSHOT_UUID --context lab --yes
zstack-cli delete snapshot SNAPSHOT_UUID --context lab --yes
zstack-cli volume resize VOLUME_UUID --context lab --size 2G --dry-run
zstack-cli volume resize VOLUME_UUID --context lab --size 2G
zstack-cli volume detach VOLUME_UUID --context lab
```

扩容的目标总容量必须大于当前值，不支持缩容。快照回滚会影响数据。数据盘 Deleted 后才能 recover，实例 Destroyed 后才能 recover，根盘通过实例恢复；expunge 后无法恢复。

```sh
zstack-cli volume recover VOLUME_UUID --context lab
zstack-cli instance recover VM_UUID --context lab
zstack-cli job resume JOB_UUID --context lab --dry-run -o json
zstack-cli job resume JOB_UUID --context lab --yes
zstack-cli wait job JOB_UUID --context lab --timeout 10m -o json
zstack-cli job clean JOB_UUID --context lab --yes
```

job resume 要求 Suspended，clean 要求 Succeeded/Failed/Canceled。wait 本身不执行恢复或启动。
