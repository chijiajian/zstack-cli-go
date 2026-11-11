function Assert-CloudInventoryContracts {
    param($Connection, [string] $Endpoint, [object[]] $VMs, [pscredential] $Credential, [psmoduleinfo] $Module, [switch] $ZStackOnly)

    $root, $other, $noDisks = $VMs
    $disks = @($root, $other, $root | Get-ZStackHardDisk)
    $after = Invoke-RestMethod "http://$Endpoint/test/reads"
    Assert-CloudContract ($disks.Count -eq 2 -and $after.resourceRequests.volumes -eq 1 -and
        $after.resourceConditions.volumes -match '^vmInstanceUuid\?=' -and
        $disks[0].PSTypeNames[0] -ceq 'Cloud.Automation.HardDisk' -and
        $disks[0].VMId -ceq 'vm-a' -and $disks[0].CapacityBytes -eq 9007199254740993L -and
        $disks[0].ActualSizeBytes -eq 0 -and $null -eq $disks[1].ActualSizeBytes) 'VM disk pipeline batches live objects, deduplicates and preserves exact bytes'
    $report = @($disks | Select-Object VMId, Name, CapacityBytes | ConvertTo-Csv -NoTypeInformation)
    Assert-CloudContract ($report.Count -eq 3 -and $report[1].Contains('9007199254740993') -and
        @($noDisks | Get-ZStackHardDisk).Count -eq 0) 'disk objects flow through CSV and empty VM inventory emits no rows'
    $forged = $root | ConvertTo-Json | ConvertFrom-Json
    Assert-CloudContract (Test-CloudRejected { $forged | Get-ZStackHardDisk -ErrorAction Stop } 'QueryVolume.Conflict') 'detached VM objects cannot be used for disk lookup'
    Assert-CloudContract (Test-CloudRejected { Get-ZStackVolume -Id wrong-inventory -Server $Connection } 'QueryVolume.Protocol') 'volume lookup rejects unrelated API rows'
    $volume = Get-ZStackVolume -Name 'spa*' -Unattached -Server $Connection
    Assert-CloudContract ($volume.PSTypeNames[0] -ceq 'Cloud.Automation.Volume' -and $volume.Id -ceq 'disk-c' -and
        -not $volume.Attached -and $null -eq $volume.CapacityBytes -and @(Get-ZStackVolume -Attached -Server $Connection).Count -eq 2) 'standalone volume inventory distinguishes current attachment and omitted capacity'

    $secondConnection = Connect-ZStackServer -Server $Endpoint -Credential $Credential
    try {
        $before = Invoke-RestMethod "http://$Endpoint/test/reads"
        Assert-CloudContract (Test-CloudRejected { $root | Get-ZStackHardDisk -Server $secondConnection -ErrorAction Stop } 'QueryVolume.Conflict') 'pipeline refuses an explicit conflicting server'
        $after = Invoke-RestMethod "http://$Endpoint/test/reads"
        Assert-CloudContract ($after.resourceRequests.volumes -eq $before.resourceRequests.volumes -and
            (Test-CloudRejected { Get-ZStackVolume } 'ConnectionRequired')) 'multi-connection reads cannot use the wrong environment'
        $sameId = Get-ZStackVM -Id vm-a -Server $secondConnection
        $separate = @($root, $sameId | Get-ZStackHardDisk)
        Assert-CloudContract ($separate.Count -eq 2 -and $separate[0].Id -ceq $separate[1].Id -and
            $separate[0].ConnectionId -cne $separate[1].ConnectionId) 'same UUID on two connections remains separate in a pipeline'
    } finally { Disconnect-ZStackServer -Server $secondConnection }

    $adapters = @($root, $other | Get-ZStackNetworkAdapter -IncludeNetworkName)
    $requests = Invoke-RestMethod "http://$Endpoint/test/reads"
    Assert-CloudContract ($adapters.Count -eq 2 -and $adapters[0].NetworkName -ceq 'app-net' -and $adapters[0].MacAddress -ceq '00:00:5e:00:53:01' -and
        $adapters[1].VMId -ceq 'vm-b' -and $requests.resourceRequests.'vm-instances/nics' -eq 1 -and $requests.resourceRequests.'l3-networks' -eq 1) 'VM network adapter pipeline batches NIC and verified L3 name lookups'
    Assert-CloudContract ($null -eq ($root | Get-ZStackNetworkAdapter).NetworkName -and
        @($noDisks | Get-ZStackNetworkAdapter).Count -eq 0) 'NIC reports remain usable without L3 enrichment or matching NICs'

    $zone = Get-ZStackZone -Name 'zo*' -Server $Connection
    $cluster = Get-ZStackCluster -Zone $zone
    $vmhost = Get-ZStackVMHost -Cluster $cluster
    Assert-CloudContract ($cluster.ZoneId -ceq $zone.Id -and $vmhost.ClusterId -ceq $cluster.Id -and
        $vmhost.NumCpu -eq 8 -and $vmhost.CpuTotalHz -eq 3000000000L -and $vmhost.MemoryTotalGB -eq 8 -and
        $vmhost.MemoryAvailableBytes -eq 0 -and $null -eq $vmhost.CpuAvailableHz -and $vmhost.ConnectionState -ceq 'Connected') 'zone, cluster and VM host objects support related inventory and numeric reports'
    $primary = Get-ZStackPrimaryStorage -Zone $zone
    $backup = Get-ZStackImageStorage -Id storage-a -Server $Connection
    Assert-CloudContract ($primary.CapacityGB -eq 4 -and $primary.FreeSpaceBytes -eq 0 -and $primary.PhysicalCapacityBytes -eq 2147483648L -and
        $primary.SystemUsedCapacityBytes -eq -4294967297L -and $null -eq $primary.PhysicalFreeSpaceBytes -and
        $backup.CapacityGB -eq 1 -and $null -eq $backup.FreeSpaceGB) 'storage reports preserve logical and physical capacity, signed usage, zero and unknown values'
    $image = Get-ZStackImage -MediaType ISO -Server $Connection
    $compute = Get-ZStackInstanceOffering -Id offering-a -Server $Connection
    $diskOffering = Get-ZStackDiskOffering -Name 'one*' -Server $Connection
    Assert-CloudContract ($image.Id -ceq 'image-a' -and $image.CapacityGB -eq 1 -and $compute.NumCpu -eq 2 -and $compute.MemoryGB -eq 2 -and
        $diskOffering.CapacityBytes -eq 1073741824L -and (Get-ZStackImage -MediaType RootVolumeTemplate -Server $Connection).Id -ceq 'image-root') 'deployment inventory distinguishes image media, compute and disk offerings'
    $l2 = Get-ZStackL2Network -Zone $zone
    $l3 = Get-ZStackL3Network -L2Network $l2
    $range = Get-ZStackIPRange -L3Network $l3
    Assert-CloudContract ($l2.VirtualNetworkId -eq 0 -and $l3.L2NetworkId -ceq $l2.Id -and $range.L3NetworkId -ceq $l3.Id -and
        $range.CIDR -ceq '192.0.2.0/24') 'L2, L3 and IP range queries preserve typed hierarchy and network ID zero'

    $detail = Get-ZStackVM -Id vm-a -Detail -Server $Connection
    Assert-CloudContract ($detail.Id -ceq $root.Id -and $detail.HostName -ceq 'esx-like' -and $detail.Description -ceq 'fixture detail' -and
        $detail.MemoryGB -eq 1.5 -and $detail.L3NetworkIds[0] -ceq 'l3-a' -and $detail.Warnings.Count -eq 1 -and
        $detail.Warnings[0] -ceq 'image missing-image unavailable: not_found' -and
        $null -eq $root.PSObject.Properties['Warnings'] -and (Test-CloudRejected { Get-ZStackVM -Name '开发*' -Detail -Server $Connection } 'DescribeVM.InvalidArguments')) 'VM detail preserves default list shape and exposes related-resource warnings'
    $legacyWarnings = & $Module {
        Convert-CloudDetailWarning 'image missing-image unavailable: GET v1/images/missing-image: HTTP 404: synthetic-related-response-marker'
        Convert-CloudDetailWarning 'synthetic-related-response-marker'
    }
    Assert-CloudContract ($legacyWarnings.Count -eq 2 -and $legacyWarnings[0] -ceq 'image missing-image unavailable: operation_failed' -and
        $legacyWarnings[1] -ceq 'related resource unavailable: operation_failed') 'VM detail drops legacy raw response text and unknown warning content'

    if ($ZStackOnly) { return }
    $oem = Connect-AcmeServer -Server $Endpoint -Credential $Credential
    try {
        $oemVm = Get-AcmeVM -Id vm-a -Server $oem
        $oemDisk = $oemVm | Get-AcmeHardDisk
        Assert-CloudContract ($oemDisk.Id -ceq $disks[0].Id -and $oemDisk.PSTypeNames[0] -ceq $disks[0].PSTypeNames[0] -and
            (Test-CloudRejected { $root | Get-AcmeHardDisk -ErrorAction Stop } 'QueryVolume.Conflict')) 'OEM inventory shares object shape and rejects another brand live VM'
    } finally { Disconnect-AcmeServer -Server $oem }
}
