# Loopa

MVP: upload audio/video, process asynchronously, show transcript, export TXT/DOCX,
and keep a per-session history.

## Repository Layout

- `frontend/` React + Redux Toolkit UI.
- `backend/` Go API and worker.
- `infra/` Docker Compose for local dev.

## Quick Start (Local)

1. Install Docker + Docker Compose and `ffmpeg` on the host.
2. Run: `docker compose -f infra/docker-compose.yml up --build`
3. Open `http://localhost:5173`

## Env Vars

Backend/Worker:

- `DB_DSN` (default: `root:root@tcp(mysql:3306)/loopa?parseTime=true`)
- `UPLOAD_DIR` (default: `/data/uploads`)
- `MAX_UPLOAD_BYTES` (default: `1073741824`)
- `MOCK_DELAY_MS` (default: `2000`)
 
## License

TBD
