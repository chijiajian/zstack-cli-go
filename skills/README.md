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
| Deep Code | `.agents/skills` | `.deepcode/skills` |

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

## WorkBuddy、DeepSeek 与其他客户端

deepcode、custom 安装目标和 WorkBuddy ZIP 导出从 CLI v1.1.2 起提供，v1.1.1 二进制尚未包含这些入口。使用前先检查 `skill install --help` 与 `skill export --help`。

WorkBuddy 支持导入本地技能包。用 CLI 导出带 WorkBuddy 展示元数据的 ZIP，然后在 WorkBuddy 的技能页面选择“添加技能 → 上传技能”导入，并按客户端提示完成检查和启用。

```sh
zstack-cli skill export --agent workbuddy --file zstack-operations-workbuddy.zip
```

导出包含完整 SKILL.md、references 和可选检查脚本，目标 ZIP 必须是新文件。客户执行云操作时，WorkBuddy 需要能在同一执行环境运行 zstack-cli，并能连接对应 ZStack endpoint。配置凭据通过 CLI 登录。

DeepSeek 模型可通过不同 agent 客户端使用，Skill 的安装位置取决于客户端。DeepSeek 官方集成指南列出的 Deep Code 可这样安装。

```sh
zstack-cli skill install --agent deepcode
zstack-cli skill install --agent deepcode --scope project --project "/path/to/project"
```

Deep Code 用户目录为 `.agents/skills`，项目目录为 `.deepcode/skills`。若用 DeepSeek 模型驱动 Codex 或 Claude Code，继续选择对应的 codex 或 claude-code 安装目标。其他支持本地 Skill 和终端执行的客户端，可指定其实际 Skill 父目录。

```sh
zstack-cli skill install --agent custom --dest "/path/to/agent/skills"
```

DeepSeek 网页聊天中的模型对话本身不足以保证加载本地 Skill 或执行 CLI，需要配套客户端提供这些能力。WorkBuddy 导入及 Deep Code 目录按官方文档适配，目前验证范围为 ZIP 结构、元数据和本地安装，实际客户端的导入与端到端操作待验证。

来源见 [WorkBuddy 技能说明](https://www.workbuddy.ai/docs/workbuddy/From-Beginner-to-Expert-Guide/Function-Description/Skills)、[WorkBuddy 包结构](https://open.workbuddy.cn/docs/skill)与 [DeepSeek 的 Deep Code 集成指南](https://github.com/deepseek-ai/awesome-deepseek-agent/blob/main/docs/deepcode.md)。

## 打包与维护

```sh
make skill-check
make skill-package
```

`skill-check` 构建 CLI，运行五个本地能力 profile 与辅助脚本回归；不连接云端。`skill-package` 生成 `dist/zstack-operations-skill.tar.gz`，顶层为 zstack-operations 目录，不包含凭据、源码仓库文档或 Python 缓存。解压到临时目录检查后，再复制到目标 Skill 目录。

新增 CLI 行为时同步修改本 Skill 参考和必要的兼容检查。Skill 与 CLI 在同一 Git tag 下发布，但 Skill 的 metadata.version 独立表达其内容版本。不要因帮助检查通过宣称真实操作已测试。
