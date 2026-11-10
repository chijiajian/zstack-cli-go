function Get-CloudImage {
    [CmdletBinding()]
    param(
        [Parameter(Position = 0)] [ValidateNotNullOrEmpty()] [string[]] $Name,
        [ValidateNotNullOrEmpty()] [ValidatePattern('\A[A-Za-z0-9-]{1,128}\z')] [string[]] $Id,
        [object] $Server,
        [ValidateNotNullOrEmpty()] [string[]] $MediaType
    )

    Get-CloudInventory -Type Image -Server $Server -Id $Id -Name $Name | Where-Object { -not $MediaType -or $MediaType -contains $_.MediaType }
}
