function Invoke-CloudZql {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory, Position = 0)] [ValidateNotNullOrEmpty()] [string] $Query,
        [ValidateRange(1, 3600)] [int] $TimeoutSeconds = 120,
        [object] $Server
    )

    $connection = Get-CloudConnection $Server
    if (-not (Test-CloudZqlReadStatement $Query)) {
        throw (New-CloudError -Operation Zql -ExitCode 2 -Target $connection.Id)
    }
    $null = Get-CloudCliPath 'zqlV1'
    # The CLI submits one argument to the server's read-only GET /v1/zql parser.
    $result = Invoke-CloudConnectedCli $connection @('zql', '--request-timeout', "${TimeoutSeconds}s", '-o', 'json', '--', $Query) ([Math]::Min(3600, $TimeoutSeconds + 1))
    $reply = Convert-CliResult $result 'Zql' -PreserveDates -Target $connection.Id
    if ($reply['results'] -isnot [array]) {
        throw (New-CloudError -Operation Zql -Failure InvalidResultSets -Target $connection.Id)
    }
    foreach ($set in $reply['results']) {
        if ($set -isnot [System.Collections.IDictionary] -or ($null -ne $set['inventories'] -and $set['inventories'] -isnot [array])) {
            throw (New-CloudError -Operation Zql -Failure InvalidResultSet -Target $connection.Id)
        }
    }
    $index = 0
    foreach ($set in $reply['results']) {
        [pscustomobject] @{
            PSTypeName = 'Cloud.Automation.ZqlResult'
            Index = $index; Name = $set['name']; Records = $set['inventories']; Total = $set['total']
            Counts = $set['inventoryCounts']; Aggregates = $set['inventoryAggregateFunctions']
            ReturnWith = $set['returnWith']; Data = $set
            Server = $connection.Endpoint; ConnectionId = $connection.Id
        }
        $index++
    }
}
