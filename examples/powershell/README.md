# PowerShell migration examples

[English](README.md) | [简体中文](README.zh-CN.md) | [PowerShell guide](../../docs/powershell.md) | [Inventory and reporting examples](inventory.md)

The following example assumes the preview package is installed and the endpoint and account have been authorized. Substitute the module name and command prefix for an OEM package.

```powershell
Import-Module ZStack.PowerCLI
$server = Connect-ZStackServer -Server '192.0.2.10:8080' -Credential (Get-Credential)
try {
    Get-ZStackVM -Name 'web*' -Server $server |
        Where-Object PowerState -eq 'PoweredOn' |
        Select-Object Name, NumCpu, MemoryGB |
        Export-Csv .\inventory.csv -NoTypeInformation -Encoding utf8

    Get-ZStackVM -Id '<VM_UUID>' -Server $server | Start-ZStackVM -WhatIf -ErrorAction Stop
} finally {
    Disconnect-ZStackServer -Server $server
}
```

When porting a VMware report, the filter and CSV stages remain the same. Replace the connection and VM lookup. Change the VM identifier to the UUID of the authorized ZStack resource before running the preview. [Supported parameter subset](../../powershell/contracts/compatibility.md)

## Before and after

In the VMware column, `$vi` is the connection returned by `Connect-VIServer`. In the ZStack column, `$server` is the connection from the example above. Replace `<VM_UUID>` with the UUID of a VM you are authorized to manage. Start examples use `-WhatIf` so you can review the target. Remove it only when you have authorization to start the VM.

| Scenario | VMware PowerCLI | ZStack PowerCLI preview |
| --- | --- | --- |
| Inventory report | `Get-VM -Server $vi \| Select-Object Name, NumCpu, MemoryGB \| Export-Csv .\inventory.csv -NoTypeInformation` | `Get-ZStackVM -Server $server \| Select-Object Name, NumCpu, MemoryGB \| Export-Csv .\inventory.csv -NoTypeInformation` |
| Explicit connection | `Get-VM -Name 'web*' -Server $vi` | `Get-ZStackVM -Name 'web*' -Server $server` |
| Start from a VM pipeline | `Get-VM -Id $vSphereId -Server $vi \| Start-VM -WhatIf` | `Get-ZStackVM -Id '<VM_UUID>' -Server $server \| Start-ZStackVM -WhatIf` |
| Review targets with WhatIf | `Start-VM -VM $viVm -WhatIf` | `Start-ZStackVM -Id '<VM_UUID>' -Server $server -WhatIf` |
| Stop a script when a command fails | `Get-VM -Name 'web-prod' -Server $vi -ErrorAction Stop` | `Get-ZStackVM -Id '<VM_UUID>' -Server $server -ErrorAction Stop` |
| Opt into legacy names in an OEM session | `Import-Module VCF.PowerCLI` followed by `Get-VM -Name 'web*' -Server $vi` | In a separate session, `Import-Module $oemCompatibilityManifest` followed by `Get-VM -Name 'web*' -Server $oemConnection` |

Set `$oemCompatibilityManifest` to the manifest in the installed OEM package, such as `Compatibility/Acme.Cloud.Compatibility.psd1`. Create `$oemConnection` using the same package's `Connect-AcmeServer`. When using the native OEM commands, call `Get-AcmeVM -Server $oemConnection` and run `Get-Help Get-AcmeVM -Full` for package details. Requery by UUID when a VM object came from CSV or an earlier session.

## VM lifecycle

Use a package built from current source and UUIDs for an authorized test environment. The pipeline uses live ZStack objects and applies `-WhatIf` to each write. The create preview emits no VM, so use the UUID of a separately approved VM for the remaining previews.

```powershell
$image = Get-ZStackImage -Id '<IMAGE_UUID>' -Server $server
$network = Get-ZStackL3Network -Id '<L3_UUID>' -Server $server
$offering = Get-ZStackInstanceOffering -Id '<OFFERING_UUID>' -Server $server
New-ZStackVM -Name 'migration-qa' -Image $image -L3Network $network `
    -InstanceOffering $offering -State Stopped -Server $server -WhatIf

$vm = Get-ZStackVM -Id '<VM_UUID>' -Server $server
$vm | Set-ZStackVM -NumCpu 4 -MemoryGB 8 -WhatIf
$vm | Suspend-ZStackVM -WhatIf
$vm | Resume-ZStackVM -WhatIf
$vm | Restart-ZStackVM -WhatIf
$vm | Stop-ZStackVM -WhatIf
$vm | Remove-ZStackVM -WhatIf
$vm | Restore-ZStackVM -WhatIf
Wait-ZStackVM -Id $vm.Id -State Stopped -TimeoutSeconds 30 -Server $server
```
