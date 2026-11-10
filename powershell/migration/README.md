# Synthetic migration cases

The 12 before/after pairs cover four query cases and two each for object pipelines, lifecycle, safety, and OEM coexistence. The snippets are data and depend on the connection and VM selection shown in the [examples](../../examples/powershell/README.md). The evaluator reads them as text. It does not execute customer scripts.

Run `Assess-Migration.ps1 -ContractReport <loopback report>` to associate each case with its behavior check. `command-only` retains the rest of the pipeline. `rewrite` requires data, parameters or workflow changes. `review` requires checking environment-dependent behavior. `Passed` describes the associated synthetic check. A count from these examples is not a customer migration rate.

For customer intake, ask for sanitized scripts and record the source module version, commands, bound parameters, pipeline input and output properties. Flag module-qualified calls, dynamically constructed commands, VMware types, `Get-View`, `.ExtensionData`, and asynchronous tasks for manual review. Keep customer scripts out of the repository unless their owner authorizes sharing.
