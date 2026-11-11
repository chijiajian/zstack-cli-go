function Get-CloudReportInteger {
    param($Value, $Connection)
    if (($Value -isnot [int] -and $Value -isnot [long]) -or $Value -lt 0) {
        throw (New-CloudError -Operation Usage -Failure InvalidCounter -Target $Connection.Id)
    }
    return [long] $Value
}

function Get-CloudReportPercent {
    param($Row, [string] $Key, $Connection)
    $value = $Row[$Key]
    if (-not $Row.Contains($Key) -or ($null -ne $value -and
        ($value -isnot [int] -and $value -isnot [long] -and $value -isnot [double]))) {
        throw (New-CloudError -Operation Usage -Failure InvalidPercent -Target $Connection.Id)
    }
    if ($null -ne $value) { return [double] $value }
}

function Convert-CloudCapacityRow {
    param($Row, $Connection)
    $units = @{ cpu = 'vCPU'; memory = 'bytes'; primaryStorageLogical = 'bytes'; primaryStoragePhysical = 'bytes' }
    if ($Row -isnot [System.Collections.IDictionary] -or $Row['metric'] -isnot [string] -or
        -not $units.ContainsKey($Row['metric']) -or $Row['unit'] -cne $units[$Row['metric']]) {
        throw (New-CloudError -Operation Usage -Failure InvalidCapacity -Target $Connection.Id)
    }
    [pscustomobject] @{
        PSTypeName = 'Cloud.Automation.UsageCapacity'
        Metric = $Row['metric']; Unit = $Row['unit']
        Total = Get-CloudReportInteger $Row['total'] $Connection
        Used = Get-CloudReportInteger $Row['used'] $Connection
        Available = Get-CloudReportInteger $Row['available'] $Connection
        Percent = Get-CloudReportPercent $Row 'percent' $Connection
        Server = $Connection.Endpoint; ConnectionId = $Connection.Id
    }
}

function Convert-CloudUsageLoad {
    param($Row, $Connection)
    if ($Row -isnot [System.Collections.IDictionary] -or -not $Row.Contains('sampleTime')) {
        throw (New-CloudError -Operation Usage -Failure InvalidLoad -Target $Connection.Id)
    }
    $sampleTime = [datetimeoffset]::MinValue
    if ($null -ne $Row['sampleTime'] -and
        ($Row['sampleTime'] -isnot [string] -or -not [datetimeoffset]::TryParse($Row['sampleTime'],
            [Globalization.CultureInfo]::InvariantCulture, [Globalization.DateTimeStyles]::None, [ref] $sampleTime))) {
        throw (New-CloudError -Operation Usage -Failure InvalidSampleTime -Target $Connection.Id)
    }
    [pscustomobject] @{
        PSTypeName = 'Cloud.Automation.UsageLoad'
        SampleTime = if ($null -ne $Row['sampleTime']) { $sampleTime }
        ConnectedHosts = Get-CloudReportInteger $Row['connectedHosts'] $Connection
        SampledHosts = Get-CloudReportInteger $Row['sampledHosts'] $Connection
        CpuPercent = Get-CloudReportPercent $Row 'cpuPercent' $Connection
        MemoryPercent = Get-CloudReportPercent $Row 'memoryPercent' $Connection
        Server = $Connection.Endpoint; ConnectionId = $Connection.Id
    }
}

function Test-CloudZqlReadStatement {
    param([string] $Query)
    $prefix = '\A\s*(?:query|count)\s+\S'
    $quote = $null
    $start = 0
    for ($i = 0; $i -lt $Query.Length; $i++) {
        $character = $Query[$i]
        if ($null -ne $quote) {
            if ($character -ceq '\') { $i++; continue }
            if ($character -ceq $quote) { $quote = $null }
        } elseif ($character -in "'", '"') {
            $quote = $character
        } elseif ($character -ceq ';') {
            if ($Query.Substring($start, $i - $start) -notmatch $prefix) { return $false }
            $start = $i + 1
        }
    }
    if ($null -ne $quote) { return $false }
    $tail = $Query.Substring($start)
    return ($tail -match $prefix -or ($start -gt 0 -and [string]::IsNullOrWhiteSpace($tail)))
}
