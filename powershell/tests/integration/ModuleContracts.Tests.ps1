param(
    [string] $WindowsCliPath,
    [string] $DevelopmentCliPath,
    [string] $ReportPath,
    [bool] $Benchmark
)

Describe 'PowerShell module and packaged CLI' {
    It 'passes the loopback CLI and PowerShell contract suite' {
        $shell = [System.Diagnostics.Process]::GetCurrentProcess().MainModule.FileName
        $arguments = @('-NoLogo', '-NoProfile', '-NonInteractive', '-File', (Join-Path $PSScriptRoot '../Run-Contracts.ps1'), '-WindowsCliPath', $WindowsCliPath)
        if ($DevelopmentCliPath) { $arguments += @('-DevelopmentCliPath', $DevelopmentCliPath) }
        if ($ReportPath) { $arguments += @('-ReportPath', $ReportPath) }
        if ($Benchmark) { $arguments += '-Benchmark' }
        # Exercise compatibility imports in the same kind of clean session a customer uses.
        $result = & $shell @arguments
        $LASTEXITCODE | Should -Be 0
        $json = $result -join "`n"
        $report = $json.Substring($json.IndexOf("{", [StringComparison]::Ordinal)) | ConvertFrom-Json
        $report.cloud | Should -Be 'loopback mock'
        $report.passed.Count | Should -BeGreaterOrEqual 26
        if ($Benchmark) {
            $benchmarkResult = $report.benchmark
            $benchmarkResult.samples | Should -Be 20
            $benchmarkResult.vmCount | Should -Be 10000
            $benchmarkResult.inventoryRequests | Should -Be 4000
            $benchmarkResult.memoryTargetBytes | Should -Be 512MB
            if ($IsWindows) {
                $benchmarkResult.memoryStatus | Should -Be 'measured'
                # One PowerShell process and at least two CLI calls for each pair and the warmup.
                $benchmarkResult.benchmarkProcessCount | Should -BeGreaterOrEqual (1 + 2 * ($benchmarkResult.samples + 1))
                $benchmarkResult.largestProcessPeakCommitBytes | Should -BeGreaterThan 0
                $benchmarkResult.peakCombinedCommitBytes | Should -BeGreaterOrEqual $benchmarkResult.largestProcessPeakCommitBytes
                $benchmarkResult.memoryTargetMet | Should -Be ($benchmarkResult.peakCombinedCommitBytes -le 512MB)
            } else {
                $benchmarkResult.memoryStatus | Should -Be 'unsupported-platform'
                $benchmarkResult.peakCombinedCommitBytes | Should -BeNullOrEmpty
                $benchmarkResult.memoryTargetMet | Should -BeNullOrEmpty
            }
        }
    }
}
