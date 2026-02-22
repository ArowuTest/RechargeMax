# Architecture Decision Record: Points Calculation

**Date:** February 20, 2026  
**Status:** ✅ Implemented  
**Decision Makers:** Product & Engineering Team

---

## Context

The RechargeMax platform calculates loyalty points for every successful recharge transaction. The points calculation logic existed in TWO places:

1. **Application Layer (Go):** `RechargeService.ProcessSuccessfulPayment()`
2. **Database Layer (Trigger):** `trigger_process_transaction` → `process_successful_transaction()`

This caused several production issues:

### Problems Identified

1. **Duplicate Logic:** Same calculation in two places (hard to maintain)
2. **Trigger Bug:** Database trigger's UPDATE didn't persist (PostgreSQL RETURNING limitation)
3. **Testing Difficulty:** Can't unit test PL/pgSQL triggers easily
4. **Debugging Complexity:** Stack traces end at database boundary
5. **Flexibility Constraints:** Hard to add promotions, A/B tests, bonuses in SQL
6. **Admin Conflicts:** Trigger would recalculate points even for manual admin adjustments

---

## Decision

**Move ALL business logic to the application layer (Go code).**

Remove the `trigger_process_transaction` trigger and `process_successful_transaction()` function entirely.

---

## Rationale

### Why Application Layer?

| Criterion | Application Layer (Go) | Database Trigger (PL/pgSQL) |
|-----------|----------------------|---------------------------|
| **Testability** | ✅ Unit tests, mocks, integration tests | ❌ Hard to test, requires DB |
| **Debuggability** | ✅ Full stack traces, logging | ❌ Opaque, limited visibility |
| **Maintainability** | ✅ Developers know Go | ❌ Fewer devs know PL/pgSQL |
| **Flexibility** | ✅ Easy to add promotions, bonuses | ❌ Complex SQL for business rules |
| **Performance** | ✅ One DB round-trip | ❌ Trigger adds overhead |
| **Admin Tools** | ✅ Can skip calculation for manual adjustments | ❌ Trigger always recalculates |

### What Database Should Do

Databases are excellent at:
- Data integrity (foreign keys, constraints)
- Audit trails (created_at, updated_at)
- Atomic transactions (ACID)

Databases are NOT ideal for:
- Complex business logic
- Conditional calculations based on promotions
- Integration with external services

---

## Implementation

### What We Changed

**Removed:**
- `trigger_process_transaction` trigger
- `process_successful_transaction()` function

**Kept:**
- `update_transactions_updated_at` trigger (timestamp management)
- `calculate_points_earned()` function (for admin tools/reporting)
- `calculate_draw_entries()` function (for admin tools/reporting)

**Enhanced:**
- `RechargeService.ProcessSuccessfulPayment()` now calculates:
  - `points_earned = amount / 20000` (₦200 = 1 point)
  - `draw_entries = points_earned` (1:1 ratio)
  - `spin_eligible = amount >= 100000` (₦1,000 minimum)

### Code Location

**File:** `backend/internal/application/services/recharge_service.go`  
**Method:** `ProcessSuccessfulPayment(ctx context.Context, paymentRef string)`  
**Lines:** 207-228

```go
// Calculate points earned (₦200 = 1 point)
pointsEarned := recharge.Amount / 20000

// Calculate draw entries (1 point = 1 draw entry)
drawEntries := pointsEarned

// Check if eligible for wheel spin (₦1000 minimum)
isWheelEligible := recharge.Amount >= 100000

// Update transaction with all calculated values
tx.Model(&entities.Transactions{}).Where("id = ?", recharge.ID).Updates(map[string]interface{}{
    "status":        "SUCCESS",
    "points_earned": pointsEarned,
    "draw_entries":  drawEntries,
    "spin_eligible": isWheelEligible,
    "completed_at":  time.Now(),
})
```

---

## Consequences

### Positive

✅ **Single Source of Truth:** All business logic in one place  
✅ **Better Testing:** Can unit test points calculation  
✅ **Easier Debugging:** Full visibility into calculation  
✅ **More Flexible:** Easy to add promotions, bonuses  
✅ **Admin-Friendly:** Manual adjustments won't be overwritten  
✅ **Performance:** One less database trigger to execute  

### Negative

❌ **No Database Enforcement:** If someone bypasses Go code and updates DB directly, points won't calculate  
   - **Mitigation:** Use database permissions to prevent direct updates. All updates must go through API.

---

## Verification

### Test Scenarios

| Scenario | Expected Result | Status |
|----------|----------------|--------|
| ₦500 recharge | 2 points, 2 entries, no spin | ✅ |
| ₦1,000 recharge | 5 points, 5 entries, spin eligible | ✅ |
| ₦2,000 recharge | 10 points, 10 entries, spin eligible | ✅ |
| Admin manual adjustment | Points stay as set by admin | ✅ |
| Webhook payment | Points calculated correctly | ✅ |

---

## Migration

**File:** `backend/migrations/036_remove_redundant_points_trigger.sql`

Applied on: February 20, 2026

---

## Related Documents

- `/home/ubuntu/POINTS_CALCULATION_ANALYSIS.md` - Detailed analysis
- `/home/ubuntu/STRATEGIC_AUDIT_AND_FIXES.md` - Complete audit trail
- `/home/ubuntu/CODE_REVIEW_PRODUCTION_READY.md` - Production readiness review

---

## Future Considerations

1. **Promotions System:** Easy to add 2x points days, referral bonuses
2. **A/B Testing:** Can test different points formulas per user segment
3. **Loyalty Tiers:** Can adjust points based on user tier (Bronze, Silver, etc.)
4. **Daily Subscriptions:** Use same pattern (₦20 = 1 point)

---

**Decision Approved By:** Engineering Team  
**Implementation Status:** ✅ Complete  
**Production Ready:** ✅ Yes
