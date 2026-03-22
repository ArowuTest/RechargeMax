import React, { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';

import { getUserDashboard, claimPrize } from '@/lib/api';
import { apiClient } from '@/lib/api-client';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useAuthContext } from '@/contexts/AuthContext';

const NIGERIAN_BANKS = [
  { name: 'Access Bank', code: '044' },
  { name: 'Citibank Nigeria', code: '023' },
  { name: 'Ecobank Nigeria', code: '050' },
  { name: 'Fidelity Bank', code: '070' },
  { name: 'First Bank of Nigeria', code: '011' },
  { name: 'First City Monument Bank (FCMB)', code: '214' },
  { name: 'Globus Bank', code: '00103' },
  { name: 'Guaranty Trust Bank (GTBank)', code: '058' },
  { name: 'Heritage Bank', code: '030' },
  { name: 'Jaiz Bank', code: '301' },
  { name: 'Keystone Bank', code: '082' },
  { name: 'Kuda Microfinance Bank', code: '50211' },
  { name: 'Lotus Bank', code: '303' },
  { name: 'OPay Digital Services', code: '999992' },
  { name: 'Palmpay', code: '999991' },
  { name: 'Parallex Bank', code: '526' },
  { name: 'Polaris Bank', code: '076' },
  { name: 'Providus Bank', code: '101' },
  { name: 'Stanbic IBTC Bank', code: '221' },
  { name: 'Standard Chartered Bank', code: '068' },
  { name: 'Sterling Bank', code: '232' },
  { name: 'Titan Trust Bank', code: '102' },
  { name: 'Union Bank of Nigeria', code: '032' },
  { name: 'United Bank for Africa (UBA)', code: '033' },
  { name: 'Unity Bank', code: '215' },
  { name: 'Wema Bank', code: '035' },
  { name: 'Zenith Bank', code: '057' },
];
import { formatCurrency, formatDate, getNetworkColor } from '@/lib/utils';
import { useToast } from '@/hooks/useToast';
import { useNavigate } from 'react-router-dom';
import { SpinWheel } from '@/components/games/SpinWheel';
import { SpinUpgradeNudge } from '@/components/games/SpinUpgradeNudge';
import {
  CreditCard, Gift, TrendingUp, Calendar, Smartphone, Trophy,
  User, Loader2, CheckCircle, Clock, AlertCircle, ArrowLeft,
  DollarSign, Phone, Download, Search, Copy, RefreshCw, Award,
  Zap, Star, ChevronRight, Sparkles, Wallet, Layers, Plus, XCircle
} from 'lucide-react';

interface DashboardData {
  user: {
    id: string;
    msisdn: string;
    first_name?: string;
    last_name?: string;
    email: string;
    loyalty_tier: string;
    total_points: number;
    referral_code: string;
  };
  stats: {
    total_recharges: number;
    total_spins: number;
    total_wins: number;
  };
  summary: {
    total_transactions: number;
    total_prizes: number;
    pending_prizes: number;
    claimed_prizes: number;
    total_amount_recharged: number;
    total_subscriptions: number;
    total_subscription_amount: number;
    total_subscription_entries: number;
    total_subscription_points: number;
  };
  recent_transactions: Array<{
    id: string;
    created_at: string;
    network_provider: string;
    recharge_type: string;
    amount: number;
    points_earned: number;
    status: string;
  }>;
  subscriptions: Array<{
    id: string;
    subscription_code: string;
    transaction_date: string;
    amount: number;
    entries: number;
    points_earned: number;
    status: string;
  }>;
  prizes: Array<{
    id: string;
    prize_name: string;
    prize_value: number;
    prize_type: string;
    won_date: string;
    claim_date?: string;
    status: string;
    fulfillment_mode?: string;
    fulfillment_attempts?: number;
    fulfillment_error?: string;
    claim_reference?: string;
  }>;
}

interface BankDetails {
  account_number: string;
  account_name: string;
  bank_name: string;
  bank_code: string;
}

export const UserDashboard: React.FC = () => {
  const { user, isAuthenticated, logout } = useAuthContext();
  const navigate = useNavigate();
  const { toast } = useToast();
  const [loading, setLoading] = useState(true);
  const [dashboardData, setDashboardData] = useState<DashboardData | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState('overview');
  const [searchTerm, setSearchTerm] = useState('');
  const [claimingPrize, setClaimingPrize] = useState<string | null>(null);
  const [activeLines, setActiveLines] = useState<{
    lines: Array<{ id: string; code: string; entries: number; daily_amount_ngn: number; status: string; next_billing: string }>;
    total_active_lines: number;
    total_daily_entries: number;
    total_daily_cost_ngn: number;
  } | null>(null);
  const [cancellingLine, setCancellingLine] = useState<string | null>(null);
    const [bankDetails, setBankDetails] = useState<BankDetails>({
    account_number: '',
    account_name: '',
    bank_name: '',
    bank_code: '',
  });
  const [showBankForm, setShowBankForm] = useState<string | null>(null);
  const [editingEmail, setEditingEmail] = useState(false);
  const [newEmail, setNewEmail] = useState('');
  const [updatingEmail, setUpdatingEmail] = useState(false);
  const [showSpinWheel, setShowSpinWheel] = useState(false);
  const [availableSpins, setAvailableSpins] = useState(0);
  const [checkingSpins, setCheckingSpins] = useState(false);
  // Upgrade nudge state — shown when spins are exhausted instead of the wheel
  const [showUpgradeNudge, setShowUpgradeNudge] = useState(false);
  const [nudgeData, setNudgeData] = useState<{
    spinsGranted: number;
    spinsUsed: number;
    nextTierName?: string;
    nextTierMinAmount?: number;
    amountToNextTier?: number;
    nextTierSpins?: number;
  } | null>(null);

  const fetchDashboardData = useCallback(async () => {
    if (!user?.msisdn) {
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError(null);
      
      const response = await getUserDashboard(user.msisdn);

      if (response.success && response.data) {
        setDashboardData(response.data);
      } else {
        setError(!response.success ? response.error : 'Failed to load dashboard');
      }

      // Fetch active subscription lines (multi-line support)
      try {
        const linesRes = await apiClient.get<{ success: boolean; data: typeof activeLines }>(
          `/subscription/active-lines?msisdn=${encodeURIComponent(user.msisdn)}`
        );
        if (linesRes.data?.success) setActiveLines(linesRes.data.data);
      } catch { /* non-critical */ }

    } catch (err: any) {
      console.error('Dashboard fetch error:', err);
      setError(err.message || 'An error occurred');
    } finally {
      setLoading(false);
    }
  }, [user?.msisdn]);

  useEffect(() => {
    if (isAuthenticated && user) {
      fetchDashboardData();
    } else {
      setLoading(false);
    }
  }, [isAuthenticated, user, fetchDashboardData]);

  // Check for pending spins after dashboard data is loaded.
  // We use sessionStorage to ensure the auto-popup fires at most ONCE per browser
  // session for a given MSISDN.  The user can always open it manually via the
  // "Prizes" tab or a dashboard button, but we won't interrupt them repeatedly.
  useEffect(() => {
    if (dashboardData && user?.msisdn && !checkingSpins) {
      checkPendingSpins();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dashboardData?.user?.id]); // Only run when dashboard data first loads

  const checkPendingSpins = async () => {
    if (!user?.msisdn || checkingSpins) return;

    try {
      setCheckingSpins(true);
      const response = await apiClient.get('/spin/eligibility');
      const data = response.data;

      if (data.success && data.data.eligible && data.data.available_spins > 0) {
        // Spins available — update count and show banner (NOT auto-popup)
        setAvailableSpins(data.data.available_spins);

      } else if (
        data.success &&
        !data.data.eligible &&
        (data.data.spins_granted_today > 0 || data.data.spins_used_today > 0)
      ) {
        // Spins exhausted — store nudge data ready if user opens it
        setAvailableSpins(0);
        setNudgeData({
          spinsGranted:      data.data.spins_granted_today ?? 0,
          spinsUsed:         data.data.spins_used_today    ?? 0,
          nextTierName:      data.data.next_tier_name,
          nextTierMinAmount: data.data.next_tier_min_amount,
          amountToNextTier:  data.data.amount_to_next_tier,
          nextTierSpins:     data.data.next_tier_spins,
        });
      }
    } catch (error) {
      console.error('Failed to check pending spins:', error);
    } finally {
      setCheckingSpins(false);
    }
  };

  const handleClaimPrize = async (prizeId: string, prizeType: string) => {
    if (!user?.msisdn) return;

    try {
      setClaimingPrize(prizeId);

      let claimData: any = {
        prize_id: prizeId,
        msisdn: user.msisdn
      };

      // For cash prizes, require bank details
      if (prizeType === 'CASH') {
        if (!bankDetails.account_number || !bankDetails.account_name || !bankDetails.bank_name) {
          setShowBankForm(prizeId);
          setClaimingPrize(null);
          return;
        }
        // Send flat body — backend expects account_number, account_name, bank_name at top level
        claimData.account_number = bankDetails.account_number;
        claimData.account_name   = bankDetails.account_name;
        claimData.bank_name      = bankDetails.bank_name;
        claimData.bank_code      = bankDetails.bank_code;
      }

      const result = await claimPrize(prizeId, claimData);
      if (result.success) {
        toast({
          title: "Claim Submitted! ✅",
          description: prizeType === 'CASH'
            ? 'Your bank details have been submitted. Admin will process your payment within 24–48 hours.'
            : prizeType === 'AIRTIME' || prizeType === 'DATA'
            ? 'Your airtime/data claim has been submitted. It will be credited to your phone — check back in a few minutes. If not received within 24 hours, our team will process it manually.'
            : 'Your claim has been submitted successfully!',
        });
        
        // Reset bank form
        setBankDetails({ account_number: '', account_name: '', bank_name: '', bank_code: '' });
        setShowBankForm(null);
        
        // Refresh dashboard data
        fetchDashboardData();
      } else {
        throw new Error(result.error);
      }
    } catch (error) {
      console.error('Failed to claim prize:', error);
      toast({
        title: "Claim Failed",
        description: error instanceof Error ? error.message : "Failed to claim prize",
        variant: "destructive"
      });
    } finally {
      setClaimingPrize(null);
    }
  };

  const copyReferralCode = () => {
    if (dashboardData?.user?.referral_code) {
      navigator.clipboard.writeText(dashboardData.user.referral_code);
      toast({
        title: 'Copied!',
        description: 'Referral code copied to clipboard',
      });
    }
  };

  const handleUpdateEmail = async () => {
    if (!newEmail || !user?.msisdn) return;

    // Basic email validation
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(newEmail)) {
      toast({
        title: 'Invalid Email',
        description: 'Please enter a valid email address',
        variant: 'destructive',
      });
      return;
    }

    try {
      setUpdatingEmail(true);
      const response = await apiClient.put('/user/profile', { email: newEmail });

      const data = response.data;

      if (data.success) {
        toast({
          title: 'Email Updated!',
          description: 'Your email has been successfully updated',
        });
        setEditingEmail(false);
        setNewEmail('');
        fetchDashboardData();
      } else {
        throw new Error(data.error || 'Failed to update email');
      }
    } catch (error) {
      console.error('Failed to update email:', error);
      toast({
        title: 'Update Failed',
        description: error instanceof Error ? error.message : 'Failed to update email',
        variant: 'destructive',
      });
    } finally {
      setUpdatingEmail(false);
    }
  };
  if (!isAuthenticated) {
    return (
      <div className="container mx-auto p-6">
        <Card>
          <CardContent className="p-6">
            <p>Please log in to view your dashboard.</p>
            <Button onClick={() => navigate('/login')} className="mt-4">
              Go to Login
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="container mx-auto p-6 flex items-center justify-center min-h-screen">
        <div className="text-center">
          <Loader2 className="w-12 h-12 animate-spin mx-auto mb-4" />
          <p>Loading dashboard...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto p-6">
        <Card>
          <CardContent className="p-6">
            <p className="text-red-600">Error: {error}</p>
            <Button onClick={fetchDashboardData} className="mt-4">
              Retry
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!dashboardData) {
    return (
      <div className="container mx-auto p-6">
        <Card>
          <CardContent className="p-6">
            <p>No dashboard data available.</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  const fullName = `${dashboardData.user.first_name || ''} ${dashboardData.user.last_name || ''}`.trim() || 'User';

  // Filter transactions based on search
  const filteredTransactions = dashboardData.recent_transactions?.filter(tx =>
    tx.network_provider?.toLowerCase().includes(searchTerm.toLowerCase()) ||
    tx.status?.toLowerCase().includes(searchTerm.toLowerCase())
  ) || [];

  const TIER_STYLES: Record<string, { badge: string; glow: string; label: string }> = {
    BRONZE:   { badge: 'tier-bronze',   glow: 'rgba(205,127,50,0.3)',  label: '🥉 Bronze'   },
    SILVER:   { badge: 'tier-silver',   glow: 'rgba(168,169,173,0.3)', label: '🥈 Silver'   },
    GOLD:     { badge: 'tier-gold',     glow: 'rgba(255,215,0,0.4)',   label: '🥇 Gold'     },
    PLATINUM: { badge: 'tier-platinum', glow: 'rgba(229,228,226,0.3)', label: '💎 Platinum' },
  };

  const tierKey = (dashboardData?.user?.loyalty_tier || 'BRONZE').toUpperCase();
  const tierStyle = TIER_STYLES[tierKey] ?? TIER_STYLES["BRONZE"]!;

  const greetingHour = new Date().getHours();
  const greeting = greetingHour < 12 ? 'Good morning' : greetingHour < 17 ? 'Good afternoon' : 'Good evening';

  const TABS = ['overview', 'transactions', 'subscriptions', 'prizes', 'profile'];

  /* stagger animation helper */
  const fadeUp = (delay = 0) => ({
    initial: { opacity: 0, y: 16 },
    animate: { opacity: 1, y: 0 },
    transition: { duration: 0.4, delay, ease: [0.16, 1, 0.3, 1] },
  });

  return (
    <div className="min-h-screen" style={{ background: 'linear-gradient(160deg, #f5f3ff 0%, #faf5ff 40%, #fff7ed 100%)' }}>
      <div className="max-w-screen-xl mx-auto px-4 py-6 space-y-6">

        {/* ── Hero greeting bar ─────────────────────────────────────────── */}
        <motion.div
          className="relative overflow-hidden rounded-3xl p-6 text-white shadow-xl"
          style={{ background: 'linear-gradient(135deg, #1a0b3b 0%, #3b0764 50%, #7c3aed 100%)' }}
          {...fadeUp(0)}
        >
          {/* bg decoration */}
          <div className="absolute top-0 right-0 w-64 h-64 rounded-full opacity-10"
               style={{ background: 'radial-gradient(circle, #f59e0b, transparent)', transform: 'translate(30%, -30%)' }} />
          <div className="absolute bottom-0 left-0 w-48 h-48 rounded-full opacity-10"
               style={{ background: 'radial-gradient(circle, #a855f7, transparent)', transform: 'translate(-30%, 30%)' }} />

          <div className="relative flex items-start justify-between flex-wrap gap-4">
            <div className="space-y-1">
              <p className="text-purple-300 text-sm font-medium">{greeting} 👋</p>
              <h1 className="text-2xl font-extrabold" style={{ fontFamily: "'Plus Jakarta Sans', sans-serif" }}>
                {fullName}
              </h1>
              <p className="text-purple-200 text-sm">{dashboardData.user.msisdn}</p>
            </div>
            <div className="flex items-center gap-3">
              <motion.span
                className={`text-xs font-bold px-3 py-1.5 rounded-full shadow-md ${tierStyle.badge}`}
                initial={{ scale: 0.7, opacity: 0 }}
                animate={{ scale: 1, opacity: 1 }}
                transition={{ delay: 0.2, type: 'spring', stiffness: 300 }}
              >
                {tierStyle.label}
              </motion.span>
              <motion.button
                onClick={fetchDashboardData}
                className="w-9 h-9 rounded-xl bg-white/10 hover:bg-white/20 flex items-center justify-center text-white/80 hover:text-white transition-colors"
                whileHover={{ scale: 1.1, rotate: 180 }}
                whileTap={{ scale: 0.9 }}
                transition={{ duration: 0.3 }}
              >
                <RefreshCw className="w-4 h-4" />
              </motion.button>
            </div>
          </div>

          {/* points + progress */}
          <div className="relative mt-5 pt-4 border-t border-white/10">
            <div className="flex items-center justify-between mb-2">
              <span className="text-purple-300 text-xs font-semibold uppercase tracking-wider flex items-center gap-1">
                <Star className="w-3 h-3" /> Loyalty Points
              </span>
              <span className="text-yellow-300 font-black text-lg">{(dashboardData.user.total_points || 0).toLocaleString()} pts</span>
            </div>
            <div className="w-full h-2 rounded-full bg-white/10 overflow-hidden">
              <motion.div
                className="h-full rounded-full gradient-gold"
                initial={{ width: 0 }}
                animate={{ width: `${Math.min(100, ((dashboardData.user.total_points || 0) % 500) / 5)}%` }}
                transition={{ duration: 1, delay: 0.4, ease: 'easeOut' }}
              />
            </div>
            <p className="text-purple-400 text-xs mt-1">
              {500 - ((dashboardData.user.total_points || 0) % 500)} pts to next tier milestone
            </p>
          </div>
        </motion.div>

        {/* ── Stat cards ───────────────────────────────────────────────── */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {[
            {
              label: 'Total Points',
              value: (dashboardData.user.total_points || 0).toLocaleString(),
              sub: tierStyle.label,
              icon: Award,
              gradient: 'from-violet-500 to-purple-600',
              glow: 'rgba(124,58,237,0.25)',
              delay: 0.05,
            },
            {
              label: 'Total Recharges',
              value: (dashboardData.stats?.total_recharges || 0).toString(),
              sub: formatCurrency(dashboardData.summary?.total_amount_recharged || 0) + ' total',
              icon: Smartphone,
              gradient: 'from-blue-500 to-cyan-500',
              glow: 'rgba(59,130,246,0.25)',
              delay: 0.1,
            },
            {
              label: 'Prizes Won',
              value: (dashboardData.summary?.total_prizes || 0).toString(),
              sub: `${dashboardData.summary?.pending_prizes || 0} pending`,
              icon: Trophy,
              gradient: 'from-amber-400 to-orange-500',
              glow: 'rgba(245,158,11,0.25)',
              delay: 0.15,
            },
            {
              label: 'Subscriptions',
              value: (dashboardData.summary?.total_subscriptions || 0).toString(),
              sub: `${dashboardData.summary?.total_subscription_entries || 0} entries`,
              icon: Calendar,
              gradient: 'from-emerald-400 to-teal-500',
              glow: 'rgba(16,185,129,0.25)',
              delay: 0.2,
            },
          ].map(({ label, value, sub, icon: Icon, gradient, glow, delay }) => (
            <motion.div
              key={label}
              className="stat-card relative overflow-hidden rounded-2xl bg-white p-4 shadow-sm border border-gray-100"
              style={{ boxShadow: `0 4px 20px ${glow}` }}
              {...fadeUp(delay)}
            >
              <div className="flex items-start justify-between">
                <div className="space-y-1">
                  <p className="text-xs text-gray-500 font-semibold uppercase tracking-wide">{label}</p>
                  <motion.p
                    className="text-3xl font-black text-gray-900"
                    initial={{ opacity: 0, y: 8 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: delay + 0.15, duration: 0.4 }}
                    style={{ fontFamily: "'Plus Jakarta Sans', sans-serif" }}
                  >
                    {value}
                  </motion.p>
                  <p className="text-xs text-gray-400">{sub}</p>
                </div>
                <div className={`w-10 h-10 rounded-xl bg-gradient-to-br ${gradient} flex items-center justify-center shadow-md flex-shrink-0`}>
                  <Icon className="w-5 h-5 text-white" />
                </div>
              </div>
              {/* colour accent bar */}
              <div className={`absolute bottom-0 left-0 right-0 h-0.5 bg-gradient-to-r ${gradient} opacity-40`} />
            </motion.div>
          ))}
        </div>

        {/* ── Spin banner ──────────────────────────────────────────────── */}
        <AnimatePresence>
          {availableSpins > 0 && (
            <motion.div
              className="relative overflow-hidden rounded-2xl p-[2px] shadow-lg"
              style={{ background: 'linear-gradient(90deg, #7c3aed, #f59e0b, #7c3aed)', backgroundSize: '200% 100%' }}
              initial={{ opacity: 0, scale: 0.97 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 0.97 }}
              transition={{ duration: 0.3 }}
            >
              <div className="rounded-[14px] px-5 py-4 flex items-center justify-between gap-4"
                   style={{ background: 'linear-gradient(135deg, #faf5ff, #fffbeb)' }}>
                <div className="flex items-center gap-3">
                  <div className="relative">
                    <div className="w-12 h-12 rounded-xl gradient-brand flex items-center justify-center">
                      <Zap className="w-6 h-6 text-white" />
                    </div>
                    <span className="absolute -top-1 -right-1 w-5 h-5 rounded-full bg-yellow-400 text-[10px] font-black text-gray-900 flex items-center justify-center">
                      {availableSpins}
                    </span>
                  </div>
                  <div>
                    <p className="font-bold text-gray-900 text-sm sm:text-base">
                      {availableSpins} free spin{availableSpins > 1 ? 's' : ''} waiting!
                    </p>
                    <p className="text-xs text-gray-500 mt-0.5">
                      Earned from today's recharge — spin to win airtime, data or cash
                    </p>
                  </div>
                </div>
                <motion.button
                  onClick={() => setShowSpinWheel(true)}
                  className="flex-shrink-0 btn-claim px-5 py-2.5 rounded-xl text-sm font-bold"
                  whileHover={{ scale: 1.04 }}
                  whileTap={{ scale: 0.97 }}
                >
                  Spin Now ⚡
                </motion.button>
              </div>
            </motion.div>
          )}
        </AnimatePresence>

        {/* ── Exhausted spins nudge ─────────────────────────────────── */}
        {availableSpins === 0 && nudgeData && (nudgeData.spinsGranted > 0 || nudgeData.spinsUsed > 0) && (
          <motion.div
            className="rounded-2xl border border-purple-200 bg-gradient-to-r from-purple-50 to-violet-50 px-5 py-4 flex items-center justify-between gap-4"
            {...fadeUp(0.1)}
          >
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-xl bg-purple-100 flex items-center justify-center flex-shrink-0">
                <CheckCircle className="w-5 h-5 text-purple-600" />
              </div>
              <div>
                <p className="font-bold text-purple-900 text-sm">
                  All {nudgeData.spinsGranted} spin{nudgeData.spinsGranted > 1 ? 's' : ''} used today
                </p>
                <p className="text-xs text-purple-600 mt-0.5">
                  {nudgeData.nextTierName && nudgeData.amountToNextTier
                    ? `Recharge ₦${Math.ceil(nudgeData.amountToNextTier / 100).toLocaleString()} more to unlock ${nudgeData.nextTierSpins} spins (${nudgeData.nextTierName})`
                    : 'Come back tomorrow for fresh spins!'}
                </p>
              </div>
            </div>
            {nudgeData.nextTierName && (
              <motion.button
                onClick={() => setShowUpgradeNudge(true)}
                className="flex-shrink-0 text-xs border border-purple-300 text-purple-700 hover:bg-purple-100 font-bold px-3 py-2 rounded-xl transition-all"
                whileHover={{ scale: 1.04 }}
                whileTap={{ scale: 0.97 }}
              >
                See how
              </motion.button>
            )}
          </motion.div>
        )}

        {/* ── Tab navigation ───────────────────────────────────────────── */}
        <motion.div className="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden" {...fadeUp(0.15)}>
          <div className="flex overflow-x-auto scrollbar-none">
            {TABS.map((tab, i) => (
              <button
                key={tab}
                onClick={() => setActiveTab(tab)}
                className={`relative flex-shrink-0 px-5 py-4 text-sm font-semibold capitalize transition-colors whitespace-nowrap ${
                  activeTab === tab ? 'text-purple-700' : 'text-gray-500 hover:text-gray-800 hover:bg-gray-50'
                }`}
              >
                {tab}
                {activeTab === tab && (
                  <motion.div
                    layoutId="tab-indicator"
                    className="absolute bottom-0 left-0 right-0 h-0.5 gradient-brand rounded-full"
                    transition={{ type: 'spring', stiffness: 400, damping: 30 }}
                  />
                )}
              </button>
            ))}
          </div>
        </motion.div>

        {/* ── Tab content ──────────────────────────────────────────────── */}
        <AnimatePresence mode="wait">
          <motion.div
            key={activeTab}
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -8 }}
            transition={{ duration: 0.25, ease: [0.16, 1, 0.3, 1] }}
          >

            {/* ════════ OVERVIEW ════════ */}
            {activeTab === 'overview' && (
              <div className="space-y-5">
                <div className="grid gap-5 md:grid-cols-2">
                  {/* Account summary */}
                  <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-5 space-y-3">
                    <h3 className="font-bold text-gray-900 flex items-center gap-2">
                      <User className="w-4 h-4 text-purple-600" /> Account Summary
                    </h3>
                    {[
                      { label: 'Phone', value: dashboardData.user.msisdn },
                      { label: 'Email', value: dashboardData.user.email || 'Not set' },
                      { label: 'Points', value: (dashboardData.user.total_points || 0).toLocaleString() },
                    ].map(({ label, value }) => (
                      <div key={label} className="flex justify-between items-center py-2 border-b border-gray-50 last:border-0">
                        <span className="text-sm text-gray-500">{label}</span>
                        <span className="text-sm font-semibold text-gray-900">{value}</span>
                      </div>
                    ))}
                    <div className="flex justify-between items-center py-2">
                      <span className="text-sm text-gray-500">Loyalty Tier</span>
                      <span className={`text-xs font-bold px-2.5 py-1 rounded-full ${tierStyle.badge}`}>{tierStyle.label}</span>
                    </div>
                  </div>

                  {/* Referral program */}
                  <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-5 space-y-3">
                    <h3 className="font-bold text-gray-900 flex items-center gap-2">
                      <Sparkles className="w-4 h-4 text-amber-500" /> Referral Program
                    </h3>
                    <p className="text-xs text-gray-500">Share your code and earn commission when friends recharge</p>
                    <div className="flex items-center gap-2 mt-2">
                      <div className="flex-1 bg-purple-50 border border-purple-200 rounded-xl px-4 py-2.5 font-mono text-sm font-bold text-purple-700 tracking-widest">
                        {dashboardData.user.referral_code || 'N/A'}
                      </div>
                      <motion.button
                        onClick={copyReferralCode}
                        className="w-10 h-10 rounded-xl gradient-brand flex items-center justify-center text-white flex-shrink-0"
                        whileHover={{ scale: 1.08 }}
                        whileTap={{ scale: 0.9 }}
                      >
                        <Copy className="w-4 h-4" />
                      </motion.button>
                    </div>
                  </div>
                </div>

                {/* Recent activity */}
                <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-5">
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="font-bold text-gray-900">Recent Activity</h3>
                    <button onClick={() => setActiveTab('transactions')} className="text-xs text-purple-600 hover:text-purple-800 font-semibold flex items-center gap-1">
                      View all <ChevronRight className="w-3 h-3" />
                    </button>
                  </div>
                  {dashboardData.recent_transactions && dashboardData.recent_transactions.length > 0 ? (
                    <div className="space-y-2">
                      {dashboardData.recent_transactions.slice(0, 5).map((tx, i) => (
                        <motion.div
                          key={tx.id}
                          className="flex justify-between items-center p-3 rounded-xl hover:bg-gray-50 transition-colors"
                          initial={{ opacity: 0, x: -12 }}
                          animate={{ opacity: 1, x: 0 }}
                          transition={{ delay: i * 0.06, duration: 0.3 }}
                        >
                          <div className="flex items-center gap-3">
                            <div className={`w-10 h-10 rounded-xl flex items-center justify-center ${getNetworkColor(tx.network_provider)}`}>
                              <Phone className="w-4 h-4 text-white" />
                            </div>
                            <div>
                              <p className="text-sm font-semibold text-gray-900">{tx.network_provider} {tx.recharge_type}</p>
                              <p className="text-xs text-gray-400">{formatDate(tx.created_at)}</p>
                            </div>
                          </div>
                          <div className="text-right">
                            <p className="text-sm font-bold text-gray-900">{formatCurrency(tx.amount)}</p>
                            <div className="flex items-center gap-1.5 justify-end">
                              <span className={`text-xs font-bold px-1.5 py-0.5 rounded-full ${
                                tx.status === 'SUCCESS'
                                  ? 'bg-emerald-100 text-emerald-700'
                                  : 'bg-gray-100 text-gray-600'
                              }`}>{tx.status}</span>
                              {tx.points_earned > 0 && (
                                <span className="text-xs text-amber-600 font-bold">+{tx.points_earned}pts</span>
                              )}
                            </div>
                          </div>
                        </motion.div>
                      ))}
                    </div>
                  ) : (
                    <div className="text-center py-10 space-y-3">
                      <Smartphone className="w-10 h-10 text-gray-300 mx-auto" />
                      <p className="text-gray-400 text-sm">No recharges yet</p>
                      <motion.button
                        onClick={() => window.location.href = '/recharge'}
                        className="btn-claim px-5 py-2.5 rounded-xl text-sm font-bold text-white"
                        whileHover={{ scale: 1.04 }}
                        whileTap={{ scale: 0.97 }}
                      >
                        Recharge Now
                      </motion.button>
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* ════════ TRANSACTIONS ════════ */}
            {activeTab === 'transactions' && (
              <div className="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden">
                <div className="p-5 border-b border-gray-100 flex flex-wrap items-center justify-between gap-3">
                  <h3 className="font-bold text-gray-900">Transaction History</h3>
                  <div className="flex gap-2">
                    <div className="relative">
                      <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
                      <Input placeholder="Search..." value={searchTerm} onChange={(e) => setSearchTerm(e.target.value)} className="pl-9 h-9 w-52 text-sm rounded-xl" />
                    </div>
                    <Button variant="outline" size="sm" className="h-9 rounded-xl gap-1.5">
                      <Download className="h-3.5 w-3.5" /> Export
                    </Button>
                  </div>
                </div>
                <div className="overflow-x-auto">
                  <Table>
                    <TableHeader>
                      <TableRow className="bg-gray-50">
                        <TableHead className="font-semibold text-xs uppercase tracking-wide">Date</TableHead>
                        <TableHead className="font-semibold text-xs uppercase tracking-wide">Network</TableHead>
                        <TableHead className="font-semibold text-xs uppercase tracking-wide">Type</TableHead>
                        <TableHead className="font-semibold text-xs uppercase tracking-wide">Amount</TableHead>
                        <TableHead className="font-semibold text-xs uppercase tracking-wide">Points</TableHead>
                        <TableHead className="font-semibold text-xs uppercase tracking-wide">Status</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {filteredTransactions.length > 0 ? filteredTransactions.map((tx) => (
                        <TableRow key={tx.id} className="hover:bg-purple-50/30 transition-colors">
                          <TableCell className="text-sm text-gray-600">{formatDate(tx.created_at)}</TableCell>
                          <TableCell>
                            <span className={`text-xs font-bold px-2 py-1 rounded-lg ${getNetworkColor(tx.network_provider)} text-white`}>{tx.network_provider}</span>
                          </TableCell>
                          <TableCell className="text-sm">{tx.recharge_type}</TableCell>
                          <TableCell className="font-semibold text-sm">{formatCurrency(tx.amount)}</TableCell>
                          <TableCell>
                            {tx.points_earned > 0
                              ? <span className="text-xs font-bold text-amber-600">+{tx.points_earned}</span>
                              : <span className="text-gray-300">—</span>
                            }
                          </TableCell>
                          <TableCell>
                            <span className={`text-xs font-bold px-2 py-1 rounded-full ${
                              tx.status === 'SUCCESS' ? 'bg-emerald-100 text-emerald-700' : 'bg-gray-100 text-gray-600'
                            }`}>{tx.status}</span>
                          </TableCell>
                        </TableRow>
                      )) : (
                        <TableRow>
                          <TableCell colSpan={6} className="text-center py-12 text-gray-400">No transactions found</TableCell>
                        </TableRow>
                      )}
                    </TableBody>
                  </Table>
                </div>
              </div>
            )}

            {/* ════════ SUBSCRIPTIONS ════════ */}
            {activeTab === 'subscriptions' && (
              <div className="space-y-5">

                {/* ── Summary stats ── */}
                <div className="grid gap-4 md:grid-cols-4">
                  {/* Active lines count (live) */}
                  <div className="bg-white rounded-2xl shadow-sm border-2 border-green-100 p-5 flex items-center gap-4">
                    <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-green-400 to-emerald-600 flex items-center justify-center flex-shrink-0">
                      <Layers className="w-6 h-6 text-white" />
                    </div>
                    <div>
                      <p className="text-xs text-gray-500 font-medium">Active Lines</p>
                      <p className="text-2xl font-black text-gray-900">{activeLines?.total_active_lines ?? 0}</p>
                    </div>
                  </div>
                  {/* Total daily entries (live = sum of active lines) */}
                  <div className="bg-white rounded-2xl shadow-sm border-2 border-amber-100 p-5 flex items-center gap-4">
                    <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-amber-400 to-orange-500 flex items-center justify-center flex-shrink-0">
                      <Trophy className="w-6 h-6 text-white" />
                    </div>
                    <div>
                      <p className="text-xs text-gray-500 font-medium">Daily Entries</p>
                      <p className="text-2xl font-black text-gray-900">{activeLines?.total_daily_entries ?? dashboardData.summary?.total_subscription_entries ?? 0}</p>
                    </div>
                  </div>
                  {/* All-time subscription count */}
                  <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-5 flex items-center gap-4">
                    <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center flex-shrink-0">
                      <Calendar className="w-6 h-6 text-white" />
                    </div>
                    <div>
                      <p className="text-xs text-gray-500 font-medium">All-time Lines</p>
                      <p className="text-2xl font-black text-gray-900">{dashboardData.summary?.total_subscriptions ?? 0}</p>
                    </div>
                  </div>
                  {/* Points earned */}
                  <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-5 flex items-center gap-4">
                    <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-violet-500 to-purple-600 flex items-center justify-center flex-shrink-0">
                      <Award className="w-6 h-6 text-white" />
                    </div>
                    <div>
                      <p className="text-xs text-gray-500 font-medium">Points Earned</p>
                      <p className="text-2xl font-black text-gray-900">{dashboardData.summary?.total_subscription_points ?? 0}</p>
                    </div>
                  </div>
                </div>

                {/* ── Active subscription lines ── */}
                {activeLines && activeLines.total_active_lines > 0 && (
                  <div className="bg-white rounded-2xl shadow-sm border-2 border-green-100 overflow-hidden">
                    <div className="p-5 border-b border-green-100 bg-green-50 flex items-center justify-between">
                      <div>
                        <h3 className="font-bold text-green-900 flex items-center gap-2">
                          <Layers className="w-4 h-4" /> Active Subscription Lines
                        </h3>
                        <p className="text-xs text-green-700 mt-0.5">
                          {activeLines.total_daily_entries} guaranteed entries/day · ₦{activeLines.total_daily_cost_ngn}/day total
                        </p>
                      </div>
                      <motion.button
                        onClick={() => window.location.href = '/subscription'}
                        className="flex items-center gap-1.5 bg-green-600 text-white text-sm font-bold px-4 py-2 rounded-xl"
                        whileHover={{ scale: 1.03 }} whileTap={{ scale: 0.97 }}
                      >
                        <Plus className="w-4 h-4" /> Add Line
                      </motion.button>
                    </div>
                    <div className="divide-y divide-gray-50">
                      {activeLines.lines.map(line => (
                        <div key={line.id} className="flex items-center justify-between px-5 py-3 hover:bg-gray-50">
                          <div className="flex items-center gap-3">
                            <div className="w-9 h-9 rounded-full bg-green-100 flex items-center justify-center">
                              <Trophy className="w-4 h-4 text-green-600" />
                            </div>
                            <div>
                              <p className="font-semibold text-gray-800 text-sm">
                                {line.entries} {line.entries === 1 ? 'entry' : 'entries'}/day
                                <span className="text-gray-400 font-normal ml-1">— ₦{line.daily_amount_ngn}/day</span>
                              </p>
                              <p className="text-xs text-gray-400 font-mono">{line.code}</p>
                            </div>
                          </div>
                          <div className="flex items-center gap-3">
                            <span className="text-xs text-gray-400">
                              Next: {new Date(line.next_billing).toLocaleDateString('en-NG', { month: 'short', day: 'numeric' })}
                            </span>
                            <span className="text-xs font-bold px-2 py-0.5 rounded-full bg-emerald-100 text-emerald-700">active</span>
                            <button
                              onClick={async () => {
                                if (!confirm(`Cancel ${line.code}?`)) return;
                                setCancellingLine(line.id);
                                try {
                                  await apiClient.post(`/subscription/cancel/${line.id}`, { msisdn: user?.msisdn });
                                  toast({ title: 'Line cancelled', description: `${line.code} has been cancelled.` });
                                  fetchDashboardData();
                                } catch {
                                  toast({ title: 'Cancel failed', variant: 'destructive' });
                                } finally { setCancellingLine(null); }
                              }}
                              disabled={cancellingLine === line.id}
                              className="text-red-400 hover:text-red-600 transition-colors p-1 rounded"
                              title="Cancel this line"
                            >
                              {cancellingLine === line.id
                                ? <Loader2 className="w-4 h-4 animate-spin" />
                                : <XCircle className="w-4 h-4" />}
                            </button>
                          </div>
                        </div>
                      ))}
                      {/* Total bar */}
                      <div className="flex items-center justify-between bg-green-700 text-white px-5 py-3">
                        <span className="font-bold text-sm">Total daily guaranteed</span>
                        <div className="flex items-center gap-4 text-sm font-bold">
                          <span>{activeLines.total_daily_entries} entries/day</span>
                          <span className="opacity-75">·</span>
                          <span>₦{activeLines.total_daily_cost_ngn}/day</span>
                        </div>
                      </div>
                    </div>
                  </div>
                )}

                {/* ── Subscription history table ── */}
                <div className="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden">
                  <div className="p-5 border-b border-gray-100">
                    <h3 className="font-bold text-gray-900">Subscription History</h3>
                    <p className="text-xs text-gray-500 mt-0.5">All lines ever created (points awarded on confirmed daily payments only)</p>
                  </div>
                  <div className="overflow-x-auto">
                    <Table>
                      <TableHeader>
                        <TableRow className="bg-gray-50">
                          <TableHead className="text-xs font-semibold uppercase tracking-wide">Date</TableHead>
                          <TableHead className="text-xs font-semibold uppercase tracking-wide">Code</TableHead>
                          <TableHead className="text-xs font-semibold uppercase tracking-wide">Amount</TableHead>
                          <TableHead className="text-xs font-semibold uppercase tracking-wide">Entries</TableHead>
                          <TableHead className="text-xs font-semibold uppercase tracking-wide">Points</TableHead>
                          <TableHead className="text-xs font-semibold uppercase tracking-wide">Status</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {dashboardData.subscriptions && dashboardData.subscriptions.length > 0
                          ? dashboardData.subscriptions.map((sub) => (
                            <TableRow key={sub.id} className="hover:bg-purple-50/30">
                              <TableCell className="text-sm">{formatDate(sub.transaction_date)}</TableCell>
                              <TableCell className="font-mono text-xs text-purple-700">{sub.subscription_code}</TableCell>
                              <TableCell className="font-semibold">{formatCurrency(sub.amount)}</TableCell>
                              <TableCell>{sub.entries}</TableCell>
                              <TableCell className="text-amber-600 font-bold">+{sub.points_earned}</TableCell>
                              <TableCell>
                                <span className={`text-xs font-bold px-2 py-1 rounded-full ${
                                  sub.status === 'active'    ? 'bg-emerald-100 text-emerald-700' :
                                  sub.status === 'pending'   ? 'bg-amber-100 text-amber-700' :
                                  sub.status === 'cancelled' ? 'bg-red-100 text-red-600' :
                                  sub.status === 'paused'    ? 'bg-orange-100 text-orange-700' :
                                  'bg-gray-100 text-gray-600'
                                }`}>{sub.status}</span>
                              </TableCell>
                            </TableRow>
                          ))
                          : (
                            <TableRow>
                              <TableCell colSpan={6} className="text-center py-12 text-gray-400">No subscriptions yet</TableCell>
                            </TableRow>
                          )}
                      </TableBody>
                    </Table>
                  </div>
                </div>

                {/* ── Add more CTA ── */}
                <div className="rounded-2xl p-6 text-white flex items-center justify-between"
                     style={{ background: 'linear-gradient(135deg, #7c3aed, #a855f7)' }}>
                  <div>
                    <h4 className="font-bold text-lg">Add Another Subscription Line</h4>
                    <p className="text-purple-200 text-sm mt-0.5">
                      Stack lines to get more daily draw entries. Each line is billed independently.
                    </p>
                  </div>
                  <motion.button
                    onClick={() => window.location.href = '/subscription'}
                    className="bg-white text-purple-700 font-bold px-5 py-2.5 rounded-xl text-sm flex items-center gap-2 shrink-0"
                    whileHover={{ scale: 1.04 }}
                    whileTap={{ scale: 0.97 }}
                  >
                    <Plus className="w-4 h-4" /> Subscribe
                  </motion.button>
                </div>
              </div>
            )}

            {/* ════════ PRIZES ════════ */}
            {activeTab === 'prizes' && dashboardData && (
              <div className="space-y-5">
                <div className="grid gap-4 md:grid-cols-3">
                  {[
                    { label: 'Total Prizes', value: dashboardData?.summary?.total_prizes || 0, color: 'from-violet-500 to-purple-600' },
                    { label: 'Pending', value: dashboardData?.summary?.pending_prizes || 0, color: 'from-amber-400 to-orange-500' },
                    { label: 'Claimed', value: (dashboardData?.summary?.total_prizes || 0) - (dashboardData?.summary?.pending_prizes || 0), color: 'from-emerald-400 to-teal-500' },
                  ].map(({ label, value, color }) => (
                    <div key={label} className="bg-white rounded-2xl shadow-sm border border-gray-100 p-5 flex items-center justify-between">
                      <div>
                        <p className="text-xs text-gray-500 font-medium">{label}</p>
                        <p className="text-3xl font-black text-gray-900 mt-1" style={{ fontFamily: "'Plus Jakarta Sans', sans-serif" }}>{value}</p>
                      </div>
                      <div className={`w-2 h-12 rounded-full bg-gradient-to-b ${color}`} />
                    </div>
                  ))}
                </div>

                <div className="space-y-3">
                  <h3 className="font-bold text-gray-900 px-1">Prize History</h3>
                  {dashboardData?.prizes && dashboardData.prizes.length > 0 ? (
                    dashboardData.prizes.map((prize, index) => (
                      <motion.div
                        key={prize?.id || index}
                        className="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden"
                        initial={{ opacity: 0, y: 10 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ delay: index * 0.06, duration: 0.3 }}
                      >
                        <div className="p-4 space-y-3">
                          {/* Prize header */}
                          <div className="flex justify-between items-start">
                            <div>
                              <div className="flex items-center gap-2">
                                <div className="w-8 h-8 rounded-xl bg-gradient-to-br from-violet-500 to-purple-600 flex items-center justify-center">
                                  <Gift className="w-4 h-4 text-white" />
                                </div>
                                <p className="font-bold text-gray-900">{prize?.prize_name || 'Unknown Prize'}</p>
                              </div>
                              <p className="text-xs text-gray-400 mt-1 ml-10">
                                Won {prize?.won_date ? formatDate(prize.won_date) : 'N/A'}
                                {prize?.claim_date && <span className="text-emerald-600"> · Claimed {formatDate(prize.claim_date)}</span>}
                              </p>
                            </div>
                            <div className="text-right flex-shrink-0">
                              <p className="font-black text-gray-900 text-lg" style={{ fontFamily: "'Plus Jakarta Sans', sans-serif" }}>
                                {prize?.prize_value ? formatCurrency(prize.prize_value) : 'N/A'}
                              </p>
                              <span className={`inline-block text-[10px] font-bold px-2 py-0.5 rounded-full mt-1 ${
                                prize?.status === 'PENDING'              ? 'bg-amber-100 text-amber-700' :
                                prize?.status === 'PENDING_ADMIN_REVIEW' ? 'bg-yellow-100 text-yellow-800' :
                                prize?.status === 'CLAIMED'              ? 'bg-emerald-100 text-emerald-700' :
                                prize?.status === 'APPROVED'             ? 'bg-blue-100 text-blue-700' :
                                prize?.status === 'REJECTED'             ? 'bg-red-100 text-red-700' :
                                'bg-gray-100 text-gray-600'
                              }`}>{prize?.status || 'PENDING'}</span>
                            </div>
                          </div>

                          {/* Fulfillment info for airtime/data */}
                          {(prize?.prize_type === 'AIRTIME' || prize?.prize_type === 'DATA') && (
                            <div className="ml-10 space-y-1">
                              {prize?.fulfillment_mode && (
                                <p className="text-xs text-gray-400">Mode: <span className="font-medium text-gray-600">{prize.fulfillment_mode}</span></p>
                              )}
                              {prize?.claim_reference && (
                                <p className="text-xs text-emerald-600 font-mono">Ref: {prize.claim_reference}</p>
                              )}
                              {prize?.fulfillment_error && (
                                <p className="text-xs text-red-500 bg-red-50 rounded-lg px-3 py-2">⚠️ {prize.fulfillment_error}</p>
                              )}
                            </div>
                          )}

                          {/* Status chips */}
                          {prize?.status === 'PENDING_ADMIN_REVIEW' && (
                            <div className="rounded-xl bg-amber-50 border border-amber-200 px-4 py-3 flex items-start gap-3">
                              <Clock className="w-4 h-4 text-amber-600 mt-0.5 flex-shrink-0" />
                              <div>
                                <p className="font-semibold text-amber-800 text-sm">Under admin review</p>
                                <p className="text-xs text-amber-600 mt-0.5">Your claim is being processed. We'll complete it within 24–48 hours.</p>
                              </div>
                            </div>
                          )}
                          {prize?.status === 'CLAIMED' && (
                            <div className="rounded-xl bg-emerald-50 border border-emerald-200 px-4 py-3 flex items-center gap-3">
                              <CheckCircle className="w-4 h-4 text-emerald-600 flex-shrink-0" />
                              <p className="font-semibold text-emerald-800 text-sm">Prize delivered ✓</p>
                            </div>
                          )}
                          {prize?.status === 'APPROVED' && (
                            <div className="rounded-xl bg-blue-50 border border-blue-200 px-4 py-3 flex items-center gap-3">
                              <CheckCircle className="w-4 h-4 text-blue-600 flex-shrink-0" />
                              <p className="font-semibold text-blue-800 text-sm">Approved — payment processing shortly</p>
                            </div>
                          )}
                          {prize?.status === 'REJECTED' && (
                            <div className="rounded-xl bg-red-50 border border-red-200 px-4 py-3 flex items-start gap-3">
                              <AlertCircle className="w-4 h-4 text-red-600 mt-0.5 flex-shrink-0" />
                              <div>
                                <p className="font-semibold text-red-800 text-sm">Claim rejected</p>
                                <p className="text-xs text-red-600 mt-0.5">Contact support if you believe this is an error.</p>
                              </div>
                            </div>
                          )}

                          {/* Bank form for cash prizes */}
                          {prize?.status === 'PENDING' && prize?.prize_type === 'CASH' && showBankForm === prize.id && (
                            <div className="bg-purple-50 border border-purple-200 rounded-xl p-4 space-y-3">
                              <p className="text-sm font-semibold text-purple-900">Enter bank details to claim:</p>
                              <div className="grid gap-3 sm:grid-cols-2">
                                <div className="space-y-1">
                                  <label className="text-xs font-semibold text-gray-700">Account Name</label>
                                  <Input value={bankDetails.account_name} onChange={(e) => setBankDetails(p => ({ ...p, account_name: e.target.value }))} placeholder="John Doe" className="h-9 rounded-xl" />
                                </div>
                                <div className="space-y-1">
                                  <label className="text-xs font-semibold text-gray-700">Account Number</label>
                                  <Input value={bankDetails.account_number} onChange={(e) => setBankDetails(p => ({ ...p, account_number: e.target.value }))} placeholder="1234567890" className="h-9 rounded-xl" />
                                </div>
                                <div className="sm:col-span-2 space-y-1">
                                  <label className="text-xs font-semibold text-gray-700">Bank</label>
                                  <Select value={bankDetails.bank_name} onValueChange={(val) => { const bank = NIGERIAN_BANKS.find(b => b.name === val); setBankDetails(p => ({ ...p, bank_name: val, bank_code: bank?.code ?? '' })); }}>
                                    <SelectTrigger className="h-9 rounded-xl">
                                      <SelectValue placeholder="Select your bank" />
                                    </SelectTrigger>
                                    <SelectContent>
                                      {NIGERIAN_BANKS.map((b) => <SelectItem key={b.code} value={b.name}>{b.name}</SelectItem>)}
                                    </SelectContent>
                                  </Select>
                                </div>
                              </div>
                              <div className="flex gap-2">
                                <Button onClick={() => handleClaimPrize(prize.id, prize.prize_type)} disabled={claimingPrize === prize.id} className="btn-claim border-0 flex-1">
                                  {claimingPrize === prize.id ? <><Loader2 className="w-4 h-4 animate-spin mr-2"/>Submitting…</> : 'Submit Claim'}
                                </Button>
                                <Button variant="outline" onClick={() => { setShowBankForm(null); setBankDetails({ account_number: '', account_name: '', bank_name: '', bank_code: '' }); }} className="rounded-xl">Cancel</Button>
                              </div>
                            </div>
                          )}

                          {/* Claim button for PENDING prizes */}
                          {prize?.status === 'PENDING' && showBankForm !== prize.id && (
                            <motion.button
                              onClick={() => handleClaimPrize(prize.id, prize.prize_type || 'OTHER')}
                              disabled={claimingPrize === prize.id}
                              className="w-full btn-claim py-3 rounded-xl font-bold text-white flex items-center justify-center gap-2 disabled:opacity-60"
                              whileHover={{ scale: 1.01 }}
                              whileTap={{ scale: 0.98 }}
                            >
                              {claimingPrize === prize.id
                                ? <><Loader2 className="w-4 h-4 animate-spin"/>Claiming…</>
                                : <><Gift className="w-4 h-4"/>Claim Now</>
                              }
                            </motion.button>
                          )}
                        </div>
                      </motion.div>
                    ))
                  ) : (
                    <div className="bg-white rounded-2xl shadow-sm border border-gray-100 text-center py-16 space-y-4">
                      <Trophy className="w-12 h-12 text-gray-200 mx-auto" />
                      <div>
                        <p className="text-gray-500 font-semibold">No prizes yet</p>
                        <p className="text-gray-400 text-sm mt-1">Recharge to spin the wheel and win</p>
                      </div>
                      <motion.button
                        onClick={() => window.location.href = '/recharge'}
                        className="btn-claim px-6 py-3 rounded-xl font-bold text-white text-sm inline-flex items-center gap-2"
                        whileHover={{ scale: 1.04 }}
                        whileTap={{ scale: 0.97 }}
                      >
                        <Zap className="w-4 h-4" /> Recharge Now
                      </motion.button>
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* ════════ PROFILE ════════ */}
            {activeTab === 'profile' && dashboardData && (
              <div className="space-y-5">
                <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-5 space-y-5">
                  <h3 className="font-bold text-gray-900 flex items-center gap-2">
                    <User className="w-4 h-4 text-purple-600" /> Profile Information
                  </h3>
                  <div className="grid gap-4 sm:grid-cols-2">
                    {[
                      { label: 'First Name', value: dashboardData.user.first_name || '' },
                      { label: 'Last Name',  value: dashboardData.user.last_name || '' },
                      { label: 'Phone',      value: dashboardData.user.msisdn },
                      { label: 'Tier',       value: dashboardData.user.loyalty_tier },
                      { label: 'Points',     value: dashboardData.user.total_points.toString() },
                    ].map(({ label, value }) => (
                      <div key={label} className="space-y-1">
                        <label className="text-xs font-semibold text-gray-500 uppercase tracking-wide">{label}</label>
                        <Input value={value} readOnly className="h-10 rounded-xl bg-gray-50 border-gray-100 text-sm font-medium" />
                      </div>
                    ))}
                    <div className="space-y-1 sm:col-span-2">
                      <label className="text-xs font-semibold text-gray-500 uppercase tracking-wide">Email</label>
                      {editingEmail ? (
                        <div className="flex gap-2">
                          <Input value={newEmail} onChange={(e) => setNewEmail(e.target.value)} placeholder="your@email.com" type="email" className="h-10 rounded-xl" />
                          <Button onClick={handleUpdateEmail} disabled={updatingEmail} className="btn-claim border-0 h-10 rounded-xl px-4">
                            {updatingEmail ? <Loader2 className="w-4 h-4 animate-spin" /> : 'Save'}
                          </Button>
                          <Button variant="outline" onClick={() => { setEditingEmail(false); setNewEmail(''); }} className="h-10 rounded-xl px-4">Cancel</Button>
                        </div>
                      ) : (
                        <div className="flex gap-2">
                          <Input value={dashboardData.user.email || 'Not set'} readOnly className="h-10 rounded-xl bg-gray-50 border-gray-100 text-sm font-medium" />
                          <Button variant="outline" onClick={() => { setEditingEmail(true); setNewEmail(dashboardData.user.email || ''); }} className="h-10 rounded-xl px-4">Edit</Button>
                        </div>
                      )}
                    </div>
                  </div>
                  <div className="flex gap-3 pt-2">
                    <Button variant="outline" className="rounded-xl">Edit Profile</Button>
                    <Button variant="outline" onClick={logout} className="rounded-xl text-red-500 hover:text-red-600 hover:bg-red-50">Logout</Button>
                  </div>
                </div>
                <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-5 space-y-3">
                  <h3 className="font-bold text-gray-900 flex items-center gap-2">
                    <Sparkles className="w-4 h-4 text-amber-500" /> Referral Code
                  </h3>
                  <p className="text-xs text-gray-500">Share this code with friends to earn commission</p>
                  <div className="flex items-center gap-2">
                    <div className="flex-1 bg-purple-50 border border-purple-200 rounded-xl px-4 py-3 font-mono text-base font-bold text-purple-700 tracking-widest">
                      {dashboardData.user.referral_code || 'N/A'}
                    </div>
                    <motion.button onClick={copyReferralCode} className="w-11 h-11 rounded-xl gradient-brand flex items-center justify-center text-white" whileHover={{ scale: 1.08 }} whileTap={{ scale: 0.9 }}>
                      <Copy className="w-4 h-4" />
                    </motion.button>
                  </div>
                </div>
              </div>
            )}

          </motion.div>
        </AnimatePresence>
      </div>

      {/* Spin Wheel Modal */}
      {showSpinWheel && availableSpins > 0 && (
        <SpinWheel
          isOpen={showSpinWheel}
          onClose={async () => {
            setShowSpinWheel(false);
            fetchDashboardData();
            try {
              const res = await apiClient.get('/spin/eligibility');
              const d = res.data?.data ?? {};
              const remaining: number = d.available_spins ?? 0;
              setAvailableSpins(remaining);
              if (remaining <= 0 && (d.spins_granted_today ?? 0) > 0) {
                setTimeout(() => {
                  setNudgeData({ spinsGranted: d.spins_granted_today ?? 0, spinsUsed: d.spins_used_today ?? 0, nextTierName: d.next_tier_name, nextTierMinAmount: d.next_tier_min_amount, amountToNextTier: d.amount_to_next_tier, nextTierSpins: d.next_tier_spins });
                  setShowUpgradeNudge(true);
                }, 400);
              }
            } catch { setAvailableSpins(0); }
          }}
          transactionAmount={1000}
          userPhone={user?.msisdn || ''}
          onPrizeWon={async (_prize) => {
            try { const res = await apiClient.get('/spin/eligibility'); setAvailableSpins(res.data?.data?.available_spins ?? 0); } catch {}
          }}
          onSpinLimitReached={() => {
            // Backend rejected spin due to stale frontend state — reset and show nudge
            setShowSpinWheel(false);
            setAvailableSpins(0);
            checkPendingSpins(); // re-fetch nudge data and update availableSpins
          }}
        />
      )}

      {/* Upgrade Nudge */}
      {showUpgradeNudge && nudgeData && (
        <SpinUpgradeNudge
          isOpen={showUpgradeNudge}
          onClose={() => { setShowUpgradeNudge(false); setNudgeData(null); }}
          spinsGranted={nudgeData.spinsGranted}
          spinsUsed={nudgeData.spinsUsed}
          nextTierName={nudgeData.nextTierName}
          nextTierMinAmount={nudgeData.nextTierMinAmount}
          amountToNextTier={nudgeData.amountToNextTier}
          nextTierSpins={nudgeData.nextTierSpins}
        />
      )}
    </div>
  );
};
