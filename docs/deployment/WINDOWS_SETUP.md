# RechargeMax - Windows Setup Guide

## 🪟 Quick Start for Windows

### Step 1: Install Docker Desktop

1. **Download Docker Desktop for Windows**
   - Visit: https://www.docker.com/products/docker-desktop
   - Click "Download for Windows"
   - Minimum Requirements:
     - Windows 10 64-bit: Pro, Enterprise, or Education (Build 19041 or higher)
     - Windows 11 64-bit
     - 4GB RAM minimum (8GB recommended)
     - WSL 2 feature enabled

2. **Install Docker Desktop**
   - Run the installer (`Docker Desktop Installer.exe`)
   - Follow the installation wizard
   - Enable "Use WSL 2 instead of Hyper-V" (recommended)
   - Restart your computer when prompted

3. **Verify Installation**
   - Open PowerShell or Command Prompt
   - Run:
     ```powershell
     docker --version
     docker-compose --version
     ```

### Step 2: Extract the Package

1. Extract `RechargeMax_Docker_Package.zip` to a folder, e.g., `C:\RechargeMax`
2. Open PowerShell as Administrator
3. Navigate to the folder:
   ```powershell
   cd C:\RechargeMax
   ```

### Step 3: Configure Environment Variables

1. **Backend Configuration:**
   ```powershell
   cd backend
   copy .env.example .env
   notepad .env
   ```

2. **Update the following in `.env`:**
   - `PAYSTACK_SECRET_KEY` - Your Paystack secret key
   - `PAYSTACK_PUBLIC_KEY` - Your Paystack public key
   - `VTPASS_API_KEY` - Your VTPass API key
   - `VTPASS_PUBLIC_KEY` - Your VTPass public key
   - `VTPASS_SECRET_KEY` - Your VTPass secret key
   - `JWT_SECRET` - Change to a strong random string

3. **Frontend Configuration:**
   ```powershell
   cd ..\frontend
   copy .env.example .env
   ```

### Step 4: Start the Application

```powershell
# Navigate to project root
cd C:\RechargeMax

# Start all services
docker-compose up -d

# View logs
docker-compose logs -f
```

### Step 5: Access the Application

- **User Portal:** http://localhost:8081
- **Admin Portal:** http://localhost:8081/#/admin/login
  - Email: `admin@rechargemax.ng`
  - Password: `Admin@123`
- **Backend API:** http://localhost:8080

---

## 🔧 Common Windows Issues

### Issue 1: WSL 2 Not Enabled

**Error:** "WSL 2 installation is incomplete"

**Solution:**
```powershell
# Run in PowerShell as Administrator
wsl --install
wsl --set-default-version 2

# Restart computer
```

### Issue 2: Hyper-V Conflicts

**Error:** "Hyper-V and Containers features are not enabled"

**Solution:**
```powershell
# Run in PowerShell as Administrator
Enable-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V -All
Enable-WindowsOptionalFeature -Online -FeatureName Containers -All

# Restart computer
```

### Issue 3: Port Already in Use

**Error:** "Bind for 0.0.0.0:8080 failed"

**Solution:**
```powershell
# Find process using port
netstat -ano | findstr :8080

# Kill process (replace PID with actual process ID)
taskkill /PID <PID> /F

# Or change port in docker-compose.yml
```

### Issue 4: Docker Desktop Won't Start

**Solution:**
1. Open Task Manager (Ctrl+Shift+Esc)
2. End all Docker processes
3. Restart Docker Desktop
4. If still failing, reinstall Docker Desktop

### Issue 5: Slow Performance

**Solution:**
1. Open Docker Desktop
2. Go to Settings → Resources
3. Increase:
   - Memory to 4GB or more
   - CPUs to 2 or more
   - Disk image size to 60GB or more
4. Click "Apply & Restart"

---

## 📝 Useful Windows Commands

### PowerShell Commands

```powershell
# Check Docker status
docker info

# List running containers
docker ps

# View all containers
docker ps -a

# Stop all containers
docker-compose down

# Restart services
docker-compose restart

# View logs
docker-compose logs backend
docker-compose logs frontend
docker-compose logs postgres

# Access database
docker-compose exec postgres psql -U rechargemax -d rechargemax_db

# Backup database
docker-compose exec postgres pg_dump -U rechargemax rechargemax_db > backup.sql

# Clean up Docker
docker system prune -a
```

### File Paths

Windows uses backslashes (`\`) for paths, but Docker uses forward slashes (`/`).

**Correct:**
```powershell
# In PowerShell
cd C:\RechargeMax

# In Docker volumes (docker-compose.yml)
./backend:/app
```

---

## 🚀 Production Deployment on Windows Server

### Using Windows Server 2019/2022

1. **Install Docker Enterprise:**
   ```powershell
   Install-Module -Name DockerMsftProvider -Repository PSGallery -Force
   Install-Package -Name docker -ProviderName DockerMsftProvider
   Restart-Computer -Force
   ```

2. **Configure Firewall:**
   ```powershell
   New-NetFirewallRule -DisplayName "Docker API" -Direction Inbound -LocalPort 2375 -Protocol TCP -Action Allow
   New-NetFirewallRule -DisplayName "HTTP" -Direction Inbound -LocalPort 80 -Protocol TCP -Action Allow
   New-NetFirewallRule -DisplayName "HTTPS" -Direction Inbound -LocalPort 443 -Protocol TCP -Action Allow
   ```

3. **Deploy Application:**
   ```powershell
   docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
   ```

---

## 🔐 Security Best Practices

1. **Change Default Passwords:**
   - Admin password: `Admin@123` → Change immediately
   - Database password: Update in `.env`
   - JWT secret: Generate strong random string

2. **Enable Windows Defender:**
   ```powershell
   Set-MpPreference -DisableRealtimeMonitoring $false
   ```

3. **Keep Docker Updated:**
   - Check for updates in Docker Desktop
   - Update regularly for security patches

4. **Use HTTPS in Production:**
   - Obtain SSL certificate (Let's Encrypt)
   - Configure nginx with SSL
   - Redirect HTTP to HTTPS

---

## 📊 Monitoring on Windows

### Using Windows Performance Monitor

1. Open Performance Monitor (`perfmon`)
2. Add counters:
   - Processor Time
   - Memory Usage
   - Network Utilization
   - Disk I/O

### Using Docker Desktop Dashboard

1. Open Docker Desktop
2. Click on container name
3. View:
   - CPU usage
   - Memory usage
   - Network I/O
   - Disk I/O

---

## 🆘 Getting Help

### Check Logs

```powershell
# All services
docker-compose logs

# Specific service
docker-compose logs backend

# Follow logs in real-time
docker-compose logs -f

# Last 100 lines
docker-compose logs --tail=100
```

### Check Service Health

```powershell
# Service status
docker-compose ps

# Container details
docker inspect rechargemax_backend

# Resource usage
docker stats
```

### Reset Everything

```powershell
# Stop all services
docker-compose down

# Remove volumes (deletes database!)
docker-compose down -v

# Remove all images
docker-compose down --rmi all

# Clean Docker system
docker system prune -a --volumes

# Rebuild from scratch
docker-compose up -d --build
```

---

## 📚 Additional Resources

- **Docker Desktop for Windows:** https://docs.docker.com/desktop/windows/
- **WSL 2 Documentation:** https://docs.microsoft.com/en-us/windows/wsl/
- **PowerShell Documentation:** https://docs.microsoft.com/en-us/powershell/
- **Windows Server Containers:** https://docs.microsoft.com/en-us/virtualization/windowscontainers/

---

## ✅ Checklist

- [ ] Docker Desktop installed
- [ ] WSL 2 enabled (for Windows 10/11)
- [ ] Repository extracted
- [ ] Environment variables configured
- [ ] Services started with `docker-compose up -d`
- [ ] Application accessible at http://localhost:8081
- [ ] Admin login working
- [ ] Backend API responding
- [ ] Database initialized

---

**For detailed deployment instructions, see `DOCKER_DEPLOYMENT.md`**

**Last Updated:** February 14, 2026  
**Version:** 1.0.0
