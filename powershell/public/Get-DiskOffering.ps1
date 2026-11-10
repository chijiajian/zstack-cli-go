function Get-CloudDiskOffering {
    [CmdletBinding()]
    param(
        [Parameter(Position = 0)] [ValidateNotNullOrEmpty()] [string[]] $Name,
        [ValidateNotNullOrEmpty()] [ValidatePattern('\A[A-Za-z0-9-]{1,128}\z')] [string[]] $Id,
        [object] $Server
    )

    Get-CloudInventory -Type DiskOffering -Server $Server -Id $Id -Name $Name
}
