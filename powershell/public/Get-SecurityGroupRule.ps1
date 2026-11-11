function Get-CloudSecurityGroupRule {
    [CmdletBinding(DefaultParameterSetName = 'Inventory')]
    param(
        [Parameter(Mandatory, ParameterSetName = 'ByParent', ValueFromPipeline)] [ValidateNotNullOrEmpty()] [object[]] $SecurityGroup,
        [ValidateNotNullOrEmpty()] [ValidatePattern('\A[A-Za-z0-9-]{1,128}\z')] [string[]] $Id,
        [object] $Server
    )

    begin { $targets = [System.Collections.Generic.List[object]]::new() }
    process { if ($PSCmdlet.ParameterSetName -ceq 'ByParent') { $targets.AddRange($SecurityGroup) } }
    end {
        if ($PSCmdlet.ParameterSetName -ceq 'ByParent') {
            Get-CloudInventory -Type SecurityGroupRule -Server $Server -Id $Id -ParentField securityGroupUuid -ParentType SecurityGroup -Parent ($targets.ToArray())
        } else {
            Get-CloudInventory -Type SecurityGroupRule -Server $Server -Id $Id
        }
    }
}
