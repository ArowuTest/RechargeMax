/**
 * Enterprise-Grade Admin API Type Definitions
 * 
 * These types match the backend API responses exactly to ensure type safety.
 * All types are derived from the Go backend structures.
 */

// ============================================================================
// CORE API RESPONSE TYPES
// ============================================================================

/**
 * Successful API response with data
 */
export interface ApiSuccessResponse<T = any> {
  success: true;
  data: T;
  message?: string;
}

/**
 * Failed API response with error
 */
export interface ApiErrorResponse {
  success: false;
  error: string;
  message?: string;
}

/**
 * Discriminated union type for all API responses
 * Can be used as ApiResponse<T> where T is the data type
 */
export type ApiResponse<T = any> = ApiSuccessResponse<T> | ApiErrorResponse;

/**
 * Paginated response wrapper
 */
export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

// ============================================================================
// NETWORK & DATA PLAN TYPES
// ============================================================================

/**
 * Network configuration (matches backend NetworkResponse)
 */
export interface NetworkConfig {
  id: string;
  name: string;
  network_name?: string;  // Alternative field name
  code: string;
  network_code?: string;  // Alternative field name
  logo: string;
  is_active: boolean;
  support_data: boolean;
  data_enabled?: boolean;  // Alternative field name
  support_airtime: boolean;
  airtime_enabled?: boolean;  // Alternative field name
  commission_rate?: number;
  minimum_amount?: number;
  maximum_amount?: number;
}

/**
 * Data plan configuration
 */
export interface DataPlan {
  id: string;
  plan_code: string;
  plan_name: string;
  network_code: string;
  network_id?: string;  // Alternative field name
  network_provider?: string;  // Alternative field name
  data_amount: string;
  validity: string;
  validity_days?: number;  // Alternative field name (numeric)
  price: number;
  is_active: boolean;
  display_order: number;
  sort_order?: number;  // Alternative field name
  network_configs_2025_11_10_13_30?: {  // Relation field
    network_name: string;
    network_code: string;
  };
  created_at: string;
  updated_at: string;
}

// ============================================================================
// WHEEL PRIZE TYPES
// ============================================================================

/**
 * Wheel prize configuration
 */
export interface WheelPrize {
  id: string;
  prize_name: string;
  prize_type: 'airtime' | 'data' | 'points' | 'cash' | 'physical_goods' | 'AIRTIME' | 'DATA' | 'POINTS' | 'CASH' | 'PHYSICAL_GOODS';
  prize_value: number;
  probability_weight: number;
  probability?: number;  // Alternative field name (percentage)
  minimum_recharge?: number;
  is_active: boolean;
  display_order: number;
  sort_order?: number;  // Alternative field name
  color: string;
  color_scheme?: string;  // Alternative field name
  icon?: string;
  icon_name?: string;  // Alternative field name
  description?: string;
  created_at: string;
  updated_at: string;
}

// ============================================================================
// DASHBOARD STATISTICS TYPES
// ============================================================================

/**
 * Admin dashboard statistics
 */
export interface DashboardStats {
  total_users: number;
  new_users_today: number;
  total_transactions: number;
  transactions_today: number;
  total_prizes: number;
  total_affiliates: number;
  pending_affiliates: number;
  approved_affiliates: number;
  total_revenue: number;
  today_revenue: number;
  total_commissions: number;
}

// ============================================================================
// USER MANAGEMENT TYPES
// ============================================================================

/**
 * User entity
 */
export interface User {
  id: string;
  phone_number: string;
  email?: string;
  full_name?: string;
  points: number;
  total_recharges: number;
  total_spent: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  last_login?: string;
}

/**
 * User with additional stats
 */
export interface UserWithStats extends User {
  total_spins: number;
  total_prizes_won: number;
  affiliate_code?: string;
  referred_by?: string;
}

// ============================================================================
// TRANSACTION TYPES
// ============================================================================

/**
 * Transaction entity
 */
export interface Transaction {
  id: string;
  user_id: string;
  transaction_type: 'recharge' | 'subscription' | 'prize_claim' | 'commission_payout';
  amount: number;
  points_earned: number;
  status: 'pending' | 'completed' | 'failed' | 'cancelled';
  payment_reference?: string;
  payment_method?: string;
  network_provider?: string;
  phone_number?: string;
  description?: string;
  created_at: string;
  completed_at?: string;
}

// ============================================================================
// AFFILIATE TYPES
// ============================================================================

/**
 * Affiliate entity
 */
export interface Affiliate {
  id: string;
  user_id: string;
  affiliate_code: string;
  status: 'pending' | 'approved' | 'rejected' | 'suspended';
  commission_rate: number;
  total_referrals: number;
  total_commissions_earned: number;
  total_commissions_paid: number;
  pending_commissions: number;
  bank_name?: string;
  account_number?: string;
  account_name?: string;
  created_at: string;
  approved_at?: string;
  rejected_at?: string;
  rejection_reason?: string;
}

/**
 * Affiliate with user details
 */
export interface AffiliateWithUser extends Affiliate {
  user_phone: string;
  user_email?: string;
  user_name?: string;
}

// ============================================================================
// PRIZE TYPES
// ============================================================================

/**
 * Prize claim entity
 */
export interface PrizeClaim {
  id: string;
  user_id: string;
  prize_id: string;
  prize_name: string;
  prize_type: string;
  prize_value: number;
  claim_status: 'pending' | 'approved' | 'rejected' | 'paid' | 'shipped';
  claim_date: string;
  processed_date?: string;
  payment_reference?: string;
  tracking_number?: string;
  rejection_reason?: string;
}

// ============================================================================
// ADMIN USER TYPES
// ============================================================================

/**
 * Admin user entity
 */
export interface AdminUser {
  id: string;
  username: string;
  email: string;
  full_name?: string;
  role: 'SUPER_ADMIN' | 'ADMIN' | 'SUPPORT' | 'VIEWER';
  is_active: boolean;
  permissions: string[];
  created_at: string;
  updated_at: string;
  last_login?: string;
}

/**
 * Admin creation request
 */
export interface CreateAdminRequest {
  username: string;
  email: string;
  password: string;
  role: 'SUPER_ADMIN' | 'ADMIN' | 'SUPPORT' | 'VIEWER';
  permissions?: string[];
}

// ============================================================================
// SETTINGS TYPES
// ============================================================================

/**
 * System settings
 */
export interface SystemSettings {
  id: string;
  key: string;
  setting_key?: string;  // Alternative field name
  value: string;
  setting_value?: any;  // Alternative field name (can be any type)
  description?: string;
  updated_at: string;
  updated_by: string;
}

/**
 * Platform settings (alias for SystemSettings)
 */
export type PlatformSetting = SystemSettings;

// ============================================================================
// DAILY SUBSCRIPTION TYPES
// ============================================================================

/**
 * Daily subscription entity (matches backend exactly)
 */
export interface DailySubscription {
  id: string;
  user_id?: string;
  msisdn?: string;
  tier_id?: string;
  tier_name?: string;
  bundle_quantity?: number;
  total_entries?: number;
  price_per_entry?: number;
  daily_amount?: number;
  status?: 'active' | 'paused' | 'cancelled' | 'failed' | 'pending';
  next_billing_date?: string;
  created_at?: string;
  updated_at?: string;
  // Additional fields used in some contexts
  amount?: number;
  draw_entries_earned?: number;
  is_paid?: boolean;
  subscription_code?: string;
}

/**
 * Subscription tier configuration
 */
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

/**
 * Subscription pricing configuration
 */
export interface SubscriptionPricing {
  id: string;
  price_per_entry: number;
  effective_from: string;
  is_active: boolean;
  created_by: string;
  created_at: string;
}

// ============================================================================
// TYPE GUARDS
// ============================================================================

/**
 * Type guard to check if API response is successful
 */
export function isApiSuccess<T>(response: ApiResponse<T>): response is ApiSuccessResponse<T> {
  return response.success === true;
}

/**
 * Type guard to check if API response has data
 */
export function hasData<T>(response: ApiResponse<T>): response is ApiSuccessResponse<T> {
  return isApiSuccess(response) && response.data !== undefined;
}

/**
 * Type guard to check if API response is an error
 */
export function isApiError<T>(response: ApiResponse<T>): response is ApiErrorResponse {
  return response.success === false;
}

// ============================================================================
// UTILITY TYPES
// ============================================================================

/**
 * Extract data type from API response
 */
export type ExtractData<T> = T extends ApiResponse<infer U> ? U : never;

/**
 * Make all properties optional recursively
 */
export type DeepPartial<T> = {
  [P in keyof T]?: T[P] extends object ? DeepPartial<T[P]> : T[P];
};
