function Restore-CloudVM {
    [CmdletBinding(DefaultParameterSetName = 'VM', SupportsShouldProcess, ConfirmImpact = 'Medium')]
    param(
        [Parameter(Mandatory, Position = 0, ValueFromPipeline, ParameterSetName = 'VM')] [ValidateNotNullOrEmpty()] [object[]] $VM,
        [Parameter(Mandatory, ParameterSetName = 'Id')] [ValidateNotNullOrEmpty()] [ValidatePattern('\A[A-Za-z0-9-]{1,128}\z')] [string[]] $Id,
        [ValidateRange(1, 3600)] [int] $TimeoutSeconds = 300,
        [ValidateRange(1, 60)] [int] $PollSeconds = 2,
        [object] $Server
    )
    process {
        $parameter = $PSCmdlet.ParameterSetName
        Invoke-CloudVMLifecycle -Caller $PSCmdlet -ParameterSet $parameter -Selectors $PSBoundParameters[$parameter] -Server $Server -Command recover -Operation RestoreVM -Description 'Recover deleted VM' -State Stopped -TimeoutSeconds $TimeoutSeconds -PollSeconds $PollSeconds
    }
}
