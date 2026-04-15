param(
  [switch]$RequireFullE2E,
  [string]$VariantId,
  [string]$OrgId,
  [string]$UserId,
  [int]$ApiPort
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

if (-not $PSBoundParameters.ContainsKey("VariantId")) {
  $VariantId = ""
}
if (-not $PSBoundParameters.ContainsKey("OrgId")) {
  $OrgId = "org_phase"
}
if (-not $PSBoundParameters.ContainsKey("UserId")) {
  $UserId = "phase-validator"
}
if (-not $PSBoundParameters.ContainsKey("ApiPort")) {
  $ApiPort = 18081
}

$repoRoot = Split-Path -Parent $PSScriptRoot

function Write-Section {
  param([string]$Text)
  Write-Host ""
  Write-Host "=== $Text ===" -ForegroundColor Cyan
}

function Invoke-Step {
  param(
    [string]$Name,
    [scriptblock]$Action
  )

  Write-Host "-> $Name" -ForegroundColor Yellow
  & $Action
  Write-Host "OK: $Name" -ForegroundColor Green
}

function Get-GoExe {
  $cmd = Get-Command go -ErrorAction SilentlyContinue
  if ($cmd) {
    return $cmd.Source
  }

  $fallback = "C:\Program Files\Go\bin\go.exe"
  if (Test-Path $fallback) {
    return $fallback
  }

  throw "Go executable not found on PATH or default location."
}

function Get-CargoExe {
  $cmd = Get-Command cargo -ErrorAction SilentlyContinue
  if ($cmd) {
    return $cmd.Source
  }
  throw "Cargo executable not found on PATH."
}

function Test-EnvPresent {
  param([string]$Name)
  $v = (Get-Item -Path ("Env:" + $Name) -ErrorAction SilentlyContinue).Value
  return -not [string]::IsNullOrWhiteSpace($v)
}

function New-HmacSha256Hex {
  param(
    [string]$Secret,
    [string]$Body
  )

  $hmac = [System.Security.Cryptography.HMACSHA256]::new([System.Text.Encoding]::UTF8.GetBytes($Secret))
  try {
    $bytes = [System.Text.Encoding]::UTF8.GetBytes($Body)
    $hash = $hmac.ComputeHash($bytes)
    return ([System.BitConverter]::ToString($hash)).Replace("-", "").ToLowerInvariant()
  }
  finally {
    $hmac.Dispose()
  }
}

Write-Section "Toolchain"
$goExe = Get-GoExe
$cargoExe = Get-CargoExe
Write-Host "Go: $goExe"
Write-Host "Cargo: $cargoExe"
Write-Host "Node: $((Get-Command node).Source)"

Write-Section "Static Validation"
Invoke-Step "API Go Test" {
  Set-Location (Join-Path $repoRoot "src/services/api")
  & $goExe test ./... | Out-Host
  if ($LASTEXITCODE -ne 0) {
    throw "API Go Test failed with exit code $LASTEXITCODE"
  }
}

Invoke-Step "Postprocess Cargo Check" {
  Set-Location (Join-Path $repoRoot "src/services/postprocess")
  & $cargoExe check | Out-Host
  if ($LASTEXITCODE -ne 0) {
    throw "Postprocess Cargo Check failed with exit code $LASTEXITCODE"
  }
}

Invoke-Step "Web Typecheck" {
  Set-Location (Join-Path $repoRoot "src/apps/web")
  npm run typecheck | Out-Host
  if ($LASTEXITCODE -ne 0) {
    throw "Web Typecheck failed with exit code $LASTEXITCODE"
  }
}

Write-Section "Runtime API Smoke"
$apiDir = Join-Path $repoRoot "src/services/api"
$apiBase = "http://localhost:$ApiPort"

$job = Start-Job -ScriptBlock {
  param($Dir, $Go, $Port)
  Set-Location $Dir
  $env:PORT = "$Port"
  & $Go run ./cmd/api
} -ArgumentList $apiDir, $goExe, $ApiPort

try {
  $started = $false
  for ($i = 0; $i -lt 30; $i++) {
    try {
      $healthProbe = Invoke-RestMethod -Method Get -Uri "$apiBase/api/v1/health" -TimeoutSec 2
      if ($healthProbe.status -eq "ok") {
        $started = $true
        break
      }
    } catch {
      # retry
    }
  }

  if (-not $started) {
    throw "API did not become healthy on port $ApiPort."
  }

  $headers = @{
    "X-User-Id" = $UserId
    "X-Org-Id" = $OrgId
    "X-Org-Role" = "member"
    "Content-Type" = "application/json"
  }

  Invoke-Step "Health endpoint" {
    $health = Invoke-RestMethod -Method Get -Uri "$apiBase/api/v1/health"
    if ($health.status -ne "ok") {
      throw "Unexpected health status: $($health.status)"
    }
  }

  Invoke-Step "Submit job route" {
    $submit = Invoke-RestMethod -Method Post -Uri "$apiBase/api/v1/jobs" -Headers $headers -Body '{"product_url":"https://example.com/products/phase3-week4","model":"veo3"}'
    if ([string]::IsNullOrWhiteSpace($submit.job_id)) {
      throw "Expected job_id in submit response"
    }
  }

  $canRunFull = $RequireFullE2E -and (Test-EnvPresent "DATABASE_URL") -and (Test-EnvPresent "MUX_WEBHOOK_SECRET") -and -not [string]::IsNullOrWhiteSpace($VariantId)

  if ($canRunFull) {
    Write-Section "Full Week 4 E2E"

    $muxSecret = (Get-Item -Path "Env:MUX_WEBHOOK_SECRET").Value
    $assetId = "asset_" + ([guid]::NewGuid().ToString("N").Substring(0, 16))
    $playbackId = "play_" + ([guid]::NewGuid().ToString("N").Substring(0, 16))

    $webhookPayload = @{
      type = "video.asset.ready"
      id = "evt_" + [guid]::NewGuid().ToString("N")
      attemptnum = 1
      created_at = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
      data = @{
        id = $assetId
        passthrough = $VariantId
        playback_ids = @(
          @{ id = $playbackId; policy = "signed" }
        )
      }
    } | ConvertTo-Json -Depth 8 -Compress

    $sigHex = New-HmacSha256Hex -Secret $muxSecret -Body $webhookPayload
    $webhookHeaders = @{
      "X-Mux-Signature-V2" = "mux_v2 $sigHex"
      "Content-Type" = "application/json"
    }

    Invoke-Step "Mux webhook callback" {
      $resp = Invoke-RestMethod -Method Post -Uri "$apiBase/webhooks/mux" -Headers $webhookHeaders -Body $webhookPayload
      if (-not $resp.received) {
        throw "Expected webhook response received=true"
      }
    }

    Invoke-Step "Playback URL endpoint" {
      $playback = Invoke-RestMethod -Method Get -Uri "$apiBase/api/v1/variants/$VariantId/playback-url" -Headers $headers
      if ([string]::IsNullOrWhiteSpace($playback.playback_url)) {
        throw "Expected non-empty playback_url"
      }
      if ($playback.playback_id -ne $playbackId) {
        throw "Playback ID mismatch. Expected $playbackId got $($playback.playback_id)"
      }
    }
  }
  else {
    Write-Section "Full Week 4 E2E (Skipped)"
    Write-Host "Set -RequireFullE2E and provide prerequisites to run full path:" -ForegroundColor Yellow
    Write-Host "- DATABASE_URL"
    Write-Host "- MUX_WEBHOOK_SECRET"
    Write-Host "- VariantId argument with an existing variant UUID"
  }
}
finally {
  if ($job -and $job.State -eq "Running") {
    Stop-Job $job -ErrorAction SilentlyContinue | Out-Null
  }
  if ($job) {
    Remove-Job $job -Force -ErrorAction SilentlyContinue | Out-Null
  }
}

Write-Section "Result"
Write-Host "Phase 3 Week 4 validation finished successfully." -ForegroundColor Green
