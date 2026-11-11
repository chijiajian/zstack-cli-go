function ConvertTo-CloudMemoryBytes {
    param([decimal] $MemoryGB, [string] $Operation)
    if ($MemoryGB -le 0 -or $MemoryGB -gt ([decimal] [long]::MaxValue / 1GB)) {
        throw (New-CloudError -Operation $Operation -ExitCode 2)
    }
    $bytes = $MemoryGB * 1GB
    if ($bytes -ne [decimal]::Truncate($bytes)) {
        throw (New-CloudError -Operation $Operation -ExitCode 2)
    }
    return [long] $bytes
}

function Resolve-CloudDeploymentId {
    param($Reference, [string] $Type, $Connection)
    if ($Reference -is [string]) {
        if ($Reference -cmatch '\A[A-Za-z0-9-]{1,128}\z') { return $Reference }
        throw (New-CloudError -Operation NewVM -ExitCode 2)
    }
    $null = Get-CloudObjectConnection $Reference $Type $Connection.Public NewVM
    return $Reference.Id
}

function Invoke-CloudVMOperation {
    param($Connection, [string[]] $Arguments, [string] $Command, [string] $Operation, [string] $Id,
        [System.Diagnostics.Stopwatch] $Clock, [int] $Budget, [switch] $AllowSkipped)

    $remaining = Get-CloudRemainingTime $Clock $Budget $Operation $Id
    $milliseconds = [long] [Math]::Floor($remaining * 1000)
    $options = @('--uuid-only', '--yes', '--request-timeout', "${milliseconds}ms", '-o', 'json')
    $result = Invoke-CloudConnectedCli $Connection ($Arguments[0..1] + $options + $Arguments[2..($Arguments.Count - 1)]) $remaining
    $response = Convert-CliResult $result $Operation -Mutation -Target $Id
    if ($response['operation'] -cne $Command -or $response['items'] -isnot [array] -or
        $response['items'].Count -ne 1 -or $response['items'][0] -isnot [System.Collections.IDictionary] -or
        $response['items'][0]['uuid'] -cne $Id) {
        throw (New-CloudError -Operation $Operation -Failure InvalidWriteResult -Mutation -Target $Id)
    }
    $item = $response['items'][0]
    if ($AllowSkipped -and $item['status'] -ceq 'skipped') { return }
    if ($item['status'] -cne 'succeeded' -or
        ($Command -cne 'delete' -and ($item['inventory'] -isnot [System.Collections.IDictionary] -or $item['inventory']['uuid'] -cne $Id))) {
        throw (New-CloudError -Operation $Operation -Failure InvalidWriteResult -Mutation -Target $Id)
    }
}

function Invoke-CloudVMLifecycle {
    param([System.Management.Automation.PSCmdlet] $Caller, [string] $ParameterSet, [object[]] $Selectors,
        $Server, [string] $Command, [string] $Operation, [string] $Description, [string] $State,
        [int] $TimeoutSeconds, [int] $PollSeconds = 2, [hashtable] $Changes = @{}, [string[]] $Options = @())

    $null = Get-CloudCliPath 'vmLifecycleV1'
    foreach ($selector in $Selectors) {
        $clock = [System.Diagnostics.Stopwatch]::StartNew()
        try {
            $selection = @{ Server = $Server; TimeoutSeconds = $TimeoutSeconds }
            $selection[$ParameterSet] = $selector
            $target = Resolve-CloudVMTarget @selection
        } catch { $Caller.WriteError($_); continue }
        $id = $target.VM.Id
        if (-not (Invoke-CloudShouldProcess $Caller $target $Description)) { continue }

        $arguments = if ($Command -ceq 'delete') { @('delete', 'instance', $id) } else { @('instance', $Command, $id) }
        $retire = $Command -cin 'delete', 'recover'
        try {
            Invoke-CloudVMOperation $target.Connection ($arguments + $Options) $Command $Operation $id $clock $TimeoutSeconds -AllowSkipped:($Command -cin 'pause', 'resume', 'stop', 'delete')
        } catch {
            if ($retire -and $_.FullyQualifiedErrorId -match '\.OutcomeUnknown(?:,|$)') {
                $target.Connection.VMEpochs[$id] = 1 + [int] $target.Connection.VMEpochs[$id]
            }
            $Caller.WriteError($_)
            continue
        }
        if ($retire) { $target.Connection.VMEpochs[$id] = 1 + [int] $target.Connection.VMEpochs[$id] }

        try {
            if ($State) {
                Wait-CloudVMState $target $clock $TimeoutSeconds $PollSeconds $State
            } else {
                $remaining = Get-CloudRemainingTime $clock $TimeoutSeconds $Operation $id
                $rows = Read-CloudVMInventory $target.Connection -Id $id -TimeoutSeconds $remaining -IncludeDescription:$Changes.ContainsKey('description')
                if ($rows.Count -ne 1) { throw 'Updated VM was not observed.' }
                $row = $rows[0]
                foreach ($field in $Changes.Keys) {
                    if ($row[$field] -cne $Changes[$field]) { throw 'Requested VM field did not match the observed value.' }
                }
                $vm = Convert-CloudVM $row $target.Connection
                if ($Changes.ContainsKey('description')) { $vm | Add-Member -NotePropertyName Description -NotePropertyValue $row['description'] }
                $vm
            }
        } catch { $Caller.WriteError((New-CloudError -Operation $Operation -Failure StateUnverified -Mutation -Target $id)) }
    }
}
