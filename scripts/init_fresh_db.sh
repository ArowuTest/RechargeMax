#!/bin/bash
# ============================================================================
# RechargeMax Fresh Database Initialiser
# ============================================================================
# Bootstraps a brand-new empty database:
#   1. Runs all table definition files (database/01_*.sql ... 46_*.sql)
#   2. Runs incremental migrations (database/migrations/)
#   3. Optionally seeds reference data
#
# Usage:
#   ./scripts/init_fresh_db.sh <DATABASE_URL> [--with-seeds] [--with-test-data]
#   DATABASE_URL=postgres://... ./scripts/init_fresh_db.sh --with-seeds
# ============================================================================
set -euo pipefail

DB_URL="${1:-${DATABASE_URL:-}}"
WITH_SEEDS=false
WITH_TEST=false

for arg in "$@"; do
  case $arg in
    --with-seeds)     WITH_SEEDS=true ;;
    --with-test-data) WITH_TEST=true  ;;
  esac
done

if [ -z "$DB_URL" ]; then
  echo "❌  No database URL. Set DATABASE_URL or pass as first argument."
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DB_DIR="$SCRIPT_DIR/../database"

echo "🚀 Initialising fresh database"
echo "🔗 $(echo "$DB_URL" | sed 's|://.*@|://***@|')"
echo ""

# ── Step 1: Create tables ────────────────────────────────────────────────────
echo "📋 Step 1: Creating tables..."
for f in $(ls "$DB_DIR"/[0-9]*.sql | sort); do
  printf "  %-50s " "$(basename $f)"
  psql "$DB_URL" -v ON_ERROR_STOP=0 -f "$f" > /dev/null 2>&1 && echo "✓" || echo "⚠ (check manually)"
done
echo ""

# ── Step 2: Apply incremental migrations ────────────────────────────────────
echo "🔄 Step 2: Applying migrations..."
bash "$SCRIPT_DIR/run_migrations.sh" "$DB_URL"
echo ""

# ── Step 3: Seeds ────────────────────────────────────────────────────────────
if [ "$WITH_SEEDS" = true ]; then
  echo "🌱 Step 3: Seeding reference data..."
  for f in \
    "$DB_DIR/seeds/004_reference_data.sql" \
    "$DB_DIR/seeds/005_notification_templates.sql" \
    "$DB_DIR/seeds/006_platform_settings.sql"
  do
    printf "  %-50s " "$(basename $f)"
    psql "$DB_URL" -v ON_ERROR_STOP=0 -f "$f" > /dev/null 2>&1 && echo "✓" || echo "⚠"
  done
fi

if [ "$WITH_TEST" = true ]; then
  echo "🧪 Applying test data..."
  psql "$DB_URL" -v ON_ERROR_STOP=0 -f "$DB_DIR/seeds/002_test_data.sql" > /dev/null 2>&1 && echo "  test_data ✓" || echo "  test_data ⚠"
  psql "$DB_URL" -v ON_ERROR_STOP=0 -f "$DB_DIR/seeds/003_test_numbers.sql" > /dev/null 2>&1 && echo "  test_numbers ✓" || echo "  test_numbers ⚠"
fi

echo ""
echo "✅ Done"
