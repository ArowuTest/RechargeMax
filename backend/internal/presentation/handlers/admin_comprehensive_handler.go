package handlers

import (
	"gorm.io/gorm"

	"rechargemax/internal/application/services"
	"rechargemax/internal/domain/repositories"
)

// AdminComprehensiveHandler handles all new admin endpoints
type AdminComprehensiveHandler struct {
	subscriptionTierService *services.SubscriptionTierService
	ussdRechargeService     *services.USSDRechargeService
	pointsService           *services.PointsService
	drawService             *services.DrawService
	winnerService           *services.WinnerService
	spinService             *services.SpinService
	// NEW: Services for recharge/user/affiliate management
	rechargeService      *services.RechargeService
	userService          *services.UserService
	affiliateService     *services.AffiliateService
	telecomService       *services.TelecomService
	networkConfigService *services.NetworkConfigService
	// Repositories for direct data access
	networkRepo         repositories.NetworkRepository
	dataPlanRepo        repositories.DataPlanRepository
	subscriptionService *services.SubscriptionService
	// Prize Tier System Services
	drawTypeService      *services.DrawTypeService
	prizeTemplateService *services.PrizeTemplateService
	// Settings service for platform configuration persistence
	settingsSvc *services.PlatformSettingsService
	// Direct DB access still needed by some comprehensive handlers
	db *gorm.DB
}

// NewAdminComprehensiveHandler creates a new comprehensive admin handler
func NewAdminComprehensiveHandler(
	subscriptionTierService *services.SubscriptionTierService,
	ussdRechargeService *services.USSDRechargeService,
	pointsService *services.PointsService,
	drawService *services.DrawService,
	winnerService *services.WinnerService,
	spinService *services.SpinService,
	rechargeService *services.RechargeService,
	userService *services.UserService,
	affiliateService *services.AffiliateService,
	telecomService *services.TelecomService,
	networkConfigService *services.NetworkConfigService,
	networkRepo repositories.NetworkRepository,
	dataPlanRepo repositories.DataPlanRepository,
	subscriptionService *services.SubscriptionService,
	drawTypeService *services.DrawTypeService,
	prizeTemplateService *services.PrizeTemplateService,
	settingsSvc *services.PlatformSettingsService,
	db *gorm.DB,
) *AdminComprehensiveHandler {
	return &AdminComprehensiveHandler{
		subscriptionTierService: subscriptionTierService,
		ussdRechargeService:     ussdRechargeService,
		pointsService:           pointsService,
		drawService:             drawService,
		winnerService:           winnerService,
		spinService:             spinService,
		rechargeService:         rechargeService,
		userService:             userService,
		affiliateService:        affiliateService,
		telecomService:          telecomService,
		networkConfigService:    networkConfigService,
		networkRepo:             networkRepo,
		dataPlanRepo:            dataPlanRepo,
		subscriptionService:     subscriptionService,
		drawTypeService:         drawTypeService,
		prizeTemplateService:    prizeTemplateService,
		settingsSvc:             settingsSvc,
		db:                      db,
	}
}
