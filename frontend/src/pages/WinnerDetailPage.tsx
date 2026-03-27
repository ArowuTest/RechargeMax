import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import {
  Trophy, Gift, Smartphone, DollarSign, ArrowLeft,
  Calendar, CheckCircle, Clock, AlertCircle, Share2
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { apiClient } from '@/lib/api-client';
import { formatCurrency } from '@/lib/utils';
import { toast } from '@/components/ui/sonner';

interface PublicWinner {
  id: string;
  draw_id: string;
  draw_name: string;
  draw_type: string;
  masked_msisdn: string;
  position: number;
  prize_type: string;
  prize_description: string;
  prize_amount: number;
  claim_status: string;
  won_at: string;
}

const CLAIM_STATUS_CONFIG: Record<string, { label: string; icon: React.ReactNode; className: string }> = {
  CLAIMED:              { label: 'Claimed',          icon: <CheckCircle className="w-4 h-4" />, className: 'bg-green-100 text-green-700 border-green-300'  },
  APPROVED:             { label: 'Approved',         icon: <CheckCircle className="w-4 h-4" />, className: 'bg-blue-100 text-blue-700 border-blue-300'    },
  PENDING:              { label: 'Pending Claim',    icon: <Clock className="w-4 h-4"        />, className: 'bg-yellow-100 text-yellow-700 border-yellow-300' },
  PENDING_ADMIN_REVIEW: { label: 'Under Review',     icon: <Clock className="w-4 h-4"        />, className: 'bg-orange-100 text-orange-700 border-orange-300' },
  REJECTED:             { label: 'Rejected',         icon: <AlertCircle className="w-4 h-4"  />, className: 'bg-red-100 text-red-700 border-red-300'       },
  EXPIRED:              { label: 'Expired',          icon: <AlertCircle className="w-4 h-4"  />, className: 'bg-gray-100 text-gray-500 border-gray-300'    },
};

const PRIZE_ICONS: Record<string, React.ReactNode> = {
  cash:    <DollarSign className="w-8 h-8" />,
  airtime: <Smartphone className="w-8 h-8" />,
  data:    <Smartphone className="w-8 h-8" />,
  goods:   <Gift className="w-8 h-8" />,
};

const PRIZE_BG: Record<string, string> = {
  cash:    'from-green-500 to-emerald-600',
  airtime: 'from-blue-500 to-cyan-600',
  data:    'from-purple-500 to-violet-600',
  goods:   'from-orange-500 to-amber-600',
};

const POSITION_LABELS = ['', '1st Place 🥇', '2nd Place 🥈', '3rd Place 🥉'];

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('en-NG', {
    weekday: 'long', year: 'numeric', month: 'long', day: 'numeric',
  });
}

function formatTime(iso: string) {
  return new Date(iso).toLocaleTimeString('en-NG', { hour: '2-digit', minute: '2-digit' });
}

function formatPrize(w: PublicWinner) {
  if (w.prize_type === 'cash' && w.prize_amount > 0) {
    return formatCurrency(w.prize_amount / 100);
  }
  return w.prize_description;
}

export const WinnerDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [winner, setWinner] = useState<PublicWinner | null>(null);
  const [loading, setLoading] = useState(true);
  const [notFound, setNotFound] = useState(false);

  useEffect(() => {
    if (!id) return;
    const fetch = async () => {
      setLoading(true);
      try {
        const res = await apiClient.get(`/winners/${id}`) as any;
        const body = res?.data ?? res;
        const winnerData = body?.data ?? null;
        if (winnerData?.id) {
          setWinner(winnerData as PublicWinner);
        } else {
          setNotFound(true);
        }
      } catch {
        setNotFound(true);
      } finally {
        setLoading(false);
      }
    };
    fetch();
  }, [id]);

  const handleShare = () => {
    const url = window.location.href;
    if (navigator.share) {
      navigator.share({ title: 'RechargeMax Winner!', text: `Someone just won a prize on RechargeMax! 🎉`, url });
    } else {
      navigator.clipboard.writeText(url);
      toast.success('Link copied to clipboard!');
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-purple-50 to-blue-50">
        <div className="text-center">
          <Trophy className="w-14 h-14 mx-auto mb-4 text-purple-400 animate-bounce" />
          <p className="text-gray-500 text-lg">Loading winner details...</p>
        </div>
      </div>
    );
  }

  if (notFound || !winner) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-purple-50 to-blue-50">
        <div className="text-center">
          <AlertCircle className="w-14 h-14 mx-auto mb-4 text-gray-400" />
          <h2 className="text-2xl font-bold text-gray-700 mb-2">Winner Not Found</h2>
          <p className="text-gray-500 mb-6">This winner record doesn't exist or has been removed.</p>
          <Link to="/winners">
            <Button variant="outline" className="flex items-center gap-2 mx-auto">
              <ArrowLeft className="w-4 h-4" /> Back to Winners Wall
            </Button>
          </Link>
        </div>
      </div>
    );
  }

  const claimCfg = (CLAIM_STATUS_CONFIG[winner.claim_status] ?? CLAIM_STATUS_CONFIG['PENDING'])!;
  const prizeGradient = PRIZE_BG[winner.prize_type] || 'from-purple-500 to-blue-600';

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 via-white to-blue-50 py-8 px-4">
      <div className="max-w-xl mx-auto">
        {/* Back nav */}
        <Link to="/winners">
          <Button variant="ghost" size="sm" className="mb-6 flex items-center gap-2 text-gray-600 hover:text-purple-700">
            <ArrowLeft className="w-4 h-4" /> All Winners
          </Button>
        </Link>

        {/* Prize hero card */}
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.4 }}
        >
          <Card className="overflow-hidden shadow-xl border-0 mb-6">
            {/* Gradient header */}
            <div className={`bg-gradient-to-r ${prizeGradient} p-8 text-white text-center`}>
              <div className="w-16 h-16 rounded-full bg-white/20 flex items-center justify-center mx-auto mb-4">
                {PRIZE_ICONS[winner.prize_type] || <Trophy className="w-8 h-8" />}
              </div>
              <p className="text-white/80 text-sm font-medium uppercase tracking-widest mb-1">
                {winner.position <= 3 ? POSITION_LABELS[winner.position] : `Position #${winner.position}`}
              </p>
              <h1 className="text-3xl font-extrabold mb-1">{formatPrize(winner)}</h1>
              <p className="text-white/80">{winner.prize_description}</p>
            </div>

            <CardContent className="p-6 space-y-4">
              {/* Claim status */}
              <div className={`flex items-center gap-2 px-4 py-3 rounded-xl border text-sm font-medium ${claimCfg.className}`}>
                {claimCfg.icon}
                {claimCfg.label}
              </div>

              <Separator />

              {/* Details grid */}
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <p className="text-gray-500 text-xs uppercase tracking-wide mb-1">Draw</p>
                  <p className="font-semibold text-gray-800">{winner.draw_name}</p>
                </div>
                <div>
                  <p className="text-gray-500 text-xs uppercase tracking-wide mb-1">Draw Type</p>
                  <Badge variant="outline" className="capitalize">{winner.draw_type.toLowerCase()}</Badge>
                </div>
                <div>
                  <p className="text-gray-500 text-xs uppercase tracking-wide mb-1">Winner</p>
                  <p className="font-mono font-semibold text-gray-800">{winner.masked_msisdn}</p>
                </div>
                <div>
                  <p className="text-gray-500 text-xs uppercase tracking-wide mb-1">Prize Type</p>
                  <Badge variant="outline" className="capitalize">{winner.prize_type}</Badge>
                </div>
              </div>

              <Separator />

              {/* Date/time */}
              <div className="flex items-center gap-3 text-sm text-gray-600">
                <Calendar className="w-4 h-4 text-purple-500 flex-shrink-0" />
                <div>
                  <span className="font-medium">{formatDate(winner.won_at)}</span>
                  <span className="text-gray-400 ml-2">at {formatTime(winner.won_at)}</span>
                </div>
              </div>
            </CardContent>
          </Card>
        </motion.div>

        {/* Share + CTA */}
        <div className="flex flex-col sm:flex-row gap-3">
          <Button
            variant="outline"
            className="flex-1 flex items-center justify-center gap-2"
            onClick={handleShare}
          >
            <Share2 className="w-4 h-4" /> Share
          </Button>
          <Link to="/" className="flex-1">
            <Button className="w-full bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-700 hover:to-blue-700 text-white">
              Recharge & Enter Now
            </Button>
          </Link>
        </div>

        {/* Footer note */}
        <p className="text-center text-xs text-gray-400 mt-6">
          Phone numbers are partially masked to protect privacy. Winners are selected randomly and verified by RechargeMax.
        </p>
      </div>
    </div>
  );
};

export default WinnerDetailPage;
