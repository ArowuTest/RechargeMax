#!/bin/bash
# Stage 1: Apply base schema (runs on first container init only)
set -e
echo "▶ Applying base schema..."
psql -v ON_ERROR_STOP=1 \
     --username "$POSTGRES_USER" \
     --dbname   "$POSTGRES_DB" \
     -f /docker-entrypoint-initdb.d/schema.sql
echo "✓ Schema applied"
