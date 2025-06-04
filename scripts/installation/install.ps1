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

# Import all helper functions
. "$PSScriptRoot\utils.ps1"

# Configuration
$repoOwner = "SameerJS6"
$repoName = "zed-cli-win-unofficial"
$apiUrl = "https://api.github.com/repos/$repoOwner/$repoName/releases/latest"


# Set default install path
$InstallPath = Join-Path $env:LOCALAPPDATA $repoName

# Main Installation Process
Write-Status "Starting installation of $repoName..."

# Check if already installed
if ((Test-Path $InstallPath) -and -not $Force) {
  Write-Status "Already installed at: $InstallPath" "Warning"
  $choice = Read-Host "Continue anyway? (y/n)"
  if ($choice -notmatch '^y(es)?$') {
    Write-Status "Installation cancelled" "Warning"
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

  Write-Status "Verified installation files" "Success"

  # Add to PATH
  Write-Status "Adding to PATH..."
  if (Add-ToPath $InstallPath) {
    Write-Status "Installation completed successfully!" "Success" -SuppressDebug
    Write-Status "⚠️  You may need to restart your terminal to use the commands" "Warning" -SuppressDebug
  }
  else {
    Write-Status "Installation completed but PATH update failed" "Warning"
    Write-Status "Manual PATH setup required: $InstallPath" "Warning"
  }
}
catch {
  Write-Status "Installation failed: $($_.Exception.Message)" "Error"
  exit 1
}
finally {
  # Cleanup
  if (Test-Path $tempDir) {
    Remove-Item $tempDir -Recurse -Force -ErrorAction SilentlyContinue
  }
}

Write-Status "Installation complete! [SUCCESS]" "Success"
