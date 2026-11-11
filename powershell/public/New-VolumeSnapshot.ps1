function New-CloudVolumeSnapshot {
    [CmdletBinding(SupportsShouldProcess)]
    param(
        [Parameter(Mandatory, ValueFromPipeline)] [ValidateNotNullOrEmpty()] [object[]] $Volume,
        [Parameter(Mandatory)] [ValidateNotNullOrEmpty()] [string] $Name,
        [string] $Description,
        [ValidatePattern('\A[a-fA-F0-9]{32}\z')] [string] $ResourceId,
        [ValidateRange(1, 3600)] [int] $TimeoutSeconds = 300,
        [object] $Server
    )
    begin { $count = 0 }
    process {
        foreach ($item in $Volume) {
            try {
                $count++
                if ($ResourceId -and $count -gt 1) { throw (New-CloudError -Operation NewVolumeSnapshot -ExitCode 2 -Target $ResourceId) }
                $parent = Get-CloudDeviceSelection $item @('Volume', 'HardDisk') $Server NewVolumeSnapshot
                $id = if ($ResourceId) { $ResourceId } else { [guid]::NewGuid().ToString('N') }
                $arguments = @('snapshot', 'create', $parent.Id, "--name=$Name", '--resource-uuid', $id)
                if ($PSBoundParameters.ContainsKey('Description')) { $arguments += "--description=$Description" }
                $target = @{ Connection = $parent.Connection; Id = $id; Name = "$Name on $($parent.Id)" }
                Invoke-CloudDeviceWrite $PSCmdlet $target $arguments create-snapshot NewVolumeSnapshot 'Create volume snapshot' $TimeoutSeconds VolumeSnapshot
            } catch { $PSCmdlet.WriteError($_) }
        }
    }
}
