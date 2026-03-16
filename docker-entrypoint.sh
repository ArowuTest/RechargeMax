#!/bin/bash
# ============================================================================
# RechargeMax Backend Entrypoint
# ============================================================================
# Do NOT use set -e - we want to continue even if pg_isready fails

echo "⏳ Waiting for database to be ready..."

if [ -n "${DATABASE_URL:-}" ]; then
  # Extract host from DATABASE_URL
  # URL format: postgresql://user:pass@host[:port]/dbname
  DB_HOST_FROM_URL=$(echo "$DATABASE_URL" | sed -E 's|.*@([^:/]+).*|\1|')
  
  # Extract port - default to 5432 if not found
  DB_PORT_RAW=$(echo "$DATABASE_URL" | sed -E 's|.*@[^:]+:([0-9]+)/.*|\1|')
  if echo "$DB_PORT_RAW" | grep -qE '^[0-9]+$'; then
    DB_PORT_FROM_URL="$DB_PORT_RAW"
  else
    DB_PORT_FROM_URL="5432"
  fi

  echo "  → Host: $DB_HOST_FROM_URL, Port: $DB_PORT_FROM_URL"
  
  MAX_RETRIES=30
  i=0
  while [ $i -lt $MAX_RETRIES ]; do
    if pg_isready -h "$DB_HOST_FROM_URL" -p "$DB_PORT_FROM_URL" -q 2>/dev/null; then
      echo "✓ Database is ready"
      break
    fi
    i=$((i + 1))
    echo "  waiting... ($i/$MAX_RETRIES)"
    sleep 2
  done
  
  if [ $i -ge $MAX_RETRIES ]; then
    echo "⚠ Database did not become ready in time, proceeding anyway..."
  fi
else
  echo "⚠ DATABASE_URL not set — skipping DB wait"
fi

# Run base schema SQL files
if [ -n "${DATABASE_URL:-}" ] && [ -d "/app/database" ]; then
  echo "▶ Applying base schema..."
  for f in /app/database/*.sql; do
    [ -f "$f" ] || continue
    echo "  → $(basename $f)"
    psql "$DATABASE_URL" -v ON_ERROR_STOP=0 -f "$f" 2>/dev/null || true
  done
  echo "✓ Base schema done"
fi

# Run versioned migrations
if [ -n "${DATABASE_URL:-}" ] && [ -d "/app/database/migrations" ]; then
  echo "▶ Applying migrations..."
  for f in /app/database/migrations/*.sql; do
    [ -f "$f" ] || continue
    echo "  → $(basename $f)"
    psql "$DATABASE_URL" -v ON_ERROR_STOP=0 -f "$f" 2>/dev/null || true
  done
  echo "✓ Migrations done"
fi

echo "🚀 Starting RechargeMax..."
exec ./rechargemax
