function Invoke-CliProcess {
    param(
        [Parameter(Mandatory)] [string] $Path,
        [Parameter(Mandatory)] [AllowEmptyCollection()] [AllowEmptyString()] [string[]] $ArgumentList,
        [string] $ConfigPath,
        [string] $ApiTimeZone,
        [byte[]] $StdinBytes,
        [ValidateRange(0.001, 3600)] [double] $TimeoutSeconds = 120,
        [System.Threading.CancellationToken] $CancellationToken = [System.Threading.CancellationToken]::None,
        [int] $StdoutLimit = 128MB,
        [int] $StderrLimit = 8MB
    )

    if (-not [System.IO.Path]::IsPathFullyQualified($Path)) { throw 'CLI executable path must be absolute.' }
    if ($ConfigPath -and -not [System.IO.Path]::IsPathFullyQualified($ConfigPath)) { throw 'CLI configuration path must be absolute.' }
    $info = [System.Diagnostics.ProcessStartInfo]::new($Path)
    $info.UseShellExecute = $false
    $info.CreateNoWindow = $true
    $info.RedirectStandardInput = $true
    $info.RedirectStandardOutput = $true
    $info.RedirectStandardError = $true
    foreach ($argument in $ArgumentList) { $info.ArgumentList.Add($argument) }
    foreach ($key in @($info.Environment.Keys)) {
        if ($key.StartsWith('ZSTACK_', [StringComparison]::OrdinalIgnoreCase)) { $null = $info.Environment.Remove($key) }
    }
    if ($ConfigPath) { $info.Environment['ZSTACK_CONFIG'] = $ConfigPath }
    if ($ApiTimeZone) { $info.Environment['ZSTACK_API_TIMEZONE'] = $ApiTimeZone }

    $result = [ordered]@{ ExitCode = $null; Stdout = ''; Stderr = ''; TimedOut = $false; Canceled = $false; Failure = $null }
    $process = [System.Diagnostics.Process]::new()
    $process.StartInfo = $info
    $started = $false
    $clock = [System.Diagnostics.Stopwatch]::StartNew()
    $streams = @()
    try {
        if ($CancellationToken.IsCancellationRequested) {
            $result.Canceled = $true
            return [pscustomobject] $result
        }
        $started = $process.Start()
        if (-not $started) { throw 'CLI process did not start.' }
        $stdin = $process.StandardInput.BaseStream
        $streams = @(
            @{ Name = 'Stdout'; Stream = $process.StandardOutput.BaseStream; Limit = $StdoutLimit; Bytes = [System.IO.MemoryStream]::new(); Buffer = [byte[]]::new(16384); Read = $null; Done = $false }
            @{ Name = 'Stderr'; Stream = $process.StandardError.BaseStream; Limit = $StderrLimit; Bytes = [System.IO.MemoryStream]::new(); Buffer = [byte[]]::new(16384); Read = $null; Done = $false }
        )
        foreach ($stream in $streams) { $stream.Read = $stream.Stream.ReadAsync($stream.Buffer, 0, $stream.Buffer.Length) }
        $inputWrite = $null
        if ($null -ne $StdinBytes -and $StdinBytes.Length -gt 0) {
            $inputWrite = $stdin.WriteAsync($StdinBytes, 0, $StdinBytes.Length)
        } else { $stdin.Dispose() }

        while ($true) {
            if ($CancellationToken.IsCancellationRequested) { $result.Canceled = $true; break }
            if ($clock.Elapsed.TotalSeconds -ge $TimeoutSeconds) { $result.TimedOut = $true; break }
            if ($null -ne $inputWrite -and $inputWrite.IsCompleted) {
                # A child may reject its arguments and close stdin before reading.
                # Its exit code and diagnostic remain authoritative in that case.
                if ($inputWrite.IsFaulted) { $null = $inputWrite.Exception }
                $stdin.Dispose()
                $inputWrite = $null
            }
            $progress = $false
            foreach ($stream in $streams) {
                if ($stream.Done -or -not $stream.Read.IsCompleted) { continue }
                $progress = $true
                $count = $stream.Read.GetAwaiter().GetResult()
                if ($count -eq 0) { $stream.Done = $true; continue }
                if ($stream.Bytes.Length + $count -gt $stream.Limit) {
                    $result.Failure = 'OutputLimit'
                    break
                }
                $stream.Bytes.Write($stream.Buffer, 0, $count)
                $stream.Read = $stream.Stream.ReadAsync($stream.Buffer, 0, $stream.Buffer.Length)
            }
            if ($result.Failure -or ($process.HasExited -and $streams[0].Done -and $streams[1].Done)) { break }
            if (-not $progress) { Start-Sleep -Milliseconds 5 }
        }
        if (-not ($result.Failure -or $result.TimedOut -or $result.Canceled)) {
            $encoding = [System.Text.UTF8Encoding]::new($false, $true)
            foreach ($stream in $streams) {
                $result[$stream.Name] = $encoding.GetString($stream.Bytes.GetBuffer(), 0, [int] $stream.Bytes.Length)
            }
            $result.ExitCode = $process.ExitCode
        }
    } catch [System.Management.Automation.PipelineStoppedException] {
        throw
    } catch {
        $result.Failure = if ($started) { 'ProcessProtocol' } else { 'ProcessStart' }
    } finally {
        if ($started) {
            if (-not $process.HasExited) {
                try { $process.Kill($true) } catch [System.InvalidOperationException] { }
                $null = $process.WaitForExit(2000)
            }
            $process.StandardInput.BaseStream.Dispose()
        }
        foreach ($stream in $streams) { $stream.Bytes.Dispose() }
        $process.Dispose()
    }
    return [pscustomobject] $result
}
