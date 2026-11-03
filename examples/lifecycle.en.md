# Resource lifecycle

[English](lifecycle.en.md) | [简体中文](lifecycle.md) | [Examples](README.en.md)

These are independent examples. Replace UUIDs with authorized test resources. Inspect current state before choosing an operation.

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

Resize specifies the new total capacity and only supports growth. Snapshot revert changes disk contents. Recover a Deleted data volume or Destroyed VM. Recover a root volume through its VM. Expunged resources cannot be recovered.

```sh
zstack-cli volume recover VOLUME_UUID --context lab
zstack-cli instance recover VM_UUID --context lab
zstack-cli job resume JOB_UUID --context lab --dry-run -o json
zstack-cli job resume JOB_UUID --context lab --yes
zstack-cli wait job JOB_UUID --context lab --timeout 10m -o json
zstack-cli job clean JOB_UUID --context lab --yes
```

Job resume requires Suspended. Clean accepts Succeeded, Failed or Canceled. Wait polls state without starting or recovering the resource.
