import React, { useState, useEffect } from 'react';
import { apiClient } from '@/lib/api-client';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Progress } from '@/components/ui/progress';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  CheckCircle,
  XCircle,
  TrendingUp,
  RefreshCw,
  AlertCircle,
  Info,
  Shield,
  Database,
  Wifi
} from 'lucide-react';



interface ValidationStats {
  summary: {
    total_validations: number;
    successful_validations: number;
    failed_validations: number;
    success_rate: number;
    mismatch_rate: number;
  };
  validation_sources: {
    hlr_api_count: number;
    prefix_count: number;
    cache_count: number;
    hlr_api_percentage: number;
    prefix_percentage: number;
    cache_percentage: number;
  };
  by_network: Array<{
    network: string;
    total_validations: number;
    successful_validations: number;
    failed_validations: number;
    success_rate: number;
    common_mismatches: string[];
  }>;
  mismatch_patterns: Array<{
    selected_network: string;
    actual_network: string;
    count: number;
    percentage: number;
  }>;
  validation_trend: Array<{
    date: string;
    total_validations: number;
    successful_validations: number;
    failed_validations: number;
    success_rate: number;
  }>;
}

const ValidationStatsDashboard: React.FC = () => {
  const [startDate, setStartDate] = useState(
    new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0]
  );
  const [endDate, setEndDate] = useState(new Date().toISOString().split('T')[0]);
  const [isLoading, setIsLoading] = useState(false);
  const [stats, setStats] = useState<ValidationStats | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadStats();
  }, []);

  const loadStats = async () => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await apiClient.post('/admin/validation/stats', {
        start_date: startDate,
        end_date: endDate,
      });

      const result = response.data;
      if (result.success) {
        setStats(result.data);
      } else {
        throw new Error(result.error || 'Failed to load data');
      }
    } catch (err: any) {
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Network Validation Statistics</h1>
          <p className="text-gray-600 mt-1">
            Monitor validation accuracy and identify mismatch patterns
          </p>
        </div>
      </div>

      {/* Date Range Filter */}
      <Card>
        <CardHeader>
          <CardTitle>Date Range</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-4 items-end">
            <div className="space-y-2 flex-1">
              <Label htmlFor="startDate">Start Date</Label>
              <Input
                id="startDate"
                type="date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
              />
            </div>
            <div className="space-y-2 flex-1">
              <Label htmlFor="endDate">End Date</Label>
              <Input
                id="endDate"
                type="date"
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
              />
            </div>
            <Button onClick={loadStats} disabled={isLoading}>
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

      {stats && (
        <>
          {/* Summary Cards */}
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-gray-600 flex items-center gap-2">
                  <Shield className="w-4 h-4" />
                  Total Validations
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {stats.summary.total_validations.toLocaleString()}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-gray-600 flex items-center gap-2">
                  <CheckCircle className="w-4 h-4 text-green-600" />
                  Successful
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-green-600">
                  {stats.summary.successful_validations.toLocaleString()}
                </div>
                <Progress value={stats.summary.success_rate} className="mt-2" />
                <p className="text-xs text-gray-600 mt-1">
                  {stats.summary.success_rate.toFixed(1)}% success rate
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-gray-600 flex items-center gap-2">
                  <XCircle className="w-4 h-4 text-red-600" />
                  Failed
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-red-600">
                  {stats.summary.failed_validations.toLocaleString()}
                </div>
                <Progress value={stats.summary.mismatch_rate} className="mt-2" />
                <p className="text-xs text-gray-600 mt-1">
                  {stats.summary.mismatch_rate.toFixed(1)}% mismatch rate
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-gray-600 flex items-center gap-2">
                  <TrendingUp className="w-4 h-4" />
                  Performance
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {stats.summary.success_rate >= 95 ? 'Excellent' : stats.summary.success_rate >= 90 ? 'Good' : 'Needs Improvement'}
                </div>
                <Badge
                  variant={stats.summary.success_rate >= 95 ? 'default' : stats.summary.success_rate >= 90 ? 'secondary' : 'destructive'}
                  className="mt-2"
                >
                  {stats.summary.success_rate.toFixed(1)}%
                </Badge>
              </CardContent>
            </Card>
          </div>

          {/* Validation Sources */}
          <Card>
            <CardHeader>
              <CardTitle>Validation Sources</CardTitle>
              <CardDescription>
                Breakdown of validation methods used
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-2">
                      <Database className="w-4 h-4" />
                      <span className="font-medium">HLR API</span>
                    </div>
                    <span className="text-sm text-gray-600">
                      {stats.validation_sources.hlr_api_count.toLocaleString()} ({stats.validation_sources.hlr_api_percentage.toFixed(1)}%)
                    </span>
                  </div>
                  <Progress value={stats.validation_sources.hlr_api_percentage} />
                </div>

                <div>
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-2">
                      <Wifi className="w-4 h-4" />
                      <span className="font-medium">Prefix Detection</span>
                    </div>
                    <span className="text-sm text-gray-600">
                      {stats.validation_sources.prefix_count.toLocaleString()} ({stats.validation_sources.prefix_percentage.toFixed(1)}%)
                    </span>
                  </div>
                  <Progress value={stats.validation_sources.prefix_percentage} />
                </div>

                <div>
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-2">
                      <Database className="w-4 h-4" />
                      <span className="font-medium">Cache (Transaction History)</span>
                    </div>
                    <span className="text-sm text-gray-600">
                      {stats.validation_sources.cache_count.toLocaleString()} ({stats.validation_sources.cache_percentage.toFixed(1)}%)
                    </span>
                  </div>
                  <Progress value={stats.validation_sources.cache_percentage} />
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Validation by Network */}
          <Card>
            <CardHeader>
              <CardTitle>Validation Success by Network</CardTitle>
              <CardDescription>
                Success rates for each network provider
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Network</TableHead>
                    <TableHead className="text-right">Total</TableHead>
                    <TableHead className="text-right">Successful</TableHead>
                    <TableHead className="text-right">Failed</TableHead>
                    <TableHead className="text-right">Success Rate</TableHead>
                    <TableHead>Common Mismatches</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {stats.by_network.map((item) => (
                    <TableRow key={item.network}>
                      <TableCell className="font-medium">
                        <Badge variant="outline">{item.network}</Badge>
                      </TableCell>
                      <TableCell className="text-right">
                        {item.total_validations.toLocaleString()}
                      </TableCell>
                      <TableCell className="text-right text-green-600">
                        {item.successful_validations.toLocaleString()}
                      </TableCell>
                      <TableCell className="text-right text-red-600">
                        {item.failed_validations.toLocaleString()}
                      </TableCell>
                      <TableCell className="text-right">
                        <Badge
                          variant={item.success_rate >= 95 ? 'default' : item.success_rate >= 90 ? 'secondary' : 'destructive'}
                        >
                          {item.success_rate.toFixed(1)}%
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <div className="flex gap-1">
                          {item.common_mismatches.map((network) => (
                            <Badge key={network} variant="outline" className="text-xs">
                              {network}
                            </Badge>
                          ))}
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>

          {/* Mismatch Patterns */}
          <Card>
            <CardHeader>
              <CardTitle>Common Mismatch Patterns</CardTitle>
              <CardDescription>
                Most frequent network selection errors
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Selected Network</TableHead>
                    <TableHead>Actual Network</TableHead>
                    <TableHead className="text-right">Count</TableHead>
                    <TableHead className="text-right">Percentage</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {stats.mismatch_patterns.map((pattern, index) => (
                    <TableRow key={index}>
                      <TableCell>
                        <Badge variant="destructive">{pattern.selected_network}</Badge>
                      </TableCell>
                      <TableCell>
                        <Badge variant="default">{pattern.actual_network}</Badge>
                      </TableCell>
                      <TableCell className="text-right font-medium">
                        {pattern.count.toLocaleString()}
                      </TableCell>
                      <TableCell className="text-right">
                        {pattern.percentage.toFixed(1)}%
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>

          {/* Insights and Recommendations */}
          <Alert>
            <Info className="h-4 w-4" />
            <AlertDescription>
              <strong>Insights & Recommendations:</strong>
              <ul className="list-disc list-inside mt-2 space-y-1">
                <li>
                  Success rate of {stats.summary.success_rate.toFixed(1)}% is {stats.summary.success_rate >= 95 ? 'excellent' : stats.summary.success_rate >= 90 ? 'good but can be improved' : 'below target (95%)'}
                </li>
                <li>
                  {stats.validation_sources.prefix_percentage > 70 ? 'Consider activating Termii HLR API for higher accuracy' : 'Good mix of validation sources'}
                </li>
                <li>
                  Most common mismatches: {stats.mismatch_patterns[0]?.selected_network} → {stats.mismatch_patterns[0]?.actual_network} ({stats.mismatch_patterns[0]?.count} cases)
                </li>
                <li>
                  {stats.summary.mismatch_rate < 5 ? 'Mismatch rate is within acceptable range' : 'Consider improving user education on network selection'}
                </li>
              </ul>
            </AlertDescription>
          </Alert>
        </>
      )}
    </div>
  );
};

export default ValidationStatsDashboard;
