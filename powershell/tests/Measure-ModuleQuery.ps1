#Requires -Version 7.4
[CmdletBinding()]
param(
    [Parameter(Mandatory)] [string] $ManifestPath,
    [Parameter(Mandatory)] [string] $Endpoint,
    [Parameter(Mandatory)] [string] $ReportPath,
    [string] $DevelopmentCliPath
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
. (Join-Path $PSScriptRoot 'ContractSupport.ps1')

function Measure-CloudModuleQuery {
    param($Module, $Connection, [string] $Endpoint)

    $arguments = @('get', 'instances', '--all-pages', '--sort', '+uuid', '-o', 'json')
    $cli = {
        & $Module {
            param($Connection, $Arguments)
            $result = Invoke-CloudConnectedCli (Get-CloudConnection $Connection) $Arguments
            if ($result.ExitCode -ne 0) { throw 'Benchmark CLI query failed.' }
        } $Connection $arguments
    }
    & $cli
    $null = Get-ZStackVM -Server $Connection
    $added = @()
    $totals = @()
    $firsts = @()
    $before = (Invoke-RestMethod "http://$Endpoint/test/reads").inventoryReads
    for ($i = 0; $i -lt 20; $i++) {
        $baseline = (Measure-Command { & $cli }).TotalMilliseconds
        $clock = [System.Diagnostics.Stopwatch]::StartNew()
        $count = 0
        $first = 0
        Get-ZStackVM -Server $Connection | ForEach-Object {
            if ($count -eq 0) { $first = $clock.Elapsed.TotalMilliseconds }
            $count++
        }
        $total = $clock.Elapsed.TotalMilliseconds
        if ($count -ne 10000) { throw "Benchmark returned $count records instead of 10000." }
        $firsts += $first
        $totals += $total
        $added += $total - $baseline
    }
    $after = (Invoke-RestMethod "http://$Endpoint/test/reads").inventoryReads
    $requestCount = $after - $before
    if ($requestCount -gt 10000) { throw 'Benchmark detected per-VM API calls.' }
    $p95Index = 18
    return [ordered]@{
        samples = 20; vmCount = 10000; inventoryRequests = $requestCount
        firstItemP95Ms = [Math]::Round(($firsts | Sort-Object)[$p95Index], 2)
        totalP95Ms = [Math]::Round(($totals | Sort-Object)[$p95Index], 2)
        addedP95Ms = [Math]::Round(($added | Sort-Object)[$p95Index], 2)
        latencyTargetMet = ($added | Sort-Object)[$p95Index] -le 500
    }
}

try {
    $module = Import-Module -Name $ManifestPath -PassThru
    Set-CloudContractDevelopmentCli $module $DevelopmentCliPath
    $connection = Connect-ZStackServer -Server $Endpoint -Credential (New-CloudContractCredential)
    try {
        $report = Measure-CloudModuleQuery $module $connection $Endpoint
    } finally {
        Disconnect-ZStackServer -Server $connection
    }
    $report | ConvertTo-Json | Set-Content -LiteralPath $ReportPath -Encoding utf8
    exit 0
} catch {
    @{ failure = $_.Exception.Message } | ConvertTo-Json | Set-Content -LiteralPath $ReportPath -Encoding utf8
    exit 1
}
