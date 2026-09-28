Saya memiliki project bernama **VisionCore / GSIVision**:

Repository:
`https://github.com/asfah21/visioncores.git`

Saya ingin kamu **mempelajari repository ini terlebih dahulu secara menyeluruh sebelum melakukan perubahan kode**.

## Tujuan

Saya ingin menambahkan fitur CCTV/IP Camera ke project yang sudah ada, terutama:

1. PTZ Control
2. Video Recording
3. Recording Playback
4. Camera Management
5. Tetap mempertahankan seluruh fitur existing GSIVision

**JANGAN membuat project CCTV baru.**
**JANGAN mengganti arsitektur existing jika tidak diperlukan.**
Implementasikan semuanya sebagai modul/fitur tambahan di dalam VisionCore.

---

# 1. Analisis Repository Terlebih Dahulu

Sebelum menulis kode:

* Baca struktur repository.
* Identifikasi frontend.
* Identifikasi Go backend.
* Identifikasi Python/AI service.
* Identifikasi PostgreSQL.
* Identifikasi konfigurasi Docker/Compose.
* Identifikasi penggunaan `go2rtc`.
* Identifikasi API routing yang sudah ada.
* Identifikasi authentication/authorization.
* Identifikasi model/database yang sudah ada.
* Identifikasi cara frontend berkomunikasi dengan backend.
* Identifikasi websocket/SSE jika sudah tersedia.

Jangan berasumsi tentang struktur project.

Gunakan struktur dan pola kode yang **sudah ada di repository**.

Setelah analisis, buat ringkasan singkat:

```text
Existing Architecture
Frontend:
Backend:
AI:
Database:
Media:
Authentication:
Current Camera/Stream:
```

Kemudian jelaskan file mana yang akan ditambah dan file mana yang akan dimodifikasi.

---

# 2. Prinsip Implementasi

Ikuti prinsip:

* Minimal modification.
* Reuse existing code.
* Reuse existing architecture.
* Jangan melakukan refactor besar tanpa alasan.
* Jangan menghapus fitur existing.
* Jangan mengganti framework.
* Jangan menambahkan dependency yang sebenarnya tidak diperlukan.
* Jangan menggunakan Python untuk kontrol PTZ jika bisa dilakukan langsung dari Go.
* Jangan menyimpan file video di PostgreSQL.
* PostgreSQL hanya menyimpan metadata recording.
* Jangan hardcode credential kamera.
* Credential kamera harus disimpan secara aman sesuai mekanisme konfigurasi existing.

Gunakan Go untuk:

* Camera API
* ONVIF
* PTZ
* Recording orchestration
* Camera management
* Playback metadata
* WebSocket/API

Gunakan Python yang sudah ada hanya untuk AI/ML seperti:

* YOLO
* InsightFace
* OCR
* processing lainnya

---

# 3. Camera Management

Tambahkan konsep camera/IP camera jika belum ada.

Minimal informasi:

```text
id
name
description
host
rtsp_port
rtsp_url
username
password
onvif_port
enabled
created_at
updated_at
```

Sesuaikan dengan model database existing.

Jangan membuat duplicate model jika repository sudah memiliki camera model.

API minimal:

```text
GET    /api/cameras
GET    /api/cameras/:id
POST   /api/cameras
PUT    /api/cameras/:id
DELETE /api/cameras/:id
```

Tambahkan:

```text
Test Connection
Get Camera Information
```

jika sesuai dengan architecture existing.

---

# 4. RTSP / go2rtc

Project sudah menggunakan atau berpotensi menggunakan `go2rtc`.

Pelajari konfigurasi `go2rtc` yang sudah ada.

Jangan membuat media pipeline kedua jika `go2rtc` existing sudah bisa digunakan.

Target:

```text
IP Camera
    │
    │ RTSP
    ▼
 go2rtc
    │
    ├── Live View
    ├── Recording
    └── AI processing
```

Gunakan mekanisme stream yang konsisten dengan project existing.

Frontend harus mendapatkan stream melalui mekanisme yang sesuai dengan `go2rtc` yang sudah digunakan.

---

# 5. ONVIF

Implementasikan ONVIF client menggunakan Go.

Fitur minimal:

```text
Device Information
Media Profiles
PTZ Profiles
PTZ capabilities
```

Jika kamera mendukung ONVIF:

```text
GetProfiles
GetStatus
ContinuousMove
Stop
```

Tambahkan abstraction seperti:

```go
type PTZController interface {
    Move(ctx context.Context, cameraID string, pan, tilt, zoom float64) error
    Stop(ctx context.Context, cameraID string) error
}
```

Sesuaikan dengan coding style project.

---

# 6. PTZ Control

Tambahkan API:

```text
POST /api/cameras/:id/ptz/move
POST /api/cameras/:id/ptz/stop
```

Request contoh:

```json
{
  "pan": 0.2,
  "tilt": 0,
  "zoom": 0,
  "duration": 1000
}
```

Nilai:

```text
pan:
-1 = left
 0 = stop
+1 = right

tilt:
-1 = down
 0 = stop
+1 = up

zoom:
-1 = zoom out
 0 = stop
+1 = zoom in
```

Implementasikan validasi nilai.

Jangan membiarkan PTZ command berjalan tanpa batas.

Gunakan timeout/context.

Setelah duration selesai, kirim:

```text
Stop
```

UI:

```text
        ▲
        │
    ◀───┼───▶
        │
        ▼

Zoom:
[-] [+]
```

Tambahkan:

```text
Stop
Home
Preset
```

hanya jika kamera/API existing mendukungnya.

---

# 7. Recording

Tambahkan kemampuan recording IP camera.

Jangan menyimpan video ke database.

Gunakan media pipeline yang sudah ada jika memungkinkan.

Metadata database minimal:

```text
id
camera_id
start_time
end_time
duration
file_path
file_size
format
status
created_at
```

Sesuaikan naming dengan database existing.

Recording harus bisa:

```text
Start
Stop
Automatic recording
```

Jika project sudah memiliki konfigurasi recording, reuse konfigurasi tersebut.

---

# 8. Recording Storage

Gunakan filesystem untuk video.

Contoh struktur:

```text
recordings/
    camera-id/
        2026/
            09/
                28/
                    15/
                        recording.mp4
```

Jangan hardcode path jika project memiliki configuration system.

Gunakan configuration:

```text
RECORDING_PATH
```

atau mekanisme konfigurasi existing.

Pastikan:

* filename aman
* path traversal tidak mungkin
* directory dibuat otomatis
* file tidak ditulis ke PostgreSQL
* metadata recording tetap konsisten

---

# 9. Playback

Tambahkan API:

```text
GET /api/cameras/:id/recordings
GET /api/recordings/:id
GET /api/recordings/:id/stream
```

Support filter:

```text
camera
date
start_time
end_time
```

Frontend harus mempunyai halaman:

```text
CCTV
 ├── Live
 └── Playback
```

Playback UI minimal:

```text
┌──────────────────────────────┐
│                              │
│        VIDEO PLAYER          │
│                              │
└──────────────────────────────┘

Camera: Camera 01

Date:
[ 2026-09-28 ]

Timeline:

00  03  06  09  12  15  18  21  24
────────────────────────────────────
       ███████       █████████

[▶] [⏸] [⏪] [⏩]

15:32:10
```

User dapat memilih recording berdasarkan timeline.

---

# 10. Recording Timeline

Jika memungkinkan, tampilkan recording availability pada timeline.

Contoh:

```text
00:00 ───────────────────────── 24:00

       █████
              █████████
                        ███
```

Area yang memiliki recording harus terlihat berbeda dari area kosong.

Klik area recording:

```text
15:32:10
```

maka player berpindah ke waktu tersebut.

---

# 11. Multi Camera

Dashboard CCTV harus mendukung beberapa camera.

Contoh:

```text
┌──────────────┬──────────────┐
│ Camera 01    │ Camera 02    │
│              │              │
│    LIVE      │    LIVE      │
│              │              │
├──────────────┼──────────────┤
│ Camera 03    │ Camera 04    │
│              │              │
│    LIVE      │    LIVE      │
└──────────────┴──────────────┘
```

Klik salah satu camera:

```text
Camera Detail
├── Live
├── PTZ
└── Playback
```

---

# 12. AI Integration

Jangan mengubah pipeline AI existing.

Jika camera stream bisa dipakai oleh YOLO/InsightFace, buat integration point:

```text
Camera
   │
   ▼
go2rtc
   │
   ├── Live
   ├── Recording
   │
   └── AI
        │
        ├── YOLO
        └── InsightFace
```

Event AI nantinya bisa dikaitkan dengan recording:

```text
15:32:10
Person detected

15:32:15
Face recognized

15:34:21
Vehicle detected
```

Tetapi **jangan mengimplementasikan ulang AI yang sudah ada**.

---

# 13. Security

Perhatikan:

* camera username/password
* API authentication
* authorization
* RTSP URL jangan dikirim ke frontend jika mengandung password
* jangan log password
* jangan expose credential dalam error message
* validasi semua camera ID
* validasi file path
* gunakan context timeout untuk ONVIF
* PTZ harus mempunyai timeout

Jika project sudah mempunyai encryption/secret mechanism, gunakan mekanisme tersebut.

---

# 14. Testing

Tambahkan testing minimal untuk:

### Go

```text
Camera service
ONVIF
PTZ validation
PTZ timeout
Recording metadata
Playback API
```

### Frontend

Pastikan:

```text
Camera list
Live view
PTZ controls
Playback
Timeline
```

tidak merusak halaman existing.

---

# 15. Development Order

Kerjakan bertahap.

### Phase 1

Camera model + Camera CRUD

### Phase 2

RTSP → go2rtc → Live View

### Phase 3

ONVIF

### Phase 4

PTZ

### Phase 5

Recording

### Phase 6

Playback

### Phase 7

Timeline

### Phase 8

AI event → recording integration

Jangan langsung membuat semuanya sekaligus.

Setelah setiap phase:

1. Compile
2. Run tests
3. Verify API
4. Verify frontend
5. Pastikan fitur existing tetap berjalan

---

# 16. Important Existing Test Camera

Saya memiliki IP camera Bardi/PPS PTZ yang sudah terbukti mendukung:

```text
RTSP:
port 8554

ONVIF:
port 8000

PTZ:
Pan
Tilt
Zoom
```

RTSP stream yang sudah berhasil:

```text
/Streaming/Channels/101
```

ONVIF device:

```text
Manufacturer: PPS IPC based on ONVIF
Model: Speed 5T
Firmware: ppstrong-c95-tuya2_general-5.5.2.20230621
```

Camera sudah terbukti:

```text
RTSP          ✅
ONVIF         ✅
Media Profile ✅
PTZ           ✅
Pan           ✅
Tilt          ✅
```

Jangan hardcode IP/password tersebut ke source code.

Gunakan configuration/environment/database sesuai architecture project.

---

# 17. Very Important

**Sebelum coding, baca repository terlebih dahulu.**

Jangan langsung membuat file baru hanya berdasarkan prompt ini.

Cari apakah project sudah memiliki:

```text
camera
stream
go2rtc
websocket
recording
storage
auth
postgres models
API router
frontend dashboard
```

Jika sudah ada, **extend existing implementation**.

Jika belum ada, baru buat module baru.

Jangan:

```text
❌ membuat project baru
❌ membuat backend kedua
❌ membuat frontend kedua
❌ mengganti Next.js
❌ mengganti Go
❌ mengganti PostgreSQL
❌ mengganti AI pipeline
❌ mengganti go2rtc tanpa alasan
❌ menggunakan Python subprocess untuk PTZ
```

Tujuan akhirnya:

```text
GSIVision
│
├── Existing Features
│   ├── Attendance
│   ├── Face Recognition
│   ├── YOLO
│   └── AI
│
└── CCTV
    ├── Camera Management
    ├── Live View
    ├── ONVIF
    ├── PTZ
    ├── Recording
    ├── Playback
    └── Timeline
```

**Mulai dengan audit repository. Jangan melakukan perubahan kode sebelum audit selesai.**
