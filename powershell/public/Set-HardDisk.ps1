function Set-CloudHardDisk {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'Medium')]
    param(
        [Parameter(Mandatory, Position = 0, ValueFromPipeline)] [ValidateNotNullOrEmpty()] [object[]] $HardDisk,
        [Parameter(Mandatory)] [decimal] $CapacityGB,
        [ValidateRange(1, 3600)] [int] $TimeoutSeconds = 300,
        [object] $Server
    )
    process {
        foreach ($item in $HardDisk) {
            try {
                $target = Get-CloudResourceSelection $item @('Volume', 'HardDisk') $Server SetHardDisk
                $bytes = ConvertTo-CloudMemoryBytes $CapacityGB SetHardDisk
                $arguments = @('volume', 'resize', $target.Id, '--size', [string] $bytes)
                Invoke-CloudDeviceWrite $PSCmdlet $target $arguments resize-volume SetHardDisk 'Expand virtual disk to total capacity' $TimeoutSeconds Volume HardDisk
            } catch { $PSCmdlet.WriteError($_) }
        }
    }
}
