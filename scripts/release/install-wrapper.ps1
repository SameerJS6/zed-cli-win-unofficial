$BaseGitHubRequest = "https://raw.githubusercontent.com/SameerJS6/zed-cli-win-unofficial/refs/heads/add-mutliple-instance-feature/scripts"
$TempInstallerPath = Join-Path $env:TEMP "CLI-only-installer-$(Get-Random)"

$isDebugging = $false

function Write-Log {
  param(
    [Parameter(Mandatory)]
    [string]$Message,
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

  # Download all required scripts
  $InstallerScriptUrl = "$BaseGitHubRequest/installation/install.ps1"
  $UtilsScriptUrl = "$BaseGitHubRequest/utils/utils.ps1"

  $InstallerScriptLocalPath = Join-Path $TempInstallerPath "install.ps1"
  $UtilsScriptLocalPath = Join-Path $TempInstallerPath "utils.ps1"

  Write-Log "Downloading installation script..." -Color Blue
  Invoke-WebRequest -Uri $InstallerScriptUrl -OutFile $InstallerScriptLocalPath -ErrorAction Stop

  Write-Log "Downloading utility script..." -Color Blue
  Invoke-WebRequest -Uri $UtilsScriptUrl -OutFile $UtilsScriptLocalPath -ErrorAction Stop

  Push-Location $TempInstallerPath

  Write-Log "Executing Installation Script..." -Color Blue
  . ".\\install.ps1"

}
catch {
  Write-Host "An error occured while installation setup: $($_.Exception.Message)" -ForegroundColor Red
}
finally {
  try { Pop-Location } catch {}
} 
