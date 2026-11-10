function Wait-CloudVM {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory, Position = 0, ValueFromPipeline)] [ValidateNotNullOrEmpty()] [object[]] $VM,
        [ValidateSet('Running')] [string] $State = 'Running',
        [ValidateRange(1, 3600)] [int] $TimeoutSeconds = 300,
        [ValidateRange(1, 60)] [int] $PollSeconds = 2,
        [object] $Server
    )

    process {
        foreach ($item in $VM) {
            $clock = [System.Diagnostics.Stopwatch]::StartNew()
            try {
                $target = Resolve-CloudVMTarget -VM $item -Server $Server -TimeoutSeconds $TimeoutSeconds
                Wait-CloudVMState $target $clock $TimeoutSeconds $PollSeconds
            } catch { $PSCmdlet.WriteError($_) }
        }
    }
}
