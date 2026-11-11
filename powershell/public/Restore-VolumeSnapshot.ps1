function Restore-CloudVolumeSnapshot {
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
                $target = Get-CloudResourceSelection $item VolumeSnapshot $Server RestoreVolumeSnapshot
                $volumeId = Get-CloudDeviceParentId $target VolumeId $Volume @('Volume', 'HardDisk') RestoreVolumeSnapshot
                $arguments = @('snapshot', 'revert', $target.Id, '--volume', $volumeId)
                Invoke-CloudDeviceWrite $PSCmdlet $target $arguments revert-snapshot RestoreVolumeSnapshot "Revert volume from snapshot for $volumeId" $TimeoutSeconds VolumeSnapshot
            } catch { $PSCmdlet.WriteError($_) }
        }
    }
}
