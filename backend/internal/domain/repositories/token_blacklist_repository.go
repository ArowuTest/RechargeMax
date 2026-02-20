package repositories

import (
	"context"

	"rechargemax/internal/domain/entities"
)

// TokenBlacklistRepository defines methods for token blacklist operations
type TokenBlacklistRepository interface {
	// Create adds a token to the blacklist
	Create(ctx context.Context, blacklist *entities.TokenBlacklist) error
	
	// IsBlacklisted checks if a token is blacklisted
	IsBlacklisted(ctx context.Context, token string) (bool, error)
	
	// DeleteExpired removes expired tokens from the blacklist
	DeleteExpired(ctx context.Context) error
	
	// DeleteByAdminID removes all blacklisted tokens for an admin
	DeleteByAdminID(ctx context.Context, adminID string) error
}
