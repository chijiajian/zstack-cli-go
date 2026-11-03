# ZStack CLI Skills

[English](README.en.md) | [简体中文](README.md)

这里是随 CLI 源码一起维护的 Skill 分发源。[zstack-operations](zstack-operations/SKILL.md) 指导 agent 使用已安装的 CLI 操作 ZStack；根目录 AGENTS.md 指导 agent 开发本仓库。两者用途不同。

执行路径为 Agent + Skill → zstack-cli → ZStack API，不要求 MCP。Skill 目录可独立复制；不复制整个 CLI 源码或配置文件。

## 安装

先安装可信的 zstack-cli 二进制。Skill 默认通过 PATH 查找，也可使用明确的绝对路径。可选兼容检查脚本需要 Python 3 标准库，正常操作只需要 agent 有终端执行能力。

使用 CLI 内置安装器，Windows、macOS 和 Linux 共用这些命令。先把发布包里的二进制加入 PATH，Windows 也可使用 `./zstack-cli.exe` 的实际文件名。

```sh
zstack-cli skill install --agent codex
zstack-cli skill install --agent claude-code
zstack-cli skill install --agent cursor --scope project --project /path/to/project
zstack-cli skill install --agent copilot --dry-run -o json
zstack-cli skill install --agent codex --replace
```

PowerShell 中显式调用当前目录的程序时使用 `./` 或 `.\` 前缀。项目路径含空格时加引号，Windows 项目路径可以写成 `"C:/work/my project"`。

| Agent | 用户级目录（相对用户主目录） | 项目级目录 |
| --- | --- | --- |
| Codex | `.agents/skills` | `.agents/skills` |
| Claude Code | `.claude/skills` | `.claude/skills` |
| Cursor | `.cursor/skills` | `.cursor/skills` |
| GitHub Copilot | `.copilot/skills` | `.github/skills` |

默认 scope 为 user。project scope 必须给出已有项目目录。`--dest` 可指定其他 agent 的 Skill 父目录，安装器会在其下创建 `zstack-operations`。agent 的发现路径取决于客户端版本，可按实际配置覆盖。

Skill 嵌入 CLI 二进制，安装时自动复制完整目录。安装过程使用本地文件系统。Python 仅用于可选的 `check_cli.py`，安装和正常使用不需要 Python。已有目录需要显式 `--replace`，更新时将原目录整体保留到同级 backup 目录，返回结果中包含备份位置。用 `--dry-run` 可先检查目标路径。

手动安装也可复制整个 `skills/zstack-operations` 到上表目录，或从 Release 下载独立 Skill 压缩包后解压复制。Windows 可以使用资源管理器完成复制。替换前先保留自己的定制内容。

本仓库不同时提交 `.agents/skills/` 副本，避免维护两份来源，也不自动写入用户目录。安装后由客户端刷新或新建会话发现 Skill，可显式请求：

> 使用 $zstack-operations，只查询 lab context 中停止的实例，返回 UUID、名称和状态，不修改资源或续期会话。

兼容检查命令（替换二进制路径）：

```sh
python3 "$HOME/.agents/skills/zstack-operations/scripts/check_cli.py" \
  --cli /absolute/path/zstack-cli --profile query
```

还可使用 zql、monitor、lifecycle、iam profile。脚本只执行 version/help，过滤 ZSTACK 环境变量，使用临时空配置；成功不代表云端认证、许可或操作状态有效。Skill 1.2.0 对应当前源码能力，没有虚构最低 CLI release tag；详见包内 references/compatibility.md。

## 打包与维护

```sh
make skill-check
make skill-package
```

`skill-check` 构建 CLI，运行五个本地能力 profile 与辅助脚本回归；不连接云端。`skill-package` 生成 `dist/zstack-operations-skill.tar.gz`，顶层为 zstack-operations 目录，不包含凭据、源码仓库文档或 Python 缓存。解压到临时目录检查后，再复制到目标 Skill 目录。

新增 CLI 行为时同步修改本 Skill 参考和必要的兼容检查。Skill 与 CLI 在同一 Git tag 下发布，但 Skill 的 metadata.version 独立表达其内容版本。不要因帮助检查通过宣称真实操作已测试。
