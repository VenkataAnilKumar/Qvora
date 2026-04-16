Param(
  [string]$WikiRoot = "./wiki",
  [string]$OutputPath = "./.cache/wiki-memory-sync.json"
)

$ErrorActionPreference = "Stop"

function Get-SectionLines {
  Param(
    [string[]]$Lines,
    [string]$StartHeader,
    [string]$EndHeader
  )

  $start = -1
  $end = $Lines.Count

  for ($i = 0; $i -lt $Lines.Count; $i++) {
    if ($Lines[$i].Trim() -eq $StartHeader) {
      $start = $i + 1
      break
    }
  }

  if ($start -lt 0) {
    return @()
  }

  if ($EndHeader) {
    for ($j = $start; $j -lt $Lines.Count; $j++) {
      if ($Lines[$j].Trim() -eq $EndHeader) {
        $end = $j
        break
      }
    }
  }

  return $Lines[$start..($end - 1)]
}

function Parse-LastUpdated {
  Param([string[]]$IndexLines)

  foreach ($line in $IndexLines) {
    if ($line -match "Last updated:\s*([^|]+)\|\s*(.+)$") {
      return [pscustomobject]@{
        date = $Matches[1].Trim()
        note = $Matches[2].Trim()
      }
    }
  }

  return [pscustomobject]@{
    date = "unknown"
    note = "not found"
  }
}

function Parse-NotYetIngested {
  Param([string[]]$IndexLines)

  $section = Get-SectionLines -Lines $IndexLines -StartHeader "## Not Yet Ingested" -EndHeader ""
  if ($section.Count -eq 0) {
    return @()
  }

  $rows = @()
  foreach ($line in $section) {
    $trimmed = $line.Trim()
    if (-not $trimmed.StartsWith("|")) {
      continue
    }

    if ($trimmed -match "^\|\s*Source\s*\|" -or $trimmed -match "^\|---") {
      continue
    }

    $parts = $trimmed.Trim('|').Split('|') | ForEach-Object { $_.Trim() }
    if ($parts.Count -ge 2) {
      $rows += [pscustomobject]@{
        source = $parts[0]
        notes = $parts[1]
      }
    }
  }

  return $rows
}

function Parse-PhaseStatus {
  Param([string[]]$ChecklistLines)

  $statuses = @()
  foreach ($line in $ChecklistLines) {
    if ($line -match "^\|\s*Phase\s+([0-7])\s*\|\s*([^|]+)\|\s*([^|]+)\|\s*([^|]+)\|$") {
      $statuses += [pscustomobject]@{
        phase = [int]$Matches[1]
        name = $Matches[2].Trim()
        duration = $Matches[3].Trim()
        status = $Matches[4].Trim()
      }
    }
  }

  return $statuses
}

$indexPath = Join-Path $WikiRoot "index.md"
$checklistPath = Join-Path $WikiRoot "architecture/implementation-checklist.md"

if (-not (Test-Path $indexPath)) {
  throw "Wiki index not found at '$indexPath'."
}

if (-not (Test-Path $checklistPath)) {
  throw "Implementation checklist not found at '$checklistPath'."
}

$indexLines = Get-Content -Path $indexPath
$checklistLines = Get-Content -Path $checklistPath

$lastUpdated = Parse-LastUpdated -IndexLines $indexLines
$notYetIngested = Parse-NotYetIngested -IndexLines $indexLines
$phaseStatus = Parse-PhaseStatus -ChecklistLines $checklistLines

$snapshot = [pscustomobject]@{
  generatedAt = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssK")
  repo = (Get-Location).Path
  wiki = [pscustomobject]@{
    indexPath = $indexPath
    checklistPath = $checklistPath
    lastUpdated = $lastUpdated
    pendingIngest = @($notYetIngested)
    pendingCount = $notYetIngested.Count
  }
  phases = $phaseStatus
  syncHints = @(
    "Wiki is canonical for facts; memory is execution accelerator.",
    "After wiki ingest/update, refresh repo memory from this snapshot.",
    "Keep user memory for stable preferences and non-negotiables only."
  )
}

$outDir = Split-Path -Parent $OutputPath
if ($outDir -and -not (Test-Path $outDir)) {
  New-Item -Path $outDir -ItemType Directory -Force | Out-Null
}

$snapshot | ConvertTo-Json -Depth 8 | Set-Content -Path $OutputPath -Encoding UTF8

Write-Host "Generated sync snapshot at: $OutputPath"
Write-Host "Pending ingest items: $($notYetIngested.Count)"
Write-Host ""
Write-Host "Next step prompt (paste to Copilot):"
Write-Host "Sync memory from $OutputPath and update /memories/repo with latest wiki-derived deltas."
