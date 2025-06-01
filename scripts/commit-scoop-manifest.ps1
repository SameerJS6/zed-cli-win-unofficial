Write-Host "🚀 Committing scoop manifest changes..."

if (Test-Path 'zed-cli-win-unofficial.json') {
  Write-Host "✅ Found fixed manifest: zed-cli-win-unofficial.json"
  Write-Host "   Size: $((Get-Item 'zed-cli-win-unofficial.json').Length) bytes"
  
  # Configure git user (required for commits)
  git config user.name "SameerJS6"
  git config user.email "sameerjs6@users.noreply.github.com"
  
  # Add the fixed manifest file
  git add zed-cli-win-unofficial.json
  Write-Host "📝 Added file to git staging"
  
  # Check if there are changes to commit
  $gitStatus = git status --porcelain
  if ($gitStatus) {
    # Commit with descriptive message
    $commitMessage = "Scoop update for zed-cli-win-unofficial version ${env:GITHUB_REF_NAME} with zed.bat"
    git commit -m $commitMessage
    Write-Host "✅ Committed changes with message: '$commitMessage'"
    
    # Push to main branch
    git push origin main
    Write-Host "🚀 Pushed changes to main branch"
    
    Write-Host ""
    Write-Host "🎉 Successfully updated scoop manifest!"
  }
  else {
    Write-Host "ℹ️ No changes to commit (manifest may already be up to date)"
  }
  
}
else {
  Write-Host "❌ No manifest file found to commit"
  exit 1
} 