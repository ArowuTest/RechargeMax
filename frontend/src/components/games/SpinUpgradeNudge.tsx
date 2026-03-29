/**
 * SpinUpgradeNudge
 *
 * Shown INSTEAD of the spin wheel when a user has exhausted all their
 * spins for today's recharge(s). It tells them:
 *   - How many spins they earned and used today
 *   - Exactly how much to recharge next to unlock the next tier
 *   - A clear CTA to recharge again
 *
 * Tier table is fetched live from GET /spins/tiers so admin changes
 * (amounts, spins, names) are reflected instantly without a frontend deploy.
 */
import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { formatCurrency } from '@/lib/utils';
import { Zap, Trophy, ArrowRight, X, RotateCcw, Ticket, Loader2 } from 'lucide-react';
import apiClient from '@/lib/api-client';

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

interface SpinTier {
  id: string;
  tier_name: string;
  tier_display_name: string;
  min_daily_amount: number; // kobo
  max_daily_amount: number; // kobo  (-1 = unlimited)
  spins_per_day: number;
  tier_icon: string;
  tier_badge: string;
  description: string;
  sort_order: number;
  is_active: boolean;
}

const TIER_COLORS: Record<string, string> = {
  Bronze:   'from-amber-600  to-yellow-500',
  Silver:   'from-slate-400  to-gray-300',
  Gold:     'from-yellow-500 to-amber-400',
  Platinum: 'from-purple-500 to-indigo-400',
};

/** Format kobo range into a display string like "₦1,000 – ₦2,499" */
function formatTierRange(minKobo: number, maxKobo: number): string {
  const min = formatCurrency(minKobo / 100);
  if (maxKobo < 0 || maxKobo === 0) return `${min}+`;
  return `${min} – ${formatCurrency(maxKobo / 100)}`;
}

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
  const [tiers, setTiers] = useState<SpinTier[]>([]);
  const [tiersLoading, setTiersLoading] = useState(false);

  // Fetch live tier data when the nudge opens
  useEffect(() => {
    if (!isOpen) return;
    setTiersLoading(true);
    apiClient
      .get<any>('/spins/tiers')
      .then((res) => {
        // res is AxiosResponse<any> — the backend wraps in { success, data: { tiers: [...] } }
        // res.data is the parsed JSON body
        const body = res?.data ?? {};
        const fetched: SpinTier[] = body?.data?.tiers ?? body?.tiers ?? [];
        if (Array.isArray(fetched) && fetched.length > 0) {
          setTiers(fetched.filter((t) => t.is_active).sort((a, b) => a.sort_order - b.sort_order));
        }
      })
      .catch(() => {
        // Silently fall back to empty — component still renders without the table
      })
      .finally(() => setTiersLoading(false));
  }, [isOpen]);

  if (!isOpen) return null;

  const hasNextTier    = !!nextTierName && !!nextTierSpins;
  const tierGradient   = nextTierName ? (TIER_COLORS[nextTierName] ?? 'from-green-500 to-emerald-400') : '';
  const tierEmoji      = nextTierName
    ? (tiers.find((t) => t.tier_display_name === nextTierName)?.tier_icon ?? '🏆')
    : '🏆';

  const nudgeAmountNaira = amountToNextTier
    ? amountToNextTier / 100
    : nextTierMinAmount
      ? nextTierMinAmount / 100
      : 0;

  const nudgeDrawEntries = nudgeAmountNaira > 0 ? Math.floor(nudgeAmountNaira / 200) : 0;

  return (
    <div className="fixed inset-0 bg-black/80 flex items-start sm:items-center justify-center z-50 p-4 overflow-y-auto">
      <Card className="w-full max-w-md shadow-2xl border-0 overflow-hidden my-auto">

        {/* Header */}
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

          {/* Draw entries motivation */}
          <div className="rounded-xl bg-gradient-to-r from-indigo-50 to-purple-50 border border-indigo-200 p-4 space-y-2">
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-full bg-indigo-100 flex items-center justify-center flex-shrink-0">
                <Ticket className="w-4 h-4 text-indigo-600" />
              </div>
              <p className="text-sm font-bold text-indigo-900">Boost your daily prize draw entries too!</p>
            </div>
            <p className="text-xs text-indigo-700 leading-relaxed">
              Every ₦200 you recharge earns <span className="font-bold">1 draw entry</span> into the daily jackpot prize draw.
              {nudgeDrawEntries > 0 && (
                <> Top up <span className="font-bold">{formatCurrency(nudgeAmountNaira)}</span> more and you'll
                also earn at least <span className="font-bold">{nudgeDrawEntries} extra draw {nudgeDrawEntries === 1 ? 'entry' : 'entries'}</span> —
                more entries = better chances of winning big!</>
              )}
              {nudgeDrawEntries === 0 && (
                <> The more you recharge, the more entries you earn and the better your chances of winning the daily jackpot!</>
              )}
            </p>
            <p className="text-xs text-indigo-500 italic">
              🎯 Higher loyalty tiers earn bonus multiplied entries on every recharge
            </p>
          </div>

          {/* Live tier ladder — fetched from API */}
          <div className="space-y-1.5">
            <div className="flex items-center justify-between">
              <p className="text-xs font-semibold text-gray-500 uppercase tracking-wide">Spin tiers</p>
              {tiersLoading && <Loader2 className="w-3 h-3 text-gray-400 animate-spin" />}
            </div>

            {tiers.length > 0 ? (
              tiers.map((t) => (
                <div
                  key={t.id}
                  className={`flex items-center justify-between px-3 py-1.5 rounded-lg text-xs ${
                    t.tier_display_name === nextTierName
                      ? 'bg-yellow-50 border border-yellow-300 font-semibold'
                      : 'bg-gray-50 text-gray-600'
                  }`}
                >
                  <span>
                    {t.tier_icon} {t.tier_display_name}{' '}
                    <span className="text-gray-400 font-normal">
                      ({formatTierRange(t.min_daily_amount, t.max_daily_amount)})
                    </span>
                  </span>
                  <span className="font-bold text-gray-700">
                    {t.spins_per_day} spin{t.spins_per_day > 1 ? 's' : ''}
                  </span>
                </div>
              ))
            ) : !tiersLoading ? (
              /* Skeleton rows shown while loading or if API failed silently */
              [1, 2, 3, 4].map((i) => (
                <div key={i} className="h-7 bg-gray-100 rounded-lg animate-pulse" />
              ))
            ) : null}
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
