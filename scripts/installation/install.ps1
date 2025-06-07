#Requires -Version 5.0

<#
.SYNOPSIS
    Install zed-cli-win-unofficial from GitHub releases
.DESCRIPTION
    Downloads the latest release of zed-cli-win-unofficial, extracts it to a safe location,
    and adds it to the user's PATH environment variable.
.PARAMETER Force
    Force reinstallation even if already installed
#>

param(
  [switch]$Force
)

$Global:ScriptDebugMode = $false

# Import all helper functions
. "./utils.ps1"

# Configuration
$repoOwner = "SameerJS6"
$repoName = "zed-cli-win-unofficial"
$apiUrl = "https://api.github.com/repos/$repoOwner/$repoName/releases/latest"


# Set default install path
$InstallPath = Join-Path $env:LOCALAPPDATA $repoName

# Main Installation Process
Write-Info "Starting installation of $repoName..."

# Check if already installed
if ((Test-Path $InstallPath) -and -not $Force) {
  Write-Warning "Already installed at: $InstallPath"
  $choice = Read-Host "Continue anyway? (y/n)"
  if ($choice -notmatch '^y(es)?$') {
    Write-Warning "Installation cancelled"
    exit 0
  }
}

$tempDir = New-TempDirectory -Prefix "zed-cli-install"

try {
  # Get latest release info
  $releaseInfo = Get-LatestRelease -ApiUrl $apiUrl -Component "CLI"
  $windowsAsset = Find-WindowsAsset -Assets $releaseInfo.assets -Pattern "x86_64"
  Install-FromZip -DownloadUrl $windowsAsset.browser_download_url -InstallPath $InstallPath -TempDir $tempDir -Component "CLI" -ExtractedFolderPattern $repoName -DeleteZipAfterExtraction 
  
  # Verify installation
  $exePath = Join-Path $InstallPath "$repoName.exe"
  $batPath = Join-Path $InstallPath "zed.bat"

  if (-not (Test-Path $exePath)) {
    throw "Main executable not found: $exePath"
  }

  if (-not (Test-Path $batPath)) {
    throw "Batch wrapper not found: $batPath"
  }

  Write-Success "Verified installation files"
  Write-Debug "Testing Debug Logs in Debug Mode working or not!!!"

  # Add to PATH
  Write-Info "Adding to PATH..."
  if (Add-ToPath $InstallPath) {
    Write-Success "Installation completed successfully!"
    Write-Warning "You may need to restart your terminal to use the commands"
    Send-AnalyticsEvent -EventType "zed_cli_installation_completed"
  }
  else {
    Send-AnalyticsEvent -EventType "zed_cli_installation_completed_with_path_update_failed"
    Write-Warning "Installation completed but PATH update failed"
    Write-Warning "Manual PATH setup required: $InstallPath"
  }
}
catch {
  Write-Error "Installation failed: $($_.Exception.Message)"
  Send-AnalyticsEvent -EventType "zed_cli_installation_failed"
  exit 1
}
finally {
  # Cleanup
  if (Test-Path $tempDir) {
    Remove-Item $tempDir -Recurse -Force -ErrorAction SilentlyContinue
  }
}
