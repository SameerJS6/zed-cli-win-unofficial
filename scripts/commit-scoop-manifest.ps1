Write-Host "🚀 Committing scoop manifest changes..."

if (Test-Path 'zed-cli-win-unofficial.json') {
  Write-Host "✅ Found fixed manifest: zed-cli-win-unofficial.json"
  Write-Host "   Size: $((Get-Item 'zed-cli-win-unofficial.json').Length) bytes"
  
  # Debug: Show git status
  Write-Host "🔍 Git status:"
  git status
  git branch -a
  
  # Configure git user (required for commits)
  git config user.name "SameerJS6"
  git config user.email "sameerjs6@users.noreply.github.com"
  
  # For tag-based workflows, we need to checkout the main branch
  Write-Host "🔄 Ensuring we're on the main branch..."
  git fetch origin main
  git checkout main
  if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Failed to checkout main branch"
    exit 1
  }
  
  # Get the current branch name
  $currentBranch = git branch --show-current
  Write-Host "📋 Current branch: $currentBranch"
  
  # Add the fixed manifest file
  git add zed-cli-win-unofficial.json
  if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Failed to add file to git staging"
    exit 1
  }
  Write-Host "📝 Added file to git staging"
  
  # Check if there are changes to commit
  $gitStatus = git status --porcelain
  if ($gitStatus) {
    # Commit with descriptive message
    $commitMessage = "Scoop update for zed-cli-win-unofficial version ${env:GITHUB_REF_NAME} with zed.bat"
    git commit -m $commitMessage
    if ($LASTEXITCODE -ne 0) {
      Write-Host "❌ Failed to commit changes"
      exit 1
    }
    Write-Host "✅ Committed changes with message: '$commitMessage'"
    
    # Push to the current branch
    git push origin $currentBranch
    if ($LASTEXITCODE -ne 0) {
      Write-Host "❌ Failed to push changes to $currentBranch"
      exit 1
    }
    Write-Host "🚀 Pushed changes to $currentBranch branch"
    
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