$BaseGitHubRequest = "https://raw.githubusercontent.com/SameerJS6/zed-cli-win-unofficial/refs/heads/main/scripts"
$TempInstallerPath = Join-Path $env:TEMP "CLI-only-installer-$(Get-Random)"

try {
  Write-Host "Creating temporary installation directory..."
  New-Item -ItemType Directory -Path $TempInstallerPath -ErrorAction Stop | Out-Null

  # Download all required scripts
  $InstallerScriptUrl = "$BaseGitHubRequest/installation/install.ps1"
  $UtilsScriptUrl = "$BaseGitHubRequest/utils/utils.ps1"

  $InstallerScriptLocalPath = Join-Path $TempInstallerPath "install.ps1"
  $UtilsScriptLocalPath = Join-Path $TempInstallerPath "utils.ps1"

  Write-Host "Downloading installation script..."
  Invoke-WebRequest -Uri $InstallerScriptUrl -OutFile $InstallerScriptLocalPath -ErrorAction Stop

  Write-Host "Downloading utility script..."
  Invoke-WebRequest -Uri $UtilsScriptUrl -OutFile $UtilsScriptLocalPath -ErrorAction Stop

  Push-Location $TempInstallerPath

  Write-Host "Executing Installation Script..."
  . ".\\install.ps1"

}
catch {
  Write-Host "An error occured while installation setup: $($_.Exception.Message)" -ForegroundColor Red
}
finally {
  try { Pop-Location } catch {}
} 
