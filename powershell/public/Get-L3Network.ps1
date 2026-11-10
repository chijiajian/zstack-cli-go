function Get-CloudL3Network {
    [CmdletBinding()]
    param(
        [Parameter(Position = 0)] [ValidateNotNullOrEmpty()] [string[]] $Name,
        [ValidateNotNullOrEmpty()] [ValidatePattern('\A[A-Za-z0-9-]{1,128}\z')] [string[]] $Id,
        [object] $Server,
        [ValidateNotNullOrEmpty()] [object[]] $L2Network
    )

    if ($PSBoundParameters.ContainsKey('L2Network')) {
        Get-CloudInventory -Type L3Network -Server $Server -Id $Id -Name $Name -ParentField l2NetworkUuid -ParentType L2Network -Parent $L2Network
    } else {
        Get-CloudInventory -Type L3Network -Server $Server -Id $Id -Name $Name
    }
}
