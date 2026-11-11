function Assert-CloudDeviceContracts {
    param($Connection, [string] $Endpoint, [pscredential] $Credential)
    $vm = Get-ZStackVM -Id vm-b -Server $Connection
    $snapshotVM = New-ZStackVM -Name device-fixture -Image image-root -L3Network l3-a -NumCpu 1 -MemoryGB 1 -State Stopped -Server $Connection -Confirm:$false
    $id = '997a713f4cd04700825dd117b6ec212f'
    $opts = @{ Confirm = $false; ErrorAction = 'Stop' }
    $calls = "http://$Endpoint/test/lifecycle"
    $before = (Invoke-WebRequest $calls).Content | ConvertFrom-Json -AsHashtable
    New-ZStackVolume -Name data -CapacityGB 1 -ResourceId $id -Server $Connection -WhatIf @opts
    $after = (Invoke-WebRequest $calls).Content | ConvertFrom-Json -AsHashtable
    Assert-CloudContract ($before.requests['create-volume'] -eq $after.requests['create-volume']) 'volume WhatIf performs no writes'
    $volume = New-ZStackVolume -Name data -CapacityGB 1 -ResourceId $id -Server $Connection @opts
    Assert-CloudContract ($volume.Id -ceq $id -and $volume.CapacityBytes -eq 1GB -and -not $volume.Attached) 'create data volume exact identity and bytes'
    Assert-CloudContract (Test-CloudRejected { New-ZStackVolume -Name data -CapacityGB 1 -ResourceId $id -Server $Connection @opts } 'Conflict') 'reserved volume duplicate rejected'
    Assert-CloudContract (Test-CloudRejected { New-ZStackVolume -Name reject-volume -CapacityGB 1 -Server $Connection @opts } 'AuthenticationFailed') 'volume permission denial is sanitized'
    Assert-CloudContract (Test-CloudRejected {
        New-ZStackVolume -Name misreported-volume -CapacityGB 1 -Server $Connection @opts
    } 'OutcomeUnknown') 'volume size mismatch does not report success'
    $disk = $volume | Add-ZStackHardDisk -VM $snapshotVM @opts
    $again = $disk | Add-ZStackHardDisk -VM $snapshotVM @opts
    Assert-CloudContract ($again.VMId -ceq $snapshotVM.Id -and (Invoke-RestMethod $calls).requests.'attach-volume' -eq 1) 'attachment is idempotent'
    Assert-CloudContract (Test-CloudRejected { $disk | Add-ZStackHardDisk -VM $vm @opts } 'Conflict') 'another VM cannot attach the same data volume'
    $other = Connect-ZStackServer -Server $Endpoint -Credential $Credential
    try {
        Assert-CloudContract (Test-CloudRejected { $disk | Set-ZStackHardDisk -CapacityGB 2 -Server $other @opts } 'Conflict') 'device rejects cross-connection mutation'
    } finally { Disconnect-ZStackServer -Server $other }
    $disk = $disk | Set-ZStackHardDisk -CapacityGB 1.5 @opts
    Assert-CloudContract ($disk.CapacityBytes -eq 1610612736) 'resize transmits an exact total byte count'
    Assert-CloudContract (Test-CloudRejected { $disk | Set-ZStackHardDisk -CapacityGB 1 @opts } 'Conflict') 'shrink rejected'
    $snap = $disk | New-ZStackVolumeSnapshot -Name smoke @opts
    Assert-CloudContract (($disk | Get-ZStackVolumeSnapshot).Id -ceq $snap.Id -and $snap.VolumeId -ceq $disk.Id) 'snapshot query is scoped to volume'
    Assert-CloudContract (Test-CloudRejected { $snap | Restore-ZStackVolumeSnapshot -Volume disk-b @opts } 'Conflict') 'snapshot rejects the wrong parent volume'
    Assert-CloudContract (($snap | Restore-ZStackVolumeSnapshot @opts).Id -ceq $snap.Id) 'snapshot revert targets the original volume'
    $snap | Remove-ZStackVolumeSnapshot @opts
    Assert-CloudContract (@($disk | Get-ZStackVolumeSnapshot).Count -eq 0) 'snapshot delete verified absent'
    $volume = $disk | Remove-ZStackHardDisk @opts
    $again = $disk | Remove-ZStackHardDisk @opts
    Assert-CloudContract (-not $volume.Attached -and $volume.Id -ceq $again.Id -and (Invoke-RestMethod $calls).requests.'detach-volume' -eq 1) 'detach preserves the same volume'
    $volume | Remove-ZStackVolume @opts
    $deleted = Get-ZStackVolume -Id $id -Server $Connection
    Assert-CloudContract ($deleted.Status -ceq 'Deleted') 'volume reaches recoverable Deleted state'
    $restored = $deleted | Restore-ZStackVolume @opts
    Assert-CloudContract ($restored.Id -ceq $id -and $restored.Status -ceq 'Ready') 'volume recovery preserves UUID'

    $nic = $vm | New-ZStackNetworkAdapter -L3Network l3-a -MacAddress '02:00:5E:10:00:AB' -StaticIP '10.255.255.201' @opts
    Assert-CloudContract ($nic.MacAddress -ceq '02:00:5e:10:00:ab' -and $nic.IPAddress -ceq '10.255.255.201') 'new NIC reads back the requested MAC and IP'
    Assert-CloudContract ($nic.Id -ceq 'nic-new' -and @($vm | Get-ZStackNetworkAdapter).Count -eq 2) 'new NIC distinct on same L3'
    $nic | Remove-ZStackNetworkAdapter @opts
    Assert-CloudContract (($vm | Get-ZStackNetworkAdapter).Id -ceq 'nic-b') 'remove exact NIC preserves sibling on same L3'

    Assert-CloudContract (Test-CloudRejected {
        $vm | New-ZStackNetworkAdapter -L3Network l3-a -StaticIP '10.255.255.202' @opts
    } 'OutcomeUnknown') 'static IP mismatch does not report success'
    $vm | Get-ZStackNetworkAdapter -Id nic-new | Remove-ZStackNetworkAdapter @opts

    $cdVM = Get-ZStackVM -Id vm-a -Server $Connection
    $cd = $cdVM | Get-ZStackCDDrive -Id cdrom-a
    $cd = $cd | Set-ZStackCDDrive -IsoImage image-a @opts
    Assert-CloudContract ($cd.IsoImageId -ceq 'image-a') 'selected CD drive receives requested ISO'
    $cd = $cd | Set-ZStackCDDrive -NoMedia @opts
    $otherCD = $cdVM | Get-ZStackCDDrive -Id cdrom-b
    $eject = (Invoke-RestMethod $calls).payloads.'detach-iso'
    Assert-CloudContract (-not $cd.IsoImageId -and $otherCD.IsoImageId -ceq 'image-other' -and
        $eject.path -ceq 'vm-instances/vm-a/iso' -and $eject.query.isoUuid[0] -ceq 'image-a') 'ISO eject preserves VM and other drive'
}
