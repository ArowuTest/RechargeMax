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

// ─── OTPRepository mock ───────────────────────────────────────────────────────

type mockOTPRepo struct{ mock.Mock }

func (m *mockOTPRepo) Create(ctx context.Context, otp *entities.OTP) error {
	return m.Called(ctx, otp).Error(0)
}
func (m *mockOTPRepo) FindByID(ctx context.Context, id uuid.UUID) (*entities.OTP, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.OTP), args.Error(1)
}
func (m *mockOTPRepo) FindValidOTP(ctx context.Context, msisdn, code string) (*entities.OTP, error) {
	args := m.Called(ctx, msisdn, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.OTP), args.Error(1)
}
func (m *mockOTPRepo) FindValidOTPWithPurpose(ctx context.Context, msisdn, code, purpose string) (*entities.OTP, error) {
	args := m.Called(ctx, msisdn, code, purpose)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.OTP), args.Error(1)
}
func (m *mockOTPRepo) FindRecentOTPs(ctx context.Context, msisdn string, since time.Time) ([]*entities.OTP, error) {
	args := m.Called(ctx, msisdn, since)
	return args.Get(0).([]*entities.OTP), args.Error(1)
}
func (m *mockOTPRepo) CountRecentOTPs(ctx context.Context, msisdn string, since time.Time) (int64, error) {
	args := m.Called(ctx, msisdn, since)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockOTPRepo) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockOTPRepo) Update(ctx context.Context, otp *entities.OTP) error {
	return m.Called(ctx, otp).Error(0)
}
func (m *mockOTPRepo) DeleteExpired(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}
func (m *mockOTPRepo) DeleteOld(ctx context.Context, older time.Time) error {
	return m.Called(ctx, older).Error(0)
}
func (m *mockOTPRepo) FindLatestPendingOTP(ctx context.Context, msisdn, purpose string) (*entities.OTP, error) {
	args := m.Called(ctx, msisdn, purpose)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.OTP), args.Error(1)
}
func (m *mockOTPRepo) IncrementFailedAttempts(ctx context.Context, id uuid.UUID) (int, error) {
	args := m.Called(ctx, id)
	return args.Int(0), args.Error(1)
}
func (m *mockOTPRepo) InvalidateByMSISDN(ctx context.Context, msisdn, purpose string) error {
	return m.Called(ctx, msisdn, purpose).Error(0)
}

// ─── UserRepository mock (shared with spin tests, reused via same package) ───

type mockUserRepoAuth struct{ mock.Mock }

func (m *mockUserRepoAuth) Create(ctx context.Context, u *entities.Users) error {
	return m.Called(ctx, u).Error(0)
}
func (m *mockUserRepoAuth) CreateBatch(ctx context.Context, us []*entities.Users) error {
	return m.Called(ctx, us).Error(0)
}
func (m *mockUserRepoAuth) CreateUserWithDefaults(ctx context.Context, msisdn string, ref *uuid.UUID) (*entities.Users, error) {
	args := m.Called(ctx, msisdn, ref)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Users), args.Error(1)
}
func (m *mockUserRepoAuth) FindByID(ctx context.Context, id uuid.UUID) (*entities.Users, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Users), args.Error(1)
}
func (m *mockUserRepoAuth) FindByMSISDN(ctx context.Context, msisdn string) (*entities.Users, error) {
	args := m.Called(ctx, msisdn)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Users), args.Error(1)
}
func (m *mockUserRepoAuth) FindByReferralCode(ctx context.Context, code string) (*entities.Users, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Users), args.Error(1)
}
func (m *mockUserRepoAuth) FindByEmail(ctx context.Context, email string) (*entities.Users, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Users), args.Error(1)
}
func (m *mockUserRepoAuth) FindAll(ctx context.Context, limit, offset int) ([]*entities.Users, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*entities.Users), args.Error(1)
}
func (m *mockUserRepoAuth) FindByLoyaltyTier(ctx context.Context, tier string) ([]*entities.Users, error) {
	args := m.Called(ctx, tier)
	return args.Get(0).([]*entities.Users), args.Error(1)
}
func (m *mockUserRepoAuth) FindActiveUsers(ctx context.Context, limit, offset int) ([]*entities.Users, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*entities.Users), args.Error(1)
}
func (m *mockUserRepoAuth) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockUserRepoAuth) Update(ctx context.Context, u *entities.Users) error {
	return m.Called(ctx, u).Error(0)
}
func (m *mockUserRepoAuth) UpdateStatus(ctx context.Context, id uuid.UUID, active bool) error {
	return m.Called(ctx, id, active).Error(0)
}
func (m *mockUserRepoAuth) UpdatePoints(ctx context.Context, id uuid.UUID, pts int) error {
	return m.Called(ctx, id, pts).Error(0)
}
func (m *mockUserRepoAuth) UpdateLoyaltyTier(ctx context.Context, id uuid.UUID, tier string) error {
	return m.Called(ctx, id, tier).Error(0)
}
func (m *mockUserRepoAuth) UpdateRechargeStats(ctx context.Context, id uuid.UUID, amount float64) error {
	return m.Called(ctx, id, amount).Error(0)
}
func (m *mockUserRepoAuth) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// ─── Helper ───────────────────────────────────────────────────────────────────

func newAuthSvc(otpRepo *mockOTPRepo, userRepo *mockUserRepoAuth) *services.AuthService {
	return services.NewAuthService(
		otpRepo,
		userRepo,
		"test-jwt-secret-that-is-long-enough",
		"test-admin-jwt-secret-long-enough",
		15*time.Minute,
		"", // smsAPIKey — not used in unit tests
		"test",
	)
}

// ─── SendOTP tests ────────────────────────────────────────────────────────────

func TestSendOTP_InvalidMSISDN_ReturnsBadRequest(t *testing.T) {
	otpRepo := &mockOTPRepo{}
	userRepo := &mockUserRepoAuth{}

	svc := newAuthSvc(otpRepo, userRepo)
	_, err := svc.SendOTP(context.Background(), "not-a-phone", "login")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid phone number")
	// No DB calls expected for an invalid number
	otpRepo.AssertNotCalled(t, "CountRecentOTPs")
	otpRepo.AssertNotCalled(t, "Create")
}

func TestSendOTP_RateLimitExceeded_ReturnsError(t *testing.T) {
	otpRepo := &mockOTPRepo{}
	userRepo := &mockUserRepoAuth{}

	// CountRecentOTPs returns >= 3 → rate limit
	otpRepo.On("CountRecentOTPs", mock.Anything, mock.Anything, mock.Anything).
		Return(int64(3), nil)

	svc := newAuthSvc(otpRepo, userRepo)
	_, err := svc.SendOTP(context.Background(), "08012345678", "login")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RATE_LIMIT_EXCEEDED")
	otpRepo.AssertNotCalled(t, "Create")
}

func TestSendOTP_UnderRateLimit_CreatesOTPRecord(t *testing.T) {
	otpRepo := &mockOTPRepo{}
	userRepo := &mockUserRepoAuth{}

	otpRepo.On("CountRecentOTPs", mock.Anything, mock.Anything, mock.Anything).
		Return(int64(0), nil)
	otpRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.OTP")).
		Return(nil)

	svc := newAuthSvc(otpRepo, userRepo)
	_, err := svc.SendOTP(context.Background(), "08012345678", "login")

	// May error on SMS send (no real key), but the OTP record must be created
	otpRepo.AssertCalled(t, "Create", mock.Anything, mock.AnythingOfType("*entities.OTP"))
	// The error (if any) is from SMS, not from the OTP creation path
	_ = err
}

func TestSendOTP_CountRecentOTPsDBError_ReturnsError(t *testing.T) {
	otpRepo := &mockOTPRepo{}
	userRepo := &mockUserRepoAuth{}

	otpRepo.On("CountRecentOTPs", mock.Anything, mock.Anything, mock.Anything).
		Return(int64(0), errors.New("db connection lost"))

	svc := newAuthSvc(otpRepo, userRepo)
	_, err := svc.SendOTP(context.Background(), "08012345678", "login")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit")
	otpRepo.AssertNotCalled(t, "Create")
}

// ─── VerifyOTP tests ──────────────────────────────────────────────────────────

func TestVerifyOTP_OTPNotFound_ReturnsFailure(t *testing.T) {
	otpRepo := &mockOTPRepo{}
	userRepo := &mockUserRepoAuth{}

	otpRepo.On("FindValidOTPWithPurpose", mock.Anything, mock.Anything, "111111", "login").
		Return(nil, errors.New("not found"))
	// FindLatestPendingOTP may be called to increment failed attempt counter
	otpRepo.On("FindLatestPendingOTP", mock.Anything, mock.Anything, "login").
		Return(nil, errors.New("not found"))

	svc := newAuthSvc(otpRepo, userRepo)
	token, user, newUser, err := svc.VerifyOTP(context.Background(), "08012345678", "111111", "login")

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Nil(t, user)
	assert.False(t, newUser)
}

func TestVerifyOTP_InvalidMSISDN_ReturnsBadRequest(t *testing.T) {
	otpRepo := &mockOTPRepo{}
	userRepo := &mockUserRepoAuth{}

	svc := newAuthSvc(otpRepo, userRepo)
	token, user, newUser, err := svc.VerifyOTP(context.Background(), "bad-msisdn", "123456", "login")

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Nil(t, user)
	assert.False(t, newUser)
}
