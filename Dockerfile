# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

# Cache-bust ARG: increment to force full rebuild when Render artifact cache is stale
ARG CACHE_BUST=20260323-v6

RUN apk add --no-cache git

WORKDIR /app

# Copy go module files from backend/
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source
COPY backend/ .

# Emit build timestamp to invalidate Render's go build artifact cache
RUN echo "Build timestamp: $(date -u) | Cache: ${CACHE_BUST}" && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o rechargemax ./cmd/server

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM alpine:latest

RUN apk --no-cache add ca-certificates wget postgresql-client bash

RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Binary
COPY --from=builder /app/rechargemax .

# Entrypoint
COPY backend/docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

# Migrations (from repo root /database AND backend/database)
COPY database/ /app/database/

RUN mkdir -p /app/logs && chown -R appuser:appuser /app

USER appuser
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=120s --retries=5 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./docker-entrypoint.sh"]
