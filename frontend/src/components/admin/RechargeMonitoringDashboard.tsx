import React, { useState, useEffect } from 'react';
import { rechargeMonitoringApi } from '@/lib/api-client-extensions';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useToast } from '@/hooks/useToast';
import { 
  Phone, 
  RefreshCw, 
  Search, 
  Filter,
  CheckCircle,
  XCircle,
  Clock,
  Loader2,
  AlertCircle,
  DollarSign,
  TrendingUp,
  Activity,
  Eye,
  RotateCcw,
  Ban,
  CheckCheck
} from 'lucide-react';

interface RechargeTransaction {
  id: string;
  msisdn: string;
  network_provider: string;
  recharge_type: string;
  amount: number;
  status: string;
  payment_reference: string;
  payment_method: string;
  created_at: string;
  completed_at?: string;
  failure_reason?: string;
  provider_reference?: string;
  retry_count?: number;
  user_id: string;
  customer_name?: string;
  customer_email?: string;
}

interface RechargeStats {
  total_today: number;
  success_today: number;
  failed_today: number;
  pending_today: number;
  total_amount_today: number;
  success_rate: number;
  avg_processing_time: number;
  stuck_count: number;
}

interface VTUTransaction {
  id: string;
  parent_transaction_id: string;
  status: string;
  provider_name?: string;
  provider_reference?: string;
  provider_response?: any;
  retry_count: number;
  max_retries: number;
  error_message?: string;
  created_at: string;
  completed_at?: string;
  failed_at?: string;
}

const RechargeMonitoringDashboard: React.FC = () => {
  const { toast } = useToast();
  const [transactions, setTransactions] = useState<RechargeTransaction[]>([]);
  const [stats, setStats] = useState<RechargeStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [selectedTransaction, setSelectedTransaction] = useState<RechargeTransaction | null>(null);
  const [vtuDetails, setVtuDetails] = useState<VTUTransaction | null>(null);
  const [detailsOpen, setDetailsOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState('ALL');
  const [networkFilter, setNetworkFilter] = useState('ALL');
  const [refreshing, setRefreshing] = useState(false);
  const [txPage, setTxPage] = useState(1);
  const [txTotal, setTxTotal] = useState(0);
  const TX_PAGE_SIZE = 20;

  useEffect(() => {
    loadData();
    // Auto-refresh every 30 seconds
    const interval = setInterval(loadData, 30000);
    return () => clearInterval(interval);
  }, [statusFilter, networkFilter, txPage]);

  const loadData = async () => {
    try {
      setRefreshing(true);
      
      // Load stats
      const statsResponse = await rechargeMonitoringApi.getStats();
      
      if (statsResponse.success) {
        setStats(statsResponse.data);
      }

      // Load transactions
      const params = new URLSearchParams({
        limit: '100',
        ...(statusFilter !== 'ALL' && { status: statusFilter }),
        ...(networkFilter !== 'ALL' && { network: networkFilter }),
        ...(searchQuery && { search: searchQuery })
      });

      const transactionsResponse = await rechargeMonitoringApi.getTransactions({
        page: txPage,
        per_page: TX_PAGE_SIZE,
        status: statusFilter !== 'ALL' ? statusFilter : undefined,
        network: networkFilter !== 'ALL' ? networkFilter : undefined,
        search: searchQuery || undefined,
      });

      if (transactionsResponse.success) {
        setTransactions(transactionsResponse.data);
        setTxTotal(transactionsResponse.pagination?.total ?? transactionsResponse.total ?? transactionsResponse.data?.length ?? 0);
      }
    } catch (error) {
      console.error('Error loading recharge data:', error);
      toast({
        title: 'Error',
        description: 'Failed to load recharge data',
        variant: 'destructive'
      });
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  const viewDetails = async (transaction: RechargeTransaction) => {
    setSelectedTransaction(transaction);
    setDetailsOpen(true);
    
    // Load VTU transaction details
    try {
      const response = await rechargeMonitoringApi.getTransactionDetails(transaction.id);
      
      if (response.success) {
        setVtuDetails(response.data);
      }
    } catch (error) {
      console.error('Error loading VTU details:', error);
    }
  };

  const retryTransaction = async (transactionId: string) => {
    try {
      const response = await rechargeMonitoringApi.retryTransaction(transactionId);

      if (response.success) {
        toast({
          title: 'Success',
          description: 'Transaction retry initiated'
        });
        loadData();
        setDetailsOpen(false);
      } else {
        throw new Error(response.error);
      }
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.message || 'Failed to retry transaction',
        variant: 'destructive'
      });
    }
  };

  const refundTransaction = async (transactionId: string) => {
    if (!confirm('Are you sure you want to refund this transaction?')) return;

    try {
      const response = await rechargeMonitoringApi.refundTransaction(transactionId);

      if (response.success) {
        toast({
          title: 'Success',
          description: 'Refund initiated successfully'
        });
        loadData();
        setDetailsOpen(false);
      } else {
        throw new Error(response.error);
      }
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.message || 'Failed to initiate refund',
        variant: 'destructive'
      });
    }
  };

  const markAsSuccess = async (transactionId: string) => {
    if (!confirm('Are you sure you want to mark this as successful? This should only be done if you have verified the recharge was completed.')) return;

    try {
      const response = await rechargeMonitoringApi.markSuccess(transactionId);

      if (response.success) {
        toast({
          title: 'Success',
          description: 'Transaction marked as successful'
        });
        loadData();
        setDetailsOpen(false);
      } else {
        throw new Error(response.error);
      }
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.message || 'Failed to mark as successful',
        variant: 'destructive'
      });
    }
  };

  const markAsFailed = async (transactionId: string) => {
    if (!confirm('Are you sure you want to mark this as failed?')) return;

    try {
      const response = await rechargeMonitoringApi.markFailed(transactionId);

      if (response.success) {
        toast({
          title: 'Success',
          description: 'Transaction marked as failed'
        });
        loadData();
        setDetailsOpen(false);
      } else {
        throw new Error(response.error);
      }
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.message || 'Failed to mark as failed',
        variant: 'destructive'
      });
    }
  };

  const getStatusBadge = (status: string) => {
    const statusConfig: Record<string, { variant: any; icon: any; label: string }> = {
      SUCCESS: { variant: 'default', icon: CheckCircle, label: 'Success' },
      COMPLETED: { variant: 'default', icon: CheckCircle, label: 'Completed' },
      PENDING: { variant: 'secondary', icon: Clock, label: 'Pending' },
      PROCESSING: { variant: 'secondary', icon: Loader2, label: 'Processing' },
      FAILED: { variant: 'destructive', icon: XCircle, label: 'Failed' },
      REVERSED: { variant: 'outline', icon: RotateCcw, label: 'Reversed' }
    };

    const config = statusConfig[status] || { variant: 'outline', icon: AlertCircle, label: status };
    const Icon = config.icon;

    return (
      <Badge variant={config.variant} className="flex items-center gap-1">
        <Icon className="w-3 h-3" />
        {config.label}
      </Badge>
    );
  };

  const formatAmount = (amount: number) => {
    return `₦${(amount / 100).toLocaleString('en-NG', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('en-NG');
  };

  const getProcessingTime = (created: string, completed?: string) => {
    if (!completed) return 'N/A';
    const diff = new Date(completed).getTime() - new Date(created).getTime();
    return `${(diff / 1000).toFixed(1)}s`;
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Stats Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Today</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.total_today || 0}</div>
            <p className="text-xs text-muted-foreground">
              {formatAmount(stats?.total_amount_today || 0)} volume
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Success Rate</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.success_rate?.toFixed(1) || 0}%</div>
            <p className="text-xs text-muted-foreground">
              {stats?.success_today || 0} successful
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Avg Processing</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.avg_processing_time?.toFixed(1) || 0}s</div>
            <p className="text-xs text-muted-foreground">
              Average time
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Stuck Transactions</CardTitle>
            <AlertCircle className="h-4 w-4 text-destructive" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-destructive">{stats?.stuck_count || 0}</div>
            <p className="text-xs text-muted-foreground">
              Pending &gt;1 hour
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Filters and Search */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Recharge Transactions</CardTitle>
              <CardDescription>Monitor and manage all recharge transactions</CardDescription>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={loadData}
              disabled={refreshing}
            >
              <RefreshCw className={`w-4 h-4 mr-2 ${refreshing ? 'animate-spin' : ''}`} />
              Refresh
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex gap-4 mb-4">
            <div className="flex-1">
              <div className="relative">
                <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search by phone, reference, or user..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  onKeyPress={(e) => e.key === 'Enter' && loadData()}
                  className="pl-8"
                />
              </div>
            </div>
            <Select value={statusFilter} onValueChange={(v) => { setTxPage(1); setStatusFilter(v); }}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="ALL">All Status</SelectItem>
                <SelectItem value="SUCCESS">Success</SelectItem>
                <SelectItem value="PENDING">Pending</SelectItem>
                <SelectItem value="PROCESSING">Processing</SelectItem>
                <SelectItem value="FAILED">Failed</SelectItem>
                <SelectItem value="REVERSED">Reversed</SelectItem>
              </SelectContent>
            </Select>
            <Select value={networkFilter} onValueChange={(v) => { setTxPage(1); setNetworkFilter(v); }}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Network" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="ALL">All Networks</SelectItem>
                <SelectItem value="MTN">MTN</SelectItem>
                <SelectItem value="AIRTEL">Airtel</SelectItem>
                <SelectItem value="GLO">Glo</SelectItem>
                <SelectItem value="NINE_MOBILE">9mobile</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Phone</TableHead>
                  <TableHead>Network</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Amount</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Time</TableHead>
                  <TableHead>Processing</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {transactions.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={8} className="text-center text-muted-foreground">
                      No transactions found
                    </TableCell>
                  </TableRow>
                ) : (
                  transactions.map((transaction) => (
                    <TableRow key={transaction.id}>
                      <TableCell className="font-medium">
                        <div className="flex items-center gap-2">
                          <Phone className="w-4 h-4 text-muted-foreground" />
                          {transaction.msisdn}
                        </div>
                      </TableCell>
                      <TableCell>{transaction.network_provider}</TableCell>
                      <TableCell>
                        <Badge variant="outline">{transaction.recharge_type}</Badge>
                      </TableCell>
                      <TableCell>{formatAmount(transaction.amount)}</TableCell>
                      <TableCell>{getStatusBadge(transaction.status)}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {formatDate(transaction.created_at)}
                      </TableCell>
                      <TableCell>
                        {getProcessingTime(transaction.created_at, transaction.completed_at)}
                      </TableCell>
                      <TableCell>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => viewDetails(transaction)}
                        >
                          <Eye className="w-4 h-4" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
          {/* Pagination controls */}
          <div className="flex items-center justify-between mt-4 px-1">
            <p className="text-xs text-muted-foreground">
              Page {txPage} · Showing {transactions.length} of {txTotal} transactions
            </p>
            <div className="flex gap-2">
              <button
                className="px-3 py-1 text-xs border rounded disabled:opacity-40"
                disabled={txPage <= 1}
                onClick={() => setTxPage(p => Math.max(1, p - 1))}
              >
                ← Prev
              </button>
              <button
                className="px-3 py-1 text-xs border rounded disabled:opacity-40"
                disabled={transactions.length < TX_PAGE_SIZE}
                onClick={() => setTxPage(p => p + 1)}
              >
                Next →
              </button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Transaction Details Dialog */}
      <Dialog open={detailsOpen} onOpenChange={setDetailsOpen}>
        <DialogContent className="max-w-3xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Transaction Details</DialogTitle>
            <DialogDescription>
              Complete information and actions for this recharge
            </DialogDescription>
          </DialogHeader>

          {selectedTransaction && (
            <div className="space-y-6">
              {/* Basic Info */}
              <Card>
                <CardHeader>
                  <CardTitle className="text-lg">Basic Information</CardTitle>
                </CardHeader>
                <CardContent className="grid gap-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <p className="text-sm font-medium text-muted-foreground">Transaction ID</p>
                      <p className="text-sm font-mono">{selectedTransaction.id}</p>
                    </div>
                    <div>
                      <p className="text-sm font-medium text-muted-foreground">Payment Reference</p>
                      <p className="text-sm font-mono">{selectedTransaction.payment_reference}</p>
                    </div>
                    <div>
                      <p className="text-sm font-medium text-muted-foreground">Phone Number</p>
                      <p className="text-sm">{selectedTransaction.msisdn}</p>
                    </div>
                    <div>
                      <p className="text-sm font-medium text-muted-foreground">Network</p>
                      <p className="text-sm">{selectedTransaction.network_provider}</p>
                    </div>
                    <div>
                      <p className="text-sm font-medium text-muted-foreground">Type</p>
                      <Badge variant="outline">{selectedTransaction.recharge_type}</Badge>
                    </div>
                    <div>
                      <p className="text-sm font-medium text-muted-foreground">Amount</p>
                      <p className="text-sm font-bold">{formatAmount(selectedTransaction.amount)}</p>
                    </div>
                    <div>
                      <p className="text-sm font-medium text-muted-foreground">Status</p>
                      {getStatusBadge(selectedTransaction.status)}
                    </div>
                    <div>
                      <p className="text-sm font-medium text-muted-foreground">Payment Method</p>
                      <p className="text-sm">{selectedTransaction.payment_method}</p>
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* VTU Details */}
              {vtuDetails && (
                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg">VTU Provider Details</CardTitle>
                  </CardHeader>
                  <CardContent className="grid gap-4">
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <p className="text-sm font-medium text-muted-foreground">Provider</p>
                        <p className="text-sm">{vtuDetails.provider_name || 'N/A'}</p>
                      </div>
                      <div>
                        <p className="text-sm font-medium text-muted-foreground">Provider Reference</p>
                        <p className="text-sm font-mono">{vtuDetails.provider_reference || 'N/A'}</p>
                      </div>
                      <div>
                        <p className="text-sm font-medium text-muted-foreground">Retry Count</p>
                        <p className="text-sm">{vtuDetails.retry_count} / {vtuDetails.max_retries}</p>
                      </div>
                      <div>
                        <p className="text-sm font-medium text-muted-foreground">VTU Status</p>
                        {getStatusBadge(vtuDetails.status)}
                      </div>
                    </div>
                    {vtuDetails.error_message && (
                      <div>
                        <p className="text-sm font-medium text-muted-foreground mb-1">Error Message</p>
                        <p className="text-sm text-destructive bg-destructive/10 p-2 rounded">
                          {vtuDetails.error_message}
                        </p>
                      </div>
                    )}
                  </CardContent>
                </Card>
              )}

              {/* Timestamps */}
              <Card>
                <CardHeader>
                  <CardTitle className="text-lg">Timeline</CardTitle>
                </CardHeader>
                <CardContent className="grid gap-2">
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Created:</span>
                    <span>{formatDate(selectedTransaction.created_at)}</span>
                  </div>
                  {selectedTransaction.completed_at && (
                    <div className="flex justify-between text-sm">
                      <span className="text-muted-foreground">Completed:</span>
                      <span>{formatDate(selectedTransaction.completed_at)}</span>
                    </div>
                  )}
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Processing Time:</span>
                    <span>{getProcessingTime(selectedTransaction.created_at, selectedTransaction.completed_at)}</span>
                  </div>
                </CardContent>
              </Card>

              {/* Actions */}
              <div className="flex gap-2 justify-end">
                {selectedTransaction.status === 'FAILED' && (
                  <>
                    <Button
                      variant="outline"
                      onClick={() => retryTransaction(selectedTransaction.id)}
                    >
                      <RotateCcw className="w-4 h-4 mr-2" />
                      Retry
                    </Button>
                    <Button
                      variant="outline"
                      onClick={() => refundTransaction(selectedTransaction.id)}
                    >
                      <DollarSign className="w-4 h-4 mr-2" />
                      Refund
                    </Button>
                  </>
                )}
                {selectedTransaction.status === 'PENDING' && (
                  <>
                    <Button
                      variant="outline"
                      onClick={() => markAsSuccess(selectedTransaction.id)}
                    >
                      <CheckCheck className="w-4 h-4 mr-2" />
                      Mark Success
                    </Button>
                    <Button
                      variant="outline"
                      onClick={() => markAsFailed(selectedTransaction.id)}
                    >
                      <Ban className="w-4 h-4 mr-2" />
                      Mark Failed
                    </Button>
                  </>
                )}
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default RechargeMonitoringDashboard;
