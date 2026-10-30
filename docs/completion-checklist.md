# SDK CLI 开发收尾清单

范围：完成 sdk-v2-cli-roadmap.md 已明确的主干能力及一致性收尾。SDK 最新 tag
已核实为 v0.0.10。IAM 写操作在原路线图中仍待产品定位，不纳入本轮；IAM 查询
已实现。真实云资源变更 E2E、推送后的远程 CI 和发布不在当前只读授权内。

- [x] 核对主干命令并增加注册覆盖，纠正过期路线图说明。
- [x] 补齐本地主存储 add-local，包含离线预览、SDK 请求适配与回归。
- [x] 补齐 legacy delete/expunge 的 --yes 及 snapshot revert 确认，保证输入/输出错误不触发变更。
- [x] 创建结果输出统一返回错误，覆盖 JSON/YAML/wide/name/text。
- [x] 增加可取消的显式会话退出，隔离 context，避免自动登录与覆盖并发会话。
- [x] 本地统一验收（make verify、工作流校验、Windows 编译）。
- [x] 修复显式登录错误中 SDK 打印密码摘要的问题，复用可取消 HTTP 适配层并补回归。
- [x] 真实只读 smoke 与 logout 实测。

本文件记录本轮明确的开发交付，不把未授权的真实写 E2E 描述为已验证。

## 最终验收记录

网络恢复后确认凭据有效。此前 CLI 登录失败包含两处兼容性问题，均已修复并补测试：

- 登录和续期仅发送必要字段，避免 SDK 空 `accountType` 等可选字段导致认证失败。
- 显式登录仅解码保存所需的会话身份字段，不让服务端本地化日期格式影响登录成功结果。

- `make verify`：通过（全量 Go 竞态测试、vet、23 项 Python 测试、构建）。
- 工作流 actionlint、Windows CLI 与配置测试交叉编译：通过。
- 172.31.29.11 真实只读 smoke：68 passed、0 failed、1 skipped。
- 跳过项为 global-configs 分页：当前输出无支持的稳定身份标识，未计作通过。
- CLI logout 返回成功；独立查询远端会话 valid=false；隔离配置中会话和密码已清空。
- 所有验证使用临时配置；未执行云资源变更、推送或发布。
