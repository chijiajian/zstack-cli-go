# PowerShell 子进程调用

用 `scripts/check_cli.py --profile powershell` 或子命令的 `--help` 核对当前二进制。帮助检查不执行登录或写请求，也不能证明已发布模块可用。

- `login --username USER --password-stdin` 从私有 stdin 读取 UTF-8 密码字节，读到 EOF。不要加入换行，不在命令参数、环境变量或日志中放入密码。使用专用配置和 `--context-name NAME --no-switch`。此模式只保存会话。
- `login --access-key-id KEY_ID --access-key-secret-stdin` 从私有 stdin 读取 UTF-8 SK，读到 EOF。AK/SK 写入所选配置，配置必须放在受保护的目录内。PowerShell 模块用独立私有目录，并在断开时清理。
- 获授权且核对 VM UUID 后可执行 `instance start VM_UUID --uuid-only --yes -o json --error-format json`。`--uuid-only` 不回退到名称匹配。成功流是 `operation` 与 `items`，逐项检查 `status` 和退出码。需要预览时使用 `--dry-run`，可能查询目标，且不会启动。超时后先查状态，避免重复写请求。
- `get instances -o json` 提供原始字节字段 `memorySizeBytes`，未报告时为 `null`。不要从展示字段 `memorySize` 反算容量。
