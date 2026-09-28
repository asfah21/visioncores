# CCTV (Camera, PTZ, Recording, Playback)

PostgreSQL `cameras` adalah **source of truth**. `config/cameras.yaml` hanya seed
sekali saat tabel kosong (backend `SeedFromYAML`, `${VAR}` di-expand dari env).

## API (semua di bawah `/api`, wajib JWT `Authorization: Bearer <vc_session>`)

Frontend browser memanggil same-origin `/api/cctv/*` (Next proxy
`src/app/api/cctv/[[...path]]/route.ts` membaca cookie `vc_session` dan
meneruskannya sebagai Bearer ke `http://backend:3001`). Jangan panggil
backend langsung dari browser.

```text
GET    /api/cameras
GET    /api/cameras/:id
POST   /api/cameras                    {id,name,host,rtsp_port,rtsp_url,username,password,onvif_port,enabled,auto_record}
PUT    /api/cameras/:id                (password kosong = pertahankan lama)
DELETE /api/cameras/:id
POST   /api/cameras/:id/test-connection
GET    /api/cameras/:id/stream-url     {webrtc,mse,hls}
GET    /api/cameras/:id/info           {device,profiles,capabilities}
POST   /api/cameras/:id/ptz/move       {pan,tilt,zoom,duration}  // ±1, 100..5000ms, auto-stop
POST   /api/cameras/:id/ptz/stop
GET    /api/cameras/:id/ptz/status
POST   /api/cameras/:id/recordings/start
POST   /api/cameras/:id/recordings/stop
GET    /api/cameras/:id/recordings?date=YYYY-MM-DD&start=&end=&limit=
GET    /api/cameras/:id/recordings/timeline?date=YYYY-MM-DD
GET    /api/recordings/:id
GET    /api/recordings/:id/stream?file=<cam>/YYYY/MM/DD/HH/mm-ss.mp4   (Range)
GET    /api/recordings/:id/events      (detections AI dalam window rekaman)
POST   /api/go2rtc/sync                (tulis ulang media/go2rtc.yaml dari DB)
GET    /api/health
```

Legacy tetap tanpa auth: `POST /detection` (opsional `X-Internal-Token`),
`GET /count|/heatmap|/daily?camera_id=`, `GET /ws`.

## Keamanan

- `rtsp_url` full hanya di DB + memori backend (ffmpeg/go2rtc). API hanya
  kirim redacted. Password terenkripsi AES-GCM (`CREDENTIAL_KEY`), tidak
  pernah di-log (`SanitizeRTSPForLog`).
- PTZ selalu timeout (dial 8-10 dtk, move maks 5 dtk + auto-Stop).
- Playback stream validasi `filepath.Rel` terhadap `RECORDING_PATH`, hanya `.mp4`.

## Recording

- `ffmpeg -rtsp_transport tcp -i <rtsp> -c:v copy -an -f segment
  -segment_time 300 -strftime 1 $RECORDING_PATH/<id>/%Y/%m/%d/%H/%M-%S.mp4`
- Satu baris `recordings` per sesi, ukuran dihitung saat ffmpeg exit.
- Janitor harian hapus file + baris `> RECORDING_RETENTION_DAYS` (default 7).
- `auto_record=true` di-resume saat backend boot.

## ONVIF

`github.com/use-go/onvif` hanya diimport di `backend/internal/onvif/goclient.go`.
Semua kode lain memakai `internal/onvif.OnvifClient` + `internal/ptz.Controller`.
Tested against Bardi/PPS Speed 5T (ONVIF :8000, RTSP :8554
`/Streaming/Channels/101`).

## Env

Lihat `.env.example`: `JWT_SECRET` (wajib sama frontend+backend),
`INTERNAL_API_TOKEN`, `CREDENTIAL_KEY`, `RECORDING_RETENTION_DAYS=7`.

Wajib di root `.env`: `CAM1_RTSP_URL` (dst. `CAM2_RTSP_URL`, …). File ini
dibaca via `env_file` oleh service `backend` (seed DB + sync go2rtc) dan
`media` (ekspansi `${CAM1_RTSP_URL}` di `go2rtc.yaml`). Tanpa ini go2rtc
error `unsupported scheme: ${CAM1_RTSP_URL}` dan seed kamera dilewati.

## Troubleshooting

- `mse: streams: unsupported scheme: ${CAM1_RTSP_URL}` → `.env` belum ada /
  belum terbaca container. Pastikan root `.env` berisi `CAM1_RTSP_URL=...`,
  lalu `docker compose up -d --force-recreate media backend`.
- Playback kosong + video tidak bisa diputar → pastikan recording pernah
  jalan (`Start Recording` di detail kamera) dan backend sehat
  (`GET /api/health` via `/api/cctv/health`, 502 = backend down).

## Halaman frontend

- `/dashboard/cctv` — grid live 2x2 (iframe go2rtc, fallback generated-cameras).
- `/dashboard/cctv/[id]` — Live + PTZ pad + playback terbaru + timeline.
- `/dashboard/cctv/playback` — filter kamera/tanggal, video + timeline 24 jam + event AI.
- `/dashboard/cctv/cameras` — CRUD kamera + test connection.
