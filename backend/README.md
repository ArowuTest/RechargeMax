# RechargeMax Backend - Complete Go Application

## 📦 Package Contents

This package contains the complete RechargeMax backend built with Go, following clean architecture principles.

### File Structure
```
backend/
├── cmd/server/main.go              # Application entry point
├── internal/
│   ├── domain/
│   │   ├── entities/               # 30 GORM entity models
│   │   └── repositories/           # 11 repository interfaces
│   ├── application/
│   │   └── services/               # Business logic services
│   ├── infrastructure/
│   │   └── persistence/            # 11 GORM repository implementations
│   └── presentation/
│       ├── handlers/               # HTTP handlers
│       └── middleware/             # Auth, CORS, logging middleware
├── migrations/                     # 12 SQL migration files
├── go.mod                          # Go module definition
└── go.sum                          # Dependency checksums
```

## 🗄️ Database Schema

**30 Tables Total:**
- 19 core business tables
- 4 notification system tables
- 1 file storage table
- 6 additional tables (OTP, admin logs, affiliate payouts, etc.)

**All tables use suffix:** `_2026_01_30_14_00`

### Migration Files (12 total)
1. `01_core_tables_schema_2026_01_30_14_00.sql` - Core 19 tables
2. `02_rls_policies_2026_01_30_14_00.sql` - Row Level Security
3. `03_seeded_data_2026_01_30_14_00.sql` - Initial data
4. `04_functions_triggers_2026_01_30_14_00.sql` - Database functions
5. `05_storage_buckets_2026_01_30_14_00.sql` - File storage
6. `06_notification_system_2026_01_30_14_00.sql` - Notifications
7. `07_notification_templates_seed_fixed_2026_01_30_14_00.sql` - Templates
8. `08_otp_verifications_2026_01_30_14_00.sql` - OTP system
9. `09_admin_activity_logs_2026_01_30_14_00.sql` - Admin logging
10. `10_affiliate_payouts_2026_01_30_14_00.sql` - Affiliate payouts
11. `11_affiliate_analytics_bank_2026_01_30_14_00.sql` - Analytics
12. `12_payment_logs_2026_01_30_14_00.sql` - Payment logging

## 🚀 Quick Start

### Prerequisites
- Go 1.21 or higher
- PostgreSQL 14+ database
- Environment variables configured

### Installation

1. **Extract the ZIP file**
   ```bash
   unzip rechargemax-backend-windows-*.zip
   cd rechargemax-rebuild/backend
   ```

2. **Install Go dependencies**
   ```bash
   go mod download
   go mod tidy
   ```

3. **Set up environment variables**
   Create `.env` file:
   ```env
   DATABASE_URL=postgresql://user:pass@localhost:5432/rechargemax
   PORT=8080
   ENV=production
   JWT_SECRET=your-secret-key
   PAYSTACK_SECRET_KEY=sk_live_xxx
   PAYSTACK_PUBLIC_KEY=pk_live_xxx
   ```

4. **Run database migrations**
   ```bash
   psql -U postgres -d rechargemax -f migrations/01_core_tables_schema_2026_01_30_14_00.sql
   psql -U postgres -d rechargemax -f migrations/02_rls_policies_2026_01_30_14_00.sql
   # ... run all 12 migration files in order
   ```

5. **Build the application**
   ```bash
   go build -o rechargemax cmd/server/main.go
   ```

6. **Run the server**
   ```bash
   ./rechargemax
   ```

   Server will start on `http://localhost:8080`

## 📡 API Endpoints

### Public Endpoints
- `GET /` - API info
- `GET /health` - Health check
- `POST /api/v1/auth/send-otp` - Send OTP
- `POST /api/v1/auth/verify-otp` - Verify OTP & login
- `GET /api/v1/user/profile` - Get user profile (auth required)

### Admin Endpoints
- `GET /api/admin/dashboard` - Admin dashboard (admin auth required)

## 🏗️ Architecture

### Clean Architecture Layers
1. **Domain Layer** - Entities and repository interfaces
2. **Application Layer** - Business logic services
3. **Infrastructure Layer** - GORM implementations, external APIs
4. **Presentation Layer** - HTTP handlers and middleware

### Repository Pattern
All data access goes through repository interfaces for testability.

### Dependency Injection
Constructor-based dependency injection throughout.

## 🔧 Development

### Testing
```bash
go test ./...
go test -cover ./...
```

## 📦 Dependencies

- **Gin** - HTTP web framework
- **GORM** - ORM library
- **PostgreSQL Driver** - Database driver
- **UUID** - UUID generation

## 🔐 Security Features

- OTP-based authentication
- JWT token authentication (to be completed)
- SQL injection prevention
- CORS middleware
- Rate limiting on OTP requests

## 🌐 Direct Telecom Integration

Designed for direct integration with Nigerian telecom providers:
- MTN, Airtel, Glo, 9mobile

## 🚀 Production Deployment

```bash
# Build optimized binary
go build -ldflags="-s -w" -o rechargemax cmd/server/main.go
```

## 📝 TODO / Next Steps

### High Priority
- [ ] Complete JWT token generation and validation
- [ ] Implement remaining HTTP handlers
- [ ] Add Paystack payment integration
- [ ] Implement direct telecom provider APIs
- [ ] Add SMS provider integration

### Medium Priority
- [ ] Add comprehensive unit tests
- [ ] Implement background jobs
- [ ] Add request validation
- [ ] Add structured logging

---

**Generated:** 2026-01-30
**Version:** 1.0.0
**Go Version:** 1.21+
**Database:** PostgreSQL 14+
