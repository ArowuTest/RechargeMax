package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
	"rechargemax/internal/validation"
)

// AuthService handles authentication operations
type AuthService struct {
	otpRepo       repositories.OTPRepository
	userRepo      repositories.UserRepository
	jwtSecret     string
	jwtExpiration time.Duration
	smsAPIKey     string
	environment   string
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID string `json:"user_id"`
	MSISDN string `json:"msisdn"`
	Role   string `json:"role"` // "user", "admin", "super_admin"
	Type   string `json:"type"` // access, refresh
	jwt.RegisteredClaims
}

// NewAuthService creates a new authentication service
func NewAuthService(
	otpRepo repositories.OTPRepository,
	userRepo repositories.UserRepository,
	jwtSecret string,
	jwtExpiration time.Duration,
	smsAPIKey string,
	environment string,
) *AuthService {
	return &AuthService{
		otpRepo:       otpRepo,
		userRepo:      userRepo,
		jwtSecret:     jwtSecret,
		jwtExpiration: jwtExpiration,
		smsAPIKey:     smsAPIKey,
		environment:   environment,
	}
}

// SendOTP sends an OTP to the user's phone number
func (s *AuthService) SendOTP(ctx context.Context, msisdn string) error {
	// Validate MSISDN format
	if len(msisdn) < 10 {
		return fmt.Errorf("invalid phone number format")
	}

	// Check rate limiting (max 3 OTPs per 10 minutes)
	tenMinutesAgo := time.Now().Add(-10 * time.Minute)
	recentCount, err := s.otpRepo.CountRecentOTPs(ctx, msisdn, tenMinutesAgo)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}

	if recentCount >= 3 {
		return fmt.Errorf("too many OTP requests. Please wait 10 minutes before trying again")
	}

	// Generate 6-digit OTP
	otpCode, err := s.generateOTP()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Create OTP record
	otp := &entities.OTP{
		ID:        uuid.New(),
		Msisdn:    msisdn,
		Code:      otpCode,
		Purpose:   "login",
		ExpiresAt: time.Now().Add(10 * time.Minute), // 10 minutes expiry
		IsUsed:    false,
	}

	if err := s.otpRepo.Create(ctx, otp); err != nil {
		return fmt.Errorf("failed to create OTP record: %w", err)
	}

	// Send SMS
	if err := s.sendSMS(ctx, msisdn, otpCode); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to send SMS to %s: %v\n", msisdn, err)
	}

	return nil
}

// VerifyOTP verifies an OTP and returns authentication tokens and user info
func (s *AuthService) VerifyOTP(ctx context.Context, msisdn, code string) (string, *entities.User, bool, error) {
	// Normalize MSISDN to format 234XXXXXXXXXX
	normalizedMSISDN, err := validation.NormalizeMSISDN(msisdn)
	if err != nil {
		return "", nil, false, fmt.Errorf("invalid phone number format: %w", err)
	}
	
	// Find valid OTP (using original msisdn for OTP lookup)
	otp, err := s.otpRepo.FindValidOTP(ctx, msisdn, code)
	if err != nil {
		return "", nil, false, fmt.Errorf("invalid or expired OTP")
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
		fmt.Printf("Failed to update last login time: %v\n", err)
	}

	// Generate JWT token (using normalized MSISDN)
	fmt.Printf("[DEBUG] Generating JWT for normalized MSISDN: %s (original: %s)\n", normalizedMSISDN, msisdn)
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

// sendSMS sends SMS (integrates with SMS provider)
func (s *AuthService) sendSMS(ctx context.Context, msisdn, code string) error {
	message := fmt.Sprintf("Your RechargeMax verification code is: %s. Valid for 10 minutes. Do not share this code.", code)
	
	// In development, just log the SMS
	if s.environment == "development" {
		fmt.Printf("SMS to %s: %s\n", msisdn, message)
		return nil
	}

	// Implement actual SMS sending via Termii
	// In production, this would:
	// 1. Use NotificationService.SendSMS() if available
	// 2. Or directly call Termii API with smsAPIKey
	// 3. Handle errors and retry logic
	//
	// Example:
	// if s.notificationService != nil {
	//     return s.notificationService.SendSMS(ctx, msisdn, message)
	// }
	// 
	// // Or direct Termii API call:
	// payload := map[string]interface{}{
	//     "to":      msisdn,
	//     "from":    "RechargeMax",
	//     "sms":     message,
	//     "type":    "plain",
	//     "channel": "generic",
	//     "api_key": s.smsAPIKey,
	// }
	// // ... HTTP POST to Termii API
	
	// For now, log the SMS (when integrated with NotificationService, uncomment above)
	fmt.Printf("[SMS-PROD] To: %s, Message: %s\n", msisdn, message)
	
		return nil
	}

// GenerateAdminToken generates a JWT token for admin users
func (s *AuthService) GenerateAdminToken(ctx context.Context, adminID, msisdn string, role string) (string, error) {
	now := time.Now()
	
	// Validate role
	if role != "admin" && role != "super_admin" {
		return "", fmt.Errorf("invalid admin role: %s", role)
	}
	
	// Admin token claims
	claims := JWTClaims{
		UserID: adminID,
		MSISDN: msisdn, // Admin's phone number for tracking
		Role:   role,   // "admin" or "super_admin"
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   adminID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtExpiration)),
			Issuer:    "rechargemax-admin",
		},
	}
	
	// Generate token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
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
