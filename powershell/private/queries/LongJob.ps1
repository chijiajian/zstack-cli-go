function Read-CloudLongJobResult {
    param($Target, [Diagnostics.Stopwatch] $Clock, [int] $Budget, [string] $Operation,
        [string[]] $AllowedState, [switch] $Mutation)

    try {
        $rows = @(Read-CloudResourceRows $Target.Connection LongJob -Id @($Target.Id) -Clock $Clock -Budget $Budget -Operation $Operation)
        if ($rows.Count -ne 1 -or $rows[0]['state'] -isnot [string] -or $rows[0]['state'] -notin $AllowedState) {
            throw (New-CloudError -Operation $Operation -Failure InvalidJob -Target $Target.Id)
        }
        Convert-CloudResource $rows[0] $Target.Connection LongJob LongJob $null
    } catch {
        if (-not $Mutation) { throw }
        throw (New-CloudError -Operation $Operation -Failure StateUnverified -Mutation -Target $Target.Id)
    }
}

function New-CloudJobTerminalError {
    param([string] $State, [string] $Id)
    $kind, $category = switch -CaseSensitive ($State) {
        'Failed' { 'JobFailed', 'InvalidResult' }
        'Canceled' { 'JobCanceled', 'OperationStopped' }
        default { 'UnexpectedTerminalState', 'ResourceBusy' }
    }
    $exception = [System.InvalidOperationException]::new("The cloud job reached terminal state $State.")
    [System.Management.Automation.ErrorRecord]::new($exception, "CloudAutomation.WaitLongJob.$kind", $category, $Id)
}
