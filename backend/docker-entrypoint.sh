#!/bin/bash
# RechargeMax Backend Entrypoint

echo "🚀 Starting RechargeMax..."
echo "  DATABASE_URL set: ${DATABASE_URL:+yes}"
echo "  DATABASE_URL length: ${#DATABASE_URL}"

# Run migrations if DATABASE_URL is set and psql is available
if [ -n "${DATABASE_URL:-}" ] && command -v psql >/dev/null 2>&1; then
  echo "▶ psql found, running database migrations..."

  # Test connectivity first
  echo "  Testing DB connection..."
  if psql "$DATABASE_URL" -c "SELECT 1" 2>&1; then
    echo "  ✓ DB connection OK"
  else
    echo "  ✗ DB connection FAILED - skipping migrations"
  fi

  # Base schema files
  if [ -d "/app/database" ]; then
    SCHEMA_COUNT=$(ls /app/database/*.sql 2>/dev/null | wc -l)
    echo "  Found $SCHEMA_COUNT base schema files"
    for f in $(ls /app/database/*.sql 2>/dev/null | sort); do
      [ -f "$f" ] || continue
      echo "  → $(basename $f)"
      psql "$DATABASE_URL" -v ON_ERROR_STOP=0 -f "$f" 2>&1 | tail -2 || true
    done
  else
    echo "  ✗ /app/database directory NOT FOUND"
  fi

  # Versioned migrations
  if [ -d "/app/database/migrations" ]; then
    MIG_COUNT=$(ls /app/database/migrations/*.sql 2>/dev/null | wc -l)
    echo "  Found $MIG_COUNT migration files"
    for f in $(ls /app/database/migrations/*.sql 2>/dev/null | sort); do
      [ -f "$f" ] || continue
      echo "  → $(basename $f)"
      psql "$DATABASE_URL" -v ON_ERROR_STOP=0 -f "$f" 2>&1 | tail -2 || true
    done
  else
    echo "  ✗ /app/database/migrations directory NOT FOUND"
  fi

  echo "✓ Migrations complete"
else
  echo "▶ Skipping migrations: DATABASE_URL=${DATABASE_URL:+set} psql=$(command -v psql 2>/dev/null || echo missing)"
fi

exec ./rechargemax
