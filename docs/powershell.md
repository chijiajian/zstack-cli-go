# PowerShell module preview

[English](powershell.md) | [简体中文](powershell.zh-CN.md) | [Documentation](README.md)

The preview targets Windows x64 and PowerShell 7.4 or later. See the [evidence record](../powershell/release/acceptance-checklist.md) for the exact tested combinations. `ZStack.PowerCLI` is the example brand. The package contains the CLI executable, so module users install only the module package. Go, Python and Pester are development dependencies.

## Install and load

Obtain the complete ZIP and SHA-256 from your provider. On Windows, run the [offline installer](../powershell/packaging/Install-Offline.ps1) against that ZIP. The installer defaults to your user module directory. Alternatively, extract the ZIP into a trusted directory and import the manifest by full path.

```powershell
# Replace the path, version and checksum with the supplied package details.
$manifest = .\Install-Offline.ps1 -Archive .\ZStack.PowerCLI-0.1.2-windows-amd64.zip -SHA256 '<provider checksum>'
Import-Module $manifest
```

After installation into a directory on `$env:PSModulePath`, you can use `Import-Module ZStack.PowerCLI`. Import loads local files. Installation from a configured module repository is a separate step once the provider publishes there. Keep each version in its own directory and open a fresh session when switching versions. [Upgrade and rollback](../powershell/packaging/Upgrade-Policy.md)

Installing an upgrade preserves the CLI used by existing connections. If the upgrade package fails validation, the installed versions remain available. To roll back, start a new session, import the previous version and reconnect.

## Migrate a script

Use a ZStack account and the API endpoint in `host[:port]` form. See [migration examples](../examples/powershell/README.md) and the [parameter compatibility matrix](../powershell/contracts/compatibility.md).

| Task | Native preview | Migration notes |
| --- | --- | --- |
| Connect | `Connect-ZStackServer -Server HOST -Credential (Get-Credential)` | Account login saves only the session. Pass the returned connection explicitly when more than one is active. |
| Inventory | `Get-ZStackVM -Name 'web*' -Server $server` | `Name`, `NumCpu`, `MemoryGB`, `PowerState` and `Id` are available. `Id` is a ZStack UUID. `MemoryGB` uses 2³⁰ bytes per unit. |
| Start | `Get-ZStackVM -Id VM_UUID -Server $server \| Start-ZStackVM -WhatIf` | Preview first. When authorized, remove `-WhatIf` and use `-Confirm` as appropriate. Each VM has a shared 300-second deadline for target resolution, the start request and verification of `Running`. |
| Wait | `$vm \| Wait-ZStackVM -State Running` | Returns a refreshed VM. This does not expose a VMware task object. |
| Disconnect | `Disconnect-ZStackServer -Server $server` | Attempts remote logout and clears the module-owned local session. |

Use `-ErrorAction Stop` when a failure should terminate the pipeline. After an unknown write result, query the VM state before retrying. Live objects carry their connection. Pass an explicit `-Id` and `-Server` when using data read back from a CSV or serialized session.

Starting with 0.1.2, run `Connect-ZStackServer -Server HOST -AccessKeyId KEY_ID -AccessKeySecret (Read-Host 'SK' -AsSecureString)` to use an existing key pair. The module checks the signature with a VM query. Signing depends on an accurate system clock. Use `-ApiTimeZone Asia/Shanghai` if the endpoint needs a specific signing time zone. The module saves the key pair in the connection's private temporary directory and deletes that directory on disconnect.

For the limited legacy command names, import the `Compatibility/ZStack.PowerCLI.Compatibility.psd1` manifest from the same package in a separate PowerShell session. Loaded VMware, Hyper-V or user commands with those names cause import to fail. Unsupported VMware parameters and objects need script changes. Stop, guest shutdown, snapshots, `-RunAsync`, `Get-View` and `.ExtensionData` are outside the preview.

For an OEM package, use the module name and command prefix supplied by the provider. Object properties and script pipeline behavior are shared across brands.
