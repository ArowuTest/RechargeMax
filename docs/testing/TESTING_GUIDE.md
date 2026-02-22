# RechargeMax Testing Guide

**Date:** February 20, 2026  
**Status:** Ready for Testing  
**Environment:** Development (Paystack Test Mode, VTPass Sandbox)

---

## Prerequisites

### Backend Running
```bash
cd /home/ubuntu/RechargeMax_Clean/backend
go run cmd/api/main.go
```
**Port:** 8080  
**Health Check:** http://localhost:8080/health

### Frontend Running
```bash
cd /home/ubuntu/RechargeMax_Clean/frontend
pnpm run dev
```
**Port:** 5173  
**URL:** http://localhost:5173

### Database
- **Host:** localhost:5432
- **Database:** rechargemax_db
- **User:** rechargemax_user
- **All migrations applied:** 36 migrations

---

## Test Scenarios

### 1. User Registration & Login

**Test Case 1.1: New User Registration**

1. Navigate to http://localhost:5173/register
2. Fill in:
   - Full Name: "Test User"
   - Email: "testuser@example.com"
   - Phone: "08012345678"
   - Password: "Test123!"
3. Click "Register"
4. **Expected:** User created, redirected to dashboard

**Test Case 1.2: User Login**

1. Navigate to http://localhost:5173/login
2. Enter credentials from 1.1
3. Click "Login"
4. **Expected:** Redirected to dashboard with user info

---

### 2. Recharge Flow (Core Feature)

**Test Case 2.1: Small Recharge (No Spin)**

1. Login to dashboard
2. Click "Recharge" or "Buy Airtime"
3. Fill in:
   - Phone: "08011111111" (VTPass test number)
   - Network: MTN
   - Amount: ₦500
4. Click "Proceed to Payment"
5. **Expected:**
   - Redirected to Paystack payment page
   - Amount shown: ₦500
6. Complete payment (use Paystack test card)
7. **Expected:**
   - Redirected back to dashboard
   - Transaction status: SUCCESS
   - Points earned: 2 points (₦500 / ₦200 = 2.5 → 2)
   - Draw entries: 2 entries
   - Spin eligible: NO (minimum ₦1,000)

**Verification:**
```sql
SELECT id, user_id, amount, status, points_earned, draw_entries, spin_eligible
FROM transactions
WHERE phone_number = '08011111111'
ORDER BY created_at DESC
LIMIT 1;
```

**Test Case 2.2: Large Recharge (Spin Eligible)**

1. Login to dashboard
2. Create recharge:
   - Phone: "08011111111"
   - Network: MTN
   - Amount: ₦2,000
3. Complete Paystack payment
4. **Expected:**
   - Transaction status: SUCCESS
   - Points earned: 10 points (₦2,000 / ₦200 = 10)
   - Draw entries: 10 entries
   - Spin eligible: YES (₦2,000 >= ₦1,000)
   - "Spin the Wheel" button appears

**Verification:**
```sql
SELECT amount, points_earned, draw_entries, spin_eligible
FROM transactions
WHERE phone_number = '08011111111' AND amount = 200000
ORDER BY created_at DESC
LIMIT 1;
```

**Test Case 2.3: Edge Case - Exactly ₦1,000**

1. Create recharge for ₦1,000
2. **Expected:**
   - Points: 5 points
   - Draw entries: 5 entries
   - Spin eligible: YES (exactly ₦1,000)

---

### 3. Points Calculation Verification

**Test Matrix:**

| Amount | Amount (kobo) | Points Expected | Draw Entries | Spin Eligible |
|--------|--------------|----------------|--------------|---------------|
| ₦100 | 10,000 | 0 | 0 | NO |
| ₦200 | 20,000 | 1 | 1 | NO |
| ₦399 | 39,900 | 1 | 1 | NO |
| ₦400 | 40,000 | 2 | 2 | NO |
| ₦500 | 50,000 | 2 | 2 | NO |
| ₦1,000 | 100,000 | 5 | 5 | YES |
| ₦2,000 | 200,000 | 10 | 10 | YES |
| ₦5,000 | 500,000 | 25 | 25 | YES |

**Formula:**
- Points = `amount_in_kobo / 20000` (rounded down)
- Draw Entries = Points (1:1 ratio)
- Spin Eligible = `amount >= 100000` (₦1,000)

---

### 4. Spin Wheel Functionality

**Test Case 4.1: Get Spin Tiers**

**API Call:**
```bash
curl http://localhost:8080/api/v1/spins/tiers
```

**Expected Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "Bronze",
      "min_amount": 100000,
      "max_amount": 499999,
      "prizes": [...],
      "created_at": "..."
    },
    {
      "id": 2,
      "name": "Silver",
      "min_amount": 500000,
      "max_amount": 999999,
      ...
    },
    ...
  ]
}
```

**Test Case 4.2: Get User Tier Progress**

**API Call:**
```bash
curl -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8080/api/v1/spins/tier-progress
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "current_tier": "Bronze",
    "current_tier_id": 1,
    "total_recharged": 200000,
    "next_tier": "Silver",
    "amount_to_next_tier": 300000,
    "progress_percentage": 40
  }
}
```

**Test Case 4.3: Play Spin Wheel**

1. Complete a ₦1,000+ recharge
2. Click "Spin the Wheel" button
3. **Expected:**
   - Wheel animation plays
   - Prize is awarded
   - Prize saved to database
   - Spin count decremented

**API Call:**
```bash
curl -X POST \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"transaction_id": 123}' \
  http://localhost:8080/api/v1/spin/play
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "prize": "₦500 Airtime",
    "prize_value": 50000,
    "tier": "Bronze",
    "spin_id": 456
  }
}
```

---

### 5. Webhook Testing

**Test Case 5.1: Paystack Webhook**

**Simulate Webhook:**
```bash
curl -X POST http://localhost:8080/api/v1/webhooks/paystack \
  -H "Content-Type: application/json" \
  -H "x-paystack-signature: <SIGNATURE>" \
  -d '{
    "event": "charge.success",
    "data": {
      "reference": "RCH_1111_1771623126",
      "amount": 200000,
      "status": "success"
    }
  }'
```

**Expected:**
1. Backend logs: "Processing Paystack webhook"
2. Transaction updated to SUCCESS
3. VTPass API called for recharge
4. Points calculated and saved

**Verification:**
```bash
# Check backend logs
tail -f /home/ubuntu/RechargeMax_Clean/backend/logs/app.log

# Check VTPass dashboard
# Visit: https://sandbox.vtpass.com/transactions
```

---

### 6. VTPass Integration

**Test Case 6.1: Airtime Recharge**

1. Create recharge for MTN airtime
2. Complete payment
3. **Expected:**
   - VTPass API called with:
     - request_id: RCH_1111_xxx
     - serviceID: mtn
     - amount: 500
     - phone: 08011111111
   - VTPass response: "000" (success)
   - Transaction status: SUCCESS

**Test Case 6.2: Data Bundle**

1. Create recharge for data bundle
2. Select bundle (e.g., "1GB - ₦500")
3. Complete payment
4. **Expected:**
   - VTPass API called with variation_code
   - Transaction successful

---

### 7. Dashboard Features

**Test Case 7.1: Transaction History**

1. Login to dashboard
2. Navigate to "Transaction History"
3. **Expected:**
   - All user transactions listed
   - Columns: Date, Amount, Points, Status
   - Filter by status (SUCCESS, PENDING, FAILED)

**Test Case 7.2: Points Balance**

1. Dashboard shows total points
2. **Expected:**
   - Sum of all points_earned from SUCCESS transactions
   - Matches database query:
     ```sql
     SELECT SUM(points_earned) FROM transactions
     WHERE user_id = <USER_ID> AND status = 'SUCCESS';
     ```

**Test Case 7.3: Prize Claims**

1. Navigate to "My Prizes"
2. **Expected:**
   - All won prizes listed
   - Status: PENDING, CLAIMED, EXPIRED
   - "Claim" button for PENDING prizes

---

### 8. Affiliate Program

**Test Case 8.1: Generate Referral Link**

1. Dashboard → "Referral Program"
2. **Expected:**
   - Unique referral code shown
   - Referral link: `https://rechargemax.com/register?ref=ABC123`

**Test Case 8.2: Referral Commission**

1. User A refers User B
2. User B registers with referral code
3. User B makes ₦1,000 recharge
4. **Expected:**
   - User A earns commission (e.g., 5% = ₦50)
   - Commission added to User A's wallet

---

### 9. Daily Subscription

**Test Case 9.1: Subscribe via Airtime (MTN)**

1. User with MTN number clicks "Daily Draw"
2. Selects "Subscribe - ₦20/day"
3. **Expected:**
   - USSD prompt: *Dial *461*4*<CODE>#*
   - After dialing: Subscription active
   - 1 draw entry added daily

**Test Case 9.2: Subscribe via Paystack (Non-MTN)**

1. User with Airtel number subscribes
2. **Expected:**
   - Redirected to Paystack for ₦20 payment
   - After payment: Subscription active
   - 1 draw entry added daily

---

## Automated Testing

### Unit Tests

```bash
cd /home/ubuntu/RechargeMax_Clean/backend
go test ./internal/services/... -v
```

**Key Tests:**
- `TestCalculatePoints` - Points calculation
- `TestProcessSuccessfulPayment` - Payment processing
- `TestGetTierProgress` - Spin tier logic

### Integration Tests

```bash
go test ./internal/handlers/... -v
```

**Key Tests:**
- `TestRechargeFlow` - End-to-end recharge
- `TestWebhookProcessing` - Paystack webhook
- `TestSpinWheel` - Spin functionality

---

## Database Verification Queries

### Check Points Calculation

```sql
SELECT 
  id,
  amount,
  amount / 20000 AS calculated_points,
  points_earned,
  CASE 
    WHEN amount / 20000 = points_earned THEN '✅ CORRECT'
    ELSE '❌ MISMATCH'
  END AS verification
FROM transactions
WHERE status = 'SUCCESS'
ORDER BY created_at DESC
LIMIT 10;
```

### Check Spin Eligibility

```sql
SELECT 
  id,
  amount,
  spin_eligible,
  CASE 
    WHEN amount >= 100000 AND spin_eligible = true THEN '✅ CORRECT'
    WHEN amount < 100000 AND spin_eligible = false THEN '✅ CORRECT'
    ELSE '❌ MISMATCH'
  END AS verification
FROM transactions
WHERE status = 'SUCCESS'
ORDER BY created_at DESC;
```

### Check User Points Balance

```sql
SELECT 
  u.id,
  u.full_name,
  u.email,
  SUM(t.points_earned) AS total_points,
  COUNT(t.id) AS total_transactions
FROM users u
LEFT JOIN transactions t ON u.id = t.user_id AND t.status = 'SUCCESS'
GROUP BY u.id
ORDER BY total_points DESC;
```

---

## Known Issues & Workarounds

### Issue 1: Paystack Test Cards

**Problem:** Some test cards fail in sandbox  
**Workaround:** Use official Paystack test card:
- Card: 4084 0840 8408 4081
- Expiry: Any future date
- CVV: 408
- PIN: 0000
- OTP: 123456

### Issue 2: VTPass Sandbox Limits

**Problem:** Sandbox has transaction limits  
**Workaround:** Use test phone: 08011111111

---

## Success Criteria

✅ All test scenarios pass  
✅ Points calculated correctly (₦200 = 1 point)  
✅ Spin wheel activates for ₦1,000+ recharges  
✅ VTPass recharges complete successfully  
✅ Webhooks process without errors  
✅ Database queries match expected results  
✅ No console errors in frontend  
✅ All API endpoints return 200 OK  

---

## Next Steps After Testing

1. **Fix any bugs found**
2. **Deploy to staging environment**
3. **Load testing (simulate 1000 concurrent users)**
4. **Security audit (penetration testing)**
5. **Production deployment**

---

**Prepared By:** Engineering Team  
**Last Updated:** February 20, 2026  
**Version:** 1.0
