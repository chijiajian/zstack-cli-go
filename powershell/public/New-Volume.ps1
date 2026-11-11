function New-CloudVolume {
    [CmdletBinding(DefaultParameterSetName = 'Capacity', SupportsShouldProcess)]
    param(
        [Parameter(Mandatory, Position = 0)] [ValidateNotNullOrEmpty()] [string] $Name,
        [Parameter(Mandatory, ParameterSetName = 'Capacity')] [decimal] $CapacityGB,
        [Parameter(Mandatory, ParameterSetName = 'Offering')] [ValidateNotNullOrEmpty()] [object] $DiskOffering,
        [object] $PrimaryStorage,
        [string] $Description,
        [ValidatePattern('\A[a-fA-F0-9]{32}\z')] [string] $ResourceId = [guid]::NewGuid().ToString('N'),
        [ValidateRange(1, 3600)] [int] $TimeoutSeconds = 300,
        [object] $Server
    )
    $connection = Get-CloudConnection $Server
    $arguments = @('create', 'volume', '--resource-uuid', $ResourceId)
    if ($PSCmdlet.ParameterSetName -ceq 'Capacity') {
        $arguments += @('--size', [string] (ConvertTo-CloudMemoryBytes $CapacityGB NewVolume))
    } else {
        $offering = Get-CloudDeviceSelection $DiskOffering DiskOffering $connection.Public NewVolume
        $arguments += @('--disk-offering', $offering.Id)
    }
    if ($null -ne $PrimaryStorage) {
        $storage = Get-CloudDeviceSelection $PrimaryStorage PrimaryStorage $connection.Public NewVolume
        $arguments += @('--primary-storage', $storage.Id)
    }
    if ($PSBoundParameters.ContainsKey('Description')) { $arguments += "--description=$Description" }
    $target = @{ Connection = $connection; Id = $ResourceId; Name = $Name }
    Invoke-CloudDeviceWrite $PSCmdlet $target ($arguments + @('--', $Name)) create-volume NewVolume 'Create independent data volume' $TimeoutSeconds Volume
}
