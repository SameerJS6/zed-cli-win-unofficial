<#
.SYNOPSIS
    Install Zed editor and zed-cli-win-unofficial together
.DESCRIPTION
    Downloads and installs the Unofficial Zed editor (Windows build) & Zed CLI, then configures szed-cli-win-unofficial to use the Zed installation.
.PARAMETER ZedInstallPath
    Custom installation directory for Zed editor (default: %LOCALAPPDATA%\Zed)
.PARAMETER CliInstallPath
    Custom installation directory for zed-cli-win-unofficial (default: %LOCALAPPDATA%\zed-cli-win-unofficial)
.PARAMETER Force
    Force reinstallation even if already installed
#>

param(
  [string]$ZedInstallPath = "",
  [string]$CliInstallPath = "",
  [switch]$Force
)

$Global:ScriptDebugMode = $false

# Import all helper functions
. "./utils.ps1"

# Configuration
$zedRepoOwner = "pirafrank"
$zedRepoName = "zed_unofficial_win_builds"
$zedApiUrl = "https://api.github.com/repos/$zedRepoOwner/$zedRepoName/releases/latest"

$cliRepoOwner = "SameerJS6"
$cliRepoName = "zed-cli-win-unofficial"
$cliApiUrl = "https://api.github.com/repos/$cliRepoOwner/$cliRepoName/releases/latest"


# Set default install paths if not provided
if (-not $ZedInstallPath) {
  $ZedInstallPath = Join-Path $env:LOCALAPPDATA "Programs\ZedTesting" # TODO: Change the end path to `Zed` from `ZedTesting`
}
if (-not $CliInstallPath) {
  $CliInstallPath = Join-Path $env:LOCALAPPDATA $cliRepoName
}

function Install-ZedEditor {
  Write-Info "[Zed] Starting Zed editor installation..."

  # Check if already installed
  if ((Test-Path $ZedInstallPath) -and -not $Force) {
    $zedExe = Join-Path $ZedInstallPath "zed.exe"
    if (Test-Path $zedExe) {
      Write-Success "[Zed] Zed editor already installed at: $ZedInstallPath"
      return $zedExe
    }
  }

 
  $tempDir = New-TempDirectory -Prefix "zed-install"

  try {
    $releaseInfo = Get-LatestRelease -ApiUrl $zedApiUrl -Component "Zed"
    $windowsAsset = Find-WindowsAsset -Assets $releaseInfo.assets -Pattern ".zip"
    Install-FromZip -DownloadUrl $windowsAsset.browser_download_url -InstallPath $ZedInstallPath -TempDir $tempDir -Component "Zed" -DeleteZipAfterExtraction | Out-Null
    
    # Verify installation
    $zedExePath = Join-Path $ZedInstallPath "zed.exe"
    if (-not (Test-Path $zedExePath)) {
      throw "Zed executable not found after installation: $zedExePath"
    }

    Write-Success "[Zed] Verified Zed installation"

    return $zedExePath
  }
  catch {
    Write-Error "[Zed] Zed installation failed: $($_.Exception.Message)"
    throw
  }
  finally {
    # Cleanup
    if (Test-Path $tempDir) {
      Remove-Item $tempDir -Recurse -Force -ErrorAction SilentlyContinue
      Write-Debug "[Zed] Cleaned up Zed temp files"
    }
  }
}

function Install-ZedCli {
  Write-Info "[CLI] Starting zed-cli-win-unofficial installation..."

  # Check if already installed
  if ((Test-Path $CliInstallPath) -and -not $Force) {
    $cliExe = Join-Path $CliInstallPath "$cliRepoName.exe"
    Write-Debug "[Debug] Existing CLI exe path = '$cliExe'"
    if (Test-Path $cliExe) {
      Write-Success "[CLI] CLI already installed at: $CliInstallPath"
      return $cliExe
    }
  }

  $tempDir = New-TempDirectory -Prefix "zed-cli-install"

  try {
    $releaseInfo = Get-LatestRelease -ApiUrl $cliApiUrl -Component "CLI"
    $windowsAsset = Find-WindowsAsset -Assets $releaseInfo.assets -Pattern "x86_64"
    $actualInstallPath = Install-FromZip -DownloadUrl $windowsAsset.browser_download_url -InstallPath $CliInstallPath -TempDir $tempDir -Component "CLI" -ExtractedFolderPattern $cliRepoName

    Write-Debug "Install-FromZip returned: '$actualInstallPath'"
    Write-Debug "Expected CliInstallPath: '$CliInstallPath'"

    # Use the expected path for verification
    $finalInstallPath = $CliInstallPath

    $cliExePath = Join-Path $finalInstallPath "$cliRepoName.exe"
    $batPath = Join-Path $finalInstallPath "zed.bat"

    Write-Debug "[Debug] Final cliExePath = '$cliExePath'"
    Write-Debug "[Debug] Final batPath = '$batPath'"

    if (-not (Test-Path $cliExePath)) {
      throw "CLI executable not found: $cliExePath"
    }

    if (-not (Test-Path $batPath)) {
      throw "CLI batch wrapper not found: $batPath"
    }

    Write-Success "[CLI] Verified CLI installation"

    # Add CLI to PATH
    Write-Info "[CLI] Adding CLI to PATH..."
    Add-ToPath $CliInstallPath | Out-Null

    return $cliExePath
  }
  catch {
    Write-Error "[CLI] CLI installation failed: $($_.Exception.Message)"
    throw
  }
  finally {
    # Cleanup
    if (Test-Path $tempDir) {
      Remove-Item $tempDir -Recurse -Force -ErrorAction SilentlyContinue
      Write-Debug "[CLI] Cleaned up CLI temp files"
    }
  }
}

function Set-ZedCli {
  param([string]$ZedExePath, [string]$CliExePath)

  Write-Info "[CLI] Configuring zed-cli-win-unofficial..."

  try {
    Write-Info "[CLI] Refreshing PATH environment variable..."
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path", "User")
    
    # Get the CLI directory for execution
    $cliDir = Split-Path $CliExePath -Parent
    $cliExeName = Split-Path $CliExePath -Leaf
    
    Write-Debug "[DEBUG] [CLI] CLI executable: $cliExeName"
    Write-Debug "[DEBUG] [CLI] CLI directory: $cliDir"
    Write-Debug "[DEBUG] [CLI] Zed path to configure: $ZedExePath"
    
    $configSuccess = $false
    
    # Approach 1: Use full path with & operator
    try {
      Write-Info "[CLI] Attempting configuration with full path..."
      $configResult = & "$CliExePath" "config" "set" "$ZedExePath" 2>&1
      if ($LASTEXITCODE -eq 0) {
        $configSuccess = $true
        Write-Success "[CLI] Configuration successful using full path"
      }
    }
    catch {
      Write-Warning "[CLI] Full path approach failed: $($_.Exception.Message)"
    }
    
    # Approach 2: Try using just the executable name (if it's in PATH)
    if (-not $configSuccess) {
      try {
        Write-Info "[CLI] Attempting configuration using PATH..."
        $configResult = & "$cliExeName" "config" "set" "$ZedExePath" 2>&1
        if ($LASTEXITCODE -eq 0) {
          $configSuccess = $true
          Write-Debug "[CLI] Configuration successful using PATH"
        }
      }
      catch {
        Write-Warning "[CLI] PATH approach failed: $($_.Exception.Message)"
      }
    }
    
    if ($configSuccess) {
      Write-Debug "[CLI] CLI successfully configured with Zed path"
      return $true
    }
    else {
      Write-Error "[CLI] All configuration attempts failed"
      Write-Debug "[CLI] Output: $configResult"
      Write-Debug "[CLI] Exit code: $LASTEXITCODE"
      return $false
    }
  }
  catch {
    Write-Error "[CLI] Configuration failed: $($_.Exception.Message)"
    Write-Warning "[CLI] You may need to run manually: $cliExeName config set `"$ZedExePath`""
    return $false
  }
}

# Main Installation Process
Write-Info "Starting combined Zed + CLI installation..."

$zedExePath = $null
$cliExePath = $null

try {
  # Install Zed Editor
  $zedExePath = Install-ZedEditor
  Write-Success "[Zed] Zed editor installation completed"

  # If installation failed, try to find existing Zed installation
  if (-not $zedExePath) {
    Write-Warning "[Zed] Zed installation failed, looking for existing installation..."
    $possibleZedPaths = @(
      (Join-Path $ZedInstallPath "zed.exe"),
      (Join-Path $env:LOCALAPPDATA "Zed\zed.exe"),
      (Join-Path $env:PROGRAMFILES "Zed\zed.exe")
    )

    foreach ($path in $possibleZedPaths) {
      if (Test-Path $path) {
        $zedExePath = $path
        Write-Success "[Zed] Found existing Zed installation: $zedExePath"
        break
      }
    }

    if (-not $zedExePath) {
      Write-Warning "[Zed] No existing Zed installation found."
    }
  }

  # Install CLI
  $cliExePath = Install-ZedCli
  Write-Success "[CLI] CLI installation completed"

  # Configure CLI if we have both components
  if ($zedExePath -and $cliExePath) {
    # Verify CLI executable exists before configuration
    if (Test-Path $cliExePath) {
      Write-Info "[CLI] Configuring CLI to use Zed installation..."
      $configSuccess = Set-ZedCli -ZedExePath $zedExePath -CliExePath $cliExePath
      if ($configSuccess) {
        Write-Success "[CLI] Configuration completed"
      }
      else {
        Write-Warning "[CLI] Configuration failed - manual setup may be required"
      }
    }
    else {
      Write-Error "[CLI] CLI executable not found at: $cliExePath"
      Write-Warning "[CLI] Configuration skipped - manual setup required"
    }
  }
  else {
    Write-Warning "[CLI] Configuration skipped - missing components"
  }

  # Success summary
  Write-Success "Installation completed successfully!"
  Send-AnalyticsEvent -EventType "zed_with_cli_installation_completed"
  Write-Info "Installed components:"
  if ($zedExePath) {
    Write-Success "Zed Editor: $zedExePath"
  }
  if ($cliExePath) {
    Write-Success "CLI Launcher: $cliExePath"
  }
  Write-Warning "You may need to restart your terminal to use the commands"
  Write-Host "
╔═══════════════════════════════════════════════════╗
║                                                   ║
║  ███████╗███████╗██████╗      ██████╗██╗     ██╗  ║
║  ╚══███╔╝██╔════╝██╔══██╗    ██╔════╝██║     ██║  ║
║    ███╔╝ █████╗  ██║  ██║    ██║     ██║     ██║  ║
║   ███╔╝  ██╔══╝  ██║  ██║    ██║     ██║     ██║  ║
║  ███████╗███████╗██████╔╝    ╚██████╗███████╗██║  ║
║  ╚══════╝╚══════╝╚═════╝      ╚═════╝╚══════╝╚═╝  ║
║                                                   ║
║    🚀 Unofficial Windows CLI for Zed Editor 🚀    ║
║                                                   ║
║                   Version: 1.0.0                  ║
║            https://zedcli.sameerjs.com            ║
║                                                   ║
╚═══════════════════════════════════════════════════╝
" -ForegroundColor Green
}
catch {
  Write-Error "Installation failed: $($_.Exception.Message)"
  Send-AnalyticsEvent -EventType "zed_with_cli_installation_failed"
  exit 1
}

Write-Success "Setup complete! Happy coding with Zed!"
