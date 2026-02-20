# RechargeMax Rewards Platform - Deployment Guide

**Version:** Production-Ready v1.0  
**Date:** February 12, 2026  
**Status:** ✅ All Fixes Applied & Tested

---

## 📋 Table of Contents

1. [System Requirements](#system-requirements)
2. [Quick Start](#quick-start)
3. [Database Setup](#database-setup)
4. [Backend Setup](#backend-setup)
5. [Frontend Setup](#frontend-setup)
6. [Testing](#testing)
7. [Production Deployment](#production-deployment)
8. [Troubleshooting](#troubleshooting)

---

## 🖥️ System Requirements

### Minimum Requirements
- **OS:** Windows 10/11, macOS 12+, or Ubuntu 20.04+
- **RAM:** 4GB minimum, 8GB recommended
- **Disk Space:** 2GB free space
- **Network:** Internet connection for package downloads

### Required Software
- **Go:** Version 1.21+ ([Download](https://go.dev/dl/))
- **Node.js:** Version 18+ ([Download](https://nodejs.org/))
- **PostgreSQL:** Version 14+ ([Download](https://www.postgresql.org/download/))
- **Git:** Latest version ([Download](https://git-scm.com/downloads))

---

## 🚀 Quick Start

### 1. Extract the Package
```bash
# Extract the zip file to your desired location
unzip RechargeMax_Production_Ready.zip
cd RechargeMax_Updated
```

### 2. Database Setup
```bash
# Start PostgreSQL service
# Windows: Start from Services
# macOS: brew services start postgresql
# Linux: sudo systemctl start postgresql

# Create database and user
psql -U postgres
CREATE DATABASE rechargemax_db;
CREATE USER rechargemax WITH PASSWORD 'rechargemax123';
GRANT ALL PRIVILEGES ON DATABASE rechargemax_db TO rechargemax;
\q
```

### 3. Load Database Schema & Seed Data
```bash
# Navigate to backend directory
cd backend

# Run the backend once to create tables (GORM AutoMigrate)
go run cmd/server/main.go
# Press Ctrl+C after you see "RechargeMax Rewards Platform - READY!"

# Load production seed data
psql -U rechargemax -d rechargemax_db -f ../database/seeds/MASTER_PRODUCTION_SEED_CORRECTED.sql
```

### 4. Start Backend
```bash
# In backend directory
go run cmd/server/main.go

# Backend will start on http://localhost:8080
```

### 5. Start Frontend
```bash
# In a new terminal, navigate to frontend directory
cd frontend

# Install dependencies (first time only)
npm install

# Start development server
npm run dev

# Frontend will start on http://localhost:5173
```

### 6. Access the Platform
- **User Portal:** http://localhost:5173
- **Admin Portal:** http://localhost:5173/admin
- **API Documentation:** http://localhost:8080/api/v1
- **Health Check:** http://localhost:8080/health

---

## 🗄️ Database Setup

### PostgreSQL Installation

#### Windows
1. Download PostgreSQL from https://www.postgresql.org/download/windows/
2. Run the installer
3. Set password for `postgres` user
4. Add PostgreSQL to PATH

#### macOS
```bash
brew install postgresql@14
brew services start postgresql@14
```

#### Linux (Ubuntu/Debian)
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

### Create Database
```sql
-- Connect to PostgreSQL
psql -U postgres

-- Create database
CREATE DATABASE rechargemax_db;

-- Create user
CREATE USER rechargemax WITH PASSWORD 'rechargemax123';

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE rechargemax_db TO rechargemax;

-- Connect to the database
\c rechargemax_db

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Exit
\q
```

### Load Seed Data
```bash
# The seed file contains:
# - 4 Network Configurations (MTN, Airtel, Glo, 9mobile)
# - 66 Data Plans
# - 4 Subscription Tiers
# - 15 Wheel Prizes
# - 1 Admin User (admin@rechargemax.ng / Admin@123456)

psql -U rechargemax -d rechargemax_db -f database/seeds/MASTER_PRODUCTION_SEED_CORRECTED.sql
```

---

## 🔧 Backend Setup

### Environment Configuration

The backend uses environment variables from `.env` file. The file is already configured with development defaults.

**Key Environment Variables:**
```env
DATABASE_URL=postgresql://rechargemax:rechargemax123@localhost:5432/rechargemax_db?sslmode=disable
JWT_SECRET=rechargemax_super_secret_key_2026_production_ready
PORT=8080
GIN_MODE=debug
```

### Install Dependencies
```bash
cd backend
go mod download
```

### Build the Backend
```bash
# Development build
go build -o rechargemax cmd/server/main.go

# Production build (optimized)
go build -ldflags="-s -w" -o rechargemax cmd/server/main.go
```

### Run the Backend
```bash
# Development mode (with hot reload)
go run cmd/server/main.go

# Production mode (using compiled binary)
./rechargemax
```

### Verify Backend is Running
```bash
# Health check
curl http://localhost:8080/health

# Test admin login
curl -X POST http://localhost:8080/api/v1/admin/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@rechargemax.ng","password":"Admin@123456"}'
```

---

## 🎨 Frontend Setup

### Install Dependencies
```bash
cd frontend
npm install
```

### Environment Configuration

The frontend is pre-configured to connect to `http://localhost:8080` for the backend API.

If you need to change the API URL, update `frontend/src/lib/api-client.ts`:
```typescript
const API_BASE_URL = 'http://localhost:8080';
```

### Development Server
```bash
npm run dev
# Frontend will start on http://localhost:5173
```

### Production Build
```bash
npm run build
# Build output will be in frontend/dist/
```

### Preview Production Build
```bash
npm run preview
```

---

## 🧪 Testing

### Backend Tests
```bash
cd backend
go test ./...
```

### Frontend Tests
```bash
cd frontend
npm run test
```

### Manual Testing Checklist

#### Admin Portal
1. ✅ Login at http://localhost:5173/admin
   - Email: `admin@rechargemax.ng`
   - Password: `Admin@123456`

2. ✅ Dashboard
   - View statistics
   - Check charts and graphs

3. ✅ User Management
   - View all users
   - Search users
   - Update user status

4. ✅ Recharge Monitoring
   - View transactions
   - Filter by status
   - Retry failed transactions

5. ✅ Affiliate Management
   - View affiliates
   - Approve/reject applications
   - View statistics

6. ✅ Network Configuration
   - View networks
   - View data plans
   - Update configurations

7. ✅ Wheel Prizes
   - View prizes
   - Update probabilities
   - Manage inventory

#### User Portal
1. ✅ Homepage
2. ✅ Recharge flow
3. ✅ Spin-to-win
4. ✅ Points system
5. ✅ Affiliate program

---

## 🚀 Production Deployment

### Backend Deployment (Render.com)

1. **Create PostgreSQL Database**
   - Go to Render Dashboard
   - Create new PostgreSQL database
   - Copy the Internal Database URL

2. **Create Web Service**
   - Create new Web Service
   - Connect your Git repository
   - Build Command: `go build -o rechargemax cmd/server/main.go`
   - Start Command: `./rechargemax`
   - Add environment variables:
     ```
     DATABASE_URL=<your_render_postgres_url>
     JWT_SECRET=<generate_with_openssl_rand_-hex_32>
     PORT=8080
     GIN_MODE=release
     ENVIRONMENT=production
     ```

3. **Load Seed Data**
   ```bash
   # Connect to Render PostgreSQL
   psql <your_render_postgres_url>
   
   # Load seed data
   \i database/seeds/MASTER_PRODUCTION_SEED_CORRECTED.sql
   ```

### Frontend Deployment (Vercel/Netlify)

1. **Build the Frontend**
   ```bash
   cd frontend
   npm run build
   ```

2. **Deploy to Vercel**
   ```bash
   npm install -g vercel
   vercel --prod
   ```

3. **Update API URL**
   - Update `frontend/src/lib/api-client.ts` with your Render backend URL
   - Rebuild and redeploy

---

## 🐛 Troubleshooting

### Backend Issues

#### "Database connection failed"
```bash
# Check PostgreSQL is running
# Windows: Check Services
# macOS: brew services list
# Linux: sudo systemctl status postgresql

# Verify credentials
psql -U rechargemax -d rechargemax_db
```

#### "Port 8080 already in use"
```bash
# Find process using port 8080
# Windows: netstat -ano | findstr :8080
# macOS/Linux: lsof -i :8080

# Kill the process or change PORT in .env
```

#### "JWT_SECRET is required"
```bash
# Generate a new JWT secret
openssl rand -hex 32

# Add to backend/.env
JWT_SECRET=<generated_secret>
```

### Frontend Issues

#### "Cannot connect to backend"
```bash
# Verify backend is running
curl http://localhost:8080/health

# Check API_BASE_URL in frontend/src/lib/api-client.ts
```

#### "npm install fails"
```bash
# Clear npm cache
npm cache clean --force

# Delete node_modules and package-lock.json
rm -rf node_modules package-lock.json

# Reinstall
npm install
```

### Database Issues

#### "Extension uuid-ossp does not exist"
```sql
-- Connect to database
psql -U rechargemax -d rechargemax_db

-- Enable extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

#### "Permission denied for database"
```sql
-- Connect as postgres user
psql -U postgres

-- Grant all privileges
GRANT ALL PRIVILEGES ON DATABASE rechargemax_db TO rechargemax;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO rechargemax;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO rechargemax;
```

---

## 📞 Support

For issues or questions:
- **Documentation:** Check this guide first
- **Logs:** Check `backend/server.log` for backend errors
- **Browser Console:** Check for frontend errors

---

## ✅ What's Included

### Backend (100% Complete)
- ✅ 15 New Admin API Endpoints
- ✅ 4 New Services (Recharge, User, Affiliate, Telecom)
- ✅ Database Schema Aligned with Entities
- ✅ GORM AutoMigrate for Table Creation
- ✅ Enterprise-Grade Error Handling
- ✅ JWT Authentication
- ✅ Admin Authorization

### Frontend (100% Complete)
- ✅ Supabase Dependencies Removed
- ✅ 3 New Admin APIs Integrated
- ✅ All Components with Default Exports
- ✅ Admin Portal Fully Functional
- ✅ 20 Admin Modules Accessible

### Database (100% Complete)
- ✅ 48 Tables with Clean Names
- ✅ Schema Aligned with Entities
- ✅ Production Seed Data
- ✅ Admin User Created
- ✅ Networks & Data Plans Loaded

---

**Champion Developer**  
**February 12, 2026**
