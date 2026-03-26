#!/usr/bin/env bash
set -euo pipefail

CONTAINER_NAME="${CONTAINER_NAME:-transfer-db}"
DB_NAME="${POSTGRES_DB:-transfer}"
DB_USER="${POSTGRES_USER:-postgres}"

if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
  echo "Container ${CONTAINER_NAME} is not running."
  echo "Run: docker start ${CONTAINER_NAME}"
  exit 1
fi

echo "Resetting database '${DB_NAME}' in container '${CONTAINER_NAME}'..."
docker exec -i "${CONTAINER_NAME}" psql -U "${DB_USER}" -d "${DB_NAME}" -v ON_ERROR_STOP=1 <<'SQL'
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO public;
SQL

echo "Database reset complete."
