function Assert-CloudPlatformContracts {
    param($Connection, [string] $Endpoint, [pscredential] $Credential)

    $usage = Get-ZStackUsage -Server $Connection
    $memory = $usage.Capacity | Where-Object Metric -eq 'memory'
    $cpu = $usage.Capacity | Where-Object Metric -eq 'cpu'
    Assert-CloudContract ($usage.Capacity.Count -eq 4 -and $memory.Unit -ceq 'bytes' -and
        $memory.Total -eq 9007199254740993L -and $cpu.Unit -ceq 'vCPU' -and $null -eq $cpu.Percent) 'usage exact counters, units and null percentage'
    Assert-CloudContract ($usage.Load.ConnectedHosts -eq 1 -and $usage.Load.SampledHosts -eq 0 -and
        $null -eq $usage.Load.SampleTime -and $null -eq $usage.Load.CpuPercent) 'usage keeps missing load separate from zero'
    Assert-CloudContract ($null -eq (Get-ZStackUsage -Section Capacity -Server $Connection).Load -and
        $null -eq (Get-ZStackUsage -Section Load -Server $Connection).Capacity) 'usage report sections'

    $sets = @(Invoke-ZStackZql "query vmInstance.uuid where name='中文;quoted'; count vmInstance" -Server $Connection -TimeoutSeconds 3600)
    Assert-CloudContract ($sets.Count -eq 2 -and $sets[0].Index -eq 0 -and $sets[0].Records[0]['uuid'] -ceq 'vm-a' -and
        $sets[0].Records[0]['dateText'] -is [string] -and $sets[0].ReturnWith.cpu[0].value -eq 0 -and
        $sets[1].Index -eq 1 -and $sets[1].Total -eq 0 -and $sets[1].Aggregates.max -eq 7) 'ZQL preserves quoted arguments and typed result sets'
    Assert-CloudContract (Test-CloudRejected { Invoke-ZStackZql 'query invalid' -Server $Connection } 'Protocol') 'malformed ZQL results rejected'
    Assert-CloudContract (Test-CloudRejected { Invoke-ZStackZql 'query denied' -Server $Connection } 'AuthenticationFailed') 'ZQL permission failure is stable'
    $before = (Invoke-RestMethod "http://$Endpoint/test/reads").resourceRequests.zql
    Assert-CloudContract ((Test-CloudRejected { Invoke-ZStackZql 'delete vmInstance' -Server $Connection } 'InvalidArguments') -and
        (Invoke-RestMethod "http://$Endpoint/test/reads").resourceRequests.zql -eq $before) 'unsupported ZQL verb never reaches API'

    Assert-CloudContract ((Test-CloudRejected { Invoke-ZStackZql 'query vmInstance.uuid; delete vmInstance' -Server $Connection } 'InvalidArguments') -and
        (Invoke-RestMethod "http://$Endpoint/test/reads").resourceRequests.zql -eq $before) 'unsupported later ZQL statement never reaches API'

    $job = Get-ZStackLongJob -JobId job-succeeded -Server $Connection
    $waited = $job | Wait-ZStackLongJob -TimeoutSeconds 3600
    $waitedById = Wait-ZStackLongJob -Id $job.Id -Server $Connection
    Assert-CloudContract (-not [object]::ReferenceEquals($job, $waited) -and $job.JobId -ceq $job.Id -and $waited.JobId -ceq $job.Id -and
        $waited.JobName -ceq 'Fixture' -and $waited.TargetResourceId -ceq 'vm-a' -and
        $waited.Description -ceq 'fixture job metadata' -and $waited.ExecuteTime -eq 1234 -and
        $waited.CreateDate -ceq $job.CreateDate -and $waited.JobResult -ceq 'synthetic-job-result' -and
        $waitedById.JobName -ceq $waited.JobName -and $waitedById.TargetResourceId -ceq $waited.TargetResourceId) 'wait preserves full long job inventory'
    Assert-CloudContract ((Wait-ZStackLongJob -Id job-unnamed -Server $Connection).Id -ceq 'job-unnamed') 'an unnamed cloud job can be waited on by ID'
    Assert-CloudContract (Test-CloudRejected { Wait-ZStackLongJob -Id job-failed -Server $Connection -ErrorAction Stop } 'JobFailed') 'wait distinguishes failed cloud job'
    Assert-CloudContract (Test-CloudRejected { Wait-ZStackLongJob -Id job-canceled -Server $Connection -ErrorAction Stop } 'JobCanceled') 'wait distinguishes canceled cloud job'
    Assert-CloudContract ((Wait-ZStackLongJob -Id job-failed -State Failed -Server $Connection).State -ceq 'Failed') 'explicit failed state is a valid wait target'
    Assert-CloudContract (Test-CloudRejected { Wait-ZStackLongJob -Id job-running -TimeoutSeconds 1 -Server $Connection -ErrorAction Stop } 'Timeout') 'wait timeout does not cancel the cloud job'
    Assert-CloudContract ((Get-ZStackLongJob -Id job-running -Server $Connection).State -ceq 'Running') 'cloud job remains running after local timeout'
    Assert-CloudContract (Test-CloudRejected { Wait-ZStackLongJob -Id fixture-succeeded -Server $Connection -ErrorAction Stop } 'NotFound') 'wait refuses to fall back to names'
    $suspended = Get-ZStackLongJob -Id job-suspended -Server $Connection
    $suspended | Resume-ZStackLongJob -WhatIf
    Assert-CloudContract (-not (Invoke-RestMethod "http://$Endpoint/test/lifecycle").requests.PSObject.Properties['resume-long-job']) 'resume WhatIf performs no writes'
    $resumed = @(@($suspended, $suspended) | Resume-ZStackLongJob -Confirm:$false -ErrorAction Stop -TimeoutSeconds 3600)
    Assert-CloudContract ($resumed.Count -eq 1 -and ($resumed[0] | Wait-ZStackLongJob).State -ceq 'Succeeded' -and
        (Invoke-RestMethod "http://$Endpoint/test/lifecycle").requests.'resume-long-job' -eq 1) 'resume deduplicates and returns a waitable job'
    Assert-CloudContract ($resumed[0].JobId -ceq $suspended.Id -and $resumed[0].Name -ceq $suspended.Name -and
        $resumed[0].State -ceq 'Succeeded' -and $resumed[0].JobName -ceq 'Fixture' -and
        $resumed[0].TargetResourceId -ceq 'vm-a' -and $resumed[0].Description -ceq 'fixture job metadata' -and
        $resumed[0].ExecuteTime -eq 1234 -and $resumed[0].CreateDate -ceq $suspended.CreateDate -and
        $resumed[0].JobResult -ceq 'synthetic-job-result') 'resume returns the verified state and full long job inventory'
    Assert-CloudContract ((Test-CloudRejected { $suspended | Resume-ZStackLongJob -Confirm:$false -ErrorAction Stop } 'Conflict') -and
        (Invoke-RestMethod "http://$Endpoint/test/lifecycle").requests.'resume-long-job' -eq 1) 'resume rechecks the current state before submitting'

    Assert-CloudContract ((Test-CloudRejected {
        Resume-ZStackLongJob -Id job-resume-stuck -Server $Connection -Confirm:$false -ErrorAction Stop
    } 'OutcomeUnknown') -and (Get-ZStackLongJob -Id job-resume-stuck -Server $Connection).State -ceq 'Suspended' -and
        (Invoke-RestMethod "http://$Endpoint/test/lifecycle").requests.'resume-long-job' -eq 2) 'resume rejects a success response when readback is still suspended'

    $accounts = @(Get-ZStackAccount -Server $Connection)
    $user = Get-ZStackUser 'ali*' -Server $Connection
    Assert-CloudContract ($accounts.Count -eq 105 -and $user.AccountId -ceq $accounts[0].Id -and
        -not $accounts[0].PSObject.Properties['Password'] -and -not $user.PSObject.Properties['Password']) 'IAM pagination and field allowlist'
    Assert-CloudContract ((Get-ZStackRole -Id role-a -Server $Connection).Identity -ceq 'fixture-role' -and
        (Get-ZStackProject -Id project-a -Server $Connection).LinkedAccountId -ceq 'acct-000' -and
        (Get-ZStackPolicy -Id policy-a -Server $Connection).AccountId -ceq 'acct-000') 'IAM role, project and policy projections'
    Assert-CloudContract (@(Get-ZStackPolicy -Id absent -Server $Connection).Count -eq 0 -and
        (Test-CloudRejected { Get-ZStackPolicy -Id denied -Server $Connection } 'AuthenticationFailed') -and
        (Get-ZStackAccount -Id acct-000 -Server $Connection).Id -ceq 'acct-000') 'IAM denial differs from empty inventory and preserves the session'

    $group = Get-ZStackSecurityGroup -Id sg-a -Server $Connection
    $rules = @($group | Get-ZStackSecurityGroupRule)
    Assert-CloudContract ($rules.Count -eq 1 -and $rules[0].SecurityGroupId -ceq $group.Id -and
        $rules[0].Direction -ceq 'Ingress' -and $rules[0].StartPort -eq 443) 'security group scopes rule inventory'
    $lb = Get-ZStackVip -Id vip-a -Server $Connection | Get-ZStackLoadBalancer
    $listener = $lb | Get-ZStackLoadBalancerListener
    $pool = $lb | Get-ZStackLoadBalancerServerGroup
    Assert-CloudContract ($lb.VipId -ceq 'vip-a' -and $listener.LoadBalancerId -ceq $lb.Id -and
        $pool.LoadBalancerId -ceq $lb.Id -and $listener.ServerGroupId -ceq $pool.Id) 'VIP, load balancer, listener and server group identities link'
    $other = Connect-ZStackServer -Server $Endpoint -Credential $Credential
    try {
        Assert-CloudContract ((Test-CloudRejected { $lb | Get-ZStackLoadBalancerListener -Server $other } 'Conflict') -and
            (Test-CloudRejected { $job | Wait-ZStackLongJob -Server $other -ErrorAction Stop } 'Conflict')) 'platform objects refuse cross-connection queries'
    } finally { Disconnect-ZStackServer -Server $other }
}
