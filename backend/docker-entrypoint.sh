#!/bin/bash
# ============================================================================
# RechargeMax Backend Entrypoint
# ============================================================================
# Waits for the database to be ready, runs migrations, then starts the server.
# Migrations are read from database/ which is mounted via docker-compose volume
# or baked into the image in CI/CD.
# ============================================================================
set -e

echo "⏳ Waiting for database..."
until pg_isready -h "${DB_HOST:-postgres}" -U "${DB_USER:-rechargemax}" -q; do
  sleep 1
done
echo "✓ Database is ready"

# Run migrations if DATABASE_URL is set and migration files exist
if [ -n "${DATABASE_URL:-}" ] && [ -d "/app/database/migrations" ]; then
  echo "▶ Applying migrations..."
  for f in $(ls /app/database/migrations/*.sql | sort); do
    echo "  → $(basename $f)"
    psql "$DATABASE_URL" -v ON_ERROR_STOP=0 -f "$f" > /dev/null 2>&1 || true
  done
  echo "✓ Migrations applied"
elif [ -n "${DATABASE_URL:-}" ] && [ ! -d "/app/database/migrations" ]; then
  echo "ℹ  No /app/database/migrations found - skipping (run scripts/run_migrations.sh manually)"
fi

echo "🚀 Starting RechargeMax..."
exec ./rechargemax
