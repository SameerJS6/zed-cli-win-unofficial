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
. "./utils.ps1"

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
  $choice = Read-Host "Continue anyway? (y/N)"
  if ($choice -notmatch '^y(es)?$') {
    Write-Status "Installation cancelled" "Warning"
    exit 0
  }
}

# Create temp directory
# $tempDir = Join-Path $env:TEMP "zed-cli-install-$(Get-Random)"
# New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

$tempDir = New-TempDirectory -Prefix "zed-cli-install"

try {
  # Get latest release info
  $releaseInfo = Get-LatestRelease -ApiUrl $apiUrl -Component "CLI"
  $windowsAsset = Find-WindowsAsset -Assets $releaseInfo.assets -Pattern "x86_64"
  Install-FromZip -DownloadUrl $windowsAsset.browser_download_url -InstallPath $InstallPath -TempDir $tempDir -Component "CLI" -ExtractedFolderPattern $repoName -DeleteZipAfterExtraction 
  # Write-Status "Fetching latest release information..."
  # $releaseInfo = Invoke-RestMethod -Uri $apiUrl -ErrorAction Stop
  # $version = $releaseInfo.tag_name
  # Write-Status "Latest version: $version" "Success"

  # Find Windows zip asset
  # $windowsAsset = $releaseInfo.assets | Where-Object {
  #   $_.name -match "x86_64"
  # } | Select-Object -First 1

  # if (-not $windowsAsset) {
  #   throw "No Windows zip asset found in release"
  # }

  # $downloadUrl = $windowsAsset.browser_download_url
  # $fileName = $windowsAsset.name
  # $downloadPath = Join-Path $tempDir $fileName

  # Write-Status "Downloading: $fileName"
  # Write-Status "From: $downloadUrl"

  # # Download with progress using Get-FileFromWeb
  # Get-FileFromWeb -URL $downloadUrl -File $downloadPath

  # Write-Status "Downloaded: $([math]::Round((Get-Item $downloadPath).Length / 1MB, 2)) MB" "Success"

  # # Create installation directory
  # if (Test-Path $InstallPath) {
  #   Write-Status "Removing existing installation..."
  #   Remove-FromPath $InstallPath
  #   Remove-Item $InstallPath -Recurse -Force
  # }

  # New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
  # Write-Status "Created installation directory: $InstallPath"

  # # Extract zip file
  # Write-Status "Extracting archive..."
  # Expand-Archive -Path $downloadPath -DestinationPath $tempDir -Force

  # # Find the extracted folder
  # $extractedFolder = Get-ChildItem $tempDir -Directory | Where-Object { $_.Name -match $repoName } | Select-Object -First 1

  # if (-not $extractedFolder) {
  #   throw "Could not find extracted folder"
  # }

  # # Copy contents to installation directory
  # Copy-Item "$($extractedFolder.FullName)\*" $InstallPath -Recurse -Force
  # Write-Status "Installed files to: $InstallPath" "Success"

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
    Write-Status "Installation completed successfully!" "Success"
    Write-Status ""
    Write-Status "⚠️  You may need to restart your terminal to use the commands" "Warning"
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

Write-Status "Installation complete! 🎉" "Success"
