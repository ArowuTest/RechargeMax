# RechargeMax API Documentation

**Version:** 1.0.0  
**Base URL:** `https://api.rechargemax.ng/api/v1`  
**Date:** February 1, 2026

---

## Table of Contents

1. [Authentication](#authentication)
2. [Public Endpoints](#public-endpoints)
3. [User Endpoints](#user-endpoints)
4. [Recharge Endpoints](#recharge-endpoints)
5. [Spin Endpoints](#spin-endpoints)
6. [Draw Endpoints](#draw-endpoints)
7. [Affiliate Endpoints](#affiliate-endpoints)
8. [Admin Endpoints](#admin-endpoints)
9. [Error Codes](#error-codes)
10. [Rate Limits](#rate-limits)

---

## Authentication

### OTP-Based Authentication

RechargeMax uses OTP (One-Time Password) authentication via SMS.

#### Send OTP
```http
POST /auth/send-otp
Content-Type: application/json

{
  "msisdn": "08012345678"  // or "2348012345678"
}
```

**Response:**
```json
{
  "success": true,
  "message": "OTP sent successfully",
  "expires_in": 300
}
```

**Rate Limit:** 5 requests per 5 minutes per phone number

---

#### Verify OTP
```http
POST /auth/verify-otp
Content-Type: application/json

{
  "msisdn": "08012345678",
  "otp": "123456"
}
```

**Response:**
```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "msisdn": "08012345678",
    "full_name": "John Doe",
    "email": "john@example.com"
  }
}
```

**Rate Limit:** 5 attempts per OTP

---

### Using JWT Token

Include the JWT token in the Authorization header for protected endpoints:

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

---

## Public Endpoints

### Get Networks
```http
GET /networks
```

**Response:**
```json
{
  "networks": [
    {
      "id": "uuid",
      "name": "MTN",
      "code": "MTN",
      "is_active": true
    }
  ]
}
```

---

### Get Data Plans
```http
GET /networks/:network_id/data-plans
```

**Response:**
```json
{
  "data_plans": [
    {
      "id": "uuid",
      "network_id": "uuid",
      "name": "1GB Daily",
      "volume": "1GB",
      "validity": "1 Day",
      "price": 30000,  // in kobo (₦300)
      "is_active": true
    }
  ]
}
```

---

### Health Check
```http
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2026-02-01T12:00:00Z",
  "service": "rechargemax-api",
  "version": "1.0.0"
}
```

---

### Detailed Health Check
```http
GET /health/detailed
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2026-02-01T12:00:00Z",
  "checks": {
    "database": "healthy",
    "write_capability": "healthy"
  },
  "version": "1.0.0"
}
```

---

## User Endpoints

**All user endpoints require authentication.**

### Get Dashboard
```http
GET /user/dashboard
Authorization: Bearer {token}
```

**Response:**
```json
{
  "user": {
    "id": "uuid",
    "msisdn": "08012345678",
    "full_name": "John Doe",
    "total_points": 1500,
    "total_recharge_amount": 5000000  // in kobo (₦50,000)
  },
  "wallet": {
    "balance": 100000  // in kobo (₦1,000)
  },
  "recent_transactions": [],
  "spin_eligibility": {
    "eligible": true,
    "spins_available": 2,
    "tier": "Silver"
  }
}
```

---

### Get Profile
```http
GET /user/profile
Authorization: Bearer {token}
```

**Response:**
```json
{
  "id": "uuid",
  "msisdn": "08012345678",
  "full_name": "John Doe",
  "email": "john@example.com",
  "total_points": 1500,
  "total_recharge_amount": 5000000,
  "created_at": "2026-01-01T00:00:00Z"
}
```

---

### Get Wallet
```http
GET /user/wallet
Authorization: Bearer {token}
```

**Response:**
```json
{
  "balance": 100000,  // in kobo
  "currency": "NGN",
  "last_updated": "2026-02-01T12:00:00Z"
}
```

---

### Get Transactions
```http
GET /user/transactions?page=1&limit=20
Authorization: Bearer {token}
```

**Response:**
```json
{
  "transactions": [
    {
      "id": "uuid",
      "type": "RECHARGE",
      "amount": 100000,  // in kobo
      "status": "SUCCESS",
      "phone_number": "08012345678",
      "network": "MTN",
      "created_at": "2026-02-01T12:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 50,
    "total_pages": 3
  }
}
```

---

## Recharge Endpoints

### Initiate Recharge
```http
POST /recharge/initiate
Authorization: Bearer {token}
Content-Type: application/json

{
  "phoneNumber": "08012345678",
  "networkProvider": "MTN",
  "rechargeType": "AIRTIME",  // or "DATA"
  "amount": 100000,  // in kobo (₦1,000)
  "dataBundle": "uuid",  // required if rechargeType is DATA
  "customerEmail": "john@example.com",
  "customerName": "John Doe"
}
```

**Response:**
```json
{
  "success": true,
  "transaction_id": "uuid",
  "payment_reference": "REF-123456789",
  "paymentUrl": "https://checkout.paystack.com/...",
  "amount": 100000
}
```

---

### Get Recharge History
```http
GET /recharge/history?page=1&limit=20
Authorization: Bearer {token}
```

**Response:**
```json
{
  "recharges": [
    {
      "id": "uuid",
      "phone_number": "08012345678",
      "network": "MTN",
      "type": "AIRTIME",
      "amount": 100000,
      "status": "SUCCESS",
      "payment_reference": "REF-123456789",
      "created_at": "2026-02-01T12:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 50
  }
}
```

---

## Spin Endpoints

### Check Spin Eligibility
```http
GET /spin/eligibility
Authorization: Bearer {token}
```

**Response:**
```json
{
  "eligible": true,
  "spins_available": 2,
  "tier": "Silver",
  "message": "You have 2 spins available!",
  "total_recharge_today": 500000,  // in kobo
  "min_recharge_required": 100000
}
```

---

### Play Spin
```http
POST /spin/play
Authorization: Bearer {token}
```

**Response:**
```json
{
  "success": true,
  "spin_id": "uuid",
  "prize": {
    "id": "uuid",
    "name": "₦100 Airtime",
    "type": "AIRTIME",
    "value": 10000,  // in kobo
    "claim_status": "PROVISIONED"
  },
  "message": "Congratulations! You won ₦100 Airtime"
}
```

**Rate Limit:** 10 requests per minute

---

### Get Spin History
```http
GET /spin/history?page=1&limit=20
Authorization: Bearer {token}
```

**Response:**
```json
{
  "spins": [
    {
      "id": "uuid",
      "prize_name": "₦100 Airtime",
      "prize_type": "AIRTIME",
      "prize_value": 10000,
      "claim_status": "PROVISIONED",
      "created_at": "2026-02-01T12:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 10
  }
}
```

---

## Draw Endpoints

### Get Active Draws
```http
GET /draws/active
Authorization: Bearer {token}
```

**Response:**
```json
{
  "draws": [
    {
      "id": "uuid",
      "title": "Monthly Mega Draw",
      "description": "Win amazing prizes!",
      "draw_date": "2026-03-01T00:00:00Z",
      "status": "ACTIVE",
      "total_entries": 1000,
      "my_entries": 5
    }
  ]
}
```

---

### Get My Draw Entries
```http
GET /draws/my-entries
Authorization: Bearer {token}
```

**Response:**
```json
{
  "entries": [
    {
      "draw_id": "uuid",
      "draw_title": "Monthly Mega Draw",
      "entries_count": 5,
      "entry_date": "2026-02-01T12:00:00Z"
    }
  ]
}
```

---

## Affiliate Endpoints

### Get Referral Code
```http
GET /affiliate/code
Authorization: Bearer {token}
```

**Response:**
```json
{
  "referral_code": "JOHN123",
  "referral_url": "https://rechargemax.ng?ref=JOHN123"
}
```

---

### Get Affiliate Stats
```http
GET /affiliate/stats
Authorization: Bearer {token}
```

**Response:**
```json
{
  "total_referrals": 10,
  "active_referrals": 8,
  "total_commission": 50000,  // in kobo
  "pending_payout": 30000,
  "lifetime_earnings": 200000
}
```

---

### Get Referrals
```http
GET /affiliate/referrals?page=1&limit=20
Authorization: Bearer {token}
```

**Response:**
```json
{
  "referrals": [
    {
      "user_id": "uuid",
      "full_name": "Jane Doe",
      "joined_date": "2026-01-15T00:00:00Z",
      "total_recharge": 100000,
      "commission_earned": 5000,
      "status": "ACTIVE"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 10
  }
}
```

---

## Admin Endpoints

**All admin endpoints require admin authentication with role verification.**

### Get Dashboard Stats
```http
GET /admin/dashboard
Authorization: Bearer {admin_token}
```

**Response:**
```json
{
  "total_users": 10000,
  "total_transactions": 50000,
  "total_revenue": 5000000000,  // in kobo
  "active_draws": 2,
  "pending_withdrawals": 5,
  "today_stats": {
    "new_users": 50,
    "transactions": 500,
    "revenue": 50000000
  }
}
```

---

### Get Users
```http
GET /admin/users?page=1&limit=50&search=john
Authorization: Bearer {admin_token}
```

**Response:**
```json
{
  "users": [
    {
      "id": "uuid",
      "msisdn": "08012345678",
      "full_name": "John Doe",
      "email": "john@example.com",
      "total_recharge": 500000,
      "status": "ACTIVE",
      "created_at": "2026-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 50,
    "total": 10000
  }
}
```

---

### Create Draw
```http
POST /admin/draws
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "title": "Monthly Mega Draw",
  "description": "Win amazing prizes!",
  "draw_date": "2026-03-01T00:00:00Z",
  "status": "ACTIVE"
}
```

**Response:**
```json
{
  "success": true,
  "draw_id": "uuid",
  "message": "Draw created successfully"
}
```

---

### Get Wheel Prizes
```http
GET /admin/wheel-prizes
Authorization: Bearer {admin_token}
```

**Response:**
```json
{
  "prizes": [
    {
      "id": "uuid",
      "name": "₦100 Airtime",
      "type": "AIRTIME",
      "value": 10000,
      "probability": 30.0,
      "is_active": true
    }
  ]
}
```

---

## Error Codes

### HTTP Status Codes

- `200` - Success
- `201` - Created
- `400` - Bad Request (validation error)
- `401` - Unauthorized (missing or invalid token)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found
- `429` - Too Many Requests (rate limit exceeded)
- `500` - Internal Server Error

### Error Response Format

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {}
}
```

### Common Error Codes

- `INVALID_OTP` - OTP is incorrect
- `OTP_EXPIRED` - OTP has expired (5 minutes)
- `OTP_LOCKED` - Too many failed attempts (15 minute lockout)
- `INSUFFICIENT_BALANCE` - Wallet balance too low
- `NOT_ELIGIBLE_TO_SPIN` - User hasn't met spin requirements
- `INVALID_PHONE_NUMBER` - Phone number format invalid
- `INVALID_AMOUNT` - Amount outside valid range
- `PAYMENT_FAILED` - Payment processing failed
- `NETWORK_ERROR` - External network provider error

---

## Rate Limits

### Global Rate Limits
- **Default:** 100 requests per minute per IP
- **Authenticated:** 200 requests per minute per user

### Endpoint-Specific Limits
- **Send OTP:** 5 requests per 5 minutes per phone number
- **Verify OTP:** 5 attempts per OTP
- **Play Spin:** 10 requests per minute per user
- **Initiate Recharge:** 20 requests per minute per user

### Rate Limit Headers

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1643724000
```

---

## CSRF Protection

For state-changing operations (POST, PUT, DELETE), include CSRF token:

### Get CSRF Token
```http
GET /csrf-token
Authorization: Bearer {token}
```

**Response:**
```json
{
  "csrf_token": "abc123...",
  "expires_in": 3600
}
```

### Use CSRF Token
```http
POST /recharge/initiate
Authorization: Bearer {token}
X-CSRF-Token: abc123...
Content-Type: application/json
```

---

## Pagination

All list endpoints support pagination:

**Query Parameters:**
- `page` - Page number (default: 1)
- `limit` - Items per page (default: 20, max: 100)

**Response Format:**
```json
{
  "data": [],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

---

## Amount Format

**All amounts are in kobo (1/100 of Naira):**

- ₦1 = 100 kobo
- ₦10 = 1,000 kobo
- ₦100 = 10,000 kobo
- ₦1,000 = 100,000 kobo

**Example:**
```json
{
  "amount": 100000  // This is ₦1,000
}
```

---

## Phone Number Format

**Accepted formats:**
- Local: `08012345678` (11 digits)
- International: `2348012345678` (13 digits)
- With plus: `+2348012345678`

**All formats are normalized to local format (08012345678) for storage.**

---

## Webhooks

### Payment Webhook (Paystack)
```http
POST /payment/webhook
Content-Type: application/json
X-Paystack-Signature: {signature}

{
  "event": "charge.success",
  "data": {
    "reference": "REF-123456789",
    "amount": 100000,
    "status": "success"
  }
}
```

**Note:** Webhook signature is verified for security. Idempotency is enforced to prevent duplicate processing.

---

## Support

For API support, contact:
- **Email:** api@rechargemax.ng
- **Documentation:** https://docs.rechargemax.ng
- **Status Page:** https://status.rechargemax.ng

---

**Last Updated:** February 1, 2026  
**API Version:** 1.0.0
