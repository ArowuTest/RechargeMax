/**
 * SpinUpgradeNudge
 *
 * Shown INSTEAD of the spin wheel when a user has exhausted all their
 * spins for today's recharge(s). It tells them:
 *   - How many spins they earned and used today
 *   - Exactly how much to recharge next to unlock the next tier
 *   - A clear CTA to recharge again
 *
 * This replaces the current bad UX of: wheel popup → hit SPIN → "Spin Failed" toast.
 */
import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { formatCurrency } from '@/lib/utils';
import { Zap, Trophy, ArrowRight, X, RotateCcw } from 'lucide-react';

interface SpinUpgradeNudgeProps {
  isOpen: boolean;
  onClose: () => void;
  /** Max daily spins allowed at current cumulative tier */
  spinsGranted: number;
  /** Spins already played today */
  spinsUsed: number;
  /** Next tier name, e.g. "Gold" */
  nextTierName?: string;
  /** Absolute min cumulative amount (kobo) to reach next tier */
  nextTierMinAmount?: number;
  /** How much MORE to recharge today (kobo) to cross into next tier */
  amountToNextTier?: number;
  /** Number of spins the next tier allows per day */
  nextTierSpins?: number;
}

const TIER_COLORS: Record<string, string> = {
  Bronze:   'from-amber-600  to-yellow-500',
  Silver:   'from-slate-400  to-gray-300',
  Gold:     'from-yellow-500 to-amber-400',
  Platinum: 'from-purple-500 to-indigo-400',
};

const TIER_EMOJI: Record<string, string> = {
  Bronze:   '🥉',
  Silver:   '🥈',
  Gold:     '🥇',
  Platinum: '💎',
};

export const SpinUpgradeNudge: React.FC<SpinUpgradeNudgeProps> = ({
  isOpen,
  onClose,
  spinsGranted,
  spinsUsed,
  nextTierName,
  nextTierMinAmount,
  amountToNextTier,
  nextTierSpins,
}) => {
  if (!isOpen) return null;

  const hasNextTier    = !!nextTierName && !!nextTierSpins;
  const tierGradient   = nextTierName ? (TIER_COLORS[nextTierName] ?? 'from-green-500 to-emerald-400') : '';
  const tierEmoji      = nextTierName ? (TIER_EMOJI[nextTierName] ?? '🏆') : '🏆';
  // "Recharge X MORE today" is more actionable than the absolute minimum
  const nudgeAmountNaira = amountToNextTier
    ? amountToNextTier / 100
    : nextTierMinAmount
      ? nextTierMinAmount / 100
      : 0;

  return (
    <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4">
      <Card className="w-full max-w-md shadow-2xl border-0 overflow-hidden">

        {/* Header — spinning coins animation feel */}
        <CardHeader className="bg-gradient-to-br from-gray-900 to-gray-800 text-white pb-6 relative">
          <button
            onClick={onClose}
            className="absolute top-3 right-3 text-gray-400 hover:text-white transition-colors"
          >
            <X className="w-5 h-5" />
          </button>

          <div className="flex flex-col items-center gap-3 pt-2">
            <div className="relative">
              <div className="w-16 h-16 rounded-full bg-gradient-to-br from-orange-400 to-pink-500 flex items-center justify-center shadow-lg">
                <RotateCcw className="w-8 h-8 text-white" />
              </div>
              {/* Checkmark badge */}
              <div className="absolute -bottom-1 -right-1 w-6 h-6 bg-green-500 rounded-full flex items-center justify-center text-white text-xs font-bold shadow">
                ✓
              </div>
            </div>

            <CardTitle className="text-center text-xl font-bold">
              All Spins Used Today!
            </CardTitle>

            <p className="text-gray-300 text-sm text-center">
              You used all <span className="font-bold text-white">{spinsUsed}/{spinsGranted}</span> spin{spinsGranted !== 1 ? 's' : ''} from today's recharge{spinsGranted !== 1 ? 's' : ''}.
            </p>
          </div>
        </CardHeader>

        <CardContent className="p-5 space-y-5 bg-white">

          {/* Today's spin summary */}
          <div className="flex items-center justify-between bg-gray-50 rounded-xl p-4 border border-gray-100">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-green-100 flex items-center justify-center">
                <Trophy className="w-5 h-5 text-green-600" />
              </div>
              <div>
                <p className="text-sm font-semibold text-gray-900">Today's spins</p>
                <p className="text-xs text-gray-500">Earned from your recharge(s)</p>
              </div>
            </div>
            <div className="text-right">
              <p className="text-2xl font-bold text-gray-900">{spinsUsed}/{spinsGranted}</p>
              <Badge variant="outline" className="text-xs text-green-600 border-green-300 bg-green-50">
                All claimed
              </Badge>
            </div>
          </div>

          {/* Upgrade nudge */}
          {hasNextTier ? (
            <div className={`rounded-xl bg-gradient-to-br ${tierGradient} p-[2px] shadow-lg`}>
                <div className="rounded-[10px] bg-white p-4 space-y-3">
                <div className="flex items-center gap-2">
                  <span className="text-2xl">{tierEmoji}</span>
                  <div>
                    <p className="font-bold text-gray-900 text-sm">
                      Unlock {nextTierSpins} spin{nextTierSpins! > 1 ? 's' : ''}/day with {nextTierName} tier
                    </p>
                    <p className="text-xs text-gray-500">Based on your cumulative recharges today</p>
                  </div>
                </div>

                <div className="bg-gradient-to-r from-gray-50 to-gray-100 rounded-lg p-3 flex items-center justify-between">
                  <div>
                    <p className="text-xs text-gray-500 mb-0.5">
                      {amountToNextTier ? 'Recharge this much MORE today' : 'Minimum cumulative today'}
                    </p>
                    <p className="text-2xl font-extrabold text-gray-900">
                      {formatCurrency(nudgeAmountNaira)}
                    </p>
                  </div>
                  <div className="text-right">
                    <p className="text-xs text-gray-500 mb-0.5">Daily cap rises to</p>
                    <p className="text-lg font-bold text-green-600">
                      {nextTierSpins} spin{nextTierSpins! > 1 ? 's' : ''}/day
                    </p>
                  </div>
                </div>

                <div className="flex items-start gap-2 text-xs text-gray-600">
                  <Zap className="w-3.5 h-3.5 text-yellow-500 mt-0.5 flex-shrink-0" />
                  <p>
                    {amountToNextTier
                      ? <>Top up <span className="font-semibold">{formatCurrency(nudgeAmountNaira)}+</span> more today and your daily spin limit jumps to <span className="font-semibold">{nextTierSpins} {nextTierName} spin{nextTierSpins! > 1 ? 's' : ''}</span>!</>
                      : <>Reach a cumulative daily total of <span className="font-semibold">{formatCurrency(nudgeAmountNaira)}</span> to unlock {nextTierSpins} spins/day.</>
                    }
                  </p>
                </div>
              </div>
            </div>
          ) : (
            /* Already at max tier */
            <div className="rounded-xl bg-gradient-to-br from-purple-500 to-indigo-400 p-[2px] shadow-lg">
              <div className="rounded-[10px] bg-white p-4 text-center space-y-2">
                <span className="text-3xl">💎</span>
                <p className="font-bold text-gray-900">You're at the top tier!</p>
                <p className="text-sm text-gray-500">
                  Come back tomorrow for a fresh batch of spins, or recharge again today for bonus spins.
                </p>
              </div>
            </div>
          )}

          {/* Recharge tier ladder (quick reference) */}
          <div className="space-y-1.5">
            <p className="text-xs font-semibold text-gray-500 uppercase tracking-wide">Spin tiers</p>
            {[
              { name: 'Bronze',   range: '₦1,000 – ₦2,499',  spins: 1, emoji: '🥉' },
              { name: 'Silver',   range: '₦2,500 – ₦4,999',  spins: 2, emoji: '🥈' },
              { name: 'Gold',     range: '₦5,000 – ₦9,999',  spins: 3, emoji: '🥇' },
              { name: 'Platinum', range: '₦10,000+',          spins: 5, emoji: '💎' },
            ].map((t) => (
              <div
                key={t.name}
                className={`flex items-center justify-between px-3 py-1.5 rounded-lg text-xs ${
                  t.name === nextTierName
                    ? 'bg-yellow-50 border border-yellow-300 font-semibold'
                    : 'bg-gray-50 text-gray-600'
                }`}
              >
                <span>{t.emoji} {t.name} <span className="text-gray-400 font-normal">({t.range})</span></span>
                <span className="font-bold text-gray-700">{t.spins} spin{t.spins > 1 ? 's' : ''}</span>
              </div>
            ))}
          </div>

          {/* Action buttons */}
          <div className="flex gap-3 pt-1">
            <Button
              className="flex-1 bg-gradient-to-r from-green-500 to-emerald-600 hover:from-green-600 hover:to-emerald-700 text-white font-bold"
              onClick={onClose}
            >
              <Zap className="w-4 h-4 mr-1.5" />
              Recharge Now
              <ArrowRight className="w-4 h-4 ml-1.5" />
            </Button>
            <Button variant="outline" onClick={onClose} className="px-4">
              Close
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};
