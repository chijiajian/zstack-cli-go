# 生命周期与任务

以下 UUID/context 是占位符，示例不是可以直接连续执行的脚本。先查询目标身份、关联和状态，并确认操作在用户授权范围内。

```sh
zstack-cli wait instance VM_UUID --context lab --for Stopped --timeout 5m -o json --error-format json
zstack-cli wait job JOB_UUID --context lab --timeout 10m -o json --error-format json
zstack-cli volume resize VOLUME_UUID --context lab --size 40G --dry-run -o json
zstack-cli volume resize VOLUME_UUID --context lab --size 40G -o json --error-format json
zstack-cli instance recover VM_UUID --context lab -o json --error-format json
zstack-cli volume recover VOLUME_UUID --context lab -o json --error-format json
```

wait 不启动、恢复或取消资源。job 默认等待 Succeeded，其他资源需要 for；显式等待 Failed 时到达 Failed 是条件成功。等待成功遇到 Failed/Canceled 则非零。进度写 stderr，自动解析诊断时不要混用 progress。

实例恢复要求 Destroyed；数据盘恢复要求 Deleted；根盘随实例恢复，expunge 后不能恢复。扩容要求 Ready 且目标总容量大于当前值，G 按 1024³ 字节计算。扩容不调整 guest 分区/文件系统。先读取当前容量，不把 size 当增量。恢复/扩容预览是离线的，不能证明服务器会接受请求。

```sh
zstack-cli volume attach VOLUME_UUID --instance VM_UUID --context lab
zstack-cli volume detach VOLUME_UUID --context lab
zstack-cli snapshot create VOLUME_UUID --name TEST_SNAPSHOT --context lab -o json
zstack-cli snapshot revert SNAPSHOT_UUID --context lab --yes
zstack-cli delete snapshot SNAPSHOT_UUID --context lab --yes
zstack-cli job resume JOB_UUID --context lab --yes -o json
zstack-cli job clean JOB_UUID --context lab --yes -o json
```

snapshot revert 会改变盘内容。job resume 仅接受 Suspended，clean 仅接受 Succeeded/Failed/Canceled；resume 返回不等于作业完成，必要时继续 wait。不要把 get 查询出的失败任务自动 clean。

写后查询核对 UUID 和目标属性：挂载的 vmInstanceUuid、扩容后的 size、恢复后的 state/status。缺少字段不能算验证成功。服务器可能直接彻底清理删除对象，此时报告无法验证 recover，不擅自创建替代对象。
