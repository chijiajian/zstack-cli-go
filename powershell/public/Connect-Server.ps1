function Connect-CloudServer {
    [CmdletBinding(DefaultParameterSetName = 'Account')]
    param(
        [Parameter(Mandatory, Position = 0)] [ValidateNotNullOrEmpty()] [string] $Server,
        [Parameter(Mandatory, ParameterSetName = 'Account')] [pscredential] $Credential,
        [Parameter(Mandatory, ParameterSetName = 'AccessKey')] [ValidateNotNullOrEmpty()] [string] $AccessKeyId,
        [Parameter(Mandatory, ParameterSetName = 'AccessKey')] [ValidateNotNull()] [securestring] $AccessKeySecret,
        [Parameter(ParameterSetName = 'AccessKey')] [ValidateNotNullOrEmpty()] [string] $ApiTimeZone
    )

    $useAccessKey = $PSCmdlet.ParameterSetName -eq 'AccessKey'
    $capability = if ($useAccessKey) { 'accessKeySecretStdin' } else { 'passwordStdin' }
    $cliPath = Get-CloudCliPath $capability
    $directory = New-CloudConnectionDirectory
    $configPath = Join-Path $directory 'config.yaml'
    $id = [guid]::NewGuid().ToString('N')
    $registered = $false
    $bytes = $null
    try {
        $arguments = @('login', $Server, '--no-switch', '--context-name', $id, '--error-format', 'json')
        if ($useAccessKey) {
            $key = [pscredential]::new($AccessKeyId, $AccessKeySecret)
            $bytes = [System.Text.UTF8Encoding]::new($false, $true).GetBytes($key.GetNetworkCredential().Password)
            $arguments += @('--access-key-id', $AccessKeyId, '--access-key-secret-stdin')
        } else {
            $bytes = [System.Text.UTF8Encoding]::new($false, $true).GetBytes($Credential.GetNetworkCredential().Password)
            $arguments += @('--username', $Credential.UserName, '--password-stdin', '--save-password=false')
        }
        $result = Invoke-CliProcess -Path $cliPath -ArgumentList $arguments -ConfigPath $configPath -StdinBytes $bytes
        $failure = Get-CloudProcessError $result 'Connect' -Target $Server
        if ($failure) { $PSCmdlet.ThrowTerminatingError($failure) }
        if (-not [System.IO.File]::Exists($configPath)) {
            $PSCmdlet.ThrowTerminatingError((New-CloudError -Operation Connect -Failure MissingSession -Target $Server))
        }
        if ($useAccessKey) {
            # The CLI saves a key without contacting the server. Verify it with a bounded inventory read.
            $probe = Invoke-CliProcess -Path $cliPath -ArgumentList @('get', 'instances', '--context', $id, '--limit', '1', '--fields', 'uuid', '-o', 'json', '--error-format', 'json') -ConfigPath $configPath -ApiTimeZone $ApiTimeZone -TimeoutSeconds 30
            $null = Convert-CliResult $probe 'Connect' -Shape Array -Target $Server
        }
        $connection = [pscustomobject]@{
            PSTypeName = 'Cloud.Automation.Connection'
            Id = $id
            Server = $Server
            BrandId = [string] $script:Package.brand.id
            IsConnected = $true
        }
        $script:Connections[$id] = @{
            Id = $id; BrandId = $connection.BrandId; Endpoint = $Server; Generation = $id
            Directory = $directory; ConfigPath = $configPath; ContextName = $id; CliPath = $cliPath; ApiTimeZone = $ApiTimeZone
            Public = $connection
            VMEpochs = [System.Collections.Generic.Dictionary[string,int]]::new([StringComparer]::Ordinal)
            Objects = [System.Runtime.CompilerServices.ConditionalWeakTable[object,object]]::new()
        }
        $registered = $true
        return $connection
    } finally {
        if ($null -ne $bytes) { [Array]::Clear($bytes) }
        if (-not $registered) { [System.IO.Directory]::Delete($directory, $true) }
    }
}
