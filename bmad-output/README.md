# PowerShell / OEM / PowerCLI Migration Planning

[English](README.md) | [简体中文](README.zh-CN.md) | [Documentation index](../docs/README.md)

The first slice is implemented as the 0.1.2 preview. See the [PowerShell guide](../docs/powershell.md) and [acceptance evidence](../powershell/release/acceptance-checklist.md).

The user selected BMad Method on 2026-11-09. The detailed planning artifacts are in Chinese; this index provides English navigation. The first increment covers connections, VM queries, start/wait, OEM packaging, and a bounded PowerCLI-style compatibility entry point.

| Artifact | Purpose |
| --- | --- |
| [PRD](prd.md) | 10 functional and 6 non-functional requirements, priorities and traceability |
| [Architecture](architecture.md) | 7 ADRs, CLI-first versus SDK, object and process contracts |
| [Epics and stories](epics.md) | 6 epics, 24 story context objects and dependencies |
| [Project context](project-context.md) | Repository evidence and constraints |
| [Decision log](decision-log.md) | Decisions and rationale |
| [Open decisions / roadmap](addendum.md) | OEM identity, customer samples and deferred features |
| [Validation notes](validation-notes.md) | Planning checks and untested boundaries |
| [Windows contract evidence](windows-contract-evidence-2026-11-10.md) | Native preview packages, exact build identities and synthetic migration metrics |

The module bundles an existing CLI binary. Users install one module package; Import-Module only loads it. The existing Go SDK remains behind the CLI, avoiding a second cloud client implementation. The preview includes PowerShell binding, object projection, error semantics, the bounded compatibility layer and OEM packaging.

ready-for-dev means the story contract is complete, not that prerequisites are implemented or a release is approved. The authorized Windows live API checks passed for both authentication methods. Customer scripts and OEM publishing identities remain separate gates.
