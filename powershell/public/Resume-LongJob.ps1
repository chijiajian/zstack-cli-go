function Resume-CloudLongJob {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'High', DefaultParameterSetName = 'Job')]
    param(
        [Parameter(Mandatory, Position = 0, ValueFromPipeline, ParameterSetName = 'Job')] [ValidateNotNullOrEmpty()] [object[]] $Job,
        [Parameter(Mandatory, ParameterSetName = 'Id')] [Alias('JobId')] [ValidateNotNullOrEmpty()] [ValidatePattern('\A[A-Za-z0-9-]{1,128}\z')] [string[]] $Id,
        [ValidateRange(1, 3600)] [int] $TimeoutSeconds = 120,
        [object] $Server
    )

    begin { $resumed = [System.Collections.Generic.HashSet[string]]::new([StringComparer]::Ordinal) }
    process {
        foreach ($item in $PSBoundParameters[$PSCmdlet.ParameterSetName]) {
            try {
                if ($PSCmdlet.ParameterSetName -eq 'Job' -and $item -is [string]) {
                    throw (New-CloudError -Operation ResumeLongJob -ExitCode 2)
                }
                $target = Get-CloudResourceSelection $item LongJob $Server 'ResumeLongJob'
                $key = "$($target.Connection.Id)/$($target.Id)"
                if ($resumed.Contains($key)) { continue }
                $label = '{0} at {1}' -f $target.Id, $target.Connection.Endpoint
                if (-not $PSCmdlet.ShouldProcess($label, 'Resume suspended cloud job')) { continue }
                $null = Get-CloudCliPath 'longJobV1'
                # Track submitted IDs even if the response is lost. A later invocation will recheck cloud state.
                $null = $resumed.Add($key)
                $clock = [Diagnostics.Stopwatch]::StartNew()
                $result = Invoke-CloudConnectedCli $target.Connection @('job', 'resume', '--uuid-only', '--yes',
                    '--request-timeout', "${TimeoutSeconds}s", '-o', 'json', '--', $target.Id) ([Math]::Min(3600, $TimeoutSeconds + 1))
                $row = Convert-CliResult $result 'ResumeLongJob' -Mutation -Target $target.Id
                if ($row['uuid'] -cne $target.Id) {
                    throw (New-CloudError -Operation ResumeLongJob -Failure InvalidJob -Mutation -Target $target.Id)
                }
                $states = @('Waiting', 'Running', 'Succeeded', 'Failed', 'Canceling', 'Canceled')
                Read-CloudLongJobResult $target $clock $TimeoutSeconds 'ResumeLongJob' -Mutation -AllowedState $states
            } catch { $PSCmdlet.WriteError($_) }
        }
    }
}
