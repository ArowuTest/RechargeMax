#!/bin/bash
# Stage 1: Create all tables (runs on first container start only)
set -e
echo "▶ Creating tables..."
for f in $(ls /docker-entrypoint-initdb.d/tables/*.sql | sort); do
  echo "  → $(basename $f)"
  psql -v ON_ERROR_STOP=0 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -f "$f" || true
done

# Apply functions/triggers from schema.sql
echo "  → schema.sql (functions & triggers)"
psql -v ON_ERROR_STOP=0 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" \
  -f /docker-entrypoint-initdb.d/schema.sql || true

echo "✓ Tables created"
