function Remove-CloudVolume {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'High')]
    param(
        [Parameter(Mandatory, Position = 0, ValueFromPipeline)] [ValidateNotNullOrEmpty()] [object[]] $Volume,
        [ValidateRange(1, 3600)] [int] $TimeoutSeconds = 300,
        [object] $Server
    )
    process {
        foreach ($item in $Volume) {
            try {
                $target = Get-CloudDeviceSelection $item Volume $Server RemoveVolume
                $arguments = @('delete', 'volume', $target.Id)
                Invoke-CloudDeviceWrite $PSCmdlet $target $arguments delete-volume RemoveVolume 'Delete detached data volume to recycle bin' $TimeoutSeconds
            } catch { $PSCmdlet.WriteError($_) }
        }
    }
}
