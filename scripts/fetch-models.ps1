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

# OpenVINO model directory used by deployments that select the OpenVINO backend.
# The archive URL is intentionally configurable because model releases may change.
$openVinoUrl = if ($env:VISIONCORE_OPENVINO_MODEL_URL) {
    $env:VISIONCORE_OPENVINO_MODEL_URL
} else {
    "https://github.com/ultralytics/assets/releases/download/v8.3.0/yolo11n_openvino_model.zip"
}

$openVinoZip = Join-Path $env:TEMP "visioncore-yolo11n-openvino.zip"
$openVinoDirectory = Join-Path $YoloDirectory "yolo11n_openvino_model"
if (Test-Path -LiteralPath $openVinoDirectory) {
    Write-Host "Already present: $openVinoDirectory"
} else {
    Write-Host "Downloading $openVinoUrl"
    Invoke-WebRequest -Uri $openVinoUrl -OutFile $openVinoZip
    Expand-Archive -LiteralPath $openVinoZip -DestinationPath $YoloDirectory -Force
    Remove-Item -LiteralPath $openVinoZip -Force
}

# InsightFace downloads model packs on first use. The directory is created here so
# Docker volume permissions and deployment checks are deterministic.
$insightFaceDirectory = Join-Path $YoloDirectory ".insightface"
New-Item -ItemType Directory -Force -Path $insightFaceDirectory | Out-Null

Write-Host "Model directories are ready under $YoloDirectory."
Write-Host "InsightFace model packs may download automatically on first container start."
