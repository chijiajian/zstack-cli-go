function Get-CloudVolumeSnapshot {
    [CmdletBinding(DefaultParameterSetName = 'Inventory')]
    param(
        [Parameter(Mandatory, ParameterSetName = 'ByVolume', ValueFromPipeline)] [ValidateNotNullOrEmpty()] [object[]] $Volume,
        [ValidateNotNullOrEmpty()] [string[]] $Name,
        [ValidatePattern('\A[A-Za-z0-9-]{1,128}\z')] [string[]] $Id,
        [object] $Server
    )
    begin { $targets = [System.Collections.Generic.List[object]]::new() }
    process { if ($PSCmdlet.ParameterSetName -ceq 'ByVolume') { $targets.AddRange($Volume) } }
    end {
        if ($PSCmdlet.ParameterSetName -ceq 'Inventory') {
            Get-CloudInventory -Type VolumeSnapshot -Server $Server -Name $Name -Id $Id
        } else {
            Get-CloudInventory -Type VolumeSnapshot -Server $Server -Name $Name -Id $Id -ParentField volumeUuid -ParentType @('Volume', 'HardDisk') -Parent ($targets.ToArray())
        }
    }
}
