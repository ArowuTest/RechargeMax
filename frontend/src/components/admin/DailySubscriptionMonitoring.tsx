/**
 * Daily Subscription Monitoring Component
 * Enterprise-grade monitoring for daily draw entry subscriptions
 * 
 * Features:
 * - Active subscription tracking
 * - Revenue analytics
 * - Subscription tier performance
 * - Churn analysis
 * - Billing status monitoring
 * - Failed payment retry
 * - Cancellation tracking
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
import { useToast } from '@/hooks/useToast';
import {
  Loader2,
  CheckCircle2,
  XCircle,
  Clock,
  TrendingUp,
  TrendingDown,
  DollarSign,
  Users,
  RefreshCw,
  Download,
  Calendar,
  Gift,
  AlertTriangle,
} from 'lucide-react';
import {
  subscriptionMonitoringApi,
  type DailySubscription,
  type SubscriptionStatistics,
  type SubscriptionBilling,
} from '@/lib/api-client-extensions';

type StatusFilter = 'all' | 'active' | 'paused' | 'cancelled';
type BillingStatusFilter = 'all' | 'success' | 'failed' | 'pending';

export default function DailySubscriptionMonitoring() {
  const { toast } = useToast();

  // State
  const [subscriptions, setSubscriptions] = useState<DailySubscription[]>([]);
  const [billings, setBillings] = useState<SubscriptionBilling[]>([]);
  const [statistics, setStatistics] = useState<SubscriptionStatistics | null>(null);
  const [loading, setLoading] = useState(true);
  const [exporting, setExporting] = useState(false);
  const [activeTab, setActiveTab] = useState<'subscriptions' | 'billings'>('subscriptions');

  // Filters
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all');
  const [billingStatusFilter, setBillingStatusFilter] = useState<BillingStatusFilter>('all');
  const [searchTerm, setSearchTerm] = useState('');
  const [dateFrom, setDateFrom] = useState('');
  const [dateTo, setDateTo] = useState('');

  useEffect(() => {
    fetchData();
  }, [statusFilter, billingStatusFilter, dateFrom, dateTo, activeTab]);

  const fetchData = async () => {
    setLoading(true);
    try {
      if (activeTab === 'subscriptions') {
        // Fetch subscriptions
        const subscriptionsResponse = await subscriptionMonitoringApi.getSubscriptions({
          status: statusFilter !== 'all' ? statusFilter : undefined,
          date_from: dateFrom || undefined,
          date_to: dateTo || undefined,
        });

        if (subscriptionsResponse.success && subscriptionsResponse.data) {
          setSubscriptions(subscriptionsResponse.data);
        }
      } else {
        // Fetch billings
        const billingsResponse = await subscriptionMonitoringApi.getBillings({
          billing_status: billingStatusFilter !== 'all' ? billingStatusFilter : undefined,
          date_from: dateFrom || undefined,
          date_to: dateTo || undefined,
        });

        if (billingsResponse.success && billingsResponse.data) {
          setBillings(billingsResponse.data);
        }
      }

      // Fetch statistics
      const statsResponse = await subscriptionMonitoringApi.getStatistics({
        date_from: dateFrom || undefined,
        date_to: dateTo || undefined,
      });

      if (statsResponse.success && statsResponse.data) {
        setStatistics(statsResponse.data);
      }
    } catch (error) {
      console.error('Failed to fetch subscription data:', error);
      toast({
        title: 'Error',
        description: 'Failed to load subscription data',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleRetryBilling = async (billingId: string) => {
    try {
      const response = await subscriptionMonitoringApi.retryBilling(billingId);

      if (response.success) {
        toast({
          title: 'Success',
          description: 'Billing retry initiated successfully',
        });
        await fetchData();
      }
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.error || 'Failed to retry billing',
        variant: 'destructive',
      });
    }
  };

  const handleExport = async () => {
    setExporting(true);
    try {
      const endpoint = activeTab === 'subscriptions'
        ? subscriptionMonitoringApi.exportSubscriptions
        : subscriptionMonitoringApi.exportBillings;

      const response = await endpoint({
        status: statusFilter !== 'all' ? statusFilter : undefined,
        billing_status: billingStatusFilter !== 'all' ? billingStatusFilter : undefined,
        date_from: dateFrom || undefined,
        date_to: dateTo || undefined,
      });

      if (response.success && response.data) {
        // Create blob and download
        const blob = new Blob([response.data], { type: 'text/csv' });
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `${activeTab}_${new Date().toISOString().split('T')[0]}.csv`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);

        toast({
          title: 'Success',
          description: `${activeTab === 'subscriptions' ? 'Subscriptions' : 'Billings'} exported successfully`,
        });
      }
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.error || 'Failed to export data',
        variant: 'destructive',
      });
    } finally {
      setExporting(false);
    }
  };

  const getStatusBadge = (status: string) => {
    const statusConfig = {
      active: { variant: 'default' as const, icon: CheckCircle2, label: 'Active' },
      paused: { variant: 'secondary' as const, icon: Clock, label: 'Paused' },
      cancelled: { variant: 'destructive' as const, icon: XCircle, label: 'Cancelled' },
    };

    const config = statusConfig[status as keyof typeof statusConfig] || statusConfig.active;
    const Icon = config.icon;

    return (
      <Badge variant={config.variant} className="flex items-center gap-1">
        <Icon className="h-3 w-3" />
        {config.label}
      </Badge>
    );
  };

  const getBillingStatusBadge = (status: string | undefined) => {
    const statusConfig = {
      success: { variant: 'default' as const, icon: CheckCircle2, label: 'Success' },
      failed: { variant: 'destructive' as const, icon: XCircle, label: 'Failed' },
      pending: { variant: 'secondary' as const, icon: Clock, label: 'Pending' },
    };

    const config = statusConfig[status as keyof typeof statusConfig] || statusConfig.pending;
    const Icon = config.icon;

    return (
      <Badge variant={config.variant} className="flex items-center gap-1">
        <Icon className="h-3 w-3" />
        {config.label}
      </Badge>
    );
  };

  const filteredSubscriptions = subscriptions.filter((subscription) => {
    if (searchTerm) {
      const search = searchTerm.toLowerCase();
      return (
        subscription.msisdn.includes(search) ||
        subscription.tier_name.toLowerCase().includes(search)
      );
    }
    return true;
  });

  const filteredBillings = billings.filter((billing) => {
    if (searchTerm) {
      const search = searchTerm.toLowerCase();
      return (
        billing.msisdn.includes(search) ||
        billing.payment_reference?.toLowerCase().includes(search)
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
          <h2 className="text-2xl font-bold">Daily Subscription Monitoring</h2>
          <p className="text-muted-foreground">
            Track subscription performance, revenue, and billing status
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
                Active Subscriptions
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between">
                <p className="text-2xl font-bold">{statistics.active_subscriptions}</p>
                <Users className="h-8 w-8 text-blue-600" />
              </div>
              <div className="flex items-center gap-1 mt-2 text-sm">
                {(statistics.subscription_growth?.growth_rate ?? 0) >= 0 ? (
                  <TrendingUp className="h-4 w-4 text-green-600" />
                ) : (
                  <TrendingDown className="h-4 w-4 text-red-600" />
                )}
                <span className={(statistics.subscription_growth?.growth_rate ?? 0) >= 0 ? 'text-green-600' : 'text-red-600'}>
                  {Math.abs(statistics.subscription_growth?.growth_rate ?? 0)}%
                </span>
                <span className="text-muted-foreground">vs last period</span>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Daily Revenue
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between">
                <p className="text-2xl font-bold">
                  ₦{statistics.daily_revenue?.toLocaleString()}
                </p>
                <DollarSign className="h-8 w-8 text-green-600" />
              </div>
              <div className="flex items-center gap-1 mt-2 text-sm">
                {(statistics.revenue_growth?.growth_rate ?? 0) >= 0 ? (
                  <TrendingUp className="h-4 w-4 text-green-600" />
                ) : (
                  <TrendingDown className="h-4 w-4 text-red-600" />
                )}
                <span className={(statistics.revenue_growth?.growth_rate ?? 0) >= 0 ? 'text-green-600' : 'text-red-600'}>
                  {Math.abs(statistics.revenue_growth?.growth_rate ?? 0)}%
                </span>
                <span className="text-muted-foreground">vs last period</span>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Total Entries
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between">
                <p className="text-2xl font-bold">{statistics.total_entries}</p>
                <Gift className="h-8 w-8 text-purple-600" />
              </div>
              <p className="text-sm text-muted-foreground mt-2">
                Draw entries from subscriptions
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Churn Rate
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between">
                <p className="text-2xl font-bold">{statistics.churn_rate?.toFixed(1)}%</p>
                {(statistics.churn_rate ?? 0) > 10 ? (
                  <AlertTriangle className="h-8 w-8 text-orange-600" />
                ) : (
                  <CheckCircle2 className="h-8 w-8 text-green-600" />
                )}
              </div>
              <p className="text-sm text-muted-foreground mt-2">
                Cancelled in last 30 days
              </p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Tier Performance */}
      {statistics?.tier_performance && (
        <Card>
          <CardHeader>
            <CardTitle>Tier Performance</CardTitle>
            <CardDescription>
              Subscription performance by tier
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {Object.entries(statistics.tier_performance).map(([tierName, data]: [string, any]) => (
                <div key={tierName} className="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
                  <div className="flex-1">
                    <div className="flex items-center gap-3 mb-2">
                      <h4 className="font-semibold">{tierName}</h4>
                      <Badge variant="outline">{data.entries} entries</Badge>
                    </div>
                    <div className="flex items-center gap-6 text-sm text-muted-foreground">
                      <div className="flex items-center gap-2">
                        <Users className="h-4 w-4" />
                        <span>{data.subscribers} subscribers</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <DollarSign className="h-4 w-4" />
                        <span>₦{data.revenue?.toLocaleString()} revenue</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <TrendingUp className="h-4 w-4" />
                        <span>{data.conversion_rate?.toFixed(1)}% conversion</span>
                      </div>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="text-2xl font-bold">{data.percentage?.toFixed(1)}%</p>
                    <p className="text-sm text-muted-foreground">of total</p>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Tabs */}
      <div className="flex gap-2 border-b">
        <button
          className={`px-4 py-2 font-medium transition-colors ${
            activeTab === 'subscriptions'
              ? 'border-b-2 border-primary text-primary'
              : 'text-muted-foreground hover:text-foreground'
          }`}
          onClick={() => setActiveTab('subscriptions')}
        >
          Subscriptions
        </button>
        <button
          className={`px-4 py-2 font-medium transition-colors ${
            activeTab === 'billings'
              ? 'border-b-2 border-primary text-primary'
              : 'text-muted-foreground hover:text-foreground'
          }`}
          onClick={() => setActiveTab('billings')}
        >
          Billings
        </button>
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="pt-6">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            {activeTab === 'subscriptions' ? (
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
                    <SelectItem value="active">Active</SelectItem>
                    <SelectItem value="paused">Paused</SelectItem>
                    <SelectItem value="cancelled">Cancelled</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            ) : (
              <div>
                <Label htmlFor="billing-status-filter">Billing Status</Label>
                <Select
                  value={billingStatusFilter}
                  onValueChange={(value) => setBillingStatusFilter(value as BillingStatusFilter)}
                >
                  <SelectTrigger id="billing-status-filter">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Statuses</SelectItem>
                    <SelectItem value="success">Success</SelectItem>
                    <SelectItem value="failed">Failed</SelectItem>
                    <SelectItem value="pending">Pending</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            )}

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
                placeholder={activeTab === 'subscriptions' ? 'MSISDN or Tier' : 'MSISDN or Reference'}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Subscriptions Table */}
      {activeTab === 'subscriptions' && (
        <Card>
          <CardHeader>
            <CardTitle>Subscriptions ({filteredSubscriptions.length})</CardTitle>
            <CardDescription>
              Active and historical subscription records
            </CardDescription>
          </CardHeader>
          <CardContent>
            {filteredSubscriptions.length === 0 ? (
              <p className="text-center text-muted-foreground py-8">
                No subscriptions found matching the current filters
              </p>
            ) : (
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>MSISDN</TableHead>
                      <TableHead>Tier</TableHead>
                      <TableHead>Bundles</TableHead>
                      <TableHead>Daily Cost</TableHead>
                      <TableHead>Entries/Day</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Start Date</TableHead>
                      <TableHead>Next Billing</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredSubscriptions.map((subscription) => (
                      <TableRow key={subscription.id}>
                        <TableCell className="font-mono">
                          {subscription.msisdn}
                        </TableCell>
                        <TableCell className="font-medium">
                          {subscription.tier_name}
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline">{subscription.bundle_quantity}x</Badge>
                        </TableCell>
                        <TableCell className="font-medium">
                          ₦{subscription.daily_cost?.toLocaleString()}
                        </TableCell>
                        <TableCell>
                          <Badge variant="secondary">{subscription.entries_per_day} entries</Badge>
                        </TableCell>
                        <TableCell>
                          {getStatusBadge(subscription.status)}
                        </TableCell>
                        <TableCell>
                          {subscription.start_date ? new Date(subscription.start_date).toLocaleDateString() : 'N/A'}
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2 text-sm text-muted-foreground">
                            <Calendar className="h-3 w-3" />
                            {subscription.next_billing_date
                              ? new Date(subscription.next_billing_date).toLocaleDateString()
                              : 'N/A'}
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Billings Table */}
      {activeTab === 'billings' && (
        <Card>
          <CardHeader>
            <CardTitle>Billings ({filteredBillings.length})</CardTitle>
            <CardDescription>
              Daily billing records and payment status
            </CardDescription>
          </CardHeader>
          <CardContent>
            {filteredBillings.length === 0 ? (
              <p className="text-center text-muted-foreground py-8">
                No billings found matching the current filters
              </p>
            ) : (
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>MSISDN</TableHead>
                      <TableHead>Amount</TableHead>
                      <TableHead>Entries</TableHead>
                      <TableHead>Payment Reference</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Billing Date</TableHead>
                      <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredBillings.map((billing) => (
                      <TableRow key={billing.id}>
                        <TableCell className="font-mono">
                          {billing.msisdn}
                        </TableCell>
                        <TableCell className="font-medium">
                          ₦{billing.amount.toLocaleString()}
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline">{billing.entries_allocated} entries</Badge>
                        </TableCell>
                        <TableCell className="font-mono text-sm">
                          {billing.payment_reference || 'N/A'}
                        </TableCell>
                        <TableCell>
                          {getBillingStatusBadge(billing.billing_status)}
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2 text-sm text-muted-foreground">
                            <Calendar className="h-3 w-3" />
                            {new Date(billing.billing_date).toLocaleDateString()}
                          </div>
                        </TableCell>
                        <TableCell className="text-right">
                          {billing.billing_status === 'failed' && (
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => handleRetryBilling(billing.id)}
                            >
                              <RefreshCw className="h-4 w-4" />
                            </Button>
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
      )}
    </div>
  );
}
