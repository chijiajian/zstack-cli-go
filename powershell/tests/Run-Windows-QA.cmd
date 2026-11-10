@echo off
setlocal
set "QA_ROOT=%~dp0"
set "QA_PWSH=7.4.20"
if /i "%~1"=="7.6.6" set "QA_PWSH=7.6.6"
if not "%~1"=="" if /i not "%~1"=="7.4.20" if /i not "%~1"=="7.6.6" exit /b 2
set "PATH=%QA_ROOT%runtime\python;%PATH%"
"%QA_ROOT%runtime\pwsh-%QA_PWSH%\pwsh.exe" -NoLogo -NoProfile -NonInteractive -File "%QA_ROOT%powershell\tests\Invoke-ContractPester.ps1" -PesterManifest "%QA_ROOT%runtime\pester\Pester.psd1" -WindowsCliPath "%QA_ROOT%zstack-cli.exe" -ReportPath "%QA_ROOT%contract-report.json" -Benchmark
exit /b %ERRORLEVEL%
