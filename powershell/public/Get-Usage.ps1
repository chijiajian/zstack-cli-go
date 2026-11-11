function Get-CloudUsage {
    [CmdletBinding()]
    param(
        [ValidateSet('All', 'Capacity', 'Load')] [string] $Section = 'All',
        [ValidateRange(1, 3600)] [int] $TimeoutSeconds = 120,
        [object] $Server
    )

    $connection = Get-CloudConnection $Server
    $null = Get-CloudCliPath 'usageV1'
    $result = Invoke-CloudConnectedCli $connection @('monitor', 'usage', '--section', $Section.ToLowerInvariant(), '-o', 'json') $TimeoutSeconds
    $report = Convert-CliResult $result 'Usage' -PreserveDates -Target $connection.Id
    $capacity = $null
    $load = $null
    if ($Section -ne 'Load') {
        if ($report['capacity'] -isnot [array]) {
            throw (New-CloudError -Operation Usage -Failure InvalidCapacity -Target $connection.Id)
        }
        $capacity = @(foreach ($row in $report['capacity']) { Convert-CloudCapacityRow $row $connection })
    }
    if ($Section -ne 'Capacity') { $load = Convert-CloudUsageLoad $report['load'] $connection }
    [pscustomobject] @{
        PSTypeName = 'Cloud.Automation.Usage'
        Server = $connection.Endpoint; ConnectionId = $connection.Id
        Capacity = $capacity; Load = $load
    }
}
