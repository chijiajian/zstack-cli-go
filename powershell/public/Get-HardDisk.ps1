function Get-CloudHardDisk {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory, Position = 0, ValueFromPipeline)] [ValidateNotNullOrEmpty()] [object[]] $VM,
        [ValidateNotNullOrEmpty()] [string[]] $Name,
        [ValidateNotNullOrEmpty()] [ValidatePattern('\A[A-Za-z0-9-]{1,128}\z')] [string[]] $Id,
        [object] $Server
    )

    begin { $targets = [System.Collections.Generic.List[object]]::new() }
    process { $targets.AddRange($VM) }
    end {
        Get-CloudInventory -Type Volume -Server $Server -Id $Id -Name $Name -ParentField vmInstanceUuid -ParentType VM -Parent ($targets.ToArray()) -OutputType HardDisk
    }
}
