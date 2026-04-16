Param(
  [string]$SnapshotPath = "./.cache/wiki-memory-sync.json",
  [string]$IndexPath = "./wiki/index.md",
  [string]$ChecklistPath = "./wiki/architecture/implementation-checklist.md",
  [switch]$IncludeSyncScript
)

$ErrorActionPreference = "Stop"

function Add-IfExists {
  Param([string]$Path)

  if (Test-Path $Path) {
    git add -- $Path
    Write-Host "Staged: $Path"
  } else {
    Write-Host "Skipped (missing): $Path"
  }
}

Add-IfExists -Path $SnapshotPath
Add-IfExists -Path $IndexPath
Add-IfExists -Path $ChecklistPath

if ($IncludeSyncScript) {
  Add-IfExists -Path "./scripts/sync-wiki-memory.ps1"
}

Write-Host ""
Write-Host "Staged files summary:"
git diff --cached --name-only
