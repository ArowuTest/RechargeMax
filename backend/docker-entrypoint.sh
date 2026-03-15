#!/bin/bash
# ============================================================================
# RechargeMax Backend Entrypoint
# ============================================================================
set -e

echo "⏳ Waiting for database..."

# Parse host from DATABASE_URL for pg_isready check
# DATABASE_URL format: postgresql://user:pass@host:port/dbname
if [ -n "${DATABASE_URL:-}" ]; then
  # Extract host from DATABASE_URL
  DB_HOST_FROM_URL=$(echo "$DATABASE_URL" | sed -E 's|.*@([^:/]+).*|\1|')
  DB_PORT_FROM_URL=$(echo "$DATABASE_URL" | sed -E 's|.*:([0-9]+)/.*|\1|')
  DB_USER_FROM_URL=$(echo "$DATABASE_URL" | sed -E 's|.*://([^:]+):.*|\1|')
  DB_PORT_FROM_URL=${DB_PORT_FROM_URL:-5432}

  echo "  → Connecting to host: $DB_HOST_FROM_URL:$DB_PORT_FROM_URL"
  for i in $(seq 1 30); do
    if pg_isready -h "$DB_HOST_FROM_URL" -p "$DB_PORT_FROM_URL" -U "$DB_USER_FROM_URL" -q 2>/dev/null; then
      echo "✓ Database is ready"
      break
    fi
    echo "  waiting... ($i/30)"
    sleep 2
  done
else
  echo "⚠ DATABASE_URL not set — skipping DB wait"
fi

# Run base schema SQL files first (numbered files in /app/database/*.sql)
if [ -n "${DATABASE_URL:-}" ] && [ -d "/app/database" ]; then
  echo "▶ Applying base schema..."
  for f in $(ls /app/database/*.sql 2>/dev/null | sort); do
    echo "  → $(basename $f)"
    psql "$DATABASE_URL" -v ON_ERROR_STOP=0 -f "$f" > /dev/null 2>&1 || true
  done
  echo "✓ Base schema applied"
fi

# Run versioned migrations
if [ -n "${DATABASE_URL:-}" ] && [ -d "/app/database/migrations" ]; then
  echo "▶ Applying migrations..."
  for f in $(ls /app/database/migrations/*.sql 2>/dev/null | sort); do
    echo "  → $(basename $f)"
    psql "$DATABASE_URL" -v ON_ERROR_STOP=0 -f "$f" > /dev/null 2>&1 || true
  done
  echo "✓ Migrations applied"
fi

echo "🚀 Starting RechargeMax..."
exec ./rechargemax
