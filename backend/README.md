# Backend (Go + Gin)

Minimal starter backend for the Transfer Service.

## Deployment Choices

For this project, deploy with Docker first (recommended now):

- `docker compose`: fastest to ship, same env between local/ECS, easy rollback.
- Local build + `scp`: simple but manual and less reproducible.
- GitHub Actions CI/CD: best long term once your release flow is stable.

## Prerequisites

- Go 1.26+
- PostgreSQL 14+ (local or Docker)

Quick local Postgres (Docker):

```bash
docker run --name transfer-postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=transfer -p 5432:5432 -d postgres:16
```

## Run

```bash
go mod tidy
cp .env.example .env
# export env vars from .env or set them in your shell
go run ./cmd/server
```

Upload directory defaults to `./uploads` and can be changed via `UPLOAD_DIR`.

## Verify

```bash
curl http://localhost:8080/healthz

curl -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"change_me"}'
```

Use the returned JWT as:

```bash
curl http://localhost:8080/api/files \
  -H "Authorization: Bearer <token>"
```

Most file/folder endpoints currently return `501 NOT_IMPLEMENTED` and are scaffolded for next steps.

## Docker Run (Local)

```bash
cp .env.docker.example .env
# edit .env values as needed
docker compose up -d --build
docker compose ps
```

Stop:

```bash
docker compose down
```

## Deploy To ECS (Current Recommended)

1. Push code to your git repo.
2. SSH into ECS and pull code.
3. In `backend/`, prepare env and start:

```bash
cp .env.docker.example .env
# set BASE_URL/JWT_SECRET/POSTGRES_PASSWORD/DEMO_PASSWORD
docker compose up -d --build
```

4. Verify:

```bash
curl http://127.0.0.1:8080/healthz
docker compose logs -f api
```
