function Get-CloudInstanceOffering {
    [CmdletBinding()]
    param(
        [Parameter(Position = 0)] [ValidateNotNullOrEmpty()] [string[]] $Name,
        [ValidateNotNullOrEmpty()] [ValidatePattern('\A[A-Za-z0-9-]{1,128}\z')] [string[]] $Id,
        [object] $Server
    )

    Get-CloudInventory -Type InstanceOffering -Server $Server -Id $Id -Name $Name
}
