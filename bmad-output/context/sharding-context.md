# Sharding Context

Read ../project-context.md, ../prd.md, ../architecture.md and ../decision-log.md first. Track BMad Method. Planning only. 24 stories, 6 epics, 4 stories each. Each one-session 2–8h scope. Status backlog until root validation. No implementation/tests/builds. Stories require Story, 3–5 numbered Acceptance Criteria (LOCKED), Tasks mapped AC, Dev Notes with source citations (LOCKED), Testing per AC (strategy only, LOCKED), Dependency Maps with explicit **Blocked by:**, Owned File/Module Scope exact repo paths, Learnings, empty Dev Agent Record.

IDs fixed:
1.1 compatibility-baseline; 1.2 password-stdin; 1.3 start-json-contract; 1.4 vm-memory-bytes.
2.1 module-loader; 2.2 process-adapter; 2.3 error-records; 2.4 connection-registry.
3.1 vm-projection; 3.2 vm-query; 3.3 vm-selection; 3.4 query-pipeline.
4.1 should-process; 4.2 start-vm; 4.3 wait-vm; 4.4 compatibility-module.
5.1 brand-schema; 5.2 bundled-package; 5.3 branded-diagnostics; 5.4 offline-upgrade.
6.1 bilingual-guides; 6.2 windows-contract-ci; 6.3 migration-corpus; 6.4 release-evidence.

Dependency direction: 1 and 2 foundations → 3.1→3.2→3.3→3.4; 4.1 depends 3.3, 4.3 depends 2.4/3.1, 4.2 depends 1.3/4.1/4.3, 4.4 depends 4.2/3.4/1.1. No circular deps. Branding/package may follow module skeleton. Release follows all.

Each story owns precise planned powershell paths plus its own tests. Avoid module bootstrap shared edits: loader discovers explicitly listed generated export manifest; package builder generates it. Scope overlaps allowed only with explicit dependency. CLI story reference docs/usage-en.md, docs/usage-zh.md, docs/reference/commands.md, docs/reference/commands.json, skill checks as needed; serialize CLI stories 1.2→1.3→1.4 for shared docs.

No real customer corpus yet, no final neutral name, no actual publication or credentials. Ready for dev means specified, not all dependencies completed. Release-evidence story may produce blocked/no-go result when external inputs absent; cannot certify release.
