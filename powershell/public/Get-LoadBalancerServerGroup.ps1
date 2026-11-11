function Get-CloudLoadBalancerServerGroup {
    [CmdletBinding(DefaultParameterSetName = 'Inventory')]
    param(
        [Parameter(Mandatory, ParameterSetName = 'ByParent', ValueFromPipeline)] [ValidateNotNullOrEmpty()] [object[]] $LoadBalancer,
        [Parameter(Position = 0)] [ValidateNotNullOrEmpty()] [string[]] $Name,
        [ValidateNotNullOrEmpty()] [ValidatePattern('\A[A-Za-z0-9-]{1,128}\z')] [string[]] $Id,
        [object] $Server
    )

    begin { $targets = [System.Collections.Generic.List[object]]::new() }
    process { if ($PSCmdlet.ParameterSetName -ceq 'ByParent') { $targets.AddRange($LoadBalancer) } }
    end {
        if ($PSCmdlet.ParameterSetName -ceq 'ByParent') {
            Get-CloudInventory -Type LoadBalancerServerGroup -Server $Server -Id $Id -Name $Name -ParentField loadBalancerUuid -ParentType LoadBalancer -Parent ($targets.ToArray())
        } else {
            Get-CloudInventory -Type LoadBalancerServerGroup -Server $Server -Id $Id -Name $Name
        }
    }
}
