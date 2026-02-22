# 🚀 Affiliate System Enterprise Deployment Guide

## Overview

This guide covers the deployment of the enterprise-grade affiliate system with all 15 critical issues fixed and production-ready features implemented.

---

## 📋 Pre-Deployment Checklist

### ✅ Database Requirements
- [x] PostgreSQL 14+ installed
- [x] UUID extension enabled (`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
- [x] Database backup completed
- [x] Migration rollback plan prepared

### ✅ API Keys & Credentials
- [ ] Paystack Secret Key configured
- [ ] Paystack webhook URL registered
- [ ] Database connection string secured
- [ ] Environment variables set

### ✅ Infrastructure
- [ ] Production database provisioned
- [ ] Backup strategy configured
- [ ] Monitoring tools setup
- [ ] Logging infrastructure ready

---

## 🗄️ Database Migration

### Step 1: Apply Migrations in Order

```bash
# Navigate to migrations directory
cd /home/ubuntu/rechargemax-production-OriginalBuild/database/migrations

# Apply migrations in sequence
psql -U postgres -d rechargemax -f 027_fix_system_config.sql
psql -U postgres -d rechargemax -f 028_affiliate_enterprise_fixes_p0.sql
psql -U postgres -d rechargemax -f 029_affiliate_enterprise_fixes_p1.sql
psql -U postgres -d rechargemax -f 030_affiliate_fixes_final.sql
```

### Step 2: Verify Migration Success

```sql
-- Check all functions exist
SELECT proname FROM pg_proc WHERE proname LIKE '%affiliate%' OR proname LIKE '%commission%';

-- Verify triggers
SELECT tgname FROM pg_trigger WHERE tgname LIKE '%affiliate%';

-- Check system config
SELECT key, value FROM system_config WHERE key LIKE 'affiliate_%';

-- Verify tier data
SELECT * FROM affiliate_commission_tiers ORDER BY min_referrals;
```

### Step 3: Run Test Suite

```bash
# Run database tests
psql -U postgres -d rechargemax -f tests/affiliate_system_tests.sql
```

---

## 🔧 Backend Configuration

### Step 1: Update Environment Variables

```bash
# .env file
DATABASE_URL=postgresql://user:password@localhost:5432/rechargemax
PAYSTACK_SECRET_KEY=sk_live_xxxxxxxxxxxxx
PAYSTACK_WEBHOOK_SECRET=whsec_xxxxxxxxxxxxx
AFFILIATE_PAYOUT_ENABLED=true
AFFILIATE_AUTO_APPROVAL=false
```

### Step 2: Initialize Services

```go
// main.go - Add payout service initialization

// Initialize payout service
payoutService := services.NewPayoutService(
    db,
    affiliateRepo,
    payoutRepo,
    commissionRepo,
    bankAccountRepo,
    os.Getenv("PAYSTACK_SECRET_KEY"),
    notificationService,
)

// Initialize enterprise affiliate service
affiliateServiceV2 := services.NewAffiliateServiceV2(
    db,
    affiliateRepo,
    commissionRepo,
    userRepo,
    transactionRepo,
    payoutService,
    notificationService,
)
```

### Step 3: Register Webhook Endpoint

```go
// routes.go - Add Paystack webhook handler

router.POST("/webhooks/paystack", func(c *gin.Context) {
    var event map[string]interface{}
    if err := c.BindJSON(&event); err != nil {
        c.JSON(400, gin.H{"error": "Invalid payload"})
        return
    }
    
    // Verify webhook signature
    signature := c.GetHeader("X-Paystack-Signature")
    if !verifyPaystackSignature(c.Request.Body, signature) {
        c.JSON(401, gin.H{"error": "Invalid signature"})
        return
    }
    
    // Handle webhook
    if err := payoutService.HandlePaystackWebhook(c.Request.Context(), event); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"status": "success"})
})
```

---

## 📊 Scheduled Jobs Setup

### Daily Jobs (Cron)

```bash
# Add to crontab
0 0 * * * psql -U postgres -d rechargemax -c "SELECT update_active_referrals();"
0 1 * * * psql -U postgres -d rechargemax -c "SELECT aggregate_affiliate_analytics(CURRENT_DATE - INTERVAL '1 day');"
```

### Alternative: Go Scheduler

```go
// scheduler.go

import "github.com/robfig/cron/v3"

func StartScheduler(db *sql.DB) {
    c := cron.New()
    
    // Update active referrals daily at midnight
    c.AddFunc("0 0 * * *", func() {
        _, err := db.Exec("SELECT update_active_referrals()")
        if err != nil {
            log.Printf("Failed to update active referrals: %v", err)
        }
    })
    
    // Aggregate analytics daily at 1 AM
    c.AddFunc("0 1 * * *", func() {
        _, err := db.Exec("SELECT aggregate_affiliate_analytics($1)", time.Now().AddDate(0, 0, -1))
        if err != nil {
            log.Printf("Failed to aggregate analytics: %v", err)
        }
    })
    
    c.Start()
}
```

---

## 🧪 Testing

### Test 1: Commission Processing

```bash
# Create test transaction
curl -X POST http://localhost:8080/api/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "test-user-id",
    "amount": 100000,
    "type": "RECHARGE"
  }'

# Verify commission created
psql -U postgres -d rechargemax -c "
SELECT * FROM affiliate_commissions 
WHERE transaction_id = 'transaction-id' 
ORDER BY created_at DESC LIMIT 1;
"
```

### Test 2: Payout Initiation

```bash
# Request payout
curl -X POST http://localhost:8080/api/affiliates/payouts \
  -H "Authorization: Bearer token" \
  -H "Content-Type: application/json" \
  -d '{
    "affiliate_id": "affiliate-id",
    "amount": 500000
  }'

# Check payout status
psql -U postgres -d rechargemax -c "
SELECT * FROM affiliate_payouts 
WHERE affiliate_id = 'affiliate-id' 
ORDER BY created_at DESC LIMIT 1;
"
```

### Test 3: Tier Upgrade

```bash
# Simulate referrals
for i in {1..30}; do
  curl -X POST http://localhost:8080/api/users/register \
    -H "Content-Type: application/json" \
    -d "{
      \"msisdn\": \"23480${i}0000000\",
      \"referred_by\": \"affiliate-user-id\"
    }"
done

# Verify tier upgrade
psql -U postgres -d rechargemax -c "
SELECT tier, commission_rate, total_referrals 
FROM affiliates 
WHERE user_id = 'affiliate-user-id';
"
```

### Test 4: Fraud Prevention

```bash
# Simulate click spam (should be blocked after 10 clicks)
for i in {1..15}; do
  curl -X POST http://localhost:8080/api/affiliates/clicks \
    -H "Content-Type: application/json" \
    -d '{
      "affiliate_code": "TEST123",
      "ip_address": "192.168.1.1"
    }'
done

# Should return error after 10th click
```

---

## 📈 Monitoring

### Key Metrics to Track

1. **Commission Processing**
   - Total commissions processed per day
   - Average commission amount
   - Commission success rate

2. **Payout Operations**
   - Total payouts initiated
   - Payout success rate
   - Average payout processing time

3. **Fraud Detection**
   - Blocked IPs count
   - Suspicious activity alerts
   - Referral loop attempts

4. **Performance**
   - Database query performance
   - API response times
   - Webhook processing latency

### Monitoring Queries

```sql
-- Daily commission summary
SELECT 
    DATE(created_at) as date,
    COUNT(*) as total_commissions,
    SUM(commission_amount) as total_amount,
    AVG(commission_amount) as avg_amount
FROM affiliate_commissions
WHERE created_at > NOW() - INTERVAL '30 days'
GROUP BY DATE(created_at)
ORDER BY date DESC;

-- Payout status summary
SELECT 
    payout_status,
    COUNT(*) as count,
    SUM(total_amount) as total_amount
FROM affiliate_payouts
WHERE created_at > NOW() - INTERVAL '30 days'
GROUP BY payout_status;

-- Fraud detection summary
SELECT 
    COUNT(*) as blocked_ips,
    SUM(click_count) as total_blocked_clicks
FROM affiliate_click_rate_limits
WHERE is_blocked = true;

-- Top affiliates
SELECT * FROM get_top_affiliates(10);
```

---

## 🔒 Security Checklist

- [ ] Paystack webhook signature verification enabled
- [ ] Database credentials secured (not in code)
- [ ] Rate limiting configured on API endpoints
- [ ] SQL injection prevention verified
- [ ] HTTPS enforced on all endpoints
- [ ] Sensitive data encrypted at rest
- [ ] Audit logging enabled
- [ ] Access control implemented

---

## 🚨 Rollback Plan

### If Migration Fails

```sql
-- Rollback migrations in reverse order
BEGIN;

-- Drop new functions
DROP FUNCTION IF EXISTS aggregate_affiliate_analytics(DATE);
DROP FUNCTION IF EXISTS update_active_referrals();
DROP FUNCTION IF EXISTS check_commission_limits(UUID, INTEGER);
DROP FUNCTION IF EXISTS reverse_affiliate_commission(UUID, TEXT);
DROP FUNCTION IF EXISTS validate_affiliate_click();
DROP FUNCTION IF EXISTS validate_affiliate_referral();

-- Drop new tables
DROP TABLE IF EXISTS affiliate_click_rate_limits;
DROP TABLE IF EXISTS affiliate_commission_tiers;

-- Restore original schema
-- (Run backup restore script)

ROLLBACK; -- or COMMIT if restore successful
```

### If Backend Deployment Fails

```bash
# Revert to previous version
git revert HEAD
git push origin main

# Redeploy previous version
./deploy.sh --version previous
```

---

## 📞 Support & Troubleshooting

### Common Issues

**Issue 1: Commission not created**
- Check if affiliate is APPROVED
- Verify transaction status is SUCCESS
- Confirm it's not the first recharge
- Check commission limits

**Issue 2: Payout fails**
- Verify bank account is verified
- Check Paystack API key is valid
- Ensure sufficient balance
- Review Paystack logs

**Issue 3: Tier not upgrading**
- Check total_referrals count
- Verify trigger is enabled
- Review audit log for errors

**Issue 4: Webhook not processing**
- Verify webhook URL is correct
- Check signature verification
- Review server logs
- Test with Paystack webhook tester

### Debug Commands

```bash
# Check database connections
psql -U postgres -d rechargemax -c "SELECT COUNT(*) FROM pg_stat_activity;"

# View recent errors
tail -f /var/log/rechargemax/error.log

# Check service status
systemctl status rechargemax-backend

# Test Paystack connectivity
curl -H "Authorization: Bearer $PAYSTACK_SECRET_KEY" \
  https://api.paystack.co/bank
```

---

## 📚 Additional Resources

- [Paystack Transfer API Documentation](https://paystack.com/docs/transfers/single-transfers)
- [PostgreSQL Trigger Documentation](https://www.postgresql.org/docs/current/sql-createtrigger.html)
- [Go Cron Library](https://github.com/robfig/cron)

---

## ✅ Post-Deployment Verification

### Checklist

- [ ] All migrations applied successfully
- [ ] Database tests passing
- [ ] Backend services running
- [ ] Webhook endpoint accessible
- [ ] Scheduled jobs configured
- [ ] Monitoring dashboards setup
- [ ] Alerts configured
- [ ] Documentation updated
- [ ] Team trained on new features
- [ ] Rollback plan tested

### Success Criteria

- ✅ Commission processing: 100% success rate
- ✅ Payout automation: <5 minute processing time
- ✅ Fraud detection: 0 false positives
- ✅ Tier upgrades: Automatic within 1 second
- ✅ Analytics: Updated daily
- ✅ API response time: <200ms p95
- ✅ Database queries: <100ms p95

---

## 🎉 Deployment Complete!

Your enterprise-grade affiliate system is now live with:

- ✅ **15 critical issues fixed**
- ✅ **Automated payout system**
- ✅ **Fraud prevention**
- ✅ **Tier automation**
- ✅ **Analytics aggregation**
- ✅ **Commission reversal**
- ✅ **Rate limiting**
- ✅ **Audit logging**

**Next Steps:**
1. Monitor system for 24 hours
2. Review analytics dashboard
3. Test with real affiliates
4. Gather feedback
5. Iterate and improve

---

**Questions or Issues?**  
Contact: support@rechargemax.com  
Documentation: https://docs.rechargemax.com/affiliate
