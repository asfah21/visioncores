# Exports PostgreSQL cameras (source of truth) to generated files.
# Usage: $env:BACKEND_URL="http://localhost:3001"; $env:API_TOKEN="<jwt>"; .\scripts\export-cameras-from-db.ps1
param(
  [string]$BackendUrl = $env:BACKEND_URL,
  [string]$Token = $env:API_TOKEN
)
$ErrorActionPreference = "Stop"
if (-not $BackendUrl) { $BackendUrl = "http://localhost:3001" }
if (-not $Token) { throw "Set API_TOKEN (vc_session JWT) in env." }

$headers = @{ Authorization = "Bearer $Token" }
$cams = Invoke-RestMethod -Uri "$BackendUrl/api/cameras" -Headers $headers
if (-not $cams) { throw "No cameras returned from $BackendUrl/api/cameras" }

# 1. media/go2rtc.yaml (backend also syncs this on boot via /api/go2rtc/sync)
$yaml = @("# GENERATED FILE - DO NOT EDIT. Source: PostgreSQL cameras", "streams:")
foreach ($c in $cams) { $yaml += "  $($c.id): `"$($c.host)`"" }
$yaml += "", "api:", '  listen: ":1984"', "", "rtsp:", '  listen: ":8554"', "  protocols:", "    - tcp  # Force TCP"
$yaml -join "`n" | Set-Content -Encoding utf8 (Join-Path $PSScriptRoot "..\media\go2rtc.yaml")
Write-Host "Exported $($cams.Count) camera(s). NOTE: go2rtc needs full RTSP urls — prefer POST $BackendUrl/api/go2rtc/sync (backend writes them from DB)."
