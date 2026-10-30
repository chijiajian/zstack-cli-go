# 回归运行

在仓库根目录执行。session.yaml 是专用配置，不纳入版本控制，账户登录须 `--save-password=false`。脚本使用该文件的当前 context，运行前核对 endpoint。

```sh
make build
python3 scripts/readonly_smoke.py --cli ./zstack-cli --config ./session.yaml > readonly-report.json
make verify
```

`make verify` 不访问真实云。以下显式创建/删除测试资源，需要独立测试环境和授权；每次创建流程使用新的 journal 文件：

```sh
python3 scripts/write_e2e.py --cli ./zstack-cli --config ./session.yaml --journal ./metadata-run.json --execute
python3 scripts/write_e2e.py --cli ./zstack-cli --config ./session.yaml --journal ./compute-run.json --execute --profile compute --image ROOT_IMAGE_UUID --l3-network L3_UUID
```

失败时先保留证据并等待云端异步请求稳定，再清理同一 journal：

```sh
python3 scripts/write_e2e.py --cli ./zstack-cli --config ./session.yaml --journal ./compute-run.json --execute --cleanup-only
```

不得并发复用 journal。所有权不匹配、未知创建结果、根盘残留均不算清理通过。不要编辑状态绕过检查。脚本不注销输入会话，验证结束后通过对应配置执行 logout。
