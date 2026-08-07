# VisionCore

VisionCore is a single-repository intelligent CCTV platform containing the web dashboard, backend API, real-time media gateway, and YOLO/InsightFace detection workers.

## Repository layout

```text
visioncore/
├── backend/       # Go/Fiber API and WebSocket server
├── frontend/      # Next.js dashboard
├── yolo/          # Python YOLO + InsightFace workers
├── media/         # go2rtc configuration
├── config/        # Camera configuration
├── scripts/       # Project automation and model provisioning
├── docs/          # Operational documentation
└── docker-compose.yml
```

## Prerequisites

- Docker Desktop or Docker Engine with Compose
- PowerShell 7+ for the model provisioning helper on Windows
- Access to the configured RTSP cameras and PostgreSQL storage

## Deployment

### Requirements

- Docker Engine or Docker Desktop with Compose v2.
- Network access from the host to the RTSP cameras.
- A PostgreSQL data volume with enough disk space.

### 1. Configure the repository

Create a root `.env` file. Do not commit it because it contains camera credentials:

   ```dotenv
   CAM1_RTSP_URL=rtsp://username:password@camera-host:8554/Streaming/Channels/101
   ```

The variable name must match the camera identifier in [`config/cameras.yaml`](config/cameras.yaml). Add variables such as `CAM2_RTSP_URL` for additional cameras.

### 2. Provision YOLO and InsightFace models

Model files are intentionally excluded from Git. Because Compose mounts the host [`yolo/`](yolo) directory into the worker, provision these files before starting:

```text
yolo/yolov8n.pt
yolo/.insightface/
```

On Windows, run the included provisioning script from the repository root:

```powershell
./scripts/fetch-models.ps1
```

On Linux, download `yolov8n.pt` using the URL in [`scripts/fetch-models.ps1`](scripts/fetch-models.ps1), place it in `yolo/`, and create the InsightFace cache directory:

```bash
mkdir -p yolo/.insightface
```

InsightFace may download additional data during the first worker startup. The Compose mount persists that cache in `yolo/.insightface`.

### 3. Build and start the stack

   ```bash
   docker compose up -d --build
   ```

docker compose build
docker compose up -d
```

Or use `docker compose up -d --build`.

Check service status and YOLO logs with:

```bash
docker compose ps
docker compose logs -f yolo_cam1
```

Open the dashboard at `http://localhost:3000`.

The API is exposed at `http://localhost:3001`, go2rtc at `http://localhost:1984`, and PostgreSQL at host port `5433`.

### Service endpoints

| Service | URL/port |
| --- | --- |
| Frontend dashboard | `http://HOST:3000` |
| Backend API | `http://HOST:3001` |
| go2rtc | `http://HOST:1984` |
| PostgreSQL | host port `5433` |

Replace `HOST` with the deployment server address when accessing from another machine.

### Updates

Pull the new source and rebuild the affected services:

```bash
git pull
docker compose up -d --build
```

The PostgreSQL data is stored in the named `pgdata` volume. Do not remove volumes during routine updates. Back up PostgreSQL and the host-side `yolo/` model cache before destructive maintenance.

### Troubleshooting

If a model is not found, verify the three host paths listed above, then run:

```bash
docker compose build yolo_cam1
docker compose up -d yolo_cam1
```

If Compose reports `Set CAM1_RTSP_URL in .env`, add that variable to the root `.env`. URL-encode passwords containing characters such as `@`, `#`, or `:`.

After editing [`config/cameras.yaml`](config/cameras.yaml), regenerate the camera services with [`scripts/generate-cameras.ps1`](scripts/generate-cameras.ps1) and validate with `docker compose config`.

### Security checklist

- Keep `.env`, camera credentials, face photos, face embeddings, and inference output outside Git.
- Change the sample PostgreSQL credentials before exposing the stack outside a trusted network.
- Restrict public access to ports `3001`, `1984`, `8554`, `8555`, and `5433`.
- Use HTTPS and authentication at the edge for untrusted networks.

## Frontend development

```bash
cd frontend
npm install
npm run dev
```

## Camera configuration

Camera definitions live in [`config/cameras.yaml`](config/cameras.yaml). After changing them, regenerate the frontend camera configuration with:

```powershell
cd frontend
npm run generate:cameras
```

Keep production camera passwords out of tracked files. The root `.gitignore` excludes `.env` files.

## Repository policy

This repository uses one clean Git history for all services. Runtime model assets, local environment files, generated build output, dependency directories, and backups are excluded by the root `.gitignore`.
