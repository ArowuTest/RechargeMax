#!/bin/bash
# Stage 2: Apply reference seeds (runs on first container init only)
set -e
echo "▶ Applying reference seeds..."
for f in \
  /docker-entrypoint-initdb.d/seeds/004_reference_data.sql \
  /docker-entrypoint-initdb.d/seeds/005_notification_templates.sql \
  /docker-entrypoint-initdb.d/seeds/006_platform_settings.sql
do
  echo "  seeding: $(basename $f)"
  psql -v ON_ERROR_STOP=0 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -f "$f" || true
done
echo "✓ Seeds applied"
