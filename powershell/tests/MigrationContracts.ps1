function Assert-CloudMigrationContracts {
    param($Connection, [string] $Endpoint, [pscredential] $Credential, [psmoduleinfo] $Compatibility)

    $exports = 'Connect-VIServer', 'Disconnect-VIServer', 'Get-Cluster', 'Get-HardDisk', 'Get-NetworkAdapter', 'Get-VM', 'Get-VMHost', 'Start-VM'
    Assert-CloudContract ((@($Compatibility.ExportedFunctions.Keys | Sort-Object) -join ',') -ceq ($exports -join ',')) 'S01-S04 compatibility exports the explicit ZStack subset'
    $vms = @(Get-VM -Name '开发*', 'web-*' -Server $Connection)
    $csv = @($vms | Select-Object Name, NumCpu, MemoryGB, PowerState, Id | ConvertTo-Csv -NoTypeInformation)
    Assert-CloudContract ($csv.Count -eq 3 -and $csv[0] -ceq '"Name","NumCpu","MemoryGB","PowerState","Id"' -and
        ($csv | ConvertFrom-Csv)[0].MemoryGB -ceq '1.5' -and @((Get-VM 'missing*' -Server $Connection) | Get-HardDisk).Count -eq 0) 'S01 legacy VM filters preserve CSV columns and empty pipelines'

    $cluster = Get-Cluster -Name 'clus*' -Server $Connection
    $vmhost = Get-VMHost -Location $cluster
    Assert-CloudContract ($cluster.Id -ceq 'cluster-a' -and $vmhost.ClusterId -ceq $cluster.Id -and
        $vmhost.ConnectionState -ceq 'Connected' -and $vmhost.MemoryTotalBytes -eq 8589934592L -and
        $null -eq $vmhost.CpuAvailableHz) 'S02 legacy host and cluster reports preserve associations and units'

    $before = Invoke-RestMethod "http://$Endpoint/test/reads"
    $disks = @($vms | Get-HardDisk)
    $adapters = @($vms | Get-NetworkAdapter)
    $after = Invoke-RestMethod "http://$Endpoint/test/reads"
    Assert-CloudContract ($disks.Count -eq 2 -and $disks[0].CapacityBytes -eq 9007199254740993L -and
        $adapters.Count -eq 2 -and $adapters[0].NetworkName -ceq 'app-net' -and
        $adapters[0].MacAddress -ceq '00:00:5e:00:53:01' -and
        $after.resourceRequests.volumes -eq ($before.resourceRequests.volumes + 1) -and
        $after.resourceRequests.'vm-instances/nics' -eq ($before.resourceRequests.'vm-instances/nics' + 1)) 'S03 legacy disk and NIC pipelines batch parents and preserve report fields'
    $copy = $vms[0] | ConvertTo-Json | ConvertFrom-Json
    $other = Connect-VIServer -Server $Endpoint -Credential $Credential
    try {
        Assert-CloudContract ((Test-CloudRejected { $vms | Get-HardDisk -Server $other -ErrorAction Stop } 'QueryVolume.Conflict') -and
            (Test-CloudRejected { $copy | Get-NetworkAdapter -ErrorAction Stop } 'QueryNetworkAdapter.Conflict') -and
            (Test-CloudRejected { Get-VMHost -Location $cluster -Server $other -ErrorAction Stop } 'QueryVMHost.Conflict')) 'S03 legacy inventory rejects detached objects and wrong servers'
    } finally { $other | Disconnect-VIServer }

    $id = '19191919191919191919191919191919'
    $spec = @{ Name = 'migration-slow-start'; Image = 'image-root'; L3Network = 'l3-a'; InstanceOffering = 'offering-a'
        ResourceId = $id; State = 'Stopped'; Server = $Connection }
    $slow = New-ZStackVM @spec -Confirm:$false -ErrorAction Stop
    $null = $slow | Start-VM -WhatIf -ErrorAction Stop
    $preview = Invoke-RestMethod "http://$Endpoint/test/lifecycle"
    Assert-CloudContract ($null -eq $preview.requests.PSObject.Properties['startVmInstance'] -and
        (Get-VM -Id $id -Server $Connection).RawState -ceq 'Stopped') 'S04 legacy Start-VM WhatIf makes no write'
    $timeout = Test-CloudRejected {
        @($slow, $vms[0]) | Start-VM -TimeoutSeconds 2 -Confirm:$false -ErrorAction Stop
    } 'StartVM.OutcomeUnknown'
    $outcome = Invoke-RestMethod "http://$Endpoint/test/lifecycle"
    $reads = Invoke-RestMethod "http://$Endpoint/test/reads"
    Assert-CloudContract ($timeout -and $outcome.requests.startVmInstance -eq 1 -and
        $reads.lastInventoryCondition -ceq "uuid=$id" -and
        (Get-VM -Id $id -Server $Connection).RawState -ceq 'Stopped') 'S04 legacy start timeout reports unknown outcome without a retry'
}

function Assert-CloudMigrationAssessmentContracts {
    param([string] $Evaluator, [string] $Work, [string[]] $Passed)

    $path = Join-Path $Work 'migration-assessment.json'
    $report = @{ cloud = 'loopback mock'; passed = @($Passed | Where-Object { $_ -cne 'OEM export scopes' }); zstackOnly = $true }
    $report | ConvertTo-Json | Set-Content -LiteralPath $path
    $scoped = & $Evaluator -ContractReport $path
    $oem = $scoped.Cases | Where-Object Id -ceq 'oem-native'
    $collision = $scoped.Cases | Where-Object Id -ceq 'legacy-collision'
    Assert-CloudContract ($scoped.ZStackOnly -and $scoped.Summary.Skipped -eq 1 -and $scoped.SkippedCases[0] -ceq 'oem-native' -and
        $oem.SyntheticCheck -ceq 'Skipped' -and $collision.SyntheticCheck -ceq 'Passed' -and
        $scoped.Summary.InScopePassRate -eq 1 -and $scoped.UnrecordedCases.Count -eq 0) 'ZStack-only assessment excludes the OEM fixture and retains collision evidence'

    $report.passed = @($report.passed | Where-Object { $_ -cne 'S01 legacy VM filters preserve CSV columns and empty pipelines' })
    $report | ConvertTo-Json | Set-Content -LiteralPath $path
    $missing = & $Evaluator -ContractReport $path
    Assert-CloudContract ($missing.Summary.InScope -eq $scoped.Summary.InScope -and
        $missing.Summary.InScopePassed -eq ($scoped.Summary.InScopePassed - 1) -and
        $missing.UnrecordedCases -ccontains 's01-vm-csv') 'unrecorded selected case remains in the assessment denominator'

    $report.Remove('zstackOnly')
    $report.passed = @($Passed) + 'OEM export scopes'
    $report | ConvertTo-Json | Set-Content -LiteralPath $path
    $full = & $Evaluator -ContractReport $path
    Assert-CloudContract (-not $full.ZStackOnly -and $full.SkippedCases.Count -eq 0 -and
        $full.Summary.InScope -eq ($scoped.Summary.InScope + 1) -and
        ($full.Cases | Where-Object Id -ceq 'oem-native').SyntheticCheck -ceq 'Passed') 'legacy contract reports retain the full assessment scope'
}
