#!/bin/bash
# ============================================================================
# RechargeMax Fresh Database Initialiser
# ============================================================================
# Bootstraps a BRAND NEW empty database:
#   1. Applies schema.sql (all CREATE TABLE / functions / triggers)
#   2. Applies incremental migrations (ALTER TABLE / fixes)
#   3. Optionally applies seeds
#
# Usage:
#   ./scripts/init_fresh_db.sh <DATABASE_URL> [--with-seeds] [--with-test-data]
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
  echo "❌ No database URL provided."
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DB_DIR="$SCRIPT_DIR/../database"

echo "🚀 Initialising fresh database..."
echo "🔗 $(echo "$DB_URL" | sed 's|://.*@|://***@|')"
echo ""

# 1. Schema
echo "📋 Step 1: Applying schema..."
psql "$DB_URL" -v ON_ERROR_STOP=1 -f "$DB_DIR/schema.sql"
echo "✓ Schema applied"
echo ""

# 2. Migrations
echo "🔄 Step 2: Applying incremental migrations..."
bash "$SCRIPT_DIR/run_migrations.sh" "$DB_URL"
echo ""

# 3. Seeds
if [ "$WITH_SEEDS" = true ]; then
  echo "🌱 Step 3: Applying reference seeds..."
  for f in \
    "$DB_DIR/seeds/004_reference_data.sql" \
    "$DB_DIR/seeds/005_notification_templates.sql" \
    "$DB_DIR/seeds/006_platform_settings.sql"
  do
    echo "  → $(basename $f)"
    psql "$DB_URL" -v ON_ERROR_STOP=0 -f "$f" || true
  done
  echo "✓ Seeds applied"
fi

if [ "$WITH_TEST" = true ]; then
  echo "🧪 Applying test data..."
  psql "$DB_URL" -v ON_ERROR_STOP=0 -f "$DB_DIR/seeds/002_test_data.sql" || true
  psql "$DB_URL" -v ON_ERROR_STOP=0 -f "$DB_DIR/seeds/003_test_numbers.sql" || true
  echo "✓ Test data applied"
fi

echo ""
echo "✅ Database initialised successfully"
