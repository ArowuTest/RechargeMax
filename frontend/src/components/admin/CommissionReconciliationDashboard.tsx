import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  DollarSign,
  TrendingUp,
  Download,
  Calendar,
  Filter,
  RefreshCw,
  CheckCircle,
  AlertCircle,
  Info
} from 'lucide-react';

const API_BASE = import.meta.env.VITE_API_BASE_URL || '/api/v1';

interface CommissionSummary {
  total_transactions: number;
  total_recharge_amount: number;
  total_commission: number;
  average_commission: number;
  commission_rate: number;
}

interface NetworkCommission {
  network: string;
  transaction_count: number;
  total_amount: number;
  total_commission: number;
  average_commission: number;
  commission_rate: number;
}

interface ProviderCommission {
  provider: string;
  transaction_count: number;
  total_amount: number;
  total_commission: number;
  average_commission: number;
  commission_rate: number;
}

interface DailyCommission {
  date: string;
  transaction_count: number;
  total_amount: number;
  total_commission: number;
}

interface RecentTransaction {
  id: string;
  msisdn: string;
  network: string;
  provider: string;
  amount: number;
  commission: number;
  commission_rate: number;
  status: string;
  created_at: string;
}

interface CommissionData {
  summary: CommissionSummary;
  by_network: NetworkCommission[];
  by_provider: ProviderCommission[];
  by_date: DailyCommission[];
  recent_transactions: RecentTransaction[];
}

export const CommissionReconciliationDashboard: React.FC = () => {
  const [startDate, setStartDate] = useState(
    new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0]
  );
  const [endDate, setEndDate] = useState(new Date().toISOString().split('T')[0]);
  const [selectedNetwork, setSelectedNetwork] = useState<string>('');
  const [selectedProvider, setSelectedProvider] = useState<string>('');
  const [isLoading, setIsLoading] = useState(false);
  const [commissionData, setCommissionData] = useState<CommissionData | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadCommissionData();
  }, []);

  const loadCommissionData = async () => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await fetch(`${API_BASE}/admin/commissions/reconciliation`,  {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          start_date: startDate,
          end_date: endDate,
          network: selectedNetwork || undefined,
          provider: selectedProvider || undefined,
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to load commission data');
      }

      const result = await response.json();
      if (result.success) {
        setCommissionData(result.data);
      } else {
        throw new Error(result.error || 'Failed to load data');
      }
    } catch (err: any) {
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  const exportReport = async () => {
    try {
      const response = await fetch(`${API_BASE}/admin/commissions/export`,  {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          start_date: startDate,
          end_date: endDate,
          network: selectedNetwork || undefined,
          provider: selectedProvider || undefined,
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to export report');
      }

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `commission_report_${startDate}_to_${endDate}.csv`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch (err: any) {
      setError(err.message);
    }
  };

  const formatCurrency = (amount: number) => {
    return `₦${(amount / 100).toLocaleString('en-NG', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
  };

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Commission Reconciliation</h1>
          <p className="text-gray-600 mt-1">
            Track and reconcile commissions from network providers and VTU aggregators
          </p>
        </div>
        <Button onClick={exportReport} disabled={isLoading || !commissionData}>
          <Download className="w-4 h-4 mr-2" />
          Export CSV
        </Button>
      </div>

      {/* Filters */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Filter className="w-5 h-5" />
            Filters
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div className="space-y-2">
              <Label htmlFor="startDate">Start Date</Label>
              <Input
                id="startDate"
                type="date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="endDate">End Date</Label>
              <Input
                id="endDate"
                type="date"
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="network">Network</Label>
              <Select value={selectedNetwork} onValueChange={setSelectedNetwork}>
                <SelectTrigger>
                  <SelectValue placeholder="All Networks" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">All Networks</SelectItem>
                  <SelectItem value="MTN">MTN</SelectItem>
                  <SelectItem value="GLO">GLO</SelectItem>
                  <SelectItem value="AIRTEL">Airtel</SelectItem>
                  <SelectItem value="9MOBILE">9mobile</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="provider">Provider</Label>
              <Select value={selectedProvider} onValueChange={setSelectedProvider}>
                <SelectTrigger>
                  <SelectValue placeholder="All Providers" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">All Providers</SelectItem>
                  <SelectItem value="VTPass">VTPass</SelectItem>
                  <SelectItem value="MTN_DIRECT">MTN Direct</SelectItem>
                  <SelectItem value="GLO_DIRECT">GLO Direct</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="mt-4 flex gap-2">
            <Button onClick={loadCommissionData} disabled={isLoading}>
              {isLoading ? (
                <>
                  <RefreshCw className="w-4 h-4 mr-2 animate-spin" />
                  Loading...
                </>
              ) : (
                <>
                  <RefreshCw className="w-4 h-4 mr-2" />
                  Refresh
                </>
              )}
            </Button>
          </div>
        </CardContent>
      </Card>

      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {commissionData && (
        <>
          {/* Summary Cards */}
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-gray-600">
                  Total Transactions
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {commissionData.summary.total_transactions.toLocaleString()}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-gray-600">
                  Total Recharge Amount
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {formatCurrency(commissionData.summary.total_recharge_amount)}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-gray-600">
                  Total Commission Earned
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-green-600">
                  {formatCurrency(commissionData.summary.total_commission)}
                </div>
                <p className="text-xs text-gray-600 mt-1">
                  Avg: {formatCurrency(commissionData.summary.average_commission)} per transaction
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-gray-600">
                  Average Commission Rate
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {commissionData.summary.commission_rate.toFixed(2)}%
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Commission by Network */}
          <Card>
            <CardHeader>
              <CardTitle>Commission by Network</CardTitle>
              <CardDescription>
                Breakdown of commissions earned from each network provider
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Network</TableHead>
                    <TableHead className="text-right">Transactions</TableHead>
                    <TableHead className="text-right">Total Amount</TableHead>
                    <TableHead className="text-right">Commission</TableHead>
                    <TableHead className="text-right">Rate</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {commissionData.by_network.map((item) => (
                    <TableRow key={item.network}>
                      <TableCell className="font-medium">
                        <Badge variant="outline">{item.network}</Badge>
                      </TableCell>
                      <TableCell className="text-right">
                        {item.transaction_count.toLocaleString()}
                      </TableCell>
                      <TableCell className="text-right">
                        {formatCurrency(item.total_amount)}
                      </TableCell>
                      <TableCell className="text-right font-semibold text-green-600">
                        {formatCurrency(item.total_commission)}
                      </TableCell>
                      <TableCell className="text-right">
                        {item.commission_rate.toFixed(2)}%
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>

          {/* Commission by Provider */}
          <Card>
            <CardHeader>
              <CardTitle>Commission by Provider</CardTitle>
              <CardDescription>
                Breakdown of commissions from VTU aggregators and direct integrations
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Provider</TableHead>
                    <TableHead className="text-right">Transactions</TableHead>
                    <TableHead className="text-right">Total Amount</TableHead>
                    <TableHead className="text-right">Commission</TableHead>
                    <TableHead className="text-right">Rate</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {commissionData.by_provider.map((item) => (
                    <TableRow key={item.provider}>
                      <TableCell className="font-medium">
                        <Badge>{item.provider}</Badge>
                      </TableCell>
                      <TableCell className="text-right">
                        {item.transaction_count.toLocaleString()}
                      </TableCell>
                      <TableCell className="text-right">
                        {formatCurrency(item.total_amount)}
                      </TableCell>
                      <TableCell className="text-right font-semibold text-green-600">
                        {formatCurrency(item.total_commission)}
                      </TableCell>
                      <TableCell className="text-right">
                        {item.commission_rate.toFixed(2)}%
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>

          {/* Recent Transactions */}
          <Card>
            <CardHeader>
              <CardTitle>Recent Transactions with Commission</CardTitle>
              <CardDescription>
                Latest transactions showing commission breakdown
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Date</TableHead>
                    <TableHead>Phone</TableHead>
                    <TableHead>Network</TableHead>
                    <TableHead>Provider</TableHead>
                    <TableHead className="text-right">Amount</TableHead>
                    <TableHead className="text-right">Commission</TableHead>
                    <TableHead className="text-right">Rate</TableHead>
                    <TableHead>Status</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {commissionData.recent_transactions.map((txn) => (
                    <TableRow key={txn.id}>
                      <TableCell className="text-sm">
                        {new Date(txn.created_at).toLocaleDateString()}
                      </TableCell>
                      <TableCell className="font-mono text-sm">{txn.msisdn}</TableCell>
                      <TableCell>
                        <Badge variant="outline">{txn.network}</Badge>
                      </TableCell>
                      <TableCell>
                        <Badge variant="secondary">{txn.provider}</Badge>
                      </TableCell>
                      <TableCell className="text-right">
                        {formatCurrency(txn.amount)}
                      </TableCell>
                      <TableCell className="text-right font-semibold text-green-600">
                        {formatCurrency(txn.commission)}
                      </TableCell>
                      <TableCell className="text-right text-sm">
                        {txn.commission_rate.toFixed(2)}%
                      </TableCell>
                      <TableCell>
                        <Badge
                          variant={txn.status === 'SUCCESS' ? 'default' : 'destructive'}
                        >
                          {txn.status}
                        </Badge>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>

          {/* Reconciliation Notes */}
          <Alert>
            <Info className="h-4 w-4" />
            <AlertDescription>
              <strong>Reconciliation Notes:</strong>
              <ul className="list-disc list-inside mt-2 space-y-1">
                <li>Commission rates may vary by network and provider agreement</li>
                <li>VTPass commissions are calculated at 3.5% of transaction amount</li>
                <li>Direct network integrations may have different commission structures</li>
                <li>Export CSV for detailed reconciliation with provider statements</li>
              </ul>
            </AlertDescription>
          </Alert>
        </>
      )}
    </div>
  );
};

export default CommissionReconciliationDashboard;
