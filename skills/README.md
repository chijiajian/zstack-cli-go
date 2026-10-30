# ZStack CLI Skills

这里是随 CLI 源码一起维护的 Skill 分发源。[zstack-operations](zstack-operations/SKILL.md) 指导 agent 使用已安装的 CLI 操作 ZStack；根目录 AGENTS.md 指导 agent 开发本仓库。两者用途不同。

执行路径为 Agent + Skill → zstack-cli → ZStack API，不要求 MCP。Skill 目录可独立复制；不复制整个 CLI 源码或配置文件。

## 安装

先安装可信的 zstack-cli 二进制。Skill 默认通过 PATH 查找，也可使用明确的绝对路径。可选兼容检查脚本需要 Python 3 标准库，正常操作只需要 agent 有终端执行能力。

从仓库根目录安装到支持 `.agents/skills` 的 agent 用户目录（POSIX shell）：

```sh
mkdir -p "$HOME/.agents/skills"
test ! -e "$HOME/.agents/skills/zstack-operations" && \
  cp -R skills/zstack-operations "$HOME/.agents/skills/zstack-operations"
```

已有同名目录时不会覆盖；更新前检查并备份本地修改，再替换整个目录。项目级安装时，将整个目录放到目标项目的 `.agents/skills/zstack-operations`。其他客户端按其 Skill 目录约定安装同一份内容；不要只复制 SKILL.md。Windows 可通过文件管理器复制整个目录到客户端支持的 Skill 路径。

本仓库不同时提交 `.agents/skills/` 副本，避免维护两份来源，也不自动写入用户目录。安装后由客户端刷新或新建会话发现 Skill，可显式请求：

> 使用 $zstack-operations，只查询 lab context 中停止的实例，返回 UUID、名称和状态，不修改资源或续期会话。

兼容检查命令（替换二进制路径）：

```sh
python3 "$HOME/.agents/skills/zstack-operations/scripts/check_cli.py" \
  --cli /absolute/path/zstack-cli --profile query
```

还可使用 lifecycle、iam profile。脚本只执行 version/help，过滤 ZSTACK 环境变量，使用临时空配置；成功不代表云端认证、许可或操作状态有效。Skill 1.0.0 对应当前源码能力，没有虚构最低 CLI release tag；详见包内 references/compatibility.md。

## 打包与维护

```sh
make skill-check
make skill-package
```

`skill-check` 构建 CLI，运行三个本地能力 profile 与辅助脚本回归；不连接云端。`skill-package` 生成 `dist/zstack-operations-skill.tar.gz`，顶层为 zstack-operations 目录，不包含凭据、源码仓库文档或 Python 缓存。解压到临时目录检查后，再复制到目标 Skill 目录。

新增 CLI 行为时同步修改本 Skill 参考和必要的兼容检查。Skill 与 CLI 在同一 Git tag 下发布，但 Skill 的 metadata.version 独立表达其内容版本。不要因帮助检查通过宣称真实操作已测试。
