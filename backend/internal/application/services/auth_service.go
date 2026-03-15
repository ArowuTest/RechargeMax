package services

import (
	"go.uber.org/zap"
	"rechargemax/internal/logger"
	"context"
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
	"rechargemax/internal/errors"
	"rechargemax/internal/validation"
)

// AuthService handles authentication operations
type AuthService struct {
	otpRepo             repositories.OTPRepository
	userRepo            repositories.UserRepository
	jwtSecret           string
	adminJWTSecret      string // separate secret for admin tokens
	jwtExpiration       time.Duration
	smsAPIKey           string
	environment         string
	notificationService *NotificationService // used for production SMS delivery (BUG-004)
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID  string `json:"user_id"`
	AdminID string `json:"admin_id"` // Populated in admin tokens; same as UserID for user tokens
	MSISDN  string `json:"msisdn"`
	Role    string `json:"role"` // "user", "admin", "super_admin"
	Type    string `json:"type"` // "access" for users, "admin" for admin tokens
	jwt.RegisteredClaims
}

// NewAuthService creates a new authentication service
func NewAuthService(
	otpRepo repositories.OTPRepository,
	userRepo repositories.UserRepository,
	jwtSecret string,
	adminJWTSecret string,
	jwtExpiration time.Duration,
	smsAPIKey string,
	environment string,
	notificationService ...*NotificationService, // variadic so existing call sites need no changes
) *AuthService {
	svc := &AuthService{
		otpRepo:        otpRepo,
		userRepo:       userRepo,
		jwtSecret:      jwtSecret,
		adminJWTSecret: adminJWTSecret,
		jwtExpiration:  jwtExpiration,
		smsAPIKey:      smsAPIKey,
		environment:    environment,
	}
	if len(notificationService) > 0 {
		svc.notificationService = notificationService[0]
	}
	return svc
}

// SendOTP sends an OTP to the user's phone number
func (s *AuthService) SendOTP(ctx context.Context, msisdn string, purpose string) error {
	// Normalise MSISDN to canonical international format (2348XXXXXXXXX)
	// This ensures OTPs are always stored and looked up using the same format
	// regardless of whether the user supplied 08012345678, 2348012345678, or +2348012345678
	normalizedMSISDN, err := validation.NormalizeMSISDN(msisdn)
	if err != nil {
		return errors.BadRequest("Invalid phone number format: " + err.Error())
	}

	// Check rate limiting (max 3 OTPs per 10 minutes) — using normalised MSISDN
	tenMinutesAgo := time.Now().Add(-10 * time.Minute)
	recentCount, err := s.otpRepo.CountRecentOTPs(ctx, normalizedMSISDN, tenMinutesAgo)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}

	if recentCount >= 3 {
		return errors.RateLimitExceeded().WithDetails(map[string]interface{}{
			"message": "Too many OTP requests. Please wait 10 minutes before trying again.",
		})
	}

	// Generate 6-digit OTP
	otpCode, err := s.generateOTP()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Hash the OTP with bcrypt before storage (plaintext sent only via SMS)
	otpHash, err := bcrypt.GenerateFromPassword([]byte(otpCode), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash OTP: %w", err)
	}

	// Create OTP record — always stored with normalised international MSISDN
	otp := &entities.OTP{
		ID:        uuid.New(),
		MSISDN:    normalizedMSISDN,
		Code:      string(otpHash), // bcrypt hash — NOT plaintext
		Purpose:   purpose,
		ExpiresAt: time.Now().Add(10 * time.Minute), // 10 minutes expiry
		IsUsed:    false,
	}

	if err := s.otpRepo.Create(ctx, otp); err != nil {
		return fmt.Errorf("failed to create OTP record: %w", err)
	}

	// Send SMS (use normalised MSISDN for delivery)
	if err := s.sendSMS(ctx, normalizedMSISDN, otpCode); err != nil {
		// Log error but don't fail the request
		logger.Error("Failed to send SMS to %s: %v", zap.Any("value", normalizedMSISDN), zap.Error(err))
	}

	return nil
}

// VerifyOTP verifies an OTP and returns authentication tokens and user info
func (s *AuthService) VerifyOTP(ctx context.Context, msisdn, code, purpose string) (string, *entities.User, bool, error) {
	// Normalize MSISDN to format 234XXXXXXXXXX
	normalizedMSISDN, err := validation.NormalizeMSISDN(msisdn)
	if err != nil {
		return "", nil, false, errors.BadRequest("Invalid phone number format: " + err.Error())
	}
	
	// Find valid OTP with matching purpose (using normalised MSISDN — same as stored in SendOTP)
	otp, otpErr := s.otpRepo.FindValidOTPWithPurpose(ctx, normalizedMSISDN, code, purpose)
	if otpErr != nil {
		// SEC-008: Increment failed attempts on any pending OTP for this MSISDN.
		// Use FindLatestPendingOTP — no code check so we always find the pending row.
		pendingOTP, _ := s.otpRepo.FindLatestPendingOTP(ctx, normalizedMSISDN, purpose)
		if pendingOTP != nil {
			count, _ := s.otpRepo.IncrementFailedAttempts(ctx, pendingOTP.ID)
			const maxOTPAttempts = 5
			if count >= maxOTPAttempts {
				// Invalidate the OTP — the user must request a new one
				_ = s.otpRepo.InvalidateByMSISDN(ctx, normalizedMSISDN, purpose)
				return "", nil, false, errors.Unauthorized("Too many failed attempts — please request a new OTP")
			}
		}
		return "", nil, false, errors.Unauthorized("Invalid or expired OTP")
	}

	// Mark OTP as used
	if err := s.otpRepo.MarkAsUsed(ctx, otp.ID); err != nil {
		return "", nil, false, fmt.Errorf("failed to mark OTP as used: %w", err)
	}

	// Get or create user (using normalized MSISDN)
	user, err := s.userRepo.FindByMSISDN(ctx, normalizedMSISDN)
	isNewUser := false
	
	if err != nil {
		// User doesn't exist - create using centralized function
		// This ensures consistent user creation across the platform
		user, err = s.userRepo.CreateUserWithDefaults(ctx, normalizedMSISDN, nil)
		if err != nil {
			return "", nil, false, fmt.Errorf("failed to create user: %w", err)
		}
		isNewUser = true
	}

	// Update last login time
	now := time.Now()
	user.LastLoginAt = &now
	if err := s.userRepo.Update(ctx, user); err != nil {
		// Log error but don't fail authentication
		logger.Error("Failed to update last login time: %v", zap.Error(err))
	}

	// Generate JWT token (using normalized MSISDN)
	logger.Info("[DEBUG] Generating JWT for normalized MSISDN: %s (original: %s)", zap.Any("value", normalizedMSISDN), zap.Any("value", msisdn))
	token, err := s.GenerateToken(ctx, normalizedMSISDN)
	if err != nil {
		return "", nil, false, fmt.Errorf("failed to generate token: %w", err)
	}

	return token, user, isNewUser, nil
}

// GenerateToken generates a JWT access token
func (s *AuthService) GenerateToken(ctx context.Context, msisdn string) (string, error) {
	// Get user
	user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}

	now := time.Now()
	
	// Access token claims
	claims := JWTClaims{
		UserID: user.ID.String(),
		MSISDN: user.MSISDN,
		Role:   "user", // Regular users get "user" role
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtExpiration)),
			Issuer:    "rechargemax",
		},
	}

	// Generate token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns claims
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GetUserByMSISDN gets a user by MSISDN
func (s *AuthService) GetUserByMSISDN(ctx context.Context, msisdn string) (*entities.User, error) {
	return s.userRepo.FindByMSISDN(ctx, msisdn)
}

// generateOTP generates a 6-digit OTP
func (s *AuthService) generateOTP() (string, error) {
	// Generate cryptographically secure 6-digit number
	max := big.NewInt(999999)
	min := big.NewInt(100000)
	
	n, err := rand.Int(rand.Reader, max.Sub(max, min))
	if err != nil {
		return "", err
	}
	
	return fmt.Sprintf("%06d", n.Add(n, min).Int64()), nil
}

// sendSMS sends the OTP message via NotificationService (production) or logs it
// to stdout in development. The plaintext code is NEVER logged in production
// to prevent OTP exposure in server logs (BUG-004).
func (s *AuthService) sendSMS(ctx context.Context, msisdn, code string) error {
	message := fmt.Sprintf("Your RechargeMax verification code is: %s. Valid for 10 minutes. Do not share this code.", code)

	if s.environment != "production" {
		// Development / staging: log without the code to avoid accidental exposure
		msisdnSuffix := msisdn
		if len(msisdn) > 4 {
			msisdnSuffix = msisdn[len(msisdn)-4:]
		}
		logger.Info("[SMS-DEV] OTP dispatched", zap.String("msisdn_suffix", "..."+msisdnSuffix))
		return nil
	}

	// Production: route through NotificationService which calls Termii API
	if s.notificationService != nil {
		return s.notificationService.SendSMS(ctx, msisdn, message)
	}

	// Fallback: NotificationService not wired — log a warning (no OTP code in log)
	msisdnSuffix2 := msisdn
	if len(msisdn) > 4 {
		msisdnSuffix2 = msisdn[len(msisdn)-4:]
	}
	logger.Warn("[SMS-WARN] NotificationService unavailable", zap.String("msisdn_suffix", "..."+msisdnSuffix2))
	return fmt.Errorf("SMS delivery unavailable: NotificationService not configured")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GenerateAdminToken generates a JWT token for admin users
func (s *AuthService) GenerateAdminToken(ctx context.Context, adminID, msisdn string, role string) (string, error) {
	now := time.Now()
	
	// Validate role
	if role != "admin" && role != "super_admin" {
		return "", fmt.Errorf("invalid admin role: %s", role)
	}
	
	// Admin token claims — must use "admin" type so AdminAuthMiddleware accepts it.
	// AdminID is populated so middleware can read it without falling back to UserID.
	claims := JWTClaims{
		AdminID: adminID, // Primary field for admin tokens
		UserID:  adminID, // Also set for backward compat with any code reading UserID
		MSISDN:  msisdn,
		Role:    role, // "admin" or "super_admin"
		Type:    "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   adminID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(8 * time.Hour)), // Admin sessions expire in 8h
			Issuer:    "rechargemax-admin",
		},
	}

	// Sign with the admin-specific secret (must differ from user JWT secret)
	signingSecret := s.adminJWTSecret
	if signingSecret == "" {
		signingSecret = s.jwtSecret // dev fallback only
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(signingSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign admin token: %w", err)
	}
	
	return tokenString, nil
}

// CleanupExpiredOTPs removes expired OTP records
func (s *AuthService) CleanupExpiredOTPs(ctx context.Context) error {
	// Delete expired OTPs
	if err := s.otpRepo.DeleteExpired(ctx); err != nil {
		return fmt.Errorf("failed to delete expired OTPs: %w", err)
	}

	// Delete old used OTPs (older than 24 hours)
	oneDayAgo := time.Now().Add(-24 * time.Hour)
	if err := s.otpRepo.DeleteOld(ctx, oneDayAgo); err != nil {
		return fmt.Errorf("failed to delete old OTPs: %w", err)
	}

	return nil
}
