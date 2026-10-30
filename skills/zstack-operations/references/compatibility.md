# 兼容性与发现

Skill 版本 1.0.0，与提供此 Skill 的 CLI 源码版本一起维护。当前验证对象是使用 SDK v0.0.10 的源码构建；SDK 版本不是 CLI 版本。尚未据此声明某个已发布 CLI tag 为最低兼容版本。源码构建可能显示 `dev`，不能只比较版本号判断兼容。

以目标命令实际 `--help` 为准。辅助脚本只验证代表性命令及参数是否存在，不验证 JSON 运行时契约、认证、网络、许可、云端状态或所有命令覆盖：

```sh
python3 /path/to/zstack-operations/scripts/check_cli.py --cli /path/to/zstack-cli --profile query
python3 /path/to/zstack-operations/scripts/check_cli.py --cli /path/to/zstack-cli --profile lifecycle
python3 /path/to/zstack-operations/scripts/check_cli.py --cli /path/to/zstack-cli --profile iam
```

query 检查基础查询、详情、context、doctor、结构化错误；其他 profile 在其上增加相应操作。脚本退出 0 仅表示该 profile 的帮助能力匹配，退出 1 表示不能确认兼容。缺少一个不用于本次任务的功能，不应阻止已确认可用的独立功能；报告限制即可。

没有 Python 时手动运行 `version` 和目标命令的 `--help`。不要把旧版未知参数错误改成静默忽略，也不要为了通过检查自动升级或切换环境。

Skill 所有参考文件与脚本位于本目录内。安装时复制整个目录，不能只复制 SKILL.md。不依赖源码仓库内 docs、examples、.env 或历史验收记录。
