function New-CloudError {
    param(
        [string] $Operation,
        [int] $ExitCode = 1,
        [string] $Failure,
        [switch] $Mutation,
        [object] $Target
    )

    $details = switch ($ExitCode) {
        2 { @('InvalidArguments', 'InvalidArgument', 'The command arguments were rejected.') }
        3 { @('AuthenticationFailed', 'AuthenticationError', 'Authentication or permission check failed.') }
        4 { @('NotFound', 'ObjectNotFound', 'The requested resource was not found.') }
        5 { @('Conflict', 'ResourceBusy', 'The resource state or identity conflicts with the operation.') }
        6 { @('Timeout', 'OperationTimeout', 'The operation timed out.') }
        7 { @('ConnectionFailed', 'ConnectionError', 'The configured endpoint could not be reached.') }
        8 { @('ApiFailed', 'InvalidResult', 'The API rejected the operation or returned an invalid response.') }
        9 { @('OutputFailed', 'WriteError', 'The CLI could not write the result.') }
        130 { @('Canceled', 'OperationStopped', 'The operation was canceled.') }
        default { @('OperationFailed', 'NotSpecified', 'The operation failed.') }
    }
    if ($Failure -eq 'ProcessStart') { $details = @('ProcessStart', 'OpenError', 'The packaged CLI could not be started.') }
    elseif ($Failure) { $details = @('Protocol', 'InvalidData', 'The packaged CLI returned an invalid or oversized result.') }
    if ($Mutation -and ($ExitCode -in 6, 7, 8, 9, 130 -or ($Failure -and $Failure -ne 'ProcessStart'))) {
        $details = @('OutcomeUnknown', $details[1], 'The write outcome is unknown. Check the target state before retrying.')
    }
    $message = if (Get-Command Resolve-CloudMessage -CommandType Function -ErrorAction Ignore) {
        Resolve-CloudMessage -Code $details[0] -Fallback $details[2]
    } else { $details[2] }
    $exception = [System.InvalidOperationException]::new($message)
    return [System.Management.Automation.ErrorRecord]::new($exception, "CloudAutomation.$Operation.$($details[0])", $details[1], $Target)
}

function Get-CloudProcessError {
    param($Result, [string] $Operation, [switch] $Mutation, [object] $Target)

    $exit = if ($Result.TimedOut) { 6 } elseif ($Result.Canceled) { 130 } else { [int] $Result.ExitCode }
    if ($exit -ne 0 -or $Result.Failure) {
        New-CloudError -Operation $Operation -ExitCode $exit -Failure $Result.Failure -Mutation:$Mutation -Target $Target
    }
}

function Convert-CliResult {
    param(
        [Parameter(Mandatory)] $Result,
        [Parameter(Mandatory)] [string] $Operation,
        [ValidateSet('Array', 'Object')] [string] $Shape = 'Object',
        [switch] $Mutation,
        [object] $Target
    )

    $failure = Get-CloudProcessError $Result $Operation -Mutation:$Mutation -Target $Target
    if ($failure) { throw $failure }
    try {
        $parsed = ConvertFrom-Json -InputObject $Result.Stdout -AsHashtable -NoEnumerate -ErrorAction Stop
    } catch {
        throw (New-CloudError -Operation $Operation -Failure 'InvalidJson' -Mutation:$Mutation -Target $Target)
    }
    if (($Shape -eq 'Array' -and $parsed -isnot [array]) -or ($Shape -eq 'Object' -and $parsed -isnot [System.Collections.IDictionary])) {
        throw (New-CloudError -Operation $Operation -Failure 'InvalidShape' -Mutation:$Mutation -Target $Target)
    }
    return ,$parsed
}
