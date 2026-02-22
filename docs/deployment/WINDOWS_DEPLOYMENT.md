# RechargeMax - Windows Deployment Guide

## Prerequisites

1. **Docker Desktop for Windows**
   - Download from: https://www.docker.com/products/docker-desktop/
   - Ensure WSL 2 is enabled
   - Minimum requirements: Windows 10 64-bit Pro, Enterprise, or Education (Build 19041 or higher)

2. **Git for Windows** (optional, for cloning)
   - Download from: https://git-scm.com/download/win

## Quick Start (5 Minutes)

### Step 1: Extract the Package
```powershell
# Extract the zip file to a location like:
C:\rechargemax\
```

### Step 2: Configure Environment Variables
```powershell
# Navigate to the project directory
cd C:\rechargemax\

# Copy the example environment file
copy .env.example .env

# Edit .env with your favorite text editor (Notepad, VS Code, etc.)
notepad .env
```

**Important Environment Variables to Update:**
```env
# Database (default values work for local development)
DB_PASSWORD=rechargemax_password_change_in_production

# JWT Secret (MUST change in production)
JWT_SECRET=your-super-secret-jwt-key-min-32-characters

# Payment Gateway Keys (get from Paystack dashboard)
PAYSTACK_SECRET_KEY=sk_test_your_secret_key_here
PAYSTACK_PUBLIC_KEY=pk_test_your_public_key_here
```

### Step 3: Start the Application
```powershell
# Open PowerShell as Administrator
# Navigate to project directory
cd C:\rechargemax\

# Start all services
docker-compose up -d

# Check if services are running
docker-compose ps
```

### Step 4: Initialize the Database
```powershell
# Run database migrations
docker-compose exec backend ./rechargemax migrate

# Seed initial data (optional)
docker-compose exec postgres psql -U rechargemax_user -d rechargemax -f /docker-entrypoint-initdb.d/seed_data.sql
```

### Step 5: Access the Application
- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **API Documentation**: http://localhost:8080/swagger/index.html

## Default Admin Credentials
```
Email: superadmin@rechargemax.ng
Password: SuperAdmin123!
```

**⚠️ IMPORTANT: Change these credentials immediately after first login!**

## Common Commands

### View Logs
```powershell
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f postgres
```

### Stop the Application
```powershell
docker-compose down
```

### Restart a Service
```powershell
docker-compose restart backend
```

### Rebuild After Code Changes
```powershell
docker-compose down
docker-compose build
docker-compose up -d
```

### Database Backup
```powershell
# Create backup
docker-compose exec postgres pg_dump -U rechargemax_user rechargemax > backup_$(Get-Date -Format "yyyyMMdd_HHmmss").sql

# Restore backup
docker-compose exec -T postgres psql -U rechargemax_user rechargemax < backup_20260201_120000.sql
```

### Clean Everything (Fresh Start)
```powershell
# Stop and remove all containers, networks, and volumes
docker-compose down -v

# Remove all images
docker-compose down --rmi all

# Start fresh
docker-compose up -d
```

## Troubleshooting

### Port Already in Use
If ports 3000, 8080, or 5432 are already in use:

1. Edit `docker-compose.yml`
2. Change the port mappings:
   ```yaml
   ports:
     - "3001:80"  # Frontend (change 3000 to 3001)
     - "8081:8080"  # Backend (change 8080 to 8081)
     - "5433:5432"  # Database (change 5432 to 5433)
   ```

### Docker Desktop Not Starting
1. Enable WSL 2: `wsl --install`
2. Restart Windows
3. Open Docker Desktop

### Database Connection Issues
```powershell
# Check if PostgreSQL is healthy
docker-compose ps postgres

# View database logs
docker-compose logs postgres

# Restart database
docker-compose restart postgres
```

### Backend Not Starting
```powershell
# Check backend logs
docker-compose logs backend

# Common issues:
# 1. Database not ready - wait 30 seconds and restart
# 2. Environment variables missing - check .env file
# 3. Port conflict - change port in docker-compose.yml
```

## Production Deployment

### Security Checklist
- [ ] Change all default passwords
- [ ] Update JWT_SECRET to a strong random string (min 32 characters)
- [ ] Add real Paystack API keys
- [ ] Enable HTTPS/SSL
- [ ] Set up firewall rules
- [ ] Configure backup strategy
- [ ] Enable logging and monitoring
- [ ] Review and update CORS settings

### Performance Optimization
1. **Database**: Increase PostgreSQL shared_buffers and work_mem
2. **Backend**: Set appropriate number of workers
3. **Frontend**: Enable gzip compression in nginx
4. **Network**: Use a reverse proxy (nginx/traefik) for load balancing

## Support

For issues or questions:
- Email: support@rechargemax.ng
- Documentation: https://docs.rechargemax.ng
- GitHub Issues: https://github.com/rechargemax/rechargemax/issues

## License

Proprietary - All Rights Reserved
