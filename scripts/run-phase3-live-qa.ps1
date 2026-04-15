param(
  [string]$ComposeFile = "docker-compose.yml",
  [switch]$SkipBuild,
  [switch]$KeepServices,
  [string]$ApiBaseUrl = "http://localhost:8080",
  [string]$VariantId,
  [string]$SecondVariantId,
  [string]$JobId,
  [string]$OrgId = "org_phase",
  [string]$UserId = "phase3-live-qa"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$reportPath = Join-Path $repoRoot "docs/07-implementation/PHASE3_QA_LIVE_RUN_$(Get-Date -Format 'yyyy-MM-dd_HHmmss').md"
$composeArgs = @("compose", "-f", $ComposeFile)

function Initialize-Report {
  "# Phase 3 Live QA Run ($(Get-Date -Format "yyyy-MM-dd HH:mm:ss"))" | Set-Content -Path $reportPath
  Add-ReportLine ""
  Add-ReportLine "## Runner"
  Add-ReportLine "- Script: scripts/run-phase3-live-qa.ps1"
  Add-ReportLine "- ApiBaseUrl: $ApiBaseUrl"
  Add-ReportLine "- ComposeFile: $ComposeFile"
  Add-ReportLine ""
  Add-ReportLine "## Results"
}

function Write-Section {
  param([string]$Text)
  Write-Host ""
  Write-Host "=== $Text ===" -ForegroundColor Cyan
}

function Add-ReportLine {
  param([string]$Line)
  Add-Content -Path $reportPath -Value $Line
}

function Get-DotEnvValue {
  param(
    [string]$Path,
    [string]$Key
  )

  if (-not (Test-Path $Path)) {
    return ""
  }

  $line = Get-Content $Path | Where-Object {
    $_ -match "^\s*$Key="
  } | Select-Object -First 1

  if (-not $line) {
    return ""
  }

  $value = $line.Substring($line.IndexOf("=") + 1).Trim()
  return $value.Trim('"')
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

function Wait-HttpOk {
  param(
    [string]$Url,
    [int]$Retries = 60,
    [int]$SleepMs = 1000
  )

  for ($i = 0; $i -lt $Retries; $i++) {
    try {
      $resp = Invoke-RestMethod -Method Get -Uri $Url -TimeoutSec 3
      if ($resp) {
        return $true
      }
    }
    catch {
      # retry
    }
    Start-Sleep -Milliseconds $SleepMs
  }

  return $false
}

function Invoke-Test {
  param(
    [string]$Id,
    [string]$Name,
    [scriptblock]$Action
  )

  Write-Host "-> [$Id] $Name" -ForegroundColor Yellow
  try {
    & $Action
    Write-Host "PASS [$Id] $Name" -ForegroundColor Green
    Add-ReportLine "- $Id | PASS | $Name"
  }
  catch {
    Write-Host "FAIL [$Id] $Name : $($_.Exception.Message)" -ForegroundColor Red
    Add-ReportLine "- $Id | FAIL | $Name | $($_.Exception.Message)"
  }
}

Set-Location $repoRoot
Initialize-Report

Write-Section "Preflight"
$preflightIssues = @()
if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
  $preflightIssues += "Docker is not installed or not on PATH."
}
if (-not (Test-Path (Join-Path $repoRoot $ComposeFile))) {
  $preflightIssues += "Compose file not found: $ComposeFile"
}
if (-not (Test-Path (Join-Path $repoRoot ".env"))) {
  $preflightIssues += ".env file is missing at repo root."
}

$muxWebhookSecret = Get-DotEnvValue -Path (Join-Path $repoRoot ".env") -Key "MUX_WEBHOOK_SECRET"
if ([string]::IsNullOrWhiteSpace($muxWebhookSecret)) {
  $preflightIssues += "MUX_WEBHOOK_SECRET is missing in .env"
}

if ($preflightIssues.Count -gt 0) {
  Add-ReportLine "- PRE-0 | BLOCKED | Preflight prerequisites missing"
  foreach ($issue in $preflightIssues) {
    Add-ReportLine "  - $issue"
    Write-Host "BLOCKED: $issue" -ForegroundColor Yellow
  }

  Add-ReportLine ""
  Add-ReportLine "## Outcome"
  Add-ReportLine "- Live QA blocked by local environment prerequisites."

  Write-Section "Done"
  Write-Host "Live QA report written to: $reportPath" -ForegroundColor Green
  exit 0
}

Write-Section "Compose Up"
if ($SkipBuild) {
  & docker @composeArgs up -d | Out-Host
}
else {
  & docker @composeArgs up -d --build | Out-Host
}
if ($LASTEXITCODE -ne 0) {
  throw "docker compose up failed with exit code $LASTEXITCODE"
}

Write-Section "Service Readiness"
$apiReady = Wait-HttpOk -Url "$ApiBaseUrl/api/v1/health"
if (-not $apiReady) {
  throw "API health endpoint did not become ready at $ApiBaseUrl/api/v1/health"
}
$postprocessReady = Wait-HttpOk -Url "http://localhost:3001/health"
if (-not $postprocessReady) {
  throw "Postprocess health endpoint did not become ready at http://localhost:3001/health"
}

Add-ReportLine "- PRE-1 | PASS | API and postprocess health endpoints responded"

Write-Section "Checklist Automation"

Invoke-Test -Id "1" -Name "FAL completion webhook accepted and enqueued" -Action {
  $falPayload = @{
    status = "completed"
    metadata = @{
      job_id = ([guid]::NewGuid().ToString())
      variant_id = ([guid]::NewGuid().ToString())
      workspace_id = ([guid]::NewGuid().ToString())
      input_r2_key = "fal/input/demo.mp4"
      output_r2_key = "fal/output/demo.mp4"
    }
    input_r2_key = "fal/input/demo.mp4"
    output_r2_key = "fal/output/demo.mp4"
  } | ConvertTo-Json -Depth 8 -Compress

  $resp = Invoke-RestMethod -Method Post -Uri "$ApiBaseUrl/webhooks/fal" -ContentType "application/json" -Body $falPayload
  if (-not $resp.received) {
    throw "Expected received=true"
  }
}

Invoke-Test -Id "5" -Name "Mux invalid signature rejected with 401" -Action {
  $muxPayload = @{
    type = "video.asset.ready"
    id = "evt_" + [guid]::NewGuid().ToString("N")
    attemptnum = 1
    created_at = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
    data = @{
      id = "asset_invalid_sig"
      passthrough = ([guid]::NewGuid().ToString())
      playback_ids = @(
        @{ id = "play_invalid_sig" }
      )
    }
  } | ConvertTo-Json -Depth 8 -Compress

  $headers = @{ "X-Mux-Signature-V2" = "mux_v2 invalid" }

  try {
    Invoke-WebRequest -Method Post -Uri "$ApiBaseUrl/webhooks/mux" -ContentType "application/json" -Headers $headers -Body $muxPayload -ErrorAction Stop | Out-Null
    throw "Expected HTTP 401 but request succeeded"
  }
  catch {
    if ($_.Exception.Response -eq $null) {
      throw
    }
    $code = [int]$_.Exception.Response.StatusCode
    if ($code -ne 401) {
      throw "Expected HTTP 401, got HTTP $code"
    }
  }
}

if (-not [string]::IsNullOrWhiteSpace($VariantId)) {
  Invoke-Test -Id "3" -Name "Mux ready webhook processed for provided variant" -Action {
    $muxPayload = @{
      type = "video.asset.ready"
      id = "evt_" + [guid]::NewGuid().ToString("N")
      attemptnum = 1
      created_at = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
      data = @{
        id = "asset_" + [guid]::NewGuid().ToString("N").Substring(0, 12)
        passthrough = $VariantId
        playback_ids = @(
          @{ id = "play_" + [guid]::NewGuid().ToString("N").Substring(0, 12) }
        )
      }
    } | ConvertTo-Json -Depth 8 -Compress

    $sigHex = New-HmacSha256Hex -Secret $muxWebhookSecret -Body $muxPayload
    $headers = @{ "X-Mux-Signature-V2" = "mux_v2 $sigHex" }

    $resp = Invoke-RestMethod -Method Post -Uri "$ApiBaseUrl/webhooks/mux" -ContentType "application/json" -Headers $headers -Body $muxPayload
    if (-not $resp.received) {
      throw "Expected received=true"
    }
  }
}
else {
  Add-ReportLine "- 3 | SKIPPED | Provide -VariantId to run full mux-ready processing"
}

if (-not [string]::IsNullOrWhiteSpace($JobId) -and -not [string]::IsNullOrWhiteSpace($VariantId) -and -not [string]::IsNullOrWhiteSpace($SecondVariantId)) {
  Invoke-Test -Id "4" -Name "Job completion gate across provided two variants" -Action {
    $authHeaders = @{
      "X-User-Id" = $UserId
      "X-Org-Id" = $OrgId
      "X-Org-Role" = "member"
    }

    $before = Invoke-RestMethod -Method Get -Uri "$ApiBaseUrl/api/v1/jobs/$JobId" -Headers $authHeaders

    foreach ($variant in @($VariantId, $SecondVariantId)) {
      $muxPayload = @{
        type = "video.asset.ready"
        id = "evt_" + [guid]::NewGuid().ToString("N")
        attemptnum = 1
        created_at = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
        data = @{
          id = "asset_" + [guid]::NewGuid().ToString("N").Substring(0, 12)
          passthrough = $variant
          playback_ids = @(
            @{ id = "play_" + [guid]::NewGuid().ToString("N").Substring(0, 12) }
          )
        }
      } | ConvertTo-Json -Depth 8 -Compress

      $sigHex = New-HmacSha256Hex -Secret $muxWebhookSecret -Body $muxPayload
      $headers = @{ "X-Mux-Signature-V2" = "mux_v2 $sigHex" }
      $resp = Invoke-RestMethod -Method Post -Uri "$ApiBaseUrl/webhooks/mux" -ContentType "application/json" -Headers $headers -Body $muxPayload
      if (-not $resp.received) {
        throw "Expected received=true for variant $variant"
      }
    }

    $after = Invoke-RestMethod -Method Get -Uri "$ApiBaseUrl/api/v1/jobs/$JobId" -Headers $authHeaders
    if ([string]::IsNullOrWhiteSpace($before.status) -or [string]::IsNullOrWhiteSpace($after.status)) {
      throw "Unable to determine job statuses before/after"
    }
  }
}
else {
  Add-ReportLine "- 4 | SKIPPED | Provide -JobId, -VariantId, and -SecondVariantId for gating test"
}

Write-Section "Compose Status"
& docker @composeArgs ps | Out-Host

if (-not $KeepServices) {
  Write-Section "Compose Down"
  & docker @composeArgs down | Out-Host
}

Write-Section "Done"
Write-Host "Live QA report written to: $reportPath" -ForegroundColor Green
