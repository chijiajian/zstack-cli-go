function Wait-CloudLongJob {
    [CmdletBinding(DefaultParameterSetName = 'Job')]
    param(
        [Parameter(Mandatory, Position = 0, ValueFromPipeline, ParameterSetName = 'Job')] [ValidateNotNullOrEmpty()] [object[]] $Job,
        [Parameter(Mandatory, ParameterSetName = 'Id')] [Alias('JobId')] [ValidateNotNullOrEmpty()] [ValidatePattern('\A[A-Za-z0-9-]{1,128}\z')] [string[]] $Id,
        [ValidateSet('Succeeded', 'Failed', 'Canceled')] [string] $State = 'Succeeded',
        [ValidateRange(1, 3600)] [int] $TimeoutSeconds = 600,
        [ValidateRange(1, 60)] [int] $PollSeconds = 5,
        [object] $Server
    )

    process {
        foreach ($item in $PSBoundParameters[$PSCmdlet.ParameterSetName]) {
            try {
                if ($PSCmdlet.ParameterSetName -eq 'Job' -and $item -is [string]) {
                    throw (New-CloudError -Operation WaitLongJob -ExitCode 2)
                }
                $target = Get-CloudResourceSelection $item LongJob $Server 'WaitLongJob'
                $null = Get-CloudCliPath 'longJobV1'
                $clock = [Diagnostics.Stopwatch]::StartNew()
                $result = Invoke-CloudConnectedCli $target.Connection @('wait', 'long-job', '--uuid-only',
                    '--for', $State, '--timeout', "${TimeoutSeconds}s", '--interval', "${PollSeconds}s", '-o', 'json', '--', $target.Id) ([Math]::Min(3600, $TimeoutSeconds + 1))
                if ($result.ExitCode -eq 5 -and -not $result.TimedOut -and -not $result.Canceled -and -not $result.Failure -and $result.Stdout.Trim()) {
                    try { $terminal = ConvertFrom-Json -InputObject $result.Stdout -AsHashtable -NoEnumerate -ErrorAction Stop }
                    catch { throw (New-CloudError -Operation WaitLongJob -Failure InvalidJob -Target $target.Id) }
                    if ($terminal -is [System.Collections.IDictionary] -and $terminal['resourceType'] -ceq 'long-job' -and
                        $terminal['uuid'] -ceq $target.Id -and $terminal['state'] -cin 'Succeeded', 'Failed', 'Canceled') {
                        throw (New-CloudJobTerminalError $terminal['state'] $target.Id)
                    }
                    throw (New-CloudError -Operation WaitLongJob -Failure InvalidJob -Target $target.Id)
                }
                $row = Convert-CliResult $result 'WaitLongJob' -Target $target.Id
                if ($row['resourceType'] -cne 'long-job' -or $row['uuid'] -cne $target.Id -or $row['state'] -ne $State) {
                    throw (New-CloudError -Operation WaitLongJob -Failure InvalidJob -Target $target.Id)
                }
                Read-CloudLongJobResult $target $clock $TimeoutSeconds 'WaitLongJob' -AllowedState $State
            } catch { $PSCmdlet.WriteError($_) }
        }
    }
}
