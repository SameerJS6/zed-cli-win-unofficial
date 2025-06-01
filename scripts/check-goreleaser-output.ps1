Write-Host "📁 Checking dist directory contents..."
if (Test-Path 'dist') {
  Write-Host "✅ dist folder exists!"
  Get-ChildItem -Path 'dist' -Recurse | Select-Object FullName, Length
}
else {
  Write-Host "❌ dist folder not found!"
}

Write-Host ""
Write-Host "🔍 Looking for scoop manifest..."
if (Test-Path 'dist/scoop/zed-cli-win-unofficial.json') {
  Write-Host "✅ Scoop manifest found!"
  Write-Host "📦 Original manifest contents:"
  Get-Content 'dist/scoop/zed-cli-win-unofficial.json' | Write-Host
}
else {
  Write-Host "❌ Scoop manifest not found in expected location"
  Write-Host "📁 Checking what's in dist/scoop/:"
  if (Test-Path 'dist/scoop') {
    Get-ChildItem 'dist/scoop' | Write-Host
  }
  else {
    Write-Host "❌ dist/scoop directory doesn't exist"
  }
} 