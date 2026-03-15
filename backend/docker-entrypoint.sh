#!/bin/bash
# RechargeMax Backend Entrypoint - Simplified for Render

echo "🚀 Starting RechargeMax..."

# Run migrations if DATABASE_URL is set and psql is available
if [ -n "${DATABASE_URL:-}" ] && command -v psql >/dev/null 2>&1; then
  echo "▶ Running database migrations..."

  # Base schema files
  if [ -d "/app/database" ]; then
    for f in $(ls /app/database/*.sql 2>/dev/null | sort); do
      [ -f "$f" ] || continue
      echo "  → $(basename $f)"
      psql "$DATABASE_URL" -v ON_ERROR_STOP=0 -f "$f" 2>&1 | head -3 || true
    done
  fi

  # Versioned migrations
  if [ -d "/app/database/migrations" ]; then
    for f in $(ls /app/database/migrations/*.sql 2>/dev/null | sort); do
      [ -f "$f" ] || continue
      echo "  → $(basename $f)"
      psql "$DATABASE_URL" -v ON_ERROR_STOP=0 -f "$f" 2>&1 | head -3 || true
    done
  fi

  echo "✓ Migrations done"
fi

exec ./rechargemax
