# These inventories share the connection, paging, and identity checks of the core inventory.
foreach ($entry in ([ordered]@{
    LongJob = @{ Command = 'long-jobs'; Capability = 'longJobV1'; Properties = [ordered]@{
        Name = 'name'; State = 'state'; JobName = 'jobName'; TargetResourceId = 'targetResourceUuid'
        Description = 'description'; ExecuteTime = 'executeTime'; CreateDate = 'createDate'; JobResult = 'jobResult'
    } }
    Account = @{ Command = 'accounts'; Capability = 'platformInventoryV1'; Properties = [ordered]@{
        Name = 'name'; Type = 'type'; Description = 'description'; CreateDate = 'createDate'
    } }
    User = @{ Command = 'users'; Capability = 'platformInventoryV1'; Properties = [ordered]@{
        Name = 'name'; AccountId = 'accountUuid'; Description = 'description'; CreateDate = 'createDate'
    } }
    Role = @{ Command = 'roles'; Capability = 'platformInventoryV1'; Properties = [ordered]@{
        Name = 'name'; Type = 'type'; State = 'state'; Identity = 'identity'; Description = 'description'; CreateDate = 'createDate'
    } }
    Policy = @{ Command = 'policies'; Capability = 'platformInventoryV1'; Properties = [ordered]@{
        Name = 'name'; AccountId = 'accountUuid'; StatementCount = 'statementCount'; CreateDate = 'createDate'
    } }
    Project = @{ Command = 'projects'; Capability = 'platformInventoryV1'; Properties = [ordered]@{
        Name = 'name'; State = 'state'; LinkedAccountId = 'linkedAccountUuid'; Description = 'description'; CreateDate = 'createDate'
    } }
    SecurityGroup = @{ Command = 'security-groups'; Capability = 'platformInventoryV1'; Properties = [ordered]@{
        Name = 'name'; State = 'state'; IPVersion = 'ipVersion'; Description = 'description'
    } }
    SecurityGroupRule = @{ Command = 'security-group-rules'; Capability = 'platformInventoryV1'; Properties = [ordered]@{
        SecurityGroupId = 'securityGroupUuid'; Direction = 'type'; Protocol = 'protocol'; StartPort = 'startPort'
        EndPort = 'endPort'; AllowedCidr = 'allowedCidr'; Action = 'action'; State = 'state'
    } }
    Vip = @{ Command = 'vips'; Capability = 'platformInventoryV1'; Properties = [ordered]@{
        Name = 'name'; L3NetworkId = 'l3NetworkUuid'; IPAddress = 'ip'; State = 'state'; Gateway = 'gateway'
        Netmask = 'netmask'; PrefixLength = 'prefixLen'; ServiceProvider = 'serviceProvider'; System = 'system'; UseFor = 'useFor'
    } }
    LoadBalancer = @{ Command = 'load-balancers'; Capability = 'platformInventoryV1'; Properties = [ordered]@{
        Name = 'name'; VipId = 'vipUuid'; State = 'state'; Description = 'description'
    } }
    LoadBalancerListener = @{ Command = 'load-balancer-listeners'; Capability = 'platformInventoryV1'; Properties = [ordered]@{
        Name = 'name'; LoadBalancerId = 'loadBalancerUuid'; Protocol = 'protocol'; LoadBalancerPort = 'loadBalancerPort'
        InstancePort = 'instancePort'; ServerGroupId = 'serverGroupUuid'
    } }
    LoadBalancerServerGroup = @{ Command = 'load-balancer-server-groups'; Capability = 'platformInventoryV1'; Properties = [ordered]@{
        Name = 'name'; LoadBalancerId = 'loadBalancerUuid'; IPVersion = 'ipVersion'; Description = 'description'
    } }
}).GetEnumerator()) {
    $script:InventoryQueries[$entry.Key] = $entry.Value
}
