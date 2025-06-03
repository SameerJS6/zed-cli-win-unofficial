$toolsDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$batSource = Join-Path $toolsDir "zed.bat"
$chocoBinDir = Join-Path $env:ChocolateyInstall "bin"
$batTarget = Join-Path $chocoBinDir "zed.bat"

Copy-Item -Path $batSource -Destination $batTarget -Force
