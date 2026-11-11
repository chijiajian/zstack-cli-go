function Set-CloudCDDrive {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'Medium')]
    param(
        [Parameter(Mandatory, Position = 0, ValueFromPipeline)] [ValidateNotNullOrEmpty()] [object[]] $CD,
        [Parameter(Mandatory, ParameterSetName = 'Insert')] [ValidateNotNullOrEmpty()] [object] $IsoImage,
        [Parameter(Mandatory, ParameterSetName = 'Eject')] [switch] $NoMedia,
        [object] $VM,
        [ValidateRange(1, 3600)] [int] $TimeoutSeconds = 300,
        [object] $Server
    )
    process {
        foreach ($item in $CD) {
            try {
                $target = Get-CloudResourceSelection $item CDDrive $Server SetCDDrive
                $vmId = Get-CloudDeviceParentId $target VMId $VM VM SetCDDrive
                if ($PSCmdlet.ParameterSetName -ceq 'Eject') {
                    if (-not $NoMedia) { throw (New-CloudError -Operation SetCDDrive -ExitCode 2 -Target $target.Id) }
                    $arguments = @('instance', 'detach-iso', $vmId, '--cdrom', $target.Id)
                    Invoke-CloudDeviceWrite $PSCmdlet $target $arguments detach-iso SetCDDrive "Eject ISO from VM $vmId" $TimeoutSeconds CDDrive
                } else {
                    $image = Get-CloudResourceSelection $IsoImage Image $target.Connection.Public SetCDDrive
                    $arguments = @('instance', 'attach-iso', $vmId, '--cdrom', $target.Id, '--iso', $image.Id)
                    Invoke-CloudDeviceWrite $PSCmdlet $target $arguments attach-iso SetCDDrive "Insert ISO $($image.Id) on VM $vmId" $TimeoutSeconds CDDrive
                }
            } catch { $PSCmdlet.WriteError($_) }
        }
    }
}
