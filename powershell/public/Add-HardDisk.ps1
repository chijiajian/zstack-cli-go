function Add-CloudHardDisk {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'Medium')]
    param(
        [Parameter(Mandatory, Position = 0, ValueFromPipeline)] [ValidateNotNullOrEmpty()] [object[]] $Volume,
        [Parameter(Mandatory)] [ValidateNotNullOrEmpty()] [object] $VM,
        [ValidateRange(1, 3600)] [int] $TimeoutSeconds = 300,
        [object] $Server
    )
    process {
        foreach ($item in $Volume) {
            try {
                $target = Get-CloudResourceSelection $item @('Volume', 'HardDisk') $Server AddHardDisk
                $vmTarget = Get-CloudResourceSelection $VM VM $target.Connection.Public AddHardDisk
                $arguments = @('volume', 'attach', $target.Id, '--instance', $vmTarget.Id)
                Invoke-CloudDeviceWrite $PSCmdlet $target $arguments attach-volume AddHardDisk "Attach data volume to VM $($vmTarget.Id)" $TimeoutSeconds Volume HardDisk
            } catch { $PSCmdlet.WriteError($_) }
        }
    }
}
