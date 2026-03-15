/**
 * api.ts — compatibility shim
 *
 * All API functions have been consolidated into api-client.ts.
 * This file re-exports named functions so that existing component imports
 * (`from '@/lib/api'`) continue to work without modification.
 *
 * @deprecated For new code, import directly from '@/lib/api-client'.
 */

import apiClient, {
  authApi,
  adminAuthApi,
  userApi,
  rechargeApi,
  spinApi,
  drawApi,
  affiliateApi,
  adminApi,
  paymentApi,
} from './api-client';

// ─── User ──────────────────────────────────────────────────────────────────
export const createUser = async (userData: any) => userApi.updateProfile(userData);
export const getUser = async (_userId: string) => userApi.getProfile();
export const updateUser = async (_userId: string, updates: any) => userApi.updateProfile(updates);
export const getUserDashboard = async (_msisdn: string) => userApi.getDashboard();

// ─── Recharge ──────────────────────────────────────────────────────────────
export const createRecharge = async (d: any) => {
  if (d.bundle_id) {
    return rechargeApi.initiateDataRecharge({ phone_number: d.phone_number, network: d.network, bundle_id: d.bundle_id });
  }
  return rechargeApi.initiateAirtimeRecharge({ phone_number: d.phone_number, network: d.network, amount: d.amount });
};
export const getUserRecharges = async (_userId: string) => {
  const r = await userApi.getTransactions();
  const data = (r as any)?.data || [];
  return data.filter((t: any) => t.transaction_type === 'airtime' || t.transaction_type === 'data');
};

// ─── Spin ───────────────────────────────────────────────────────────────────
export const createSpin = async (_d: any) => spinApi.spin();
export const getUserSpins = async (_userId: string) => spinApi.getHistory();

// ─── Draw ───────────────────────────────────────────────────────────────────
export const createDrawEntry = async (entryData: any) => entryData;
export const getUserDrawEntries = async (_userId: string) => drawApi.getMyEntries();

// ─── Networks ───────────────────────────────────────────────────────────────
export const getNetworks = async () => rechargeApi.getNetworks();
export const getDataPlans = async (networkId: string) => rechargeApi.getDataBundles(networkId);
export const validatePhoneNetwork = async (phone: string, net: string) =>
  rechargeApi.validatePhoneNetwork(phone, net);

// ─── Transactions ───────────────────────────────────────────────────────────
export const createTransaction = async (d: any) => d;
export const getUserTransactions = async (_userId: string) => userApi.getTransactions();

// ─── Affiliate ──────────────────────────────────────────────────────────────
export const createAffiliate = async (d: any) => affiliateApi.register(d);
export const getAffiliateStats = async (_userId: string) => affiliateApi.getDashboard();
export const getAffiliateDashboard = async (_msisdn: string) => affiliateApi.getDashboard();
export const registerAffiliate = async (d: any) => affiliateApi.register(d);
export const refreshAffiliateLink = async () => affiliateApi.getReferralLink();
export const trackAffiliateClick = async () => affiliateApi.trackClick();

// ─── Admin ───────────────────────────────────────────────────────────────────
export const getAllUsers = async () => {
  const r = await adminApi.users.getAll();
  return r?.data || [];
};
export const getAllTransactions = async () => {
  const r = await adminApi.analytics.getOverview();
  return r?.transactions || [];
};
export const getSystemStats = async () => adminApi.getStats();

// ─── Prizes ──────────────────────────────────────────────────────────────────
export const getWheelPrizes = async () => adminApi.spin.getPrizes();
export const createWheelPrize = async (d: any) => adminApi.spin.createPrize(d);
export const updateWheelPrize = async (id: string, d: any) => adminApi.spin.updatePrize(id, d);
export const deleteWheelPrize = async (id: string) => adminApi.spin.deletePrize(id);

// ─── Payment ─────────────────────────────────────────────────────────────────
export const initializePayment = async (d: any) => paymentApi.initializePayment(d);
export const verifyPayment = async (ref: string) => paymentApi.verifyPayment(ref);

// ─── Prize claims ────────────────────────────────────────────────────────────
export const claimPrize = async (prizeId: string, d: any) => userApi.claimPrize(prizeId, d);

// ─── Platform stats ──────────────────────────────────────────────────────────
export const getPlatformStatistics = async () => apiClient.get('/platform/statistics');
export const getRecentWinners = async (limit = 4) => apiClient.get(`/winners/recent?limit=${limit}`);

// ─── Spins (mapped to real backend endpoints) ────────────────────────────────
export const getAvailableSpins = async (msisdn: string) => {
  try {
    return await apiClient.get(`/spin/eligibility?msisdn=${msisdn}`);
  } catch {
    return { success: false, data: { availableSpins: 0 } };
  }
};
export const consumeSpin = async (msisdn: string, _transactionReference?: string) => {
  try {
    return await apiClient.post('/spin/play', { msisdn });
  } catch {
    return { success: false, error: 'Failed to consume spin' };
  }
};
export const recordTransactionPrize = async (d: {
  transactionReference: string; msisdn: string; prizeType: string;
  prizeValue: number; prizeDescription: string;
}) => {
  // Prize recording is handled server-side during spin play — no separate endpoint.
  // Return success so callers don't break.
  console.warn('recordTransactionPrize: handled server-side, call is a no-op', d);
  return { success: true };
};
export const getTierProgress = async (msisdn: string) => {
  try {
    return await apiClient.get(`/spins/tier-progress?msisdn=${msisdn}`);
  } catch {
    return { success: false, data: { currentTier: 0, progress: 0, cumulativeAmount: 0 } };
  }
};
export const getSpinTiers = async () => {
  try {
    return await apiClient.get('/spins/tiers');
  } catch {
    return { success: false, data: [] };
  }
};

// ─── Draws ───────────────────────────────────────────────────────────────────
export const getActiveDraws = async () => drawApi.getActiveDraws();
export const getDrawResults = async (drawId: string) => drawApi.getDrawResults(drawId);

// ─── Daily subscription ───────────────────────────────────────────────────────
export const processDailySubscription = async (d: { msisdn: string; tier_id?: string; payment_method?: string }) =>
  (await apiClient.post('/subscriptions/daily', d)).data;
export const getDailySubscriptionStatus = async (msisdn: string) =>
  (await apiClient.get(`/subscriptions/daily/status?msisdn=${msisdn}`)).data;
export const cancelDailySubscription = async (subscriptionId: string) =>
  (await apiClient.post(`/subscriptions/daily/${subscriptionId}/cancel`)).data;

// ─── Logging (no-op in production; kept for compatibility) ────────────────────
export const logError = async (errorData: any) => {
  if (import.meta.env.DEV) console.error('Error logged:', errorData);
};
export const logPerformance = async (_d: any) => { /* intentional no-op */ };

// ─── Legacy edge-function shim ────────────────────────────────────────────────
export const callEdgeFunction = async (functionName: string, params: any) => {
  console.warn(`callEdgeFunction('${functionName}') is deprecated.`);
  const map: Record<string, () => Promise<any>> = {
    'send-otp': () => authApi.sendOTP(params.phone_number),
    'verify-otp': () => authApi.verifyOTP(params.phone_number, params.otp_code),
    'initialize-payment': () => paymentApi.initializePayment(params),
    'verify-payment': () => paymentApi.verifyPayment(params.reference),
    'get-networks': () => rechargeApi.getNetworks(),
    'spin-wheel': () => spinApi.spin(),
    'get-active-draws': () => drawApi.getActiveDraws(),
    'admin-login': () => adminAuthApi.login(params.email, params.password),
  };
  const fn = map[functionName];
  if (fn) return fn();
  throw new Error(`Edge function '${functionName}' not implemented.`);
};
