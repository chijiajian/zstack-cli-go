function Get-CloudLoadBalancer {
    [CmdletBinding(DefaultParameterSetName = 'Inventory')]
    param(
        [Parameter(Mandatory, ParameterSetName = 'ByParent', ValueFromPipeline)] [ValidateNotNullOrEmpty()] [object[]] $Vip,
        [Parameter(Position = 0)] [ValidateNotNullOrEmpty()] [string[]] $Name,
        [ValidateNotNullOrEmpty()] [ValidatePattern('\A[A-Za-z0-9-]{1,128}\z')] [string[]] $Id,
        [object] $Server
    )

    begin { $targets = [System.Collections.Generic.List[object]]::new() }
    process { if ($PSCmdlet.ParameterSetName -ceq 'ByParent') { $targets.AddRange($Vip) } }
    end {
        if ($PSCmdlet.ParameterSetName -ceq 'ByParent') {
            Get-CloudInventory -Type LoadBalancer -Server $Server -Id $Id -Name $Name -ParentField vipUuid -ParentType Vip -Parent ($targets.ToArray())
        } else {
            Get-CloudInventory -Type LoadBalancer -Server $Server -Id $Id -Name $Name
        }
    }
}
