# zstack-sdk-go-v2 迁移计划

状态：已完成

完成日期：2026-06-14

实际迁移版本：

```text
github.com/zstackio/zstack-sdk-go-v2 v0.0.8
```

当前 `go.mod` 未保留本地 `replace`。

## 2026-10-29 增量升级

当前依赖升级至 `v0.0.10`，上文 v0.0.8 为历史迁移基线。
已通过远程 tags 核实 v0.0.10 为本次检查时的最新 tag，指向提交
`d86f3179eeeeff5922707fb6abeb02b615236c8e`。

- 适配 `DetachDataVolumeFromVm` 新参数与双返回值，并验证指定实例的解绑参数。
- 保留可取消 HTTP 读取适配层：v0.0.10 的 HTTP 实现仍未传递请求 context。
- 新增网络查询、实例/云盘详情和统一分页输出；回归测试使用本地 HTTP mock。
- 此次升级未进行真实环境 E2E。

## 完成记录（v0.0.8）

已完成：

- 依赖切换到 `github.com/zstackio/zstack-sdk-go-v2 v0.0.8`
- 移除旧 SDK import
- 适配查询和分页的 `*param.QueryParam` 签名
- 适配 `Login(context.Context)` / `Logout(context.Context)`
- 修正 `QueryInstaceOffering` 为 `QueryInstanceOffering`
- 适配 v2 view 字段命名差异
- 适配 create 参数结构差异
- 适配实例生命周期操作参数
- 修复 `get clusters` 直接打印 SDK 嵌入结构的问题

已通过自动化验证：

```bash
go test -count=1 ./...
go vet ./...
go build ./...
```

已通过真实环境只读 smoke：

```bash
zstack-cli login 172.24.248.4 -u admin -p ...
zstack-cli get zones -l 3
zstack-cli get clusters -l 3
zstack-cli get hosts -l 3
zstack-cli get images -l 3
zstack-cli get instances -l 3
zstack-cli get l3-networks -l 3
```

已通过真实环境写操作 E2E：

```text
create instance --strategy CreateStopped
instance start
instance restart
instance pause
instance resume
instance stop
delete instance
expunge instance
```

E2E 临时 VM 已清理，expunge 后查询不可见。

后续 CLI 能力建设请见：

```text
docs/sdk-v2-cli-roadmap.md
```

## 目标

将 CLI 当前依赖的旧 SDK：

```text
github.com/zstackio/zstack-sdk-go-v2 v1.0.0
```

迁移到新 SDK：

```text
github.com/zstackio/zstack-sdk-go-v2 v0.0.8
```

本地 SDK 仓库路径：

```text
/Users/michelvaillant/Developer/zstack.io/zstack-sdk-go-v2
```

本地确认到的最新 tag：

```text
v0.0.8
```

## 迁移步骤

1. 更新依赖

   在 `go.mod` 中移除旧 SDK：

   ```text
   github.com/zstackio/zstack-sdk-go-v2 v1.0.0
   ```

   增加新 SDK：

   ```text
   github.com/zstackio/zstack-sdk-go-v2 v0.0.8
   ```

   本地开发验证阶段建议临时增加：

   ```text
   replace github.com/zstackio/zstack-sdk-go-v2 => /Users/michelvaillant/Developer/zstack.io/zstack-sdk-go-v2
   ```

   注意：绝对路径 `replace` 不适合提交到公共仓库，最终是否保留需要单独决定。

2. 替换 import path

   将所有旧 SDK import：

   ```text
   github.com/zstackio/zstack-sdk-go-v2/pkg/client
   github.com/zstackio/zstack-sdk-go-v2/pkg/param
   github.com/zstackio/zstack-sdk-go-v2/pkg/view
   ```

   替换为：

   ```text
   github.com/zstackio/zstack-sdk-go-v2/pkg/client
   github.com/zstackio/zstack-sdk-go-v2/pkg/param
   github.com/zstackio/zstack-sdk-go-v2/pkg/view
   ```

3. 适配查询和分页签名

   v2 的查询/分页方法主要接收 `*param.QueryParam`。

   当前 CLI 里很多调用是：

   ```go
   zsClient.QueryImage(*queryParam)
   zsClient.PageVmInstance(*queryParam)
   ```

   需要改为：

   ```go
   zsClient.QueryImage(queryParam)
   zsClient.PageVmInstance(queryParam)
   ```

4. 适配登录和登出签名

   旧 SDK：

   ```go
   session, err := client.Login()
   err := client.Logout()
   ```

   v2 SDK：

   ```go
   session, err := client.Login(context.Background())
   err := client.Logout(context.Background())
   ```

5. 修正方法名差异

   旧 SDK 中存在拼写错误的方法：

   ```go
   QueryInstaceOffering
   ```

   v2 中应使用：

   ```go
   QueryInstanceOffering
   ```

6. 检查 view 字段命名差异

   重点检查常用结构体字段是否从 `UUID` 变为 `Uuid`，或存在其他字段名变化。

   优先检查这些路径：

   - `pkg/client/resource_helpers.go`
   - `cmd/get/*.go`
   - `cmd/create/*.go`
   - `cmd/resources/*.go`
   - `pkg/utils/vm_output.go`

7. 编译驱动适配

   先执行：

   ```bash
   go build ./...
   ```

   按编译错误逐个修复签名和字段差异。编译通过后再执行测试和 vet。

## 验收范围

必须通过：

```bash
go test ./...
go vet ./...
go build ./...
```

需要手动抽查的 CLI 路径：

```bash
zstack-cli --help
zstack-cli --version
zstack-cli login <endpoint> -u <user> -p <password>
zstack-cli get images
zstack-cli get instances
zstack-cli get zones
zstack-cli get clusters
zstack-cli get hosts
zstack-cli get l3-networks
zstack-cli create image ...
zstack-cli create instance ...
zstack-cli create disk-offering ...
zstack-cli create instance-offering ...
zstack-cli delete image ...
zstack-cli delete instance ...
zstack-cli expunge image ...
zstack-cli expunge instance ...
zstack-cli instance start ...
zstack-cli instance stop ...
zstack-cli instance pause ...
zstack-cli instance resume ...
zstack-cli instance restart ...
```

## 风险点

- `zstack-sdk-go-v2` 的 `go.mod` 声明 `go 1.25`，当前 CLI 项目之前使用的工具链是 Go 1.24.6。迁移时可能需要安装或自动下载 Go 1.25 工具链。
- v2 SDK 的查询方法接收指针参数，当前 CLI 大量旧调用需要机械替换。
- v2 SDK 的部分 view 字段可能与旧 SDK 命名不一致，编译错误能暴露大部分，但表格输出还需要手动检查。
- 如果最终移除本地 `replace`，CI 或其他开发环境需要能从 GitHub 正常拉取 `github.com/zstackio/zstack-sdk-go-v2 v0.0.8`。
- SDK 升级后登录、创建、删除、expunge 等写操作需要在真实或测试 ZStack 环境中做一次端到端验证。

## 建议执行顺序

1. 使用本地 `replace` 完成依赖迁移和编译适配。
2. 运行 `go build ./...`，修复所有编译错误。
3. 运行 `go test ./...` 和 `go vet ./...`。
4. 在无真实环境下先完成 CLI 帮助、参数校验、未登录错误路径测试。
5. 在 ZStack 测试环境中验证登录、查询、创建、删除、expunge、实例生命周期操作。
6. 决定是否移除本地 `replace`，并更新 `go.sum`。
