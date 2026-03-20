import React, { useState, useEffect, useRef, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { apiClient } from '@/lib/api-client';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { PremiumRechargeForm } from '@/components/recharge/PremiumRechargeForm';
import { DrawsList } from '@/components/draws/DrawsList';
import { SpinWheel } from '@/components/games/SpinWheel';
import { SpinUpgradeNudge } from '@/components/games/SpinUpgradeNudge';
import { DailySpinProgress } from '@/components/spin/DailySpinProgress';
import { toast } from '@/components/ui/sonner';
import { useAuthContext } from '@/contexts/AuthContext';
import { formatCurrency } from '@/lib/utils';
import { getAvailableSpins, getPlatformStatistics, getRecentWinners } from '@/lib/api';
import {
  Smartphone,
  Trophy,
  Users,
  Zap,
  Gift,
  Star,
  Shield,
  Clock,
  Sparkles,
  ArrowRight,
  CheckCircle,
  Flame,
  Award,
  DollarSign,
  ChevronDown,
  Play,
  TrendingUp,
  Crown,
  Target,
  Wifi,
  Loader2,
} from 'lucide-react';

/* ─── types ───────────────────────────────────────────────────── */
interface Stats {
  totalUsers: number;
  totalRecharges: number;
  totalPrizes: number;
  activeDraw: { name: string; prizePool: number; endTime: string; entries: number } | null;
}

interface RecentWinner {
  id: string;
  name: string;
  prize: string;
  amount: number;
  time: string;
  network: string;
}

/* ─── animated counter hook ──────────────────────────────────── */
function useCountUp(target: number, duration = 1800, trigger = true) {
  const [value, setValue] = useState(0);
  useEffect(() => {
    if (!trigger || target === 0) return;
    let start = 0;
    const step = Math.ceil(target / (duration / 16));
    const id = setInterval(() => {
      start = Math.min(start + step, target);
      setValue(start);
      if (start >= target) clearInterval(id);
    }, 16);
    return () => clearInterval(id);
  }, [target, trigger]);
  return value;
}

/* ─── floating particle ──────────────────────────────────────── */
const FloatingParticle: React.FC<{ delay: number; x: number; size: number }> = ({ delay, x, size }) => (
  <div
    className="absolute rounded-full bg-white/10 animate-float-up"
    style={{
      width: size,
      height: size,
      left: `${x}%`,
      bottom: '-20px',
      animationDelay: `${delay}s`,
      animationDuration: `${4 + delay}s`,
    }}
  />
);

/* ─── ticker item ─────────────────────────────────────────────── */
const TICKER_WINNERS = [
  '🏆 Adaeze W. just won ₦50,000!',
  '🎰 Chukwuemeka O. spun and won ₦10,000!',
  '🎉 Fatimah B. won a Data Bundle!',
  '💰 Obinna N. won ₦25,000 in daily draw!',
  '🎯 Ngozi A. won ₦5,000 instantly!',
  '🥇 Tunde F. won ₦100,000 prize!',
];

/* ══════════════════════════════════════════════════════════════ */
export const EnterpriseHomePage: React.FC = () => {
  const { user, isAuthenticated } = useAuthContext();
  // toast comes from sonner (global, no subscription needed)
  const [rechargeSuccess, setRechargeSuccess] = useState<any>(null);
  const [stats, setStats] = useState<Stats>({ totalUsers: 0, totalRecharges: 0, totalPrizes: 0, activeDraw: null });
  const [recentWinners, setRecentWinners] = useState<RecentWinner[]>([]);
  const [showSpinWheel, setShowSpinWheel] = useState(false);
  const [availableSpins, setAvailableSpins] = useState(0);
  // Upgrade nudge — shown when spins are exhausted for logged-in users
  const [showUpgradeNudge, setShowUpgradeNudge] = useState(false);
  const [nudgeData, setNudgeData] = useState<{
    spinsGranted: number;
    spinsUsed: number;
    nextTierName?: string;
    nextTierMinAmount?: number;
    amountToNextTier?: number;
    nextTierSpins?: number;
  } | null>(null);
  const [userPhone, setUserPhone] = useState('');
  const [statsVisible, setStatsVisible] = useState(false);
  const [timeRemaining, setTimeRemaining] = useState('');
  const [tickerIndex, setTickerIndex] = useState(0);
  const statsRef = useRef<HTMLDivElement>(null);

  /* animated counters */
  const countUsers = useCountUp(stats.totalUsers, 1600, statsVisible);
  const countRecharges = useCountUp(stats.totalRecharges, 1800, statsVisible);
  const countPrizes = useCountUp(stats.totalPrizes, 1400, statsVisible);

  /* intersection observer for stats section */
  useEffect(() => {
    const el = statsRef.current;
    if (!el) return;
    const obs = new IntersectionObserver((entries) => { if (entries[0]?.isIntersecting) setStatsVisible(true); }, { threshold: 0.3 });
    obs.observe(el);
    return () => obs.disconnect();
  }, []);

  /* ticker */
  useEffect(() => {
    const id = setInterval(() => setTickerIndex((i) => (i + 1) % TICKER_WINNERS.length), 3000);
    return () => clearInterval(id);
  }, []);

  /* countdown */
  useEffect(() => {
    if (!stats.activeDraw) return;
    const tick = () => {
      const diff = new Date(stats.activeDraw!.endTime).getTime() - Date.now();
      if (diff <= 0) return setTimeRemaining('Draw Ended');
      const h = Math.floor(diff / 3600000);
      const m = Math.floor((diff % 3600000) / 60000);
      const s = Math.floor((diff % 60000) / 1000);
      setTimeRemaining(`${h}h ${m}m ${s}s`);
    };
    tick();
    const id = setInterval(tick, 1000);
    return () => clearInterval(id);
  }, [stats.activeDraw]);

  // openSpinWheelIfEligible — checks eligibility with the server.
  //
  // IMPORTANT INTENT: the spin wheel should NEVER auto-popup on a logged-in user
  // without them explicitly clicking a button.  When this is called automatically
  // after a recharge, we show a toast with an action button instead of forcing the modal open.
  //
  // The only time setShowSpinWheel(true) is called here is:
  //   a) When the user already clicked an explicit "Spin Now" button (triggerByUser=true), OR
  //   b) For guest users (no auth) where we have no dashboard banner to rely on.
  //
  // Why: even if the server says "you have 3 spins available", interrupting a user
  // mid-page with a modal popup is jarring, especially right after a recharge.
  const openSpinWheelIfEligible = useCallback(async (triggeredByUser = false) => {
    if (isAuthenticated) {
      try {
        const res = await apiClient.get('/spin/eligibility');
        const d = res.data?.data ?? {};
        const spins: number = d.available_spins ?? 0;

        if (spins > 0) {
          setAvailableSpins(spins);

          if (triggeredByUser) {
            // User explicitly clicked "Spin Now" — open the modal
            setShowSpinWheel(true);
          } else {
            // Called automatically after recharge — show a non-intrusive toast
            // with a button so the user can choose to spin on their own terms.
            toast.success(`🎡 You have ${spins} free spin${spins > 1 ? 's' : ''}!`, {
              description: 'Tap "Spin Now" to play — prizes include airtime, data & cash.',
              duration: 10000,
              action: {
                label: 'Spin Now ⚡',
                onClick: () => setShowSpinWheel(true),
              },
            });
          }

        } else if (d.spins_granted_today > 0 || d.spins_used_today > 0) {
          // All spins used today — update count to zero and show upgrade nudge
          setAvailableSpins(0);
          setNudgeData({
            spinsGranted:      d.spins_granted_today  ?? 0,
            spinsUsed:         d.spins_used_today      ?? 0,
            nextTierName:      d.next_tier_name,
            nextTierMinAmount: d.next_tier_min_amount,
            amountToNextTier:  d.amount_to_next_tier,
            nextTierSpins:     d.next_tier_spins,
          });
          if (triggeredByUser) {
            setShowUpgradeNudge(true);
          }
          // When auto-triggered after recharge: the recharge success toast already
          // tells the user about points/entries — no need to immediately nag about tiers.
        } else {
          if (triggeredByUser) {
            toast('No spins available', { description: 'Recharge ₦1,000+ to unlock a spin.' });
          }
        }
      } catch {
        if (triggeredByUser) {
          toast.error('Could not check eligibility', { description: 'Please try again shortly.' });
        }
      }
    } else {
      // Guest path: show the wheel immediately.
      // /spin/play will validate the qualifying transaction (within 4 hours) server-side.
      setAvailableSpins(1);
      setShowSpinWheel(true);
    }
  }, [isAuthenticated]);

  /* recharge success */
  const handleRechargeSuccess = (result: any) => {
    setRechargeSuccess(result);
    fetchPlatformData();
    // After a successful recharge, verify eligibility server-side before showing wheel
    if (result.amount >= 1000) {
      openSpinWheelIfEligible(false); // auto-called after form recharge, not user-initiated
    }
  };

  /* ── boot: payment callback + fetch data ─────────────────────── */
  useEffect(() => {
    fetchPlatformData();

    const hash = window.location.hash;
    const hashQ = hash.includes('?') ? hash.split('?')[1] : '';
    const allParams = new URLSearchParams();
    new URLSearchParams(hashQ).forEach((v, k) => allParams.set(k, v));
    new URLSearchParams(window.location.search).forEach((v, k) => allParams.set(k, v));

    const paymentStatus = allParams.get('payment');
    const paymentSuccess = allParams.get('payment_success') === 'true' || paymentStatus === 'success';
    const subscriptionSuccess = allParams.get('subscription_success') === 'true';
    const reference = allParams.get('reference') || allParams.get('ref');

    if (paymentSuccess && reference) {
      window.history.replaceState({}, document.title, window.location.pathname);
      // Show immediate "processing" feedback so the user knows something is happening
      toast('✅ Payment Confirmed!', { description: 'Your recharge is being processed, please wait…', duration: 5000 });

      // Convert 234XXXXXXXXXX → 0XXXXXXXXXX for display
      const toLocalPhone = (msisdn: string) =>
        msisdn?.startsWith('234') ? '0' + msisdn.slice(3) : msisdn;

      // Show a persistent "processing" banner so user doesn't think it failed
      // while VTPass is working in the background (can take up to ~2 minutes).
      setRechargeSuccess({ amount: 0, points: 0, drawEntries: 0, spinEligible: false, phone: '', network: '', transactionReference: reference, pending: true });

      // Polling strategy (adaptive intervals, total window ≈ 20 minutes):
      //   Attempts  1-30  (0-60s):   every 2s  — fast VTPass responses
      //   Attempts 31-60  (60-210s): every 5s  — VTPass requery window (backend polls every 30s)
      //   Attempts 61-90 (210-510s): every 10s — extended wait for slow VTPass delivery
      //   Attempts 91-120 (510-1110s): every 20s — final safety buffer up to ~18 min
      // Total window: ~18-20 minutes, matching the backend's 15-minute requery loop.
      const pollTransaction = (attempt = 0, maxAttempts = 120) => {
        const delay = attempt < 30 ? 2000 : attempt < 60 ? 5000 : attempt < 90 ? 10000 : 20000;
        setTimeout(() => {
          apiClient.get(`/recharge/reference/${reference}`)
            .then(res => res.data)
            .then(response => {
              const txn = response.data || response;
              if (txn.status === 'SUCCESS' || txn.status === 'COMPLETED') {
                const amount       = txn.amount / 100;                          // kobo → naira
                const points       = txn.points_earned || 0;
                const drawEntries  = txn.draw_entries  || 0;
                const spinEligible = txn.spin_eligible || amount >= 1000;
                const displayPhone = toLocalPhone(txn.msisdn || '');

                if (txn.msisdn && txn.msisdn !== 'null') setUserPhone(txn.msisdn);

                setRechargeSuccess({
                  amount,
                  points,
                  drawEntries,
                  spinEligible,
                  phone:               displayPhone,
                  network:             txn.network_provider || 'MTN',
                  transactionReference: reference,
                  pending:             false,
                });

                // Build reward summary line
                const rewardParts: string[] = [];
                if (points > 0) rewardParts.push(`${points} point${points !== 1 ? 's' : ''}`);
                if (drawEntries > 0) rewardParts.push(`${drawEntries} draw entr${drawEntries !== 1 ? 'ies' : 'y'}`);
                if (spinEligible) rewardParts.push('spin wheel unlocked! 🎰');
                const rewardLine = rewardParts.length
                  ? ` You earned ${rewardParts.join(' and ')}.`
                  : '';

                toast.success('🎉 Recharge Successful!', {
                  description: `₦${amount.toLocaleString()} recharged to ${displayPhone}.${rewardLine}`,
                  duration: 8000,
                });

                if (spinEligible) {
                  // Always verify with the backend rather than trusting the frontend flag
                  setTimeout(() => openSpinWheelIfEligible(false), 800); // auto after payment callback
                }
              } else if (txn.status === 'FAILED') {
                setRechargeSuccess(null);
                toast.error('Recharge Failed', {
                  description: txn.failure_reason || 'Transaction could not be completed. Please contact support.',
                  duration: 10000,
                });
              } else if (attempt < maxAttempts - 1) {
                // Still PROCESSING — keep polling silently
                pollTransaction(attempt + 1, maxAttempts);
              } else {
                // 4-minute window exhausted. Leave the pending banner visible
                // and show a toast so the user knows to check back.
                setRechargeSuccess((prev: any) => prev ? { ...prev, pending: true, timedOut: true } : null);
                toast('⏳ Still Processing…', {
                  description: `Your recharge for Ref: ${reference} is taking longer than usual. The airtime will be delivered — check your phone or transaction history shortly.`,
                  duration: 15000,
                });
              }
            })
            .catch(() => {
              if (attempt < maxAttempts - 1) pollTransaction(attempt + 1, maxAttempts);
            });
        }, delay);
      };
      pollTransaction();
      return;
    }

    if (subscriptionSuccess && reference) {
      window.history.replaceState({}, document.title, window.location.pathname);
      apiClient.get(`/payment/callback?reference=${reference}&gateway=paystack`).catch(console.error);
      const entries = parseInt(allParams.get('entries') || '0');
      const totalEntries = parseInt(allParams.get('totalEntries') || '0');
      setTimeout(() => {
        toast.success('🎉 Subscription Activated!', {
          description: `${entries} draw entries added (${totalEntries} total today). Good luck!`,
          duration: 8000,
        });
      }, 800);
      return;
    }

    const subscriptionStatus = allParams.get('subscription');
    const subscriptionRecorded = allParams.get('subscription_recorded') === 'true';
    const subscriptionAmount = parseFloat(allParams.get('amount') || '0');
    const subscriptionEntries = parseInt(allParams.get('entries') || '0');
    const subscriptionMsisdn = allParams.get('msisdn');

    if (subscriptionStatus === 'success' && subscriptionRecorded && subscriptionAmount > 0) {
      window.history.replaceState({}, document.title, window.location.pathname);
      setTimeout(() => {
        toast.success('🎉 Subscription Successful!', {
          description: `${subscriptionMsisdn} subscribed with ${subscriptionEntries} ${subscriptionEntries === 1 ? 'entry' : 'entries'}.`,
          duration: 8000,
        });
      }, 800);
      return;
    }

    if (subscriptionStatus === 'failed') {
      window.history.replaceState({}, document.title, window.location.pathname);
      setTimeout(() => {
        toast.error('Subscription Failed', { description: 'Please try again.', duration: 6000 });
      }, 800);
    }
  }, []);

  const fetchPlatformData = async () => {
    try {
      const statsResponse = await getPlatformStatistics();
      if ('success' in statsResponse && statsResponse.success && statsResponse.data) {
        setStats({
          totalUsers: statsResponse.data.totalUsers || 0,
          totalRecharges: statsResponse.data.totalTransactions || 0,
          totalPrizes: statsResponse.data.totalPrizes || 0,
          activeDraw: statsResponse.data.activeDraw || null,
        });
      }
      const winnersResponse = await getRecentWinners(6);
      if ('success' in winnersResponse && winnersResponse.success && Array.isArray(winnersResponse.data) && winnersResponse.data.length > 0) {
        setRecentWinners(
          winnersResponse.data.map((w: any, i: number) => ({
            id: `w_${i}`,
            name: w.full_name,
            prize: w.prize_description,
            amount: w.prize_value,
            time: new Date(w.created_at).toLocaleString(),
            network: w.network_provider || 'MTN',
          }))
        );
      }
    } catch (e) {
      console.error('fetchPlatformData error', e);
    }
  };

  const scrollToRecharge = () =>
    document.getElementById('recharge-form')?.scrollIntoView({ behavior: 'smooth' });

  // Shared helper: fetch fresh eligibility, set nudgeData and open the upgrade modal.
  // Called both from onClose (after all spins are used) and from onSpinLimitReached
  // (when the backend rejects mid-spin because state was stale).
  const triggerUpgradeNudgeIfExhausted = useCallback(async () => {
    if (!isAuthenticated) return;
    try {
      const res = await apiClient.get('/spin/eligibility');
      const d = res.data?.data ?? {};
      setAvailableSpins(d.available_spins ?? 0);
      if ((d.available_spins ?? 0) <= 0 && (d.spins_granted_today ?? 0) > 0) {
        setNudgeData({
          spinsGranted:      d.spins_granted_today  ?? 0,
          spinsUsed:         d.spins_used_today      ?? 0,
          nextTierName:      d.next_tier_name,
          nextTierMinAmount: d.next_tier_min_amount,
          amountToNextTier:  d.amount_to_next_tier,
          nextTierSpins:     d.next_tier_spins,
        });
        setShowUpgradeNudge(true);
      }
    } catch { /* ignore */ }
  }, [isAuthenticated]);

  /* ═══════════════════════ render ════════════════════════════════ */
  return (
    <div className="min-h-screen bg-white overflow-x-hidden">

      {/* ── Live ticker ──────────────────────────────────────────── */}
      <div className="bg-gradient-to-r from-[#7c3aed] to-[#f59e0b] text-white py-2 overflow-hidden">
        <div className="flex items-center justify-center gap-3 text-sm font-medium px-4">
          <div className="flex items-center gap-1.5 flex-shrink-0">
            <div className="w-2 h-2 rounded-full bg-white animate-pulse" />
            <span className="font-bold tracking-widest text-xs uppercase">Live</span>
          </div>
          <div className="overflow-hidden h-5 relative flex-1 max-w-md">
            <p
              key={tickerIndex}
              className="absolute inset-0 flex items-center justify-center text-center animate-ticker-in"
            >
              {TICKER_WINNERS[tickerIndex]}
            </p>
          </div>
        </div>
      </div>

      {/* ── Hero ─────────────────────────────────────────────────── */}
      <section className="relative overflow-hidden text-white" style={{ background: 'linear-gradient(160deg, #0f0520 0%, #1a0b3b 35%, #3b0764 70%, #7c3aed 100%)' }}>
        {/* Background texture */}
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top_right,_rgba(245,158,11,0.15)_0%,_transparent_50%)]" />
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_bottom_left,_rgba(124,58,237,0.3)_0%,_transparent_60%)]" />
        {/* Subtle grid overlay */}
        <div className="absolute inset-0 opacity-5" style={{ backgroundImage: 'linear-gradient(rgba(255,255,255,0.05) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.05) 1px, transparent 1px)', backgroundSize: '40px 40px' }} />

        {/* Floating particles */}
        {[
          { delay: 0, x: 15, size: 12 },
          { delay: 1, x: 25, size: 8 },
          { delay: 2, x: 70, size: 16 },
          { delay: 0.5, x: 85, size: 10 },
          { delay: 1.5, x: 50, size: 6 },
          { delay: 3, x: 40, size: 14 },
        ].map((p, i) => <FloatingParticle key={i} {...p} />)}

        <div className="relative max-w-screen-xl mx-auto px-4 py-16 md:py-24 lg:py-28">
          <div className="grid lg:grid-cols-2 gap-12 lg:gap-16 items-center">

            {/* Left: copy */}
            <motion.div
              className="space-y-7 text-center lg:text-left"
              initial={{ opacity: 0, x: -30 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.7, ease: [0.16, 1, 0.3, 1] }}
            >
              <div className="space-y-2">
                <Badge className="bg-amber-400/20 text-amber-200 border border-amber-400/30 backdrop-blur-sm px-4 py-1 text-sm font-semibold">
                  <Flame className="w-3.5 h-3.5 mr-1.5" />
                  Nigeria's #1 Gamified Recharge Platform
                </Badge>
                <h1 className="text-5xl sm:text-6xl lg:text-7xl font-black tracking-tight leading-[1.05]">
                  Recharge &<br />
                  <span className="bg-gradient-to-r from-yellow-300 via-amber-300 to-yellow-200 bg-clip-text text-transparent">
                    Win Big!
                  </span>
                </h1>
                <p className="text-xl text-blue-100 leading-relaxed max-w-lg mx-auto lg:mx-0">
                  Turn every mobile top-up into a chance to win amazing prizes — daily draws, spin wheel, instant rewards!
                </p>
              </div>

              <div className="flex flex-wrap gap-3 justify-center lg:justify-start">
                <Button
                  size="lg"
                  onClick={scrollToRecharge}
                  className="bg-white text-purple-700 hover:bg-amber-50 font-bold px-7 py-3 text-base shadow-xl shadow-purple-900/30 hover:shadow-2xl hover:scale-105 transition-all rounded-xl"
                >
                  <Zap className="w-4 h-4 mr-2" />
                  Recharge Now
                </Button>
                <Button
                  size="lg"
                  variant="outline"
                  onClick={() => window.location.href = '/draws'}
                  className="border-white/40 text-white hover:bg-white/10 font-semibold px-7 py-3 text-base rounded-xl backdrop-blur-sm"
                >
                  <Trophy className="w-4 h-4 mr-2" />
                  View Prizes
                </Button>
              </div>

              {/* trust badges */}
              <div className="flex flex-wrap gap-4 items-center justify-center lg:justify-start text-sm text-purple-200">
                <span className="flex items-center gap-1.5"><Shield className="w-4 h-4 text-green-300" /> Secure Payments</span>
                <span className="flex items-center gap-1.5"><CheckCircle className="w-4 h-4 text-green-300" /> Instant Recharge</span>
                <span className="flex items-center gap-1.5"><Star className="w-4 h-4 text-yellow-300" /> Real Prizes</span>
              </div>

            </motion.div>

            {/* Right: hero card */}
            <motion.div
              className="relative flex justify-center"
              initial={{ opacity: 0, x: 30 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.7, delay: 0.15, ease: [0.16, 1, 0.3, 1] }}
            >
              <div className="absolute -inset-4 bg-gradient-to-r from-yellow-400/30 to-orange-500/30 rounded-3xl blur-2xl" />
              <div className="relative bg-white/10 backdrop-blur-md border border-white/20 rounded-3xl p-6 w-full max-w-sm shadow-2xl">
                {/* Prize pool card */}
                <div className="gradient-gold rounded-2xl p-5 mb-4 text-white shadow-lg shadow-amber-500/30">
                  <div className="flex items-center gap-2 mb-1">
                    <Trophy className="w-5 h-5" />
                    <span className="text-sm font-semibold opacity-90">Today's Prize Pool</span>
                  </div>
                  <div className="text-4xl font-black">
                    {formatCurrency(stats.activeDraw?.prizePool || 500000)}
                  </div>
                  {stats.activeDraw && (
                    <div className="flex items-center gap-1.5 mt-2 text-sm opacity-80">
                      <Clock className="w-3.5 h-3.5" />
                      <span>{timeRemaining || 'Loading…'}</span>
                    </div>
                  )}
                </div>

                {/* Mini stat pills */}
                <div className="grid grid-cols-3 gap-2">
                  {[
                    { label: 'Users', value: `${stats.totalUsers > 0 ? (stats.totalUsers / 1000).toFixed(0) + 'k+' : '10k+'}`, icon: Users },
                    { label: 'Prizes', value: `${stats.totalPrizes > 0 ? stats.totalPrizes + '+' : '500+'}`, icon: Gift },
                    { label: 'Networks', value: '4', icon: Wifi },
                  ].map(({ label, value, icon: Icon }) => (
                    <div key={label} className="bg-white/10 rounded-xl p-3 text-center">
                      <Icon className="w-4 h-4 mx-auto mb-1 text-purple-200" />
                      <div className="text-white font-bold text-sm">{value}</div>
                      <div className="text-blue-300 text-xs">{label}</div>
                    </div>
                  ))}
                </div>

                {/* CTA pill */}
                <button
                  onClick={scrollToRecharge}
                  className="mt-4 w-full bg-white text-purple-700 font-bold rounded-xl py-3 text-sm flex items-center justify-center gap-2 hover:bg-yellow-50 transition-colors shadow-md hover:scale-[1.02] active:scale-[0.98]"
                >
                  <Zap className="w-4 h-4" />
                  Start Recharging & Win
                  <ArrowRight className="w-4 h-4" />
                </button>
              </div>
          </motion.div>

          </div>{/* /grid */}
          {/* Scroll indicator */}
          <div className="absolute bottom-4 left-1/2 -translate-x-1/2 flex flex-col items-center gap-1 text-white/50 animate-bounce">
            <span className="text-xs">Scroll</span>
            <ChevronDown className="w-4 h-4" />
          </div>
        </div>{/* /max-w */}
      </section>

      {/* ── Live draw banner ─────────────────────────────────────── */}
      {stats.activeDraw && (
        <div className="bg-gradient-to-r from-[#3b0764] to-[#7c3aed] text-white">
          <div className="max-w-screen-xl mx-auto px-4 py-3 flex flex-wrap items-center justify-between gap-3">
            <div className="flex items-center gap-3">
              <div className="w-2.5 h-2.5 rounded-full bg-green-400 animate-pulse flex-shrink-0" />
              <span className="font-semibold text-sm">
                🔴 LIVE: {stats.activeDraw.name} — {formatCurrency(stats.activeDraw.prizePool)} Prize Pool
              </span>
              <span className="text-blue-200 text-sm flex items-center gap-1">
                <Clock className="w-3.5 h-3.5" />
                {timeRemaining}
              </span>
            </div>
            <div className="flex items-center gap-3">
              <span className="text-blue-200 text-sm">
                <strong className="text-white">{stats.activeDraw.entries.toLocaleString()}</strong> entries
              </span>
              <Button
                size="sm"
                onClick={() => window.location.href = '/draws'}
                className="bg-white/20 hover:bg-white/30 text-white border border-white/30 text-xs font-semibold"
              >
                Enter Draw <ArrowRight className="w-3 h-3 ml-1" />
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* ── Recharge Form section ─────────────────────────────────── */}
      <section id="recharge-form" className="scroll-mt-16 bg-gray-50 py-14 md:py-20">
        <div className="max-w-screen-xl mx-auto px-4">

          {/* Header */}
          <div className="text-center mb-10">
            <Badge className="mb-3 bg-purple-100 text-purple-700 border-purple-200 px-4 py-1 font-semibold">
              <Zap className="w-3.5 h-3.5 mr-1.5" />
              Instant Recharge
            </Badge>
            <h2 className="text-4xl font-black text-gray-900 mb-3">Quick Mobile Recharge</h2>
            <p className="text-gray-500 text-lg max-w-xl mx-auto">
              Recharge in seconds, earn draw entries, and unlock the spin wheel with every ₦1,000+
            </p>
          </div>

          {/* Processing banner — shown while polling (pending=true, amount=0) */}
          {rechargeSuccess?.pending && (
            <Alert className="mb-6 border-purple-200 bg-purple-50 max-w-2xl mx-auto">
              <div className="flex items-center gap-2">
                {rechargeSuccess.timedOut ? (
                  <Clock className="h-5 w-5 text-purple-500 shrink-0" />
                ) : (
                  <Loader2 className="h-5 w-5 text-blue-500 animate-spin shrink-0" />
                )}
                <AlertDescription className="text-purple-800 font-medium">
                  {rechargeSuccess.timedOut
                    ? `⏳ Your recharge (Ref: ${rechargeSuccess.transactionReference}) is still being processed by your network provider. The airtime will be delivered shortly — no action needed.`
                    : '⏳ Your recharge is being processed… Please wait, this usually takes under 60 seconds.'}
                </AlertDescription>
              </div>
            </Alert>
          )}

          {/* Success alert — shown when pending=false and amount > 0 */}
          {rechargeSuccess && !rechargeSuccess.pending && rechargeSuccess.amount > 0 && (
            <Alert className="mb-6 border-green-200 bg-green-50 max-w-2xl mx-auto">
              <CheckCircle className="h-5 w-5 text-green-600" />
              <AlertDescription className="text-green-800 font-medium">
                🎉 Recharge of ₦{rechargeSuccess.amount?.toLocaleString()} to {rechargeSuccess.phone} was successful!
                {(rechargeSuccess.points > 0 || rechargeSuccess.drawEntries > 0) && (
                  <>
                    {' '}You earned{' '}
                    {rechargeSuccess.points > 0 && `${rechargeSuccess.points} point${rechargeSuccess.points !== 1 ? 's' : ''}`}
                    {rechargeSuccess.points > 0 && rechargeSuccess.drawEntries > 0 && ' and '}
                    {rechargeSuccess.drawEntries > 0 && `${rechargeSuccess.drawEntries} draw entr${rechargeSuccess.drawEntries !== 1 ? 'ies' : 'y'}`}
                    !
                  </>
                )}
                {rechargeSuccess.spinEligible && ' 🎰 Spin wheel unlocked!'}
              </AlertDescription>
            </Alert>
          )}

          {/* Form + sidebar */}
          <div className="grid lg:grid-cols-3 gap-8 items-start">
            <div className="lg:col-span-2">
              <PremiumRechargeForm onRechargeSuccess={handleRechargeSuccess} />
            </div>

            {/* Sidebar */}
            <div className="space-y-4">
              {/* Spin wheel teaser */}
              <div className="bg-gradient-to-br from-purple-600 to-orange-500 rounded-2xl p-5 text-white shadow-lg">
                <div className="text-center space-y-3">
                  <div className="text-5xl">🎰</div>
                  <h3 className="font-bold text-lg">Spin & Win!</h3>
                  <p className="text-sm text-white/80">Recharge ₦1,000+ to unlock the spin wheel and win instant prizes</p>

                  {userPhone && (
                    <DailySpinProgress
                      msisdn={userPhone}
                      onSpinsUpdate={(spins) => {
                        // Only use this as the initial value if we haven't yet
                        // gotten a fresh count from /spin/eligibility.
                        // availableSpins from eligibility is authoritative (cap - used);
                        // tier-progress returns the raw cap which can be stale.
                        setAvailableSpins((prev) => prev === 0 ? spins : prev);
                      }}
                    />
                  )}

                  {availableSpins > 0 ? (
                    <Button
                      onClick={() => openSpinWheelIfEligible(true)} // user-initiated click
                      className="w-full bg-white text-purple-700 font-bold hover:bg-yellow-50 shadow-md"
                    >
                      🎰 Spin Now! ({availableSpins} left)
                    </Button>
                  ) : (
                    <div className="text-xs text-white/70 bg-white/10 rounded-xl px-3 py-2">
                      Recharge ₦1,000+ to unlock your spin
                    </div>
                  )}
                </div>
              </div>

              {/* Daily subscription teaser */}
              <div className="bg-gradient-to-br from-blue-50 to-indigo-50 border border-purple-200 rounded-2xl p-5">
                <div className="flex items-start gap-3">
                  <div className="w-10 h-10 rounded-xl gradient-brand flex items-center justify-center flex-shrink-0">
                    <Gift className="w-5 h-5 text-white" />
                  </div>
                  <div>
                    <h3 className="font-bold text-gray-900">Daily Subscription</h3>
                    <p className="text-sm text-gray-500 mt-0.5">Get guaranteed draw entries every day for just ₦20</p>
                    <Button
                      size="sm"
                      onClick={() => window.location.href = '/subscription'}
                      className="mt-3 btn-claim text-white text-xs font-semibold border-0"
                    >
                      Subscribe — Only ₦20/day
                    </Button>
                  </div>
                </div>
              </div>

              {/* How points work */}
              <div className="bg-white border border-gray-200 rounded-2xl p-5 space-y-3">
                <h3 className="font-bold text-gray-800 text-sm flex items-center gap-2">
                  <TrendingUp className="w-4 h-4 text-purple-600" />
                  How It Works
                </h3>
                {[
                  { icon: '📱', text: 'Recharge any Nigerian network' },
                  { icon: '🎟', text: 'Every ₦200 = 1 draw entry' },
                  { icon: '🎰', text: '₦1,000+ unlocks spin wheel' },
                  { icon: '🏆', text: 'Win daily cash prizes' },
                ].map((item, i) => (
                  <div key={i} className="flex items-center gap-2.5 text-sm text-gray-600">
                    <span className="text-base">{item.icon}</span>
                    <span>{item.text}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ── Platform Stats ──────────────────────────────────────── */}
      <section ref={statsRef} className="py-14 md:py-20 bg-gradient-to-br from-[#1a0b3b] to-[#3b0764] text-white relative overflow-hidden">
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,_rgba(255,255,255,0.05)_0%,_transparent_70%)]" />
        <div className="relative max-w-screen-xl mx-auto px-4">
          <div className="text-center mb-12">
            <h2 className="text-3xl md:text-4xl font-black mb-3">Platform Impact</h2>
            <p className="text-purple-200 text-lg">Numbers that speak for themselves</p>
          </div>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
            {[
              { label: 'Happy Users', value: countUsers.toLocaleString() + '+', icon: Users, color: 'text-blue-300' },
              { label: 'Total Recharges', value: countRecharges.toLocaleString() + '+', icon: Smartphone, color: 'text-purple-200' },
              { label: 'Prizes Distributed', value: countPrizes.toLocaleString() + '+', icon: Award, color: 'text-yellow-300' },
              { label: 'Networks Supported', value: '4', icon: Wifi, color: 'text-green-300' },
            ].map(({ label, value, icon: Icon, color }) => (
              <div key={label} className="text-center">
                <div className="w-14 h-14 rounded-2xl bg-white/10 backdrop-blur flex items-center justify-center mx-auto mb-3 border border-white/10">
                  <Icon className={`w-6 h-6 ${color}`} />
                </div>
                <div className="text-3xl md:text-4xl font-black mb-1">{value}</div>
                <div className="text-purple-200 text-sm">{label}</div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── How It Works ─────────────────────────────────────────── */}
      <section className="py-14 md:py-20 bg-white">
        <div className="max-w-screen-xl mx-auto px-4">
          <div className="text-center mb-12">
            <Badge className="mb-3 bg-purple-100 text-purple-700 border-purple-200 px-4 py-1 font-semibold">
              <Play className="w-3.5 h-3.5 mr-1.5" />
              How It Works
            </Badge>
            <h2 className="text-3xl md:text-4xl font-black text-gray-900 mb-3">Three Simple Steps</h2>
            <p className="text-gray-500 text-lg max-w-2xl mx-auto">
              Start winning in under 2 minutes — no registration required!
            </p>
          </div>

          <div className="grid md:grid-cols-3 gap-8 relative">
            {/* connector line (desktop) */}
            <div className="hidden md:block absolute top-14 left-[calc(16.666%+2rem)] right-[calc(16.666%+2rem)] h-0.5 bg-gradient-to-r from-purple-200 via-amber-200 to-purple-200" />

            {[
              {
                step: '01',
                title: 'Recharge Your Phone',
                desc: 'Enter phone number, pick network & amount, pay securely via card or bank transfer.',
                icon: Smartphone,
                color: 'from-violet-500 to-purple-600',
                bg: 'bg-purple-50',
                border: 'border-blue-100',
              },
              {
                step: '02',
                title: 'Earn Entries & Spins',
                desc: 'Every ₦200 = 1 prize draw entry. Recharge ₦1,000+ to unlock the spin wheel.',
                icon: Sparkles,
                color: 'from-purple-500 to-purple-600',
                bg: 'bg-purple-50',
                border: 'border-purple-100',
              },
              {
                step: '03',
                title: 'Win Amazing Prizes',
                desc: 'Win cash, data bundles, gadgets & more in daily draws and instant spin wins.',
                icon: Trophy,
                color: 'from-green-500 to-green-600',
                bg: 'bg-green-50',
                border: 'border-green-100',
              },
            ].map(({ step, title, desc, icon: Icon, color, bg, border }) => (
              <div
                key={step}
                className={`${bg} border ${border} rounded-3xl p-7 text-center hover:shadow-xl transition-shadow duration-300 relative`}
              >
                <div className="absolute -top-4 left-1/2 -translate-x-1/2 bg-gray-100 text-gray-400 text-xs font-bold px-3 py-1 rounded-full border border-gray-200">
                  STEP {step}
                </div>
                <div className={`w-16 h-16 rounded-2xl bg-gradient-to-br ${color} flex items-center justify-center text-white mx-auto mb-5 shadow-lg`}>
                  <Icon className="w-7 h-7" />
                </div>
                <h3 className="font-bold text-xl text-gray-900 mb-3">{title}</h3>
                <p className="text-gray-500 text-sm leading-relaxed">{desc}</p>
              </div>
            ))}
          </div>

          <div className="mt-10 text-center">
            <Button
              size="lg"
              onClick={scrollToRecharge}
              className="bg-gradient-to-r gradient-brand text-white font-bold px-10 py-3 rounded-xl shadow-lg hover:shadow-xl hover:scale-105 transition-all"
            >
              <Zap className="w-4 h-4 mr-2" />
              Try It Now — Free!
            </Button>
          </div>
        </div>
      </section>

      {/* ── Feature cards ────────────────────────────────────────── */}
      <section className="py-14 md:py-20 bg-gray-50">
        <div className="max-w-screen-xl mx-auto px-4">
          <div className="text-center mb-12">
            <h2 className="text-3xl md:text-4xl font-black text-gray-900 mb-3">Why Choose RechargeMax?</h2>
            <p className="text-gray-500 text-lg max-w-2xl mx-auto">
              More than a recharge app — it's a full rewards ecosystem
            </p>
          </div>
          <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-6">
            {[
              {
                icon: Smartphone,
                title: 'Instant Recharge',
                desc: 'Quick, secure airtime & data for all 4 Nigerian networks — MTN, Airtel, Glo, 9mobile',
                color: 'from-violet-500 to-purple-600',
                highlight: 'bg-purple-50',
              },
              {
                icon: Trophy,
                title: 'Daily Cash Draws',
                desc: 'Win up to ₦500,000 in our daily prize draws. Every recharge is an entry ticket!',
                color: 'from-yellow-500 to-orange-500',
                highlight: 'bg-orange-50',
              },
              {
                icon: Sparkles,
                title: 'Spin & Win Instantly',
                desc: 'Recharge ₦1,000+ to spin the prize wheel. Instant cash, data bundles, and more!',
                color: 'from-purple-500 to-purple-600',
                highlight: 'bg-purple-50',
              },
              {
                icon: Users,
                title: 'Refer & Earn',
                desc: 'Earn up to 5% commission on every recharge made by your referrals. Unlimited earnings!',
                color: 'from-green-500 to-teal-500',
                highlight: 'bg-green-50',
              },
            ].map(({ icon: Icon, title, desc, color, highlight }, i) => (
              <Card
                key={i}
                className={`${highlight} border-0 shadow-sm hover:shadow-xl transition-all duration-300 hover:-translate-y-1 cursor-default rounded-2xl overflow-hidden`}
              >
                <CardContent className="p-6">
                  <div className={`w-12 h-12 rounded-xl bg-gradient-to-br ${color} flex items-center justify-center text-white mb-4 shadow-md`}>
                    <Icon className="w-6 h-6" />
                  </div>
                  <h3 className="font-bold text-gray-900 text-lg mb-2">{title}</h3>
                  <p className="text-gray-500 text-sm leading-relaxed">{desc}</p>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* ── Recent Winners ───────────────────────────────────────── */}
      <section className="py-14 md:py-20 bg-white">
        <div className="max-w-screen-xl mx-auto px-4">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-10">
            <div>
              <Badge className="mb-2 bg-yellow-100 text-yellow-700 border-yellow-200 px-4 py-1 font-semibold">
                <Crown className="w-3.5 h-3.5 mr-1.5" />
                Recent Winners
              </Badge>
              <h2 className="text-3xl font-black text-gray-900">Real People. Real Prizes.</h2>
            </div>
            <Button
              variant="outline"
              onClick={() => window.location.href = '/draws'}
              className="border-gray-200 text-gray-600 hover:text-blue-600 hover:border-purple-200 font-medium"
            >
              All Winners <ArrowRight className="w-4 h-4 ml-2" />
            </Button>
          </div>

          {recentWinners.length > 0 ? (
            <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
              {recentWinners.map((w) => (
                <div
                  key={w.id}
                  className="flex items-center gap-4 p-4 rounded-2xl border border-gray-100 hover:border-yellow-200 hover:bg-yellow-50/50 transition-all"
                >
                  <div className="w-11 h-11 rounded-full bg-gradient-to-br from-yellow-400 to-orange-500 flex items-center justify-center flex-shrink-0 shadow">
                    <Trophy className="w-5 h-5 text-white" />
                  </div>
                  <div className="min-w-0">
                    <p className="font-semibold text-gray-900 truncate">{w.name}</p>
                    <p className="text-sm text-orange-600 font-medium truncate">{w.prize}</p>
                    <p className="text-xs text-gray-400 mt-0.5">{w.time} · {w.network}</p>
                  </div>
                  <div className="ml-auto text-right flex-shrink-0">
                    <p className="text-sm font-bold text-gray-900">{formatCurrency(w.amount)}</p>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            /* Placeholder cards when no data */
            <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
              {[
                { name: 'Adaeze W.', prize: '₦50,000 Cash Prize', network: 'MTN', time: '2 hours ago' },
                { name: 'Obinna N.', prize: '10GB Data Bundle', network: 'Airtel', time: '3 hours ago' },
                { name: 'Fatimah B.', prize: '₦25,000 Cash Prize', network: 'Glo', time: '5 hours ago' },
                { name: 'Tunde F.', prize: '₦100,000 Grand Prize', network: 'MTN', time: 'Yesterday' },
                { name: 'Ngozi A.', prize: '₦5,000 Instant Win', network: '9mobile', time: 'Yesterday' },
                { name: 'Chidi O.', prize: '20GB Data Bundle', network: 'Airtel', time: '2 days ago' },
              ].map((w, i) => (
                <div
                  key={i}
                  className="flex items-center gap-4 p-4 rounded-2xl border border-gray-100 hover:border-yellow-200 transition-all"
                >
                  <div className="w-11 h-11 rounded-full bg-gradient-to-br from-yellow-400 to-orange-500 flex items-center justify-center flex-shrink-0 shadow">
                    <Trophy className="w-5 h-5 text-white" />
                  </div>
                  <div className="min-w-0">
                    <p className="font-semibold text-gray-900 truncate">{w.name}</p>
                    <p className="text-sm text-orange-600 font-medium truncate">{w.prize}</p>
                    <p className="text-xs text-gray-400 mt-0.5">{w.time} · {w.network}</p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </section>

      {/* ── Active Draws ─────────────────────────────────────────── */}
      <section className="py-14 md:py-20 bg-gradient-to-br from-gray-50 to-blue-50">
        <div className="max-w-screen-xl mx-auto px-4">
          <div className="text-center mb-10">
            <Badge className="mb-3 bg-purple-100 text-purple-700 border-purple-200 px-4 py-1 font-semibold">
              <Target className="w-3.5 h-3.5 mr-1.5" />
              Prize Draws
            </Badge>
            <h2 className="text-3xl md:text-4xl font-black text-gray-900 mb-3">Active Prize Draws</h2>
            <p className="text-gray-500 text-lg">Join today's draws and win amazing prizes</p>
          </div>
          <DrawsList />
        </div>
      </section>

      {/* ── Benefits banner ──────────────────────────────────────── */}
      <section className="py-14 md:py-20 bg-white">
        <div className="max-w-screen-xl mx-auto px-4">
          <div className="grid lg:grid-cols-2 gap-12 items-center">
            <div>
              <h2 className="text-3xl md:text-4xl font-black text-gray-900 mb-6">
                Everything You Need<br />to Win
              </h2>
              <div className="space-y-3">
                {[
                  'Every ₦200 recharge = 1 draw entry',
                  'Recharge ₦1,000+ to unlock Spin Wheel',
                  'Daily ₦20 subscription for guaranteed entries',
                  'Earn points and climb loyalty tiers',
                  'Refer friends to earn commissions',
                  'Instant airtime & data delivery',
                  'Secure card & bank transfer payments',
                  'Real-time SMS delivery confirmation',
                ].map((item, i) => (
                  <div key={i} className="flex items-center gap-3">
                    <div className="w-5 h-5 rounded-full bg-green-100 flex items-center justify-center flex-shrink-0">
                      <CheckCircle className="w-3.5 h-3.5 text-green-600" />
                    </div>
                    <span className="text-gray-600 text-sm">{item}</span>
                  </div>
                ))}
              </div>
              <div className="mt-8 inline-flex items-center gap-3 bg-green-700 text-white px-6 py-4 rounded-2xl shadow-lg">
                <div className="text-2xl">🇳🇬</div>
                <div>
                  <p className="font-bold text-sm">Proudly Nigerian</p>
                  <p className="text-green-200 text-xs">Built for Nigerian mobile users</p>
                </div>
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              {[
                { icon: DollarSign, title: 'Daily Prize Pool', value: formatCurrency(stats.activeDraw?.prizePool || 500000), color: 'from-yellow-400 to-orange-400', textColor: 'text-orange-600', bg: 'bg-orange-50', border: 'border-orange-200' },
                { icon: Users, title: 'Active Users', value: `${stats.totalUsers > 0 ? stats.totalUsers.toLocaleString() : '10,000'}+`, color: 'from-violet-500 to-purple-600', textColor: 'text-blue-600', bg: 'bg-purple-50', border: 'border-purple-200' },
                { icon: Award, title: 'Prizes Given', value: `${stats.totalPrizes > 0 ? stats.totalPrizes.toLocaleString() : '500'}+`, color: 'from-green-500 to-teal-500', textColor: 'text-green-600', bg: 'bg-green-50', border: 'border-green-200' },
                { icon: Shield, title: 'Security Level', value: 'Enterprise', color: 'from-purple-500 to-purple-600', textColor: 'text-purple-600', bg: 'bg-purple-50', border: 'border-purple-200' },
              ].map(({ icon: Icon, title, value, color, textColor, bg, border }) => (
                <Card key={title} className={`${bg} border ${border} shadow-sm rounded-2xl`}>
                  <CardContent className="p-5">
                    <div className={`w-10 h-10 rounded-xl bg-gradient-to-br ${color} flex items-center justify-center text-white mb-3 shadow-sm`}>
                      <Icon className="w-5 h-5" />
                    </div>
                    <p className="text-xs text-gray-500 font-medium mb-1">{title}</p>
                    <p className={`text-2xl font-black ${textColor}`}>{value}</p>
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>
        </div>
      </section>

      {/* ── Final CTA ────────────────────────────────────────────── */}
      <section className="py-16 md:py-24 bg-gradient-to-br from-[#1a0b3b] via-[#3b0764] to-[#7c3aed] text-white relative overflow-hidden">
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,_rgba(255,255,255,0.05)_0%,_transparent_60%)]" />
        <div className="relative max-w-screen-xl mx-auto px-4 text-center space-y-7">
          <Badge className="bg-white/20 text-white border-white/30 text-sm px-4 py-1 font-semibold backdrop-blur-sm">
            <Flame className="w-3.5 h-3.5 mr-1.5 text-orange-300" />
            Join Thousands of Winners
          </Badge>
          <h2 className="text-4xl md:text-5xl font-black leading-tight max-w-2xl mx-auto">
            Start Winning With Every Recharge Today!
          </h2>
          <p className="text-xl text-blue-100 max-w-xl mx-auto">
            No complicated sign-ups. Just recharge, win, and collect your prizes.
          </p>
          <div className="flex flex-wrap gap-4 justify-center">
            <Button
              size="lg"
              onClick={scrollToRecharge}
              className="bg-white text-purple-700 hover:bg-yellow-50 font-bold px-10 py-4 text-lg shadow-2xl hover:shadow-3xl hover:scale-105 transition-all rounded-xl"
            >
              <Smartphone className="w-5 h-5 mr-2" />
              Recharge Now
            </Button>
            {!isAuthenticated && (
              <Button
                size="lg"
                variant="outline"
                onClick={() => window.location.href = '/login'}
                className="border-white/40 text-white hover:bg-white/10 font-semibold px-10 py-4 text-lg rounded-xl backdrop-blur-sm"
              >
                <Users className="w-5 h-5 mr-2" />
                Create Account
              </Button>
            )}
          </div>
          <p className="text-blue-300 text-sm">
            🔒 Secured by SSL · PCI DSS compliant · Instant delivery guaranteed
          </p>
        </div>
      </section>

      {/* ── Footer ───────────────────────────────────────────────── */}
      <footer className="bg-gray-900 text-gray-400 py-10">
        <div className="max-w-screen-xl mx-auto px-4">
          <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-8 mb-8">
            <div>
              <div className="flex items-center gap-2.5 mb-3">
                <div className="w-8 h-8 rounded-lg bg-gradient-to-br gradient-brand flex items-center justify-center">
                  <Zap className="w-4 h-4 text-white" />
                </div>
                <span className="font-bold text-white text-base">RechargeMax</span>
              </div>
              <p className="text-xs leading-relaxed">Nigeria's leading gamified mobile recharge platform. Recharge, earn, and win!</p>
            </div>
            <div>
              <p className="text-white font-semibold mb-3 text-sm">Services</p>
              <ul className="space-y-2 text-xs">
                {['Airtime Recharge', 'Data Bundles', 'Daily Draws', 'Spin Wheel', 'Daily Subscription'].map((s) => (
                  <li key={s} className="hover:text-purple-400 cursor-pointer transition-colors">{s}</li>
                ))}
              </ul>
            </div>
            <div>
              <p className="text-white font-semibold mb-3 text-sm">Networks</p>
              <ul className="space-y-2 text-xs">
                {['MTN Nigeria', 'Airtel Nigeria', 'Glo Mobile', '9mobile'].map((n) => (
                  <li key={n} className="hover:text-purple-400 cursor-pointer transition-colors">{n}</li>
                ))}
              </ul>
            </div>
            <div>
              <p className="text-white font-semibold mb-3 text-sm">Company</p>
              <ul className="space-y-2 text-xs">
                {['About Us', 'Contact', 'Privacy Policy', 'Terms of Service', 'Affiliate Program'].map((c) => (
                  <li key={c} className="hover:text-purple-400 cursor-pointer transition-colors">{c}</li>
                ))}
              </ul>
            </div>
          </div>
          <div className="border-t border-gray-800 pt-6 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs">
            <p>© 2026 RechargeMax. All rights reserved. 🇳🇬 Made in Nigeria.</p>
            <p className="text-gray-500">Secure · Fast · Rewarding</p>
          </div>
        </div>
      </footer>

      {/* ── Spin Wheel Modal ─────────────────────────────────────── */}
      {showSpinWheel && availableSpins > 0 && (
        <SpinWheel
          isOpen={showSpinWheel}
          onClose={async () => {
            setShowSpinWheel(false);
            setAvailableSpins(0);
            // Re-check eligibility after the user closes the wheel normally.
            // If all spins are exhausted show the upgrade nudge after a short delay.
            setTimeout(() => triggerUpgradeNudgeIfExhausted(), 400);
          }}
          transactionAmount={rechargeSuccess?.amount || 1000}
          userPhone={userPhone || ''}
          onPrizeWon={async () => {
            // Silently refresh remaining spin count while the user reads their prize.
            try {
              const res = await apiClient.get('/spin/eligibility');
              setAvailableSpins(res.data?.data?.available_spins ?? 0);
            } catch { /* ignore — onClose will re-check */ }
          }}
          onSpinLimitReached={() => {
            // Backend rejected the spin because the frontend had stale state.
            // The SpinWheel already called handleClose() + showed a brief toast.
            // We just need to show the upgrade nudge with fresh data.
            setShowSpinWheel(false);
            setAvailableSpins(0);
            triggerUpgradeNudgeIfExhausted();
          }}
        />
      )}

      {/* ── Spin Upgrade Nudge ───────────────────────────────────── */}
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

export default EnterpriseHomePage;
