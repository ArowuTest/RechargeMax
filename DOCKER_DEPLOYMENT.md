# RechargeMax Docker Deployment Guide

## 🚀 Quick Start

### Prerequisites

- **Docker Desktop** (Windows 10/11, macOS, or Linux)
  - Download: https://www.docker.com/products/docker-desktop
  - Minimum: 4GB RAM, 20GB disk space
- **Docker Compose** (included with Docker Desktop)
- **Git** (for cloning the repository)

### One-Command Deployment

```bash
# Clone the repository
git clone <repository-url>
cd RechargeMax_Clean

# Start all services
docker-compose up -d

# View logs
docker-compose logs -f
```

That's it! The application will be available at:
- **Frontend:** http://localhost:8081 or http://localhost:3000
- **Backend API:** http://localhost:8080
- **Admin Portal:** http://localhost:8081/#/admin/login

---

## 📋 Detailed Setup Instructions

### Step 1: Install Docker Desktop

#### Windows
1. Download Docker Desktop from https://www.docker.com/products/docker-desktop
2. Run the installer
3. Enable WSL 2 backend (recommended)
4. Restart your computer
5. Launch Docker Desktop
6. Verify installation:
   ```powershell
   docker --version
   docker-compose --version
   ```

#### macOS
1. Download Docker Desktop for Mac
2. Drag Docker.app to Applications
3. Launch Docker Desktop
4. Verify installation:
   ```bash
   docker --version
   docker-compose --version
   ```

#### Linux
```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker
```

### Step 2: Configure Environment Variables

1. **Backend Configuration:**
   ```bash
   cd backend
   cp .env.example .env
   ```

2. **Edit `backend/.env`** with your credentials:
   ```env
   # Database (default values work for Docker)
   DATABASE_URL=postgresql://rechargemax:rechargemax123@postgres:5432/rechargemax_db?sslmode=disable
   
   # JWT Secret (CHANGE IN PRODUCTION!)
   JWT_SECRET=your-super-secret-jwt-key-change-in-production
   
   # Paystack (Get from https://dashboard.paystack.com/#/settings/developer)
   PAYSTACK_SECRET_KEY=sk_test_your_actual_key
   PAYSTACK_PUBLIC_KEY=pk_test_your_actual_key
   
   # VTPass (Get from https://www.vtpass.com/)
   VTPASS_API_KEY=your_actual_api_key
   VTPASS_PUBLIC_KEY=your_actual_public_key
   VTPASS_SECRET_KEY=your_actual_secret_key
   VTPASS_SANDBOX_MODE=true
   ```

3. **Frontend Configuration:**
   ```bash
   cd ../frontend
   cp .env.example .env
   ```

4. **Root `.env` for Docker Compose (optional):**
   ```bash
   cd ..
   cat > .env << EOF
   DB_PASSWORD=rechargemax123
   JWT_SECRET=your-super-secret-jwt-key-change-in-production
   PAYSTACK_SECRET_KEY=sk_test_your_actual_key
   PAYSTACK_PUBLIC_KEY=pk_test_your_actual_key
   VTPASS_API_KEY=your_actual_api_key
   VTPASS_PUBLIC_KEY=your_actual_public_key
   VTPASS_SECRET_KEY=your_actual_secret_key
   VTPASS_SANDBOX_MODE=true
   VITE_API_BASE_URL=/api/v1
   EOF
   ```

### Step 3: Deploy the Stack

```bash
# Build and start all services
docker-compose up -d --build

# Check service status
docker-compose ps

# View logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f postgres
```

### Step 4: Verify Deployment

1. **Check Database:**
   ```bash
   docker-compose exec postgres psql -U rechargemax -d rechargemax_db -c "\dt"
   ```

2. **Check Backend Health:**
   ```bash
   curl http://localhost:8080/health
   ```

3. **Check Frontend:**
   ```bash
   curl http://localhost:8081/
   ```

4. **Test Admin Login:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/admin/login \
     -H "Content-Type: application/json" \
     -d '{"email":"admin@rechargemax.ng","password":"Admin@123"}'
   ```

### Step 5: Access the Application

- **User Portal:** http://localhost:8081
- **Admin Portal:** http://localhost:8081/#/admin/login
  - Email: `admin@rechargemax.ng`
  - Password: `Admin@123`
- **API Documentation:** http://localhost:8080/api/v1
- **Backend Health:** http://localhost:8080/health

---

## 🔧 Docker Commands Reference

### Service Management

```bash
# Start services
docker-compose up -d

# Stop services
docker-compose stop

# Restart services
docker-compose restart

# Stop and remove containers
docker-compose down

# Stop and remove containers + volumes (CAUTION: deletes database!)
docker-compose down -v

# Rebuild and restart
docker-compose up -d --build

# Scale a service
docker-compose up -d --scale backend=3
```

### Logs and Monitoring

```bash
# View all logs
docker-compose logs

# Follow logs in real-time
docker-compose logs -f

# View logs for specific service
docker-compose logs backend
docker-compose logs frontend
docker-compose logs postgres

# View last 100 lines
docker-compose logs --tail=100

# View logs with timestamps
docker-compose logs -t
```

### Container Management

```bash
# List running containers
docker-compose ps

# Execute command in container
docker-compose exec backend sh
docker-compose exec postgres psql -U rechargemax -d rechargemax_db

# View container resource usage
docker stats

# Inspect container
docker inspect rechargemax_backend
```

### Database Operations

```bash
# Access PostgreSQL shell
docker-compose exec postgres psql -U rechargemax -d rechargemax_db

# Backup database
docker-compose exec postgres pg_dump -U rechargemax rechargemax_db > backup.sql

# Restore database
docker-compose exec -T postgres psql -U rechargemax -d rechargemax_db < backup.sql

# View database tables
docker-compose exec postgres psql -U rechargemax -d rechargemax_db -c "\dt"

# View admin users
docker-compose exec postgres psql -U rechargemax -d rechargemax_db -c "SELECT email, role FROM admin_users;"
```

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Docker Network                          │
│                   (rechargemax_network)                      │
│                                                              │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────┐ │
│  │   Frontend   │      │   Backend    │      │ Database │ │
│  │  (React/Vite)│─────▶│   (Go/Gin)   │─────▶│PostgreSQL│ │
│  │   Port: 80   │      │  Port: 8080  │      │Port: 5432│ │
│  │  Nginx Proxy │      │              │      │          │ │
│  └──────────────┘      └──────────────┘      └──────────┘ │
│        │                                                    │
│        │ API Proxy: /api/* → backend:8080                  │
│        │                                                    │
└────────┼────────────────────────────────────────────────────┘
         │
    Host Ports:
    - 8081 → Frontend
    - 3000 → Frontend (alternative)
    - 8080 → Backend API
    - 5432 → PostgreSQL
```

### Service Details

#### Frontend Container
- **Base Image:** nginx:alpine
- **Build:** Node.js 20 (build stage)
- **Exposed Ports:** 80 (mapped to 8081, 3000 on host)
- **Features:**
  - React/Vite SPA
  - Nginx reverse proxy
  - API proxy to backend
  - Gzip compression
  - Static asset caching
  - Security headers

#### Backend Container
- **Base Image:** golang:1.22-alpine (build), alpine:latest (runtime)
- **Exposed Port:** 8080
- **Features:**
  - Go/Gin REST API
  - JWT authentication
  - GORM ORM
  - Auto-migrations
  - Health checks
  - Structured logging

#### Database Container
- **Base Image:** postgres:14-alpine
- **Exposed Port:** 5432
- **Features:**
  - PostgreSQL 14
  - Persistent volume
  - Auto-seed on first run
  - Health checks
  - Optimized configuration

---

## 🔐 Security Considerations

### Production Deployment

1. **Change Default Credentials:**
   ```bash
   # Generate strong JWT secret
   openssl rand -base64 32
   
   # Update .env files with production values
   ```

2. **Use Environment Variables:**
   - Never commit `.env` files to Git
   - Use Docker secrets or external secret management
   - Rotate credentials regularly

3. **Enable HTTPS:**
   - Use the nginx service profile for production
   - Configure SSL certificates
   - Redirect HTTP to HTTPS

4. **Database Security:**
   - Use strong passwords
   - Enable SSL connections
   - Restrict network access
   - Regular backups

5. **Container Security:**
   - Run as non-root user (already configured)
   - Scan images for vulnerabilities
   - Keep images updated
   - Use minimal base images

### Network Security

```yaml
# Example production nginx configuration
services:
  nginx:
    profiles:
      - production
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/ssl:/etc/nginx/ssl:ro
```

---

## 🐛 Troubleshooting

### Common Issues

#### 1. Port Already in Use

**Error:** `Bind for 0.0.0.0:8080 failed: port is already allocated`

**Solution:**
```bash
# Find process using the port
# Windows
netstat -ano | findstr :8080

# macOS/Linux
lsof -i :8080

# Kill the process or change port in docker-compose.yml
```

#### 2. Database Connection Failed

**Error:** `connection refused` or `database does not exist`

**Solution:**
```bash
# Check database logs
docker-compose logs postgres

# Restart database
docker-compose restart postgres

# Wait for health check
docker-compose ps

# Manually create database if needed
docker-compose exec postgres psql -U rechargemax -c "CREATE DATABASE rechargemax_db;"
```

#### 3. Backend Won't Start

**Error:** `exit code 1` or `panic: runtime error`

**Solution:**
```bash
# Check backend logs
docker-compose logs backend

# Verify environment variables
docker-compose exec backend env | grep DATABASE_URL

# Rebuild backend
docker-compose up -d --build backend
```

#### 4. Frontend Shows 404

**Error:** `Cannot GET /api/v1/...`

**Solution:**
```bash
# Check nginx proxy configuration
docker-compose exec frontend cat /etc/nginx/conf.d/default.conf

# Verify backend is running
curl http://localhost:8080/health

# Restart frontend
docker-compose restart frontend
```

#### 5. Slow Performance

**Solution:**
```bash
# Check resource usage
docker stats

# Increase Docker Desktop resources
# Settings → Resources → Advanced
# Recommended: 4GB RAM, 2 CPUs

# Prune unused resources
docker system prune -a
```

### Reset Everything

```bash
# Stop all containers
docker-compose down

# Remove volumes (CAUTION: deletes all data!)
docker-compose down -v

# Remove all images
docker-compose down --rmi all

# Clean Docker system
docker system prune -a --volumes

# Rebuild from scratch
docker-compose up -d --build
```

---

## 📊 Monitoring and Maintenance

### Health Checks

All services have built-in health checks:

```bash
# Check service health
docker-compose ps

# Manual health check
curl http://localhost:8080/health
curl http://localhost:8081/health
```

### Logs Management

```bash
# Rotate logs
docker-compose logs --tail=1000 > logs_$(date +%Y%m%d).txt

# Clear logs
truncate -s 0 $(docker inspect --format='{{.LogPath}}' rechargemax_backend)
```

### Backups

```bash
# Automated backup script
#!/bin/bash
BACKUP_DIR="./backups"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup directory
mkdir -p $BACKUP_DIR

# Backup database
docker-compose exec -T postgres pg_dump -U rechargemax rechargemax_db | gzip > $BACKUP_DIR/db_backup_$DATE.sql.gz

# Backup volumes
docker run --rm -v rechargemax_postgres_data:/data -v $(pwd)/$BACKUP_DIR:/backup alpine tar czf /backup/volume_backup_$DATE.tar.gz -C /data .

echo "Backup completed: $BACKUP_DIR"
```

### Updates

```bash
# Pull latest images
docker-compose pull

# Rebuild and restart
docker-compose up -d --build

# Remove old images
docker image prune -a
```

---

## 🚀 Production Deployment

### Using Docker Swarm

```bash
# Initialize swarm
docker swarm init

# Deploy stack
docker stack deploy -c docker-compose.yml rechargemax

# Check services
docker service ls

# Scale services
docker service scale rechargemax_backend=3
```

### Using Kubernetes

```bash
# Convert docker-compose to k8s
kompose convert -f docker-compose.yml

# Apply to cluster
kubectl apply -f .

# Check pods
kubectl get pods
```

### Cloud Deployment

#### AWS ECS
- Use AWS Fargate for serverless containers
- Configure Application Load Balancer
- Use RDS for PostgreSQL

#### Google Cloud Run
- Deploy frontend and backend separately
- Use Cloud SQL for PostgreSQL
- Configure Cloud Load Balancing

#### Azure Container Instances
- Use Azure Container Instances
- Configure Azure Database for PostgreSQL
- Use Azure Application Gateway

---

## 📝 Environment Variables Reference

### Backend (.env)

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DATABASE_URL` | PostgreSQL connection string | - | Yes |
| `PORT` | Backend server port | 8080 | No |
| `JWT_SECRET` | Secret for JWT tokens | - | Yes |
| `PAYSTACK_SECRET_KEY` | Paystack secret key | - | Yes |
| `PAYSTACK_PUBLIC_KEY` | Paystack public key | - | Yes |
| `VTPASS_API_KEY` | VTPass API key | - | Yes |
| `VTPASS_PUBLIC_KEY` | VTPass public key | - | Yes |
| `VTPASS_SECRET_KEY` | VTPass secret key | - | Yes |
| `VTPASS_SANDBOX_MODE` | Enable sandbox mode | true | No |
| `GIN_MODE` | Gin framework mode | release | No |
| `ALLOWED_ORIGINS` | CORS allowed origins | * | No |

### Frontend (.env)

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `VITE_API_BASE_URL` | API base URL | /api/v1 | Yes |

### Docker Compose (.env)

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DB_PASSWORD` | Database password | rechargemax123 | Yes |
| All backend variables | Passed to backend | - | Yes |

---

## 🎯 Performance Optimization

### Docker Desktop Settings

**Recommended Settings:**
- **Memory:** 4-8 GB
- **CPUs:** 2-4 cores
- **Disk:** 20-50 GB
- **Swap:** 1-2 GB

### Build Optimization

```dockerfile
# Use multi-stage builds (already implemented)
# Use .dockerignore (already implemented)
# Cache dependencies
# Minimize layers
```

### Runtime Optimization

```yaml
# Resource limits in docker-compose.yml
services:
  backend:
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
```

---

## 📚 Additional Resources

- **Docker Documentation:** https://docs.docker.com/
- **Docker Compose:** https://docs.docker.com/compose/
- **PostgreSQL:** https://www.postgresql.org/docs/
- **Go:** https://golang.org/doc/
- **React:** https://reactjs.org/docs/
- **Nginx:** https://nginx.org/en/docs/

---

## 🆘 Support

For issues or questions:
1. Check the troubleshooting section above
2. Review Docker logs: `docker-compose logs`
3. Check service status: `docker-compose ps`
4. Verify environment variables
5. Consult the main README.md

---

**Last Updated:** February 14, 2026  
**Version:** 1.0.0  
**Maintainer:** RechargeMax Development Team
