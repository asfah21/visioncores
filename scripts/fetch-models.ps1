[CmdletBinding()]
param(
    [string]$YoloDirectory = (Join-Path $PSScriptRoot "..\yolo")
)

$ErrorActionPreference = "Stop"

function Download-FileIfMissing {
    param(
        [Parameter(Mandatory = $true)][string]$Url,
        [Parameter(Mandatory = $true)][string]$Destination
    )

    if (Test-Path -LiteralPath $Destination) {
        Write-Host "Already present: $Destination"
        return
    }

    $parent = Split-Path -Parent $Destination
    New-Item -ItemType Directory -Force -Path $parent | Out-Null
    Write-Host "Downloading $Url"
    Invoke-WebRequest -Uri $Url -OutFile $Destination
}

New-Item -ItemType Directory -Force -Path $YoloDirectory | Out-Null

# Ultralytics model used by yolo/main.py.
Download-FileIfMissing `
    -Url "https://github.com/ultralytics/assets/releases/download/v8.3.0/yolov8n.pt" `
    -Destination (Join-Path $YoloDirectory "yolov8n.pt")

# InsightFace downloads model packs on first use. The directory is created here so
# Docker volume permissions and deployment checks are deterministic.
$insightFaceDirectory = Join-Path $YoloDirectory ".insightface"
New-Item -ItemType Directory -Force -Path $insightFaceDirectory | Out-Null

Write-Host "Model directories are ready under $YoloDirectory."
Write-Host "InsightFace model packs may download automatically on first container start."
