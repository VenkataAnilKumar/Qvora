Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

param(
  [switch]$RequireFullE2E
)

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

function Assert-Contains {
  param(
    [string]$Value,
    [string]$Needle,
    [string]$ErrorMessage
  )

  if ($Value -notmatch [Regex]::Escape($Needle)) {
    throw $ErrorMessage
  }
}

Write-Section "Toolchain"
$goExe = Get-GoExe
Write-Host "Go: $goExe"
Write-Host "Node: $((Get-Command node).Source)"

Write-Section "Static Validation"
Invoke-Step "Web Typecheck" {
  Set-Location $repoRoot
  npm run typecheck --workspace=web | Out-Host
}

Invoke-Step "API Go Test" {
  Set-Location (Join-Path $repoRoot "src/services/api")
  & $goExe test ./... | Out-Host
}

Invoke-Step "Worker Go Test" {
  Set-Location (Join-Path $repoRoot "src/services/worker")
  & $goExe test ./... | Out-Host
}

Write-Section "Runtime API Smoke"
$apiDir = Join-Path $repoRoot "src/services/api"
$apiPort = 18080
$apiBase = "http://localhost:$apiPort/api/v1"

$job = Start-Job -ScriptBlock {
  param($Dir, $Go, $Port)
  Set-Location $Dir
  $env:PORT = "$Port"
  & $Go run ./cmd/api
} -ArgumentList $apiDir, $goExe, $apiPort

try {
  $started = $false
  for ($i = 0; $i -lt 25; $i++) {
    try {
      $healthProbe = Invoke-RestMethod -Method Get -Uri "$apiBase/health" -TimeoutSec 2
      if ($healthProbe.status -eq "ok") {
        $started = $true
        break
      }
    } catch {
      Start-Sleep -Milliseconds 400
    }
  }

  if (-not $started) {
    throw "API did not become healthy on port $apiPort."
  }

  $headers = @{
    "X-User-Id" = "phase-validator"
    "X-Org-Id" = "org_phase"
    "X-Org-Role" = "member"
    "Content-Type" = "application/json"
  }

  Invoke-Step "Health Endpoint" {
    $health = Invoke-RestMethod -Method Get -Uri "$apiBase/health"
    if ($health.status -ne "ok") {
      throw "Unexpected health status: $($health.status)"
    }
  }

  $script:jobId = ""

  Invoke-Step "Submit/List/Get/Patch Job" {
    $submit = Invoke-RestMethod -Method Post -Uri "$apiBase/jobs" -Headers $headers -Body '{"product_url":"https://example.com/products/phase-test","model":"veo3"}'
    $script:jobId = $submit.job_id

    if ([string]::IsNullOrWhiteSpace($script:jobId)) {
      throw "Submit did not return job_id"
    }

    $list = Invoke-RestMethod -Method Get -Uri "$apiBase/jobs?limit=5" -Headers $headers
    if ($list.jobs.Count -lt 1) {
      throw "ListJobs returned no jobs"
    }

    $get1 = Invoke-RestMethod -Method Get -Uri "$apiBase/jobs/$script:jobId" -Headers $headers
    if ($get1.status -ne "queued") {
      throw "Expected queued, got $($get1.status)"
    }

    $patch = Invoke-RestMethod -Method Patch -Uri "$apiBase/jobs/$script:jobId/status" -Headers $headers -Body '{"status":"briefing"}'
    if ($patch.status -ne "briefing") {
      throw "Patch did not persist briefing status"
    }

    $get2 = Invoke-RestMethod -Method Get -Uri "$apiBase/jobs/$script:jobId" -Headers $headers
    if ($get2.status -ne "briefing") {
      throw "Get after patch did not return briefing"
    }
  }

  Invoke-Step "SSE Stream Sample" {
    $sse = curl.exe -sN --max-time 3 -H "X-User-Id: phase-validator" -H "X-Org-Id: org_phase" -H "X-Org-Role: member" "$apiBase/jobs/$script:jobId/stream"
    Assert-Contains -Value $sse -Needle "data:" -ErrorMessage "SSE stream did not emit data frames"
  }

  Invoke-Step "Guardrails" {
    try {
      Invoke-RestMethod -Method Get -Uri "$apiBase/jobs" -ErrorAction Stop | Out-Null
      throw "Expected unauthorized access to fail"
    } catch {
      if ($_.Exception.Response.StatusCode.value__ -ne 401) {
        throw "Expected 401 for unauthorized jobs list"
      }
    }

    try {
      Invoke-RestMethod -Method Post -Uri "$apiBase/jobs" -Headers $headers -Body '{"product_url":"https://example.com","model":"invalid-model"}' -ErrorAction Stop | Out-Null
      throw "Expected invalid model to fail"
    } catch {
      if ($_.Exception.Response.StatusCode.value__ -ne 400) {
        throw "Expected 400 for invalid model"
      }
    }
  }
}
finally {
  if ($job -and $job.State -eq "Running") {
    Stop-Job $job -Force | Out-Null
  }
  if ($job) {
    Remove-Job $job -Force | Out-Null
  }
}

Write-Section "Full E2E Prerequisites"
$required = @(
  "RAILWAY_REDIS_URL",
  "MODAL_SCRAPER_ENDPOINT",
  "OPENAI_API_KEY",
  "API_BASE_URL",
  "INTERNAL_API_KEY"
)

$missing = @()
foreach ($name in $required) {
  $v = (Get-Item -Path ("Env:" + $name) -ErrorAction SilentlyContinue).Value
  if ([string]::IsNullOrWhiteSpace($v)) {
    $missing += $name
  }
}

if ($missing.Count -eq 0) {
  Write-Host "All full E2E prerequisites are set in this shell." -ForegroundColor Green
} else {
  Write-Host "Missing full E2E prerequisites:" -ForegroundColor Yellow
  $missing | ForEach-Object { Write-Host "- $_" }

  if ($RequireFullE2E) {
    throw "Full E2E required but prerequisite environment variables are missing."
  }
}

Write-Section "Result"
Write-Host "Phase 1 and 2 validation command finished successfully." -ForegroundColor Green
