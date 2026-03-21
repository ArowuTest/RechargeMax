// Package routes wires all HTTP routes onto a gin.Engine.
//
// Design principles (enterprise-grade):
//   - main.go owns configuration and DI; routes/ owns routing only.
//   - Each route file in this package registers one logical domain.
//   - The public Register function is the single entry-point called from main.go.
//   - No business logic lives here; every line is a router.METHOD → handler.Method call.
package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rechargemax/internal/application/services"
	"rechargemax/internal/middleware"
	"rechargemax/internal/presentation/handlers"
)

// Register configures the full route tree and returns the ready Engine.
func Register(
	hdlrs *handlers.Registry,
	svcs  *services.Registry,
	db    *gorm.DB,
) *gin.Engine {
	if gin.Mode() == "" {
		gin.SetMode(gin.DebugMode)
	}
	router := gin.New()

	// ── Global middleware ────────────────────────────────────────────────────
	router.Use(gin.Recovery())
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.SecurityHeadersMiddleware())
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.RequestSizeLimitMiddleware(10 * 1024 * 1024)) // 10 MB
	router.Use(middleware.RateLimitMiddleware(100))                      // 100 req/min per IP

	// Initialise PostgreSQL-backed CSRF token store (INFRA-001)
	middleware.InitCSRF(db)

	// ── Infrastructure endpoints (no auth, no versioning) ───────────────────
	registerInfra(router)

	// ── Debug endpoint (temp) ────────────────────────────────────────────────
	debugHandler := handlers.NewHealthHandler(db)
	router.GET("/debug/db", debugHandler.DebugDB)

	// ── API v1 ───────────────────────────────────────────────────────────────
	v1 := router.Group("/api/v1")
	registerPublic(v1, hdlrs, db)
	registerProtected(v1, hdlrs, svcs)
	registerAdmin(v1, hdlrs, svcs, db)

	return router
}

// ────────────────────────────────────────────────────────────────────────────
// Infrastructure
// ────────────────────────────────────────────────────────────────────────────

func registerInfra(r *gin.Engine) {
	// Liveness probe
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"service":   "rechargemax-api",
			"version":   "1.1.0",
			"build":     "20260318-process-stuck-fix",
		})
	})

	// CSRF token issuance (SEC-007) — call before any state-changing request
	r.GET("/csrf-token", middleware.GetCSRFTokenHandler())
}

// ────────────────────────────────────────────────────────────────────────────
// Public routes  (no authentication required)
// ────────────────────────────────────────────────────────────────────────────

func registerPublic(v1 *gin.RouterGroup, hdlrs *handlers.Registry, db *gorm.DB) {
	otpLimiter := middleware.NewOTPRateLimiter(db, 5, time.Minute)

	// Auth
	auth := v1.Group("/auth")
	{
		auth.POST("/send-otp",   middleware.OTPRateLimit(otpLimiter), hdlrs.Auth.SendOTP)
		auth.POST("/verify-otp", middleware.OTPRateLimit(otpLimiter), hdlrs.Auth.VerifyOTP)
		auth.POST("/logout",     hdlrs.Auth.Logout) // user logout — clears httpOnly cookie
	}

	// Admin auth (public — login doesn't require a token)
	adminAuth := v1.Group("/admin/auth")
	{
		adminAuth.POST("/login",  hdlrs.AdminAuth.Login)
		adminAuth.POST("/logout", hdlrs.AdminAuth.Logout)
	}
	v1.POST("/admin/login", hdlrs.AdminAuth.Login) // legacy alias

	// Networks
	networks := v1.Group("/networks")
	{
		networks.GET("",                     hdlrs.Network.GetNetworks)
		networks.GET("/:networkId/bundles",  hdlrs.Network.GetDataBundles)
		networks.POST("/validate",           hdlrs.Network.ValidatePhoneNetwork)
		networks.POST("/cached",             hdlrs.Network.GetCachedNetwork)
		networks.POST("/validate-selection", hdlrs.Network.ValidateNetworkSelection)
	}

	// Guest recharge (no account needed)
	recharge := v1.Group("/recharge")
	{
		recharge.POST("/airtime",             hdlrs.Recharge.InitiateAirtimeRecharge)
		recharge.POST("/data",                hdlrs.Recharge.InitiateDataRecharge)
		recharge.GET("/:id",                  hdlrs.Recharge.GetRecharge)
		recharge.GET("/reference/:reference", hdlrs.Recharge.GetRechargeByReference)
		recharge.POST("/process/:reference",   hdlrs.Recharge.ProcessStuckRecharge)
	}

	// Payment (Paystack callbacks are public)
	payment := v1.Group("/payment")
	{
		payment.POST("/initialize",       hdlrs.Payment.InitializePayment)
		payment.GET("/verify/:reference", hdlrs.Payment.VerifyPayment)
		payment.POST("/webhook",          hdlrs.Payment.HandleWebhook)
		payment.GET("/callback",          hdlrs.Payment.HandleCallback)
		payment.GET("/callback/success",  hdlrs.Payment.HandleSuccess)
		payment.GET("/callback/cancel",   hdlrs.Payment.HandleCancel)
	}

	// Webhooks (payment gateway push notifications)
	v1.POST("/webhooks/paystack", hdlrs.Webhook.HandlePaystackWebhook)

	// Platform statistics (public homepage data)
	v1.GET("/platform/statistics", hdlrs.Platform.GetStatistics)
	v1.GET("/winners/recent",       hdlrs.Platform.GetRecentWinners)

	// Subscription — public & guest-accessible
	// Config is always public (pricing display before sign-up).
	// Create / status / cancel / history use OptionalAuth: if a JWT is present the
	// user's MSISDN is taken from the token; otherwise the request body MSISDN is used.
	v1.GET("/subscription/config", hdlrs.Subscription.GetConfig)

	subPublic := v1.Group("/subscription", middleware.OptionalAuthMiddleware())
	{
		subPublic.POST("/create",  hdlrs.Subscription.CreateSubscription)
		subPublic.GET("/status",   hdlrs.Subscription.GetSubscription)
		subPublic.POST("/cancel",  hdlrs.Subscription.CancelSubscription)
		subPublic.GET("/history",  hdlrs.Subscription.GetHistory)
	}

	// /subscriptions/daily/* aliases — frontend api.ts also calls these paths
	subDailyPublic := v1.Group("/subscriptions/daily", middleware.OptionalAuthMiddleware())
	{
		subDailyPublic.POST("",                        hdlrs.Subscription.CreateSubscription)
		subDailyPublic.GET("/status",                  hdlrs.Subscription.GetSubscription)
		subDailyPublic.POST("/:subscriptionId/cancel", hdlrs.Subscription.CancelSubscription)
		subDailyPublic.GET("/history",                 hdlrs.Subscription.GetHistory)
	}

	// Draws (public browsing)
	draws := v1.Group("/draws")
	{
		draws.GET("",             hdlrs.Draw.GetDraws)
		draws.GET("/active",      hdlrs.Draw.GetActiveDraws)
		draws.GET("/:id",         hdlrs.Draw.GetDrawByID)
		draws.GET("/:id/winners", hdlrs.Draw.GetDrawWinners)
		draws.GET("/:id/results", hdlrs.Draw.GetDrawWinners) // alias — frontend calls /results
		draws.GET("/my-entries",  hdlrs.Draw.GetMyEntries)
	}

	// Spin wheel — public endpoints only (prizes list, tiers).
	// /spin/play, /spin/eligibility, /spin/history are moved to registerProtected
	// so they always require a valid JWT.
	spin := v1.Group("/spin")
	{
		spin.GET("/prizes", hdlrs.Spin.GetPrizes)
		// Guest-accessible play: OptionalAuth — MSISDN comes from JWT if present,
		// otherwise from request body (with strict 4-hour transaction window check).
		spin.POST("/play", middleware.OptionalAuthMiddleware(), hdlrs.Spin.PlaySpin)
	}

	spins := v1.Group("/spins", middleware.OptionalAuthMiddleware())
	{
		spins.GET("/tiers",         hdlrs.Spin.GetTiers)
		spins.GET("/tier-progress", hdlrs.Spin.GetTierProgress)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Protected routes  (user JWT required)
// ────────────────────────────────────────────────────────────────────────────

func registerProtected(v1 *gin.RouterGroup, hdlrs *handlers.Registry, svcs *services.Registry) {
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(svcs.Auth, svcs.Token))
	protected.Use(middleware.CSRFMiddleware()) // SEC-007

	// User profile & wallet
	user := protected.Group("/user")
	{
		user.GET("/dashboard",    hdlrs.User.GetDashboard)
		user.GET("/profile",      hdlrs.User.GetProfile)
		user.POST("/profile",     hdlrs.User.UpdateProfile) // profile update
		user.GET("/wallet",       hdlrs.User.GetWallet)
		user.GET("/transactions", hdlrs.User.GetTransactions)
		user.GET("/prizes",       hdlrs.User.GetPrizes)
	}

	// Authenticated recharge
	recharge := protected.Group("/recharge")
	{
		recharge.POST("/initiate", hdlrs.Recharge.InitiateRecharge)
		recharge.GET("/history",   hdlrs.Recharge.GetHistory)
	}

	// Subscription routes are fully handled by the OptionalAuth group registered
	// in registerPublic (/subscription/* and /subscriptions/daily/*).
	// No protected aliases are needed — OptionalAuthMiddleware already sets
	// msisdn/user_id from the JWT when one is present.

	// Affiliate programme
	affiliate := protected.Group("/affiliate")
	{
		affiliate.GET("/code",          hdlrs.Affiliate.GetReferralCode)
		affiliate.GET("/stats",         hdlrs.Affiliate.GetStats)
		affiliate.GET("/referrals",     hdlrs.Affiliate.GetReferrals)
		affiliate.GET("/dashboard",     hdlrs.Affiliate.GetDashboard)
		affiliate.POST("/register",     hdlrs.Affiliate.Register)
		affiliate.GET("/referral-link", hdlrs.Affiliate.GetReferralCode)
		affiliate.GET("/link",          hdlrs.Affiliate.GetReferralLink)
		affiliate.GET("/commissions",   hdlrs.Affiliate.GetCommissions)
		affiliate.GET("/earnings",      hdlrs.Affiliate.GetEarnings)
		affiliate.POST("/payout",       hdlrs.Affiliate.RequestPayout)
		affiliate.POST("/track-click",  hdlrs.Affiliate.TrackClick) // click attribution
	}

	// Prize claims
	winner := protected.Group("/winner")
	{
		winner.GET("/my-wins",    hdlrs.Winner.GetMyWins)
		winner.POST("/:id/claim", hdlrs.Winner.ClaimPrize)
	}

	// Spin wheel (authenticated actions) — eligibility and history require JWT.
	// /spin/play lives in registerPublic with OptionalAuth to allow guest spins.
	spinProtected := protected.Group("/spin")
	{
		spinProtected.GET("/eligibility", hdlrs.Spin.CheckEligibility)
		spinProtected.GET("/history",     hdlrs.Spin.GetHistory)
	}

	// Notifications
	notifications := protected.Group("/notifications")
	{
		notifications.GET("",              hdlrs.Notification.GetNotifications)
		notifications.GET("/unread-count", hdlrs.Notification.GetUnreadCount)
		notifications.POST("/:id/read",    hdlrs.Notification.MarkAsRead)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Admin routes  (admin JWT + audit logging required)
// ────────────────────────────────────────────────────────────────────────────

func registerAdmin(v1 *gin.RouterGroup, hdlrs *handlers.Registry, svcs *services.Registry, db *gorm.DB) {
	admin := v1.Group("/admin")
	admin.Use(middleware.AdminAuthMiddleware(svcs.Auth, svcs.Token))
	admin.Use(middleware.AdminAuditMiddleware(db))
	admin.Use(middleware.CSRFMiddleware()) // SEC-007 — admin routes mutate state and use httpOnly cookies

	// Dashboard
	admin.GET("/dashboard", hdlrs.Admin.GetDashboardStats)
	admin.GET("/users",     hdlrs.Admin.GetUsers)

	// ── Draws ────────────────────────────────────────────────────────────────
	admin.GET("/draws",                          hdlrs.Admin.GetDraws)
	admin.POST("/draws",                         hdlrs.Draw.CreateDraw)
	admin.PUT("/draws/:id",                      hdlrs.Draw.UpdateDraw)
	admin.POST("/draws/:id/execute",             hdlrs.Draw.ExecuteDraw)
	admin.GET("/draws/:id/export",               hdlrs.Draw.ExportEntries)
	admin.POST("/draws/:id/import-winners",      hdlrs.Draw.ImportWinners)
	admin.GET("/draws/:id/csv/export",           hdlrs.AdminComprehensive.ExportDrawToCSV)
	admin.POST("/draws/:id/csv/import-winners",  hdlrs.AdminComprehensive.ImportWinnersFromCSV)
	admin.GET("/draws/export-history",           hdlrs.AdminComprehensive.GetDrawExportHistory)

	// ── Prize tier system ────────────────────────────────────────────────────
	admin.GET("/draw-types",                         hdlrs.AdminComprehensive.GetDrawTypes)
	admin.GET("/prize-templates",                    hdlrs.AdminComprehensive.GetPrizeTemplates)
	admin.GET("/prize-templates/:id",                hdlrs.AdminComprehensive.GetPrizeTemplate)
	admin.POST("/prize-templates",                   hdlrs.AdminComprehensive.CreatePrizeTemplate)
	admin.PUT("/prize-templates/:id",                hdlrs.AdminComprehensive.UpdatePrizeTemplate)
	admin.DELETE("/prize-templates/:id",             hdlrs.AdminComprehensive.DeletePrizeTemplate)
	admin.POST("/prize-templates/:id/categories",    hdlrs.AdminComprehensive.AddPrizeCategory)
	admin.PUT("/prize-categories/:id",               hdlrs.AdminComprehensive.UpdatePrizeCategory)
	admin.DELETE("/prize-categories/:id",            hdlrs.AdminComprehensive.DeletePrizeCategory)

	// ── Subscriptions ────────────────────────────────────────────────────────
	admin.GET("/subscription-tiers",                       hdlrs.AdminComprehensive.GetSubscriptionTiers)
	admin.POST("/subscription-tiers",                      hdlrs.AdminComprehensive.CreateSubscriptionTier)
	admin.PUT("/subscription-tiers/:id",                   hdlrs.AdminComprehensive.UpdateSubscriptionTier)
	admin.DELETE("/subscription-tiers/:id",                hdlrs.AdminComprehensive.DeleteSubscriptionTier)
	admin.PATCH("/subscription-tiers/:id/toggle-active",   hdlrs.AdminComprehensive.ToggleSubscriptionTier)
	admin.GET("/subscription-pricing/current",    hdlrs.AdminComprehensive.GetCurrentPricing)
	admin.GET("/subscription-pricing/history",    hdlrs.AdminComprehensive.GetPricingHistory)
	admin.POST("/subscription-pricing",           hdlrs.AdminComprehensive.UpdatePricing)
	admin.GET("/daily-subscriptions",                  hdlrs.AdminComprehensive.GetDailySubscriptions)
	admin.GET("/daily-subscriptions/analytics",        hdlrs.AdminComprehensive.GetSubscriptionAnalytics)
	admin.GET("/daily-subscriptions/config",           hdlrs.AdminComprehensive.GetSubscriptionConfig)
	admin.PUT("/daily-subscriptions/config",           hdlrs.AdminComprehensive.UpdateSubscriptionConfig)
	admin.GET("/daily-subscriptions/:id",              hdlrs.AdminComprehensive.GetDailySubscriptionDetails)
	admin.POST("/daily-subscriptions/:id/cancel",      hdlrs.AdminComprehensive.CancelDailySubscription)
	admin.POST("/daily-subscriptions/:id/pause",       hdlrs.AdminComprehensive.PauseDailySubscription)
	admin.POST("/daily-subscriptions/:id/resume",      hdlrs.AdminComprehensive.ResumeDailySubscription)
	admin.GET("/daily-subscriptions/:id/billings",     hdlrs.AdminComprehensive.GetSubscriptionBillingsByID)
	admin.GET("/subscription-billings",                hdlrs.AdminComprehensive.GetSubscriptionBillings)
	admin.POST("/subscription-billings/:id/retry",     hdlrs.AdminComprehensive.RetrySubscriptionBilling)

	// ── USSD ─────────────────────────────────────────────────────────────────
	admin.GET("/ussd/recharges",       hdlrs.AdminComprehensive.GetUSSDRecharges)
	admin.GET("/ussd/recharges/:id",   hdlrs.AdminComprehensive.GetUSSDRechargeByID)
	admin.GET("/ussd/statistics",    hdlrs.AdminComprehensive.GetUSSDStatistics)
	admin.GET("/ussd/webhook-logs",  hdlrs.AdminComprehensive.GetUSSDWebhookLogs)
	admin.POST("/ussd/retry-failed", hdlrs.AdminComprehensive.RetryFailedUSSDWebhooks)

	// ── Points ───────────────────────────────────────────────────────────────
	admin.GET("/points/users",          hdlrs.AdminComprehensive.GetUsersWithPoints)
	admin.GET("/points/history",        hdlrs.AdminComprehensive.GetPointsHistory)
	admin.POST("/points/adjust",        hdlrs.AdminComprehensive.AdjustUserPoints)
	admin.GET("/points/statistics",     hdlrs.AdminComprehensive.GetPointsStatistics)
	admin.GET("/points/stats",          hdlrs.AdminComprehensive.GetPointsStatistics) // alias for frontend
	admin.GET("/points/export/users",   hdlrs.AdminComprehensive.ExportUsersWithPoints)
	admin.GET("/points/export/history", hdlrs.AdminComprehensive.ExportPointsHistory)

	// ── Winners — list, detail, actions ────────────────────────────────────
	admin.GET("/winners",                              hdlrs.AdminComprehensive.GetAllWinners)
	admin.GET("/winners/pending-claims",              hdlrs.AdminComprehensive.GetPendingClaims)
	admin.GET("/winners/claim-statistics",            hdlrs.AdminComprehensive.GetClaimStatistics)
	admin.GET("/winners/:id",                         hdlrs.AdminComprehensive.GetWinnerByID)
	admin.POST("/winners/:id/approve-claim",          hdlrs.AdminComprehensive.ApproveWinnerClaim)
	admin.POST("/winners/:id/reject-claim",           hdlrs.AdminComprehensive.RejectWinnerClaim)
	admin.POST("/winners/:id/process-payout",         hdlrs.AdminComprehensive.ProcessWinnerPayout)
	admin.POST("/winners/:id/mark-shipped",           hdlrs.AdminComprehensive.MarkWinnerShipped)
	admin.POST("/winners/:id/send-notification",      hdlrs.AdminComprehensive.SendWinnerNotification)
	admin.POST("/winners/:id/invoke-runner-up",       hdlrs.AdminComprehensive.InvokeWinnerRunnerUp)
	admin.GET("/prize-fulfillment/failed-provisions", hdlrs.AdminSpinClaims.GetPendingClaims)
	admin.POST("/prize-fulfillment/retry/:id",        hdlrs.AdminSpinClaims.ApproveClaim)
	admin.POST("/prize-fulfillment/retry-all",        hdlrs.AdminSpinClaims.GetPendingClaims)
	admin.POST("/prize-fulfillment/send-reminders",   hdlrs.AdminSpinClaims.SendReminders)

	// ── Spin wheel ───────────────────────────────────────────────────────────
	admin.GET("/spin/config",        hdlrs.AdminComprehensive.GetSpinConfig)
	admin.PUT("/spin/config",        hdlrs.AdminComprehensive.UpdateSpinConfig)
	admin.GET("/spin/prizes",        hdlrs.AdminComprehensive.GetAllPrizes)
	admin.POST("/spin/prizes",       hdlrs.AdminComprehensive.CreatePrize)
	admin.PUT("/spin/prizes/:id",    hdlrs.AdminComprehensive.UpdatePrize)
	admin.DELETE("/spin/prizes/:id", hdlrs.AdminComprehensive.DeletePrize)
	admin.GET("/spin-tiers",         hdlrs.AdminSpinTiers.GetAllTiers)
	admin.GET("/spin-tiers/:id",     hdlrs.AdminSpinTiers.GetTierByID)
	admin.POST("/spin-tiers",        hdlrs.AdminSpinTiers.CreateTier)
	admin.PUT("/spin-tiers/:id",     hdlrs.AdminSpinTiers.UpdateTier)
	admin.DELETE("/spin-tiers/:id",  hdlrs.AdminSpinTiers.DeleteTier)
	admin.GET("/spin/claims",              hdlrs.AdminSpinClaims.ListClaims)
	admin.GET("/spin/claims/pending",      hdlrs.AdminSpinClaims.GetPendingClaims)
	admin.GET("/spin/claims/statistics",   hdlrs.AdminSpinClaims.GetStatistics)
	admin.GET("/spin/claims/export",       hdlrs.AdminSpinClaims.ExportClaims)
	admin.GET("/spin/claims/:id",          hdlrs.AdminSpinClaims.GetClaimDetails)
	admin.POST("/spin/claims/:id/approve", hdlrs.AdminSpinClaims.ApproveClaim)
	admin.POST("/spin/claims/:id/reject",  hdlrs.AdminSpinClaims.RejectClaim)

	// ── Recharge monitoring ──────────────────────────────────────────────────
	admin.GET("/recharge/transactions",        hdlrs.AdminComprehensive.GetRechargeTransactions)
	admin.GET("/recharge/transactions/:id",    hdlrs.AdminComprehensive.GetRechargeByID)
	admin.GET("/recharge/stats",               hdlrs.AdminComprehensive.GetRechargeStats)
	admin.POST("/recharge/:id/retry",          hdlrs.AdminComprehensive.RetryFailedRecharge)
	admin.POST("/recharge/bulk-retry-processing", hdlrs.AdminComprehensive.BulkRetryProcessingTransactions)
	admin.POST("/recharge/:id/refund",         hdlrs.AdminComprehensive.RefundRecharge)
	admin.POST("/recharge/:id/mark-success",   hdlrs.AdminComprehensive.MarkRechargeSuccess)
	admin.POST("/recharge/:id/mark-failed",    hdlrs.AdminComprehensive.MarkRechargeFailed)
	admin.GET("/recharge/vtpass/status",       hdlrs.AdminComprehensive.GetVTPassStatus)
	admin.PUT("/recharge/provider-config",     hdlrs.AdminComprehensive.UpdateProviderConfig)
	admin.GET("/recharge/network-configs",     hdlrs.AdminComprehensive.GetNetworkConfigurations)
	admin.GET("/recharge/data-plans",          hdlrs.AdminComprehensive.GetDataPlans)

	// ── Network & data plan CRUD ─────────────────────────────────────────────
	admin.GET("/networks",        hdlrs.AdminComprehensive.GetNetworkConfigurations) // list
	admin.POST("/networks",       hdlrs.AdminComprehensive.CreateNetwork)
	admin.PUT("/networks/:id",    hdlrs.AdminComprehensive.UpdateNetwork)
	admin.DELETE("/networks/:id", hdlrs.AdminComprehensive.DeleteNetwork)
	admin.POST("/data-plans",       hdlrs.AdminComprehensive.CreateDataPlan)
	admin.PUT("/data-plans/:id",    hdlrs.AdminComprehensive.UpdateDataPlan)
	admin.DELETE("/data-plans/:id", hdlrs.AdminComprehensive.DeleteDataPlan)

	// ── User management ──────────────────────────────────────────────────────
	admin.GET("/users/all",                  hdlrs.AdminComprehensive.GetAllUsers)
	admin.GET("/users/:id/details",          hdlrs.AdminComprehensive.GetUserDetails)
	admin.GET("/users/:id",                  hdlrs.AdminComprehensive.GetUser)           // alias — no /details suffix
	admin.PUT("/users/:id",                  hdlrs.AdminComprehensive.UpdateUser)        // status + tier update
	admin.PUT("/users/:id/status",           hdlrs.AdminComprehensive.UpdateUserStatus)
	admin.DELETE("/users/:id",               hdlrs.AdminComprehensive.DeleteUser)
	admin.POST("/users/:id/suspend",         hdlrs.AdminComprehensive.SuspendUser)
	admin.POST("/users/:id/activate",        hdlrs.AdminComprehensive.ActivateUser)
	admin.GET("/users/:id/points-history",   hdlrs.AdminComprehensive.GetUserPointsHistory)

	// ── Affiliate management ─────────────────────────────────────────────────
	admin.GET("/affiliates/all",                   hdlrs.AdminComprehensive.GetAllAffiliates)
	admin.POST("/affiliates",                      hdlrs.AdminComprehensive.CreateAffiliate)
	admin.PUT("/affiliates/:id",                   hdlrs.AdminComprehensive.UpdateAffiliate)
	admin.DELETE("/affiliates/:id",                hdlrs.AdminComprehensive.DeleteAffiliate)
	admin.GET("/affiliates/:id/stats",             hdlrs.AdminComprehensive.GetAffiliateStats)
	admin.POST("/affiliates/:id/approve",          hdlrs.AdminComprehensive.ApproveAffiliate)
	admin.POST("/affiliates/:id/reject",           hdlrs.AdminComprehensive.RejectAffiliate)
	admin.POST("/affiliates/:id/suspend",          hdlrs.AdminComprehensive.SuspendAffiliate)
	admin.GET("/affiliates/:id/commissions",       hdlrs.AdminComprehensive.GetAffiliateCommissions)
	admin.PUT("/affiliates/:id/commission-rate",   hdlrs.AdminComprehensive.UpdateAffiliateCommissionRate)
	admin.GET("/affiliates/:id/payouts",           hdlrs.AdminComprehensive.GetAffiliatePayouts)
	admin.POST("/affiliates/:id/payout",           hdlrs.AdminComprehensive.ProcessAffiliatePayout)
	admin.POST("/affiliates/:id/process-payout",   hdlrs.AdminComprehensive.ProcessAffiliatePayout)   // alias
	admin.GET("/affiliates/:id/payout-history",    hdlrs.AdminComprehensive.GetAffiliatePayoutHistory)
	admin.GET("/affiliates/analytics",             hdlrs.AdminComprehensive.GetAffiliateAnalytics)

	// ── Admin user management ────────────────────────────────────────────────
	admin.GET("/admins",            hdlrs.AdminUserManagement.GetAllAdmins)
	admin.GET("/admins/:id",        hdlrs.AdminUserManagement.GetAdminByID)
	admin.POST("/admins",           hdlrs.AdminUserManagement.CreateAdmin)
	admin.PUT("/admins/:id",        hdlrs.AdminUserManagement.UpdateAdmin)
	admin.DELETE("/admins/:id",     hdlrs.AdminUserManagement.DeleteAdmin)
	admin.PUT("/admins/:id/status", hdlrs.AdminUserManagement.UpdateAdminStatus)

	// ── Platform settings ────────────────────────────────────────────────────
	admin.GET("/settings",                    hdlrs.PlatformSettings.GetAllSettings)
	admin.PUT("/settings",                    hdlrs.PlatformSettings.UpdateSettings)
	admin.GET("/settings/category/:category", hdlrs.PlatformSettings.GetSettingsByCategory)
	admin.PUT("/settings/category/:category", hdlrs.PlatformSettings.UpdateCategorySettings)
	admin.GET("/settings/:key",               hdlrs.PlatformSettings.GetSetting)
	admin.PUT("/settings/:key",               hdlrs.PlatformSettings.UpdateSetting)

	// ── Commissions ──────────────────────────────────────────────────────────
	admin.POST("/commissions/reconciliation", hdlrs.Commission.GetCommissionReconciliation)
	admin.POST("/commissions/export",         hdlrs.Commission.ExportCommissionReport)

	// ── Validation stats ─────────────────────────────────────────────────────
	admin.GET("/validation/stats",  hdlrs.ValidationStats.GetValidationStats)
	admin.POST("/validation/stats", hdlrs.ValidationStats.GetValidationStats)

	// ── Transaction limits ───────────────────────────────────────────────────
	admin.GET("/transaction-limits",        hdlrs.TransactionLimits.ListTransactionLimits)
	admin.GET("/transaction-limits/:id",    hdlrs.TransactionLimits.GetTransactionLimit)
	admin.POST("/transaction-limits",       hdlrs.TransactionLimits.CreateTransactionLimit)
	admin.PUT("/transaction-limits/:id",    hdlrs.TransactionLimits.UpdateTransactionLimit)
	admin.DELETE("/transaction-limits/:id", hdlrs.TransactionLimits.DeleteTransactionLimit)

	// ── Audit logs ───────────────────────────────────────────────────────────
	admin.GET("/audit-logs", middleware.AdminAuditLogsList(db))
}
