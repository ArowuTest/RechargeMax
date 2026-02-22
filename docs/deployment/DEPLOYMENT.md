# RechargeMax Production Deployment Guide

## 🚀 Quick Start

This guide covers deploying RechargeMax to production with all champion-level fixes applied.

---

## 📋 Prerequisites

### System Requirements
- **OS:** Ubuntu 22.04 LTS or later
- **CPU:** 2+ cores (4+ recommended)
- **RAM:** 4GB minimum (8GB+ recommended)
- **Storage:** 20GB+ SSD
- **Database:** PostgreSQL 14+
- **Go:** 1.21+
- **Node.js:** 18+

### External Services
- **Payment Gateway:** Paystack (live keys)
- **SMS Provider:** Termii (live API key)
- **VTU Provider:** VTPass (live credentials)
- **Domain:** Registered domain with SSL certificate

---

## 🔧 Environment Setup

### 1. Generate JWT Secret

```bash
cd /home/ubuntu/rechargemax-production-OriginalBuild
./scripts/generate-jwt-secret.sh
```

Copy the generated secret for use in `.env`.

### 2. Create Environment File

```bash
cp .env.example .env
nano .env
```

### 3. Required Environment Variables

```bash
# Server Configuration
PORT=8080
ENVIRONMENT=production  # IMPORTANT: Must be "production"
APP_URL=https://yourdomain.com

# Database (PostgreSQL)
DATABASE_URL=postgresql://username:password@localhost:5432/rechargemax?sslmode=require

# JWT Authentication (REQUIRED - No default!)
JWT_SECRET=<paste-generated-secret-here>  # Must be 32+ characters

# Payment Gateway (Paystack - LIVE keys)
PAYSTACK_SECRET_KEY=sk_live_xxxxxxxxxxxxx  # NOT sk_test_!
PAYSTACK_PUBLIC_KEY=pk_live_xxxxxxxxxxxxx

# SMS Provider (Termii - LIVE)
TERMII_API_KEY=xxxxxxxxxxxxx
TERMII_SENDER_ID=RechargeMax

# VTU Provider (VTPass - LIVE)
VTPASS_API_KEY=xxxxxxxxxxxxx
VTPASS_PUBLIC_KEY=xxxxxxxxxxxxx
VTPASS_SECRET_KEY=xxxxxxxxxxxxx

# Admin Configuration
ADMIN_EMAIL=admin@yourdomain.com
ADMIN_PHONE=08012345678
```

### 4. Validate Configuration

The server will automatically validate configuration on startup:
- JWT_SECRET must be set and 32+ characters
- In production, API keys must be live keys (not test keys)
- Database connection must be valid

**Server will FAIL to start if validation fails!**

---

## 🗄️ Database Setup

### 1. Create Database

```bash
sudo -u postgres psql
```

```sql
CREATE DATABASE rechargemax;
CREATE USER rechargemax_user WITH ENCRYPTED PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE rechargemax TO rechargemax_user;
\q
```

### 2. Run Migrations

**IMPORTANT:** Run migrations in order!

```bash
cd database/migrations

# Core schema
psql -d rechargemax -f 01_core_tables_schema_2026_01_30_14_00.sql

# ... (run all existing migrations)

# Champion Developer Fixes (NEW)
psql -d rechargemax -f 23_fix_spin_race_condition.sql
psql -d rechargemax -f 24_standardize_amounts_to_kobo.sql
psql -d rechargemax -f 25_add_performance_indexes.sql
psql -d rechargemax -f 26_fix_phone_validation.sql
```

### 3. Verify Migrations

```bash
psql -d rechargemax -c "SELECT indexname FROM pg_indexes WHERE schemaname = 'public' AND indexname LIKE 'idx_%' ORDER BY indexname;"
```

You should see 40+ performance indexes.

---

## 🏗️ Backend Deployment

### 1. Build Backend

```bash
cd backend
go mod download
go build -o rechargemax cmd/server/main.go
```

### 2. Test Locally

```bash
./rechargemax
```

Expected output:
```
✅ Configuration validated successfully
✅ Database connected
✅ JWT secret validated (64 characters)
✅ Production API keys validated
Server starting on :8080
```

### 3. Create Systemd Service

```bash
sudo nano /etc/systemd/system/rechargemax.service
```

```ini
[Unit]
Description=RechargeMax Backend Service
After=network.target postgresql.service

[Service]
Type=simple
User=rechargemax
WorkingDirectory=/home/rechargemax/rechargemax-production-OriginalBuild/backend
ExecStart=/home/rechargemax/rechargemax-production-OriginalBuild/backend/rechargemax
Restart=on-failure
RestartSec=5s
StandardOutput=journal
StandardError=journal

# Environment
Environment="PORT=8080"
Environment="ENVIRONMENT=production"
EnvironmentFile=/home/rechargemax/rechargemax-production-OriginalBuild/.env

[Install]
WantedBy=multi-user.target
```

### 4. Start Service

```bash
sudo systemctl daemon-reload
sudo systemctl enable rechargemax
sudo systemctl start rechargemax
sudo systemctl status rechargemax
```

### 5. Check Logs

```bash
sudo journalctl -u rechargemax -f
```

---

## 🌐 Frontend Deployment

### 1. Build Frontend

```bash
cd frontend
npm install
npm run build
```

### 2. Configure Nginx

```bash
sudo nano /etc/nginx/sites-available/rechargemax
```

```nginx
server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com www.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    # Frontend
    root /home/rechargemax/rechargemax-production-OriginalBuild/frontend/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    # Backend API
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
}
```

### 3. Enable Site

```bash
sudo ln -s /etc/nginx/sites-available/rechargemax /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

---

## 🔒 SSL Certificate (Let's Encrypt)

```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com
```

---

## ✅ Health Checks

### Basic Health Check
```bash
curl https://yourdomain.com/api/v1/health
```

Expected response:
```json
{
  "status": "ok",
  "timestamp": "2026-02-01T12:00:00Z"
}
```

### Detailed Health Check
```bash
curl https://yourdomain.com/api/v1/health/detailed
```

Expected response:
```json
{
  "status": "healthy",
  "timestamp": "2026-02-01T12:00:00Z",
  "version": "1.0.0",
  "checks": {
    "database": {
      "status": "healthy",
      "message": "Database connection is healthy",
      "latency": "2.5ms"
    },
    "database_write": {
      "status": "healthy",
      "message": "Database write capability is healthy",
      "latency": "1.2ms"
    }
  }
}
```

---

## 🧪 Testing Champion Fixes

### Test 1: Server Won't Start Without JWT Secret

```bash
# Remove JWT_SECRET from .env
./rechargemax
```

Expected: Server fails with error message.

### Test 2: Admin Auth Works

```bash
# Try accessing admin endpoint with user token
curl -H "Authorization: Bearer <USER_TOKEN>" \
  https://yourdomain.com/api/v1/admin/dashboard
```

Expected: `403 Forbidden` with `insufficient_privileges` error.

### Test 3: Race Condition Fixed

```bash
# Spam spin button (simulate with curl)
for i in {1..10}; do
  curl -X POST -H "Authorization: Bearer <TOKEN>" \
    https://yourdomain.com/api/v1/spin/play &
done

# Check database
psql -d rechargemax -c "SELECT COUNT(*) FROM wheel_spins WHERE user_id = '<USER_ID>' AND DATE(created_at) = CURRENT_DATE;"
```

Expected: Count = 1 (not 10!)

### Test 4: OTP Rate Limiting

```bash
# Try to request OTP 6 times rapidly
for i in {1..6}; do
  curl -X POST https://yourdomain.com/api/v1/auth/request-otp \
    -H "Content-Type: application/json" \
    -d '{"msisdn":"08012345678"}'
done
```

Expected: 6th request returns `429 Too Many Requests`.

### Test 5: Phone Validation Fixed

```bash
# Try a 071X number (previously rejected)
curl -X POST https://yourdomain.com/api/v1/auth/request-otp \
  -H "Content-Type: application/json" \
  -d '{"msisdn":"07112345678"}'
```

Expected: `200 OK` (not validation error).

---

## 📊 Monitoring

### Application Logs

```bash
# Real-time logs
sudo journalctl -u rechargemax -f

# Filter by level
sudo journalctl -u rechargemax | grep ERROR

# Last 100 lines
sudo journalctl -u rechargemax -n 100
```

### Database Performance

```bash
# Check slow queries
psql -d rechargemax -c "SELECT query, mean_exec_time, calls FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;"

# Check index usage
psql -d rechargemax -c "SELECT schemaname, tablename, indexname, idx_scan FROM pg_stat_user_indexes ORDER BY idx_scan DESC;"
```

### System Resources

```bash
# CPU and memory
htop

# Disk usage
df -h

# Database connections
psql -d rechargemax -c "SELECT count(*) FROM pg_stat_activity;"
```

---

## 🔄 Backup and Recovery

### Database Backup

```bash
# Daily backup
pg_dump -U rechargemax_user -d rechargemax > backup_$(date +%Y%m%d).sql

# Automated backup (cron)
0 2 * * * pg_dump -U rechargemax_user -d rechargemax > /backups/rechargemax_$(date +\%Y\%m\%d).sql
```

### Database Restore

```bash
psql -U rechargemax_user -d rechargemax < backup_20260201.sql
```

---

## 🚨 Troubleshooting

### Server Won't Start

**Check logs:**
```bash
sudo journalctl -u rechargemax -n 50
```

**Common issues:**
- JWT_SECRET not set or too short
- Database connection failed
- Port 8080 already in use
- Test API keys in production

### Database Connection Failed

```bash
# Test connection
psql -U rechargemax_user -d rechargemax -h localhost

# Check PostgreSQL status
sudo systemctl status postgresql
```

### High Memory Usage

```bash
# Check Go process
ps aux | grep rechargemax

# Restart service
sudo systemctl restart rechargemax
```

---

## 📈 Performance Tuning

### Database

```sql
-- Increase shared buffers (25% of RAM)
ALTER SYSTEM SET shared_buffers = '2GB';

-- Increase work memory
ALTER SYSTEM SET work_mem = '64MB';

-- Reload configuration
SELECT pg_reload_conf();
```

### Nginx

```nginx
# In nginx.conf
worker_processes auto;
worker_connections 2048;

# Enable gzip compression
gzip on;
gzip_types text/plain text/css application/json application/javascript;
```

---

## 🎯 Champion Fixes Summary

### What Was Fixed:

1. **Admin Auth Vulnerability** ✅
   - Role-based access control
   - User tokens rejected from admin endpoints

2. **JWT Secret Validation** ✅
   - No hardcoded defaults
   - Server fails to start without strong secret

3. **Race Condition in Spins** ✅
   - Advisory locks
   - Database unique constraint

4. **Secure Random** ✅
   - Crypto/rand for prize selection

5. **Rate Limiting** ✅
   - General, OTP, and spin rate limits
   - OTP attempt limiting (5 attempts, 15 min lock)

6. **Database Indexes** ✅
   - 40+ performance indexes
   - Optimized for common queries

7. **Phone Validation** ✅
   - Accepts all valid Nigerian numbers
   - Fixed restrictive regex

8. **Amount Standardization** ✅ (Partial)
   - Database migration created
   - Entities updated to int64
   - Tier calculator uses kobo
   - Currency helper functions

---

## 📞 Support

For issues or questions:
- Check logs: `sudo journalctl -u rechargemax -f`
- Review this guide
- Check environment variables
- Verify migrations ran successfully

---

**Last Updated:** February 1, 2026  
**Version:** 1.0.0 (Champion Developer Review)
