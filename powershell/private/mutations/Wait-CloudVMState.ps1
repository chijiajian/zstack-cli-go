function Get-CloudRemainingTime {
    param([System.Diagnostics.Stopwatch] $Clock, [double] $Budget, [string] $Operation, [string] $Id)

    $remaining = $Budget - $Clock.Elapsed.TotalSeconds
    if ($remaining -lt 0.001) { throw (New-CloudError -Operation $Operation -ExitCode 6 -Target $Id) }
    return $remaining
}

function Wait-CloudVMState {
    param($Target, [System.Diagnostics.Stopwatch] $Clock, [double] $Budget, [int] $PollSeconds)

    $id = $Target.VM.Id
    $remaining = Get-CloudRemainingTime $Clock $Budget 'WaitVM' $id
    $milliseconds = [Math]::Max(1, [long] [Math]::Floor($remaining * 1000))
    $arguments = @('wait', 'instance', $id, '--for', 'Running', '--timeout', "${milliseconds}ms", '--interval', "${PollSeconds}s", '-o', 'json')
    $result = Invoke-CloudConnectedCli $Target.Connection $arguments $remaining
    $observation = Convert-CliResult $result 'WaitVM' -Target $id
    if ($observation['uuid'] -cne $id -or $observation['state'] -cne 'Running') {
        throw (New-CloudError -Operation WaitVM -Failure WrongObservation -Target $id)
    }
    $remaining = Get-CloudRemainingTime $Clock $Budget 'WaitVM' $id
    $vm = Get-CloudTargetVM $Target.Connection -Id $id -TimeoutSeconds $remaining
    if ($vm.RawState -cne 'Running') { throw (New-CloudError -Operation WaitVM -ExitCode 5 -Target $id) }
    return $vm
}
