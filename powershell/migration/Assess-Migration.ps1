#Requires -Version 7.4
[CmdletBinding()]
param([string] $ContractReport)

$ErrorActionPreference = 'Stop'

function Measure-ChangedLines {
    param([string] $Before, [string] $After)
    $old = $Before -split '\r\n|\n|\r'
    $new = $After -split '\r\n|\n|\r'
    # Match lines in order so an insertion does not count all later lines as changed.
    $common = [int[,]]::new($old.Count + 1, $new.Count + 1)
    for ($i = 1; $i -le $old.Count; $i++) {
        for ($j = 1; $j -le $new.Count; $j++) {
            $common[$i, $j] = if ($old[$i - 1] -ceq $new[$j - 1]) {
                1 + $common[($i - 1), ($j - 1)]
            } else {
                [Math]::Max($common[($i - 1), $j], $common[$i, ($j - 1)])
            }
        }
    }
    $same = $common[$old.Count, $new.Count]
    return @{ Added = $new.Count - $same; Removed = $old.Count - $same }
}

function Get-StaticCommandCounts {
    param([string] $Text)
    $parseErrors = $null
    $tree = [System.Management.Automation.Language.Parser]::ParseInput($Text, [ref] $null, [ref] $parseErrors)
    if ($parseErrors.Count) { throw 'A synthetic snippet does not parse.' }
    $counts = @{}
    foreach ($command in $tree.FindAll({ param($node) $node -is [System.Management.Automation.Language.CommandAst] }, $true)) {
        $name = $command.GetCommandName()
        if ($name) { $counts[$name] = 1 + $counts[$name] }
    }
    return $counts
}

$catalog = Get-Content -LiteralPath (Join-Path $PSScriptRoot 'corpus-manifest.json') -Raw | ConvertFrom-Json -AsHashtable
$passed = @()
if ($ContractReport) {
    $report = Get-Content -LiteralPath $ContractReport -Raw | ConvertFrom-Json -AsHashtable
    if ($report.cloud -cne 'loopback mock') { throw 'This assessment only accepts a synthetic contract report.' }
    $passed = $report.passed
}
$rows = @(foreach ($sample in $catalog.cases) {
    $pair = Get-Content -LiteralPath (Join-Path $PSScriptRoot $sample.file) -Raw | ConvertFrom-Json -AsHashtable
    $lines = Measure-ChangedLines $pair.before $pair.after
    $oldCommands = Get-StaticCommandCounts $pair.before
    $newCommands = Get-StaticCommandCounts $pair.after
    $replacementCount = 0
    foreach ($replacement in $pair.replacements) {
        $removed = $oldCommands[$replacement.from] - $newCommands[$replacement.from]
        $added = $newCommands[$replacement.to] - $oldCommands[$replacement.to]
        if ($removed -lt 1 -or $added -lt 1) { throw "Invalid command replacement in sample $($sample.id)." }
        $replacementCount += [Math]::Min($removed, $added)
    }
    $missing = @($sample.verifiedBy | Where-Object { $passed -cnotcontains $_ })
    [pscustomobject]@{
        Id = $sample.id; Group = $sample.group; Migration = $sample.migration
        SupportedBehavior = $pair.supportedBehavior; Unsupported = @($pair.unsupported)
        Changes = @($pair.changes); AddedLines = $lines.Added; RemovedLines = $lines.Removed
        ChangedLines = $lines.Added + $lines.Removed; CommandReplacements = $replacementCount
        SyntheticCheck = if ($missing.Count -eq 0) { 'Passed' } else { 'NotRecorded' }
        UnrecordedChecks = $missing
        Before = $pair.before; After = $pair.after
    }
})
$inScope = @($rows | Where-Object { $_.Migration -ne 'review' -and $_.Unsupported.Count -eq 0 })
$inScopePassed = @($inScope | Where-Object SyntheticCheck -eq 'Passed').Count
[pscustomobject]@{
    Kind = 'synthetic'
    Summary = [ordered]@{
        Samples = $rows.Count
        CommandOnly = @($rows | Where-Object Migration -eq 'command-only').Count
        Rewrite = @($rows | Where-Object Migration -eq 'rewrite').Count
        ManualReview = @($rows | Where-Object Migration -eq 'review').Count
        Unsupported = @($rows | Where-Object { $_.Unsupported.Count -gt 0 }).Count
        InScope = $inScope.Count
        InScopePassed = $inScopePassed
        InScopePassRate = if ($ContractReport -and $inScope.Count) { $inScopePassed / $inScope.Count }
        AddedLines = ($rows | Measure-Object AddedLines -Sum).Sum
        RemovedLines = ($rows | Measure-Object RemovedLines -Sum).Sum
        ChangedLines = ($rows | Measure-Object ChangedLines -Sum).Sum
        CommandReplacements = ($rows | Measure-Object CommandReplacements -Sum).Sum
    }
    Cases = $rows
}
