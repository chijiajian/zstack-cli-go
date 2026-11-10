function Get-CloudNetworkAdapter {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory, Position = 0, ValueFromPipeline)] [ValidateNotNullOrEmpty()] [object[]] $VM,
        [ValidateNotNullOrEmpty()] [ValidatePattern('\A[A-Za-z0-9-]{1,128}\z')] [string[]] $Id,
        [object] $Server,
        [switch] $IncludeNetworkName
    )

    begin { $targets = [System.Collections.Generic.List[object]]::new() }
    process { $targets.AddRange($VM) }
    end {
        Get-CloudInventory -Type NetworkAdapter -Server $Server -Id $Id -ParentField vmInstanceUuid -ParentType VM -Parent ($targets.ToArray()) -IncludeNetworkName:$IncludeNetworkName
    }
}
