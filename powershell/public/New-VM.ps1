function New-CloudVM {
    [CmdletBinding(DefaultParameterSetName = 'Offering', SupportsShouldProcess, ConfirmImpact = 'Medium')]
    param(
        [Parameter(Mandatory, Position = 0)] [ValidateNotNullOrEmpty()] [string] $Name,
        [Parameter(Mandatory)] [ValidateNotNullOrEmpty()] [object] $Image,
        [Parameter(Mandatory)] [ValidateNotNullOrEmpty()] [object[]] $L3Network,
        [Parameter(Mandatory, ParameterSetName = 'Offering')] [ValidateNotNullOrEmpty()] [object] $InstanceOffering,
        [Parameter(Mandatory, ParameterSetName = 'Sizing')] [ValidateRange(1, 2147483647)] [int] $NumCpu,
        [Parameter(Mandatory, ParameterSetName = 'Sizing')] [decimal] $MemoryGB,
        [Parameter(Mandatory)] [ValidateSet('Stopped', 'Running')] [string] $State,
        [AllowEmptyString()] [string] $Description,
        [ValidatePattern('\A[a-fA-F0-9]{32}\z')] [string] $ResourceId = [guid]::NewGuid().ToString('N'),
        [ValidateRange(1, 3600)] [int] $TimeoutSeconds = 300,
        [ValidateRange(1, 60)] [int] $PollSeconds = 2,
        [object] $Server
    )
    $null = Get-CloudCliPath 'vmLifecycleV1'
    $clock = [System.Diagnostics.Stopwatch]::StartNew()
    if ([string]::IsNullOrWhiteSpace($Name)) { throw (New-CloudError -Operation NewVM -ExitCode 2 -Target $ResourceId) }
    $connection = Get-CloudConnection $Server
    $imageId = Resolve-CloudDeploymentId $Image Image $connection
    $networks = @(foreach ($network in $L3Network) { Resolve-CloudDeploymentId $network L3Network $connection })
    $stateName = if ($State -eq 'Running') { 'Running' } else { 'Stopped' }
    $strategy = if ($stateName -ceq 'Running') { 'InstantStart' } else { 'CreateStopped' }
    $arguments = @('create', 'instance', '--image', $imageId, '--l3-network', ($networks -join ','),
        '--resource-uuid', $ResourceId, '--strategy', $strategy)
    if ($PSCmdlet.ParameterSetName -ceq 'Offering') {
        $arguments += @('--instance-offering', (Resolve-CloudDeploymentId $InstanceOffering InstanceOffering $connection))
    } else {
        $arguments += @('--cpu', [string] $NumCpu, '--memory', [string] (ConvertTo-CloudMemoryBytes $MemoryGB NewVM))
    }
    if ($PSBoundParameters.ContainsKey('Description')) { $arguments += "--description=$Description" }
    $target = @{ Connection = $connection; VM = @{ Id = $ResourceId; Name = $Name } }
    if (-not (Invoke-CloudShouldProcess $PSCmdlet $target 'Create VM')) { return }

    Invoke-CloudVMOperation $connection ($arguments + @('--', $Name)) create NewVM $ResourceId $clock $TimeoutSeconds
    try { Wait-CloudVMState $target $clock $TimeoutSeconds $PollSeconds $stateName }
    catch { throw (New-CloudError -Operation NewVM -Failure StateUnverified -Mutation -Target $ResourceId) }
}
