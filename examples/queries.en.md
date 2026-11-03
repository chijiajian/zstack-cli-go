# Queries and automation

[English](queries.en.md) | [简体中文](queries.md) | [Examples](README.en.md)

```sh
zstack-cli config get-contexts
zstack-cli doctor --context lab --offline -o json
zstack-cli doctor --context lab --request-timeout 15s -o json
zstack-cli get instances --context lab -q 'state=Stopped' --sort '+uuid' --limit 20 -o json
zstack-cli get instances --context lab --all-pages --limit 100 --fields uuid,name,state -o json
zstack-cli get instances --context lab --start 20 --limit 20 --reply-with-count -o json
zstack-cli describe instance VM_UUID --context lab -o json
```

Preserve the CLI exit code and keep resource output separate from diagnostics.

```sh
if zstack-cli get volumes --context lab -o json --error-format json > volumes.json 2> error.json; then
  cat volumes.json
else
  status=$?
  printf 'exit=%s\n' "$status"
  cat error.json
fi
```

An empty list is a successful query. Pagination can be affected by concurrent changes. Inspect detail warnings separately. Queries can renew account sessions with saved passwords. Strict read-only work uses session-only credentials or AK/SK.

## Compare environment usage

Query each context and compare load.sampleTime, sampledHosts and connectedHosts alongside the values.

```sh
zstack-cli monitor usage --context lab
zstack-cli monitor usage --context staging --section load -o json --error-format json
zstack-cli monitor usage --context lab --section capacity -o yaml
```

## Individual VMs and ZQL

```sh
zstack-cli get instances web-01 --context lab
zstack-cli describe instance VM_UUID --context lab -o json
zstack-cli zql "query vmInstance.uuid where name='web-01' limit 10" --context lab
zstack-cli zql "count vmInstance where state='Running'" --context lab
```

For five minutes of CPU history, replace both VM_UUID values with the same verified UUID.

```sh
zstack-cli zql "query vmInstance.uuid where uuid='VM_UUID' return with (zwatch{resultName='cpu',metricName='CPUAverageUsedUtilization',offsetAheadOfCurrentTime=300,period=60,labels='VMUuid=VM_UUID'})" --context lab
```

Read time, value and labels in results[0].returnWith.cpu. A zero value and missing samples have different meanings.
