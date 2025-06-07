$BaseGitHubRequest = "https://raw.githubusercontent.com/SameerJS6/zed-cli-win-unofficial/refs/heads/main/scripts"
$TempInstallerPath = Join-Path $env:TEMP "Zed-With-CLI-Installer-$(Get-Random)"

$isDebugging = $false

function Write-Log {
  param(
    [Parameter(Mandatory)]
    [string]$Message,
    [Parameter(Mandatory)]
    [ValidateSet("Black", "Blue", "Cyan", "DarkBlue", "DarkCyan", "DarkGray", "DarkGreen", "DarkMagenta", "DarkRed", "DarkYellow", "Gray", "Green", "Magenta", "Red", "White", "Yellow")]
    [string]$Color
  )
  
  if ($isDebugging) {
    Write-Host "[DEBUG] $Message" -ForegroundColor $Color
  }
}

try {
  Write-Log "Creating temporary installation directory..." -Color Blue
  New-Item -ItemType Directory -Path $TempInstallerPath -ErrorAction Stop | Out-Null

  $InstallerScriptUri = "$BaseGitHubRequest/installation/install-with-zed.ps1"
  $UtilsScriptUri = "$BaseGitHubRequest/utils/utils.ps1"

  $InstallerScriptLocalPath = Join-Path $TempInstallerPath "install-with-zed.ps1"
  $UtilsScriptLocalPath = Join-Path $TempInstallerPath "utils.ps1"

  Write-Log "Downloading Installation Script..." -Color Blue
  Invoke-WebRequest -Uri $InstallerScriptUri -OutFile $InstallerScriptLocalPath -ErrorAction Stop

  Write-Log "Downloading Utility Script..." -Color Blue
  Invoke-WebRequest -Uri $UtilsScriptUri -OutFile $UtilsScriptLocalPath -ErrorAction Stop | Out-Null

  Push-Location $TempInstallerPath

  Write-Log "Executing Installation Script..." -Color Blue
  . ".\\install-with-zed.ps1"

}
catch {
  Write-Host "An error occured during installation setup: $($_.Exception.Message)" -ForegroundColor Red
}
finally {
  try { Pop-Location } catch {}
}