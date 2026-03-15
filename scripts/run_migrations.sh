#!/bin/bash
# ============================================================================
# RechargeMax Migration Runner
# ============================================================================
# Applies incremental migrations from database/migrations/ to an existing DB.
# For a FRESH database, run schema.sql first, then this script.
#
# Usage:
#   ./scripts/run_migrations.sh                        # uses DATABASE_URL env var
#   DATABASE_URL=postgres://... ./scripts/run_migrations.sh
#   ./scripts/run_migrations.sh postgres://user:pass@host/db
# ============================================================================

set -euo pipefail

DB_URL="${1:-${DATABASE_URL:-}}"

if [ -z "$DB_URL" ]; then
  echo "❌ No database URL provided."
  echo "   Set DATABASE_URL or pass it as first argument."
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MIGS_DIR="$SCRIPT_DIR/../database/migrations"

if [ ! -d "$MIGS_DIR" ]; then
  echo "❌ Migrations directory not found: $MIGS_DIR"
  exit 1
fi

echo "🗄  Running migrations from: $MIGS_DIR"
echo "🔗 Database: $(echo "$DB_URL" | sed 's|://.*@|://***@|')"
echo ""

ERRORS=0
COUNT=0

for f in $(ls "$MIGS_DIR"/*.sql | sort); do
  fname=$(basename "$f")
  printf "  %-55s " "$fname"
  
  result=$(psql "$DB_URL" -v ON_ERROR_STOP=0 -f "$f" 2>&1)
  
  if echo "$result" | grep -q "^ERROR:"; then
    echo "❌ FAILED"
    echo "$result" | grep "^ERROR:" | head -3 | sed 's/^/     /'
    ERRORS=$((ERRORS + 1))
  else
    echo "✓"
    COUNT=$((COUNT + 1))
  fi
done

echo ""
if [ $ERRORS -eq 0 ]; then
  echo "✅ All $COUNT migrations applied successfully"
else
  echo "⚠️  $COUNT applied, $ERRORS failed"
  exit 1
fi
