function Write-Status {
  param([string]$Message, [string]$Type = "Info", [string]$Component = "", [switch]$Debug)
  $color = switch ($Type) {
    "Success" { "Green" }
    "Warning" { "Yellow" }
    "Error" { "Red" }
    "Debug" { "Gray" }
    default { "Cyan" }
  }

  $icon = switch ($Component) {
    "Zed" { "[ZED]" }
    "CLI" { "[CLI]" }
    "Debug" { "[DEBUG]" }
    default { "[INFO]" }
  }

  if (-not $Debug) { 
    Write-Host "$icon $Message" -ForegroundColor $color
  }
}

function Add-ToPath {
  param([string]$Directory)

  # Get current user PATH
  $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
  
  # Check if already in PATH
  if ($currentPath -split ';' | Where-Object { $_ -eq $Directory }) {
    Write-Status "Directory already in PATH" "Success" 
    return $true
  }

  # Add to PATH
  $newPath = if ($currentPath) { "$currentPath;$Directory" } else { $Directory }

  try {
    [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
    Write-Status "Added to PATH: $Directory" "Success"
    return $true
  }
  catch {
    Write-Status "Failed to update PATH: $($_.Exception.Message)" "Error" 
    return $false
  }
}

# function Remove-FromPath {
#   param([string]$Directory)
  
#   $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
#   $pathEntries = $currentPath -split ';' | Where-Object { $_ -ne $Directory -and $_ -ne "" }
#   $newPath = $pathEntries -join ';'
  
#   try {
#     [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
#     Write-Status "Removed from PATH: $Directory" "Success"
#     return $true
#   }
#   catch {
#     Write-Status "Failed to remove from PATH: $($_.Exception.Message)" "Warning"
#     return $false
#   }
# }


function New-TempDirectory {
  param([string]$Prefix = "install")
  
  $tempDir = Join-Path $env:TEMP "$Prefix-$(Get-Random)"
  New-Item -ItemType Directory -Path $tempDir -Force | Out-Null
  return $tempDir
}

function Get-LatestRelease {
  param(
    [Parameter(Mandatory)]
    [string]$ApiUrl,
    
    [Parameter(Mandatory)]
    [string]$Component
  )
  
  Write-Status "Fetching latest release information..." "Info" $Component -Debug
  $releaseInfo = Invoke-RestMethod -Uri $ApiUrl -ErrorAction Stop
  $version = $releaseInfo.tag_name
  Write-Status "Latest version: $version" "Success" $Component -Debug
  
  return $releaseInfo
}

function Find-WindowsAsset {
  param(
    [Parameter(Mandatory)]
    [array]$Assets,
    
    [string]$Pattern = "x86_64"    
  )
  
  $windowsAsset = $Assets | Where-Object { 
    $_.name -match $Pattern
  } | Select-Object -First 1
    
  if (-not $windowsAsset) {
    throw "No Windows asset found matching pattern '$Pattern'"
  }
  
  return $windowsAsset
}

function Install-FromZip {
  param(
    [Parameter(Mandatory)]
    [string]$DownloadUrl,
    
    [Parameter(Mandatory)]
    [string]$InstallPath,
    
    [Parameter(Mandatory)]
    [string]$TempDir,
    
    [string]$Component = "",
    
    [string]$ExtractedFolderPattern = "",
    
    [switch]$DeleteZipAfterExtraction
  )
  
  $fileName = Split-Path $DownloadUrl -Leaf
  $downloadPath = Join-Path $TempDir $fileName
  
  Write-Status "Downloading: $fileName" "Info" $Component
  Write-Status "From: $DownloadUrl" "Info" $Component
  
  # Download with progress
  Get-FileFromWeb -URL $DownloadUrl -File $downloadPath
  
  Write-Status "Downloaded: $([math]::Round((Get-Item $downloadPath).Length / 1MB, 2)) MB" "Success" $Component
  
  # Create installation directory
  if (Test-Path $InstallPath) {
    Write-Status "Removing existing installation..." "Info" $Component
    Remove-Item $InstallPath -Recurse -Force
  }
  
  New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
  Write-Status "Created installation directory: $InstallPath" "Info" $Component
  
  # Extract zip file
  Write-Status "Extracting archive..." "Info" $Component
  Expand-Archive -Path $downloadPath -DestinationPath $TempDir -Force
  
  # Delete ZIP if requested
  if ($DeleteZipAfterExtraction -and (Test-Path $downloadPath)) {
    Remove-Item $downloadPath -Force
    Write-Status "Removed downloaded ZIP from temp directory." "Info" $Component
  }
  
  # Find extracted content
  if ($ExtractedFolderPattern) {
    $extractedFolder = Get-ChildItem $TempDir -Directory | Where-Object { $_.Name -match $ExtractedFolderPattern } | Select-Object -First 1
    if (-not $extractedFolder) {
      throw "Could not find extracted folder matching pattern '$ExtractedFolderPattern'"
    }
    $sourcePath = $extractedFolder.FullName
  }
  else {
    # Use all items except the ZIP file
    $extractedItems = Get-ChildItem $TempDir -Exclude "*.zip"
    if ($extractedItems.Count -eq 1 -and $extractedItems[0].PSIsContainer) {
      $sourcePath = $extractedItems[0].FullName
    }
    else {
      $sourcePath = $TempDir
    }
  }
  
  # Copy contents to installation directory
  Copy-Item "$sourcePath\*" $InstallPath -Recurse -Force
  Write-Status "Installed files to: $InstallPath" "Success" $Component
  
  return $InstallPath
}

function Get-FileFromWeb {
  param (
    # Parameter help description
    [Parameter(Mandatory)]
    [string]$URL,
  
    # Parameter help description
    [Parameter(Mandatory)]
    [string]$File 
  )
  Begin {
    function Show-Progress {
      param (
        # Enter total value
        [Parameter(Mandatory)]
        [Single]$TotalValue,
        
        # Enter current value
        [Parameter(Mandatory)]
        [Single]$CurrentValue,
        
        # Enter custom progresstext
        [Parameter(Mandatory)]
        [string]$ProgressText,
        
        # Enter value suffix
        [Parameter()]
        [string]$ValueSuffix,
        
        # Enter bar lengh suffix
        [Parameter()]
        [int]$BarSize = 40,

        # show complete bar
        [Parameter()]
        [switch]$Complete
      )
            
      # calc %
      $percent = $CurrentValue / $TotalValue
      $percentComplete = $percent * 100
      if ($ValueSuffix) {
        $ValueSuffix = " $ValueSuffix" # add space in front
      }
      if ($psISE) {
        Write-Progress "$ProgressText $CurrentValue$ValueSuffix of $TotalValue$ValueSuffix" -id 0 -percentComplete $percentComplete            
      }
      else {
        # build progressbar with string function
        $curBarSize = $BarSize * $percent
        $progbar = ""
        $progbar = $progbar.PadRight($curBarSize, [char]9608)
        $progbar = $progbar.PadRight($BarSize, [char]9617)
        
        if (!$Complete.IsPresent) {
          Write-Host -NoNewLine "`r$ProgressText $progbar [ $($CurrentValue.ToString("#.###").PadLeft($TotalValue.ToString("#.###").Length))$ValueSuffix / $($TotalValue.ToString("#.###"))$ValueSuffix ] $($percentComplete.ToString("##0.00").PadLeft(6)) % complete"
        }
        else {
          Write-Host -NoNewLine "`r$ProgressText $progbar [ $($TotalValue.ToString("#.###").PadLeft($TotalValue.ToString("#.###").Length))$ValueSuffix / $($TotalValue.ToString("#.###"))$ValueSuffix ] $($percentComplete.ToString("##0.00").PadLeft(6)) % complete"                    
        }                
      }   
    }
  }
  Process {
    try {
      $storeEAP = $ErrorActionPreference
      $ErrorActionPreference = 'Stop'
        
      # invoke request
      $request = [System.Net.HttpWebRequest]::Create($URL)
      $response = $request.GetResponse()
  
      if ($response.StatusCode -eq 401 -or $response.StatusCode -eq 403 -or $response.StatusCode -eq 404) {
        throw "Remote file either doesn't exist, is unauthorized, or is forbidden for '$URL'."
      }
  
      if ($File -match '^\.\\') {
        $File = Join-Path (Get-Location -PSProvider "FileSystem") ($File -Split '^\.')[1]
      }
            
      if ($File -and !(Split-Path $File)) {
        $File = Join-Path (Get-Location -PSProvider "FileSystem") $File
      }

      if ($File) {
        $fileDirectory = $([System.IO.Path]::GetDirectoryName($File))
        if (!(Test-Path($fileDirectory))) {
          [System.IO.Directory]::CreateDirectory($fileDirectory) | Out-Null
        }
      }

      [long]$fullSize = $response.ContentLength
      $fullSizeMB = $fullSize / 1024 / 1024
  
      # define buffer
      [byte[]]$buffer = new-object byte[] 1048576
      [long]$total = [long]$count = 0
  
      # create reader / writer
      $reader = $response.GetResponseStream()
      $writer = new-object System.IO.FileStream $File, "Create"
  
      # start download
      $finalBarCount = 0 #show final bar only one time
      do {
          
        $count = $reader.Read($buffer, 0, $buffer.Length)
          
        $writer.Write($buffer, 0, $count)
              
        $total += $count
        $totalMB = $total / 1024 / 1024
          
        if ($fullSize -gt 0) {
          Show-Progress -TotalValue $fullSizeMB -CurrentValue $totalMB -ProgressText "Downloading" -ValueSuffix "MB"
        }

        if ($total -eq $fullSize -and $count -eq 0 -and $finalBarCount -eq 0) {
          Show-Progress -TotalValue $fullSizeMB -CurrentValue $totalMB -ProgressText "Downloading" -ValueSuffix "MB" -Complete
          $finalBarCount++
        }

      } while ($count -gt 0)
            
      Write-Host "" # New line after progress bar
    }
  
    catch {
      $ExeptionMsg = $_.Exception.Message
      Write-Host "Download breaks with error : $ExeptionMsg"
      throw
    }
  
    finally {
      # cleanup
      if ($reader) { $reader.Close() }
      if ($writer) { $writer.Flush(); $writer.Close() }
        
      $ErrorActionPreference = $storeEAP
      [GC]::Collect()
    }    
  }
}
