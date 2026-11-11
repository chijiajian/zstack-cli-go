function Assert-CloudLifecycleContracts {
    param($Connection, [string] $Endpoint, [pscredential] $Credential)

    $image = Get-ZStackImage -Id image-root -Server $Connection
    $network = Get-ZStackL3Network -Id l3-a -Server $Connection
    $offering = Get-ZStackInstanceOffering -Id offering-a -Server $Connection
    $id = '11111111222233334444555555555555'
    $spec = @{ Name = '-QA VM'; Image = $image; L3Network = $network; InstanceOffering = $offering
        State = 'Stopped'; ResourceId = $id; Server = $Connection }
    $preview = @(New-ZStackVM @spec -WhatIf -ErrorAction Stop)
    Assert-CloudContract ($preview.Count -eq 0 -and @(Get-ZStackVM -Id $id -Server $Connection).Count -eq 0) 'VM creation WhatIf returns no VM and performs no write'
    $vm = New-ZStackVM @spec -Confirm:$false -ErrorAction Stop
    $data = Invoke-RestMethod "http://$Endpoint/test/lifecycle"
    Assert-CloudContract ($vm.Id -ceq $id -and $vm.Name -ceq '-QA VM' -and $vm.RawState -ceq 'Stopped' -and
        $data.requests.create -eq 1 -and $data.payloads.create.strategy -ceq 'CreateStopped' -and
        $data.payloads.create.instanceOfferingUuid -ceq $offering.Id -and $data.payloads.create.l3NetworkUuids[0] -ceq $network.Id) 'VM creation uses exact deployment objects, explicit state and a reserved UUID'
    Assert-CloudContract (Test-CloudRejected { New-ZStackVM @spec -Confirm:$false -ErrorAction Stop } 'NewVM.Conflict') 'VM creation rejects a duplicate reserved UUID'
    $other = Connect-ZStackServer -Server $Endpoint -Credential $Credential
    try {
        $mismatch = $spec.Clone()
        $mismatch.Server = $other
        Assert-CloudContract (Test-CloudRejected { New-ZStackVM @mismatch -WhatIf -ErrorAction Stop } 'NewVM.Conflict') 'VM creation rejects deployment objects from another connection'
    } finally { Disconnect-ZStackServer -Server $other }
    $dynamic = @{ Image = 'image-root'; L3Network = 'l3-a'; NumCpu = 2; MemoryGB = 1.5
        State = 'Running'; Server = $Connection; Confirm = $false; ErrorAction = 'Stop' }
    $running = New-ZStackVM -Name running-vm @dynamic
    Assert-CloudContract ($running.RawState -ceq 'Running' -and $running.MemoryGB -eq 1.5) 'VM creation accepts explicit sizing and InstantStart'
    Assert-CloudContract (Test-CloudRejected { New-ZStackVM -Name reject-vm @dynamic } 'NewVM.AuthenticationFailed') 'VM creation surfaces a rejected write'
    $unknownId = '22222222333344445555666666666666'
    Assert-CloudContract ((Test-CloudRejected { New-ZStackVM -Name unknown-vm -ResourceId $unknownId @dynamic } 'NewVM.OutcomeUnknown') -and
        (Get-ZStackVM -Id $unknownId -Server $Connection).Id -ceq $unknownId) 'VM creation identifies the attempted UUID after an unknown outcome'

    $renamed = $vm | Set-ZStackVM -Name 'renamed' -Description '' -Confirm:$false -ErrorAction Stop
    $data = Invoke-RestMethod "http://$Endpoint/test/lifecycle"
    Assert-CloudContract ($renamed.Name -ceq 'renamed' -and $renamed.Description -ceq '' -and
        @($data.payloads.updateVmInstance.PSObject.Properties).Count -eq 2 -and $renamed.NumCpu -eq $vm.NumCpu) 'VM metadata update sends only bound fields and preserves an empty description'
    $sized = $renamed | Set-ZStackVM -NumCpu 3 -MemoryGB 2.5 -Confirm:$false -ErrorAction Stop
    Assert-CloudContract ($sized.NumCpu -eq 3 -and $sized.MemoryGB -eq 2.5 -and
        (Test-CloudRejected { $running | Set-ZStackVM -NumCpu 3 -Confirm:$false -ErrorAction Stop } 'SetVM.Conflict')) 'VM sizing succeeds when stopped and rejects hot resizing'

    $null = $running | Suspend-ZStackVM -WhatIf -ErrorAction Stop
    $paused = $running | Suspend-ZStackVM -Confirm:$false -ErrorAction Stop
    $pausedAgain = $paused | Suspend-ZStackVM -Confirm:$false -ErrorAction Stop
    $resumed = $paused | Resume-ZStackVM -Confirm:$false -ErrorAction Stop
    $resumedAgain = $resumed | Resume-ZStackVM -Confirm:$false -ErrorAction Stop
    $data = Invoke-RestMethod "http://$Endpoint/test/lifecycle"
    Assert-CloudContract ($pausedAgain.RawState -ceq 'Paused' -and $resumedAgain.RawState -ceq 'Running' -and
        $data.requests.pauseVmInstance -eq 1 -and $data.requests.resumeVmInstance -eq 1) 'pause and resume are idempotent and WhatIf sends no write'
    $runningId = $running.Id
    $beforeBoot = $data.bootIds.$runningId
    $restarted = $resumed | Restart-ZStackVM -Confirm:$false -ErrorAction Stop
    $data = Invoke-RestMethod "http://$Endpoint/test/lifecycle"
    Assert-CloudContract ($restarted.RawState -ceq 'Running' -and $data.requests.rebootVmInstance -eq 1 -and
        $data.bootIds.$runningId -eq ($beforeBoot + 1)) 'VM reboot waits for the accepted API job and independent fixture boot marker'

    $stopped = $restarted | Stop-ZStackVM -Confirm:$false -ErrorAction Stop
    $data = Invoke-RestMethod "http://$Endpoint/test/lifecycle"
    Assert-CloudContract ($stopped.RawState -ceq 'Stopped' -and $data.payloads.stopVmInstance.type -ceq 'grace' -and
        $null -eq $data.payloads.stopVmInstance.PSObject.Properties['stopHA']) 'VM stop uses graceful shutdown and omits HA override by default'
    Assert-CloudContract ((Wait-ZStackVM -Id $stopped.Id -State stopped -Server $Connection -ErrorAction Stop).RawState -ceq 'Stopped' -and
        (Test-CloudRejected { Wait-ZStackVM -Id missing-vm -State Stopped -Server $Connection -ErrorAction Stop } 'NotFound') -and
        (Test-CloudRejected { Wait-ZStackVM -Id $stopped.Id -State Running -TimeoutSeconds 1 -PollSeconds 1 -Server $Connection -ErrorAction Stop } 'Timeout')) 'VM wait observes exact states and reports missing targets and timeouts'

    $root = $sized | Get-ZStackHardDisk
    $null = $sized | Remove-ZStackVM -WhatIf -ErrorAction Stop
    $destroyed = $sized | Remove-ZStackVM -Confirm:$false -ErrorAction Stop
    Assert-CloudContract ($destroyed.RawState -ceq 'Destroyed' -and
        (Wait-ZStackVM -VM $destroyed -State Destroyed -ErrorAction Stop).RawState -ceq 'Destroyed' -and
        (Test-CloudRejected { Wait-ZStackVM -VM $destroyed -ErrorAction Stop } 'WaitVM.Conflict') -and
        (Test-CloudRejected { $sized | Wait-ZStackVM -State Destroyed -ErrorAction Stop } 'ResolveVM.Conflict')) 'VM removal reports Destroyed and invalidates objects from before deletion'
    $restored = $destroyed | Restore-ZStackVM -Confirm:$false -ErrorAction Stop
    $restoredRoot = $restored | Get-ZStackHardDisk
    $data = Invoke-RestMethod "http://$Endpoint/test/lifecycle"
    Assert-CloudContract ($restored.RawState -ceq 'Stopped' -and $restoredRoot.Id -ceq $root.Id -and
        $restoredRoot.Status -ceq 'Ready' -and $data.requests.delete -eq 1 -and $data.payloads.delete.mode -ceq 'Permissive' -and
        (Test-CloudRejected { $destroyed | Wait-ZStackVM -State Stopped -ErrorAction Stop } 'ResolveVM.Conflict')) 'VM recovery returns the same UUID and root volume and invalidates the destroyed object'
}
