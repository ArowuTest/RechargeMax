import React, { useState, useEffect } from 'react';
import { apiClient } from '@/lib/api-client';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { 
  Gift, 
  RefreshCw, 
  CheckCircle2, 
  Clock,
  Download,
  Filter,
  Bell
} from 'lucide-react';



interface UnclaimedPrize {
  id: string;
  msisdn: string;
  prizeType: string;
  prizeDescription: string;
  prizeAmount?: number;
  airtimeAmount?: number;
  dataPackage?: string;
  claimStatus: string;
  claimDeadline: string;
  daysUntilDeadline: number;
  drawName: string;
  createdAt: string;
}

const UnclaimedPrizesDashboard: React.FC = () => {
  const [prizes, setPrizes] = useState<UnclaimedPrize[]>([]);
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState<string>('all');
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPrizes, setTotalPrizes] = useState(0);
  const PAGE_SIZE = 20;

  useEffect(() => {
    fetchUnclaimedPrizes();
    // Auto-refresh every 60 seconds
    const interval = setInterval(fetchUnclaimedPrizes, 60000);
    return () => clearInterval(interval);
  }, []);

  const fetchUnclaimedPrizes = async () => {
    setLoading(true);
    try {
      const response = await apiClient.get(`/admin/winners/pending-claims?page=${currentPage}&per_page=${PAGE_SIZE}`);
      const data = response.data;
      const list = data.prizes
        || (Array.isArray(data.data) ? data.data : null)
        || data.data?.winners
        || [];
      setPrizes(list);
      setTotalPrizes(data.total || data.data?.total || list.length);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load unclaimed prizes');
    } finally {
      setLoading(false);
    }
  };

  const handleSendReminders = async () => {
    setLoading(true);
    setSuccess(null);
    setError(null);
    
    try {
      const response = await apiClient.post('/admin/prize-fulfillment/send-reminders');
      const data = response.data;
      setSuccess(`Sent ${data.count} reminder notifications`);
      setTimeout(() => setSuccess(null), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to send reminders');
    } finally {
      setLoading(false);
    }
  };

  const handleExport = () => {
    const csv = [
      ['MSISDN', 'Prize Type', 'Prize Description', 'Draw', 'Deadline', 'Days Left', 'Status'],
      ...prizes.map(p => [
        p.msisdn,
        p.prizeType,
        p.prizeDescription,
        p.drawName,
        new Date(p.claimDeadline).toLocaleDateString(),
        p.daysUntilDeadline.toString(),
        p.claimStatus
      ])
    ].map(row => row.join(',')).join('\n');
    
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `unclaimed-prizes-${new Date().toISOString().split('T')[0]}.csv`;
    a.click();
  };

  const filteredPrizes = prizes.filter(p => {
    if (filter === 'all') return true;
    if (filter === 'urgent') return p.daysUntilDeadline <= 7;
    if (filter === 'critical') return p.daysUntilDeadline <= 3;
    return true;
  });

  const stats = {
    total: prizes.length,
    urgent: prizes.filter(p => p.daysUntilDeadline <= 7).length,
    critical: prizes.filter(p => p.daysUntilDeadline <= 3).length
  };

  const getUrgencyBadge = (days: number) => {
    if (days <= 3) return <Badge className="bg-red-500">Critical</Badge>;
    if (days <= 7) return <Badge className="bg-orange-500">Urgent</Badge>;
    return <Badge variant="outline">Pending</Badge>;
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold flex items-center gap-2">
            <Gift className="w-8 h-8 text-blue-500" />
            Unclaimed Prizes
          </h2>
          <p className="text-gray-600 mt-1">
            Monitor prizes waiting for user claim
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={fetchUnclaimedPrizes}
            disabled={loading}
          >
            <RefreshCw className={`w-4 h-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          <Button
            variant="outline"
            onClick={handleExport}
            disabled={prizes.length === 0}
          >
            <Download className="w-4 h-4 mr-2" />
            Export CSV
          </Button>
        </div>
      </div>

      {/* Alerts */}
      {success && (
        <Alert className="bg-green-50 border-green-200">
          <CheckCircle2 className="w-4 h-4 text-green-600" />
          <AlertDescription className="text-green-800">{success}</AlertDescription>
        </Alert>
      )}
      
      {error && (
        <Alert className="bg-red-50 border-red-200">
          <AlertDescription className="text-red-800">{error}</AlertDescription>
        </Alert>
      )}

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardContent className="pt-6">
            <div className="text-center">
              <div className="text-4xl font-bold text-blue-600">{stats.total}</div>
              <div className="text-gray-600 mt-1">Total Unclaimed</div>
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="pt-6">
            <div className="text-center">
              <div className="text-4xl font-bold text-orange-600">{stats.urgent}</div>
              <div className="text-gray-600 mt-1">Urgent (≤7 days)</div>
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="pt-6">
            <div className="text-center">
              <div className="text-4xl font-bold text-red-600">{stats.critical}</div>
              <div className="text-gray-600 mt-1">Critical (≤3 days)</div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Filters and Actions */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Filter className="w-4 h-4 text-gray-500" />
              <select
                value={filter}
                onChange={(e) => setFilter(e.target.value)}
                className="border rounded px-3 py-2"
              >
                <option value="all">All ({stats.total})</option>
                <option value="urgent">Urgent ({stats.urgent})</option>
                <option value="critical">Critical ({stats.critical})</option>
              </select>
            </div>
            
            {prizes.length > 0 && (
              <Button
                onClick={handleSendReminders}
                disabled={loading}
                className="bg-blue-600 hover:bg-blue-700"
              >
                <Bell className="w-4 h-4 mr-2" />
                Send Reminders
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Unclaimed Prizes Table */}
      <Card>
        <CardHeader>
          <CardTitle>Unclaimed Prizes ({filteredPrizes.length})</CardTitle>
        </CardHeader>
        <CardContent>
          {filteredPrizes.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              {prizes.length === 0 
                ? '🎉 No unclaimed prizes! All prizes have been claimed or auto-provisioned.'
                : 'No prizes match the selected filter.'}
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b">
                    <th className="text-left p-3">MSISDN</th>
                    <th className="text-left p-3">Prize</th>
                    <th className="text-left p-3">Draw</th>
                    <th className="text-center p-3">Deadline</th>
                    <th className="text-center p-3">Days Left</th>
                    <th className="text-center p-3">Status</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredPrizes.map((prize) => (
                    <tr key={prize.id} className="border-b hover:bg-gray-50">
                      <td className="p-3 font-mono">{prize.msisdn}</td>
                      <td className="p-3">
                        <div>
                          <Badge variant="outline">{prize.prizeType}</Badge>
                          <div className="text-sm text-gray-600 mt-1">
                            {prize.prizeDescription}
                          </div>
                          {prize.airtimeAmount && (
                            <div className="text-sm font-semibold text-green-600 mt-1">
                              ₦{(prize.airtimeAmount / 100).toFixed(2)}
                            </div>
                          )}
                          {prize.dataPackage && (
                            <div className="text-sm font-semibold text-blue-600 mt-1">
                              {prize.dataPackage}
                            </div>
                          )}
                        </div>
                      </td>
                      <td className="p-3 text-sm text-gray-600">
                        {prize.drawName}
                      </td>
                      <td className="p-3 text-center text-sm">
                        {new Date(prize.claimDeadline).toLocaleDateString()}
                      </td>
                      <td className="p-3 text-center">
                        <div className="flex items-center justify-center gap-2">
                          <Clock className="w-4 h-4 text-gray-500" />
                          <span className={`font-semibold ${
                            prize.daysUntilDeadline <= 3 ? 'text-red-600' :
                            prize.daysUntilDeadline <= 7 ? 'text-orange-600' :
                            'text-gray-600'
                          }`}>
                            {prize.daysUntilDeadline}
                          </span>
                        </div>
                      </td>
                      <td className="p-3 text-center">
                        {getUrgencyBadge(prize.daysUntilDeadline)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        {/* Pagination */}
        <div className="flex items-center justify-between mt-4 px-1">
          <p className="text-xs text-gray-500">Page {currentPage} · {prizes.length} of {totalPrizes} unclaimed prizes</p>
          <div className="flex gap-2">
            <button className="px-3 py-1 text-xs border rounded disabled:opacity-40"
              disabled={currentPage <= 1} onClick={() => setCurrentPage(p => Math.max(1, p - 1))}>← Prev</button>
            <button className="px-3 py-1 text-xs border rounded disabled:opacity-40"
              disabled={prizes.length < PAGE_SIZE} onClick={() => setCurrentPage(p => p + 1)}>Next →</button>
          </div>
        </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default UnclaimedPrizesDashboard;
