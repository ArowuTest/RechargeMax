#!/bin/bash
# RechargeMax Backend Entrypoint

echo "🚀 Starting RechargeMax..."
echo "  DATABASE_URL set: ${DATABASE_URL:+yes}"

# NOTE: Database migrations are handled by the Go binary on startup (embed.go).
# We do NOT run psql migrations here to avoid delaying the HTTP server start.
# Render's health check needs /health to respond quickly after container start.

exec ./rechargemax
