# RechargeMax Platform - Deployment Guide
**Version:** Fixed Build (Feb 16, 2026)  
**Status:** Production Ready

## Quick Start

### 1. Extract Archive
```bash
tar -xzf RechargeMax_Fixed_20260216_015656.tar.gz
cd RechargeMax_Clean
```

### 2. Database Setup
```bash
# Create database
createdb rechargemax_db

# Run migrations (if using migration tool)
# OR manually run the schema SQL files

# Apply critical fix
psql -d rechargemax_db -c "ALTER TABLE draw_winners ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id);"
```

### 3. Backend Setup
```bash
cd backend

# Install dependencies
go mod download

# Configure environment
cp .env.example .env
# Edit .env with your credentials

# Required environment variables:
# BACKEND_URL=http://localhost:8080
# FRONTEND_URL=http://localhost:5173
# DATABASE_URL=postgresql://user:pass@localhost:5432/rechargemax_db
# PAYSTACK_SECRET_KEY=sk_test_...
# VTPASS_API_KEY=...
# VTPASS_PUBLIC_KEY=...

# Run backend
go run cmd/server/main.go
```

### 4. Frontend Setup
```bash
cd frontend

# Install dependencies
npm install

# Configure environment
cp .env.example .env
# Edit .env with your API URL

# Required environment variables:
# VITE_API_URL=http://localhost:8080/api/v1

# Run frontend
npm run dev
```

## Critical Fixes Included

1. **Paystack Callback Fix** - Redirects to backend for VTU processing
2. **Toast Data Fix** - Shows real transaction amounts and points
3. **Winners Endpoint Fix** - Added user_id column to draw_winners table

## Testing

Test a recharge with:
- Amount: ₦1,500
- Network: MTN
- Phone: 08011111111

Expected: Success popup with correct amount (₦1,500) and points (7)

## Documentation

- `FIXES_APPLIED.md` - Detailed fix documentation
- `CHANGELOG_Feb16_2026.md` - Complete changelog

## Support

For issues, check the backend logs and database status.
