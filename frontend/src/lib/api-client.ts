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
  DataPlan,
  WheelPrize,
  User,
  Transaction,
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

// ---------------------------------------------------------------------------
// CSRF token management (SEC-007)
// ---------------------------------------------------------------------------

// Simple in-memory CSRF token cache (per browser tab session).
// The token is fetched once from GET /csrf-token and reused until it expires or
// the server rejects it (403), at which point it is cleared and refetched.
const csrfCache: { token: string | null; fetchingPromise: Promise<string> | null } = {
  token: null,
  fetchingPromise: null,
};

// Determines if a request method requires a CSRF token.
const requiresCSRF = (method?: string): boolean =>
  ['post', 'put', 'patch', 'delete'].includes((method || '').toLowerCase());

// Fetches a fresh CSRF token from the backend and caches it.
async function fetchCSRFToken(): Promise<string> {
  // Deduplicate concurrent requests
  if (csrfCache.fetchingPromise) return csrfCache.fetchingPromise;
  csrfCache.fetchingPromise = (async () => {
    try {
      // Use the raw axios instance to avoid triggering the interceptors recursively
      const baseURL = import.meta.env.VITE_API_BASE_URL?.replace('/api/v1', '') || 'http://localhost:8080';
      const res = await axios.get<{ csrf_token: string }>(`${baseURL}/csrf-token`, {
        withCredentials: true,
      });
      const token = res.data.csrf_token;
      csrfCache.token = token;
      return token;
    } finally {
      csrfCache.fetchingPromise = null;
    }
  })();
  return csrfCache.fetchingPromise;
}

// Pre-warm the CSRF token on load so the first mutation does not pay the round-trip cost.
// Errors are silently swallowed; the request interceptor below will retry if needed.
fetchCSRFToken().catch(() => {});

// ---------------------------------------------------------------------------
// Request interceptor — attach CSRF token + rely on cookies for auth
// ---------------------------------------------------------------------------
apiClient.interceptors.request.use(
  async (config) => {
    // httpOnly cookies are sent automatically via withCredentials: true.
    // However, SameSite=Lax blocks cookies on cross-origin XHR/fetch (vercel → onrender).
    // So we also store the token in localStorage and send it as Bearer for ALL routes.
    const adminToken = localStorage.getItem('rechargemax_admin_token');
    const userToken = localStorage.getItem('rechargemax_user_token');
    if (adminToken && config.url?.includes('/admin/')) {
      config.headers = config.headers || {};
      config.headers['Authorization'] = `Bearer ${adminToken}`;
    } else if (userToken && !config.url?.includes('/admin/')) {
      config.headers = config.headers || {};
      config.headers['Authorization'] = `Bearer ${userToken}`;
    }
    // For state-changing methods, also attach the CSRF token header (SEC-007)
    if (requiresCSRF(config.method)) {
      try {
        const token = csrfCache.token || (await fetchCSRFToken());
        config.headers = config.headers || {};
        config.headers['X-CSRF-Token'] = token;
      } catch {
        // Could not fetch CSRF token — request will likely fail with 403 on the server.
        // We do not block the request here; the server-side error is surfaced to the caller.
      }
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// ---------------------------------------------------------------------------
// Response interceptor — handle auth expiry, rate limiting, CSRF expiry
// ---------------------------------------------------------------------------
apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const status = error.response?.status;
    const requestUrl = (error.config as any)?.url || '';

    if (status === 401) {
      const isAdminRoute = requestUrl.includes('/admin/');
      // Silent endpoints: 401 just means "not logged in" — never redirect for these
      const isSilentEndpoint =
        requestUrl.includes('/user/profile') ||
        requestUrl.includes('/user/me');
      // Public pages: don't redirect if the user is browsing unauthenticated
      const publicPages = ['/', '/recharge', '/draws', '/subscription', '/affiliate', '/login'];
      const isOnPublicPage = publicPages.some(
        (p) => window.location.pathname === p || window.location.pathname.startsWith(p + '/')
      );

      if (isAdminRoute) {
        localStorage.removeItem('rechargemax_admin_user');
        localStorage.removeItem('rechargemax_admin_token');
        // Don't redirect if already on login page
        if (window.location.pathname !== '/admin/login') {
          window.location.href = '/admin/login';
        }
      } else if (!isSilentEndpoint && !isOnPublicPage) {
        // Only redirect to /login if we're on a protected page
        localStorage.removeItem('rechargemax_user');
        localStorage.removeItem('rechargemax_user_token');
        window.location.href = '/login';
      } else {
        // Silent: just clear the user state without redirecting
        localStorage.removeItem('rechargemax_user');
        localStorage.removeItem('rechargemax_user_token');
      }
    }

    // CSRF token expired or invalid — clear cache and retry once (SEC-007)
    if (status === 403 && requiresCSRF(error.config?.method)) {
      const responseData = error.response?.data as any;
      if (responseData?.error?.toLowerCase().includes('csrf')) {
        csrfCache.token = null;
        try {
          const newToken = await fetchCSRFToken();
          const retryConfig = { ...error.config } as any;
          retryConfig.headers['X-CSRF-Token'] = newToken;
          return apiClient(retryConfig); // single automatic retry
        } catch {
          // Retry failed — fall through to reject
        }
      }
    }

    // 429 Too Many Requests — surface a human-readable message
    if (status === 429) {
      const enhanced = new Error('Too many requests — please wait a moment before trying again') as any;
      enhanced.status = 429;
      enhanced.originalError = error;
      return Promise.reject(enhanced);
    }

    // 413 Request Entity Too Large
    if (status === 413) {
      const enhanced = new Error('The request is too large. Please reduce the file or data size and try again') as any;
      enhanced.status = 413;
      enhanced.originalError = error;
      return Promise.reject(enhanced);
    }

    return Promise.reject(error);
  }
);

// ============================================================================
// TYPES (Re-exported from admin-api.types.ts)
// ============================================================================

// Re-export shared types from admin-api.types (not locally declared here)
export type {
  ApiResponse,
  ApiSuccessResponse,
  ApiErrorResponse,
  PaginatedResponse,
  DataPlan,
  WheelPrize,
  User,
  Transaction,
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
    
    // Store token in localStorage so it can be sent as Bearer header on
    // cross-origin requests (cookie SameSite=Lax is blocked by browsers for XHR/fetch).
    if (response.data.success && response.data.data) {
      if (response.data.data.token) {
        localStorage.setItem('rechargemax_user_token', response.data.data.token);
      }
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
      // Clear user profile and token
      localStorage.removeItem('rechargemax_user');
      localStorage.removeItem('rechargemax_user_token');
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
    
    // Store admin profile data and JWT token (cross-domain requires Bearer token)
    if (response.data.success) {
      // The response has: { success, token, admin } at top level (not nested under data)
      const adminData = response.data.data?.admin || (response.data as any).admin;
      const jwtToken = response.data.data?.token || (response.data as any).token;
      if (adminData) {
        localStorage.setItem('rechargemax_admin_user', JSON.stringify(adminData));
      }
      if (jwtToken) {
        localStorage.setItem('rechargemax_admin_token', jwtToken);
      }
    }
    
    return response.data;
  },

  // Admin logout
  logout: async () => {
    try {
      // Backend clears the httpOnly admin cookie
      await apiClient.post('/admin/auth/logout');
    } finally {
      // Clear admin session data
      localStorage.removeItem('rechargemax_admin_user');
      localStorage.removeItem('rechargemax_admin_token');
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
  // Use a generous 65 s timeout to survive a Render cold-start (free tier spins down after
  // 15 min of inactivity; first request can take 30-50 s to wake up the service).
  initiateAirtimeRecharge: async (data: {
    phone_number: string;
    network: string;
    amount: number;
  }) => {
    const response = await apiClient.post<ApiResponse>('/recharge/airtime', data, { timeout: 65000 });
    return response.data;
  },

  // Initialize data recharge (same generous timeout)
  initiateDataRecharge: async (data: {
    phone_number: string;
    network: string;
    bundle_id: string;
  }) => {
    const response = await apiClient.post<ApiResponse>('/recharge/data', data, { timeout: 65000 });
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


// ============================================================================
// ADMIN EXTENSIONS (merged from api-client-extensions.ts)
// ============================================================================

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
  daily_revenue?: number;  // Revenue for today
  total_billings: number;
  successful_billings: number;
  failed_billings: number;
  average_subscription_value: number;
  total_entries?: number;  // Total draw entries allocated
  churn_rate?: number;  // Subscription churn rate percentage
  subscription_growth?: {  // Subscription growth metrics
    new_subscriptions: number;
    cancelled_subscriptions: number;
    net_growth: number;
    growth_rate: number;
  };
  revenue_growth?: {  // Revenue growth metrics
    current_period: number;
    previous_period: number;
    growth_amount: number;
    growth_rate: number;
  };
  tier_performance?: Array<{  // Performance by tier
    tier_id: string;
    tier_name: string;
    subscriber_count: number;
    revenue: number;
  }>;
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
    const response = await apiClient.get<ApiResponse>('/admin/daily-subscriptions/analytics');
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
    // No dedicated export endpoint — fetch full list for client-side CSV generation
    const response = await apiClient.get<ApiResponse>('/admin/daily-subscriptions', { params: filters });
    return response.data;
  },

  // Export billings
  exportBillings: async (filters?: any) => {
    // No dedicated export endpoint — fetch full list for client-side CSV generation
    const response = await apiClient.get<ApiResponse>('/admin/subscription-billings', { params: filters });
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
  points_allocated?: number;  // Alternative field name
  transaction_id?: string;  // Alternative field name for transaction_reference
  transaction_reference: string;
  transaction_date: string;
  status: 'pending' | 'completed' | 'failed' | 'duplicate' | 'success';
  webhook_received_at: string;
  processed_at?: string;
  error_message?: string;
  created_at: string;
}

export interface USSDStatistics {
  total_recharges: number;
  completed_recharges: number;
  pending_recharges: number;
  failed_recharges: number;
  total_amount: number;
  total_points_earned: number;
  total_points?: number;  // Alternative field name
  average_amount: number;
  success_rate?: number;  // Success rate percentage
  recharges_by_network: Record<string, number>;
  network_breakdown?: Array<{  // Detailed network breakdown
    network: string;
    count: number;
    amount: number;
    points: number;
  }>;
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
    let url = `/admin/ussd/recharges?page=${page}&per_page=${perPage}`;
    if (network) url += `&network=${network}`;
    if (status) url += `&status=${status}`;
    const response = await apiClient.get<ApiResponse<PaginatedResponse<USSDRecharge>>>(url);
    return response.data;
  },

  // Get USSD recharge by ID
  getById: async (rechargeId: string) => {
    const response = await apiClient.get<ApiResponse<USSDRecharge>>(`/admin/ussd/recharges/${rechargeId}`);
    return response.data;
  },

  // Get webhook logs
  getWebhookLogs: async (page = 1, perPage = 50) => {
    const response = await apiClient.get<ApiResponse<PaginatedResponse<USSDWebhookLog>>>(`/admin/ussd/webhook-logs?page=${page}&per_page=${perPage}`);
    return response.data;
  },

  // Retry webhook
  retryWebhook: async (webhookId: string) => {
    const response = await apiClient.post<ApiResponse>(`/admin/ussd/retry-failed`);
    return response.data;
  },

  // Get USSD statistics
  getStats: async () => {
    const response = await apiClient.get<ApiResponse>('/admin/ussd/statistics');
    return response.data;
  },

  // Get recharges (alias for getAll)
  getRecharges: async (filters?: { network?: string; status?: string; date_from?: string; date_to?: string; page?: number; perPage?: number }) => {
    const { page = 1, perPage = 50, network, status } = filters || {};
    return ussdRechargeApi.getAll(page, perPage, network, status);
  },

  // Get statistics (alias for getStats)
  getStatistics: async (filters?: { date_from?: string; date_to?: string }) => {
    return ussdRechargeApi.getStats();
  },

  // Retry failed recharge
  retryRecharge: async (rechargeId: string) => {
    const response = await apiClient.post<ApiResponse>(`/admin/ussd/retry-failed`);
    return response.data;
  },

  // Export recharges to CSV
  exportRecharges: async (filters?: { network?: string; status?: string; date_from?: string; date_to?: string }) => {
    const response = await apiClient.get<ApiResponse>('/admin/ussd/recharges', { params: filters });
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
    // Backend route: GET /admin/draws/:id/csv/export
    const response = await apiClient.get<ApiResponse<{ file_url: string; total_msisdns: number }>>(`/admin/draws/${data.draw_id}/csv/export`, { params: data });
    return response.data;
  },

  // Import winners from CSV
  importWinners: async (drawId: string, file: File) => {
    const formData = new FormData();
    formData.append('draw_id', drawId);
    formData.append('file', file);

    const response = await apiClient.post<ApiResponse<{ total_winners: number; total_runners_up: number }>>(`/admin/draws/${drawId}/csv/import-winners`, formData, {
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

export interface ClaimApprovalRequest {
  winner_id: string;
  approved?: boolean;  // Can be derived from action
  action?: 'approve' | 'reject';  // Alternative field name
  reason?: string;
  notes?: string;  // Additional notes
}

export interface PayoutRequest {
  winner_id: string;
  payout_reference: string;
  amount?: number;  // Payout amount
  payout_amount?: number;  // Alternative field name
  payout_method?: string;  // Payment method used
  notes?: string;  // Additional payout notes
}

export interface ShippingUpdateRequest {
  winner_id: string;
  tracking_number: string;
  shipping_status: string;
  courier_service?: string;  // Courier service provider
  estimated_delivery?: string;  // Estimated delivery date
  notes?: string;  // Additional shipping notes
}

export interface Winner {
  id: string;
  draw_id: string;
  msisdn: string;
  prize_name: string;
  prize_type: 'airtime' | 'data' | 'points' | 'cash' | 'physical_goods';
  prize_value: number;
  claim_status: 'PENDING' | 'CLAIMED' | 'EXPIRED' | 'PENDING_ADMIN_REVIEW' | 'APPROVED' | 'REJECTED' | string;
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
    const response = await apiClient.get<ApiResponse>('/admin/winners/claim-statistics');
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
    const response = await apiClient.post<ApiResponse>('/admin/points/adjust', data);
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

export default apiClient;
export { apiClient };