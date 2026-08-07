param(
  [string]$ConfigPath = "config/cameras.yaml"
)

$ErrorActionPreference = "Stop"

function Read-RequiredValue([string]$Text, [string]$Name, [int]$LineNumber) {
  $pattern = '(?m)^\s+' + [regex]::Escape($Name) + ':\s*(?:"([^"]*)"|''([^'']*)''|(.+?))\s*$'
  $match = [regex]::Match($Text, $pattern)
  if (-not $match.Success) {
    throw "Missing '$Name' in camera block near line $LineNumber."
  }
  $value = if ($match.Groups[1].Success) { $match.Groups[1].Value } elseif ($match.Groups[2].Success) { $match.Groups[2].Value } else { $match.Groups[3].Value.Trim() }
  if ([string]::IsNullOrWhiteSpace($value)) { throw "'$Name' cannot be empty near line $LineNumber." }
  return $value
}

if (-not (Test-Path $ConfigPath)) { throw "Camera config not found: $ConfigPath" }
$config = Get-Content -Raw $ConfigPath
$blocks = [regex]::Matches($config, '(?ms)^\s+- id:\s*(?<id>[^\r\n]+)\r?\n(?<body>.*?)(?=^\s+- id:|\z)')
if ($blocks.Count -eq 0) { throw "No cameras found in $ConfigPath" }

$cameras = @()
$ids = @{}
foreach ($block in $blocks) {
  $id = $block.Groups['id'].Value.Trim()
  if ($id -notmatch '^[a-zA-Z0-9_-]+$') { throw "Invalid camera id '$id'. Use only letters, numbers, '_' or '-'." }
  if ($ids.ContainsKey($id)) { throw "Duplicate camera id '$id'." }
  $ids[$id] = $true
  $body = $block.Groups['body'].Value
  $cameras += [pscustomobject]@{
    id = $id
    label = Read-RequiredValue $body 'label' $block.Index
    description = Read-RequiredValue $body 'description' $block.Index
    rtsp_url = Read-RequiredValue $body 'rtsp_url' $block.Index
  }
}

$root = (Get-Location).Path
$mediaPath = Join-Path $root 'media/go2rtc.yaml'
$composePath = Join-Path $root 'docker-compose.yml'
$frontendPath = Join-Path $root 'frontend/src/app/(main)/dashboard/default/_components/generated-cameras.ts'

$yaml = @("# GENERATED FILE - DO NOT EDIT. Source: config/cameras.yaml", "streams:")
foreach ($camera in $cameras) { $yaml += "  $($camera.id): `"$($camera.rtsp_url)`"" }
$yaml += "", "api:", '  listen: ":1984"', "", "rtsp:", '  listen: ":8554"', "  protocols:", "    - tcp  # Force TCP"
$yaml -join "`n" | Set-Content -Encoding utf8 $mediaPath

$compose = Get-Content -Raw $composePath
$start = $compose.IndexOf("  # ================= YOLO CAMERAS =================")
$end = $compose.IndexOf("  # ================= BACKEND =================")
if ($start -lt 0 -or $end -lt 0 -or $end -le $start) { throw "Unable to locate YOLO camera section in docker-compose.yml" }
$cameraServices = @("  # ================= YOLO CAMERAS =================", "  # GENERATED FILE SECTION - Source: config/cameras.yaml")
foreach ($camera in $cameras) {
  $cameraServices += @(
    "  yolo_$($camera.id):",
    "    <<: *yolo-template",
    "    container_name: yolo_$($camera.id)",
    "    environment:",
    "      - CAMERA_ID=$($camera.id)",
    "      - RTSP_URL=$($camera.rtsp_url)",
    ""
  )
}
$compose = $compose.Substring(0, $start) + (($cameraServices -join "`n") + "`n`n") + $compose.Substring($end)
$compose | Set-Content -Encoding utf8 $composePath

$ts = @("// GENERATED FILE - DO NOT EDIT. Source: config/cameras.yaml", "", "export type CameraConfig = {", "  value: string;", "  label: string;", "  description: string;", "  url: string;", "};", "", "export const CAMERAS: CameraConfig[] = [")
foreach ($camera in $cameras) {
  $ts += @(
    "  {",
    "    value: `"$($camera.id)`",",
    "    label: `"$($camera.label.Replace('`', '\`'))`",",
    "    description: `"$($camera.description.Replace('`', '\`'))`",",
    "    url: `"http://10.10.11.5:1984/stream.html?src=$($camera.id)`",",
    "  },"
  )
}
$ts += "] as const;"
$ts -join "`n" | Set-Content -Encoding utf8 $frontendPath

Write-Host "Generated $($cameras.Count) camera(s): $($cameras.id -join ', ')"
