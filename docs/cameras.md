# Manajemen Kamera

Dokumen ini menjelaskan cara menambah, mengubah, dan menghapus kamera pada VisionCore.

## Konsep utama

Semua konfigurasi kamera disimpan di satu file sumber:

[`config/cameras.yaml`](../config/cameras.yaml)

File tersebut adalah **source of truth**. Jangan mengedit file hasil generate secara manual:

- [`media/go2rtc.yaml`](../media/go2rtc.yaml)
- [`docker-compose.yml`](../docker-compose.yml)
- [`generated-cameras.ts`](../frontend/src/app/(main)/dashboard/default/_components/generated-cameras.ts)

Ketiga file tersebut dibuat oleh generator [`generate-cameras.ps1`](../scripts/generate-cameras.ps1).

## Format konfigurasi

Setiap kamera harus memiliki empat properti:

```yaml
cameras:
  - id: cam4
    label: Cam 4
    description: Lobby Utama
    rtsp_url: "rtsp://<username>:<password>@<camera-host>:554/Streaming/Channels/101"
```

Aturan `id`:

- Harus unik.
- Hanya boleh berisi huruf, angka, underscore (`_`), atau tanda minus (`-`).
- Sebaiknya tidak diubah setelah kamera mulai mengirim data karena `camera_id` dipakai pada histori deteksi.

## Menambah kamera

### 1. Edit source of truth

Tambahkan kamera baru pada [`config/cameras.yaml`](../config/cameras.yaml):

```yaml
cameras:
  - id: cam1
    label: Cam 1
    description: Ruang Tengah
    rtsp_url: "rtsp://<username>:<password>@<camera-host>:8554/Streaming/Channels/101"

  - id: cam4
    label: Cam 4
    description: Lobby Utama
    rtsp_url: "rtsp://<username>:<password>@<camera-host>:8554/Streaming/Channels/101"
```

Jika password RTSP memiliki karakter khusus seperti `@`, encode URL tersebut. Contoh karakter `@` menjadi `%40`.

### 2. Jalankan generator

Dari root project:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -Command "& .\scripts\generate-cameras.ps1"
```

Atau dari folder [`frontend`](../frontend/package.json):

```powershell
npm run generate:cameras
```

Generator akan memperbarui:

1. Stream `cam4` di [`media/go2rtc.yaml`](../media/go2rtc.yaml).
2. Service `yolo_cam4` di [`docker-compose.yml`](../docker-compose.yml).
3. Daftar kamera frontend di [`generated-cameras.ts`](../frontend/src/app/(main)/dashboard/default/_components/generated-cameras.ts).

### 3. Validasi konfigurasi

Periksa konfigurasi Docker Compose:

```powershell
docker compose config
```

Periksa frontend:

```powershell
cd frontend
npm run check
cd ..
```

### 4. Build dan jalankan service

Misalnya konfigurasi sekarang berisi `cam1` sampai `cam4`:

```powershell
docker compose up -d --build media frontend yolo_cam1 yolo_cam2 yolo_cam3 yolo_cam4
```

Verifikasi service:

```powershell
docker compose ps
docker compose logs -f yolo_cam4
```

Stream kamera baru dapat dibuka melalui:

```text
http://HOST_MEDIA:1984/stream.html?src=cam4
```

## Mengubah kamera

Edit properti kamera di [`config/cameras.yaml`](../config/cameras.yaml). Contoh mengubah nama dan URL:

```yaml
- id: cam4
  label: Cam 4 - Lobby
  description: Lobby Utama Gedung A
  rtsp_url: "rtsp://<username>:<password>@<camera-host>:554/Streaming/Channels/101"
```

Jalankan generator ulang, lalu recreate service terkait:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -Command "& .\scripts\generate-cameras.ps1"
docker compose up -d --force-recreate media frontend yolo_cam4
```

## Menghapus kamera

Hapus seluruh blok kamera dari [`config/cameras.yaml`](../config/cameras.yaml), kemudian jalankan generator:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -Command "& .\scripts\generate-cameras.ps1"
```

Jika yang dihapus adalah `cam4`, hentikan dan hapus container lama:

```powershell
docker compose rm -sf yolo_cam4
docker compose up -d --force-recreate media frontend
```

Generator hanya menghapus konfigurasi dan service kamera. Histori deteksi lama di database tidak ikut terhapus karena backend menyimpan data pada tabel `detections` berdasarkan `camera_id`.

## Troubleshooting

### Kamera tampil di dropdown tetapi stream offline

Periksa:

```powershell
docker compose logs -f media
docker compose logs -f yolo_cam4
```

Pastikan:

- IP kamera dapat dijangkau dari host atau container.
- Username dan password benar.
- Port RTSP benar.
- URL RTSP menggunakan encoding untuk karakter khusus.
- Nama `id` sama antara konfigurasi, stream, dan service YOLO.

### Service YOLO tidak ditemukan

Jalankan generator ulang lalu cek hasil Compose:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -Command "& .\scripts\generate-cameras.ps1"
docker compose config --services | Select-String yolo_
```

### Frontend tidak menampilkan kamera baru

Pastikan generator berhasil dan rebuild frontend:

```powershell
docker compose up -d --build frontend
```

## Checklist perubahan kamera

- [ ] Edit [`config/cameras.yaml`](../config/cameras.yaml).
- [ ] Jalankan generator.
- [ ] Jalankan `docker compose config`.
- [ ] Jalankan pemeriksaan frontend.
- [ ] Rebuild atau recreate service terkait.
- [ ] Periksa log media dan YOLO.
- [ ] Buka stream dari URL go2rtc.
