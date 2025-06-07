# Controls whether Write-Debug messages are displayed.
# Set to $true for verbose debugging output, $false to suppress.
if (-not (Test-Path Variable:Global:ScriptDebugMode)) {
  $Global:ScriptDebugMode = $false
}

function Write-LogInternal {
  param(
    [Parameter(Mandatory)]
    [string]$Message,
    [Parameter(Mandatory)]
    [string]$TypePrefix,
    [Parameter(Mandatory)]
    [string]$Color
  )
  Write-Host "$TypePrefix $Message" -ForegroundColor $Color
}

# Logs a message only if $Global:ScriptDebugMode is $true.
function Write-Debug {
  param(
    [Parameter(Mandatory)]
    [string]$Message
  )
  Write-Host "[utils.ps1] Inside Write-Debug, Global:ScriptDebugMode is: '$($Global:ScriptDebugMode)'" -ForegroundColor Magenta
  if ($Global:ScriptDebugMode) {
    Write-LogInternal -Message $Message -TypePrefix "[DEBUG]" -Color "Gray"
  }
}

function Write-Info {
  param(
    [Parameter(Mandatory)]
    [string]$Message
  )
  Write-LogInternal -Message $Message -TypePrefix "[INFO]" -Color "Cyan"
}

# Logs a success message.
function Write-Success {
  param(
    [Parameter(Mandatory)]
    [string]$Message
  )
  Write-LogInternal -Message $Message -TypePrefix "[SUCCESS]" -Color "Green"
}

# Logs a warning message.
function Write-Warning {
  param(
    [Parameter(Mandatory)]
    [string]$Message
  )
  Write-LogInternal -Message $Message -TypePrefix "[WARNING]" -Color "Yellow"
}

# Logs an error message.
function Write-Error {
  param(
    [Parameter(Mandatory)]
    [string]$Message
  )
  Write-LogInternal -Message $Message -TypePrefix "[ERROR]" -Color "Red"
}
# --- End of Logging Functions ---


function Add-ToPath {
  param([string]$Directory)

  # Get current user PATH
  $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")

  # Check if already in PATH
  if ($currentPath -split ';' | Where-Object { $_ -eq $Directory }) {
    Write-Success "Directory already in PATH: $Directory"
    return $true
  }

  # Add to PATH
  $newPath = if ($currentPath) { "$currentPath;$Directory" } else { $Directory }

  try {
    [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
    Write-Success "Added to PATH: $Directory"
    return $true
  }
  catch {
    Write-Error "Failed to update PATH: $($_.Exception.Message)"
    return $false
  }
}

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

  Write-Info "[$Component] Fetching latest release information..."
  $releaseInfo = Invoke-RestMethod -Uri $ApiUrl -ErrorAction Stop
  $version = $releaseInfo.tag_name
  Write-Success "[$Component] Latest version: $version"

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

  Write-Info "[$Component] Downloading: $fileName"
  Write-Debug "[$Component] From: $DownloadUrl"

  # Download with progress
  Get-FileFromWeb -URL $DownloadUrl -File $downloadPath

  Write-Success "[$Component] Downloaded: $([math]::Round((Get-Item $downloadPath).Length / 1MB, 2)) MB"

  # Create installation directory
  if (Test-Path $InstallPath) {
    Write-Debug "[$Component] Removing existing installation at $InstallPath..."
    Remove-Item $InstallPath -Recurse -Force
  }

  New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
  Write-Debug "[$Component] Created installation directory: $InstallPath"

  # Extract zip file
  Write-Debug "[$Component] Extracting archive $fileName..."
  Expand-Archive -Path $downloadPath -DestinationPath $TempDir -Force

  # Delete ZIP if requested
  if ($DeleteZipAfterExtraction -and (Test-Path $downloadPath)) {
    Remove-Item $downloadPath -Force
    Write-Debug "[$Component] Removed downloaded ZIP $fileName from temp directory."
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
  Copy-Item "$sourcePath\\*" $InstallPath -Recurse -Force
  Write-Success "[$Component] Installed files to: $InstallPath"

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
      Write-Error "Download breaks with error: $ExeptionMsg"
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

function Send-AnalyticsEvent {
  param(
    [Parameter(Mandatory)]
    [string]$EventType
  )

  $apiUrl = "http://localhost:3000/api/analytics/event"

  # Generate anonymous user ID
  $machineId = "$env:COMPUTERNAME-$env:USERNAME"
  $hasher = [System.Security.Cryptography.MD5]::Create()
  $hashBytes = $hasher.ComputeHash([System.Text.Encoding]::UTF8.GetBytes($machineId))
  $anonymousUserId = [System.BitConverter]::ToString($hashBytes).Replace("-", "").Substring(0, 16)

  # Generate session ID
  $sessionId = [System.Guid]::NewGuid().ToString()

  $analyticsData = @{
    event      = $EventType
    properties = @{
      os             = "windows"
      os_version     = (Get-CimInstance Win32_OperatingSystem).Caption
      arch           = $env:PROCESSOR_ARCHITECTURE
      session_id     = $sessionId
      user_id        = $anonymousUserId.ToLower()
      timestamp      = $(Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ")
      cli_version    = "v1.0.0"
      install_method = "powershell"
    }
  } | ConvertTo-Json

  # Send asynchronously (non-blocking) with timeout and auto-cleanup

  Write-Debug "[ANALYTICS] Sending event '$EventType' asynchronously..."
  Write-Debug "[ANALYTICS] API URL: $apiUrl"
  Write-Debug "[ANALYTICS] Data: $($analyticsData | ConvertTo-Json -Depth 100)"

  $job = Start-Job -ScriptBlock {
    param($url, $data)
    try {
      Invoke-RestMethod -Uri $url -Method Post -Body $data -ContentType "application/json" -TimeoutSec 3 -ErrorAction SilentlyContinue
    }
    catch {

    }
  } -ArgumentList $apiUrl, $analyticsData


  Register-ObjectEvent -InputObject $job -EventName StateChanged -Action {
    if ($Event.Sender.State -eq "Completed" -or $Event.Sender.State -eq "Failed") {
      Remove-Job $Event.Sender -Force -ErrorAction SilentlyContinue
      Unregister-Event $Event.SourceIdentifier -ErrorAction SilentlyContinue
    }
  } | Out-Null

  Write-Debug "[ANALYTICS] Event '$EventType' sent asynchronously (Job ID: $($job.Id))"
}