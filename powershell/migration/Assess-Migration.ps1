#Requires -Version 7.4
[CmdletBinding()]
param([string] $ContractReport)

$ErrorActionPreference = 'Stop'
$catalog = Get-Content -LiteralPath (Join-Path $PSScriptRoot 'corpus-manifest.json') -Raw | ConvertFrom-Json -AsHashtable
$passed = @()
if ($ContractReport) {
    $report = Get-Content -LiteralPath $ContractReport -Raw | ConvertFrom-Json -AsHashtable
    if ($report.cloud -cne 'loopback mock') { throw 'This assessment only accepts a synthetic contract report.' }
    $passed = $report.passed
}
foreach ($sample in $catalog.cases) {
    $pair = Get-Content -LiteralPath (Join-Path $PSScriptRoot $sample.file) -Raw | ConvertFrom-Json -AsHashtable
    [pscustomobject]@{
        Id = $sample.id; Group = $sample.group; Migration = $sample.migration
        ChangedLines = if ($pair.before -ceq $pair.after) { 0 } else { 1 }
        SyntheticCheck = if ($passed -ccontains $sample.verifiedBy) { 'Passed' } else { 'NotRecorded' }
        Before = $pair.before; After = $pair.after
    }
}
