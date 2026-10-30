# 查询与自动化

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
