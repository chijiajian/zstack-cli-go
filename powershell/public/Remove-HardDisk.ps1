function Remove-CloudHardDisk {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'High')]
    param(
        [Parameter(Mandatory, Position = 0, ValueFromPipeline)] [ValidateNotNullOrEmpty()] [object[]] $HardDisk,
        [object] $VM,
        [ValidateRange(1, 3600)] [int] $TimeoutSeconds = 300,
        [object] $Server
    )
    process {
        foreach ($item in $HardDisk) {
            try {
                $target = Get-CloudResourceSelection $item @('Volume', 'HardDisk') $Server RemoveHardDisk
                $vmId = Get-CloudDeviceParentId $target VMId $VM VM RemoveHardDisk
                $arguments = @('volume', 'detach', $target.Id, '--instance', $vmId)
                Invoke-CloudDeviceWrite $PSCmdlet $target $arguments detach-volume RemoveHardDisk "Detach data volume from VM $vmId" $TimeoutSeconds Volume
            } catch { $PSCmdlet.WriteError($_) }
        }
    }
}
