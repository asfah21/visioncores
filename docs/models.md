# Model assets

The model files are intentionally excluded from the `visioncore` Git repository. They are large, machine-specific runtime assets and should be provisioned separately on each deployment host.

## Windows setup

From the repository root, run:

```powershell
./scripts/fetch-models.ps1
```

The script prepares:

- `yolo/yolov8n.pt`
- `yolo/.insightface/`

The script is idempotent for files and directories that already exist. InsightFace model packs may still be downloaded by the Python library the first time the detector starts.

## Linux / Docker deployment

The Docker Compose file mounts `./yolo` into the detector container. Copy or download the model assets into the host's `yolo/` directory before starting the stack:

```bash
docker compose build yolo
docker compose up -d
```

If the deployment host cannot run PowerShell, download `yolov8n.pt` using the URL in `scripts/fetch-models.ps1`, place it in `yolo/`, and create `yolo/.insightface/`.

Do not commit model binaries, face embeddings, or other sensitive inference data. The root `.gitignore` excludes the expected model paths.
