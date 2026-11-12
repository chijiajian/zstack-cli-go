#Requires -Version 7.4
[CmdletBinding()]
param(
    [Parameter(Mandatory)] [string] $WindowsCliPath,
    [string] $DevelopmentCliPath,
    [string] $PythonPath = (Get-Command python, python3 -CommandType Application -ErrorAction SilentlyContinue | Select-Object -First 1 -ExpandProperty Source),
    [string] $ReportPath,
    [version] $ModuleVersion = '0.1.2',
    [switch] $Benchmark,
    [switch] $ZStackOnly
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
. (Join-Path $PSScriptRoot 'ContractSupport.ps1')
. (Join-Path $PSScriptRoot 'InventoryContracts.ps1')
. (Join-Path $PSScriptRoot 'LifecycleContracts.ps1')
. (Join-Path $PSScriptRoot 'DeviceContracts.ps1')
. (Join-Path $PSScriptRoot 'PlatformContracts.ps1')
. (Join-Path $PSScriptRoot 'MigrationContracts.ps1')
$source = Split-Path -Parent $PSScriptRoot
$work = [System.IO.Directory]::CreateTempSubdirectory('cloud-contracts-').FullName
$service = $null
$module = $null
$acme = $null
$legacy = $null
$connection = $null
$tests = [System.Collections.Generic.List[string]]::new()

function Assert-CloudContract {
    param([bool] $Condition, [string] $Name)
    if (-not $Condition) { throw "Contract check failed: $Name" }
    $tests.Add($Name)
}

function Test-CloudRejected {
    param([scriptblock] $Action, [string] $Pattern)
    try { $null = & $Action } catch {
        if (($_.FullyQualifiedErrorId + ' ' + $_.Exception.Message) -match $Pattern) { return $true }
        throw
    }
    return $false
}

function Wait-CloudFixtureEndpoint {
    param([System.Diagnostics.Process] $Process, [string] $PortFile)

    $clock = [System.Diagnostics.Stopwatch]::StartNew()
    while ($clock.Elapsed.TotalSeconds -lt 120) {
        if ($Process.HasExited) { throw "Loopback fixture exited before it was ready (exit $($Process.ExitCode))." }
        if ([System.IO.File]::Exists($PortFile)) {
            try { $port = [System.IO.File]::ReadAllText($PortFile) }
            catch [System.IO.IOException] { $port = '' }
            if ($port -match '\A[0-9]{1,5}\z' -and [int] $port -gt 0 -and [int] $port -le 65535) {
                return "127.0.0.1:$port"
            }
        }
        Start-Sleep -Milliseconds 50
    }
    throw 'Loopback fixture did not report its port within 120 seconds.'
}

function Get-FreshCloudModuleVersion {
    param([string] $ModuleDirectory, [version] $RequiredVersion)

    $shell = [System.Diagnostics.Process]::GetCurrentProcess().MainModule.FileName
    $modulePath = $ModuleDirectory.Replace("'", "''")
    $command = '$ErrorActionPreference = "Stop"; $env:PSModulePath = ''{0}'' + [IO.Path]::PathSeparator + (Join-Path $PSHOME "Modules"); ' -f $modulePath
    if ($RequiredVersion) {
        $command += '(Import-Module -FullyQualifiedName @{ ModuleName = "ZStack.PowerCLI"; RequiredVersion = "' + $RequiredVersion + '" } -PassThru).Version.ToString()'
    } else {
        $command += '(Import-Module ZStack.PowerCLI -PassThru).Version.ToString()'
    }
    $encoded = [Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($command))
    $version = & $shell -NoLogo -NoProfile -NonInteractive -EncodedCommand $encoded
    if ($LASTEXITCODE -ne 0) { throw 'Fresh PowerShell session failed to import the offline module.' }
    return $version
}

try {
    . (Join-Path $source 'private/process/Invoke-CliProcess.ps1')
    . (Join-Path $source 'private/process/Convert-CliResult.ps1')
    $probeFile = Join-Path $work 'argv probe.py'
    @'
import json, os, sys, time
if sys.argv[1] == 'slow':
    time.sleep(5)
elif sys.argv[1] == 'dual':
    sys.stderr.write('e' * 131072)
    sys.stdout.write('o' * 131072)
else:
    print(json.dumps([sys.argv[1:], sys.stdin.buffer.read().decode('utf8'), os.environ.get('ZSTACK_CONFIG'), os.environ.get('ZSTACK_ACCESS_KEY_SECRET')]))
'@ | Set-Content -LiteralPath $probeFile -Encoding utf8
    $token = 'synthetic-only'
    $configPath = Join-Path $work 'config with spaces.yaml'
    $argsToCheck = @('roundtrip', '', '中文', 'a "quote"', 'spaces path')
    $probe = Invoke-CliProcess -Path $PythonPath -ArgumentList (@($probeFile) + $argsToCheck) -ConfigPath $configPath -StdinBytes ([Text.Encoding]::UTF8.GetBytes($token))
    $parsed = Convert-CliResult $probe Probe -Shape Array
    Assert-CloudContract (($parsed[0] -join '|') -ceq ($argsToCheck -join '|') -and $parsed[1] -ceq $token -and $parsed[2] -ceq $configPath -and $null -eq $parsed[3]) 'argv, private stdin and environment'
    $probe = Invoke-CliProcess -Path $PythonPath -ArgumentList @($probeFile, 'dual')
    Assert-CloudContract ($probe.ExitCode -eq 0 -and $probe.Stdout.Length -eq 131072 -and $probe.Stderr.Length -eq 131072) 'dual output drains without deadlock'
    $probe = Invoke-CliProcess -Path $PythonPath -ArgumentList @($probeFile, 'dual') -StderrLimit 1024
    Assert-CloudContract ($probe.Failure -eq 'OutputLimit') 'bounded process output'
    $probe = Invoke-CliProcess -Path $PythonPath -ArgumentList @($probeFile, 'slow') -TimeoutSeconds 0.25
    Assert-CloudContract ($probe.TimedOut -and (Get-CloudProcessError $probe StartVM -Mutation).FullyQualifiedErrorId -match 'OutcomeUnknown') 'timeout and unknown write outcome'
    $cancellation = [Threading.CancellationTokenSource]::new()
    try {
        $cancellation.CancelAfter(500)
        $probe = Invoke-CliProcess -Path $PythonPath -ArgumentList @($probeFile, 'slow') -CancellationToken $cancellation.Token
        Assert-CloudContract ($probe.Canceled -and -not $probe.TimedOut -and
            (Get-CloudProcessError $probe Query).FullyQualifiedErrorId -match '\.Canceled$' -and
            (Get-CloudProcessError $probe StartVM -Mutation).FullyQualifiedErrorId -match '\.OutcomeUnknown$') 'CLI cancellation distinguishes reads from unknown writes'
    } finally { $cancellation.Dispose() }
    $probe = [pscustomobject]@{ ExitCode = 0; Stdout = 'not json'; Stderr = $token; TimedOut = $false; Canceled = $false; Failure = $null }
    Assert-CloudContract (Test-CloudRejected { Convert-CliResult $probe Query -Shape Array } 'Protocol') 'malformed JSON rejected'

    $builder = Join-Path $source 'packaging/Build-Package.ps1'
    $brands = @((Join-Path $source 'branding/examples/zstack.json'))
    if (-not $ZStackOnly) { $brands += Join-Path $source 'branding/examples/acme.json' }
    $invalidBrand = Get-Content -LiteralPath $brands[0] -Raw | ConvertFrom-Json -AsHashtable
    $invalidBrand.licenseUri = 'https://sample:synthetic@example.invalid/license'
    $invalidBrandPath = Join-Path $work 'invalid-brand.json'
    $invalidBrand | ConvertTo-Json | Set-Content -LiteralPath $invalidBrandPath -Encoding utf8
    Assert-CloudContract (Test-CloudRejected {
        & $builder -BrandConfig $invalidBrandPath -CliPath $WindowsCliPath -ModuleVersion $ModuleVersion -OutputDirectory (Join-Path $work 'invalid packages')
    } 'licenseUri') 'brand license URI rejects user information'
    $rejectedPath = Join-Path $work 'rejected packages'
    $build = @{ BrandConfig = $brands; CliPath = $WindowsCliPath; ModuleVersion = $ModuleVersion; OutputDirectory = $rejectedPath }
    Assert-CloudContract (Test-CloudRejected { & $builder @build -CliVersion 'invalid-fixture-version' } 'CliVersionMismatch') 'packager rejects a mismatched CLI version'
    Assert-CloudContract (Test-CloudRejected { & $builder @build -CliCommit 'invalid-fixture-commit' } 'CliVersionMismatch') 'packager rejects a mismatched CLI commit'
    Assert-CloudContract (-not (Test-Path -LiteralPath $rejectedPath)) 'CLI mismatch leaves no package output'
    $build.OutputDirectory = Join-Path $work 'packages'
    $packages = @(& $builder @build)
    $zstackData = Import-PowerShellDataFile -LiteralPath (Join-Path $packages[0].Directory 'ZStack.PowerCLI.psd1')
    $zstackPackage = Get-Content -LiteralPath (Join-Path $packages[0].Directory 'package.json') -Raw | ConvertFrom-Json -AsHashtable
    Assert-CloudContract ($zstackData.ModuleVersion -eq $ModuleVersion.ToString() -and $zstackData.Author -eq 'ZStack' -and
        $zstackData.Description -eq 'ZStack PowerCLI' -and
        $zstackData.PrivateData.PSData.LicenseUri -eq 'https://www.apache.org/licenses/LICENSE-2.0' -and
        $zstackData.PrivateData.PSData.ProjectUri -eq 'https://github.com/terraform-zstack-modules/zstack-cli-go' -and
        $zstackPackage.brand.helpUri -eq 'https://github.com/terraform-zstack-modules/zstack-cli-go/blob/main/docs/powershell.md' -and
        @($zstackData.PrivateData.PSData.Tags).Count -eq 3) 'Gallery metadata in ZStack manifest'
    $compatibilityData = Import-PowerShellDataFile -LiteralPath (Join-Path $packages[0].Directory 'Compatibility/ZStack.PowerCLI.Compatibility.psd1')
    Assert-CloudContract ($compatibilityData.PrivateData.PSData.LicenseUri -eq $zstackData.PrivateData.PSData.LicenseUri -and
        $compatibilityData.PrivateData.PSData.ProjectUri -eq $zstackData.PrivateData.PSData.ProjectUri) 'bundled compatibility manifest metadata'
    $null = Test-ModuleManifest -Path (Join-Path $packages[0].Directory 'ZStack.PowerCLI.psd1') -ErrorAction Stop
    Assert-CloudContract $true 'generated ZStack manifest validates'
    if (-not $ZStackOnly) {
        Assert-CloudContract ($packages.Count -eq 2) 'two OEM packages'
        $acmeData = Import-PowerShellDataFile -LiteralPath (Join-Path $packages[1].Directory 'Acme.Cloud.psd1')
        Assert-CloudContract ($acmeData.PrivateData.PSData.LicenseUri -eq 'https://www.apache.org/licenses/LICENSE-2.0' -and
            -not $acmeData.PrivateData.PSData.Contains('ProjectUri')) 'optional OEM project URI omitted'
        foreach ($file in 'private/process/Invoke-CliProcess.ps1', 'private/objects/Convert-CloudVM.ps1', 'bin/windows-amd64/zstack-cli.exe') {
            $a = (Get-FileHash -LiteralPath (Join-Path $packages[0].Directory $file)).Hash
            $b = (Get-FileHash -LiteralPath (Join-Path $packages[1].Directory $file)).Hash
            Assert-CloudContract ($a -eq $b) "OEM common payload $file"
        }
    }
    if (-not $IsWindows -and -not $DevelopmentCliPath) { throw 'Supply a local native CLI for non-Windows development verification.' }
    $installer = Join-Path $source 'packaging/Install-Offline.ps1'
    $installRoot = Join-Path $work 'offline modules'
    Assert-CloudContract (Test-CloudRejected { & $installer -Archive $packages[0].Archive -SHA256 ('0' * 64) -ModuleDirectory $installRoot } 'checksum mismatch') 'bad offline checksum rejected'
    $manifest = & $installer -Archive $packages[0].Archive -SHA256 $packages[0].SHA256 -ModuleDirectory $installRoot
    Assert-CloudContract (Test-CloudRejected { & $installer -Archive $packages[0].Archive -SHA256 $packages[0].SHA256 -ModuleDirectory $installRoot } 'already installed') 'installed version not overwritten'
    $module = Import-Module -Name $manifest -PassThru
    Assert-CloudContract ($module.Version -eq $ModuleVersion) 'installed module matches the requested test version'
    Assert-CloudContract ($module.ExportedFunctions.Count -eq 57 -and -not $module.ExportedFunctions.ContainsKey('Get-VM')) 'ZStack native export scope'
    $helpCases = @(@{ Command = 'Get-ZStackVM'; Name = 'ZStack PowerCLI'; Namespace = 'zstack-powershell'; Uri = 'https://github.com/terraform-zstack-modules/zstack-cli-go/blob/main/docs/powershell.md' })
    if (-not $ZStackOnly) {
        $acme = Import-Module -Name (Join-Path $packages[1].Directory 'Acme.Cloud.psd1') -PassThru
        Assert-CloudContract ($acme.ExportedFunctions.Count -eq 57 -and $acme.ExportedFunctions.ContainsKey('New-AcmeVM') -and $acme.ExportedFunctions.ContainsKey('Get-AcmeHardDisk') -and $acme.ExportedFunctions.ContainsKey('Set-AcmeCDDrive')) 'OEM export scopes'
        $helpCases += @{ Command = 'Get-AcmeVM'; Name = 'Acme Cloud Automation'; Namespace = 'acme-powershell'; Uri = 'https://example.invalid/acme/powershell' }
        Set-CloudContractDevelopmentCli $acme $DevelopmentCliPath
    }
    foreach ($case in $helpCases) {
        $help = Get-Help -Name $case.Command -Full
        $description = @($help.Description.Text) -join ' '
        Assert-CloudContract ($description.Contains($case.Name) -and $description.Contains($case.Namespace) -and
            $description.Contains(" $ModuleVersion) uses CLI ") -and $help.RelatedLinks.NavigationLink.Uri -ceq $case.Uri) "OEM help identifies $($case.Command) and its package"
    }
    Set-CloudContractDevelopmentCli $module $DevelopmentCliPath
    $preflight = & $module {
        $expectedHash = $script:Package.cli.sha256
        $expectedVersion = $script:Package.cli.version
        try {
            $script:Package.cli.sha256 = '0' * 64
            try { $null = Get-CloudCliPath 'passwordStdin'; $hashRejected = $false }
            catch { $hashRejected = $_.Exception.Message -match 'CliIntegrity' }
            $script:Package.cli.sha256 = $expectedHash
            $script:Package.cli.version = 'wrong-version'
            try { $null = Get-CloudCliPath 'passwordStdin'; $versionRejected = $false }
            catch { $versionRejected = $_.Exception.Message -match 'CliVersionMismatch' }
            $script:Package.cli.version = $expectedVersion
            $first = Get-CloudCliPath 'passwordStdin'
            $second = Get-CloudCliPath 'vmMemoryBytes'
            @{
                HashRejected = $hashRejected
                VersionRejected = $versionRejected
                ReusedPath = $first -ceq $second -and $script:VerifiedCapabilities.ContainsKey('vmMemoryBytes')
            }
        } finally {
            $script:Package.cli.sha256 = $expectedHash
            $script:Package.cli.version = $expectedVersion
        }
    }
    Assert-CloudContract $preflight.HashRejected 'mismatched packaged CLI hash rejected'
    Assert-CloudContract $preflight.VersionRejected 'mismatched packaged CLI version rejected'
    Assert-CloudContract $preflight.ReusedPath 'new capability checks reuse verified CLI path'

    $ready = Join-Path $work 'mock-port'
    $info = [System.Diagnostics.ProcessStartInfo]::new($PythonPath)
    $info.UseShellExecute = $false
    $info.ArgumentList.Add((Join-Path $PSScriptRoot 'mock_api.py'))
    $info.ArgumentList.Add($ready)
    $service = [System.Diagnostics.Process]::Start($info)
    $endpoint = Wait-CloudFixtureEndpoint $service $ready
    $credential = New-CloudContractCredential
    $connection = Connect-ZStackServer -Server $endpoint -Credential $credential
    $badKey = ConvertTo-SecureString 'synthetic invalid secret' -AsPlainText -Force
    Assert-CloudContract (Test-CloudRejected { Connect-ZStackServer $endpoint -AccessKeyId 'synthetic-access-key' -AccessKeySecret $badKey -ApiTimeZone Asia/Shanghai } 'Connect.Authentication') 'invalid access key is rejected before registering a connection'
    $keySecret = ConvertTo-SecureString ' synthetic signing secret "quotes" ' -AsPlainText -Force
    $keyConnection = Connect-ZStackServer $endpoint -AccessKeyId 'synthetic-access-key' -AccessKeySecret $keySecret -ApiTimeZone Asia/Shanghai
    $keyDirectory = & $module { param($Server) (Get-CloudConnection $Server).Directory } $keyConnection
    try {
        Assert-CloudContract ((Get-ZStackVM -Id vm-b -Server $keyConnection).Id -eq 'vm-b') 'access key connect and inventory sign requests in the selected time zone'
    } finally { Disconnect-ZStackServer -Server $keyConnection }
    Assert-CloudContract (-not $keyConnection.IsConnected -and -not (Test-Path -LiteralPath $keyDirectory)) 'access key disconnect removes its private configuration'
    $sessionDir = & $module { @($script:Connections.Values)[0].Directory }
    if ($IsWindows) {
        $acl = Get-Acl -LiteralPath $sessionDir
        $userSid = [System.Security.Principal.WindowsIdentity]::GetCurrent().User.Value
        $rules = @($acl.GetAccessRules($true, $true, [System.Security.Principal.SecurityIdentifier]))
        Assert-CloudContract ($acl.AreAccessRulesProtected -and $rules.Count -eq 1 -and $rules[0].IdentityReference.Value -eq $userSid) 'Windows private session directory ACL'
    }
    $vms = @(Get-ZStackVM -Server $connection)
    Assert-CloudContract ($vms.Count -eq 3 -and (Get-ZStackVM '开发*').Id -eq 'vm-a') 'inventory, wildcard and Chinese arguments'
    Assert-CloudContract ((Get-ZStackVM -Name 'WEB-PROD').Id -eq 'vm-b') 'default VM name query stays case-insensitive'
    $namedVM = Get-ZStackVM -Name '开发 VM' -Exact -Server $connection
    $request = Invoke-RestMethod "http://$endpoint/test/reads"
    Assert-CloudContract ($namedVM.Id -ceq 'vm-a' -and $request.lastInventoryCondition -ceq 'name=开发 VM') 'exact VM name filters on the server'
    Assert-CloudContract (@(Get-ZStackVM -Name 'missing' -Exact -Server $connection).Count -eq 0) 'missing literal VM name emits no object'
    Assert-CloudContract (@(Get-ZStackVM -Name @('web-prod', '开发 VM') -Exact -Server $connection).Count -eq 2) 'multiple literal VM names query without scanning inventory'
    Assert-CloudContract (Test-CloudRejected { Get-ZStackVM -Name 'web*' -Exact -Server $connection } 'Specify one or more literal names') 'exact VM name rejects wildcard patterns'
    $null = Get-ZStackVM -Id vm-a
    $request = Invoke-RestMethod "http://$endpoint/test/reads"
    Assert-CloudContract ($request.lastInventoryCondition -ceq 'uuid=vm-a') 'single VM lookup filters on the server'
    Assert-CloudContract (Test-CloudRejected { Get-ZStackVM -Id wrong-identity } 'ResolveVM.Protocol') 'single VM lookup rejects a mismatched UUID'
    Assert-CloudContract ((Get-ZStackVM -Id vm-a).MemoryGB -eq 1.5 -and $null -eq (Get-ZStackVM -Id vm-c).MemoryGB) 'numeric memory and omitted memory'
    Assert-CloudContract (@(Get-ZStackVM -Name 'missing*').Count -eq 0) 'empty query pipeline'
    $csv = @($vms | Where-Object PowerState -eq 'PoweredOn' | Select-Object Name, NumCpu, MemoryGB | ConvertTo-Csv -NoTypeInformation)
    Assert-CloudContract ($csv.Count -eq 2 -and $csv[1] -match 'web-prod') 'existing CSV report pipeline'
    Assert-CloudInventoryContracts $connection $endpoint $vms $credential $module -ZStackOnly:$ZStackOnly

    $oldBinary = Join-Path (Split-Path -Parent $manifest) 'bin/windows-amd64/zstack-cli.exe'
    $oldHash = (Get-FileHash -LiteralPath $oldBinary -Algorithm SHA256).Hash
    $boundCli = & $module { param($Connection) (Get-CloudConnection $Connection).CliPath } $connection
    $upgradeVersion = [version]::new($ModuleVersion.Major, $ModuleVersion.Minor, ($ModuleVersion.Build + 1))
    $upgrade = & $builder -BrandConfig $brands[0] -CliPath $WindowsCliPath -ModuleVersion $upgradeVersion -OutputDirectory (Join-Path $work 'upgrade packages')
    $incomplete = Join-Path $work 'incomplete.zip'
    Remove-Item -LiteralPath (Join-Path $upgrade.Directory 'public/Get-VM.ps1')
    [System.IO.Compression.ZipFile]::CreateFromDirectory((Split-Path -Parent $upgrade.Directory), $incomplete, [System.IO.Compression.CompressionLevel]::Optimal, $true)
    $incompleteHash = (Get-FileHash -LiteralPath $incomplete -Algorithm SHA256).Hash
    Assert-CloudContract (Test-CloudRejected { & $installer -Archive $incomplete -SHA256 $incompleteHash -ModuleDirectory $installRoot } 'Incomplete module package') 'incomplete newer package rejected despite matching archive checksum'
    Assert-CloudContract (-not (Test-Path -LiteralPath (Join-Path $installRoot "ZStack.PowerCLI/$upgradeVersion")) -and
        @(Get-ChildItem -LiteralPath $installRoot -Force -Directory -Filter '.cloud-install-*').Count -eq 0) 'failed install leaves no discoverable upgrade or staging files'
    $interrupted = Join-Path $installRoot '.cloud-install-interrupted/payload'
    [System.IO.Compression.ZipFile]::ExtractToDirectory($upgrade.Archive, $interrupted)
    Assert-CloudContract ((Get-FreshCloudModuleVersion $installRoot) -ceq $ModuleVersion.ToString()) 'fresh session ignores interrupted staging'
    $upgradeManifest = & $installer -Archive $upgrade.Archive -SHA256 $upgrade.SHA256 -ModuleDirectory $installRoot
    Assert-CloudContract ((Get-FreshCloudModuleVersion $installRoot) -ceq $upgradeVersion.ToString() -and
        (Test-Path -LiteralPath $upgradeManifest -PathType Leaf)) 'fresh session imports the newer installed module'
    $activeCli = & $module { param($Connection) (Get-CloudConnection $Connection).CliPath } $connection
    Assert-CloudContract ($module.Version -eq $ModuleVersion -and $activeCli -ceq $boundCli -and
        (-not $IsWindows -or $activeCli -ceq $oldBinary) -and
        (Get-FileHash -LiteralPath $oldBinary -Algorithm SHA256).Hash -eq $oldHash -and
        (Get-ZStackVM -Id vm-a -Server $connection).Id -eq 'vm-a') 'active connection keeps its original CLI after upgrade'
    Assert-CloudContract ((Get-FreshCloudModuleVersion $installRoot -RequiredVersion $ModuleVersion) -ceq $ModuleVersion.ToString()) 'fresh session explicitly rolls back to the intact older module'

    $copy = $vms[0] | ConvertTo-Json | ConvertFrom-Json
    Assert-CloudContract (Test-CloudRejected { $copy | Start-ZStackVM -ErrorAction Stop } 'ResolveVM.Conflict') 'copied VM rejected before mutation'
    $other = Connect-ZStackServer -Server $endpoint -Credential $credential
    try {
        Assert-CloudContract (Test-CloudRejected { Get-ZStackVM } 'ConnectionRequired') 'multi-connection requires explicit selection'
        Assert-CloudContract (Test-CloudRejected { $vms[0] | Start-ZStackVM -Server $other -ErrorAction Stop } 'ResolveVM.Conflict') 'cross-connection write rejected'
        if (-not $ZStackOnly) {
            Assert-CloudContract (Test-CloudRejected { Get-AcmeVM -Server $connection } 'InvalidConnection') 'cross-OEM connection rejected'
        }
    } finally { Disconnect-ZStackServer -Server $other }
    $null = (Get-ZStackVM -Id vm-a) | Start-ZStackVM -WhatIf -ErrorAction Stop
    Assert-CloudContract ((Get-ZStackVM -Id vm-a).RawState -eq 'Stopped') 'WhatIf sends no write'
    $start = & $module {
        $invoke = ${function:Invoke-CloudConnectedCli}
        $observed = @{}
        function Invoke-CloudConnectedCli {
            param($Connection, [string[]] $Arguments, [double] $TimeoutSeconds)
            if ($Arguments[0] -eq 'instance' -and $Arguments[1] -eq 'start') {
                $observed.StartBudget = $TimeoutSeconds
                $index = [Array]::IndexOf($Arguments, '--request-timeout')
                $observed.RequestTimeout = if ($index -ge 0) { $Arguments[$index + 1] }
            } elseif ($Arguments[0] -eq 'wait') {
                $observed.WaitBudget = $TimeoutSeconds
            }
            & $invoke @PSBoundParameters
        }
        $observed.VM = (Get-ZStackVM -Id vm-a) | Start-ZStackVM -Confirm:$false -ErrorAction Stop
        $observed
    }
    $expectedTimeout = [long] [Math]::Floor($start.StartBudget * 1000)
    Assert-CloudContract ($start.StartBudget -gt 120 -and $start.StartBudget -le 300 -and
        $start.RequestTimeout -ceq "${expectedTimeout}ms" -and $start.WaitBudget -gt 0 -and
        $start.WaitBudget -lt $start.StartBudget) 'async start passes remaining request timeout and preserves the wait deadline'
    $mutation = Invoke-RestMethod "http://$endpoint/test/reads"
    Assert-CloudContract ($mutation.startRequests -eq 1 -and $mutation.startJobPolls -eq 2) 'async start polls one accepted request until completion'
    $started = $start.VM
    Assert-CloudContract ($started.PowerState -eq 'PoweredOn' -and ($started | Wait-ZStackVM).MemoryGB -eq 1.5) 'start and wait refresh VM projection'
    $results = @(Start-ZStackVM -Id 'missing-vm', 'vm-a' -ErrorAction Continue -Confirm:$false 2>&1)
    $errors = @($results | Where-Object { $_ -is [System.Management.Automation.ErrorRecord] })
    $continued = @($results | Where-Object { $_ -isnot [System.Management.Automation.ErrorRecord] })
    Assert-CloudContract ($continued.Count -eq 1 -and $continued[0].Id -eq 'vm-a' -and
        $errors.Count -eq 1 -and $errors[0].FullyQualifiedErrorId -match 'NotFound') 'nonterminating start error preserves the next target'
    $stopped = Test-CloudRejected { Start-ZStackVM -Id 'missing-vm', 'vm-a' -ErrorAction Stop -Confirm:$false } 'NotFound'
    $afterError = Invoke-RestMethod "http://$endpoint/test/reads"
    Assert-CloudContract ($stopped -and $afterError.lastInventoryCondition -ceq 'uuid=missing-vm' -and
        $afterError.startRequests -eq 1) 'ErrorAction Stop terminates before the next target'
    $legacyManifest = Join-Path (Split-Path -Parent $manifest) 'Compatibility/ZStack.PowerCLI.Compatibility.psd1'
    function Get-VM { 'legacy collision' }
    Assert-CloudContract (Test-CloudRejected { Import-Module -Name $legacyManifest } 'CompatibilityConflict') 'legacy name collision rejected'
    Assert-CloudContract ((Get-VM) -eq 'legacy collision') 'collision leaves existing command intact'
    Remove-Item Function:Get-VM
    function Get-HardDisk { 'disk collision' }
    Assert-CloudContract ((Test-CloudRejected { Import-Module -Name $legacyManifest } 'CompatibilityConflict') -and
        (Get-HardDisk) -ceq 'disk collision') 'new inventory name collision leaves existing command intact'
    Remove-Item Function:Get-HardDisk
    $ambientRoot = Join-Path $work 'ambient modules'
    $ambientModule = Join-Path $ambientRoot 'CloudContractAmbientVM'
    $null = New-Item -ItemType Directory -Path $ambientModule -Force
    Set-Content -LiteralPath (Join-Path $ambientModule 'CloudContractAmbientVM.psm1') -Value "function Get-VM { throw 'Unexpected ambient Get-VM invocation.' }"
    New-ModuleManifest -Path (Join-Path $ambientModule 'CloudContractAmbientVM.psd1') `
        -RootModule 'CloudContractAmbientVM.psm1' -FunctionsToExport 'Get-VM'
    $savedModulePath = $env:PSModulePath
    try {
        $env:PSModulePath = $ambientRoot + [IO.Path]::PathSeparator + $savedModulePath
        if (-not (Get-Module -Name CloudContractAmbientVM -ListAvailable)) { throw 'Ambient VM module fixture was not discoverable.' }
        $legacy = Import-Module -Name $legacyManifest -PassThru
        Assert-CloudContract (-not (Get-Module -Name CloudContractAmbientVM)) 'legacy import does not autoload an installed VM module'
    } finally {
        $env:PSModulePath = $savedModulePath
        Remove-Module CloudContractAmbientVM -ErrorAction SilentlyContinue
    }
    Assert-CloudContract ((Get-VM -Id vm-a -Server $connection).PowerState -eq 'PoweredOn') 'opt-in legacy Get-VM reuses core'
    Assert-CloudContract (Test-CloudRejected { Start-VM -VM $started -RunAsync } 'NamedParameterNotFound') 'unsupported legacy parameter rejected'
    Assert-CloudLifecycleContracts $connection $endpoint $credential
    Assert-CloudDeviceContracts $connection $endpoint $credential
    Assert-CloudPlatformContracts $connection $endpoint $credential
    Assert-CloudMigrationContracts $connection $endpoint $credential $legacy
    Disconnect-ZStackServer -Server $connection
    $connection = $null
    Assert-CloudContract (-not (Test-Path -LiteralPath $sessionDir)) 'session cleanup'
    Assert-CloudContract (Test-CloudRejected { $started | Start-ZStackVM -ErrorAction Stop } 'ResolveVM.Conflict') 'disconnected VM cannot write'
    Assert-CloudMigrationAssessmentContracts (Join-Path $source 'migration/Assess-Migration.ps1') $work ($tests.ToArray())
    $report = [ordered]@{
        powerShell = $PSVersionTable.PSVersion.ToString()
        os = [System.Runtime.InteropServices.RuntimeInformation]::OSDescription
        architecture = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString()
        developmentCli = [bool] $DevelopmentCliPath
        moduleVersion = $module.Version.ToString()
        windowsCliSha256 = (Get-FileHash -LiteralPath $WindowsCliPath -Algorithm SHA256).Hash
        cli = (Get-Content -LiteralPath (Join-Path $packages[0].Directory 'package.json') -Raw | ConvertFrom-Json -AsHashtable).cli
        packages = @($packages | Select-Object ModuleName, Version, SHA256)
        cloud = 'loopback mock'
        zstackOnly = [bool] $ZStackOnly
        passed = $tests.ToArray()
    }
    if ($Benchmark) {
        $benchPort = Join-Path $work 'bench-port'
        $info.ArgumentList.Clear()
        $info.ArgumentList.Add((Join-Path $PSScriptRoot 'mock_api.py'))
        $info.ArgumentList.Add($benchPort)
        $info.ArgumentList.Add('10000')
        $service.Kill($true)
        $service.WaitForExit()
        $service.Dispose()
        $service = [System.Diagnostics.Process]::Start($info)
        $benchEndpoint = Wait-CloudFixtureEndpoint $service $benchPort
        $benchmarkPath = Join-Path $work 'query benchmark.json'
        $worker = Join-Path $PSScriptRoot 'Measure-ModuleQuery.ps1'
        $parameters = [ordered]@{ ManifestPath = $manifest; Endpoint = $benchEndpoint; ReportPath = $benchmarkPath }
        if ($DevelopmentCliPath) { $parameters.DevelopmentCliPath = (Resolve-Path -LiteralPath $DevelopmentCliPath).Path }
        $command = "& '" + $worker.Replace("'", "''") + "'"
        foreach ($entry in $parameters.GetEnumerator()) {
            $command += " -$($entry.Key) '" + $entry.Value.Replace("'", "''") + "'"
        }
        $encoded = [Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($command + '; exit $LASTEXITCODE'))
        $shell = [System.Diagnostics.Process]::GetCurrentProcess().MainModule.FileName
        $peak = $largest = $processCount = $null
        if ($IsWindows) {
            Add-Type -Path (Join-Path $PSScriptRoot 'WindowsBenchmarkJob.cs')
            $job = [CloudContractTests.WindowsBenchmarkJob]::Run($shell, $encoded, 600000)
            $exitCode = $job.ExitCode
            $peak = $job.PeakCommitBytes
            $largest = $job.LargestProcessPeakCommitBytes
            $processCount = $job.ProcessCount
        } else {
            & $shell -NoLogo -NoProfile -NonInteractive -EncodedCommand $encoded
            $exitCode = $LASTEXITCODE
        }
        $benchmarkResult = if (Test-Path -LiteralPath $benchmarkPath) {
            Get-Content -LiteralPath $benchmarkPath -Raw | ConvertFrom-Json -AsHashtable
        }
        if ($exitCode -ne 0 -or $null -eq $benchmarkResult) {
            $reason = if ($benchmarkResult) { $benchmarkResult['failure'] }
            throw "Query benchmark worker failed (exit $exitCode). $reason"
        }
        $benchmarkResult.peakCombinedCommitBytes = $peak
        $benchmarkResult.largestProcessPeakCommitBytes = $largest
        $benchmarkResult.benchmarkProcessCount = $processCount
        $benchmarkResult.memoryTargetBytes = 512MB
        $benchmarkResult.memoryTargetMet = if ($peak -gt 0) { $peak -le 512MB }
        $benchmarkResult.memoryStatus = if ($IsWindows) { 'measured' } else { 'unsupported-platform' }
        $benchmarkResult.memoryScope = 'Windows Job Object peak commit for the benchmark PowerShell process and its CLI children'
        $report.benchmark = $benchmarkResult
    }
    if ($ReportPath) { $report | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $ReportPath -Encoding utf8 }
    $report | ConvertTo-Json -Depth 5
} finally {
    if ($null -ne $connection) { Disconnect-ZStackServer -Server $connection -ErrorAction SilentlyContinue }
    foreach ($item in @($legacy, $acme, $module)) { if ($item) { Remove-Module -ModuleInfo $item -ErrorAction SilentlyContinue } }
    if ($null -ne $service) {
        if (-not $service.HasExited) { $service.Kill($true); $service.WaitForExit() }
        $service.Dispose()
    }
    Remove-Item -LiteralPath $work -Recurse -Force
}
