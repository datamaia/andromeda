#!/usr/bin/env pwsh
<#
.SYNOPSIS
  Andromeda installer for Windows.

  irm https://raw.githubusercontent.com/datamaia/andromeda/main/scripts/install.ps1 | iex

  Downloads the release archive for your architecture from GitHub Releases, verifies its SHA256
  against the release checksums, and installs andromeda.exe onto your PATH. Configuration via
  environment variables:

    ANDROMEDA_VERSION      version tag to install (default: latest)
    ANDROMEDA_INSTALL_DIR  install directory (default: %LOCALAPPDATA%\Programs\andromeda)
    GITHUB_TOKEN           token for a private repo's releases (otherwise anonymous)
#>
$ErrorActionPreference = 'Stop'

$repo = 'datamaia/andromeda'
$bin  = 'andromeda'

function Info($msg) { Write-Host "> $msg" -ForegroundColor Magenta }
function Die($msg)  { Write-Host "x $msg" -ForegroundColor Red; exit 1 }

# --- detect architecture ---------------------------------------------------
switch ($env:PROCESSOR_ARCHITECTURE) {
  'AMD64' { $arch = 'amd64' }
  'ARM64' { $arch = 'arm64' }
  default { Die "unsupported architecture: $($env:PROCESSOR_ARCHITECTURE) (andromeda ships amd64 and arm64)" }
}
Info "platform: windows/$arch"

# --- request headers (User-Agent required by the GitHub API; auth for private repos) -------
$headers = @{ 'User-Agent' = 'andromeda-installer' }
if ($env:GITHUB_TOKEN) { $headers['Authorization'] = "Bearer $($env:GITHUB_TOKEN)" }

# --- resolve version -------------------------------------------------------
$version = if ($env:ANDROMEDA_VERSION) { $env:ANDROMEDA_VERSION } else { 'latest' }
if ($version -eq 'latest') {
  Info 'resolving latest release...'
  try {
    $rel = Invoke-RestMethod -UseBasicParsing -Headers $headers -Uri "https://api.github.com/repos/$repo/releases/latest"
    $version = $rel.tag_name
  } catch { Die "could not resolve the latest release: $($_.Exception.Message)" }
}
if (-not $version) { Die "no releases found for $repo" }
Info "version: $version"

# --- download + verify + extract -------------------------------------------
$verNoV = $version.TrimStart('v')
$asset  = "${bin}_${verNoV}_windows_${arch}.zip"
$base   = "https://github.com/$repo/releases/download/$version"
$tmp    = Join-Path ([System.IO.Path]::GetTempPath()) ("andromeda-" + [System.Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $tmp -Force | Out-Null
try {
  $zip = Join-Path $tmp $asset
  Info "downloading $asset..."
  try {
    Invoke-WebRequest -UseBasicParsing -Headers $headers -Uri "$base/$asset" -OutFile $zip
  } catch { Die "download failed: $base/$asset" }

  # Verify the archive's SHA256 against the release checksums (best-effort; skipped if absent).
  try {
    $sums = (Invoke-WebRequest -UseBasicParsing -Headers $headers -Uri "$base/checksums.txt").Content
    $line = $sums -split "`n" | Where-Object { $_ -match [regex]::Escape($asset) } | Select-Object -First 1
    $want = ($line -split '\s+' | Select-Object -First 1)
    if ($want) {
      $got = (Get-FileHash -Algorithm SHA256 -Path $zip).Hash.ToLower()
      if ($got -ne $want.ToLower()) { Die "checksum mismatch for $asset (got $got, want $want)" }
      Info 'checksum verified'
    }
  } catch { Info "checksum verification skipped: $($_.Exception.Message)" }

  Expand-Archive -Path $zip -DestinationPath $tmp -Force
  $exe = Join-Path $tmp "$bin.exe"
  if (-not (Test-Path $exe)) { Die "archive did not contain $bin.exe" }

  # --- install onto PATH ---------------------------------------------------
  $dir = if ($env:ANDROMEDA_INSTALL_DIR) { $env:ANDROMEDA_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'Programs\andromeda' }
  New-Item -ItemType Directory -Path $dir -Force | Out-Null
  Copy-Item -Path $exe -Destination (Join-Path $dir "$bin.exe") -Force
  Info "installed $bin -> $dir\$bin.exe"

  # Add the install dir to the user PATH if it is not already there.
  $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
  if (-not $userPath) { $userPath = '' }
  if (($userPath -split ';') -notcontains $dir) {
    [Environment]::SetEnvironmentVariable('Path', ($userPath.TrimEnd(';') + ';' + $dir), 'User')
    Info "added $dir to your user PATH (restart your shell to pick it up)"
  }
  $env:Path = "$env:Path;$dir"
  Info ('done: ' + (& (Join-Path $dir "$bin.exe") version 2>$null))
} finally {
  Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
