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
  $ZedInstallPath = Join-Path $env:LOCALAPPDATA "Programs\ZedTesting"
}
if (-not $CliInstallPath) {
  $CliInstallPath = Join-Path $env:LOCALAPPDATA $cliRepoName
}

function Install-ZedEditor {
  Write-Status "Starting Zed editor installation..." "Info" "Zed"

  # Check if already installed
  if ((Test-Path $ZedInstallPath) -and -not $Force) {
    $zedExe = Join-Path $ZedInstallPath "zed.exe"
    if (Test-Path $zedExe) {
      Write-Status "Zed editor already installed at: $ZedInstallPath" "Success" "Zed"
      return $zedExe
    }
  }

  # Create temp directory
  # $tempDir = Join-Path $env:TEMP "zed-install-$(Get-Random)"
  # New-Item -ItemType Directory -Path $tempDir -Force | Out-Null
  #  Write-Status "Created temp directory: $tempDir" "Info" "Zed"
 
  $tempDir = New-TempDirectory -Prefix "zed-install"

  try {
    # Get latest release info
    # Write-Status "Fetching latest Zed release information..." "Info" "Zed"
    # $releaseInfo = Invoke-RestMethod -Uri $zedApiUrl -ErrorAction Stop
    # $version = $releaseInfo.tag_name
    # Write-Status "Latest Zed version: $version" "Success" "Zed"

    $releaseInfo = Get-LatestRelease -ApiUrl $zedApiUrl -Component "Zed"

    # Find Windows zip asset (look for x86_64 Windows build)
    # $windowsAsset = $releaseInfo.assets | Where-Object {
    #   $_.name -match ".zip$"
    # } | Select-Object -First 1

    # if (-not $windowsAsset) {
    #   throw "No Windows x86_64 zip asset found in Zed release"
    # }

    $windowsAsset = Find-WindowsAsset -Assets $releaseInfo.assets -Pattern ".zip"
    Install-FromZip -DownloadUrl $windowsAsset.browser_download_url -InstallPath $ZedInstallPath -TempDir $tempDir -Component "Zed" -DeleteZipAfterExtraction
    
    # $downloadUrl = $windowsAsset.browser_download_url
    # $fileName = $windowsAsset.name
    # $downloadPath = Join-Path $tempDir $fileName

    # Write-Status "Downloading: $fileName" "Info" "Zed"
    # Write-Status "From: $downloadUrl" "Info" "Zed"

    # # Download with progress using Get-FileFromWeb
    # Get-FileFromWeb -URL $downloadUrl -File $downloadPath

    # Write-Status "Downloaded: $([math]::Round((Get-Item $downloadPath).Length / 1MB, 2)) MB" "Success" "Zed"

    # # Create installation directory
    # if (Test-Path $ZedInstallPath) {
    #   Write-Status "Removing existing Zed installation..." "Info" "Zed"
    #   Remove-Item $ZedInstallPath -Recurse -Force
    # }

    # New-Item -ItemType Directory -Path $ZedInstallPath -Force | Out-Null
    # Write-Status "Created Zed installation directory: $ZedInstallPath" "Info" "Zed"

    # # Extract zip file
    # Write-Status "Extracting Zed archive..." "Info" "Zed"
    # Expand-Archive -Path $downloadPath -DestinationPath $tempDir -Force

    # # Explicitly delete the downloaded ZIP from the temp directory before copying
    # if (Test-Path $downloadPath) {
    #   Remove-Item $downloadPath -Force
    #   Write-Status "Removed downloaded Zed ZIP from temp directory." "Info" "Zed"
    # }

    # # Find the extracted content and copy to installation directory
    # $extractedItems = Get-ChildItem $tempDir # No longer need -Exclude "*.zip" as it's deleted

    # # Look for zed.exe in extracted content
    # $zedExeFound = $null
    # foreach ($item in $extractedItems) {
    #   if ($item.PSIsContainer) {
    #     $zedExePath = Join-Path $item.FullName "zed.exe"
    #     if (Test-Path $zedExePath) {
    #       $zedExeFound = $item.FullName
    #       break
    #     }
    #   }
    #   else {
    #     if ($item.Name -eq "zed.exe") {
    #       $zedExeFound = $tempDir
    #       break
    #     }
    #   }
    # }

    # if (-not $zedExeFound) {
    #   throw "Could not find zed.exe in extracted archive"
    # }

    # # Copy contents to installation directory
    # Copy-Item "$zedExeFound\*" $ZedInstallPath -Recurse -Force
    # Write-Status "Installed Zed files to: $ZedInstallPath" "Success" "Zed"

    # Verify installation
    $zedExePath = Join-Path $ZedInstallPath "zed.exe"
    if (-not (Test-Path $zedExePath)) {
      throw "Zed executable not found after installation: $zedExePath"
    }

    Write-Status "Verified Zed installation" "Success" "Zed"

    return $zedExePath
  }
  catch {
    Write-Status "Zed installation failed: $($_.Exception.Message)" "Error" "Zed"
    throw
  }
  finally {
    # Cleanup
    if (Test-Path $tempDir) {
      Remove-Item $tempDir -Recurse -Force -ErrorAction SilentlyContinue
      Write-Status "Cleaned up Zed temp files" "Info" "Zed"
    }
  }
}

function Install-ZedCli {
  Write-Status "Starting zed-cli-win-unofficial installation..." "Info" "CLI"

  # Check if already installed
  if ((Test-Path $CliInstallPath) -and -not $Force) {
    $cliExe = Join-Path $CliInstallPath "$cliRepoName.exe"
    if (Test-Path $cliExe) {
      Write-Status "CLI already installed at: $CliInstallPath" "Success" "CLI"
      return $cliExe
    }
  }

  # Create temp directory
  # $tempDir = Join-Path $env:TEMP "zed-cli-install-$(Get-Random)"
  # New-Item -ItemType Directory -Path $tempDir -Force | Out-Null
  # Write-Status "Created temp directory: $tempDir" "Info" "CLI"

  $tempDir = New-TempDirectory -Prefix "zed-cli-install"

  try {
    # Get latest release info
    # Write-Status "Fetching latest CLI release information..." "Info" "CLI"
    # $releaseInfo = Invoke-RestMethod -Uri $cliApiUrl -ErrorAction Stop
    # $version = $releaseInfo.tag_name
    # Write-Status "Latest CLI version: $version" "Success" "CLI"

    $releaseInfo = Get-LatestRelease -ApiUrl $cliApiUrl -Component "CLI"

    $windowsAsset = Find-WindowsAsset -Assets $releaseInfo.assets -Pattern "x86_64"
    # Find Windows zip asset
    # $windowsAsset = $releaseInfo.assets | Where-Object {
    #   $_.name -match "x86_64"
    # } | Select-Object -First 1

    # if (-not $windowsAsset) {
    #   throw "No Windows zip asset found in CLI release"
    # }

    Install-FromZip -DownloadUrl $windowsAsset.browser_download_url -InstallPath $CliInstallPath -TempDir $tempDir -Component "CLI" -ExtractedFolderPattern $cliRepoName -DeleteZipAfterExtraction

    # $downloadUrl = $windowsAsset.browser_download_url
    # $fileName = $windowsAsset.name
    # $downloadPath = Join-Path $tempDir $fileName

    # Write-Status "Downloading: $fileName" "Info" "CLI"
    # Write-Status "From: $downloadUrl" "Info" "CLI"

    # # Download with progress using Get-FileFromWeb
    # Get-FileFromWeb -URL $downloadUrl -File $downloadPath

    # Write-Status "Downloaded: $([math]::Round((Get-Item $downloadPath).Length / 1MB, 2)) MB" "Success" "CLI"

    # # Create installation directory
    # if (Test-Path $CliInstallPath) {
    #   Write-Status "Removing existing CLI installation..." "Info" "CLI"
    #   Remove-Item $CliInstallPath -Recurse -Force
    # }

    # New-Item -ItemType Directory -Path $CliInstallPath -Force | Out-Null
    # Write-Status "Created CLI installation directory: $CliInstallPath" "Info" "CLI"

    # # Extract zip file
    # Write-Status "Extracting CLI archive..." "Info" "CLI"
    # Expand-Archive -Path $downloadPath -DestinationPath $tempDir -Force

    # # Find the extracted folder
    # $extractedFolder = Get-ChildItem $tempDir -Directory | Where-Object { $_.Name -match $cliRepoName } | Select-Object -First 1

    # if (-not $extractedFolder) {
    #   throw "Could not find extracted CLI folder"
    # }

    # # Copy contents to installation directory
    # Copy-Item "$($extractedFolder.FullName)\*" $CliInstallPath -Recurse -Force
    # Write-Status "Installed CLI files to: $CliInstallPath" "Success" "CLI"

    # Verify installation
    $cliExePath = Join-Path $CliInstallPath "$cliRepoName.exe"
    $batPath = Join-Path $CliInstallPath "zed.bat"

    if (-not (Test-Path $cliExePath)) {
      throw "CLI executable not found: $cliExePath"
    }

    if (-not (Test-Path $batPath)) {
      throw "CLI batch wrapper not found: $batPath"
    }

    Write-Status "Verified CLI installation" "Success" "CLI"

    # Add CLI to PATH
    Write-Status "Adding CLI to PATH..." "Info" "CLI"
    Add-ToPath $CliInstallPath | Out-Null

    return $cliExePath
  }
  catch {
    Write-Status "CLI installation failed: $($_.Exception.Message)" "Error" "CLI"
    throw
  }
  finally {
    # Cleanup
    if (Test-Path $tempDir) {
      Remove-Item $tempDir -Recurse -Force -ErrorAction SilentlyContinue
      Write-Status "Cleaned up CLI temp files" "Info" "CLI"
    }
  }
}

function Set-ZedCli {
  param([string]$ZedExePath, [string]$CliExePath)

  Write-Status "Configuring zed-cli-win-unofficial..." "Info" "CLI"

  try {
    # Refresh PATH for current session to ensure CLI is available
    Write-Status "Refreshing PATH environment variable..." "Info" "CLI"
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path", "User")
    
    # Get the CLI directory for execution
    $cliDir = Split-Path $CliExePath -Parent
    $cliExeName = Split-Path $CliExePath -Leaf
    
    Write-Status "CLI executable: $cliExeName" "Info" "CLI"
    Write-Status "CLI directory: $cliDir" "Info" "CLI"
    Write-Status "Zed path to configure: $ZedExePath" "Info" "CLI"
    
    # Try multiple approaches to run the CLI command
    $configSuccess = $false
    
    # Approach 1: Use full path with & operator
    try {
      Write-Status "Attempting configuration with full path..." "Info" "CLI"
      $configResult = & "$CliExePath" "config" "set" "$ZedExePath" 2>&1
      if ($LASTEXITCODE -eq 0) {
        $configSuccess = $true
        Write-Status "Configuration successful using full path" "Success" "CLI"
      }
    }
    catch {
      Write-Status "Full path approach failed: $($_.Exception.Message)" "Warning" "CLI"
    }
    
    # Approach 2: Try using just the executable name (if it's in PATH)
    if (-not $configSuccess) {
      try {
        Write-Status "Attempting configuration using PATH..." "Info" "CLI"
        $configResult = & "$cliExeName" "config" "set" "$ZedExePath" 2>&1
        if ($LASTEXITCODE -eq 0) {
          $configSuccess = $true
          Write-Status "Configuration successful using PATH" "Success" "CLI"
        }
      }
      catch {
        Write-Status "PATH approach failed: $($_.Exception.Message)" "Warning" "CLI"
      }
    }
    
    # Approach 3: Change directory and run locally
    if (-not $configSuccess) {
      try {
        Write-Status "Attempting configuration from CLI directory..." "Info" "CLI"
        Push-Location $cliDir
        $configResult = & ".\$cliExeName" "config" "set" "$ZedExePath" 2>&1
        if ($LASTEXITCODE -eq 0) {
          $configSuccess = $true
          Write-Status "Configuration successful from CLI directory" "Success" "CLI"
        }
        Pop-Location
      }
      catch {
        if (Get-Location) { Pop-Location }
        Write-Status "Local directory approach failed: $($_.Exception.Message)" "Warning" "CLI"
      }
    }
    
    if ($configSuccess) {
      Write-Status "CLI successfully configured with Zed path" "Success" "CLI"
      return $true
    }
    else {
      Write-Status "All configuration attempts failed" "Error" "CLI"
      Write-Status "Output: $configResult" "Info" "CLI"
      Write-Status "Exit code: $LASTEXITCODE" "Info" "CLI"
      return $false
    }
  }
  catch {
    Write-Status "Configuration failed: $($_.Exception.Message)" "Error" "CLI"
    Write-Status "You may need to run manually: $cliExeName config set `"$ZedExePath`"" "Warning" "CLI"
    return $false
  }
}

# Main Installation Process
Write-Status "🚀 Starting combined Zed + CLI installation..."

$zedExePath = $null
$cliExePath = $null

try {
  # Install Zed Editor
  $zedExePath = Install-ZedEditor
  Write-Status "✅ Zed editor installation completed" "Success" "Zed"

  # If installation failed, try to find existing Zed installation
  if (-not $zedExePath) {
    Write-Status "Zed installation failed, looking for existing installation..." "Warning" "Zed"
    $possibleZedPaths = @(
      (Join-Path $ZedInstallPath "zed.exe"),
      (Join-Path $env:LOCALAPPDATA "Zed\zed.exe"),
      (Join-Path $env:PROGRAMFILES "Zed\zed.exe")
    )

    foreach ($path in $possibleZedPaths) {
      if (Test-Path $path) {
        $zedExePath = $path
        Write-Status "Found existing Zed installation: $zedExePath" "Success" "Zed"
        break
      }
    }

    if (-not $zedExePath) {
      Write-Status "No existing Zed installation found." "Warning" "Zed"
    }
  }

  # Install CLI
  $cliExePath = Install-ZedCli
  Write-Status "✅ CLI installation completed" "Success" "CLI"

  # Configure CLI if we have both components
  if ($zedExePath -and $cliExePath) {
    # Verify CLI executable exists before configuration
    if (Test-Path $cliExePath) {
      Write-Status "Configuring CLI to use Zed installation..." "Info" "CLI"
      $configSuccess = Set-ZedCli -ZedExePath $zedExePath -CliExePath $cliExePath
      if ($configSuccess) {
        Write-Status "✅ Configuration completed" "Success" "CLI"
      }
      else {
        Write-Status "⚠️  Configuration failed - manual setup may be required" "Warning" "CLI"
      }
    }
    else {
      Write-Status "CLI executable not found at: $cliExePath" "Error" "CLI"
      Write-Status "⚠️  Configuration skipped - manual setup required" "Warning" "CLI"
    }
  }
  else {
    Write-Status "⚠️  Configuration skipped - missing components" "Warning" "CLI"
  }

  # Success summary
  Write-Status ""
  Write-Status "🎉 Installation completed successfully!" "Success"
  Write-Status ""
  Write-Status "Installed components:" "Info"
  if ($zedExePath) {
    Write-Status "  🎨 Zed Editor: $zedExePath" "Success"
  }
  if ($cliExePath) {
    Write-Status "  ⚡ CLI Launcher: $cliExePath" "Success"
  }
  Write-Status ""
  Write-Status "⚠️  You may need to restart your terminal to use the commands" "Warning"
}
catch {
  Write-Status "Installation failed: $($_.Exception.Message)" "Error"
  exit 1
}

Write-Status "🎉 Setup complete! Happy coding with Zed! 🚀" "Success"
