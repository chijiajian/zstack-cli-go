function Remove-CloudVolumeSnapshot {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'High')]
    param(
        [Parameter(Mandatory, Position = 0, ValueFromPipeline)] [ValidateNotNullOrEmpty()] [object[]] $Snapshot,
        [object] $Volume,
        [ValidateRange(1, 3600)] [int] $TimeoutSeconds = 300,
        [object] $Server
    )
    process {
        foreach ($item in $Snapshot) {
            try {
                $target = Get-CloudResourceSelection $item VolumeSnapshot $Server RemoveVolumeSnapshot
                $volumeId = Get-CloudDeviceParentId $target VolumeId $Volume @('Volume', 'HardDisk') RemoveVolumeSnapshot
                $arguments = @('delete', 'snapshot', $target.Id, '--volume', $volumeId)
                Invoke-CloudDeviceWrite $PSCmdlet $target $arguments delete-snapshot RemoveVolumeSnapshot "Delete leaf volume snapshot for $volumeId" $TimeoutSeconds
            } catch { $PSCmdlet.WriteError($_) }
        }
    }
}
