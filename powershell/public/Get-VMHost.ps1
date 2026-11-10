function Get-CloudVMHost {
    [CmdletBinding()]
    param(
        [Parameter(Position = 0)] [ValidateNotNullOrEmpty()] [string[]] $Name,
        [ValidateNotNullOrEmpty()] [ValidatePattern('\A[A-Za-z0-9-]{1,128}\z')] [string[]] $Id,
        [object] $Server,
        [ValidateNotNullOrEmpty()] [object[]] $Cluster
    )

    if ($PSBoundParameters.ContainsKey('Cluster')) {
        Get-CloudInventory -Type VMHost -Server $Server -Id $Id -Name $Name -ParentField clusterUuid -ParentType Cluster -Parent $Cluster
    } else {
        Get-CloudInventory -Type VMHost -Server $Server -Id $Id -Name $Name
    }
}
