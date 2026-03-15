/**
 * RechargeMax API Client
 * Connects to Go backend API
 */

import axios, { AxiosInstance, AxiosError, AxiosRequestConfig } from 'axios';
import type {
  ApiResponse,
  ApiSuccessResponse,
  ApiErrorResponse,
  PaginatedResponse,
  NetworkConfig,
  DataPlan,
  WheelPrize,
  User,
  Transaction,
  Affiliate,
  DailySubscription,
  isApiSuccess,
  hasData,
} from '@/types/admin-api.types';

// API Configuration
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';
const API_TIMEOUT = parseInt(import.meta.env.VITE_API_TIMEOUT || '30000');

// Create axios instance
// withCredentials: true ensures httpOnly cookies are sent with every cross-origin request
const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: API_TIMEOUT,
  withCredentials: true, // Send httpOnly auth cookies automatically
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor
// Cookies are sent automatically via withCredentials: true
// Authorization header is only added as a fallback for environments where cookies
// are not available (e.g., native mobile wrappers, Postman, server-side calls)
apiClient.interceptors.request.use(
  (config) => {
    // No-op: httpOnly cookies are attached by the browser automatically.
    // The header fallback below is intentionally kept for API/mobile clients
    // that may store the token from the response body in their own secure storage.
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor - Handle errors
apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      const requestUrl = (error.config as any)?.url || '';
      const isAdminRoute = requestUrl.includes('/admin/');
      if (isAdminRoute) {
        // Clear non-sensitive admin profile data (token is in httpOnly cookie, auto-expired by server)
        localStorage.removeItem('rechargemax_admin_user');
        window.location.href = '/#/admin/login';
      } else {
        // Clear non-sensitive user profile data
        localStorage.removeItem('rechargemax_user');
        window.location.href = '/#/login';
      }
    }
    return Promise.reject(error);
  }
);

// ============================================================================
// TYPES (Re-exported from admin-api.types.ts)
// ============================================================================

export type {
  ApiResponse,
  ApiSuccessResponse,
  ApiErrorResponse,
  PaginatedResponse,
  NetworkConfig,
  DataPlan,
  WheelPrize,
  User,
  Transaction,
  Affiliate,
  DailySubscription,
} from '@/types/admin-api.types';

export { isApiSuccess, hasData } from '@/types/admin-api.types';

// ============================================================================
// AUTHENTICATION API
// ============================================================================

export const authApi = {
  // Send OTP to phone number
  sendOTP: async (phoneNumber: string) => {
    const response = await apiClient.post<ApiResponse>('/auth/send-otp', {
      msisdn: phoneNumber,
    });
    return response.data;
  },

  // Verify OTP and login
  verifyOTP: async (phoneNumber: string, otp: string) => {
    const response = await apiClient.post<ApiResponse<{ token: string; user: any }>>('/auth/verify-otp', {
      msisdn: phoneNumber,
      otp: otp,
    });
    
    // Token is now stored as httpOnly cookie by the server
    // Only store non-sensitive user profile data in localStorage
    if (response.data.success && response.data.data) {
      localStorage.setItem('rechargemax_user', JSON.stringify(response.data.data.user));
    }
    
    return response.data;
  },

  // Logout
  logout: async () => {
    try {
      // Backend clears the httpOnly cookie on this call
      await apiClient.post('/auth/logout');
    } finally {
      // Clear non-sensitive user profile data
      localStorage.removeItem('rechargemax_user');
    }
  },

  // Get current user profile from localStorage
  getCurrentUser: () => {
    const userStr = localStorage.getItem('rechargemax_user');
    return userStr ? JSON.parse(userStr) : null;
  },
};

// ============================================================================
// ADMIN AUTHENTICATION API
// ============================================================================

export const adminAuthApi = {
  // Admin login
  login: async (email: string, password: string) => {
    const response = await apiClient.post<ApiResponse<{ token: string; admin: any }>>('/admin/auth/login', {
      email,
      password,
    });
    
    // Token stored as httpOnly cookie by server — only cache non-sensitive admin profile data
    if (response.data.success && response.data.data) {
      localStorage.setItem('rechargemax_admin_user', JSON.stringify(response.data.data.admin));
    }
    
    return response.data;
  },

  // Admin logout
  logout: async () => {
    try {
      // Backend clears the httpOnly admin cookie
      await apiClient.post('/admin/auth/logout');
    } finally {
      // Clear non-sensitive admin profile cache
      localStorage.removeItem('rechargemax_admin_user');
    }
  },
};

// ============================================================================
// RECHARGE API
// ============================================================================

export const rechargeApi = {
  // Get available networks
  getNetworks: async () => {
    const response = await apiClient.get<ApiResponse<any[]>>('/networks');
    return response.data;
  },

  // Get data bundles for a network
  getDataBundles: async (networkId: string) => {
    const response = await apiClient.get<ApiResponse<any[]>>(`/networks/${networkId}/bundles`);
    return response.data;
  },

  // Validate phone number network
  validatePhoneNetwork: async (phoneNumber: string, expectedNetwork: string) => {
    const response = await apiClient.post<ApiResponse<any>>('/networks/validate', {
      phone_number: phoneNumber,
      expected_network: expectedNetwork,
    });
    return response.data;
  },

  // Initialize airtime recharge
  initiateAirtimeRecharge: async (data: {
    phone_number: string;
    network: string;
    amount: number;
  }) => {
    const response = await apiClient.post<ApiResponse>('/recharge/airtime', data);
    return response.data;
  },

  // Initialize data recharge
  initiateDataRecharge: async (data: {
    phone_number: string;
    network: string;
    bundle_id: string;
  }) => {
    const response = await apiClient.post<ApiResponse>('/recharge/data', data);
    return response.data;
  },
};

// ============================================================================
// PAYMENT API
// ============================================================================

export const paymentApi = {
  // Initialize payment
  initializePayment: async (data: {
    amount: number;
    email: string;
    phone_number: string;
    transaction_type: string;
    metadata?: any;
  }) => {
    const response = await apiClient.post<ApiResponse<{ authorization_url: string; reference: string }>>('/payment/initialize', data);
    return response.data;
  },

  // Verify payment
  verifyPayment: async (reference: string) => {
    const response = await apiClient.get<ApiResponse>(`/payment/verify/${reference}`);
    return response.data;
  },
};

// ============================================================================
// SPIN WHEEL API
// ============================================================================

export const spinApi = {
  // Get spin eligibility
  getEligibility: async () => {
    const response = await apiClient.get<ApiResponse<{ available_spins: number; next_reset: string }>>('/spin/eligibility');
    return response.data;
  },

  // Spin the wheel
  spin: async () => {
    const response = await apiClient.post<ApiResponse<{ prize: any }>>('/spin/play');
    return response.data;
  },

  // Get spin history
  getHistory: async () => {
    const response = await apiClient.get<ApiResponse<any[]>>('/spin/history');
    return response.data;
  },
};

// ============================================================================
// DRAW API
// ============================================================================

export const drawApi = {
  // Get active draws
  getActiveDraws: async () => {
    const response = await apiClient.get<ApiResponse<any[]>>('/draws/active');
    return response.data;
  },

  // Get draw details
  getDrawDetails: async (drawId: string) => {
    const response = await apiClient.get<ApiResponse>(`/draws/${drawId}`);
    return response.data;
  },

  // Get user's draw entries
  getMyEntries: async () => {
    const response = await apiClient.get<ApiResponse<any[]>>('/draws/my-entries');
    return response.data;
  },

  // Get draw results
  getDrawResults: async (drawId: string) => {
    const response = await apiClient.get<ApiResponse>(`/draws/${drawId}/results`);
    return response.data;
  },

  // Get draw winners
  getWinners: async (drawId: string) => {
    const response = await apiClient.get<ApiResponse<any[]>>(`/draws/${drawId}/winners`);
    return response.data;
  },
};

// ============================================================================
// SUBSCRIPTION API
// ============================================================================

export const subscriptionApi = {
  // Get subscription status
  getStatus: async () => {
    const response = await apiClient.get<ApiResponse>('/subscription/status');
    return response.data;
  },

  // Subscribe (Create Subscription)
  subscribe: async (msisdn?: string) => {
    const response = await apiClient.post<ApiResponse>('/subscription/create', { msisdn });
    return response.data;
  },

  // Cancel Subscription
  cancel: async () => {
    const response = await apiClient.post<ApiResponse>('/subscription/cancel');
    return response.data;
  },

  // Get subscription history
  getHistory: async () => {
    const response = await apiClient.get<ApiResponse<any[]>>('/subscription/history');
    return response.data;
  },

  // Get public subscription config (pricing)
  getConfig: async () => {
    const response = await apiClient.get<ApiResponse>('/subscription/config');
    return response.data;
  },
};

// ============================================================================
// AFFILIATE API
// ============================================================================

export const affiliateApi = {
  // Register as affiliate
  register: async (data: {
    business_name?: string;
    bank_name: string;
    account_number: string;
    account_name: string;
  }) => {
    const response = await apiClient.post<ApiResponse>('/affiliate/register', data);
    return response.data;
  },

  // Get affiliate dashboard
  getDashboard: async () => {
    const response = await apiClient.get<ApiResponse>('/affiliate/dashboard');
    return response.data;
  },

  // Get referral link
  getReferralLink: async () => {
    const response = await apiClient.get<ApiResponse<{ referral_code: string; referral_link: string }>>('/affiliate/referral-link');
    return response.data;
  },

  // Get commissions
  getCommissions: async () => {
    const response = await apiClient.get<ApiResponse<any[]>>('/affiliate/commissions');
    return response.data;
  },

  // Request payout
  requestPayout: async (amount: number) => {
    const response = await apiClient.post<ApiResponse>('/affiliate/payout', { amount });
    return response.data;
  },

  // Track affiliate click
  trackClick: async () => {
    const affiliateCode = sessionStorage.getItem('affiliate_code');
    if (!affiliateCode) return { success: false };
    
    const response = await apiClient.post<ApiResponse>('/affiliate/track-click', {
      affiliate_code: affiliateCode,
      timestamp: new Date().toISOString(),
      referrer: document.referrer,
      landing_page: window.location.pathname
    });
    return response.data;
  },
};

// ============================================================================
// USER API
// ============================================================================

export const userApi = {
  // Get user dashboard
  getDashboard: async () => {
    const response = await apiClient.get<ApiResponse>('/user/dashboard');
    return response.data;
  },

  // Get user profile
  getProfile: async () => {
    const response = await apiClient.get<ApiResponse>('/user/profile');
    return response.data;
  },

  // Update profile
  updateProfile: async (data: { full_name?: string; email?: string }) => {
    const response = await apiClient.post<ApiResponse>('/user/profile', data);
    return response.data;
  },

  // Get transactions
  getTransactions: async () => {
    const response = await apiClient.get<ApiResponse<any[]>>('/user/transactions');
    return response.data;
  },

  // Get prizes
  getPrizes: async () => {
    const response = await apiClient.get<ApiResponse<any[]>>('/user/prizes');
    return response.data;
  },

  // Claim prize
  claimPrize: async (prizeId: string, claimData: any) => {
    const response = await apiClient.post<ApiResponse>(`/winner/${prizeId}/claim`, claimData);
    return response.data;
  },
};

// ============================================================================
// ADMIN API
// ============================================================================

export const adminApi: any = {
  // Dashboard stats
  getStats: async () => {
    const response = await apiClient.get<ApiResponse>('/admin/dashboard');
    return response.data;
  },

  // Users management
  users: {
    getAll: async (page = 1, perPage = 50) => {
      const response = await apiClient.get<ApiResponse<any>>(`/admin/users/all`);
      return response.data;
    },
    getById: async (userId: string) => {
      const response = await apiClient.get<ApiResponse>(`/admin/users/${userId}`);
      return response.data;
    },
    update: async (userId: string, data: any) => {
      const response = await apiClient.put<ApiResponse>(`/admin/users/${userId}`, data);
      return response.data;
    },
  },

  // Affiliates management
  affiliates: {
    getAll: async () => {
      const response = await apiClient.get<ApiResponse<any[]>>('/admin/affiliates/all');
      return response.data;
    },
    approve: async (affiliateId: string) => {
      const response = await apiClient.post<ApiResponse>(`/admin/affiliates/${affiliateId}/approve`);
      return response.data;
    },
    reject: async (affiliateId: string, reason: string) => {
      const response = await apiClient.post<ApiResponse>(`/admin/affiliates/${affiliateId}/reject`, { reason });
      return response.data;
    },
    suspend: async (affiliateId: string) => {
      const response = await apiClient.post<ApiResponse>(`/admin/affiliates/${affiliateId}/suspend`);
      return response.data;
    },
  },

  // Subscriptions management
  subscriptions: {
    getAll: async () => {
      const response = await apiClient.get<ApiResponse<any[]>>('/admin/daily-subscriptions');
      return response.data;
    },
    getConfig: async () => {
      const response = await apiClient.get<ApiResponse>('/admin/daily-subscriptions/config');
      return response.data;
    },
    updateConfig: async (data: any) => {
      const response = await apiClient.put<ApiResponse>('/admin/daily-subscriptions/config', data);
      return response.data;
    },
  },

  // Spin wheel management
  spin: {
    getConfig: async () => {
      const response = await apiClient.get<ApiResponse>('/admin/spin/config');
      return response.data;
    },
    updateConfig: async (data: any) => {
      const response = await apiClient.put<ApiResponse>('/admin/spin/config', data);
      return response.data;
    },
    getPrizes: async () => {
      const response = await apiClient.get<ApiResponse<any[]>>('/admin/spin/prizes');
      return response.data;
    },
    createPrize: async (data: any) => {
      const response = await apiClient.post<ApiResponse>('/admin/spin/prizes', data);
      return response.data;
    },
    updatePrize: async (prizeId: string, data: any) => {
      const response = await apiClient.put<ApiResponse>(`/admin/spin/prizes/${prizeId}`, data);
      return response.data;
    },
    deletePrize: async (prizeId: string) => {
      const response = await apiClient.delete<ApiResponse>(`/admin/spin/prizes/${prizeId}`);
      return response.data;
    },
  },

  // Data bundles management
  bundles: {
    getAll: async () => {
      // Get all data plans from the admin endpoint
      const response = await apiClient.get<ApiResponse<any[]>>('/admin/recharge/data-plans');
      return response.data;
    },
    create: async (data: any) => {
      const response = await apiClient.post<ApiResponse>('/admin/data-plans', data);
      return response.data;
    },
    update: async (bundleId: string, data: any) => {
      const response = await apiClient.put<ApiResponse>(`/admin/data-plans/${bundleId}`, data);
      return response.data;
    },
    delete: async (bundleId: string) => {
      const response = await apiClient.delete<ApiResponse>(`/admin/data-plans/${bundleId}`);
      return response.data;
    },
  },

  // Networks management
  networks: {
    getAll: async () => {
      const response = await apiClient.get<ApiResponse<any[]>>('/admin/recharge/network-configs');
      return response.data;
    },
    create: async (data: any) => {
      const response = await apiClient.post<ApiResponse>('/admin/networks', data);
      return response.data;
    },
    update: async (networkId: string, data: any) => {
      const response = await apiClient.put<ApiResponse>(`/admin/networks/${networkId}`, data);
      return response.data;
    },
    delete: async (networkId: string) => {
      const response = await apiClient.delete<ApiResponse>(`/admin/networks/${networkId}`);
      return response.data;
    },
  },

  // Draws management
  draws: {
    getAll: async () => {
      const response = await apiClient.get<ApiResponse<any[]>>('/admin/draws');
      return response.data;
    },
    create: async (data: any) => {
      const response = await apiClient.post<ApiResponse>('/admin/draws', data);
      return response.data;
    },
    update: async (drawId: string, data: any) => {
      const response = await apiClient.put<ApiResponse>(`/admin/draws/${drawId}`, data);
      return response.data;
    },
    execute: async (drawId: string) => {
      const response = await apiClient.post<ApiResponse>(`/admin/draws/${drawId}/execute`);
      return response.data;
    },
  },

  // Analytics
  analytics: {
    getOverview: async () => {
      // Use dashboard endpoint for analytics overview
      const response = await apiClient.get<ApiResponse>('/admin/dashboard');
      return response.data;
    },
    getRevenue: async (startDate?: string, endDate?: string) => {
      const params = new URLSearchParams();
      if (startDate) params.append('start_date', startDate);
      if (endDate) params.append('end_date', endDate);
      const response = await apiClient.get<ApiResponse>(`/admin/dashboard?${params}`);
      return response.data;
    },
    getUsers: async () => {
      const response = await apiClient.get<ApiResponse>('/admin/users');
      return response.data;
    },
  },

  // Configuration
  config: {
    get: async (key: string) => {
      const response = await apiClient.get<ApiResponse>(`/admin/settings/${key}`);
      return response.data;
    },
    set: async (key: string, value: any) => {
      const response = await apiClient.put<ApiResponse>('/admin/settings', { [key]: value });
      return response.data;
    },
    getCommissionRates: async () => {
      const response = await apiClient.get<ApiResponse>('/admin/settings');
      return response.data;
    },
    setCommissionRates: async (rates: any) => {
      const response = await apiClient.put<ApiResponse>('/admin/settings', rates);
      return response.data;
    },
  },

  // Convenience methods for backward compatibility
  getNetworks: async () => {
    return await adminApi.networks.getAll();
  },

  getDataPlans: async () => {
    return await adminApi.bundles.getAll();
  },

  getWheelPrizes: async () => {
    return await adminApi.spin.getPrizes();
  },

  // Spin Prize Claims Management
  getSpinClaims: async (params?: any) => {
    const response = await apiClient.get<ApiResponse>('/admin/spin/claims', { params });
    return response.data;
  },

  getSpinClaimDetails: async (claimId: string) => {
    const response = await apiClient.get<ApiResponse>(`/admin/spin/claims/${claimId}`);
    return response.data;
  },

  getPendingSpinClaims: async () => {
    const response = await apiClient.get<ApiResponse>('/admin/spin/claims/pending');
    return response.data;
  },

  getSpinClaimStatistics: async () => {
    const response = await apiClient.get<ApiResponse>('/admin/spin/claims/statistics');
    return response.data;
  },

  approveSpinClaim: async (claimId: string, data: { admin_notes?: string; payment_reference?: string }) => {
    const response = await apiClient.post<ApiResponse>(`/admin/spin/claims/${claimId}/approve`, data);
    return response.data;
  },

  rejectSpinClaim: async (claimId: string, data: { rejection_reason: string; admin_notes?: string }) => {
    const response = await apiClient.post<ApiResponse>(`/admin/spin/claims/${claimId}/reject`, data);
    return response.data;
  },

  exportSpinClaims: async (params?: any) => {
    const response = await apiClient.get('/admin/spin/claims/export', {
      params,
      responseType: 'blob',
    });
    
    // Create download link
    const url = window.URL.createObjectURL(new Blob([response.data]));
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', `spin_claims_${new Date().toISOString().split('T')[0]}.csv`);
    document.body.appendChild(link);
    link.click();
    link.remove();
  },

  // Generic HTTP methods for flexible API calls
  get: async <T = any>(url: string, config?: any) => {
    const response = await apiClient.get<ApiResponse<T>>(url, config);
    return response.data;
  },

  post: async <T = any>(url: string, data?: any, config?: any) => {
    const response = await apiClient.post<ApiResponse<T>>(url, data, config);
    return response.data;
  },

  put: async <T = any>(url: string, data?: any, config?: any) => {
    const response = await apiClient.put<ApiResponse<T>>(url, data, config);
    return response.data;
  },

  delete: async <T = any>(url: string, config?: any) => {
    const response = await apiClient.delete<ApiResponse<T>>(url, config);
    return response.data;
  },
};

// ============================================================================
// HELPER FUNCTION (Legacy compatibility)
// ============================================================================

/**
 * Legacy function for backward compatibility
 * Replaces Supabase edge function calls
 */
export async function callEdgeFunction(
  functionName: string,
  payload: any
): Promise<any> {
  // Map old edge function names to new API endpoints
  const endpointMap: Record<string, () => Promise<any>> = {
    'send-otp': () => authApi.sendOTP(payload.phone_number),
    'verify-otp': () => authApi.verifyOTP(payload.phone_number, payload.otp_code),
    'initialize-payment': () => paymentApi.initializePayment(payload),
    'verify-payment': () => paymentApi.verifyPayment(payload.reference),
    'get-networks': () => rechargeApi.getNetworks(),
    'get-data-bundles': () => rechargeApi.getDataBundles(payload.network_id),
    'spin-wheel': () => spinApi.spin(),
    'get-active-draws': () => drawApi.getActiveDraws(),
    'admin-login': () => adminAuthApi.login(payload.email, payload.password),
  };

  const apiCall = endpointMap[functionName];
  if (apiCall) {
    return await apiCall();
  }

  // Fallback: make a generic POST request
  const response = await apiClient.post(`/${functionName}`, payload);
  return response.data;
}

export default apiClient;
export { apiClient };
