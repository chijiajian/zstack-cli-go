# ZStack CLI 使用指南

本指南对应当前源码，SDK 版本为 v0.0.10。已发布二进制可能尚未包含全部功能；先用 `zstack-cli version` 和子命令 `--help` 确认。示例中的名称、UUID、地址和文件路径均需替换，未包含真实环境凭据。

完整参数与默认值见[命令参考](reference/commands.md)，可参考输入和组合流程见 [examples](../examples/README.md)，AI 自动执行约定见[agent 指南](agents.md)。

## 1. 构建与登录

在仓库根目录执行 `make build`，生成 `./zstack-cli`。以下示例假设已将该二进制加入 PATH；也可将 `zstack-cli` 替换为 `./zstack-cli`。本地构建需要 Go 1.25 或兼容版本。

```sh
zstack-cli version
zstack-cli --help
zstack-cli login 192.0.2.10:8080 --username admin --context-name lab --save-password=false
```

登录命令在终端提示输入密码，不需要把密码写进命令历史。endpoint 使用 `host` 或 `host:port`，默认端口为 8080；不要传控制台 URL、`http://` 或尾部 `/`。`--save-password=false` 只保存会话，过期后需要重新登录；省略此参数默认保存密码，后续命令可以自动续期。

使用已有 AK/SK：

```sh
zstack-cli login 192.0.2.10:8080 --ak "$LAB_AK" --sk "$LAB_SK" --context-name lab
```

配置默认在 `~/.zstack-cli/config.yaml`，可通过 `ZSTACK_CONFIG` 指定独立文件。配置包含敏感信息，不要提交到仓库。`config view` 仅展示脱敏内容，不能用其输出替代可登录的配置文件。

## 2. 多环境与诊断

```sh
zstack-cli config get-contexts
zstack-cli config use-context lab
zstack-cli config view -o json
zstack-cli get instances --context lab -o json
zstack-cli doctor --context lab --offline -o json
zstack-cli doctor --context lab --request-timeout 15s -o json
```

`--context` 只影响本次命令，不更改默认环境。登录另一个环境但不切换默认值，使用 `login ... --context-name staging --no-switch`。登录使用 `--context-name`，不是 `--context`。

`doctor --offline` 检查本地配置；在线模式额外检查已有会话和一条 zone 查询。两者都不登录、不续期、不修改配置。在线检查通过不代表所有 API 权限或模块许可可用。已有会话过期时，doctor 会失败，即使配置中保存了密码。

```sh
zstack-cli config rename-context staging test
zstack-cli config use-context lab
zstack-cli config delete-context test
zstack-cli logout --context lab
```

删除 context 只删除本地配置。普通 logout 会撤销账户会话后清除凭据；`logout --local` 只清本地凭据。AK/SK logout 始终只清本地，不撤销云端密钥。logout 保留 context 名称及 endpoint。配置锁文件不应手动删除；其他进程切换默认 context 不会改变已运行命令选定的环境。

## 3. 查询、分页与输出

```sh
zstack-cli get instances -q 'state=Stopped' --limit 20 --sort '+uuid' -o json
zstack-cli get instances --all-pages --limit 100 --fields uuid,name,state -o json
zstack-cli get instances --start 20 --limit 20 --reply-with-count -o json
zstack-cli get instances --count
zstack-cli describe instance web-01 -o json
zstack-cli describe volume data-01 --request-timeout 15s -o yaml
```

- 普通 JSON/YAML 查询返回列表，空结果为 `[]`。`--reply-with-count` 返回 `items` 和 `total`；`--count` 只输出总数。
- `--limit` 是单页大小，`--start` 是偏移；`--all-pages` 自动拉取后续页。分页不是快照，资源并发变化可能影响结果。
- `--fields` 限制输出字段，未知字段会报错。字段名以该命令实际 JSON 输出为准，详情结构与列表结构不同。
- 查询通常支持 table、text、json、yaml；具体格式以子命令帮助为准，不能假定所有命令都输出相同结构。
- 资源选择使用 UUID 或精确名称；单资源操作遇到重名会拒绝。已有实例/镜像批量操作可能选择多个同名资源，执行前检查目标列表。
- `describe` 主资源查询失败会返回非零；可选关联读取失败可能保留主资源并返回 `warnings`，自动化需同时检查这些警告。

完整资源列表：`zstack-cli get --help`。`get volumes` 与 `get disks` 共用实现，`get jobs` 可用于长任务查询。

## 4. 等待、恢复与扩容

```sh
zstack-cli wait instance web-01 --for Running --timeout 5m --interval 2s
zstack-cli wait job JOB_UUID --timeout 10m --progress -o json
zstack-cli instance recover VM_UUID
zstack-cli volume recover VOLUME_UUID -o json
zstack-cli volume resize VOLUME_UUID --size 40G --dry-run
zstack-cli volume resize VOLUME_UUID --size 40G -o json
```

wait 只查询，不启动实例或恢复任务。job 默认等待 Succeeded；等待成功时遇到 Failed/Canceled 会失败，显式 `--for Failed` 则表示等待该条件。`--progress` 写 stderr，最终结果写 stdout。

实例恢复要求 Destroyed，数据盘恢复要求 Deleted；根盘随实例恢复，彻底删除后不能恢复。扩容的 `--size` 是目标总容量，只能增大；G 按 1024³ 字节计算，不自动扩展客户机分区或文件系统。recover/resize 的 dry-run 是离线预览，不验证实时状态。

## 5. IAM 管理

| 资源 | 创建 | 更新 | 删除 | 额外操作 |
| --- | --- | --- | --- | --- |
| account、user、role | 支持 | 支持 | 支持 | user 属于当前 context 的账户 |
| policy | 支持 | 不支持 | 支持 | 用 statements 文件定义 |
| organization | 支持 | 支持 | 支持 | 创建需指定 type |
| project | 支持 | 支持 | 支持 | recover、expunge |

准备只含新账户密码的本地文件 `account-password`，建议权限 0600。`--password-file` 用于 IAM 创建/更新，login 没有该参数。读取时会去掉末尾一个换行，发送前进行 SHA-512 哈希；预览和返回结果不展示密码或哈希。

```sh
zstack-cli create account test-account --password-file ./account-password --dry-run -o json
zstack-cli create account test-account --password-file ./account-password -o json
zstack-cli update account test-account --description 'CLI test account'
zstack-cli update account test-account --description ''
```

显式空 description 用于清空描述。策略文件 `statements.json` 示例：

```json
[{"name":"query-account","effect":"Allow","actions":["org.zstack.header.identity.APIQueryAccountMsg"]}]
```

```sh
zstack-cli create policy query-only --statements-file ./statements.json
zstack-cli create role operator --policy query-only
zstack-cli create organization test-department --type Department
zstack-cli create project test-project --organization test-department
zstack-cli update project test-project --description 'Temporary test project'
```

上述策略仅表达一条权限规则，创建 policy/role 不会自动给其他账户授予权限。IAM 创建使用参数和 `--statements-file`，不支持通用 `create --file`。IAM2 组织/项目需要服务器许可。引用名称同样采用精确解析；需要固定创建身份时使用 32 位十六进制 `--resource-uuid`。IAM 请求默认超时 2 分钟，可用 `--request-timeout` 调整。

项目删除、恢复与最终清理的顺序：

```sh
zstack-cli delete project test-project --yes
zstack-cli project recover test-project --yes
# 再次删除，之后的 expunge 是不可恢复的彻底清理。
zstack-cli delete project test-project --yes
zstack-cli expunge project test-project --yes
```

其他 IAM 删除示例：`zstack-cli delete account test-account`。不加 `--yes` 时交互确认；自动化使用 `--yes` 前应已明确目标。删除模式默认 Permissive，可显式指定 `--delete-mode Enforcing`，按服务器语义处理依赖。

## 6. 长任务操作

```sh
zstack-cli get jobs -o json
zstack-cli describe job JOB_UUID -o json
zstack-cli job resume JOB_UUID --dry-run -o json
zstack-cli job resume JOB_UUID --yes --request-timeout 2m -o json
zstack-cli job clean JOB_UUID --yes -o json
```

resume 仅接受 Suspended；clean 仅接受 Succeeded、Failed、Canceled。dry-run 不核验实时状态，实际执行会解析并检查状态。resume 返回不代表任务已经成功，可继续 `wait job`。本项目已用本地测试验证状态保护和请求契约，未在真实环境操作既有 long job。

## 7. 自动化错误处理

```sh
if zstack-cli get instances -o json --error-format json > result.json 2> error.json; then
  cat result.json
else
  status=$?
  printf 'CLI exit: %s\n' "$status"
  cat error.json
fi
```

`--error-format json` 控制 stderr 的错误诊断，`-o json` 控制 stdout 的业务结果。错误包含 `error.code`、`message`、`exitCode`，可能包含 `httpStatus`、`apiCode`。不要把进度提示或交互确认与机器解析混用。

| 退出码 | 含义 |
| --- | --- |
| 0 | 成功 |
| 1 | 未分类操作失败或部分批量失败 |
| 2 | 参数错误 |
| 3 | 认证或权限错误 |
| 4 | 资源/context 不存在 |
| 5 | 状态冲突或资源选择歧义 |
| 6 | 超时 |
| 7 | 网络连接失败 |
| 8 | API 或响应契约错误 |
| 9 | 输出/编码失败 |
| 130 | 操作取消 |

写请求超时、输出失败或返回身份不匹配，都不等于服务器回滚。先按已知 UUID 查询资源，再决定下一步，不要直接重试创建。JSON 诊断中的错误信息会脱敏，但业务查询结果仍应按业务数据管理。

## 8. 回归脚本与清理

本节命令在仓库根目录执行，需要 Python 3。先用独立配置登录，不保存账户密码：

```sh
export ZSTACK_CONFIG="$PWD/session.yaml"
zstack-cli login 192.0.2.10:8080 --username admin --context-name qa --save-password=false
python3 scripts/readonly_smoke.py --cli ./zstack-cli --config "$ZSTACK_CONFIG" > readonly-report.json
```

两个脚本均复制配置到私有临时目录，清除继承的 ZSTACK 环境变量，不修改源配置。使用传入文件中的当前 context，不接受脚本级 `--context`；运行前用 CLI 选定 context。配置可以使用 session-only 或完整 AK/SK，保存账户密码的配置会被拒绝。

只读脚本覆盖 42 类规范查询命令。`--sample-size` 为 1–3，单命令 `--timeout` 默认 30 秒、范围 1–120；`--suite-timeout` 默认 600 秒、范围 1–3600。空集合及不可用关联的后续检查记为 skipped，不代表覆盖通过。global-configs 使用数值 id 验证分页。资源变动可能造成比较失败，应结合稳定样本排查。

写脚本会创建和删除测试资源，仅在有配额、许可的测试环境运行。默认 metadata 包含 volume、account、user、policy、role、organization、project：

```sh
python3 scripts/write_e2e.py --cli ./zstack-cli --config "$ZSTACK_CONFIG" \
  --journal ./metadata-run.json --execute
python3 scripts/write_e2e.py --cli ./zstack-cli --config "$ZSTACK_CONFIG" \
  --journal ./compute-run.json --execute --profile compute \
  --image ROOT_IMAGE_UUID --l3-network L3_UUID --host HOST_UUID
```

每次运行必须指定新的 journal 路径。`--resources account,user,policy,role` 可缩小 metadata 范围；`--profile all` 同时运行 metadata 和 compute，需提供计算所需 UUID。compute 创建停止状态的 1 CPU/1 GiB 实例和 1 GiB 数据盘，验证挂载、快照、扩容至 2 GiB、解绑及恢复，结束时清理并检查根盘。恢复测试必须观察删除状态和恢复后的状态，缺失状态字段不会算成功。

写脚本 `--timeout` 默认 180 秒、范围 1–600；`--suite-timeout` 默认 1200 秒、范围 1–3600。清理不受主流程剩余时间限制，每条清理命令单独限时，因此总运行时间可能超过 suite-timeout。

中断或失败后保留 journal，待 API 作业稳定后重跑清理：

```sh
python3 scripts/write_e2e.py --cli ./zstack-cli --config "$ZSTACK_CONFIG" \
  --journal ./compute-run.json --execute --cleanup-only
zstack-cli logout
```

cleanup-only 从 journal 读取资源，不需要重新传 image/network。必须连接相同 endpoint；会话过期可重新登录到该环境后再清理。脚本核对本次生成的名称和 UUID，只解绑挂在本次实例上的云盘，拒绝操作身份不匹配的资源。未知创建结果或根盘残留不会标记 cleaned；不要修改 journal 来绕过检查，也不要将同一 journal 并发交给多个运行进程。

报告的 `passed`、`failure`、`cleanup_errors` 与进程退出码一起判断本次是否通过；主测试失败但清理成功仍算失败。源会话不会被脚本注销，结束后由调用方 logout。SIGKILL 或掉电无法执行退出清理，需要保留 journal 恢复处理。

本地开发检查用 `make verify`（vet、Go race、Python 回归和构建），无需云凭据，也不会执行真实环境脚本。历史实测范围和跳过项见[交付清单](next-stage-checklist.md)。
