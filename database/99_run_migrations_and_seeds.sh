#!/bin/bash
# ============================================================================
# RechargeMax Database Boot Sequence
# ============================================================================
# Runs automatically by Docker postgres on first container start.
# Also called by scripts/init_fresh_db.sh for non-Docker setups.
#
# Order:
#   1. Tables already created by 01_*.sql ... 46_*.sql (Docker runs them first)
#   2. This script: migrations (ALTER TABLE) then reference seeds
# ============================================================================
set -e

DB_USER="${POSTGRES_USER:-rechargemax}"
DB_NAME="${POSTGRES_DB:-rechargemax_db}"
INIT_DIR="$(dirname "$0")"

run_sql() {
  local file="$1"
  echo "  → $(basename $file)"
  psql -v ON_ERROR_STOP=0 --username "$DB_USER" --dbname "$DB_NAME" -f "$file" 2>&1 \
    | grep -v "^SET$\|^CREATE\|^ALTER\|^INSERT\|^DROP\|^GRANT\|^DO\|^UPDATE\|^--\|^$" \
    | grep -i "error\|warning" || true
}

echo ""
echo "▶ Running migrations..."
for f in $(ls "$INIT_DIR/migrations"/*.sql 2>/dev/null | sort); do
  run_sql "$f"
done
echo "✓ Migrations done"

echo ""
echo "▶ Seeding reference data..."
for f in \
  "$INIT_DIR/seeds/004_reference_data.sql" \
  "$INIT_DIR/seeds/005_notification_templates.sql" \
  "$INIT_DIR/seeds/006_platform_settings.sql" \
  "$INIT_DIR/seeds/007_draw_prize_config.sql"
  "$INIT_DIR/seeds/008_network_configs.sql"
do
  [ -f "$f" ] && run_sql "$f" || true
done
echo "✓ Seeds done"

echo ""
echo "✅ Database ready"
