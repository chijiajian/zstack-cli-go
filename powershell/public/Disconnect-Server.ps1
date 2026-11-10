function Disconnect-CloudServer {
    [CmdletBinding()]
    param([Parameter(Mandatory, Position = 0, ValueFromPipeline)] [object[]] $Server)

    process {
        foreach ($item in $Server) {
            $connection = Get-CloudConnection $item
            $failure = $null
            try {
                $arguments = @('logout', '--context', $connection.ContextName, '--error-format', 'json')
                $result = Invoke-CliProcess -Path $connection.CliPath -ArgumentList $arguments -ConfigPath $connection.ConfigPath
                $failure = Get-CloudProcessError $result 'Disconnect' -Target $connection.Id
            } finally {
                try { Remove-CloudConnection $connection } catch {
                    $PSCmdlet.WriteError((New-CloudError -Operation Disconnect -ExitCode 9 -Target $connection.Id))
                }
            }
            if ($failure) { $PSCmdlet.WriteError($failure) }
        }
    }
}
