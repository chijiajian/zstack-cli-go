function Assert-CloudNameQueryContracts {
    param($Connection, [object[]] $VMs, [psmoduleinfo] $Module)

    # One row per public Get cmdlet with a Name parameter. The command set
    # assertion below makes additions to the module require a corresponding
    # name-query case instead of silently inheriting an untested selector.
    $cases = @(
        @{ Command = 'Get-ZStackAccount'; Name = 'account-104'; Id = 'acct-104' }
        @{ Command = 'Get-ZStackCluster'; Name = 'cluster'; Id = 'cluster-a' }
        @{ Command = 'Get-ZStackDiskOffering'; Name = 'one-gib'; Id = 'offering-disk' }
        @{ Command = 'Get-ZStackHardDisk'; Name = 'system'; Id = 'disk-a' }
        @{ Command = 'Get-ZStackIPRange'; Name = 'app-range'; Id = 'range-a' }
        @{ Command = 'Get-ZStackImage'; Name = 'installer'; Id = 'image-a' }
        @{ Command = 'Get-ZStackImageStorage'; Name = 'backup'; Id = 'storage-a' }
        @{ Command = 'Get-ZStackInstanceOffering'; Name = 'small'; Id = 'offering-a' }
        @{ Command = 'Get-ZStackL2Network'; Name = 'transport'; Id = 'l2-a' }
        @{ Command = 'Get-ZStackL3Network'; Name = 'app-net'; Id = 'l3-a' }
        @{ Command = 'Get-ZStackLoadBalancer'; Name = 'edge'; Id = 'lb-a' }
        @{ Command = 'Get-ZStackLoadBalancerListener'; Name = 'https'; Id = 'listener-a' }
        @{ Command = 'Get-ZStackLoadBalancerServerGroup'; Name = 'apps'; Id = 'pool-a' }
        @{ Command = 'Get-ZStackLongJob'; Name = 'fixture Succeeded'; Id = 'job-succeeded' }
        @{ Command = 'Get-ZStackPolicy'; Name = 'audit'; Id = 'policy-a' }
        @{ Command = 'Get-ZStackPrimaryStorage'; Name = 'primary'; Id = 'primary-a' }
        @{ Command = 'Get-ZStackProject'; Name = 'project'; Id = 'project-a' }
        @{ Command = 'Get-ZStackRole'; Name = 'operator'; Id = 'role-a' }
        @{ Command = 'Get-ZStackSecurityGroup'; Name = 'apps'; Id = 'sg-a' }
        @{ Command = 'Get-ZStackUser'; Name = 'alice'; Id = 'user-a' }
        @{ Command = 'Get-ZStackVM'; Name = 'web-prod'; Id = 'vm-b' }
        @{ Command = 'Get-ZStackVMHost'; Name = 'esx-like'; Id = 'host-a' }
        @{ Command = 'Get-ZStackVip'; Name = 'front door'; Id = 'vip-a' }
        @{ Command = 'Get-ZStackVolume'; Name = 'spare'; Id = 'disk-c' }
        @{ Command = 'Get-ZStackVolumeSnapshot'; Name = 'baseline'; Id = 'snap-baseline' }
        @{ Command = 'Get-ZStackZone'; Name = 'zone'; Id = 'zone-a' }
    )
    $expected = @($cases | ForEach-Object Command | Sort-Object -Unique)
    $actual = @($Module.ExportedFunctions.Values | Where-Object {
        $_.Name -clike 'Get-*' -and $_.Parameters.ContainsKey('Name')
    } | ForEach-Object Name | Sort-Object -Unique)
    Assert-CloudContract (($actual -join '|') -ceq ($expected -join '|')) 'every public Get Name selector has a fixture case'

    foreach ($case in $cases) {
        $command = [string] $case.Command
        $arguments = @{ Name = $case.Name.ToUpperInvariant(); Server = $Connection }
        if ($command -ceq 'Get-ZStackHardDisk') { $arguments.VM = $VMs[0] }
        $found = @(& $command @arguments)
        Assert-CloudContract ($found.Count -eq 1 -and $found[0].Id -ceq $case.Id -and
            $found[0].Name -ceq $case.Name) "$command matches a complete name without case sensitivity"
    }

    $missing = @(Get-ZStackVolumeSnapshot -Name 'absent' -Server $Connection)
    Assert-CloudContract ($missing.Count -eq 0) 'named resource query emits no object for an absent name'
    $wildcard = @(Get-ZStackSecurityGroup -Name 'app*' -Server $Connection)
    Assert-CloudContract ($wildcard.Count -eq 1 -and $wildcard[0].Id -ceq 'sg-a') 'named resource query keeps PowerShell wildcard matching'
}
