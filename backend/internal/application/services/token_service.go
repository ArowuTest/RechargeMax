package services

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

type TokenService struct {
	blacklistRepo repositories.TokenBlacklistRepository
}

func NewTokenService(blacklistRepo repositories.TokenBlacklistRepository) *TokenService {
	return &TokenService{
		blacklistRepo: blacklistRepo,
	}
}

// BlacklistToken adds a token to the blacklist
func (s *TokenService) BlacklistToken(ctx context.Context, tokenString string, adminID uuid.UUID, reason string) error {
	// Parse token to get expiry time
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return fmt.Errorf("invalid token format: %w", err)
	}
	
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("invalid token claims")
	}
	
	// Get expiry time from token
	exp, ok := claims["exp"].(float64)
	if !ok {
		return fmt.Errorf("token missing expiry")
	}
	expiresAt := time.Unix(int64(exp), 0)
	
	// Create blacklist entry
	blacklist := &entities.TokenBlacklist{
		ID:        uuid.New(),
		Token:     tokenString,
		AdminID:   adminID,
		Reason:    reason,
		ExpiresAt: expiresAt,
	}
	
	return s.blacklistRepo.Create(ctx, blacklist)
}

// IsTokenBlacklisted checks if a token is blacklisted
func (s *TokenService) IsTokenBlacklisted(ctx context.Context, tokenString string) (bool, error) {
	return s.blacklistRepo.IsBlacklisted(ctx, tokenString)
}

// CleanupExpiredTokens removes expired tokens from the blacklist
// This should be called periodically (e.g., via cron job)
func (s *TokenService) CleanupExpiredTokens(ctx context.Context) error {
	return s.blacklistRepo.DeleteExpired(ctx)
}

// RevokeAllAdminTokens revokes all tokens for a specific admin
// Useful when admin account is compromised or password is changed
func (s *TokenService) RevokeAllAdminTokens(ctx context.Context, adminID string) error {
	return s.blacklistRepo.DeleteByAdminID(ctx, adminID)
}
