import React, { useState, useEffect, useRef } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Button } from '@/components/ui/button';
import { formatCurrency } from '@/lib/utils';
import { useToast } from '@/hooks/useToast';
import { Gift, Zap, RotateCcw, Loader2, X, Sparkles, Trophy } from 'lucide-react';
import apiClient from '@/lib/api-client';
import { useAuthContext } from '@/contexts/AuthContext';
import confetti from 'canvas-confetti';

/* ── Fallback prizes (used when /spin/prizes is unreachable) ── */
const FALLBACK_PRIZES = [
  { name: '₦100 Airtime',  type: 'AIRTIME', value: 10000,  probability: 25, color: '#10b981' },
  { name: '₦200 Airtime',  type: 'AIRTIME', value: 20000,  probability: 20, color: '#3b82f6' },
  { name: '500MB Data',    type: 'DATA',    value: 50000,  probability: 15, color: '#8b5cf6' },
  { name: '1GB Data',      type: 'DATA',    value: 100000, probability: 15, color: '#f59e0b' },
  { name: '₦100 Cash',     type: 'CASH',    value: 10000,  probability: 10, color: '#ef4444' },
  { name: '₦200 Cash',     type: 'CASH',    value: 20000,  probability: 8,  color: '#ec4899' },
  { name: '₦500 Cash',     type: 'CASH',    value: 50000,  probability: 5,  color: '#fbbf24' },
  { name: '₦1000 Cash',    type: 'CASH',    value: 100000, probability: 2,  color: '#a855f7' },
];

interface WheelPrize {
  name: string;
  type: string;
  value: number;
  probability: number;
  color: string;
  is_no_win?: boolean;
  no_win_message?: string;
}

interface SpinWheelProps {
  isOpen: boolean;
  onClose: () => void;
  transactionAmount: number;
  userPhone: string;
  onPrizeWon?: (prize: any) => void;
  /** Called when the backend rejects the spin with a daily-limit error.
   *  The parent should close the wheel and show the upgrade nudge. */
  onSpinLimitReached?: () => void;
}

/* ── Fire confetti from both corners ── */
function fireWinConfetti() {
  const defaults = { startVelocity: 30, spread: 360, ticks: 60, zIndex: 9999 };
  const count = 200;
  const origin1 = { x: 0.2, y: 0.5 };
  const origin2 = { x: 0.8, y: 0.5 };
  confetti({ ...defaults, particleCount: count * 0.4, origin: origin1,
    colors: ['#7c3aed', '#a855f7', '#f59e0b', '#fbbf24', '#10b981'] });
  confetti({ ...defaults, particleCount: count * 0.4, origin: origin2,
    colors: ['#7c3aed', '#c084fc', '#f59e0b', '#fcd34d', '#ec4899'] });
  setTimeout(() => {
    confetti({ ...defaults, particleCount: count * 0.2, origin: { x: 0.5, y: 0.3 },
      colors: ['#ffffff', '#f59e0b', '#7c3aed'] });
  }, 300);
}

export const SpinWheel: React.FC<SpinWheelProps> = ({
  isOpen, onClose, transactionAmount, userPhone, onPrizeWon, onSpinLimitReached,
}) => {
  const { toast } = useToast();
  const { isAuthenticated } = useAuthContext();
  const [prizes, setPrizes] = useState<WheelPrize[]>(FALLBACK_PRIZES);
  const [loadingPrizes, setLoadingPrizes] = useState(true);
  const [isSpinning, setIsSpinning] = useState(false);
  const [rotation, setRotation] = useState(0);
  const [selectedPrize, setSelectedPrize] = useState<any>(null);
  const [hasSpun, setHasSpun] = useState(false);
  const [showWin, setShowWin] = useState(false);
  const [noWinResult, setNoWinResult] = useState<{ message: string } | null>(null);
  const wheelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!isOpen) return;
    setLoadingPrizes(true);
    apiClient.get('/spin/prizes')
      .then((res) => {
        const raw: any[] = res.data?.data ?? [];
        if (raw.length > 0) {
          const mapped: WheelPrize[] = raw
            .filter((p) => p.is_active !== false)
            .map((p) => ({
              name:          p.prize_name ?? p.name ?? 'Prize',
              type:          (p.prize_type ?? p.type ?? 'AIRTIME').toUpperCase(),
              value:         Number(p.prize_value ?? p.value ?? 0),
              probability:   Number(p.probability ?? 0),
              color:         p.is_no_win ? '#4b5563' : (p.color_scheme ?? p.color ?? '#6b7280'),
              is_no_win:     p.is_no_win ?? false,
              no_win_message: p.no_win_message ?? '',
            }));
          if (mapped.length > 0) setPrizes(mapped);
        }
      })
      .catch(() => {})
      .finally(() => setLoadingPrizes(false));
  }, [isOpen]);

  const segmentAngle = 360 / prizes.length;

  const spinWheel = async () => {
    if (isSpinning || hasSpun) return;
    setIsSpinning(true);
    setShowWin(false);

    try {
      const spinBody = isAuthenticated ? {} : { msisdn: userPhone };
      const response = await apiClient.post('/spin/play', spinBody);
      if (!response.data.success) throw new Error(response.data.error || 'Failed to spin');

      const spinResult = response.data.data;
      const winningPrize: WheelPrize =
        prizes.find((p) => p.type === spinResult.prize_type && p.value === spinResult.prize_value) ??
        prizes.find((p) => p.name === spinResult.prize_won) ??
        prizes[0] ?? { name: spinResult.prize_won ?? 'Prize', type: spinResult.prize_type ?? 'AIRTIME', value: 0, probability: 0, color: '#6b7280' };

      const prizeIndex = prizes.findIndex((p) => p.name === winningPrize.name);
      const targetAngle = prizeIndex * segmentAngle + segmentAngle / 2;
      const spins = 6 + Math.random() * 2;
      const finalRotation = spins * 360 + (360 - targetAngle);
      setRotation((prev) => prev + finalRotation);

      setTimeout(() => {
        setIsSpinning(false);
        setHasSpun(true);

        // ── No-win slot ──────────────────────────────────────────────────────
        if (spinResult.no_win) {
          setNoWinResult({
            message: spinResult.no_win_message || 'Better luck next time! Recharge again to spin.',
          });
          toast({
            title: 'Not this time…',
            description: spinResult.no_win_message || 'Keep recharging for more spin chances!',
            duration: 6000,
          });
          return;
        }

        // ── Prize won ────────────────────────────────────────────────────────
        setSelectedPrize({ ...winningPrize, claimStatus: spinResult.claim_status });
        setShowWin(true);
        fireWinConfetti();

        const claimInstructions =
          winningPrize.type === 'AIRTIME' || winningPrize.type === 'DATA'
            ? spinResult.claim_status === 'PROVISIONING'
              ? 'Being processed — credited within 5–10 min.'
              : 'Check Dashboard → Prizes for claim status.'
            : winningPrize.type === 'CASH'
            ? 'Go to Dashboard → Prizes to submit bank details.'
            : 'Login to see your account update.';

        toast({
          title: '🎉 You Won!',
          description: `${winningPrize.name}! ${claimInstructions}`,
          duration: 10000,
        });
        onPrizeWon?.(winningPrize);
      }, 4500);
    } catch (error: any) {
      setIsSpinning(false);
      const errMsg: string =
        error.response?.data?.error?.message ??
        error.response?.data?.message ??
        error.message ??
        'Failed to spin. Please try again.';

      // Detect daily-limit errors — close the wheel and show the upgrade nudge
      // instead of leaving the user staring at an unusable wheel.
      const isLimitError =
        errMsg.toLowerCase().includes('daily spin limit') ||
        errMsg.toLowerCase().includes('not eligible') ||
        errMsg.toLowerCase().includes('no spins') ||
        errMsg.toLowerCase().includes('limit reached') ||
        error.response?.status === 429;

      if (isLimitError && onSpinLimitReached) {
        // Small delay so the user can read the brief toast before the modal swaps
        toast({ title: 'Spins Used Up', description: errMsg, variant: 'destructive', duration: 3000 });
        setTimeout(() => {
          handleClose();
          onSpinLimitReached();
        }, 1200);
      } else {
        toast({ title: 'Spin Failed', description: errMsg, variant: 'destructive', duration: 5000 });
      }
    }
  };

  const handleClose = () => {
    setRotation(0);
    setSelectedPrize(null);
    setHasSpun(false);
    setShowWin(false);
    setNoWinResult(null);
    onClose();
  };

  return (
    <AnimatePresence>
      {isOpen && (
        <motion.div
          className="fixed inset-0 z-50 flex items-center justify-center p-4"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
        >
          {/* Backdrop */}
          <motion.div
            className="absolute inset-0"
            style={{ background: 'radial-gradient(ellipse at center, rgba(59,7,100,0.97) 0%, rgba(10,5,20,0.98) 100%)' }}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={handleClose}
          />

          {/* Modal */}
          <motion.div
            className="relative w-full max-w-md rounded-3xl overflow-hidden shadow-2xl"
            style={{ background: 'linear-gradient(160deg, #1a0b3b 0%, #0f0520 60%, #1a0b3b 100%)', border: '1px solid rgba(124,58,237,0.3)' }}
            initial={{ scale: 0.85, y: 40, opacity: 0 }}
            animate={{ scale: 1, y: 0, opacity: 1 }}
            exit={{ scale: 0.85, y: 40, opacity: 0 }}
            transition={{ type: 'spring', damping: 22, stiffness: 250 }}
          >
            {/* Glow top accent */}
            <div className="absolute top-0 left-1/2 -translate-x-1/2 w-3/4 h-1 rounded-full" style={{ background: 'linear-gradient(90deg, transparent, #7c3aed, #f59e0b, #7c3aed, transparent)' }} />

            {/* Close button */}
            <motion.button
              onClick={handleClose}
              className="absolute top-4 right-4 w-8 h-8 rounded-full flex items-center justify-center text-white/60 hover:text-white hover:bg-white/10 transition-colors z-10"
              whileHover={{ scale: 1.1 }}
              whileTap={{ scale: 0.9 }}
            >
              <X className="w-4 h-4" />
            </motion.button>

            <div className="p-6 space-y-5">
              {/* Header */}
              <div className="text-center space-y-1">
                <motion.div
                  className="flex items-center justify-center gap-2"
                  initial={{ y: -10, opacity: 0 }}
                  animate={{ y: 0, opacity: 1 }}
                  transition={{ delay: 0.1 }}
                >
                  <Sparkles className="w-5 h-5 text-yellow-400" />
                  <h2 className="text-2xl font-extrabold text-white" style={{ fontFamily: "'Plus Jakarta Sans', sans-serif" }}>
                    Spin the Wheel!
                  </h2>
                  <Sparkles className="w-5 h-5 text-yellow-400" />
                </motion.div>
                <p className="text-purple-300 text-sm">
                  You've unlocked a free spin for recharging {formatCurrency(transactionAmount)}
                </p>
              </div>

              {/* Wheel */}
              {loadingPrizes ? (
                <div className="flex justify-center items-center h-72">
                  <div className="text-center space-y-3">
                    <Loader2 className="w-10 h-10 animate-spin text-purple-400 mx-auto" />
                    <p className="text-purple-300 text-sm">Loading prizes…</p>
                  </div>
                </div>
              ) : (
                <div className="relative mx-auto w-72 h-72">
                  {/* Outer glow ring */}
                  <div className="absolute inset-0 rounded-full wheel-glow opacity-60" />

                  {/* Pointer */}
                  <div className="absolute -top-3 left-1/2 -translate-x-1/2 z-10">
                    <motion.div
                      animate={isSpinning ? { scale: [1, 1.3, 1], y: [0, -3, 0] } : {}}
                      transition={{ duration: 0.3, repeat: isSpinning ? Infinity : 0 }}
                    >
                      <div className="w-0 h-0 border-l-[10px] border-r-[10px] border-b-[20px] border-l-transparent border-r-transparent border-b-yellow-400 drop-shadow-lg" />
                    </motion.div>
                  </div>

                  {/* Spinning disc */}
                  <div
                    ref={wheelRef}
                    className="w-full h-full rounded-full relative overflow-hidden"
                    style={{
                      transform: `rotate(${rotation}deg)`,
                      transition: isSpinning ? `transform 4500ms cubic-bezier(0.17, 0.67, 0.12, 0.99)` : 'none',
                      border: '4px solid rgba(124,58,237,0.6)',
                      boxShadow: isSpinning
                        ? '0 0 40px rgba(124,58,237,0.6), 0 0 80px rgba(245,158,11,0.3)'
                        : '0 0 20px rgba(124,58,237,0.3)',
                      background: `conic-gradient(${prizes.map((p, i) =>
                        `${p.color} ${i * segmentAngle}deg ${(i + 1) * segmentAngle}deg`
                      ).join(', ')})`,
                    }}
                  >
                    {/* Segment labels */}
                    {prizes.map((prize, index) => {
                      const angle = index * segmentAngle + segmentAngle / 2;
                      const radian = (angle * Math.PI) / 180;
                      const r = 100;
                      const x = Math.cos(radian) * r;
                      const y = Math.sin(radian) * r;
                      return (
                        <div
                          key={prize.name}
                          className="absolute font-bold text-center leading-tight"
                          style={{
                            left: `calc(50% + ${x}px - 32px)`,
                            top: `calc(50% + ${y}px - 16px)`,
                            width: '64px',
                            transform: `rotate(${angle}deg)`,
                            textShadow: '1px 1px 3px rgba(0,0,0,0.9)',
                            fontSize: '9px',
                            color: '#ffffff',
                            lineHeight: '1.2',
                          }}
                        >
                          {prize.name.split(' ').map((word, i) => (
                            <div key={i}>{word}</div>
                          ))}
                        </div>
                      );
                    })}
                  </div>

                  {/* Centre spin button */}
                  <div className="absolute inset-0 flex items-center justify-center">
                    <motion.button
                      onClick={spinWheel}
                      disabled={isSpinning || hasSpun}
                      className="w-[72px] h-[72px] rounded-full flex items-center justify-center font-black text-base z-10 disabled:opacity-60"
                      style={{
                        background: 'white',
                        color: '#7c3aed',
                        boxShadow: '0 0 0 4px rgba(124,58,237,0.4), 0 4px 20px rgba(0,0,0,0.4)',
                        border: '3px solid rgba(124,58,237,0.8)',
                      }}
                      whileHover={!isSpinning && !hasSpun ? { scale: 1.1 } : {}}
                      whileTap={!isSpinning && !hasSpun ? { scale: 0.95 } : {}}
                    >
                      {isSpinning
                        ? <RotateCcw className="w-7 h-7 animate-spin text-purple-700" />
                        : <span style={{ fontFamily: "'Plus Jakarta Sans', sans-serif" }}>SPIN</span>
                      }
                    </motion.button>
                  </div>
                </div>
              )}

              {/* Win result */}
              <AnimatePresence>
                {showWin && selectedPrize && (
                  <motion.div
                    initial={{ scale: 0.7, opacity: 0, y: 20 }}
                    animate={{ scale: 1, opacity: 1, y: 0 }}
                    transition={{ type: 'spring', damping: 18, stiffness: 260 }}
                    className="rounded-2xl p-5 text-center space-y-3 animate-prize-glow"
                    style={{ background: 'linear-gradient(135deg, rgba(124,58,237,0.3), rgba(245,158,11,0.2))', border: '1px solid rgba(245,158,11,0.4)' }}
                  >
                    <motion.div
                      animate={{ rotate: [0, -10, 10, -5, 5, 0], scale: [1, 1.2, 1] }}
                      transition={{ duration: 0.6, delay: 0.2 }}
                    >
                      <Trophy className="w-10 h-10 text-yellow-400 mx-auto" />
                    </motion.div>
                    <div>
                      <p className="text-yellow-300 text-sm font-semibold uppercase tracking-wider">🎉 Congratulations!</p>
                      <p className="text-white text-2xl font-black mt-1" style={{ fontFamily: "'Plus Jakarta Sans', sans-serif" }}>
                        {selectedPrize.name}
                      </p>
                    </div>
                    <p className="text-purple-300 text-xs leading-relaxed">
                      {selectedPrize.type === 'CASH'
                        ? 'Go to Dashboard → Prizes to submit your bank details.'
                        : selectedPrize.claimStatus === 'PROVISIONING'
                        ? 'Being credited to your phone within 5–10 minutes.'
                        : 'Login and go to Dashboard → Prizes to check status.'}
                    </p>
                  </motion.div>
                )}
              </AnimatePresence>

              {/* No-win result panel */}
              <AnimatePresence>
                {noWinResult && (
                  <motion.div
                    initial={{ scale: 0.85, opacity: 0, y: 20 }}
                    animate={{ scale: 1, opacity: 1, y: 0 }}
                    transition={{ type: 'spring', damping: 20, stiffness: 250 }}
                    className="rounded-2xl p-5 text-center space-y-3"
                    style={{ background: 'linear-gradient(135deg, rgba(55,65,81,0.6), rgba(30,20,60,0.7))', border: '1px solid rgba(255,255,255,0.1)' }}
                  >
                    <motion.div
                      animate={{ rotate: [0, -15, 15, -8, 8, 0] }}
                      transition={{ duration: 0.7, delay: 0.1 }}
                    >
                      <RotateCcw className="w-10 h-10 text-purple-400 mx-auto" />
                    </motion.div>
                    <div>
                      <p className="text-white/60 text-sm font-semibold uppercase tracking-wider">Not this time</p>
                      <p className="text-white text-lg font-bold mt-1" style={{ fontFamily: "'Plus Jakarta Sans', sans-serif" }}>
                        {noWinResult.message}
                      </p>
                    </div>
                    <p className="text-purple-300 text-xs">
                      Recharge again to earn another free spin!
                    </p>
                  </motion.div>
                )}
              </AnimatePresence>

              {/* Action buttons */}
              <div className="space-y-2">
                {!hasSpun ? (
                  <div className="flex gap-3">
                    <motion.button
                      onClick={spinWheel}
                      disabled={isSpinning || loadingPrizes}
                      className="flex-1 py-3.5 rounded-2xl font-bold text-white btn-claim disabled:opacity-50 flex items-center justify-center gap-2"
                      whileHover={{ scale: 1.02 }}
                      whileTap={{ scale: 0.98 }}
                    >
                      {isSpinning
                        ? <><Loader2 className="w-4 h-4 animate-spin" /> Spinning…</>
                        : <><Zap className="w-4 h-4" /> Spin Now!</>
                      }
                    </motion.button>
                    <motion.button
                      onClick={handleClose}
                      className="px-5 py-3.5 rounded-2xl font-semibold text-white/60 hover:text-white border border-white/10 hover:border-white/20 transition-colors"
                      whileHover={{ scale: 1.02 }}
                      whileTap={{ scale: 0.98 }}
                    >
                      Skip
                    </motion.button>
                  </div>
                ) : noWinResult ? (
                  /* No-win CTA buttons */
                  <div className="space-y-2">
                    <motion.button
                      onClick={handleClose}
                      className="w-full py-3.5 rounded-2xl font-bold text-white flex items-center justify-center gap-2"
                      style={{ background: 'linear-gradient(135deg, #7c3aed, #6d28d9)' }}
                      whileHover={{ scale: 1.02 }}
                      whileTap={{ scale: 0.98 }}
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: 0.3 }}
                    >
                      <Zap className="w-4 h-4" /> Recharge to Spin Again
                    </motion.button>
                    <motion.button
                      onClick={handleClose}
                      className="w-full py-2.5 rounded-2xl font-medium text-white/50 hover:text-white/70 transition-colors text-sm"
                      whileTap={{ scale: 0.98 }}
                    >
                      Maybe Later
                    </motion.button>
                  </div>
                ) : (
                  /* Prize won CTA buttons */
                  <div className="space-y-2">
                    <motion.button
                      onClick={() => window.location.href = '/login'}
                      className="w-full py-3.5 rounded-2xl font-bold text-white btn-claim flex items-center justify-center gap-2"
                      whileHover={{ scale: 1.02 }}
                      whileTap={{ scale: 0.98 }}
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: 0.3 }}
                    >
                      <Gift className="w-4 h-4" /> Login to Claim Prize
                    </motion.button>
                    <motion.button
                      onClick={handleClose}
                      className="w-full py-2.5 rounded-2xl font-medium text-white/50 hover:text-white/70 transition-colors text-sm"
                      whileTap={{ scale: 0.98 }}
                    >
                      Close (Claim Later)
                    </motion.button>
                  </div>
                )}
              </div>

              {/* Prize list */}
              {!showWin && !noWinResult && (
                <div className="border-t border-white/10 pt-4">
                  <p className="text-center text-xs text-purple-300 font-semibold uppercase tracking-wider mb-3">Possible Prizes</p>
                  <div className="grid grid-cols-2 gap-2">
                    {prizes.map((prize) => (
                      <div key={prize.name} className="flex items-center gap-2 text-xs text-white/70">
                        <div className="w-2.5 h-2.5 rounded-full flex-shrink-0" style={{ backgroundColor: prize.color }} />
                        {(prize as any).is_no_win
                          ? <span className="text-white/40 italic">{prize.name}</span>
                          : prize.name
                        }
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
};
