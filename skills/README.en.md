# ZStack CLI Skills

[English](README.en.md) | [简体中文](README.md)

This directory contains the distributable [zstack-operations Skill](zstack-operations/SKILL.md). It guides agents using an installed CLI. The repository's AGENTS.md describes code development.

Execution follows Agent + Skill → zstack-cli → ZStack API. The Skill directory installs independently of the CLI source checkout or configuration files.

## Install

Install a trusted CLI binary and make it available on PATH, or use its absolute path. Installation uses the native binary on Windows, macOS and Linux. Python 3 is needed only for the optional capability-check helper.

```sh
zstack-cli skill install --agent codex
zstack-cli skill install --agent claude-code
zstack-cli skill install --agent cursor --scope project --project "/path/to/project"
zstack-cli skill install --agent copilot --dry-run -o json
zstack-cli skill install --agent codex --replace
```

In PowerShell, use `./` or `.\` to run an executable in the current directory. Use the archive's actual executable filename or rename it to zstack-cli.exe. Quote paths with spaces. A project path can be written as `"C:/work/my project"`.

| Agent | User directory, relative to home | Project directory |
| --- | --- | --- |
| Codex | `.agents/skills` | `.agents/skills` |
| Claude Code | `.claude/skills` | `.claude/skills` |
| Cursor | `.cursor/skills` | `.cursor/skills` |
| GitHub Copilot | `.copilot/skills` | `.github/skills` |
| Deep Code | `.agents/skills` | `.deepcode/skills` |

Default scope is user. Project scope requires an existing `--project` directory. `--dest` selects a custom parent directory and the installer creates zstack-operations beneath it. Agent discovery conventions depend on client version, so override the destination to match your actual configuration when needed.

The CLI embeds and copies the entire Skill directory. An existing installation requires explicit `--replace`, which keeps the previous directory in a sibling backup and prints its location. `--dry-run` previews the destination. Manual installation can copy the complete source directory or an extracted standalone Skill archive into the same locations. Windows users can copy it with File Explorer. Preserve local customizations before replacement.

The repository maintains the distributable source under skills/. After installation refresh the client or start a new session. An example request is the following.

> Use zstack-operations to query stopped VMs in the lab context, returning UUID, name and state using session-only credentials or AK/SK.

Optionally check the installed binary's help capabilities. This POSIX example uses the Codex user installation path. On Windows invoke the same script with your Python command and Windows path.

```sh
python3 "$HOME/.agents/skills/zstack-operations/scripts/check_cli.py" \
  --cli /absolute/path/zstack-cli --profile query
```

Other profiles are zql, monitor, lifecycle and iam. The helper executes only version/help, removes ZSTACK variables and selects an empty temporary configuration. Passing establishes help capabilities. Cloud authentication, licenses and resource state need their own checks. Skill content version 1.2.0 corresponds to current source capabilities, with release compatibility explained in references/compatibility.md inside the package.

## WorkBuddy, DeepSeek and other clients

The deepcode/custom targets and WorkBuddy ZIP export are available starting with CLI v1.1.2. Check skill install/export help before using these commands with a released binary.

WorkBuddy imports local Skill packages. Export a ZIP containing the required WorkBuddy display metadata, then choose Add Skill → Upload Skill in WorkBuddy. Complete the client's import checks and enable the Skill.

```sh
zstack-cli skill export --agent workbuddy --file zstack-operations-workbuddy.zip
```

The export includes SKILL.md, references and optional check scripts. Use a new ZIP filename. For cloud operations, WorkBuddy needs terminal access to zstack-cli and network access to the selected ZStack endpoint in the same execution environment. Authenticate through the CLI.

DeepSeek models can run inside different agent clients. Skill discovery follows the client's conventions. Deep Code, listed in DeepSeek's official integration guide, can use these commands.

```sh
zstack-cli skill install --agent deepcode
zstack-cli skill install --agent deepcode --scope project --project "/path/to/project"
```

Deep Code uses `.agents/skills` for user scope and `.deepcode/skills` for project scope. When using a DeepSeek model with Codex or Claude Code, select codex or claude-code accordingly. Other clients supporting local Skills and terminal execution can use an explicit parent directory.

```sh
zstack-cli skill install --agent custom --dest "/path/to/agent/skills"
```

A DeepSeek web chat alone does not establish access to local Skills or CLI execution. The hosting client must provide those capabilities. WorkBuddy exports and Deep Code targets follow official documentation. Validation currently covers ZIP structure, metadata and local installation; actual client import and end-to-end operation remain unverified.

Sources are the [WorkBuddy Skill guide](https://www.workbuddy.ai/docs/workbuddy/From-Beginner-to-Expert-Guide/Function-Description/Skills), [WorkBuddy package structure](https://open.workbuddy.cn/docs/skill) and [DeepSeek Deep Code integration guide](https://github.com/deepseek-ai/awesome-deepseek-agent/blob/main/docs/deepcode.md).

## Package and maintain

```sh
make skill-check
make skill-package
```

Skill-check builds the CLI and runs five capability profiles and helper regression tests locally. Skill-package creates `dist/zstack-operations-skill.tar.gz`, with a top-level zstack-operations directory. Extract it into a temporary location for inspection before manual installation.

Update Skill references and relevant capability checks when changing CLI behavior. CLI and Skill source share a Git tag. Skill metadata.version tracks content independently. Report the exact local and live validation performed.
