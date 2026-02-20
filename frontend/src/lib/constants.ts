export const ROUTE_PATHS = {
  HOME: '/',
  LOGIN: '/login',
  RECHARGE: '/recharge',
  DRAWS: '/draws',
} as const;

export const NETWORKS = {
  MTN: 'mtn',
  AIRTEL: 'airtel',
  GLO: 'glo',
  ETISALAT: '9mobile',
} as const;

export const TRANSACTION_TYPES = {
  AIRTIME: 'airtime',
  DATA: 'data',
  ELECTRICITY: 'electricity',
  CABLE_TV: 'cable_tv',
} as const;

export const TRANSACTION_STATUS = {
  PENDING: 'pending',
  SUCCESSFUL: 'successful',
  FAILED: 'failed',
} as const;

export const SPIN_PRIZES = [
  { name: '₦100 Airtime', type: 'AIRTIME', amount: 100, probability: 25, color: '#10b981' },
  { name: '₦200 Airtime', type: 'AIRTIME', amount: 200, probability: 20, color: '#3b82f6' },
  { name: '₦500 Airtime', type: 'AIRTIME', amount: 500, probability: 15, color: '#8b5cf6' },
  { name: '1GB Data', type: 'DATA', amount: 1024, probability: 15, color: '#f59e0b' },
  { name: '2GB Data', type: 'DATA', amount: 2048, probability: 10, color: '#ef4444' },
  { name: '₦1,000 Cash', type: 'CASH', amount: 1000, probability: 8, color: '#ec4899' },
  { name: '₦5,000 Cash', type: 'CASH', amount: 5000, probability: 2, color: '#fbbf24' },
  { name: 'Better Luck', type: 'NONE', amount: 0, probability: 5, color: '#6b7280' },
] as const;

export const DRAW_PRIZES = {
  FIRST: { position: 1, amount: 100000, label: '₦100,000 Grand Prize' },
  SECOND: { position: 2, amount: 50000, label: '₦50,000 Second Prize' },
  THIRD: { position: 3, amount: 25000, label: '₦25,000 Third Prize' },
  CONSOLATION: { position: 4, amount: 5000, label: '₦5,000 Consolation' },
} as const;

export const USER_TIERS = {
  BRONZE: { name: 'Bronze', minSpent: 0, benefits: ['Basic rewards', 'Standard support'] },
  SILVER: { name: 'Silver', minSpent: 10000, benefits: ['Enhanced rewards', 'Priority support', '5% bonus spins'] },
  GOLD: { name: 'Gold', minSpent: 50000, benefits: ['Premium rewards', 'VIP support', '10% bonus spins', 'Exclusive draws'] },
  PLATINUM: { name: 'Platinum', minSpent: 100000, benefits: ['Elite rewards', 'Dedicated support', '15% bonus spins', 'VIP draws', 'Special offers'] },
} as const;

export const COMMISSION_RATES = {
  LEVEL_1: 0.05, // 5% for direct referrals
  LEVEL_2: 0.02, // 2% for second level
  LEVEL_3: 0.01, // 1% for third level
} as const;

export const MINIMUM_AMOUNTS = {
  AIRTIME: 50,
  DATA: 100,
  ELECTRICITY: 1000,
  CABLE_TV: 1000,
  SPIN_ELIGIBLE: 1000,
  DRAW_ELIGIBLE: 500,
} as const;

export const API_ENDPOINTS = {
  // User endpoints
  USERS: '/users',
  USER_PROFILE: '/users/profile',
  USER_STATS: '/users/stats',
  
  // Transaction endpoints
  TRANSACTIONS: '/transactions',
  RECHARGE: '/transactions/recharge',
  VERIFY_TRANSACTION: '/transactions/verify',
  
  // Spin endpoints
  SPINS: '/spins',
  SPIN_WHEEL: '/spins/wheel',
  SPIN_HISTORY: '/spins/history',
  
  // Draw endpoints
  DRAWS: '/draws',
  DRAW_ENTRIES: '/draws/entries',
  DRAW_WINNERS: '/draws/winners',
  
  // Network endpoints
  NETWORKS: '/networks',
  DATA_PLANS: '/networks/data-plans',
  
  // Affiliate endpoints
  AFFILIATES: '/affiliates',
  AFFILIATE_STATS: '/affiliates/stats',
  REFERRALS: '/affiliates/referrals',
  
  // Admin endpoints
  ADMIN: '/admin',
  ADMIN_USERS: '/admin/users',
  ADMIN_TRANSACTIONS: '/admin/transactions',
  ADMIN_STATS: '/admin/stats',
  ADMIN_PRIZES: '/admin/prizes',
} as const;

export const STORAGE_KEYS = {
  USER_TOKEN: 'rechargemax_token',
  USER_DATA: 'rechargemax_user',
  REFERRAL_CODE: 'rechargemax_referral',
  THEME: 'rechargemax_theme',
} as const;

export const VALIDATION_RULES = {
  PHONE: /^(\+234|234|0)?[789][01]\d{8}$/,
  EMAIL: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
  PASSWORD: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d@$!%*?&]{8,}$/,
  REFERRAL_CODE: /^[A-Z0-9]{6}$/,
} as const;

export const ERROR_MESSAGES = {
  NETWORK_ERROR: 'Network error. Please check your connection.',
  INVALID_PHONE: 'Please enter a valid Nigerian phone number.',
  INVALID_EMAIL: 'Please enter a valid email address.',
  WEAK_PASSWORD: 'Password must be at least 8 characters with uppercase, lowercase, and number.',
  INSUFFICIENT_BALANCE: 'Insufficient balance for this transaction.',
  TRANSACTION_FAILED: 'Transaction failed. Please try again.',
  SPIN_NOT_ELIGIBLE: 'Minimum ₦1,000 recharge required to spin.',
  DRAW_NOT_ELIGIBLE: 'Minimum ₦500 recharge required for draw entry.',
  UNAUTHORIZED: 'Please log in to continue.',
  SERVER_ERROR: 'Server error. Please try again later.',
} as const;

export const SUCCESS_MESSAGES = {
  REGISTRATION_SUCCESS: 'Account created successfully!',
  LOGIN_SUCCESS: 'Welcome back!',
  TRANSACTION_SUCCESS: 'Transaction completed successfully!',
  SPIN_SUCCESS: 'Congratulations on your spin!',
  DRAW_ENTRY_SUCCESS: 'You have been entered into the draw!',
  PROFILE_UPDATED: 'Profile updated successfully!',
  PASSWORD_CHANGED: 'Password changed successfully!',
} as const;