function Set-CloudVM {
    [CmdletBinding(DefaultParameterSetName = 'VM', SupportsShouldProcess, ConfirmImpact = 'Medium')]
    param(
        [Parameter(Mandatory, Position = 0, ValueFromPipeline, ParameterSetName = 'VM')] [ValidateNotNullOrEmpty()] [object[]] $VM,
        [Parameter(Mandatory, ParameterSetName = 'Id')] [ValidateNotNullOrEmpty()] [ValidatePattern('\A[A-Za-z0-9-]{1,128}\z')] [string[]] $Id,
        [ValidateNotNullOrEmpty()] [string] $Name,
        [AllowEmptyString()] [string] $Description,
        [ValidateRange(1, 2147483647)] [int] $NumCpu,
        [decimal] $MemoryGB,
        [ValidateRange(1, 3600)] [int] $TimeoutSeconds = 300,
        [object] $Server
    )
    begin {
        $options = @()
        $changes = @{}
        if ($PSBoundParameters.ContainsKey('Name')) {
            if ([string]::IsNullOrWhiteSpace($Name)) { throw (New-CloudError -Operation SetVM -ExitCode 2) }
            $options += "--name=$Name"
            $changes.name = $Name
        }
        if ($PSBoundParameters.ContainsKey('Description')) {
            $options += "--description=$Description"
            $changes.description = $Description
        }
        if ($PSBoundParameters.ContainsKey('NumCpu')) {
            $options += @('--cpu', [string] $NumCpu)
            $changes.cpuNum = $NumCpu
        }
        if ($PSBoundParameters.ContainsKey('MemoryGB')) {
            $bytes = ConvertTo-CloudMemoryBytes $MemoryGB SetVM
            $options += @('--memory', [string] $bytes)
            $changes.memorySizeBytes = $bytes
        }
        if (-not $options.Count) { throw (New-CloudError -Operation SetVM -ExitCode 2) }
    }
    process {
        $parameter = $PSCmdlet.ParameterSetName
        Invoke-CloudVMLifecycle -Caller $PSCmdlet -ParameterSet $parameter -Selectors $PSBoundParameters[$parameter] -Server $Server -Command update -Operation SetVM -Description 'Update VM' -TimeoutSeconds $TimeoutSeconds -Options $options -Changes $changes
    }
}
