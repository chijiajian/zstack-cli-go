function Invoke-CloudShouldProcess {
    param([System.Management.Automation.PSCmdlet] $Caller, $Target, [string] $Operation)

    $description = '{0} ({1}) at {2}' -f $Target.VM.Name, $Target.VM.Id, $Target.Connection.Endpoint
    return $Caller.ShouldProcess($description, $Operation)
}
