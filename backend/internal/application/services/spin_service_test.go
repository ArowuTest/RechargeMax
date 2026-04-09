package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"rechargemax/internal/application/services"
	"rechargemax/internal/domain/entities"
)

// ─── SpinRepository mock ──────────────────────────────────────────────────────

type mockSpinRepo struct{ mock.Mock }

func (m *mockSpinRepo) Create(ctx context.Context, e *entities.SpinResults) error {
	return m.Called(ctx, e).Error(0)
}
func (m *mockSpinRepo) FindByID(ctx context.Context, id uuid.UUID) (*entities.SpinResults, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.SpinResults), args.Error(1)
}
func (m *mockSpinRepo) FindAll(ctx context.Context, limit, offset int) ([]*entities.SpinResults, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*entities.SpinResults), args.Error(1)
}
func (m *mockSpinRepo) Update(ctx context.Context, e *entities.SpinResults) error {
	return m.Called(ctx, e).Error(0)
}
func (m *mockSpinRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockSpinRepo) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockSpinRepo) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.SpinResults, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*entities.SpinResults), args.Error(1)
}
func (m *mockSpinRepo) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockSpinRepo) CountPendingByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockSpinRepo) CountTodayByMSISDN(ctx context.Context, msisdn string, since time.Time) (int64, error) {
	args := m.Called(ctx, msisdn, since)
	return args.Get(0).(int64), args.Error(1)
}

// ─── TransactionRepository mock (minimal — only methods used by CheckEligibility) ─

type mockRechargeRepo struct{ mock.Mock }

func (m *mockRechargeRepo) Create(ctx context.Context, e *entities.Transactions) error {
	return m.Called(ctx, e).Error(0)
}
func (m *mockRechargeRepo) FindByID(ctx context.Context, id uuid.UUID) (*entities.Transactions, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Transactions), args.Error(1)
}
func (m *mockRechargeRepo) FindAll(ctx context.Context, limit, offset int) ([]*entities.Transactions, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*entities.Transactions), args.Error(1)
}
func (m *mockRechargeRepo) Update(ctx context.Context, e *entities.Transactions) error {
	return m.Called(ctx, e).Error(0)
}
func (m *mockRechargeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockRechargeRepo) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockRechargeRepo) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Transactions, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*entities.Transactions), args.Error(1)
}
func (m *mockRechargeRepo) FindByReference(ctx context.Context, ref string) (*entities.Transactions, error) {
	args := m.Called(ctx, ref)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Transactions), args.Error(1)
}
func (m *mockRechargeRepo) FindByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Transactions, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]*entities.Transactions), args.Error(1)
}
func (m *mockRechargeRepo) GetTotalRevenue(ctx context.Context) (float64, error) {
	args := m.Called(ctx)
	return args.Get(0).(float64), args.Error(1)
}
func (m *mockRechargeRepo) GetRevenueByDate(ctx context.Context, date time.Time) (float64, error) {
	args := m.Called(ctx, date)
	return args.Get(0).(float64), args.Error(1)
}
func (m *mockRechargeRepo) CountPendingWithdrawals(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Get(0).(int), args.Error(1)
}
func (m *mockRechargeRepo) CountActiveSubscriptions(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Get(0).(int), args.Error(1)
}
func (m *mockRechargeRepo) CountSpinsByDate(ctx context.Context, date time.Time) (int, error) {
	args := m.Called(ctx, date)
	return args.Get(0).(int), args.Error(1)
}
func (m *mockRechargeRepo) CountEligibleForSpin(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockRechargeRepo) FindEligibleForSpin(ctx context.Context, userID uuid.UUID) (*entities.Transactions, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Transactions), args.Error(1)
}
func (m *mockRechargeRepo) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockRechargeRepo) CountByMSISDN(ctx context.Context, msisdn string) (int64, error) {
	args := m.Called(ctx, msisdn)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockRechargeRepo) SumSuccessfulAmountByMSISDNSince(ctx context.Context, msisdn string, since time.Time) (int64, error) {
	args := m.Called(ctx, msisdn, since)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockRechargeRepo) FindByPaymentRef(ctx context.Context, ref string) (*entities.Transactions, error) {
	args := m.Called(ctx, ref)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Transactions), args.Error(1)
}
func (m *mockRechargeRepo) FindByMSISDN(ctx context.Context, msisdn string, limit, offset int) ([]*entities.Transactions, error) {
	args := m.Called(ctx, msisdn, limit, offset)
	return args.Get(0).([]*entities.Transactions), args.Error(1)
}
func (m *mockRechargeRepo) FindByPaymentReference(ctx context.Context, ref string) (*entities.Transactions, error) {
	args := m.Called(ctx, ref)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Transactions), args.Error(1)
}
func (m *mockRechargeRepo) CountByUserIDAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) (int64, error) {
	args := m.Called(ctx, userID, start, end)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockRechargeRepo) SumAmountByUserIDAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) (float64, error) {
	args := m.Called(ctx, userID, start, end)
	return args.Get(0).(float64), args.Error(1)
}

// ─── UserRepository mock (full interface) ────────────────────────────────────

type mockUserRepoSpin struct{ mock.Mock }

func (m *mockUserRepoSpin) Create(ctx context.Context, u *entities.Users) error {
	return m.Called(ctx, u).Error(0)
}
func (m *mockUserRepoSpin) CreateBatch(ctx context.Context, us []*entities.Users) error {
	return m.Called(ctx, us).Error(0)
}
func (m *mockUserRepoSpin) CreateUserWithDefaults(ctx context.Context, msisdn string, ref *uuid.UUID) (*entities.Users, error) {
	args := m.Called(ctx, msisdn, ref)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Users), args.Error(1)
}
func (m *mockUserRepoSpin) FindByID(ctx context.Context, id uuid.UUID) (*entities.Users, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Users), args.Error(1)
}
func (m *mockUserRepoSpin) FindByMSISDN(ctx context.Context, msisdn string) (*entities.Users, error) {
	args := m.Called(ctx, msisdn)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Users), args.Error(1)
}
func (m *mockUserRepoSpin) FindByReferralCode(ctx context.Context, code string) (*entities.Users, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Users), args.Error(1)
}
func (m *mockUserRepoSpin) FindByEmail(ctx context.Context, email string) (*entities.Users, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Users), args.Error(1)
}
func (m *mockUserRepoSpin) FindAll(ctx context.Context, limit, offset int) ([]*entities.Users, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*entities.Users), args.Error(1)
}
func (m *mockUserRepoSpin) FindByLoyaltyTier(ctx context.Context, tier string) ([]*entities.Users, error) {
	args := m.Called(ctx, tier)
	return args.Get(0).([]*entities.Users), args.Error(1)
}
func (m *mockUserRepoSpin) FindActiveUsers(ctx context.Context, limit, offset int) ([]*entities.Users, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*entities.Users), args.Error(1)
}
func (m *mockUserRepoSpin) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockUserRepoSpin) Update(ctx context.Context, u *entities.Users) error {
	return m.Called(ctx, u).Error(0)
}
func (m *mockUserRepoSpin) UpdateStatus(ctx context.Context, id uuid.UUID, active bool) error {
	return m.Called(ctx, id, active).Error(0)
}
func (m *mockUserRepoSpin) UpdatePoints(ctx context.Context, id uuid.UUID, pts int) error {
	return m.Called(ctx, id, pts).Error(0)
}
func (m *mockUserRepoSpin) UpdateLoyaltyTier(ctx context.Context, id uuid.UUID, tier string) error {
	return m.Called(ctx, id, tier).Error(0)
}
func (m *mockUserRepoSpin) UpdateRechargeStats(ctx context.Context, id uuid.UUID, amount float64) error {
	return m.Called(ctx, id, amount).Error(0)
}
func (m *mockUserRepoSpin) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// ─── Helper ───────────────────────────────────────────────────────────────────

// newSpinSvc creates a SpinService wired with mock repos and no DB (unit-test mode).
// The rechargeRepo argument is optional — pass nil to get the old nil-repo behaviour.
func newSpinSvc(sr *mockSpinRepo, ur *mockUserRepoSpin) *services.SpinService {
	return services.NewSpinService(sr, nil, ur, nil, nil, nil, nil, nil)
}

func newSpinSvcWithRecharge(sr *mockSpinRepo, ur *mockUserRepoSpin, rr *mockRechargeRepo) *services.SpinService {
	return services.NewSpinService(sr, nil, ur, rr, nil, nil, nil, nil)
}

// ─── CheckEligibility tests ───────────────────────────────────────────────────

func TestCheckEligibility_UserNotFound_ReturnsIneligible(t *testing.T) {
	sr := &mockSpinRepo{}
	ur := &mockUserRepoSpin{}

	ur.On("FindByMSISDN", mock.Anything, "08012345678").
		Return(nil, errors.New("not found"))

	svc := newSpinSvc(sr, ur)
	result, err := svc.CheckEligibility(context.Background(), "08012345678")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Eligible)
	ur.AssertExpectations(t)
}

func TestCheckEligibility_BelowMinimumRecharge_ReturnsIneligible(t *testing.T) {
	// User exists but today's recharge total is below ₦1,000 (100,000 kobo).
	// CheckEligibility must return ineligible without calling CountTodayByMSISDN.
	sr := &mockSpinRepo{}
	ur := &mockUserRepoSpin{}
	rr := &mockRechargeRepo{}

	user := &entities.Users{ID: uuid.New(), MSISDN: "08012345678"}
	ur.On("FindByMSISDN", mock.Anything, "08012345678").Return(user, nil)
	// Return 50,000 kobo (₦500) — below the ₦1,000 threshold
	rr.On("SumSuccessfulAmountByMSISDNSince", mock.Anything, "08012345678", mock.AnythingOfType("time.Time")).
		Return(int64(50000), nil)

	svc := newSpinSvcWithRecharge(sr, ur, rr)
	result, err := svc.CheckEligibility(context.Background(), "08012345678")

	assert.NoError(t, err)
	assert.False(t, result.Eligible)
	assert.Contains(t, result.Message, "Recharge ₦1000+")
	// Spin count must NOT be queried when the recharge threshold is not met
	sr.AssertNotCalled(t, "CountTodayByMSISDN", mock.Anything, mock.Anything, mock.Anything)
	rr.AssertExpectations(t)
	ur.AssertExpectations(t)
}

func TestCheckEligibility_HasSpinsAvailable_ReturnsEligible(t *testing.T) {
	// User has recharged ₦1,500 today (150,000 kobo) and has played 0 spins.
	// With s.db == nil the tier lookup falls back to dailyCap=1, so 1 spin is available.
	sr := &mockSpinRepo{}
	ur := &mockUserRepoSpin{}
	rr := &mockRechargeRepo{}

	user := &entities.Users{ID: uuid.New(), MSISDN: "08012345678"}
	ur.On("FindByMSISDN", mock.Anything, "08012345678").Return(user, nil)
	// ₦1,500 in kobo — above the ₦1,000 threshold
	rr.On("SumSuccessfulAmountByMSISDNSince", mock.Anything, "08012345678", mock.AnythingOfType("time.Time")).
		Return(int64(150000), nil)
	// 0 spins played today
	sr.On("CountTodayByMSISDN", mock.Anything, "08012345678", mock.AnythingOfType("time.Time")).
		Return(int64(0), nil)

	svc := newSpinSvcWithRecharge(sr, ur, rr)
	result, err := svc.CheckEligibility(context.Background(), "08012345678")

	assert.NoError(t, err)
	assert.True(t, result.Eligible)
	assert.GreaterOrEqual(t, result.AvailableSpins, int64(1))
	// Guard: the old O(N) method must NOT be called
	sr.AssertNotCalled(t, "FindByUserID", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	sr.AssertExpectations(t)
	rr.AssertExpectations(t)
	ur.AssertExpectations(t)
}

func TestCheckEligibility_AllSpinsUsed_ReturnsIneligible(t *testing.T) {
	// User has recharged ₦1,500 today but has already used all spins (1 spin used,
	// dailyCap defaults to 1 when s.db is nil for tier lookup).
	sr := &mockSpinRepo{}
	ur := &mockUserRepoSpin{}
	rr := &mockRechargeRepo{}

	user := &entities.Users{ID: uuid.New(), MSISDN: "08099887766"}
	ur.On("FindByMSISDN", mock.Anything, "08099887766").Return(user, nil)
	rr.On("SumSuccessfulAmountByMSISDNSince", mock.Anything, "08099887766", mock.AnythingOfType("time.Time")).
		Return(int64(150000), nil)
	// 1 spin already played today — equals the default dailyCap of 1
	sr.On("CountTodayByMSISDN", mock.Anything, "08099887766", mock.AnythingOfType("time.Time")).
		Return(int64(1), nil)

	svc := newSpinSvcWithRecharge(sr, ur, rr)
	result, err := svc.CheckEligibility(context.Background(), "08099887766")

	assert.NoError(t, err)
	assert.False(t, result.Eligible)
	assert.Equal(t, int64(1), result.SpinsUsed)
	assert.Contains(t, result.Message, "Daily spin limit reached")
	sr.AssertExpectations(t)
	rr.AssertExpectations(t)
	ur.AssertExpectations(t)
}

func TestCheckEligibility_RechargeRepoError_TreatedAsZero(t *testing.T) {
	// If the recharge repo returns an error, CheckEligibility treats the daily
	// total as 0 and returns ineligible with a clean message — no 500 error.
	sr := &mockSpinRepo{}
	ur := &mockUserRepoSpin{}
	rr := &mockRechargeRepo{}

	user := &entities.Users{ID: uuid.New(), MSISDN: "08011223344"}
	ur.On("FindByMSISDN", mock.Anything, "08011223344").Return(user, nil)
	rr.On("SumSuccessfulAmountByMSISDNSince", mock.Anything, "08011223344", mock.AnythingOfType("time.Time")).
		Return(int64(0), errors.New("db timeout"))

	svc := newSpinSvcWithRecharge(sr, ur, rr)
	result, err := svc.CheckEligibility(context.Background(), "08011223344")

	assert.NoError(t, err) // error is swallowed — clean response to frontend
	assert.False(t, result.Eligible)
	// Spin count must NOT be queried when the recharge sum failed
	sr.AssertNotCalled(t, "CountTodayByMSISDN", mock.Anything, mock.Anything, mock.Anything)
	rr.AssertExpectations(t)
	ur.AssertExpectations(t)
}

func TestCheckEligibility_SpinCountError_TreatedAsZero(t *testing.T) {
	// If CountTodayByMSISDN returns an error, CheckEligibility conservatively
	// treats it as 0 spins used (user is not incorrectly blocked).
	sr := &mockSpinRepo{}
	ur := &mockUserRepoSpin{}
	rr := &mockRechargeRepo{}

	user := &entities.Users{ID: uuid.New(), MSISDN: "08033445566"}
	ur.On("FindByMSISDN", mock.Anything, "08033445566").Return(user, nil)
	rr.On("SumSuccessfulAmountByMSISDNSince", mock.Anything, "08033445566", mock.AnythingOfType("time.Time")).
		Return(int64(150000), nil) // ₦1,500 — above threshold
	sr.On("CountTodayByMSISDN", mock.Anything, "08033445566", mock.AnythingOfType("time.Time")).
		Return(int64(0), errors.New("connection reset"))

	svc := newSpinSvcWithRecharge(sr, ur, rr)
	result, err := svc.CheckEligibility(context.Background(), "08033445566")

	assert.NoError(t, err)
	// With 0 spins used (error treated as 0) and dailyCap=1, user is eligible
	assert.True(t, result.Eligible)
	sr.AssertExpectations(t)
	rr.AssertExpectations(t)
	ur.AssertExpectations(t)
}

func TestCheckEligibility_EmptyMSISDN_HandledGracefully(t *testing.T) {
	sr := &mockSpinRepo{}
	ur := &mockUserRepoSpin{}

	ur.On("FindByMSISDN", mock.Anything, "").
		Return(nil, errors.New("not found"))

	svc := newSpinSvc(sr, ur)
	result, err := svc.CheckEligibility(context.Background(), "")

	assert.NoError(t, err)
	assert.False(t, result.Eligible)
}
