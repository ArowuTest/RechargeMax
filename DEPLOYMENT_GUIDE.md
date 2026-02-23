# 🚀 RechargeMax Deployment Guide

## Production Deployment to Vercel (Frontend) + Render (Backend + Database)

This guide provides step-by-step instructions for deploying the RechargeMax Rewards Platform to production using industry-standard hosting providers.

---

## 📋 Table of Contents

1. [Prerequisites](#prerequisites)
2. [Database Setup (Render PostgreSQL)](#database-setup)
3. [Backend Deployment (Render)](#backend-deployment)
4. [Frontend Deployment (Vercel)](#frontend-deployment)
5. [Post-Deployment Verification](#post-deployment-verification)
6. [Troubleshooting](#troubleshooting)

---

## ✅ Prerequisites

### Required Accounts
- ✅ GitHub account with repository access
- ✅ [Render.com](https://render.com) account (for backend + database)
- ✅ [Vercel.com](https://vercel.com) account (for frontend)
- ✅ VTPass account with API credentials (for airtime/data provisioning)
- ✅ Paystack account with API keys (for payments)

### Required Files
- ✅ `.env` file with all environment variables
- ✅ Database migrations in `backend/migrations/`
- ✅ Frontend build configuration

---

## 🗄️ Database Setup (Render PostgreSQL)

### Step 1: Create PostgreSQL Database

1. Log in to [Render Dashboard](https://dashboard.render.com)
2. Click **"New +"** → **"PostgreSQL"**
3. Configure database:
   ```
   Name: rechargemax-db
   Database: rechargemax_db
   User: rechargemax (auto-generated)
   Region: Choose closest to your users
   Plan: Starter ($7/month) or higher
   ```
4. Click **"Create Database"**
5. Wait for provisioning (2-3 minutes)

### Step 2: Get Database Connection Details

After creation, you'll see:
```
Internal Database URL: postgresql://rechargemax:xxxxx@dpg-xxxxx/rechargemax_db
External Database URL: postgresql://rechargemax:xxxxx@oregon-postgres.render.com/rechargemax_db
```

**Save both URLs** - you'll need them for:
- **Internal URL**: For backend service on Render
- **External URL**: For local development and migrations

### Step 3: Run Database Migrations

**Option A: From Local Machine**
```bash
# Set environment variable
export DATABASE_URL="<External Database URL from Render>"

# Navigate to backend directory
cd backend

# Run migrations in order
for file in migrations/*.sql; do
    echo "Running $file..."
    psql $DATABASE_URL -f "$file"
done

# Run comprehensive permissions migration
psql $DATABASE_URL -f migrations/999_grant_all_permissions.sql
```

**Option B: Using Render Shell**
```bash
# In Render Dashboard, go to your database
# Click "Connect" → "External Connection"
# Use psql command provided

# Upload migrations to a temporary location
# Run each migration file
```

### Step 4: Verify Database Setup

```bash
# Connect to database
psql $DATABASE_URL

# Check tables
\dt

# Verify permissions
SELECT grantee, privilege_type 
FROM information_schema.role_table_grants 
WHERE table_name='users' AND grantee='rechargemax';

# Should show: SELECT, INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER
```

---

## 🖥️ Backend Deployment (Render)

### Step 1: Create Web Service

1. In Render Dashboard, click **"New +"** → **"Web Service"**
2. Connect your GitHub repository
3. Configure service:
   ```
   Name: rechargemax-backend
   Region: Same as database
   Branch: main
   Root Directory: backend
   Runtime: Go
   Build Command: go build -o backend cmd/server/main.go
   Start Command: ./backend
   Plan: Starter ($7/month) or higher
   ```

### Step 2: Configure Environment Variables

In the "Environment" tab, add:

```bash
# Database
DATABASE_URL=<Internal Database URL from Step 1>
DB_HOST=<from Render database>
DB_PORT=5432
DB_USER=rechargemax
DB_PASSWORD=<from Render database>
DB_NAME=rechargemax_db

# Server
PORT=8080
GIN_MODE=release
ENVIRONMENT=production

# JWT
JWT_SECRET=<generate strong secret: openssl rand -base64 32>
JWT_EXPIRY=24h

# VTPass (Airtime/Data Provider)
VTPASS_API_KEY=<your VTPass API key>
VTPASS_PUBLIC_KEY=<your VTPass public key>
VTPASS_SECRET_KEY=<your VTPass secret key>
VTPASS_MODE=LIVE
VTPASS_BASE_URL=https://api-service.vtpass.com/api

# Paystack (Payments)
PAYSTACK_SECRET_KEY=<your Paystack secret key>
PAYSTACK_PUBLIC_KEY=<your Paystack public key>

# SMS (Optional - for OTP)
SMS_PROVIDER=vtpass
SMS_SENDER_ID=RechargeMax

# CORS
CORS_ALLOWED_ORIGINS=https://your-frontend.vercel.app

# Admin
ADMIN_DEFAULT_EMAIL=admin@rechargemax.com
ADMIN_DEFAULT_PASSWORD=<generate strong password>

# Feature Flags
ENABLE_SPIN_WHEEL=true
ENABLE_DAILY_DRAW=true
ENABLE_AFFILIATE=true
ENABLE_SUBSCRIPTIONS=true
```

### Step 3: Deploy

1. Click **"Create Web Service"**
2. Render will automatically:
   - Clone your repository
   - Run build command
   - Start your application
3. Monitor logs for any errors
4. Wait for "Live" status (3-5 minutes)

### Step 4: Get Backend URL

After deployment:
```
Your backend URL: https://rechargemax-backend.onrender.com
```

**Test it:**
```bash
curl https://rechargemax-backend.onrender.com/health
# Should return: {"status":"healthy"}
```

---

## 🌐 Frontend Deployment (Vercel)

### Step 1: Prepare Frontend

1. Update API base URL in frontend code:
   ```typescript
   // frontend/src/config/api.ts
   export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 
     'https://rechargemax-backend.onrender.com';
   ```

2. Commit and push changes:
   ```bash
   git add frontend/src/config/api.ts
   git commit -m "chore: Update API URL for production"
   git push origin main
   ```

### Step 2: Deploy to Vercel

1. Log in to [Vercel Dashboard](https://vercel.com/dashboard)
2. Click **"Add New..."** → **"Project"**
3. Import your GitHub repository
4. Configure project:
   ```
   Framework Preset: Vite
   Root Directory: frontend
   Build Command: npm run build
   Output Directory: dist
   Install Command: npm install
   ```

### Step 3: Configure Environment Variables

In Vercel project settings, add:

```bash
VITE_API_BASE_URL=https://rechargemax-backend.onrender.com
VITE_PAYSTACK_PUBLIC_KEY=<your Paystack public key>
VITE_ENVIRONMENT=production
```

### Step 4: Deploy

1. Click **"Deploy"**
2. Vercel will automatically:
   - Install dependencies
   - Build your application
   - Deploy to CDN
3. Wait for "Ready" status (2-3 minutes)

### Step 5: Get Frontend URL

After deployment:
```
Your frontend URL: https://rechargemax.vercel.app
```

### Step 6: Configure Custom Domain (Optional)

1. In Vercel project settings, go to **"Domains"**
2. Add your custom domain (e.g., `app.rechargemax.com`)
3. Update DNS records as instructed
4. Wait for SSL certificate (automatic)

---

## ✅ Post-Deployment Verification

### 1. Test Backend Health

```bash
curl https://rechargemax-backend.onrender.com/health
# Expected: {"status":"healthy","timestamp":"..."}
```

### 2. Test Database Connection

```bash
curl https://rechargemax-backend.onrender.com/api/v1/health/db
# Expected: {"database":"connected"}
```

### 3. Test Frontend

1. Open `https://rechargemax.vercel.app`
2. Verify homepage loads
3. Test user registration/login
4. Test recharge flow
5. Test spin wheel
6. Test prize claiming

### 4. Test VTPass Integration

1. Make a test recharge (₦100)
2. Verify transaction success
3. Check if airtime/data was delivered
4. Verify in VTPass dashboard

### 5. Test Paystack Integration

1. Make a test payment
2. Verify payment success
3. Check Paystack dashboard for transaction

### 6. Monitor Logs

**Backend Logs (Render):**
```
Dashboard → rechargemax-backend → Logs
```

**Frontend Logs (Vercel):**
```
Dashboard → rechargemax → Deployments → View Function Logs
```

---

## 🔧 Troubleshooting

### Issue: Database Connection Failed

**Symptoms:**
```
ERROR: connection refused
ERROR: permission denied for table users
```

**Solution:**
```bash
# 1. Verify DATABASE_URL is correct
echo $DATABASE_URL

# 2. Run permissions migration
psql $DATABASE_URL -f backend/migrations/999_grant_all_permissions.sql

# 3. Restart backend service in Render
```

### Issue: CORS Errors

**Symptoms:**
```
Access to fetch at 'https://backend.onrender.com' from origin 'https://frontend.vercel.app' 
has been blocked by CORS policy
```

**Solution:**
```bash
# Update CORS_ALLOWED_ORIGINS in Render environment variables
CORS_ALLOWED_ORIGINS=https://rechargemax.vercel.app,https://rechargemax-git-main.vercel.app

# Restart backend service
```

### Issue: VTPass Provisioning Failed

**Symptoms:**
```
Spin wheel works but airtime/data not delivered
```

**Solution:**
```bash
# 1. Verify VTPass credentials in Render environment variables
# 2. Check VTPass mode (LIVE vs TEST)
# 3. Verify VTPass account has sufficient balance
# 4. Check backend logs for VTPass API errors
```

### Issue: Frontend Build Failed

**Symptoms:**
```
Build failed: Module not found
```

**Solution:**
```bash
# 1. Verify all dependencies in package.json
# 2. Clear Vercel cache and redeploy
# 3. Check Node.js version compatibility
```

### Issue: Slow Backend Response

**Symptoms:**
```
Requests taking > 5 seconds
```

**Solution:**
```bash
# 1. Upgrade Render plan (more RAM/CPU)
# 2. Add database connection pooling
# 3. Enable Redis caching
# 4. Optimize database queries
```

---

## 📊 Monitoring & Maintenance

### Daily Checks

- ✅ Check error logs in Render and Vercel
- ✅ Monitor VTPass balance
- ✅ Verify payment transactions in Paystack
- ✅ Check database disk usage

### Weekly Tasks

- ✅ Review application metrics
- ✅ Backup database
- ✅ Update dependencies
- ✅ Review security alerts

### Monthly Tasks

- ✅ Analyze user growth
- ✅ Optimize database performance
- ✅ Review and update pricing
- ✅ Plan new features

---

## 🔐 Security Best Practices

1. **Never commit secrets to Git**
   - Use environment variables
   - Add `.env` to `.gitignore`

2. **Use strong passwords**
   - JWT_SECRET: 32+ characters
   - Admin password: 16+ characters
   - Database password: Auto-generated by Render

3. **Enable HTTPS only**
   - Vercel: Automatic
   - Render: Automatic

4. **Regular updates**
   - Update Go dependencies monthly
   - Update npm packages weekly
   - Apply security patches immediately

5. **Monitor access logs**
   - Review failed login attempts
   - Check for suspicious API calls
   - Monitor database queries

---

## 📞 Support

**Issues with deployment?**
- GitHub Issues: https://github.com/ArowuTest/RechargeMax/issues
- Email: support@rechargemax.com

**Platform Support:**
- Render: https://render.com/docs
- Vercel: https://vercel.com/docs
- VTPass: support@vtpass.com
- Paystack: support@paystack.com

---

## 🎉 Success!

Your RechargeMax platform is now live in production!

**Next Steps:**
1. ✅ Test all features thoroughly
2. ✅ Set up monitoring and alerts
3. ✅ Configure backups
4. ✅ Launch marketing campaign
5. ✅ Onboard first users

**Your URLs:**
- Frontend: https://rechargemax.vercel.app
- Backend: https://rechargemax-backend.onrender.com
- Database: Managed by Render

---

**Built with ❤️ for scale. Ready for 50 million users!** 🚀
