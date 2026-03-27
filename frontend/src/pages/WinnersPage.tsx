import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Trophy, Gift, Smartphone, DollarSign, ChevronLeft, ChevronRight, Search, Calendar, Award } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { apiClient } from '@/lib/api-client';
import { formatCurrency } from '@/lib/utils';

interface PublicWinner {
  id: string;
  draw_id: string;
  draw_name: string;
  draw_type: string;
  masked_msisdn: string;
  position: number;
  prize_type: string;
  prize_description: string;
  prize_amount: number; // kobo
  claim_status: string;
  won_at: string;
}

const PRIZE_ICONS: Record<string, React.ReactNode> = {
  cash:    <DollarSign className="w-5 h-5" />,
  airtime: <Smartphone className="w-5 h-5" />,
  data:    <Smartphone className="w-5 h-5" />,
  goods:   <Gift className="w-5 h-5" />,
};

const PRIZE_COLORS: Record<string, string> = {
  cash:    'bg-green-100 text-green-700 border-green-200',
  airtime: 'bg-blue-100 text-blue-700 border-blue-200',
  data:    'bg-purple-100 text-purple-700 border-purple-200',
  goods:   'bg-orange-100 text-orange-700 border-orange-200',
};

const POSITION_BADGE = ['', '🥇', '🥈', '🥉'];

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('en-NG', {
    year: 'numeric', month: 'short', day: 'numeric',
  });
}

function formatPrize(w: PublicWinner) {
  if (w.prize_type === 'cash' && w.prize_amount > 0) {
    return formatCurrency(w.prize_amount / 100); // kobo → naira
  }
  return w.prize_description;
}

export const WinnersPage: React.FC = () => {
  const [winners, setWinners]     = useState<PublicWinner[]>([]);
  const [total, setTotal]         = useState(0);
  const [page, setPage]           = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [loading, setLoading]     = useState(true);
  const [search, setSearch]       = useState('');
  const [filter, setFilter]       = useState<string>('all');
  const LIMIT = 12;

  useEffect(() => {
    const fetchWinners = async () => {
      setLoading(true);
      try {
        const res = await apiClient.get(`/winners?page=${page}&limit=${LIMIT}`) as any;
        const body = res?.data ?? res; // axios wraps in .data; apiClient may unwrap
        const payload = body?.data ?? body; // backend: { success, data: { winners, total, ... } }
        if (payload?.winners) {
          setWinners(payload.winners || []);
          setTotal(payload.total || 0);
          setTotalPages(payload.total_pages || 1);
        }
      } catch (e) {
        console.error('Failed to fetch winners:', e);
      } finally {
        setLoading(false);
      }
    };
    fetchWinners();
  }, [page]);

  const displayed = winners.filter(w => {
    const matchSearch =
      search === '' ||
      (w.draw_name || '').toLowerCase().includes(search.toLowerCase()) ||
      (w.prize_description || '').toLowerCase().includes(search.toLowerCase()) ||
      w.masked_msisdn.includes(search);
    const matchFilter = filter === 'all' || w.prize_type === filter;
    return matchSearch && matchFilter;
  });

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 via-white to-blue-50">
      {/* Hero */}
      <div className="bg-gradient-to-r from-purple-700 to-blue-600 text-white py-16 px-4">
        <div className="max-w-5xl mx-auto text-center">
          <motion.div
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
          >
            <Trophy className="w-16 h-16 mx-auto mb-4 text-yellow-300 drop-shadow-lg" />
            <h1 className="text-4xl font-extrabold mb-2 tracking-tight">Winners Wall</h1>
            <p className="text-purple-200 text-lg">
              Real people winning real prizes every day.
              {total > 0 && <span className="font-semibold text-white"> {total.toLocaleString()} winners and counting!</span>}
            </p>
          </motion.div>
        </div>
      </div>

      <div className="max-w-5xl mx-auto px-4 py-10">
        {/* Filters */}
        <div className="flex flex-col sm:flex-row gap-3 mb-8">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
            <Input
              placeholder="Search by draw name, prize, phone..."
              className="pl-9"
              value={search}
              onChange={e => setSearch(e.target.value)}
            />
          </div>
          <div className="flex gap-2 flex-wrap">
            {['all', 'cash', 'airtime', 'data', 'goods'].map(f => (
              <Button
                key={f}
                size="sm"
                variant={filter === f ? 'default' : 'outline'}
                onClick={() => setFilter(f)}
                className="capitalize"
              >
                {f === 'all' ? 'All Prizes' : f}
              </Button>
            ))}
          </div>
        </div>

        {/* Grid */}
        {loading ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
            {Array.from({ length: 9 }).map((_, i) => (
              <div key={i} className="h-44 bg-gray-100 rounded-2xl animate-pulse" />
            ))}
          </div>
        ) : displayed.length === 0 ? (
          <div className="text-center py-20 text-gray-500">
            <Award className="w-14 h-14 mx-auto mb-4 opacity-30" />
            <p className="text-lg font-medium">No winners found</p>
            <p className="text-sm mt-1">Check back after the next draw!</p>
          </div>
        ) : (
          <motion.div
            className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5"
            initial="hidden"
            animate="visible"
            variants={{ visible: { transition: { staggerChildren: 0.06 } } }}
          >
            {displayed.map(w => (
              <motion.div
                key={w.id}
                variants={{ hidden: { opacity: 0, y: 20 }, visible: { opacity: 1, y: 0 } }}
              >
                <Link to={`/winners/${w.id}`}>
                  <Card className="group hover:shadow-lg hover:-translate-y-1 transition-all duration-200 border border-gray-100 rounded-2xl overflow-hidden cursor-pointer">
                    <CardContent className="p-5">
                      <div className="flex items-start justify-between mb-3">
                        <div className={`p-2 rounded-xl border ${PRIZE_COLORS[w.prize_type] || 'bg-gray-100 text-gray-600 border-gray-200'}`}>
                          {PRIZE_ICONS[w.prize_type] || <Trophy className="w-5 h-5" />}
                        </div>
                        <span className="text-2xl" title={`Position ${w.position}`}>
                          {POSITION_BADGE[w.position] || `#${w.position}`}
                        </span>
                      </div>

                      <p className="font-bold text-gray-800 text-base leading-tight mb-1 group-hover:text-purple-700 transition-colors">
                        {formatPrize(w)}
                      </p>
                      <p className="text-xs text-gray-500 mb-3 truncate">{w.draw_name}</p>

                      <div className="flex items-center justify-between text-xs text-gray-500">
                        <span className="font-mono bg-gray-50 px-2 py-0.5 rounded border">{w.masked_msisdn}</span>
                        <span className="flex items-center gap-1">
                          <Calendar className="w-3 h-3" />
                          {formatDate(w.won_at)}
                        </span>
                      </div>

                      <div className="mt-3 pt-3 border-t border-gray-50 flex items-center justify-between">
                        <Badge variant="outline" className="text-xs capitalize">{w.prize_type}</Badge>
                        <Badge
                          variant="outline"
                          className={`text-xs ${w.claim_status === 'CLAIMED' ? 'bg-green-50 text-green-700 border-green-200' : 'bg-yellow-50 text-yellow-700 border-yellow-200'}`}
                        >
                          {w.claim_status}
                        </Badge>
                      </div>
                    </CardContent>
                  </Card>
                </Link>
              </motion.div>
            ))}
          </motion.div>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex items-center justify-center gap-4 mt-10">
            <Button
              variant="outline" size="sm"
              disabled={page <= 1}
              onClick={() => setPage(p => p - 1)}
              className="flex items-center gap-1"
            >
              <ChevronLeft className="w-4 h-4" /> Prev
            </Button>
            <span className="text-sm text-gray-600">
              Page <strong>{page}</strong> of <strong>{totalPages}</strong>
            </span>
            <Button
              variant="outline" size="sm"
              disabled={page >= totalPages}
              onClick={() => setPage(p => p + 1)}
              className="flex items-center gap-1"
            >
              Next <ChevronRight className="w-4 h-4" />
            </Button>
          </div>
        )}
      </div>
    </div>
  );
};

export default WinnersPage;
