package services_test

import (
	"context"
	"errors"
	"testing"

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

func newSpinSvc(sr *mockSpinRepo, ur *mockUserRepoSpin) *services.SpinService {
	return services.NewSpinService(sr, nil, ur, nil, nil, nil, nil, nil)
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

func TestCheckEligibility_UserFound_NoDB_ReturnsIneligible(t *testing.T) {
	// When the service has no DB connection (unit-test mode), CheckEligibility
	// returns ineligible after confirming the user exists. This verifies that
	// the function does not panic and returns a clean response.
	sr := &mockSpinRepo{}
	ur := &mockUserRepoSpin{}

	userID := uuid.New()
	user := &entities.Users{ID: userID, MSISDN: "08012345678"}

	ur.On("FindByMSISDN", mock.Anything, "08012345678").Return(user, nil)

	svc := newSpinSvc(sr, ur) // s.db == nil
	result, err := svc.CheckEligibility(context.Background(), "08012345678")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Eligible)
	// Spin repo must NOT be called when DB is nil
	sr.AssertNotCalled(t, "CountPendingByUserID", mock.Anything, mock.Anything)
	sr.AssertNotCalled(t, "FindByUserID", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	ur.AssertExpectations(t)
}

func TestCheckEligibility_DifferentUser_NoDB_ReturnsIneligible(t *testing.T) {
	// Verifies that ineligibility is correctly returned for a different MSISDN
	// when no DB is available — ensures no cross-user state leakage.
	sr := &mockSpinRepo{}
	ur := &mockUserRepoSpin{}

	userID := uuid.New()
	user := &entities.Users{ID: userID, MSISDN: "08099887766"}

	ur.On("FindByMSISDN", mock.Anything, "08099887766").Return(user, nil)

	svc := newSpinSvc(sr, ur) // s.db == nil
	result, err := svc.CheckEligibility(context.Background(), "08099887766")

	assert.NoError(t, err)
	assert.False(t, result.Eligible)
	sr.AssertNotCalled(t, "CountPendingByUserID", mock.Anything, mock.Anything)
	ur.AssertExpectations(t)
}

func TestCheckEligibility_NilDB_DoesNotPanic(t *testing.T) {
	// Verifies that CheckEligibility does not panic when s.db is nil,
	// and returns a clean ineligible response after confirming the user exists.
	sr := &mockSpinRepo{}
	ur := &mockUserRepoSpin{}

	userID := uuid.New()
	user := &entities.Users{ID: userID, MSISDN: "08011223344"}

	ur.On("FindByMSISDN", mock.Anything, "08011223344").Return(user, nil)

	svc := newSpinSvc(sr, ur) // s.db == nil — should not panic
	result, err := svc.CheckEligibility(context.Background(), "08011223344")

	assert.NoError(t, err)
	assert.False(t, result.Eligible)
	sr.AssertNotCalled(t, "CountPendingByUserID", mock.Anything, mock.Anything)
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
