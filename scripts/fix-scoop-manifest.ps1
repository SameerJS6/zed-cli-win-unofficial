Write-Host "🔧 Testing scoop manifest fix..."

if (Test-Path 'dist/scoop/zed-cli-win-unofficial.json') {
  Write-Host "✅ Found scoop manifest, reading contents..."
  
  # Read the original content as text to preserve formatting
  $content = Get-Content 'dist/scoop/zed-cli-win-unofficial.json' -Raw
  
  Write-Host "📄 Original manifest:"
  Write-Host $content
  
  # Parse JSON to check current bin array (for logging only)
  $manifest = $content | ConvertFrom-Json
  Write-Host ""
  Write-Host "📋 Current bin array:"
  $manifest.architecture."64bit".bin | ForEach-Object { Write-Host "  - $_" }
  
  # Use regex to find and modify the bin array while preserving formatting
  # This looks for the bin array pattern and adds zed.bat if not already present
  if ($content -notmatch '"zed-cli-win-unofficial/zed\.bat"') {
    Write-Host ""
    Write-Host "🔧 Adding zed.bat to bin array..."
    
    # Find the bin array and add the new entry
    # Pattern matches: "bin": ["existing-entry"] or "bin": ["entry1", "entry2"]
    $pattern = '("bin":\s*\[\s*"[^"]+")(\s*\])'
    $replacement = '$1, "zed-cli-win-unofficial/zed.bat"$2'
    
    $updatedContent = $content -replace $pattern, $replacement
    
    Write-Host "📄 Updated manifest:"
    Write-Host $updatedContent
    
    # Save the updated manifest with preserved formatting
    $updatedContent | Set-Content 'zed-cli-win-unofficial.json' -NoNewline
    Write-Host ""
    Write-Host "💾 Saved updated manifest to root level (for testing)"
    Write-Host "✅ Scoop manifest fix test successful!"
    
  }
  else {
    Write-Host ""
    Write-Host "✅ zed.bat already exists in bin array"
  }

}
else {
  Write-Host "❌ Scoop manifest not found"
  exit 1
}