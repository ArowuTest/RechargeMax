/**
 * api-client-extensions.ts
 *
 * All admin extension APIs have been merged into api-client.ts.
 * This file is kept as a re-export shim so existing imports continue to work
 * without requiring individual component updates.
 *
 * @deprecated Import directly from '@/lib/api-client' for new code.
 */

export type {
  SubscriptionTier,
  SubscriptionPricing,
  DailySubscription,
  SubscriptionStatistics,
  SubscriptionBilling,
  USSDRecharge,
  USSDStatistics,
  USSDWebhookLog,
  DrawExportRequest,
  DrawExportHistory,
  WinnerImportRequest,
  ClaimApprovalRequest,
  PayoutRequest,
  ShippingUpdateRequest,
  Winner,
  PointsAdjustment,
  PointsHistory,
  RechargeTransaction,
  RechargeStats,
  NetworkConfig,
  VTPassStatus,
  AdminUser,
  Affiliate,
  AffiliateStats,
} from './api-client';

export {
  subscriptionTierApi,
  subscriptionPricingApi,
  dailySubscriptionApi,
  subscriptionMonitoringApi,
  ussdRechargeApi,
  drawCSVApi,
  winnerClaimApi,
  userPointsApi,
  rechargeMonitoringApi,
  userManagementApi,
  affiliateManagementApi,
} from './api-client';
