/**
 * Legacy Supabase client file - now exports Go backend API client
 * Maintained for backward compatibility
 */

import {
  authApi,
  adminAuthApi,
  rechargeApi,
  paymentApi,
  spinApi,
  drawApi,
  subscriptionApi,
  affiliateApi,
  userApi,
  adminApi,
  callEdgeFunction,
} from '../../lib/api-client';
import apiClient from '../../lib/api-client';

// Export API client as 'supabase' for backward compatibility
export const supabase = {
  auth: authApi,
  functions: {
    invoke: callEdgeFunction,
  },
};

// Export all API modules
export {
  apiClient,
  authApi,
  adminAuthApi,
  rechargeApi,
  paymentApi,
  spinApi,
  drawApi,
  subscriptionApi,
  affiliateApi,
  userApi,
  adminApi,
  callEdgeFunction,
};

export default apiClient;
