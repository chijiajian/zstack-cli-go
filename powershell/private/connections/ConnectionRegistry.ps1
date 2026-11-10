$script:Connections = @{}

function New-CloudConnectionDirectory {
    $namespace = [string] $script:Package.brand.configNamespace
    if ($namespace -cnotmatch '\A[a-z][a-z0-9-]{0,39}\z') { throw 'Invalid connection namespace.' }
    if (-not $IsWindows) { return [System.IO.Directory]::CreateTempSubdirectory("$namespace-").FullName }

    $sid = [System.Security.Principal.WindowsIdentity]::GetCurrent().User
    $security = [System.Security.AccessControl.DirectorySecurity]::new()
    $security.SetAccessRuleProtection($true, $false)
    $inheritance = [System.Security.AccessControl.InheritanceFlags]::ContainerInherit -bor [System.Security.AccessControl.InheritanceFlags]::ObjectInherit
    $rule = [System.Security.AccessControl.FileSystemAccessRule]::new($sid, 'FullControl', $inheritance, 'None', 'Allow')
    $security.AddAccessRule($rule)
    $path = Join-Path ([System.IO.Path]::GetTempPath()) "$namespace-$([guid]::NewGuid().ToString('N'))"
    [System.IO.FileSystemAclExtensions]::Create([System.IO.DirectoryInfo]::new($path), $security)
    return $path
}

function Remove-CloudConnection {
    param([Parameter(Mandatory)] $Connection)

    $null = $script:Connections.Remove($Connection.Id)
    try { $Connection.Public.IsConnected = $false }
    finally {
        if ([System.IO.Directory]::Exists($Connection.Directory)) {
            [System.IO.Directory]::Delete($Connection.Directory, $true)
        }
    }
}

function Get-CloudConnection {
    param([object] $Server)

    if ($null -eq $Server) {
        if ($script:Connections.Count -eq 1) { return @($script:Connections.Values)[0] }
        if ($script:Connections.Count -gt 1) { throw 'CloudAutomation.ConnectionRequired: Select a connection with -Server.' }
        throw 'CloudAutomation.NotConnected: Connect to a server first.'
    }
    $property = $Server.PSObject.Properties['Id']
    if ($null -ne $property -and $property.Value -is [string] -and $script:Connections.ContainsKey($property.Value)) {
        $connection = $script:Connections[$property.Value]
        if ([object]::ReferenceEquals($Server, $connection.Public) -and
            $Server.BrandId -ceq $connection.BrandId -and $Server.Server -ceq $connection.Endpoint -and $Server.IsConnected -eq $true) {
            return $connection
        }
    }
    throw 'CloudAutomation.InvalidConnection: The connection is invalid, expired or belongs to another module.'
}

function Get-CloudObjectConnection {
    param([object] $InputObject, [string] $Type, [object] $Server, [string] $Operation)

    if ($null -eq $InputObject) { throw (New-CloudError -Operation $Operation -ExitCode 5) }
    $key = $InputObject.PSObject.Properties['ConnectionId']
    if ($null -eq $key -or $key.Value -isnot [string] -or -not $script:Connections.ContainsKey($key.Value)) {
        throw (New-CloudError -Operation $Operation -ExitCode 5)
    }
    $connection = $script:Connections[$key.Value]
    $identity = $null
    if (-not $connection.Objects.TryGetValue($InputObject.PSObject, [ref] $identity)) {
        throw (New-CloudError -Operation $Operation -ExitCode 5)
    }
    $originalId = if ($Type -ceq 'VM' -and $identity -is [string]) { $identity }
        elseif ($identity -is [hashtable] -and $identity.Type -ceq $Type) { $identity.Id }
    if (-not $originalId -or $InputObject.Id -cne $originalId -or
        $InputObject.BrandId -cne $connection.BrandId -or $InputObject.Generation -cne $connection.Generation) {
        throw (New-CloudError -Operation $Operation -ExitCode 5)
    }
    if ($null -ne $Server -and (Get-CloudConnection $Server).Id -cne $connection.Id) {
        throw (New-CloudError -Operation $Operation -ExitCode 5 -Target $originalId)
    }
    return $connection
}

function Invoke-CloudConnectedCli {
    param($Connection, [string[]] $Arguments, [double] $TimeoutSeconds = 120)

    $clock = [System.Diagnostics.Stopwatch]::StartNew()
    $result = Invoke-CliProcess -Path $Connection.CliPath -ArgumentList ($Arguments + @('--context', $Connection.ContextName, '--error-format', 'json')) -ConfigPath $Connection.ConfigPath -ApiTimeZone $Connection.ApiTimeZone -TimeoutSeconds $TimeoutSeconds
    $probeTime = [Math]::Min($TimeoutSeconds - $clock.Elapsed.TotalSeconds, 5)
    if ($result.ExitCode -eq 3 -and $probeTime -ge 0.001) {
        # Permission errors and expired sessions share an exit code. Ask the
        # read-only session probe before dropping an otherwise usable session.
        $probe = Invoke-CliProcess -Path $Connection.CliPath -ArgumentList @('doctor', '--context', $Connection.ContextName, '-o', 'json', '--error-format', 'json') -ConfigPath $Connection.ConfigPath -ApiTimeZone $Connection.ApiTimeZone -TimeoutSeconds $probeTime
        $expired = $false
        if ($probe.ExitCode -eq 3) {
            try {
                $report = ConvertFrom-Json -InputObject $probe.Stdout -AsHashtable -ErrorAction Stop
                $expired = @($report['checks'] | Where-Object { $_['name'] -in 'session', 'credentials' -and $_['status'] -eq 'failed' }).Count -gt 0
            } catch { }
        }
        if ($expired) {
            try { Remove-CloudConnection $Connection } catch { Write-Warning 'The session could not be cleaned up locally.' }
        }
    }
    return $result
}
