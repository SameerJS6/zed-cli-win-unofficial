Write-Host "🚀 Committing scoop manifest changes..."

# Check if the source manifest exists first
if (-not (Test-Path 'dist/scoop/zed-cli-win-unofficial.json')) {
  Write-Host "❌ No scoop manifest found in dist/scoop/"
  exit 1
}

# Debug: Show git status
Write-Host "🔍 Git status before checkout:"
git status
git branch -a

# Configure git user (required for commits)
git config user.name "SameerJS6"
git config user.email "sameerjs6@users.noreply.github.com"

# For tag-based workflows, we need to checkout the main branch FIRST
Write-Host "🔄 Checking out main branch..."
git fetch origin main
git checkout main
if ($LASTEXITCODE -ne 0) {
  Write-Host "❌ Failed to checkout main branch"
  exit 1
}

# Get the current branch name
$currentBranch = git branch --show-current
Write-Host "📋 Current branch: $currentBranch"

# Now that we're on main, create the fixed manifest
Write-Host "🔧 Creating fixed scoop manifest..."
if (Test-Path 'dist/scoop/zed-cli-win-unofficial.json') {
  # Read the original content as text to preserve formatting
  $content = Get-Content 'dist/scoop/zed-cli-win-unofficial.json' -Raw
  
  # Use regex to find and modify the bin array while preserving formatting
  if ($content -notmatch '"zed-cli-win-unofficial/zed\.bat"') {
    Write-Host "🔧 Adding zed.bat to bin array..."
    
    # Find the bin array and add the new entry
    $pattern = '("bin":\s*\[\s*"[^"]+")(\s*\])'
    $replacement = '$1, "zed-cli-win-unofficial/zed.bat"$2'
    
    $updatedContent = $content -replace $pattern, $replacement
    
    # Save the updated manifest
    $updatedContent | Set-Content 'zed-cli-win-unofficial.json' -NoNewline
    Write-Host "✅ Created fixed manifest: zed-cli-win-unofficial.json"
    Write-Host "   Size: $((Get-Item 'zed-cli-win-unofficial.json').Length) bytes"
  }
  else {
    Write-Host "ℹ️ zed.bat already exists in bin array, copying as-is"
    Copy-Item 'dist/scoop/zed-cli-win-unofficial.json' 'zed-cli-win-unofficial.json'
  }
}
else {
  Write-Host "❌ Source manifest disappeared"
  exit 1
}

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