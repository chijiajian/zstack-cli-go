# 查询与自动化

[English](queries.en.md) | [简体中文](queries.md) | [示例目录](README.md)

```sh
zstack-cli config get-contexts
zstack-cli doctor --context lab --offline -o json
zstack-cli doctor --context lab --request-timeout 15s -o json
zstack-cli get instances --context lab -q 'state=Stopped' --sort '+uuid' --limit 20 -o json
zstack-cli get instances --context lab --all-pages --limit 100 --fields uuid,name,state -o json
zstack-cli get instances --context lab --start 20 --limit 20 --reply-with-count -o json
zstack-cli describe instance VM_UUID --context lab -o json
```

保存退出码并分离业务结果与诊断：

```sh
if zstack-cli get volumes --context lab -o json --error-format json > volumes.json 2> error.json; then
  cat volumes.json
else
  status=$?
  printf 'exit=%s\n' "$status"
  cat error.json
fi
```

正常输出为空列表仍是成功；分页不是快照。详情中的 warnings 要单独处理。get 可能续期保存密码的账户；严格只读任务请使用 session-only 配置或 AK/SK。

## 比较环境使用率

分别查询指定 context，比较时同时核对 load.sampleTime、sampledHosts 和 connectedHosts。

```sh
zstack-cli monitor usage --context lab
zstack-cli monitor usage --context staging --section load -o json --error-format json
zstack-cli monitor usage --context lab --section capacity -o yaml
```

## 指定云主机和 ZQL

```sh
zstack-cli get instances web-01 --context lab
zstack-cli describe instance VM_UUID --context lab -o json
zstack-cli zql "query vmInstance.uuid where name='web-01' limit 10" --context lab
zstack-cli zql "count vmInstance where state='Running'" --context lab
```

查某台 VM 最近五分钟的 CPU，先将下列两个 VM_UUID 替换为同一个已确认的 UUID：

```sh
zstack-cli zql "query vmInstance.uuid where uuid='VM_UUID' return with (zwatch{resultName='cpu',metricName='CPUAverageUsedUtilization',offsetAheadOfCurrentTime=300,period=60,labels='VMUuid=VM_UUID'})" --context lab
```

读取 `results[0].returnWith.cpu` 中的 time、value 和 labels。零值和没有采样的含义不同。
