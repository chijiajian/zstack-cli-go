# PowerShell migration examples

[English](README.md) | [简体中文](README.zh-CN.md) | [PowerShell guide](../../docs/powershell.md)

The following example assumes the preview package is installed and the endpoint and account have been authorized. Substitute the module name and command prefix for an OEM package.

```powershell
Import-Module ZStack.PowerCLI
$server = Connect-ZStackServer -Server '10.255.255.21:8080' -Credential (Get-Credential)
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
