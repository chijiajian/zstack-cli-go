function Set-CloudContractDevelopmentCli {
    param($Module, [string] $DevelopmentCliPath)

    if ($IsWindows) { return }
    if (-not $DevelopmentCliPath) { throw 'Supply a local native CLI for non-Windows development verification.' }
    & $Module {
        param($Cli)
        $script:TestCli = $Cli
        $script:Package.cli.sha256 = (Get-FileHash -LiteralPath $Cli -Algorithm SHA256).Hash
        # Preserve package version and capability checks with the local executable.
        function script:Get-CloudPackagedCliPath { return $script:TestCli }
    } (Resolve-Path -LiteralPath $DevelopmentCliPath).Path
}

function New-CloudContractCredential {
    [pscredential]::new('mock', (ConvertTo-SecureString " 测试密码 `"quotes`"`n " -AsPlainText -Force))
}
