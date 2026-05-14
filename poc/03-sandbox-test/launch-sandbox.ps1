# IndraNet PoC 03 - Launch Windows Sandbox with GPU access
# Requires: Windows 10/11 Pro/Enterprise, Virtualization enabled in BIOS

param(
    [string]$SessionId = "poc-test-$(Get-Random -Maximum 9999)"
)

Write-Host "IndraNet PoC 03: Windows Sandbox GPU Test" -ForegroundColor Cyan
Write-Host "Session ID: $SessionId"

# Check if Windows Sandbox is available
$sandboxFeature = Get-WindowsOptionalFeature -Online -FeatureName "Containers-DisposableClientVM"
if ($sandboxFeature.State -ne "Enabled") {
    Write-Error "Windows Sandbox is not enabled. Run: Enable-WindowsOptionalFeature -Online -FeatureName Containers-DisposableClientVM -All"
    exit 1
}

# Create session temp directory
$sessionDir = Join-Path $env:TEMP "IndraNet\$SessionId"
New-Item -ItemType Directory -Path $sessionDir -Force | Out-Null
Write-Host "Session dir: $sessionDir"

# Create a test script that runs inside the sandbox
$testScript = @"
@echo off
echo === IndraNet PoC 03: GPU Test Inside Sandbox ===
echo.

echo Running dxdiag (saving to Desktop)...
dxdiag /t %USERPROFILE%\Desktop\dxdiag-output.txt

echo Waiting for dxdiag to complete...
timeout /t 5 /nobreak > nul

echo.
echo === DXDiag Output ===
type %USERPROFILE%\Desktop\dxdiag-output.txt | findstr /i "Card name\|Display Memory\|Feature Levels\|Driver Version"
echo.

echo Press any key to close sandbox...
pause
"@

$testScriptPath = Join-Path $sessionDir "test.bat"
Set-Content -Path $testScriptPath -Value $testScript

# Create Windows Sandbox configuration
# <VGpu>Enable</VGpu> enables GPU paravirtualization (DirectX sharing)
$wsbConfig = @"
<Configuration>
  <VGpu>Enable</VGpu>
  <Networking>Enable</Networking>
  <MappedFolders>
    <MappedFolder>
      <HostFolder>$sessionDir</HostFolder>
      <SandboxFolder>C:\TestAssets</SandboxFolder>
      <ReadOnly>true</ReadOnly>
    </MappedFolder>
  </MappedFolders>
  <LogonCommand>
    <Command>C:\TestAssets\test.bat</Command>
  </LogonCommand>
</Configuration>
"@

$wsbPath = Join-Path $sessionDir "test.wsb"
Set-Content -Path $wsbPath -Value $wsbConfig

Write-Host "`nLaunching Windows Sandbox with GPU enabled..."
Write-Host "WSB config: $wsbPath"
Write-Host "`nINSTRUCTIONS:"
Write-Host "  1. Wait for sandbox to open and test.bat to run"
Write-Host "  2. Note the GPU name shown in dxdiag output"
Write-Host "  3. Verify DirectX feature level (should be 12_0 or higher)"
Write-Host "  4. Record results in poc/03-sandbox-test/README.md"

$launchTime = Get-Date
Start-Process -FilePath $wsbPath -Wait
$elapsed = (Get-Date) - $launchTime
Write-Host "`nSandbox session duration: $($elapsed.ToString('hh\:mm\:ss'))"

# Cleanup
Remove-Item -Path $sessionDir -Recurse -Force
Write-Host "Cleaned up session directory."
