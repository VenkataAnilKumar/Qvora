param(
  [string]$SupabaseProjectRef = "",
  [string]$ApiBaseUrl = "",
  [string]$ApiAuthToken = "",
  [string]$WorkspaceId = "",
  [string]$ProductUrl = "",
  [string]$InternalApiKey = "",
  [switch]$RunStagingMigration,
  [switch]$SkipApiChecks
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$reportPath = Join-Path $repoRoot ("docs/07-implementation/PHASE0_3_FIXES_ONE_SHOT_" + (Get-Date -Format "yyyy-MM-dd_HHmmss") + ".md")

function Add-ReportLine {
  param([string]$Line)
  Add-Content -Path $reportPath -Value $Line
}

function Write-Section {
  param([string]$Text)
  Write-Host ""
  Write-Host "=== $Text ===" -ForegroundColor Cyan
  Add-ReportLine ""
  Add-ReportLine "## $Text"
}

function Invoke-Check {
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

function Assert-Command {
  param([string]$Name)
  if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
    throw "Required command not found: $Name"
  }
}

function Assert-NonEmpty {
  param(
    [string]$Value,
    [string]$Name
  )

  if ([string]::IsNullOrWhiteSpace($Value)) {
    throw "Required parameter is missing: $Name"
  }
}

Set-Location $repoRoot
"# Phase 0-3 Fixes One-Shot Validation" | Set-Content -Path $reportPath
Add-ReportLine ""
Add-ReportLine "- Timestamp: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
Add-ReportLine "- Repo: Qvora"

Write-Section "Preflight"
Invoke-Check -Id "PRE-1" -Name "PowerShell prerequisites" -Action {
  Assert-Command -Name "npx"
}

Invoke-Check -Id "PRE-2" -Name "Required env keys declared in .env.example" -Action {
  $envExample = Join-Path $repoRoot ".env.example"
  if (-not (Test-Path $envExample)) {
    throw ".env.example not found"
  }

  $requiredKeys = @(
    "HEYGEN_API_KEY",
    "TAVUS_API_KEY",
    "FAL_COST_LIMIT_STARTER",
    "FAL_COST_LIMIT_GROWTH",
    "FAL_COST_LIMIT_AGENCY",
    "INTERNAL_API_KEY",
    "API_BASE_URL",
    "ANTHROPIC_API_KEY"
  )

  $content = Get-Content $envExample -Raw
  $missing = @()
  foreach ($key in $requiredKeys) {
    if ($content -notmatch "(?m)^\s*" + [regex]::Escape($key) + "=") {
      $missing += $key
    }
  }

  if ($missing.Count -gt 0) {
    throw ("Missing keys in .env.example: " + ($missing -join ", "))
  }
}

Write-Section "Staging Migration"
Invoke-Check -Id "MIG-1" -Name "Run migration 004 on staging Supabase" -Action {
  if (-not $RunStagingMigration) {
    throw "Skipped. Re-run with -RunStagingMigration after Supabase access is restored."
  }

  Assert-NonEmpty -Value $SupabaseProjectRef -Name "SupabaseProjectRef"

  npx supabase link --project-ref $SupabaseProjectRef | Out-Host
  if ($LASTEXITCODE -ne 0) {
    throw "supabase link failed"
  }

  npx supabase db push | Out-Host
  if ($LASTEXITCODE -ne 0) {
    throw "supabase db push failed"
  }
}

if (-not $SkipApiChecks) {
  Write-Section "API Runtime Checks"

  Invoke-Check -Id "API-1" -Name "Inputs provided for runtime validation" -Action {
    Assert-NonEmpty -Value $ApiBaseUrl -Name "ApiBaseUrl"
    Assert-NonEmpty -Value $ApiAuthToken -Name "ApiAuthToken"
    Assert-NonEmpty -Value $WorkspaceId -Name "WorkspaceId"
    Assert-NonEmpty -Value $ProductUrl -Name "ProductUrl"
    Assert-NonEmpty -Value $InternalApiKey -Name "InternalApiKey"
  }

  Invoke-Check -Id "API-2" -Name "Idempotency: duplicate submit returns same job" -Action {
    $key = [guid]::NewGuid().ToString()

    $headers = @{
      "Authorization" = "Bearer $ApiAuthToken"
      "X-Idempotency-Key" = $key
      "Content-Type" = "application/json"
    }

    $body = @{
      workspace_id = $WorkspaceId
      product_url = $ProductUrl
    } | ConvertTo-Json -Depth 4 -Compress

    $resp1 = Invoke-RestMethod -Method Post -Uri "$ApiBaseUrl/api/v1/jobs" -Headers $headers -Body $body
    $resp2 = Invoke-RestMethod -Method Post -Uri "$ApiBaseUrl/api/v1/jobs" -Headers $headers -Body $body

    if (($null -eq $resp1.job_id) -or ($null -eq $resp2.job_id)) {
      throw "job_id missing in response"
    }

    if ($resp1.job_id -ne $resp2.job_id) {
      throw "idempotent submit returned different job ids"
    }
  }

  Invoke-Check -Id "API-3" -Name "Idempotency: missing header returns 400" -Action {
    $headers = @{
      "Authorization" = "Bearer $ApiAuthToken"
      "Content-Type" = "application/json"
    }

    $body = @{
      workspace_id = $WorkspaceId
      product_url = $ProductUrl
    } | ConvertTo-Json -Depth 4 -Compress

    try {
      Invoke-WebRequest -Method Post -Uri "$ApiBaseUrl/api/v1/jobs" -Headers $headers -Body $body -ErrorAction Stop | Out-Null
      throw "Expected 400 but request succeeded"
    }
    catch {
      if ($_.Exception.Response -eq $null) {
        throw
      }

      $statusCode = [int]$_.Exception.Response.StatusCode
      if ($statusCode -ne 400) {
        throw "Expected 400, got $statusCode"
      }
    }
  }

  Invoke-Check -Id "API-4" -Name "Internal perf endpoint reachable with internal key" -Action {
    $headers = @{
      "X-Internal-API-Key" = $InternalApiKey
      "Content-Type" = "application/json"
    }

    $body = @{
      workspace_id = $WorkspaceId
      variant_id = [guid]::NewGuid().ToString()
      stage = "total"
      duration_ms = 1234
      model = "manual-check"
    } | ConvertTo-Json -Depth 5 -Compress

    $resp = Invoke-RestMethod -Method Post -Uri "$ApiBaseUrl/api/v1/internal/perf-events" -Headers $headers -Body $body
    if ($null -eq $resp) {
      throw "No response from perf-events endpoint"
    }
  }

  Invoke-Check -Id "API-5" -Name "Internal cost endpoint reachable with internal key" -Action {
    $headers = @{
      "X-Internal-API-Key" = $InternalApiKey
      "Content-Type" = "application/json"
    }

    $body = @{
      workspace_id = $WorkspaceId
      source = "manual-check"
      model = "manual-check"
      estimated_usd = "0.010000"
      credits = 1
    } | ConvertTo-Json -Depth 5 -Compress

    $resp = Invoke-RestMethod -Method Post -Uri "$ApiBaseUrl/api/v1/internal/cost-events" -Headers $headers -Body $body
    if ($null -eq $resp) {
      throw "No response from cost-events endpoint"
    }
  }
}

Write-Section "Manual Checks Still Required"
Add-ReportLine "- MAN-1 | TODO | Semaphore concurrency and TTL behavior in live generation flow"
Add-ReportLine "- MAN-2 | TODO | Cost breaker threshold and hourly reset behavior"
Add-ReportLine "- MAN-3 | TODO | HeyGen success, HeyGen 429 to Tavus fallback, avatar to postprocess handoff"
Add-ReportLine "- MAN-4 | TODO | Full E2E URL to brief to video to postprocess to Mux playback"

Write-Host ""
Write-Host "Validation report written:" -ForegroundColor Green
Write-Host "$reportPath" -ForegroundColor Green
