function Wait-CloudVM {
    [CmdletBinding(DefaultParameterSetName = 'VM')]
    param(
        [Parameter(Mandatory, Position = 0, ValueFromPipeline, ParameterSetName = 'VM')] [ValidateNotNullOrEmpty()] [object[]] $VM,
        [Parameter(Mandatory, ParameterSetName = 'Id')] [ValidateNotNullOrEmpty()] [ValidatePattern('\A[A-Za-z0-9-]{1,128}\z')] [string[]] $Id,
        [ValidateSet('Running', 'Stopped', 'Paused', 'Destroyed')] [string] $State = 'Running',
        [ValidateRange(1, 3600)] [int] $TimeoutSeconds = 300,
        [ValidateRange(1, 60)] [int] $PollSeconds = 2,
        [object] $Server
    )

    process {
        $parameter = $PSCmdlet.ParameterSetName
        $stateName = (@('Running', 'Stopped', 'Paused', 'Destroyed') -eq $State)[0]
        foreach ($item in $PSBoundParameters[$parameter]) {
            $clock = [System.Diagnostics.Stopwatch]::StartNew()
            try {
                $selection = @{ Server = $Server; TimeoutSeconds = $TimeoutSeconds }
                $selection[$parameter] = $item
                $target = Resolve-CloudVMTarget @selection
                Wait-CloudVMState $target $clock $TimeoutSeconds $PollSeconds $stateName
            } catch { $PSCmdlet.WriteError($_) }
        }
    }
}
