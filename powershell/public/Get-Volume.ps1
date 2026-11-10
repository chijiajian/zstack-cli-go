function Get-CloudVolume {
    [CmdletBinding()]
    param(
        [Parameter(Position = 0)] [ValidateNotNullOrEmpty()] [string[]] $Name,
        [ValidateNotNullOrEmpty()] [ValidatePattern('\A[A-Za-z0-9-]{1,128}\z')] [string[]] $Id,
        [object] $Server,
        [switch] $Attached,
        [switch] $Unattached
    )

    if ($Attached -and $Unattached) { throw (New-CloudError -Operation QueryVolume -ExitCode 2) }
    Get-CloudInventory -Type Volume -Server $Server -Id $Id -Name $Name | Where-Object {
        (-not $Attached -or $_.Attached) -and (-not $Unattached -or -not $_.Attached)
    }
}
