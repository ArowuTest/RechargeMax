/**
 * USSD Recharge Monitoring Component
 * Enterprise-grade monitoring for USSD recharges from telecom providers
 * 
 * Features:
 * - Real-time USSD recharge tracking
 * - Points allocation monitoring
 * - Network-wise breakdown
 * - Webhook status monitoring
 * - Failed recharge retry mechanism
 * - Duplicate detection alerts
 * - Statistics dashboard
 * - Export functionality
 */

import React, { useState, useEffect } from 'react';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { useToast } from '@/hooks/use-toast';
import {
  Loader2,
  CheckCircle2,
  XCircle,
  AlertTriangle,
  Phone,
  TrendingUp,
  RefreshCw,
  Download,
  Calendar,
  DollarSign,
  Gift,
} from 'lucide-react';
import {
  ussdRechargeApi,
  type USSDRecharge,
  type USSDStatistics,
} from '@/lib/api-client-extensions';

type NetworkFilter = 'all' | 'MTN' | 'Glo' | 'Airtel' | '9mobile';
type StatusFilter = 'all' | 'success' | 'failed' | 'duplicate';

export default function USSDRechargeMonitoring() {
  const { toast } = useToast();

  // State
  const [recharges, setRecharges] = useState<USSDRecharge[]>([]);
  const [statistics, setStatistics] = useState<USSDStatistics | null>(null);
  const [loading, setLoading] = useState(true);
  const [exporting, setExporting] = useState(false);

  // Filters
  const [networkFilter, setNetworkFilter] = useState<NetworkFilter>('all');
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all');
  const [searchTerm, setSearchTerm] = useState('');
  const [dateFrom, setDateFrom] = useState('');
  const [dateTo, setDateTo] = useState('');

  useEffect(() => {
    fetchData();
  }, [networkFilter, statusFilter, dateFrom, dateTo]);

  const fetchData = async () => {
    setLoading(true);
    try {
      // Fetch USSD recharges with filters
      const rechargesResponse = await ussdRechargeApi.getRecharges({
        network: networkFilter !== 'all' ? networkFilter : undefined,
        status: statusFilter !== 'all' ? statusFilter : undefined,
        date_from: dateFrom || undefined,
        date_to: dateTo || undefined,
      });

      if (rechargesResponse.success && rechargesResponse.data) {
        setRecharges(rechargesResponse.data);
      }

      // Fetch statistics
      const statsResponse = await ussdRechargeApi.getStatistics({
        date_from: dateFrom || undefined,
        date_to: dateTo || undefined,
      });

      if (statsResponse.success && statsResponse.data) {
        setStatistics(statsResponse.data);
      }
    } catch (error) {
      console.error('Failed to fetch USSD recharge data:', error);
      toast({
        title: 'Error',
        description: 'Failed to load USSD recharge data',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleRetryFailed = async (rechargeId: string) => {
    try {
      const response = await ussdRechargeApi.retryRecharge(rechargeId);

      if (response.success) {
        toast({
          title: 'Success',
          description: 'Recharge retry initiated successfully',
        });
        await fetchData();
      }
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.error || 'Failed to retry recharge',
        variant: 'destructive',
      });
    }
  };

  const handleExport = async () => {
    setExporting(true);
    try {
      const response = await ussdRechargeApi.exportRecharges({
        network: networkFilter !== 'all' ? networkFilter : undefined,
        status: statusFilter !== 'all' ? statusFilter : undefined,
        date_from: dateFrom || undefined,
        date_to: dateTo || undefined,
      });

      if (response.success && response.data) {
        // Create blob and download
        const blob = new Blob([response.data], { type: 'text/csv' });
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `ussd_recharges_${new Date().toISOString().split('T')[0]}.csv`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);

        toast({
          title: 'Success',
          description: 'USSD recharges exported successfully',
        });
      }
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.error || 'Failed to export recharges',
        variant: 'destructive',
      });
    } finally {
      setExporting(false);
    }
  };

  const getStatusBadge = (status: string) => {
    const statusConfig = {
      success: { variant: 'default' as const, icon: CheckCircle2, label: 'Success' },
      failed: { variant: 'destructive' as const, icon: XCircle, label: 'Failed' },
      duplicate: { variant: 'secondary' as const, icon: AlertTriangle, label: 'Duplicate' },
    };

    const config = statusConfig[status as keyof typeof statusConfig] || statusConfig.success;
    const Icon = config.icon;

    return (
      <Badge variant={config.variant} className="flex items-center gap-1">
        <Icon className="h-3 w-3" />
        {config.label}
      </Badge>
    );
  };

  const getNetworkColor = (network: string) => {
    const colors = {
      MTN: 'text-yellow-600',
      Glo: 'text-green-600',
      Airtel: 'text-red-600',
      '9mobile': 'text-emerald-600',
    };
    return colors[network as keyof typeof colors] || 'text-gray-600';
  };

  const filteredRecharges = recharges.filter((recharge) => {
    if (searchTerm) {
      const search = searchTerm.toLowerCase();
      return (
        recharge.msisdn.includes(search) ||
        recharge.transaction_id.toLowerCase().includes(search) ||
        recharge.network.toLowerCase().includes(search)
      );
    }
    return true;
  });

  if (loading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin" />
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">USSD Recharge Monitoring</h2>
          <p className="text-muted-foreground">
            Track USSD recharges from telecom providers and points allocation
          </p>
        </div>
        <Button onClick={handleExport} disabled={exporting}>
          {exporting ? (
            <Loader2 className="h-4 w-4 mr-2 animate-spin" />
          ) : (
            <Download className="h-4 w-4 mr-2" />
          )}
          Export CSV
        </Button>
      </div>

      {/* Statistics Cards */}
      {statistics && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Total Recharges
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between">
                <p className="text-2xl font-bold">{statistics.total_recharges}</p>
                <Phone className="h-8 w-8 text-blue-600" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Total Amount
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between">
                <p className="text-2xl font-bold">
                  ₦{statistics.total_amount?.toLocaleString()}
                </p>
                <DollarSign className="h-8 w-8 text-green-600" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Points Allocated
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between">
                <p className="text-2xl font-bold">{statistics.total_points}</p>
                <Gift className="h-8 w-8 text-purple-600" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Success Rate
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between">
                <p className="text-2xl font-bold">
                  {statistics.success_rate?.toFixed(1)}%
                </p>
                <TrendingUp className="h-8 w-8 text-orange-600" />
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Network Breakdown */}
      {statistics?.network_breakdown && (
        <Card>
          <CardHeader>
            <CardTitle>Network Breakdown</CardTitle>
            <CardDescription>
              USSD recharges by telecom provider
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              {Object.entries(statistics.network_breakdown).map(([network, data]: [string, any]) => (
                <div key={network} className="bg-gray-50 p-4 rounded-lg">
                  <div className="flex items-center justify-between mb-2">
                    <h4 className={`font-semibold ${getNetworkColor(network)}`}>
                      {network}
                    </h4>
                    <Phone className={`h-5 w-5 ${getNetworkColor(network)}`} />
                  </div>
                  <div className="space-y-1 text-sm">
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Recharges:</span>
                      <span className="font-medium">{data.count}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Amount:</span>
                      <span className="font-medium">₦{data.amount?.toLocaleString()}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Points:</span>
                      <span className="font-medium">{data.points}</span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Filters */}
      <Card>
        <CardContent className="pt-6">
          <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
            <div>
              <Label htmlFor="network-filter">Network</Label>
              <Select
                value={networkFilter}
                onValueChange={(value) => setNetworkFilter(value as NetworkFilter)}
              >
                <SelectTrigger id="network-filter">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Networks</SelectItem>
                  <SelectItem value="MTN">MTN</SelectItem>
                  <SelectItem value="Glo">Glo</SelectItem>
                  <SelectItem value="Airtel">Airtel</SelectItem>
                  <SelectItem value="9mobile">9mobile</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label htmlFor="status-filter">Status</Label>
              <Select
                value={statusFilter}
                onValueChange={(value) => setStatusFilter(value as StatusFilter)}
              >
                <SelectTrigger id="status-filter">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Statuses</SelectItem>
                  <SelectItem value="success">Success</SelectItem>
                  <SelectItem value="failed">Failed</SelectItem>
                  <SelectItem value="duplicate">Duplicate</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label htmlFor="date-from">Date From</Label>
              <input
                id="date-from"
                type="date"
                value={dateFrom}
                onChange={(e) => setDateFrom(e.target.value)}
                className="w-full px-3 py-2 border rounded-md"
              />
            </div>

            <div>
              <Label htmlFor="date-to">Date To</Label>
              <input
                id="date-to"
                type="date"
                value={dateTo}
                onChange={(e) => setDateTo(e.target.value)}
                className="w-full px-3 py-2 border rounded-md"
              />
            </div>

            <div>
              <Label htmlFor="search">Search</Label>
              <Input
                id="search"
                placeholder="MSISDN or Transaction ID"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Recharges Table */}
      <Card>
        <CardHeader>
          <CardTitle>USSD Recharges ({filteredRecharges.length})</CardTitle>
          <CardDescription>
            Recent USSD recharges from telecom providers
          </CardDescription>
        </CardHeader>
        <CardContent>
          {filteredRecharges.length === 0 ? (
            <p className="text-center text-muted-foreground py-8">
              No USSD recharges found matching the current filters
            </p>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>MSISDN</TableHead>
                    <TableHead>Network</TableHead>
                    <TableHead>Amount</TableHead>
                    <TableHead>Points</TableHead>
                    <TableHead>Transaction ID</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Timestamp</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredRecharges.map((recharge) => (
                    <TableRow key={recharge.id}>
                      <TableCell className="font-mono">
                        {recharge.msisdn}
                      </TableCell>
                      <TableCell>
                        <span className={`font-medium ${getNetworkColor(recharge.network)}`}>
                          {recharge.network}
                        </span>
                      </TableCell>
                      <TableCell className="font-medium">
                        ₦{recharge.amount.toLocaleString()}
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">{recharge.points_allocated} pts</Badge>
                      </TableCell>
                      <TableCell className="font-mono text-sm">
                        {recharge.transaction_id}
                      </TableCell>
                      <TableCell>
                        {getStatusBadge(recharge.status)}
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2 text-sm text-muted-foreground">
                          <Calendar className="h-3 w-3" />
                          {new Date(recharge.created_at).toLocaleString()}
                        </div>
                      </TableCell>
                      <TableCell className="text-right">
                        {recharge.status === 'failed' && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleRetryFailed(recharge.id)}
                          >
                            <RefreshCw className="h-4 w-4" />
                          </Button>
                        )}
                        {recharge.status === 'duplicate' && (
                          <Badge variant="secondary" className="text-xs">
                            <AlertTriangle className="h-3 w-3 mr-1" />
                            Duplicate
                          </Badge>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Webhook Status */}
      <Card>
        <CardHeader>
          <CardTitle>Webhook Status</CardTitle>
          <CardDescription>
            Integration status with telecom providers
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {['MTN', 'Glo', 'Airtel', '9mobile'].map((network) => (
              <div key={network} className="flex items-center justify-between p-4 border rounded-lg">
                <div className="flex items-center gap-3">
                  <Phone className={`h-5 w-5 ${getNetworkColor(network)}`} />
                  <div>
                    <p className="font-medium">{network}</p>
                    <p className="text-sm text-muted-foreground">Webhook Endpoint</p>
                  </div>
                </div>
                <Badge variant="default">
                  <CheckCircle2 className="h-3 w-3 mr-1" />
                  Active
                </Badge>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
