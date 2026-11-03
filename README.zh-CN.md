# ZStack CLI

[English](README.md) | [简体中文](README.zh-CN.md)

使用 Go 开发的 ZStack 命令行工具，支持计算、存储、网络和 IAM 操作，提供资源查询、生命周期操作、容量与负载监控、ZQL、多 context，以及可安装的 AI 运维 Skill。

## 文档导航

| 主题 | English | 简体中文 |
| --- | --- | --- |
| 文档目录 | [Index](docs/README.md) | [目录](docs/README.zh-CN.md) |
| 使用手册 | [Manual](docs/usage-en.md) | [使用手册](docs/usage-zh.md) |
| Agent 执行 | [Guide](docs/agents.en.md) | [执行指南](docs/agents.md) |
| Skill 安装 | [Installation](skills/README.en.md) | [安装说明](skills/README.md) |
| 操作示例 | [Examples](examples/README.en.md) | [操作示例](examples/README.md) |

[命令参考](docs/reference/commands.md)和[机器可读命令索引](docs/reference/commands.json)由 CLI 帮助自动生成，包含当前源码的命令、参数、别名和默认值。已安装版本的能力可通过对应子命令的 `--help` 核对。

## 安装 CLI

在 [GitHub Releases](https://github.com/chijiajian/zstack-cli-go/releases) 下载对应操作系统和架构的压缩包。解压后将二进制加入 PATH，可重命名为 `zstack-cli`，Windows 使用 `zstack-cli.exe`。发布包中的文件名可能带版本和平台，也可直接使用完整文件名调用。

| 操作系统 | 选择的压缩包 |
| --- | --- |
| Linux | `linux_amd64`、`linux_arm64` 或 `linux_arm` |
| macOS Intel | `darwin_amd64` |
| macOS Apple 芯片 | `darwin_arm64` |
| Windows x64 | `windows_amd64.zip` |

从源码构建时，需要 Go 1.25 或兼容版本，在仓库根目录运行以下命令。

```sh
go build -o zstack-cli .
```

Windows 使用 `go build -o zstack-cli.exe .`。PowerShell 调用当前目录中的程序时使用 `./zstack-cli.exe`。

## 快速开始

```sh
zstack-cli version
zstack-cli login 10.255.255.21:8080 --username admin --context-name lab --save-password=false
zstack-cli get instances --context lab -o json
zstack-cli monitor usage --context lab
zstack-cli zql "count vmInstance where state='Running'" --context lab
```

登录时终端会提示输入密码。endpoint 使用 `host` 或 `host:port`，API 默认端口为 8080。配置保存在 `~/.zstack-cli/config.yaml`，也可用 `ZSTACK_CONFIG` 指定路径。AK/SK 登录、会话续期、context 选择和注销见使用手册。

## 安装 AI 运维 Skill

CLI 二进制内置完整 Skill，通过本地文件操作安装，支持 Windows、macOS 和 Linux。

```sh
zstack-cli skill install --agent codex
zstack-cli skill install --agent claude-code
zstack-cli skill install --agent cursor --scope project --project "/path/to/project"
zstack-cli skill install --agent copilot --dry-run -o json
```

更新已有安装时添加 `--replace`，安装器会保留旧目录备份。`--dest` 用于指定 Skill 父目录。安装后刷新 agent 或新建会话，要求使用 `zstack-operations`。Python 仅用于可选的能力检查脚本。各 agent 路径、Windows 示例及手动安装方式见上面的双语安装说明。

## 开发验证

```sh
make verify       # vet、Go race、Python 回归和构建
make docs-check   # 自动生成帮助的一致性
make skill-check  # Skill 能力检查
make skill-package
```

`make verify` 需要 C 编译器和 Python 3。CI 在 Windows 上原生测试 Skill 安装。真实环境回归脚本需要显式调用，操作方式见使用手册。仓库开发约定见 [AGENTS.md](AGENTS.md)。

## 许可证

[Apache License 2.0](LICENSE)。
