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

## First setup

1. Provision the ignored model assets. See [`docs/models.md`](docs/models.md).
2. Create a local `.env` file with deployment-specific camera credentials, for example:

   ```dotenv
   CAM1_RTSP_URL=rtsp://username:password@camera-host:8554/Streaming/Channels/101
   ```

3. Review [`config/cameras.yaml`](config/cameras.yaml) and add additional camera variables as needed.
4. Start the full stack:

   ```bash
   docker compose up -d --build
   ```

5. Open the dashboard at `http://localhost:3000`.

The API is exposed at `http://localhost:3001`, go2rtc at `http://localhost:1984`, and PostgreSQL at host port `5433`.

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
