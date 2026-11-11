function Get-CloudCDDrive {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory, Position = 0, ValueFromPipeline)] [ValidateNotNullOrEmpty()] [object[]] $VM,
        [ValidatePattern('\A[A-Za-z0-9-]{1,128}\z')] [string[]] $Id,
        [object] $Server
    )
    begin { $targets = [System.Collections.Generic.List[object]]::new() }
    process { $targets.AddRange($VM) }
    end { Get-CloudInventory -Type CDDrive -Server $Server -Id $Id -ParentField vmInstanceUuid -ParentType VM -Parent ($targets.ToArray()) }
}
