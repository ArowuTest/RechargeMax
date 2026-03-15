import React, { useState, useEffect } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { 
  AlertCircle, 
  RefreshCw, 
  CheckCircle2, 
  XCircle,
  Download,
  Filter
} from 'lucide-react';

const API_BASE = '/api/v1';

interface FailedProvision {
  id: string;
  msisdn: string;
  prizeType: string;
  prizeDescription: string;
  provisionStatus: string;
  provisionError: string;
  provisionAttempts: number;
  maxRetryAttempts: number;
  lastProvisionAttemptAt: string;
  createdAt: string;
  allowRetryOnFailure: boolean;
}

const FailedProvisionsDashboard: React.FC = () => {
  const [provisions, setProvisions] = useState<FailedProvision[]>([]);
  const [loading, setLoading] = useState(false);
  const [retrying, setRetrying] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState<string>('all');

  useEffect(() => {
    fetchFailedProvisions();
    // Auto-refresh every 30 seconds
    const interval = setInterval(fetchFailedProvisions, 30000);
    return () => clearInterval(interval);
  }, []);

  const fetchFailedProvisions = async () => {
    setLoading(true);
    try {
      const response = await fetch(`${API_BASE}/admin/prize-fulfillment/failed-provisions`,  { credentials: 'include' });
      
      if (!response.ok) throw new Error('Failed to fetch failed provisions');
      
      const data = await response.json();
      setProvisions(data.provisions || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load failed provisions');
    } finally {
      setLoading(false);
    }
  };

  const handleRetry = async (provisionId: string) => {
    setRetrying(provisionId);
    setSuccess(null);
    setError(null);
    
    try {
      const response = await fetch(`/api/v1/admin/prize-fulfillment/retry/${provisionId}`, {
        method: 'POST',
        credentials: 'include'
      });
      
      if (!response.ok) throw new Error('Failed to retry provision');
      
      setSuccess('Retry initiated successfully');
      setTimeout(() => setSuccess(null), 3000);
      
      // Refresh list
      await fetchFailedProvisions();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to retry provision');
    } finally {
      setRetrying(null);
    }
  };

  const handleRetryAll = async () => {
    setLoading(true);
    setSuccess(null);
    setError(null);
    
    try {
      const response = await fetch(`/api/v1/admin/prize-fulfillment/retry-all`, {
        method: 'POST',
        credentials: 'include'
      });
      
      if (!response.ok) throw new Error('Failed to retry all provisions');
      
      const data = await response.json();
      setSuccess(`Initiated retry for ${data.count} provisions`);
      setTimeout(() => setSuccess(null), 3000);
      
      // Refresh list
      await fetchFailedProvisions();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to retry all provisions');
    } finally {
      setLoading(false);
    }
  };

  const handleExport = () => {
    const csv = [
      ['MSISDN', 'Prize Type', 'Prize Description', 'Error', 'Attempts', 'Max Attempts', 'Last Attempt', 'Created'],
      ...provisions.map(p => [
        p.msisdn,
        p.prizeType,
        p.prizeDescription,
        p.provisionError,
        p.provisionAttempts.toString(),
        p.maxRetryAttempts.toString(),
        new Date(p.lastProvisionAttemptAt).toLocaleString(),
        new Date(p.createdAt).toLocaleString()
      ])
    ].map(row => row.join(',')).join('\n');
    
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `failed-provisions-${new Date().toISOString().split('T')[0]}.csv`;
    a.click();
  };

  const filteredProvisions = provisions.filter(p => {
    if (filter === 'all') return true;
    if (filter === 'retryable') return p.allowRetryOnFailure && p.provisionAttempts < p.maxRetryAttempts;
    if (filter === 'maxed-out') return p.provisionAttempts >= p.maxRetryAttempts;
    return true;
  });

  const stats = {
    total: provisions.length,
    retryable: provisions.filter(p => p.allowRetryOnFailure && p.provisionAttempts < p.maxRetryAttempts).length,
    maxedOut: provisions.filter(p => p.provisionAttempts >= p.maxRetryAttempts).length
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold flex items-center gap-2">
            <AlertCircle className="w-8 h-8 text-red-500" />
            Failed Provisions
          </h2>
          <p className="text-gray-600 mt-1">
            Monitor and retry failed prize deliveries
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={fetchFailedProvisions}
            disabled={loading}
          >
            <RefreshCw className={`w-4 h-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          <Button
            variant="outline"
            onClick={handleExport}
            disabled={provisions.length === 0}
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
          <XCircle className="w-4 h-4 text-red-600" />
          <AlertDescription className="text-red-800">{error}</AlertDescription>
        </Alert>
      )}

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardContent className="pt-6">
            <div className="text-center">
              <div className="text-4xl font-bold text-red-600">{stats.total}</div>
              <div className="text-gray-600 mt-1">Total Failed</div>
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="pt-6">
            <div className="text-center">
              <div className="text-4xl font-bold text-orange-600">{stats.retryable}</div>
              <div className="text-gray-600 mt-1">Can Retry</div>
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="pt-6">
            <div className="text-center">
              <div className="text-4xl font-bold text-gray-600">{stats.maxedOut}</div>
              <div className="text-gray-600 mt-1">Max Attempts Reached</div>
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
                <option value="retryable">Can Retry ({stats.retryable})</option>
                <option value="maxed-out">Max Attempts ({stats.maxedOut})</option>
              </select>
            </div>
            
            {stats.retryable > 0 && (
              <Button
                onClick={handleRetryAll}
                disabled={loading}
                className="bg-orange-600 hover:bg-orange-700"
              >
                <RefreshCw className="w-4 h-4 mr-2" />
                Retry All ({stats.retryable})
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Failed Provisions Table */}
      <Card>
        <CardHeader>
          <CardTitle>Failed Provisions ({filteredProvisions.length})</CardTitle>
        </CardHeader>
        <CardContent>
          {filteredProvisions.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              {provisions.length === 0 
                ? '🎉 No failed provisions! All prizes delivered successfully.'
                : 'No provisions match the selected filter.'}
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b">
                    <th className="text-left p-3">MSISDN</th>
                    <th className="text-left p-3">Prize</th>
                    <th className="text-left p-3">Error</th>
                    <th className="text-center p-3">Attempts</th>
                    <th className="text-left p-3">Last Attempt</th>
                    <th className="text-center p-3">Action</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredProvisions.map((provision) => {
                    const canRetry = provision.allowRetryOnFailure && 
                                    provision.provisionAttempts < provision.maxRetryAttempts;
                    
                    return (
                      <tr key={provision.id} className="border-b hover:bg-gray-50">
                        <td className="p-3 font-mono">{provision.msisdn}</td>
                        <td className="p-3">
                          <div>
                            <Badge variant="outline">{provision.prizeType}</Badge>
                            <div className="text-sm text-gray-600 mt-1">
                              {provision.prizeDescription}
                            </div>
                          </div>
                        </td>
                        <td className="p-3">
                          <div className="text-sm text-red-600 max-w-xs truncate" title={provision.provisionError}>
                            {provision.provisionError}
                          </div>
                        </td>
                        <td className="p-3 text-center">
                          <Badge variant={canRetry ? "default" : "secondary"}>
                            {provision.provisionAttempts}/{provision.maxRetryAttempts}
                          </Badge>
                        </td>
                        <td className="p-3 text-sm text-gray-600">
                          {new Date(provision.lastProvisionAttemptAt).toLocaleString()}
                        </td>
                        <td className="p-3 text-center">
                          {canRetry ? (
                            <Button
                              size="sm"
                              onClick={() => handleRetry(provision.id)}
                              disabled={retrying === provision.id}
                            >
                              {retrying === provision.id ? (
                                <RefreshCw className="w-4 h-4 animate-spin" />
                              ) : (
                                <>
                                  <RefreshCw className="w-4 h-4 mr-1" />
                                  Retry
                                </>
                              )}
                            </Button>
                          ) : (
                            <Badge variant="secondary">Max Reached</Badge>
                          )}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
};

export default FailedProvisionsDashboard;
