# RechargeMax Testing Guide

**Date:** February 1, 2026  
**Version:** 1.0.0  
**Status:** Production-Ready

---

## Table of Contents

1. [Testing Strategy](#testing-strategy)
2. [Manual Testing Procedures](#manual-testing-procedures)
3. [API Testing](#api-testing)
4. [Frontend Testing](#frontend-testing)
5. [Security Testing](#security-testing)
6. [Performance Testing](#performance-testing)
7. [Integration Testing](#integration-testing)
8. [Automated Testing Setup](#automated-testing-setup)
9. [Test Data](#test-data)
10. [Bug Reporting](#bug-reporting)

---

## Testing Strategy

### Testing Pyramid

```
    /\
   /  \  E2E Tests (10%)
  /----\
 / Unit \ Integration Tests (30%)
/--------\
   Unit    Unit Tests (60%)
```

### Test Levels

1. **Unit Tests** - Individual functions and methods
2. **Integration Tests** - Component interactions
3. **API Tests** - Endpoint functionality
4. **E2E Tests** - Complete user workflows
5. **Security Tests** - Vulnerability scanning
6. **Performance Tests** - Load and stress testing

---

## Manual Testing Procedures

### Pre-Deployment Checklist

#### 1. Authentication Flow ✅

**Test Steps:**
1. Navigate to login page
2. Enter phone number: `08012345678`
3. Click "Send OTP"
4. Check SMS for OTP code
5. Enter OTP code
6. Verify successful login
7. Check JWT token in localStorage

**Expected Results:**
- OTP sent within 30 seconds
- OTP valid for 5 minutes
- Successful login redirects to dashboard
- Token stored securely

**Test Cases:**
- ✅ Valid phone number
- ✅ Invalid phone number format
- ✅ Wrong OTP code (should lock after 5 attempts)
- ✅ Expired OTP
- ✅ Rate limiting (5 OTPs per 5 minutes)

---

#### 2. Recharge Flow ✅

**Test Steps:**
1. Login as user
2. Navigate to recharge page
3. Enter phone number: `08012345678`
4. Select network: MTN
5. Select type: Airtime
6. Enter amount: ₦1000
7. Click "Recharge Now"
8. Complete payment on Paystack
9. Verify recharge success

**Expected Results:**
- Amount converted to kobo (100000)
- Payment URL generated
- Redirect to Paystack
- Webhook processes payment
- User receives airtime
- Transaction appears in history

**Test Cases:**
- ✅ Airtime recharge (₦100 - ₦50,000)
- ✅ Data recharge with bundle selection
- ✅ Invalid amount (< ₦50)
- ✅ Decimal amount (should be rejected)
- ✅ Payment failure handling
- ✅ Duplicate webhook prevention

---

#### 3. Wheel Spin Flow ✅

**Test Steps:**
1. Login as user
2. Complete recharge of ₦1000+
3. Navigate to spin page
4. Check eligibility
5. Click "Spin Now"
6. Verify prize won
7. Check prize provisioning

**Expected Results:**
- Eligibility check passes
- Spin animation plays
- Prize selected randomly
- Prize provisioned automatically (airtime/data)
- Spin recorded in history
- User cannot spin again same day

**Tier Testing:**
| Tier | Amount | Spins | Test Phone |
|------|--------|-------|------------|
| Bronze | ₦1,000 | 1 | 08011111111 |
| Silver | ₦5,000 | 2 | 08022222222 |
| Gold | ₦10,000 | 3 | 08033333333 |
| Platinum | ₦20,000 | 5 | 08044444444 |
| Diamond | ₦50,000 | 10 | 08055555555 |

**Test Cases:**
- ✅ Eligible user spins
- ✅ Ineligible user (no recharge)
- ✅ Already spun today
- ✅ Race condition (concurrent spins)
- ✅ Prize provisioning success
- ✅ Prize provisioning failure (retry)

---

#### 4. Draw Entry Flow ✅

**Test Steps:**
1. Login as user
2. Complete recharge
3. Navigate to draws page
4. View active draws
5. Check draw entries
6. Verify entry count

**Expected Results:**
- Active draws displayed
- Entry count matches recharge amount
- Draw date shown correctly
- My entries visible

**Test Cases:**
- ✅ View active draws
- ✅ Entry calculation (₦100 = 1 entry)
- ✅ Multiple recharges = multiple entries
- ✅ Draw winner selection (admin)

---

#### 5. Affiliate Program ✅

**Test Steps:**
1. Login as user
2. Navigate to affiliate page
3. Copy referral code
4. Share referral link
5. New user signs up with code
6. Complete recharge as referral
7. Check commission earned

**Expected Results:**
- Unique referral code generated
- Referral link works
- New user linked to referrer
- Commission calculated correctly
- Commission visible in stats

**Test Cases:**
- ✅ Generate referral code
- ✅ Referral signup
- ✅ Commission calculation (5% default)
- ✅ Commission payout request

---

#### 6. Admin Dashboard ✅

**Test Steps:**
1. Login as admin
2. View dashboard stats
3. Manage users
4. Create draw
5. Manage wheel prizes
6. View transactions

**Expected Results:**
- Dashboard shows real-time stats
- User management functional
- Draw creation works
- Prize management works
- Transaction history complete

**Test Cases:**
- ✅ Admin authentication (role verification)
- ✅ Dashboard statistics
- ✅ User search and filtering
- ✅ Draw creation and management
- ✅ Prize CRUD operations
- ✅ Transaction monitoring

---

## API Testing

### Using cURL

#### 1. Send OTP
```bash
curl -X POST https://api.rechargemax.ng/api/v1/auth/send-otp \
  -H "Content-Type: application/json" \
  -d '{"msisdn": "08012345678"}'
```

**Expected Response:**
```json
{
  "success": true,
  "message": "OTP sent successfully",
  "expires_in": 300
}
```

---

#### 2. Verify OTP
```bash
curl -X POST https://api.rechargemax.ng/api/v1/auth/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"msisdn": "08012345678", "otp": "123456"}'
```

**Expected Response:**
```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {...}
}
```

---

#### 3. Check Spin Eligibility
```bash
curl -X GET https://api.rechargemax.ng/api/v1/spin/eligibility \
  -H "Authorization: Bearer {token}"
```

**Expected Response:**
```json
{
  "eligible": true,
  "spins_available": 2,
  "tier": "Silver"
}
```

---

#### 4. Initiate Recharge
```bash
curl -X POST https://api.rechargemax.ng/api/v1/recharge/initiate \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "phoneNumber": "08012345678",
    "networkProvider": "MTN",
    "rechargeType": "AIRTIME",
    "amount": 100000,
    "customerEmail": "test@example.com"
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "transaction_id": "uuid",
  "payment_reference": "REF-123456789",
  "paymentUrl": "https://checkout.paystack.com/..."
}
```

---

### Using Postman

**Collection:** `RechargeMax-API-Tests.postman_collection.json`

**Environment Variables:**
- `base_url`: https://api.rechargemax.ng/api/v1
- `token`: (set after login)
- `test_phone`: 08012345678

**Test Suites:**
1. Authentication Tests (5 tests)
2. Recharge Tests (8 tests)
3. Spin Tests (6 tests)
4. Draw Tests (4 tests)
5. Affiliate Tests (5 tests)
6. Admin Tests (10 tests)

**Total:** 38 automated API tests

---

## Frontend Testing

### Browser Compatibility

Test on:
- ✅ Chrome (latest)
- ✅ Firefox (latest)
- ✅ Safari (latest)
- ✅ Edge (latest)
- ✅ Mobile Safari (iOS)
- ✅ Chrome Mobile (Android)

### Responsive Testing

Test breakpoints:
- ✅ Mobile: 320px - 767px
- ✅ Tablet: 768px - 1023px
- ✅ Desktop: 1024px+

### Accessibility Testing

- ✅ Keyboard navigation
- ✅ Screen reader compatibility
- ✅ Color contrast (WCAG AA)
- ✅ Focus indicators
- ✅ Alt text for images

---

## Security Testing

### 1. Authentication Security

**Tests:**
- ✅ JWT token expiration (24 hours)
- ✅ Token refresh mechanism
- ✅ OTP brute force protection (5 attempts)
- ✅ Rate limiting on OTP endpoint
- ✅ Secure token storage (httpOnly cookies)

---

### 2. Authorization Security

**Tests:**
- ✅ User cannot access admin endpoints
- ✅ Role-based access control (RBAC)
- ✅ JWT role verification
- ✅ Admin token validation

---

### 3. Input Validation

**Tests:**
- ✅ SQL injection prevention (GORM)
- ✅ XSS prevention (HTML escaping)
- ✅ CSRF protection
- ✅ Phone number validation
- ✅ Amount validation (kobo)
- ✅ Email validation

---

### 4. API Security

**Tests:**
- ✅ HTTPS enforcement
- ✅ CORS configuration
- ✅ Rate limiting (100 req/min)
- ✅ Request size limiting (10MB)
- ✅ Webhook signature verification

---

## Performance Testing

### Load Testing

**Tool:** Apache JMeter or k6

**Scenarios:**

#### 1. Normal Load
- **Users:** 100 concurrent
- **Duration:** 10 minutes
- **Expected:** < 500ms response time

#### 2. Peak Load
- **Users:** 500 concurrent
- **Duration:** 5 minutes
- **Expected:** < 1000ms response time

#### 3. Stress Test
- **Users:** 1000+ concurrent
- **Duration:** 2 minutes
- **Expected:** Graceful degradation

---

### Database Performance

**Tests:**
- ✅ Query performance (< 100ms)
- ✅ Index effectiveness
- ✅ Connection pooling
- ✅ Transaction throughput

**Critical Queries:**
1. Spin eligibility check
2. Transaction history
3. User dashboard
4. Draw entries

---

### Caching Strategy

**Tests:**
- ✅ Network list caching
- ✅ Data plan caching
- ✅ Prize list caching
- ✅ Cache invalidation

---

## Integration Testing

### External Services

#### 1. Paystack Integration
```bash
# Test payment initialization
# Test webhook processing
# Test payment verification
```

**Test Cases:**
- ✅ Successful payment
- ✅ Failed payment
- ✅ Duplicate webhook
- ✅ Invalid signature

---

#### 2. VTPass Integration
```bash
# Test airtime purchase
# Test data purchase
# Test network validation
```

**Test Cases:**
- ✅ Successful recharge
- ✅ Failed recharge
- ✅ Network unavailable
- ✅ Invalid phone number

---

#### 3. Termii Integration
```bash
# Test OTP sending
# Test SMS delivery
```

**Test Cases:**
- ✅ OTP sent successfully
- ✅ SMS delivery failure
- ✅ Rate limiting

---

## Automated Testing Setup

### Backend Unit Tests (Go)

**Setup:**
```bash
cd backend
go test ./...
```

**Coverage Target:** 70%+

**Example Test:**
```go
func TestSpinEligibility(t *testing.T) {
    // Test spin eligibility logic
    user := &entities.User{
        TotalRechargeAmount: 100000, // ₦1000
    }
    
    eligible := spinService.CheckEligibility(user)
    assert.True(t, eligible)
}
```

---

### Frontend Unit Tests (Jest + React Testing Library)

**Setup:**
```bash
cd frontend
npm test
```

**Coverage Target:** 60%+

**Example Test:**
```javascript
test('validates phone number format', () => {
  render(<PremiumRechargeForm />);
  const input = screen.getByLabelText('Phone Number');
  
  fireEvent.change(input, { target: { value: '0801234567' } });
  expect(screen.getByText('Invalid phone number')).toBeInTheDocument();
});
```

---

### E2E Tests (Cypress or Playwright)

**Setup:**
```bash
cd e2e-tests
npm install
npx cypress open
```

**Test Scenarios:**
1. Complete recharge flow
2. Spin wheel and win prize
3. Affiliate signup and commission
4. Admin create draw

---

## Test Data

### Test Users

| Role | Phone | Email | Password |
|------|-------|-------|----------|
| User | 08011111111 | user@test.com | (OTP) |
| Admin | 08099999999 | admin@test.com | (OTP) |
| Affiliate | 08022222222 | affiliate@test.com | (OTP) |

### Test Networks

- MTN
- Airtel
- Glo
- 9mobile

### Test Amounts

- Minimum: ₦50 (5000 kobo)
- Normal: ₦100, ₦500, ₦1000
- Maximum: ₦50,000 (5000000 kobo)

---

## Bug Reporting

### Bug Report Template

```markdown
**Title:** [Component] Brief description

**Severity:** Critical / High / Medium / Low

**Environment:**
- Platform: Web / Mobile
- Browser: Chrome 120
- OS: Windows 11
- API Version: v1

**Steps to Reproduce:**
1. Login as user
2. Navigate to recharge page
3. Enter amount ₦100
4. Click submit

**Expected Result:**
Payment URL generated

**Actual Result:**
Error: "Invalid amount"

**Screenshots:**
[Attach screenshots]

**Logs:**
[Attach relevant logs]

**Additional Context:**
[Any other relevant information]
```

---

### Bug Severity Levels

- **Critical** - System down, data loss, security breach
- **High** - Major feature broken, affects many users
- **Medium** - Feature partially broken, workaround exists
- **Low** - Minor issue, cosmetic problem

---

## Regression Testing

### Before Each Release

Run full regression test suite:

1. ✅ Authentication flow
2. ✅ Recharge flow (all networks)
3. ✅ Spin flow (all tiers)
4. ✅ Draw entry flow
5. ✅ Affiliate flow
6. ✅ Admin operations
7. ✅ Payment webhook
8. ✅ API endpoints (all)

**Duration:** ~2 hours

---

## Monitoring and Alerts

### Production Monitoring

**Tools:**
- Application logs (Zap logger)
- Error tracking (to be added)
- Performance monitoring (to be added)
- Uptime monitoring (to be added)

**Alerts:**
- API error rate > 5%
- Response time > 2s
- Database connection failures
- Payment webhook failures

---

## Test Automation Roadmap

### Phase 1 (Month 1)
- ✅ Manual testing procedures documented
- ✅ API test collection created
- ⏳ Unit test setup (backend)

### Phase 2 (Month 2)
- ⏳ Unit test coverage > 50%
- ⏳ Integration test setup
- ⏳ E2E test setup

### Phase 3 (Month 3)
- ⏳ CI/CD integration
- ⏳ Automated regression tests
- ⏳ Performance test automation

---

## Contact

For testing support:
- **Email:** qa@rechargemax.ng
- **Slack:** #testing
- **Documentation:** https://docs.rechargemax.ng/testing

---

**Last Updated:** February 1, 2026  
**Next Review:** March 1, 2026
