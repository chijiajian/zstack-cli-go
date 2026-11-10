# Synthetic migration cases

The 12 before/after pairs cover four query cases and two each for object pipelines, lifecycle, safety, and OEM coexistence. The snippets are data and depend on the connection and VM selection shown in the [examples](../../examples/powershell/README.md). The evaluator reads them as text. It does not execute customer scripts.

Run `Assess-Migration.ps1 -ContractReport <loopback report>` to associate each case with its behavior check. `command-only` retains the rest of the pipeline. `rewrite` requires data, parameters or workflow changes. `review` requires checking environment-dependent behavior. `Passed` describes the associated synthetic check. A count from these examples is not a customer migration rate.

For customer intake, ask for sanitized scripts and record the source module version, commands, bound parameters, pipeline input and output properties. Flag module-qualified calls, dynamically constructed commands, VMware types, `Get-View`, `.ExtensionData`, and asynchronous tasks for manual review. Keep customer scripts out of the repository unless their owner authorizes sharing.

The report contains `Summary` and `Cases`. `InScopePassRate` divides the recorded passes by the supported samples that do not require manual review. Unsupported VMware behavior is counted separately and may occur in a rewrite case. The line counts compare the snippets in order and report inserted and deleted lines. A line replacement contributes one insertion and one deletion. Command replacements use the PowerShell parser to count the named substitutions declared in each pair. Newly added or removed commands are separate from replacements. The report records the supported behavior, change location and reason for each pair.

With a complete loopback report the corpus currently has 9 of 9 supported, automatically assessed samples passing, 5 command-only samples, 5 rewrites and 2 requiring manual review. One rewrite includes unsupported VMware task behavior. The pairs contain 12 inserted lines, 12 removed lines and 13 command replacements.
