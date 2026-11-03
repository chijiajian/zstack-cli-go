# ZStack CLI

[English](README.md) | [简体中文](README.zh-CN.md)

A Go CLI for ZStack compute, storage, networking and IAM operations. It provides resource queries, lifecycle actions, capacity and load reports, ZQL, multiple contexts and an installable AI operations Skill.

## Documentation

| Topic | English | 简体中文 |
| --- | --- | --- |
| Documentation index | [Index](docs/README.md) | [目录](docs/README.zh-CN.md) |
| User manual | [Manual](docs/usage-en.md) | [使用手册](docs/usage-zh.md) |
| Agent execution | [Guide](docs/agents.en.md) | [执行指南](docs/agents.md) |
| Skill installation | [Installation](skills/README.en.md) | [安装说明](skills/README.md) |
| Workflow examples | [Examples](examples/README.en.md) | [操作示例](examples/README.md) |

[Command reference](docs/reference/commands.md) and [machine-readable command index](docs/reference/commands.json) are generated from CLI help. They list the current source tree's commands, flags, aliases and defaults. Use your installed binary's `--help` to confirm available features.

## Install the CLI

Download an archive matching your OS and architecture from [GitHub Releases](https://github.com/chijiajian/zstack-cli-go/releases). Extract the binary and put it on PATH as `zstack-cli` (`zstack-cli.exe` on Windows). Release binaries may have the version and platform in their filenames, so rename them or use the full filename.

| OS | Archive selection |
| --- | --- |
| Linux | `linux_amd64`, `linux_arm64` or `linux_arm` |
| macOS Intel | `darwin_amd64` |
| macOS Apple silicon | `darwin_arm64` |
| Windows x64 | `windows_amd64.zip` |

To build from source, install Go 1.25 or compatible and run the following from the repository root.

```sh
go build -o zstack-cli .
```

On Windows use `go build -o zstack-cli.exe .`. PowerShell needs `./zstack-cli.exe` to run an executable in the current directory.

## Quick start

```sh
zstack-cli version
zstack-cli login 192.0.2.10:8080 --username admin --context-name lab --save-password=false
zstack-cli get instances --context lab -o json
zstack-cli monitor usage --context lab
zstack-cli zql "count vmInstance where state='Running'" --context lab
```

Login prompts for the password. Endpoints use `host` or `host:port`, with API port 8080 as the default. Configuration is stored in `~/.zstack-cli/config.yaml`, or at the path selected by `ZSTACK_CONFIG`. The manual explains AK/SK login, session renewal, context selection and logout.

## Install the AI operations Skill

The CLI embeds the Skill and installs it using local filesystem operations on Windows, macOS and Linux.

```sh
zstack-cli skill install --agent codex
zstack-cli skill install --agent claude-code
zstack-cli skill install --agent deepcode
zstack-cli skill export --agent workbuddy --file zstack-operations-workbuddy.zip
zstack-cli skill install --agent cursor --scope project --project "/path/to/project"
zstack-cli skill install --agent copilot --dry-run -o json
```

Use `--replace` to update an existing installation and retain a backup. `--dest` overrides the parent Skill directory. Refresh the agent or start a new session, then ask it to use `zstack-operations`. Python is required only for the optional capability-check helper. See the bilingual installation guide above for agent paths, Windows examples and manual installation.

## Development

```sh
make verify       # vet, Go race tests, Python runner tests and build
make docs-check   # generated help consistency
make skill-check  # Skill capability checks
make skill-package
```

`make verify` requires a C compiler and Python 3. Native Windows CI checks Skill installation. Cloud regression scripts are opt-in and documented in the user manual. Repository development instructions are in [AGENTS.md](AGENTS.md).

## License

[Apache License 2.0](LICENSE).
