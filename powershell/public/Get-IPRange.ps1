function Get-CloudIPRange {
    [CmdletBinding()]
    param(
        [Parameter(Position = 0)] [ValidateNotNullOrEmpty()] [string[]] $Name,
        [ValidateNotNullOrEmpty()] [ValidatePattern('\A[A-Za-z0-9-]{1,128}\z')] [string[]] $Id,
        [object] $Server,
        [ValidateNotNullOrEmpty()] [object[]] $L3Network
    )

    if ($PSBoundParameters.ContainsKey('L3Network')) {
        Get-CloudInventory -Type IPRange -Server $Server -Id $Id -Name $Name -ParentField l3NetworkUuid -ParentType L3Network -Parent $L3Network
    } else {
        Get-CloudInventory -Type IPRange -Server $Server -Id $Id -Name $Name
    }
}
