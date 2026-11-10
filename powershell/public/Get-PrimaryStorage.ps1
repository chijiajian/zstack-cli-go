function Get-CloudPrimaryStorage {
    [CmdletBinding()]
    param(
        [Parameter(Position = 0)] [ValidateNotNullOrEmpty()] [string[]] $Name,
        [ValidateNotNullOrEmpty()] [ValidatePattern('\A[A-Za-z0-9-]{1,128}\z')] [string[]] $Id,
        [object] $Server,
        [ValidateNotNullOrEmpty()] [object[]] $Zone
    )

    if ($PSBoundParameters.ContainsKey('Zone')) {
        Get-CloudInventory -Type PrimaryStorage -Server $Server -Id $Id -Name $Name -ParentField zoneUuid -ParentType Zone -Parent $Zone
    } else {
        Get-CloudInventory -Type PrimaryStorage -Server $Server -Id $Id -Name $Name
    }
}
