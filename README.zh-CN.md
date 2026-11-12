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
| PowerShell 模块 | [ZStack PowerCLI](docs/powershell.md) | [ZStack PowerCLI](docs/powershell.zh-CN.md) |
| 操作示例 | [Examples](examples/README.en.md) | [操作示例](examples/README.md) |

[命令参考](docs/reference/commands.md)和[机器可读命令索引](docs/reference/commands.json)由 CLI 帮助自动生成，包含当前源码的命令、参数、别名和默认值。已安装版本的能力可通过对应子命令的 `--help` 核对。

## 安装 CLI

在 [GitHub Releases](https://github.com/terraform-zstack-modules/zstack-cli-go/releases) 下载对应操作系统和架构的压缩包。解压后将二进制加入 PATH，可重命名为 `zstack-cli`，Windows 使用 `zstack-cli.exe`。发布包中的文件名可能带版本和平台，也可直接使用完整文件名调用。

| 操作系统 | 选择的压缩包 |
| --- | --- |
| Linux | `linux_amd64`、`linux_arm64` 或 `linux_arm` |
| macOS Intel | `darwin_amd64` |
| macOS Apple 芯片 | `darwin_arm64` |
| Windows x64 | `windows_amd64.zip` |

### Windows x64 上的 PowerShell 模块

若需要 PowerShell 命令而不是独立 CLI，可在 PowerShell 7.4 或更高版本中从 PowerShell Gallery 安装 [`ZStack.PowerCLI`](https://www.powershellgallery.com/packages/ZStack.PowerCLI)。模块已包含 Windows CLI，无需为模块另行下载 CLI。

```powershell
Install-PSResource -Name ZStack.PowerCLI -Repository PSGallery -Scope CurrentUser
Import-Module ZStack.PowerCLI
```

版本固定、升级、命令和兼容说明见 [PowerShell 指南](docs/powershell.zh-CN.md)。

### 解压后直接运行

进入解压目录，用下面的实际二进制文件执行。示例版本为 v1.1.4，其他版本或架构请替换对应文件名。常规运行和 Skill 安装只需要发布包内的二进制。

**macOS Apple 芯片**

```sh
mv zstack-cli-go_v1.1.4_darwin_arm64 zstack-cli
chmod +x zstack-cli
./zstack-cli version
./zstack-cli skill install --agent codex
./zstack-cli login 10.255.255.21:8080 --username admin --context-name lab --save-password=false
./zstack-cli monitor usage --context lab
```

Intel Mac 将第一行源文件名改为 `zstack-cli-go_v1.1.4_darwin_amd64`。

**Windows x64，PowerShell**

```powershell
Rename-Item ./zstack-cli-go_v1.1.4_windows_amd64.exe zstack-cli.exe
./zstack-cli.exe version
./zstack-cli.exe skill install --agent codex
./zstack-cli.exe login 10.255.255.21:8080 --username admin --context-name lab --save-password=false
./zstack-cli.exe monitor usage --context lab
```

可以在资源管理器中解压 ZIP，然后在该文件夹打开终端。路径包含空格时先用 `Set-Location "C:/path with spaces/zstack-cli"` 进入目录。调用当前目录程序需保留 `./` 前缀。

**Linux x64**

```sh
mv zstack-cli-go_v1.1.4_linux_amd64 zstack-cli
chmod +x zstack-cli
./zstack-cli version
./zstack-cli skill install --agent codex
```

ARM64 使用文件名 `zstack-cli-go_v1.1.4_linux_arm64`。

### 让终端和 agent 找到 CLI

上面的 `./zstack-cli` 可以在解压目录直接使用。若希望在其他目录或由 agent 直接调用 `zstack-cli`，将它所在的目录加入 PATH，或向 agent 提供二进制绝对路径。

macOS/Linux 可执行 `sudo install -m 0755 ./zstack-cli /usr/local/bin/zstack-cli`，然后运行 `zstack-cli version` 核对。Windows 可将 `zstack-cli.exe` 放到固定目录，将该目录添加到用户环境变量 Path，重新打开终端和 agent，再运行 `zstack-cli version`。Skill 安装器会复制 Skill 文件；CLI 的 PATH 配置按上述步骤完成。

### 从源码构建

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
zstack-cli skill install --agent deepcode
zstack-cli skill export --agent workbuddy --file zstack-operations-workbuddy.zip
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
