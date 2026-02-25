/**
 * API Client Extensions - Missing Admin Endpoints
 * These endpoints extend the existing api-client.ts with new admin functionality
 */

import { apiClient, ApiResponse, PaginatedResponse } from './api-client';

// ============================================================================
// SUBSCRIPTION TIER MANAGEMENT
// ============================================================================

export interface SubscriptionTier {
  id: string;
  name: string;
  entries_count: number;
  description?: string;
  is_active: boolean;
  display_order: number;
  created_at: string;
  updated_at: string;
}

export interface SubscriptionPricing {
  id: string;
  price_per_entry: number;
  effective_from: string;
  is_active: boolean;
  created_by: string;
  created_at: string;
}

export const subscriptionTierApi = {
  // Get all tiers
  getAll: async () => {
    const response = await apiClient.get<ApiResponse<SubscriptionTier[]>>('/admin/subscription-tiers');
    return response.data;
  },

  // Get tier by ID
  getById: async (tierId: string) => {
    const response = await apiClient.get<ApiResponse<SubscriptionTier>>(`/admin/subscription-tiers/${tierId}`);
    return response.data;
  },

  // Create tier
  create: async (data: Partial<SubscriptionTier>) => {
    const response = await apiClient.post<ApiResponse<SubscriptionTier>>('/admin/subscription-tiers', data);
    return response.data;
  },

  // Update tier
  update: async (tierId: string, data: Partial<SubscriptionTier>) => {
    const response = await apiClient.put<ApiResponse<SubscriptionTier>>(`/admin/subscription-tiers/${tierId}`, data);
    return response.data;
  },

  // Delete tier
  delete: async (tierId: string) => {
    const response = await apiClient.delete<ApiResponse>(`/admin/subscription-tiers/${tierId}`);
    return response.data;
  },

  // Toggle tier active status
  toggleActive: async (tierId: string) => {
    const response = await apiClient.patch<ApiResponse>(`/admin/subscription-tiers/${tierId}/toggle-active`);
    return response.data;
  },
};

export const subscriptionPricingApi = {
  // Get current pricing
  getCurrent: async () => {
    const response = await apiClient.get<ApiResponse<SubscriptionPricing>>('/admin/subscription-pricing/current');
    return response.data;
  },

  // Get pricing history
  getHistory: async () => {
    const response = await apiClient.get<ApiResponse<SubscriptionPricing[]>>('/admin/subscription-pricing/history');
    return response.data;
  },

  // Update pricing
  update: async (pricePerEntry: number) => {
    const response = await apiClient.post<ApiResponse<SubscriptionPricing>>('/admin/subscription-pricing', {
      price_per_entry: pricePerEntry,
    });
    return response.data;
  },
};

// ============================================================================
// DAILY SUBSCRIPTION MONITORING
// ============================================================================

export interface DailySubscription {
  id: string;
  user_id: string;
  msisdn: string;
  tier_id: string;
  tier_name: string;
  bundle_quantity: number;
  total_entries: number;
  price_per_entry: number;
  daily_amount: number;
  daily_cost?: number;  // Alternative field name
  entries_per_day?: number;  // Number of entries per day
  start_date?: string;  // Subscription start date
  status: 'active' | 'paused' | 'cancelled' | 'failed';
  next_billing_date: string;
  created_at: string;
  updated_at: string;
}

export interface SubscriptionStatistics {
  total_subscriptions: number;
  active_subscriptions: number;
  paused_subscriptions: number;
  cancelled_subscriptions: number;
  total_revenue: number;
  monthly_revenue: number;
  total_billings: number;
  successful_billings: number;
  failed_billings: number;
  average_subscription_value: number;
}

export interface SubscriptionBilling {
  id: string;
  subscription_id: string;
  msisdn: string;
  billing_date: string;
  amount: number;
  status: 'pending' | 'success' | 'failed';
  billing_status?: 'pending' | 'success' | 'failed';  // Alternative field name
  payment_reference?: string;
  error_message?: string;
  retry_count: number;
  entries_allocated?: number;  // Number of entries allocated for this billing
  created_at: string;
}

export const dailySubscriptionApi: any = {
  // Get all subscriptions
  getAll: async (page = 1, perPage = 50, status?: string) => {
    let url = `/admin/daily-subscriptions?page=${page}&per_page=${perPage}`;
    if (status) url += `&status=${status}`;
    const response = await apiClient.get<ApiResponse<PaginatedResponse<DailySubscription>>>(url);
    return response.data;
  },

  // Get subscription by ID
  getById: async (subscriptionId: string) => {
    const response = await apiClient.get<ApiResponse<DailySubscription>>(`/admin/daily-subscriptions/${subscriptionId}`);
    return response.data;
  },

  // Cancel subscription
  cancel: async (subscriptionId: string, reason?: string) => {
    const response = await apiClient.post<ApiResponse>(`/admin/daily-subscriptions/${subscriptionId}/cancel`, { reason });
    return response.data;
  },

  // Pause subscription
  pause: async (subscriptionId: string) => {
    const response = await apiClient.post<ApiResponse>(`/admin/daily-subscriptions/${subscriptionId}/pause`);
    return response.data;
  },

  // Resume subscription
  resume: async (subscriptionId: string) => {
    const response = await apiClient.post<ApiResponse>(`/admin/daily-subscriptions/${subscriptionId}/resume`);
    return response.data;
  },

  // Get billing history
  getBillings: async (subscriptionId: string) => {
    const response = await apiClient.get<ApiResponse<SubscriptionBilling[]>>(`/admin/daily-subscriptions/${subscriptionId}/billings`);
    return response.data;
  },

  // Retry failed billing
  retryBilling: async (billingId: string) => {
    const response = await apiClient.post<ApiResponse>(`/admin/subscription-billings/${billingId}/retry`);
    return response.data;
  },

  // Get subscription statistics
  getStats: async () => {
    const response = await apiClient.get<ApiResponse>('/admin/daily-subscriptions/stats');
    return response.data;
  },

  // Get subscriptions (alias for getAll)
  getSubscriptions: async (page = 1, perPage = 50, status?: string) => {
    return dailySubscriptionApi.getAll(page, perPage, status);
  },

  // Get statistics (alias)
  getStatistics: async () => {
    return dailySubscriptionApi.getStats();
  },

  // Export subscriptions
  exportSubscriptions: async (filters?: any) => {
    const response = await apiClient.get<ApiResponse>('/admin/daily-subscriptions/export', { params: filters });
    return response.data;
  },

  // Export billings
  exportBillings: async (filters?: any) => {
    const response = await apiClient.get<ApiResponse>('/admin/subscription-billings/export', { params: filters });
    return response.data;
  },
};

// ============================================================================
// USSD RECHARGE MONITORING
// ============================================================================

export interface USSDRecharge {
  id: string;
  msisdn: string;
  network: string;
  amount: number;
  points_earned: number;
  transaction_reference: string;
  transaction_date: string;
  status: 'pending' | 'completed' | 'failed';
  webhook_received_at: string;
  processed_at?: string;
  error_message?: string;
  created_at: string;
}

export interface USSDWebhookLog {
  id: string;
  network: string;
  payload: any;
  status: 'success' | 'failed';
  error_message?: string;
  retry_count: number;
  created_at: string;
}

export const ussdRechargeApi = {
  // Get all USSD recharges
  getAll: async (page = 1, perPage = 50, network?: string, status?: string) => {
    let url = `/admin/ussd-recharges?page=${page}&per_page=${perPage}`;
    if (network) url += `&network=${network}`;
    if (status) url += `&status=${status}`;
    const response = await apiClient.get<ApiResponse<PaginatedResponse<USSDRecharge>>>(url);
    return response.data;
  },

  // Get USSD recharge by ID
  getById: async (rechargeId: string) => {
    const response = await apiClient.get<ApiResponse<USSDRecharge>>(`/admin/ussd-recharges/${rechargeId}`);
    return response.data;
  },

  // Get webhook logs
  getWebhookLogs: async (page = 1, perPage = 50) => {
    const response = await apiClient.get<ApiResponse<PaginatedResponse<USSDWebhookLog>>>(`/admin/ussd-webhooks?page=${page}&per_page=${perPage}`);
    return response.data;
  },

  // Retry webhook
  retryWebhook: async (webhookId: string) => {
    const response = await apiClient.post<ApiResponse>(`/admin/ussd-webhooks/${webhookId}/retry`);
    return response.data;
  },

  // Get USSD statistics
  getStats: async () => {
    const response = await apiClient.get<ApiResponse>('/admin/ussd-recharges/stats');
    return response.data;
  },
};

// ============================================================================
// DRAW CSV MANAGEMENT
// ============================================================================

export interface DrawExportRequest {
  draw_id: string;
  start_date: string;
  end_date: string;
  include_subscription_points: boolean;
  include_ussd_points: boolean;
  include_wheel_points: boolean;
}

export interface DrawExportHistory {
  id: string;
  draw_id: string;
  exported_by: string;
  exported_at: string;
  total_msisdns: number;
  total_points: number;
  file_url: string;
}

export interface WinnerImportRequest {
  draw_id: string;
  csv_file: File;
}

export const drawCSVApi = {
  // Export draw entries to CSV
  exportCSV: async (data: DrawExportRequest) => {
    const response = await apiClient.post<ApiResponse<{ file_url: string; total_msisdns: number }>>('/admin/draws/export-csv', data);
    return response.data;
  },

  // Import winners from CSV
  importWinners: async (drawId: string, file: File) => {
    const formData = new FormData();
    formData.append('draw_id', drawId);
    formData.append('file', file);

    const response = await apiClient.post<ApiResponse<{ total_winners: number; total_runners_up: number }>>('/admin/draws/import-winners', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  },

  // Get export history
  getExportHistory: async (drawId?: string) => {
    let url = '/admin/draws/export-history';
    if (drawId) url += `?draw_id=${drawId}`;
    const response = await apiClient.get<ApiResponse<DrawExportHistory[]>>(url);
    return response.data;
  },

  // Download CSV file
  downloadCSV: async (fileUrl: string) => {
    const response = await apiClient.get(fileUrl, {
      responseType: 'blob',
    });
    return response.data;
  },
};

// ============================================================================
// WINNER CLAIM PROCESSING
// ============================================================================

export interface Winner {
  id: string;
  draw_id: string;
  msisdn: string;
  prize_name: string;
  prize_type: 'airtime' | 'data' | 'points' | 'cash' | 'physical_goods';
  prize_value: number;
  claim_status: 'unclaimed' | 'claim_submitted' | 'processing' | 'paid' | 'shipped' | 'rejected';
  claim_submitted_at?: string;
  bank_name?: string;
  bank_code?: string;
  account_number?: string;
  account_name?: string;
  shipping_address?: string;
  shipping_phone?: string;
  payout_status?: string;
  payout_reference?: string;
  notification_sent: boolean;
  tracking_number?: string;  // Shipping tracking number
  draw_date?: string;  // Date of the draw
  created_at: string;
  updated_at: string;
}

export const winnerClaimApi: any = {
  // Get all winners
  getAll: async (page = 1, perPage = 50, claimStatus?: string, prizeType?: string) => {
    let url = `/admin/winners?page=${page}&per_page=${perPage}`;
    if (claimStatus) url += `&claim_status=${claimStatus}`;
    if (prizeType) url += `&prize_type=${prizeType}`;
    const response = await apiClient.get<ApiResponse<PaginatedResponse<Winner>>>(url);
    return response.data;
  },

  // Get winner by ID
  getById: async (winnerId: string) => {
    const response = await apiClient.get<ApiResponse<Winner>>(`/admin/winners/${winnerId}`);
    return response.data;
  },

  // Approve claim
  approveClaim: async (winnerId: string) => {
    const response = await apiClient.post<ApiResponse>(`/admin/winners/${winnerId}/approve-claim`);
    return response.data;
  },

  // Reject claim
  rejectClaim: async (winnerId: string, reason: string) => {
    const response = await apiClient.post<ApiResponse>(`/admin/winners/${winnerId}/reject-claim`, { reason });
    return response.data;
  },

  // Process cash payout
  processPayout: async (winnerId: string, payoutReference: string) => {
    const response = await apiClient.post<ApiResponse>(`/admin/winners/${winnerId}/process-payout`, { payout_reference: payoutReference });
    return response.data;
  },

  // Mark as shipped
  markShipped: async (winnerId: string, trackingNumber: string) => {
    const response = await apiClient.post<ApiResponse>(`/admin/winners/${winnerId}/mark-shipped`, { tracking_number: trackingNumber });
    return response.data;
  },

  // Send notification
  sendNotification: async (winnerId: string, channels: string[]) => {
    const response = await apiClient.post<ApiResponse>(`/admin/winners/${winnerId}/send-notification`, { channels });
    return response.data;
  },

  // Get claim statistics
  getStats: async () => {
    const response = await apiClient.get<ApiResponse>('/admin/winners/stats');
    return response.data;
  },

  // Get winners (alias for getAll)
  getWinners: async (page = 1, perPage = 50, claimStatus?: string, prizeType?: string) => {
    return winnerClaimApi.getAll(page, perPage, claimStatus, prizeType);
  },

  // Get winner details (alias for getById)
  getWinnerDetails: async (winnerId: string) => {
    return winnerClaimApi.getById(winnerId);
  },

  // Get claim statistics (alias)
  getClaimStatistics: async () => {
    return winnerClaimApi.getStats();
  },

  // Update shipping information
  updateShipping: async (winnerId: string, trackingNumber: string) => {
    return winnerClaimApi.markShipped(winnerId, trackingNumber);
  },

  // Invoke runner-up (placeholder - needs backend implementation)
  invokeRunnerUp: async (winnerId: string) => {
    const response = await apiClient.post<ApiResponse>(`/admin/winners/${winnerId}/invoke-runner-up`);
    return response.data;
  },
};

// ============================================================================
// USER POINTS MANAGEMENT
// ============================================================================

export interface PointsAdjustment {
  user_id: string;
  points: number;
  reason: string;
  type: 'add' | 'subtract';
}

export interface PointsHistory {
  id: string;
  user_id: string;
  msisdn: string;
  points: number;
  source: 'recharge' | 'subscription' | 'ussd' | 'wheel_spin' | 'manual_adjustment';
  description: string;
  created_at: string;
}

export const userPointsApi = {
  // Get user points history
  getHistory: async (userId: string) => {
    const response = await apiClient.get<ApiResponse<PointsHistory[]>>(`/admin/users/${userId}/points-history`);
    return response.data;
  },

  // Adjust user points
  adjustPoints: async (data: PointsAdjustment) => {
    const response = await apiClient.post<ApiResponse>('/admin/users/adjust-points', data);
    return response.data;
  },

  // Get points statistics
  getStats: async () => {
    const response = await apiClient.get<ApiResponse>('/admin/points/stats');
    return response.data;
  },
};

// ============================================================================
// USER MANAGEMENT EXTENSIONS (OLD - MERGED BELOW)
// ============================================================================
// Removed duplicate export - merged with new userManagementApi below

// ============================================================================
// AFFILIATE MANAGEMENT EXTENSIONS (OLD - MERGED BELOW)
// ============================================================================
// Removed duplicate export - merged with new affiliateManagementApi below


// Subscription Monitoring API (alias for dailySubscriptionApi)
export const subscriptionMonitoringApi = dailySubscriptionApi;


// ============================================================================
// RECHARGE MONITORING & MANAGEMENT
// ============================================================================

export interface RechargeTransaction {
  id: string;
  msisdn: string;
  network_provider: string;
  recharge_type: string;
  amount: number;
  status: string;
  payment_reference: string;
  payment_method: string;
  created_at: string;
  completed_at?: string;
  failure_reason?: string;
  provider_reference?: string;
  retry_count?: number;
  user_id: string;
  customer_name?: string;
  customer_email?: string;
}

export interface RechargeStats {
  total_today: number;
  success_today: number;
  failed_today: number;
  pending_today: number;
  total_amount_today: number;
  success_rate: number;
  avg_processing_time: number;
  stuck_count: number;
}

export interface NetworkConfig {
  network: string;
  provider: 'vtpass' | 'direct';
  enabled: boolean;
  success_rate: number;
}

export interface VTPassStatus {
  enabled: boolean;
  api_connected: boolean;
  last_check: string;
}

export const rechargeMonitoringApi: any = {
  // Get all recharge transactions with filters
  getTransactions: async (params?: {
    page?: number;
    limit?: number;
    status?: string;
    network?: string;
    search?: string;
  }) => {
    const response = await apiClient.get<ApiResponse<RechargeTransaction[]>>('/admin/recharge/transactions', {
      params,
    });
    return response.data;
  },

  // Get recharge statistics
  getStats: async () => {
    const response = await apiClient.get<ApiResponse<RechargeStats>>('/admin/recharge/stats');
    return response.data;
  },

  // Retry failed recharge
  retry: async (transactionId: string) => {
    const response = await apiClient.post<ApiResponse>(`/admin/recharge/${transactionId}/retry`);
    return response.data;
  },

  // Get VTPass status
  getVTPassStatus: async () => {
    const response = await apiClient.get<ApiResponse<VTPassStatus>>('/admin/recharge/vtpass/status');
    return response.data;
  },

  // Update provider configuration (VTPass/Direct switching)
  updateProviderConfig: async (data: {
    provider: 'vtpass' | 'direct';
    network: string;
    enabled: boolean;
  }) => {
    const response = await apiClient.put<ApiResponse>('/admin/recharge/provider-config', data);
    return response.data;
  },

  // Get network configurations
  getNetworkConfigs: async () => {
    const response = await apiClient.get<ApiResponse<NetworkConfig[]>>('/admin/recharge/network-configs');
    return response.data;
  },

  // Get transaction details
  getTransactionDetails: async (transactionId: string) => {
    const response = await apiClient.get<ApiResponse>(`/admin/recharge/transactions/${transactionId}`);
    return response.data;
  },

  // Retry transaction (alias)
  retryTransaction: async (transactionId: string) => {
    return rechargeMonitoringApi.retry(transactionId);
  },

  // Refund transaction
  refundTransaction: async (transactionId: string, reason?: string) => {
    const response = await apiClient.post<ApiResponse>(`/admin/recharge/${transactionId}/refund`, { reason });
    return response.data;
  },

  // Mark transaction as success
  markSuccess: async (transactionId: string) => {
    const response = await apiClient.post<ApiResponse>(`/admin/recharge/${transactionId}/mark-success`);
    return response.data;
  },

  // Mark transaction as failed
  markFailed: async (transactionId: string, reason?: string) => {
    const response = await apiClient.post<ApiResponse>(`/admin/recharge/${transactionId}/mark-failed`, { reason });
    return response.data;
  },
};

// ============================================================================
// USER MANAGEMENT
// ============================================================================

export interface AdminUser {
  id: string;
  msisdn: string;
  email?: string;
  full_name?: string;
  status: 'active' | 'suspended' | 'banned';
  created_at: string;
  last_login?: string;
  total_recharges: number;
  total_spent: number;
  points_balance: number;
}

export const userManagementApi = {
  // Get all users with pagination
  getAll: async (params?: {
    page?: number;
    limit?: number;
    search?: string;
  }) => {
    const response = await apiClient.get<ApiResponse<AdminUser[]>>('/admin/users/all', {
      params,
    });
    return response.data;
  },

  // Get user details
  getDetails: async (userId: string) => {
    const response = await apiClient.get<ApiResponse<AdminUser>>(`/admin/users/${userId}/details`);
    return response.data;
  },

  // Update user status
  updateStatus: async (userId: string, data: {
    status: 'active' | 'suspended' | 'banned';
    reason?: string;
  }) => {
    const response = await apiClient.put<ApiResponse>(`/admin/users/${userId}/status`, data);
    return response.data;
  },

  // Suspend user (from old API)
  suspend: async (userId: string, reason: string) => {
    const response = await apiClient.post<ApiResponse>(`/admin/users/${userId}/suspend`, { reason });
    return response.data;
  },

  // Activate user (from old API)
  activate: async (userId: string) => {
    const response = await apiClient.post<ApiResponse>(`/admin/users/${userId}/activate`);
    return response.data;
  },

  // Delete user (from old API)
  delete: async (userId: string) => {
    const response = await apiClient.delete<ApiResponse>(`/admin/users/${userId}`);
    return response.data;
  },
};

// ============================================================================
// AFFILIATE MANAGEMENT
// ============================================================================

export interface Affiliate {
  id: string;
  user_id: string;
  code: string;
  status: 'pending' | 'active' | 'suspended' | 'rejected';
  total_referrals: number;
  active_referrals: number;
  total_commission: number;
  pending_commission: number;
  paid_commission: number;
  created_at: string;
  approved_at?: string;
  rejected_at?: string;
  rejection_reason?: string;
}

export interface AffiliateStats {
  total_referrals: number;
  active_referrals: number;
  total_commission: number;
  pending_commission: number;
  paid_commission: number;
}

export const affiliateManagementApi = {
  // Get all affiliates with pagination
  getAll: async (params?: {
    page?: number;
    limit?: number;
    status?: string;
  }) => {
    const response = await apiClient.get<ApiResponse<Affiliate[]>>('/admin/affiliates/all', {
      params,
    });
    return response.data;
  },

  // Get affiliate statistics
  getStats: async (affiliateId: string) => {
    const response = await apiClient.get<ApiResponse<AffiliateStats>>(`/admin/affiliates/${affiliateId}/stats`);
    return response.data;
  },

  // Approve affiliate
  approve: async (affiliateId: string) => {
    const response = await apiClient.post<ApiResponse>(`/admin/affiliates/${affiliateId}/approve`);
    return response.data;
  },

  // Reject affiliate
  reject: async (affiliateId: string, reason: string) => {
    const response = await apiClient.post<ApiResponse>(`/admin/affiliates/${affiliateId}/reject`, {
      reason,
    });
    return response.data;
  },

  // Create affiliate (from old API)
  create: async (data: any) => {
    const response = await apiClient.post<ApiResponse>('/admin/affiliates', data);
    return response.data;
  },

  // Update affiliate (from old API)
  update: async (affiliateId: string, data: any) => {
    const response = await apiClient.put<ApiResponse>(`/admin/affiliates/${affiliateId}`, data);
    return response.data;
  },

  // Delete affiliate (from old API)
  delete: async (affiliateId: string) => {
    const response = await apiClient.delete<ApiResponse>(`/admin/affiliates/${affiliateId}`);
    return response.data;
  },

  // Process payout (from old API)
  processPayout: async (affiliateId: string, amount: number, payoutReference: string) => {
    const response = await apiClient.post<ApiResponse>(`/admin/affiliates/${affiliateId}/process-payout`, {
      amount,
      payout_reference: payoutReference,
    });
    return response.data;
  },

  // Get payout history (from old API)
  getPayoutHistory: async (affiliateId: string) => {
    const response = await apiClient.get<ApiResponse>(`/admin/affiliates/${affiliateId}/payout-history`);
    return response.data;
  },
};
