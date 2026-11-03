# CLI examples

[English](README.en.md) | [简体中文](README.md)

These workflows use placeholders. Authenticate to the intended context following the [user manual](../docs/usage-en.md), and replace environment names, resource names, UUIDs and paths before running cloud commands. Writes require authorized targets.

| Scenario | English | 简体中文 |
| --- | --- | --- |
| Query, pagination, diagnostics and contexts | [Queries](queries.en.md) | [查询](queries.md) |
| IAM and input files | [IAM](iam.en.md) | [IAM](iam.md) |
| VM, volume, snapshot and job lifecycle | [Lifecycle](lifecycle.en.md) | [生命周期](lifecycle.md) |
| Networking and infrastructure | [Infrastructure](infrastructure.en.md) | [基础设施](infrastructure.md) |
| Regression and cleanup recovery | [Regression](regression.en.md) | [回归](regression.md) |

The [policy JSON](policy-statements.json) is shared by both languages. The [command reference](../docs/reference/commands.md) contains usage, aliases and flags for the entire command tree. Workflow coverage and actual backend validation are described separately in the manual and historical checklists.

Maintain generated references with `make docs` and check them with `make docs-check`. These targets execute CLI help locally.
