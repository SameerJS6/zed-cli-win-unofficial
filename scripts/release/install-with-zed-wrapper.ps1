$BaseGitHubRequest = "https://raw.githubusercontent.com/SameerJS6/zed-cli-win-unofficial/refs/heads/main/scripts"
$TempInstallerPath = Join-Path $env:TEMP "Zed-With-CLI-Installer-$(Get-Random)"

try {
  Write-Host "Creating temporary installation directory..."
  New-Item -ItemType Directory -Path $TempInstallerPath -ErrorAction Stop | Out-Null

  $InstallerScriptUri = "$BaseGitHubRequest/installation/install-with-zed.ps1"
  $UtilsScriptUri = "$BaseGitHubRequest/utils/utils.psm1"

  $InstallerScriptLocalPath = Join-Path $TempInstallerPath "install-with-zed.ps1"
  $UtilsScriptLocalPath = Join-Path $TempInstallerPath "utils.psm1"

  Write-Host "Downloading Installation Script..."
  Invoke-WebRequest -Uri $InstallerScriptUri -OutFile $InstallerScriptLocalPath -ErrorAction Stop

  Write-Host "Downloading Utility Script..."
  Invoke-WebRequest -Uri $UtilsScriptUri -OutFile $UtilsScriptLocalPath -ErrorAction Stop | Out-Null

  Push-Location $TempInstallerPath

  Write-Host "Executing Installation Script..."
  . "./install-with-zed.ps1"

}
catch {
  Write-Host "An error occured during installation setup: $($_.Exception.Message)" -ForegroundColor Red
}
finally {
  try { Pop-Location } catch {}
}