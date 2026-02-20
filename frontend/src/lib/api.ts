/**
 * API Functions - Updated to use Go backend
 * Replaces Supabase database calls with REST API calls
 */

import apiClient, {
  userApi,
  rechargeApi,
  spinApi,
  drawApi,
  affiliateApi,
  adminApi,
  paymentApi,
} from './api-client';

// ============================================================================
// USER MANAGEMENT
// ============================================================================

export const createUser = async (userData: any) => {
  // User creation is handled by OTP verification in Go backend
  // This function is kept for compatibility but may not be needed
  const response = await userApi.updateProfile(userData);
  return response;
};

export const getUser = async (userId: string) => {
  const response = await userApi.getProfile();
  return response;
};

export const updateUser = async (userId: string, updates: any) => {
  const response = await userApi.updateProfile(updates);
  return response;
};

export const getUserDashboard = async (msisdn: string) => {
  const response = await userApi.getDashboard();
  return response;
};

// ============================================================================
// RECHARGE OPERATIONS
// ============================================================================

export const createRecharge = async (rechargeData: any) => {
  const { phone_number, network, amount, bundle_id } = rechargeData;
  
  if (bundle_id) {
    // Data recharge
    const response = await rechargeApi.initiateDataRecharge({
      phone_number,
      network,
      bundle_id,
    });
    return response;
  } else {
    // Airtime recharge
    const response = await rechargeApi.initiateAirtimeRecharge({
      phone_number,
      network,
      amount,
    });
    return response;
  }
};

export const getUserRecharges = async (userId: string) => {
  const response = await userApi.getTransactions();
  // Filter for recharge transactions
  return response?.filter((t: any) => 
    t.transaction_type === 'airtime' || t.transaction_type === 'data'
  ) || [];
};

// ============================================================================
// SPIN OPERATIONS
// ============================================================================

export const createSpin = async (spinData: any) => {
  const response = await spinApi.spin();
  return response;
};

export const getUserSpins = async (userId: string) => {
  const response = await spinApi.getHistory();
  return response || [];
};

// ============================================================================
// DRAW OPERATIONS
// ============================================================================

export const createDrawEntry = async (entryData: any) => {
  // Draw entries are created automatically by the backend when user subscribes
  // This function is kept for compatibility
  return entryData;
};

export const getUserDrawEntries = async (userId: string) => {
  const response = await drawApi.getMyEntries();
  return response || [];
};

// ============================================================================
// NETWORK OPERATIONS
// ============================================================================

export const getNetworks = async () => {
  const response = await rechargeApi.getNetworks();
  return response || [];
};

export const getDataPlans = async (networkId: string) => {
  const response = await rechargeApi.getDataBundles(networkId);
  return response || [];
};

export const validatePhoneNetwork = async (phoneNumber: string, expectedNetwork: string) => {
  const response = await rechargeApi.validatePhoneNetwork(phoneNumber, expectedNetwork);
  return response;
};

// ============================================================================
// TRANSACTION OPERATIONS
// ============================================================================

export const createTransaction = async (transactionData: any) => {
  // Transactions are created automatically by the backend
  // This function is kept for compatibility
  return transactionData;
};

export const getUserTransactions = async (userId: string) => {
  const response = await userApi.getTransactions();
  return response || [];
};

// ============================================================================
// AFFILIATE OPERATIONS
// ============================================================================

export const createAffiliate = async (affiliateData: any) => {
  const response = await affiliateApi.register(affiliateData);
  return response;
};

export const getAffiliateStats = async (userId: string) => {
  const response = await affiliateApi.getDashboard();
  return response;
};

// ============================================================================
// ADMIN OPERATIONS
// ============================================================================

export const getAllUsers = async () => {
  const response = await adminApi.users.getAll();
  return response?.data || [];
};

export const getAllTransactions = async () => {
  // Admin analytics endpoint
  const response = await adminApi.analytics.getOverview();
  return response?.transactions || [];
};

export const getSystemStats = async () => {
  const response = await adminApi.getStats();
  return response;
};

// ============================================================================
// PRIZE OPERATIONS
// ============================================================================

export const getWheelPrizes = async () => {
  const response = await adminApi.spin.getPrizes();
  return response || [];
};

export const createWheelPrize = async (prizeData: any) => {
  const response = await adminApi.spin.createPrize(prizeData);
  return response;
};

export const updateWheelPrize = async (prizeId: string, updates: any) => {
  const response = await adminApi.spin.updatePrize(prizeId, updates);
  return response;
};

export const deleteWheelPrize = async (prizeId: string) => {
  const response = await adminApi.spin.deletePrize(prizeId);
  return response;
};

// ============================================================================
// PAYMENT OPERATIONS
// ============================================================================

export const initializePayment = async (paymentData: any) => {
  const response = await paymentApi.initializePayment(paymentData);
  return response;
};

export const verifyPayment = async (reference: string) => {
  const response = await paymentApi.verifyPayment(reference);
  return response;
};

// ============================================================================
// PRIZE CLAIM OPERATIONS
// ============================================================================

export const claimPrize = async (prizeId: string, claimData: any) => {
  const response = await userApi.claimPrize(prizeId, claimData);
  return response;
};

// ============================================================================
// AFFILIATE OPERATIONS
// ============================================================================

export const getAffiliateDashboard = async (msisdn: string) => {
  const response = await affiliateApi.getDashboard();
  return response;
};

export const registerAffiliate = async (affiliateData: any) => {
  const response = await affiliateApi.register(affiliateData);
  return response;
};

export const refreshAffiliateLink = async () => {
  const response = await affiliateApi.getReferralLink();
  return response;
};

export const trackAffiliateClick = async () => {
  const response = await affiliateApi.trackClick();
  return response;
};

// ============================================================================
// LOGGING OPERATIONS
// ============================================================================

export const logError = async (errorData: any) => {
  try {
    // Log to console in development
    console.error('Error logged:', errorData);
    // In production, send to logging service
    // await apiClient.post('/logs/error', errorData);
  } catch (err) {
    console.error('Failed to log error:', err);
  }
};

export const logPerformance = async (performanceData: any) => {
  try {
    // Log to console in development
    console.log('Performance logged:', performanceData);
    // In production, send to logging service
    // await apiClient.post('/logs/performance', performanceData);
  } catch (err) {
    console.error('Failed to log performance:', err);
  }
};

// ============================================================================
// PLATFORM STATISTICS
// ============================================================================

export const getPlatformStatistics = async () => {
  const response = await apiClient.get('/platform/statistics');
  return response;
};

export const getRecentWinners = async (limit: number = 4) => {
  const response = await apiClient.get(`/winners/recent?limit=${limit}`);
  return response;
};

// ============================================================================
// SPIN MANAGEMENT
// ============================================================================

export const getAvailableSpins = async (msisdn: string) => {
  try {
    const response = await apiClient.get(`/spins/available?msisdn=${msisdn}`);
    return response;
  } catch (err) {
    console.error('Failed to get available spins:', err);
    return { success: false, data: { availableSpins: 0 } };
  }
};

export const consumeSpin = async (msisdn: string, transactionReference?: string) => {
  try {
    const response = await apiClient.post('/spins/consume', {
      msisdn,
      transactionReference
    });
    return response;
  } catch (err) {
    console.error('Failed to consume spin:', err);
    return { success: false, error: 'Failed to consume spin' };
  }
};

export const recordTransactionPrize = async (data: {
  transactionReference: string;
  msisdn: string;
  prizeType: string;
  prizeValue: number;
  prizeDescription: string;
}) => {
  try {
    const response = await apiClient.post('/prizes/record', data);
    return response;
  } catch (err) {
    console.error('Failed to record transaction prize:', err);
    return { success: false, error: 'Failed to record prize' };
  }
};

export const getTierProgress = async (msisdn: string) => {
  try {
    const response = await apiClient.get(`/spins/tier-progress?msisdn=${msisdn}`);
    return response;
  } catch (err) {
    console.error('Failed to get tier progress:', err);
    return { success: false, data: { currentTier: 0, progress: 0, cumulativeAmount: 0 } };
  }
};

export const getSpinTiers = async () => {
  try {
    const response = await apiClient.get('/spins/tiers');
    return response;
  } catch (err) {
    console.error('Failed to get spin tiers:', err);
    return { success: false, data: [] };
  }
};

// ============================================================================
// EDGE FUNCTIONS (Legacy Supabase compatibility)
// ============================================================================

export const callEdgeFunction = async (functionName: string, params: any) => {
  // Legacy function for Supabase edge functions
  // Now routes to appropriate Go backend API endpoints
  console.warn(`callEdgeFunction('${functionName}') is deprecated. Use specific API methods instead.`);
  
  // Map legacy function names to new API endpoints
  const functionMap: Record<string, () => Promise<any>> = {
    'clean_working_admin_api_2025_11_12_17_00': () => adminApi.getStats(),
    'updated_admin_subscription_api_2026_01_08_19_02': () => adminApi.getSubscriptions(),
    'sync_frontend_settings_2025_11_10_14_30': () => adminApi.updateSettings(params),
    'strategic_affiliate_admin_api_2025_11_12_18_30': () => adminApi.getAffiliateStats(),
  };
  
  const handler = functionMap[functionName];
  if (handler) {
    return await handler();
  }
  
  throw new Error(`Edge function '${functionName}' not implemented in Go backend`);
};

// ============================================================================
// DRAW OPERATIONS (Additional)
// ============================================================================

export const getActiveDraws = async () => {
  const response = await drawApi.getActiveDraws();
  return response;
};

export const getDrawResults = async (drawId: string) => {
  const response = await drawApi.getDrawResults(drawId);
  return response;
};
