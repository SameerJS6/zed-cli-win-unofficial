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

# Configuration
$repoOwner = "SameerJS6"
$repoName = "zed-cli-win-unofficial"
$apiUrl = "https://api.github.com/repos/$repoOwner/$repoName/releases/latest"

# Import Get-FileFromWeb function
. "$PSScriptRoot\..\get-file-from-web.ps1"

# Set default install path
$InstallPath = Join-Path $env:LOCALAPPDATA $repoName

# Helper Functions
function Write-Status {
  param([string]$Message, [string]$Type = "Info")
  $color = switch ($Type) {
    "Success" { "Green" }
    "Warning" { "Yellow" }
    "Error" { "Red" }
    default { "Cyan" }
  }
  Write-Host "🔧 $Message" -ForegroundColor $color
}

function Test-Administrator {
  $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
  $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
  return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Add-ToPath {
  param (
    [string]$Directory
  )
    
  $currentUserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    
  if ($currentUserPath -split ';' | Where-Object { $_ -eq $Directory }) {
    Write-Status "Directory already in PATH" "Success"
    return $true
  }
    
  $updatedPath = ""
  if ($currentUserPath) {
    $updatedPath = $currentUserPath + ";" + $Directory
  }
  else {
    $updatedPath = $Directory
  }
    
  try {
    [Environment]::SetEnvironmentVariable("PATH", $updatedPath, "User")
    Write-Status "Added to PATH: $Directory" "Success"
    return $true
  }
  catch {
    Write-Status "Failed to update PATH: $($_.Exception.Message)" "Error"
    return $false
  }
}

function Remove-FromPath {
  param([string]$Directory)
  
  $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
  $pathEntries = $currentPath -split ';' | Where-Object { $_ -ne $Directory -and $_ -ne "" }
  $newPath = $pathEntries -join ';'
  
  try {
    [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
    Write-Status "Removed from PATH: $Directory" "Success"
    return $true
  }
  catch {
    Write-Status "Failed to remove from PATH: $($_.Exception.Message)" "Warning"
    return $false
  }
}

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
$tempDir = Join-Path $env:TEMP "zed-cli-install-$(Get-Random)"
New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

try {
  # Get latest release info
  Write-Status "Fetching latest release information..."
  $releaseInfo = Invoke-RestMethod -Uri $apiUrl -ErrorAction Stop
  $version = $releaseInfo.tag_name
  Write-Status "Latest version: $version" "Success"
    
  # Find Windows zip asset
  $windowsAsset = $releaseInfo.assets | Where-Object { 
    $_.name -match "x86_64"
  } | Select-Object -First 1
    
  if (-not $windowsAsset) {
    throw "No Windows zip asset found in release"
  }
    
  $downloadUrl = $windowsAsset.browser_download_url
  $fileName = $windowsAsset.name
  $downloadPath = Join-Path $tempDir $fileName
    
  Write-Status "Downloading: $fileName"
  Write-Status "From: $downloadUrl"
    
  # Download with progress using Get-FileFromWeb
  Get-FileFromWeb -URL $downloadUrl -File $downloadPath
    
  Write-Status "Downloaded: $([math]::Round((Get-Item $downloadPath).Length / 1MB, 2)) MB" "Success"
    
  # Create installation directory
  if (Test-Path $InstallPath) {
    Write-Status "Removing existing installation..."
    Remove-FromPath $InstallPath
    Remove-Item $InstallPath -Recurse -Force
  }
    
  New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
  Write-Status "Created installation directory: $InstallPath"
    
  # Extract zip file
  Write-Status "Extracting archive..."
  Expand-Archive -Path $downloadPath -DestinationPath $tempDir -Force
    
  # Find the extracted folder
  $extractedFolder = Get-ChildItem $tempDir -Directory | Where-Object { $_.Name -match $repoName } | Select-Object -First 1
    
  if (-not $extractedFolder) {
    throw "Could not find extracted folder"
  }
    
  # Copy contents to installation directory
  Copy-Item "$($extractedFolder.FullName)\*" $InstallPath -Recurse -Force
  Write-Status "Installed files to: $InstallPath" "Success"
    
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
