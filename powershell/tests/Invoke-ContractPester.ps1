#Requires -Version 7.4
[CmdletBinding()]
param(
    [Parameter(Mandatory)] [string] $PesterManifest,
    [Parameter(Mandatory)] [string] $WindowsCliPath,
    [string] $DevelopmentCliPath,
    [string] $ReportPath,
    [version] $ModuleVersion,
    [switch] $Benchmark,
    [switch] $ZStackOnly
)

$ErrorActionPreference = 'Stop'
$module = Import-Module -Name $PesterManifest -PassThru
if ($module.Version -ne [version] '5.7.1') { throw 'Contract tests require Pester 5.7.1.' }
$data = @{
    WindowsCliPath = $WindowsCliPath; DevelopmentCliPath = $DevelopmentCliPath; ReportPath = $ReportPath
    ModuleVersion = $ModuleVersion; Benchmark = [bool] $Benchmark; ZStackOnly = [bool] $ZStackOnly
}
$configuration = New-PesterConfiguration
$configuration.Run.Container = New-PesterContainer -Path (Join-Path $PSScriptRoot 'integration/ModuleContracts.Tests.ps1') -Data $data
$configuration.Run.PassThru = $true
$configuration.Output.Verbosity = 'Detailed'
$result = Invoke-Pester -Configuration $configuration
if ($result.Result -ne 'Passed') { throw 'PowerShell contract tests failed.' }
