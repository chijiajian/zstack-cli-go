function Restore-CloudVolume {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'Medium')]
    param(
        [Parameter(Mandatory, Position = 0, ValueFromPipeline)] [ValidateNotNullOrEmpty()] [object[]] $Volume,
        [ValidateRange(1, 3600)] [int] $TimeoutSeconds = 300,
        [object] $Server
    )
    process {
        foreach ($item in $Volume) {
            try {
                $target = Get-CloudResourceSelection $item Volume $Server RestoreVolume
                $arguments = @('volume', 'recover', $target.Id)
                Invoke-CloudDeviceWrite $PSCmdlet $target $arguments recover-volume RestoreVolume 'Recover data volume from recycle bin' $TimeoutSeconds Volume
            } catch { $PSCmdlet.WriteError($_) }
        }
    }
}
