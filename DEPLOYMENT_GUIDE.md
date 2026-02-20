# RechargeMax Deployment Guide

**Date:** February 20, 2026  
**Status:** Production Ready  
**Target Scale:** 50 Million Users

---

## Architecture Overview

```
┌─────────────────┐
│   Cloudflare    │  (CDN, DDoS Protection, SSL)
└────────┬────────┘
         │
┌────────▼────────┐
│  Load Balancer  │  (Nginx / AWS ALB)
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
┌───▼───┐ ┌──▼────┐
│ API 1 │ │ API 2 │  (Go Backend - Multiple Instances)
└───┬───┘ └──┬────┘
    │        │
    └────┬───┘
         │
┌────────▼────────┐
│   PostgreSQL    │  (Primary + Read Replicas)
│   + Redis       │  (Caching Layer)
└─────────────────┘
         │
    ┌────┴────┐
    │         │
┌───▼───┐ ┌──▼────┐
│Paystack│ │VTPass │  (External Services)
└────────┘ └───────┘
```

---

## Prerequisites

### System Requirements

**Minimum (Development):**
- CPU: 2 cores
- RAM: 4GB
- Storage: 20GB SSD
- OS: Ubuntu 22.04 LTS

**Recommended (Production):**
- CPU: 8 cores (16 for high traffic)
- RAM: 16GB (32GB for high traffic)
- Storage: 100GB SSD (with auto-scaling)
- OS: Ubuntu 22.04 LTS

### Software Dependencies

- **Go:** 1.21+
- **PostgreSQL:** 15+
- **Redis:** 7+
- **Node.js:** 20+
- **Nginx:** 1.24+
- **Docker:** 24+ (optional)
- **Git:** 2.40+

---

## Environment Setup

### 1. Clone Repository

```bash
git clone https://github.com/yourusername/RechargeMax_Clean.git
cd RechargeMax_Clean
```

### 2. Configure Environment Variables

**Backend (.env):**

```bash
cd backend
cp .env.example .env
nano .env
```

**Required Variables:**

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=rechargemax_user
DB_PASSWORD=your_secure_password_here
DB_NAME=rechargemax_db
DB_SSLMODE=disable  # Use 'require' in production

# JWT
JWT_SECRET=your_jwt_secret_here_min_32_chars
JWT_EXPIRY=24h

# Paystack (Production Keys)
PAYSTACK_SECRET_KEY=sk_live_xxxxxxxxxxxxx
PAYSTACK_PUBLIC_KEY=pk_live_xxxxxxxxxxxxx
PAYSTACK_WEBHOOK_SECRET=whsec_xxxxxxxxxxxxx

# VTPass (Production Keys)
VTPASS_API_KEY=your_vtpass_api_key
VTPASS_SECRET_KEY=your_vtpass_secret_key
VTPASS_BASE_URL=https://api.vtpass.com/api

# Server
PORT=8080
ENV=production
CORS_ORIGINS=https://yourdomain.com

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password

# Email (for notifications)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=noreply@yourdomain.com
SMTP_PASSWORD=your_smtp_password

# SMS (for OTP)
SMS_API_KEY=your_sms_api_key
SMS_SENDER_ID=RechargeMax

# Monitoring
SENTRY_DSN=https://xxxxx@sentry.io/xxxxx
```

**Frontend (.env):**

```bash
cd ../frontend
cp .env.example .env
nano .env
```

```env
VITE_API_BASE_URL=https://api.yourdomain.com
VITE_PAYSTACK_PUBLIC_KEY=pk_live_xxxxxxxxxxxxx
VITE_APP_NAME=RechargeMax Rewards
VITE_APP_URL=https://yourdomain.com
```

---

## Database Setup

### 1. Install PostgreSQL

```bash
sudo apt update
sudo apt install postgresql postgresql-contrib -y
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

### 2. Create Database and User

```bash
sudo -u postgres psql
```

```sql
CREATE DATABASE rechargemax_db;
CREATE USER rechargemax_user WITH ENCRYPTED PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE rechargemax_db TO rechargemax_user;
\q
```

### 3. Run Migrations

```bash
cd backend
go run cmd/migrate/main.go up
```

**Verify:**

```bash
psql -U rechargemax_user -d rechargemax_db -c "\dt"
```

**Expected:** 36 tables listed

### 4. Seed Essential Data

```bash
psql -U rechargemax_user -d rechargemax_db < seeds/essential_data.sql
```

**Tables Seeded:**
- `networks` (MTN, Airtel, Glo, 9mobile)
- `spin_tiers` (Bronze, Silver, Gold, Platinum, Diamond)
- `provider_configs` (VTPass configuration)
- `system_configs` (platform settings)

---

## Backend Deployment

### Option 1: Direct Deployment (Systemd)

**1. Build Binary:**

```bash
cd backend
go build -o rechargemax-api cmd/api/main.go
```

**2. Create Systemd Service:**

```bash
sudo nano /etc/systemd/system/rechargemax-api.service
```

```ini
[Unit]
Description=RechargeMax API Service
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/RechargeMax_Clean/backend
ExecStart=/home/ubuntu/RechargeMax_Clean/backend/rechargemax-api
Restart=always
RestartSec=5
StandardOutput=append:/var/log/rechargemax/api.log
StandardError=append:/var/log/rechargemax/api-error.log

Environment="ENV=production"
EnvironmentFile=/home/ubuntu/RechargeMax_Clean/backend/.env

[Install]
WantedBy=multi-user.target
```

**3. Start Service:**

```bash
sudo mkdir -p /var/log/rechargemax
sudo systemctl daemon-reload
sudo systemctl start rechargemax-api
sudo systemctl enable rechargemax-api
sudo systemctl status rechargemax-api
```

### Option 2: Docker Deployment

**1. Build Image:**

```bash
cd backend
docker build -t rechargemax-api:latest .
```

**2. Run Container:**

```bash
docker run -d \
  --name rechargemax-api \
  --restart unless-stopped \
  -p 8080:8080 \
  --env-file .env \
  rechargemax-api:latest
```

### Option 3: Docker Compose (Recommended)

**docker-compose.yml:**

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: rechargemax-db
    restart: unless-stopped
    environment:
      POSTGRES_DB: rechargemax_db
      POSTGRES_USER: rechargemax_user
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backend/migrations:/docker-entrypoint-initdb.d
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U rechargemax_user"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: rechargemax-redis
    restart: unless-stopped
    command: redis-server --requirepass ${REDIS_PASSWORD}
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  api:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: rechargemax-api
    restart: unless-stopped
    ports:
      - "8080:8080"
    env_file:
      - ./backend/.env
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_started
    volumes:
      - ./backend/logs:/app/logs

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    container_name: rechargemax-frontend
    restart: unless-stopped
    ports:
      - "3000:80"
    depends_on:
      - api

volumes:
  postgres_data:
  redis_data:
```

**Deploy:**

```bash
docker-compose up -d
docker-compose logs -f
```

---

## Frontend Deployment

### Option 1: Static Hosting (Vercel/Netlify)

**1. Build:**

```bash
cd frontend
pnpm install
pnpm run build
```

**2. Deploy to Vercel:**

```bash
vercel --prod
```

**3. Configure Environment:**

- Add environment variables in Vercel dashboard
- Set build command: `pnpm run build`
- Set output directory: `dist`

### Option 2: Nginx Hosting

**1. Build:**

```bash
cd frontend
pnpm install
pnpm run build
```

**2. Copy to Nginx:**

```bash
sudo cp -r dist/* /var/www/rechargemax/
```

**3. Configure Nginx:**

```bash
sudo nano /etc/nginx/sites-available/rechargemax
```

```nginx
server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com;

    root /var/www/rechargemax;
    index index.html;

    # SPA routing
    location / {
        try_files $uri $uri/ /index.html;
    }

    # API proxy
    location /api {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;
}
```

**4. Enable Site:**

```bash
sudo ln -s /etc/nginx/sites-available/rechargemax /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

**5. SSL Certificate (Let's Encrypt):**

```bash
sudo apt install certbot python3-certbot-nginx -y
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com
```

---

## Redis Setup

```bash
sudo apt install redis-server -y
sudo nano /etc/redis/redis.conf
```

**Configure:**

```conf
bind 127.0.0.1
requirepass your_redis_password
maxmemory 2gb
maxmemory-policy allkeys-lru
```

**Start:**

```bash
sudo systemctl start redis-server
sudo systemctl enable redis-server
```

---

## Monitoring & Logging

### 1. Application Logs

**Backend Logs:**

```bash
tail -f /var/log/rechargemax/api.log
```

**Nginx Logs:**

```bash
tail -f /var/log/nginx/access.log
tail -f /var/log/nginx/error.log
```

### 2. Database Monitoring

**Check Connections:**

```sql
SELECT count(*) FROM pg_stat_activity;
```

**Slow Queries:**

```sql
SELECT query, calls, total_time, mean_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;
```

### 3. Performance Monitoring

**Install Prometheus + Grafana:**

```bash
docker run -d --name prometheus -p 9090:9090 prom/prometheus
docker run -d --name grafana -p 3001:3000 grafana/grafana
```

**Metrics Endpoint:**

```go
// Add to main.go
import "github.com/prometheus/client_golang/prometheus/promhttp"

router.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

---

## Security Checklist

✅ **Environment Variables:** Never commit .env files  
✅ **Database:** Use strong passwords, enable SSL  
✅ **API:** Rate limiting enabled (100 req/min per IP)  
✅ **CORS:** Restrict to production domain only  
✅ **JWT:** Use strong secret (min 32 chars)  
✅ **HTTPS:** Force SSL redirect  
✅ **Webhooks:** Verify signatures (Paystack)  
✅ **SQL Injection:** Use parameterized queries (GORM)  
✅ **XSS:** Sanitize user inputs  
✅ **CSRF:** Use CSRF tokens  
✅ **Firewall:** Only expose ports 80, 443  
✅ **Backups:** Daily automated backups  

---

## Backup Strategy

### Database Backup

**Daily Backup Script:**

```bash
#!/bin/bash
BACKUP_DIR="/backups/postgres"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/rechargemax_$DATE.sql.gz"

mkdir -p $BACKUP_DIR

pg_dump -U rechargemax_user rechargemax_db | gzip > $BACKUP_FILE

# Keep only last 30 days
find $BACKUP_DIR -name "*.sql.gz" -mtime +30 -delete

echo "Backup completed: $BACKUP_FILE"
```

**Cron Job:**

```bash
crontab -e
```

```cron
0 2 * * * /home/ubuntu/scripts/backup_db.sh
```

### Restore Backup

```bash
gunzip -c /backups/postgres/rechargemax_20260220.sql.gz | \
  psql -U rechargemax_user rechargemax_db
```

---

## Scaling Strategy

### Horizontal Scaling (50M Users)

**1. Load Balancer:**
- Use AWS ALB or Nginx
- Distribute traffic across multiple API instances

**2. Database:**
- **Primary:** Write operations
- **Read Replicas:** Read operations (3-5 replicas)
- **Connection Pooling:** PgBouncer (max 1000 connections)

**3. Caching:**
- **Redis Cluster:** 3 master + 3 replica nodes
- Cache: User sessions, API responses, spin tiers

**4. CDN:**
- Cloudflare for static assets
- Edge caching for API responses

**5. Queue System:**
- RabbitMQ or AWS SQS for async tasks
- Process recharges, send emails, SMS

### Vertical Scaling

**Database:**
- Upgrade to 32GB RAM, 16 cores
- Use SSD storage (NVMe)

**API Servers:**
- 16GB RAM, 8 cores per instance
- Auto-scaling: 2-10 instances

---

## Troubleshooting

### Issue 1: Database Connection Failed

**Error:** `pq: password authentication failed`

**Solution:**
```bash
# Check .env file
cat backend/.env | grep DB_

# Test connection
psql -U rechargemax_user -d rechargemax_db -h localhost
```

### Issue 2: Migrations Failed

**Error:** `migration 036 already applied`

**Solution:**
```bash
# Check migration status
go run cmd/migrate/main.go version

# Rollback if needed
go run cmd/migrate/main.go down 1

# Re-apply
go run cmd/migrate/main.go up
```

### Issue 3: Paystack Webhook Not Received

**Solution:**
1. Check webhook URL in Paystack dashboard
2. Verify webhook secret in .env
3. Check firewall allows incoming connections
4. Test with Paystack webhook tester

---

## Production Checklist

✅ All environment variables set  
✅ Database migrations applied  
✅ Essential data seeded  
✅ SSL certificate installed  
✅ Firewall configured  
✅ Backups scheduled  
✅ Monitoring enabled  
✅ Load testing completed  
✅ Security audit passed  
✅ Documentation updated  
✅ Team trained on deployment process  

---

## Support & Maintenance

**Monitoring:**
- Check logs daily
- Monitor error rates
- Track API response times

**Updates:**
- Apply security patches weekly
- Update dependencies monthly
- Database maintenance quarterly

**Incident Response:**
- On-call rotation (24/7)
- Incident response playbook
- Post-mortem after major incidents

---

**Prepared By:** Engineering Team  
**Last Updated:** February 20, 2026  
**Version:** 1.0
