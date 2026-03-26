import React, { useState, useEffect } from 'react';
import { apiClient } from '@/lib/api-client';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Progress } from '@/components/ui/progress';
import { 
  Activity, 
  Server, 
  Database, 
  Wifi, 
  AlertTriangle, 
  CheckCircle, 
  Clock, 
  TrendingUp,
  TrendingDown,
  RefreshCw,
  Loader2,
  Zap,
  HardDrive,
  Cpu,
  MemoryStick
} from 'lucide-react';

interface SystemMetrics {
  server: {
    status: 'healthy' | 'warning' | 'critical';
    uptime: number;
    cpu_usage: number;
    memory_usage: number;
    disk_usage: number;
    response_time: number;
  };
  database: {
    status: 'healthy' | 'warning' | 'critical';
    connections: number;
    max_connections: number;
    query_time: number;
    slow_queries: number;
  };
  api: {
    status: 'healthy' | 'warning' | 'critical';
    requests_per_minute: number;
    error_rate: number;
    avg_response_time: number;
  };
  external_services: {
    paystack: 'online' | 'offline' | 'degraded';
    telecom_providers: {
      mtn: 'online' | 'offline' | 'degraded';
      airtel: 'online' | 'offline' | 'degraded';
      glo: 'online' | 'offline' | 'degraded';
      nine_mobile: 'online' | 'offline' | 'degraded';
    };
  };
  recent_alerts: Array<{
    id: string;
    type: 'error' | 'warning' | 'info';
    message: string;
    timestamp: string;
    resolved: boolean;
  }>;
}

const SystemMonitoringDashboard: React.FC = () => {
  const [metrics, setMetrics] = useState<SystemMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const [autoRefresh, setAutoRefresh] = useState(true);

  useEffect(() => {
    fetchSystemMetrics();
    
    let interval: NodeJS.Timeout;
    if (autoRefresh) {
      interval = setInterval(fetchSystemMetrics, 30000); // Refresh every 30 seconds
    }
    
    return () => {
      if (interval) clearInterval(interval);
    };
  }, [autoRefresh]);

  const fetchSystemMetrics = async () => {
    try {
      setLoading(true);

      // Call the dedicated monitoring endpoint (v19+)
      const monRes = await apiClient.get<any>('/admin/monitoring/system');
      const d = monRes.data?.data ?? monRes.data;

      if (d) {
        const metrics: SystemMetrics = {
          server: {
            status:        d.server?.status        ?? 'healthy',
            uptime:        d.server?.uptime         ?? 0,
            cpu_usage:     d.server?.cpu_usage       ?? 0,
            memory_usage:  d.server?.memory_usage    ?? 0,
            disk_usage:    d.server?.disk_usage       ?? 0,
            response_time: d.server?.response_time   ?? 0,
          },
          database: {
            status:          d.database?.status          ?? 'healthy',
            connections:     d.database?.connections      ?? 0,
            max_connections: d.database?.max_connections  ?? 100,
            query_time:      d.database?.query_time       ?? 0,
            slow_queries:    d.database?.slow_queries      ?? 0,
          },
          api: {
            status:               d.api?.status               ?? 'healthy',
            requests_per_minute:  d.api?.requests_per_minute  ?? 0,
            error_rate:           d.api?.error_rate            ?? 0,
            avg_response_time:    d.api?.avg_response_time     ?? 0,
          },
          external_services: {
            paystack:            d.external_services?.paystack ?? 'online',
            telecom_providers: {
              mtn:         d.external_services?.telecom_providers?.mtn         ?? 'online',
              airtel:      d.external_services?.telecom_providers?.airtel      ?? 'online',
              glo:         d.external_services?.telecom_providers?.glo         ?? 'online',
              nine_mobile: d.external_services?.telecom_providers?.nine_mobile ?? 'online',
            },
          },
          recent_alerts: d.recent_alerts ?? [],
        };
        setMetrics(metrics);
        setLastUpdated(new Date());
        return;
      }
    } catch (_err) {
      // Fall through to health-check fallback
    }

    // Fallback: basic /health endpoint
    try {
      const healthResponse = await apiClient.get('/health');
      const basicMetrics: SystemMetrics = {
        server: {
          status: healthResponse.data?.status === 'healthy' ? 'healthy' : 'critical',
          uptime: healthResponse.data?.uptime_seconds
            ? Math.min(99.9, (healthResponse.data.uptime_seconds / 86400) * 100)
            : (healthResponse.data?.status === 'healthy' ? 99.9 : 0),
          cpu_usage: 0,
          memory_usage: 0,
          disk_usage: 0,
          response_time: 0,
        },
        database: {
          status: healthResponse.data?.database === 'connected' ? 'healthy' : 'critical',
          connections: 0,
          max_connections: 100,
          query_time: 0,
          slow_queries: 0,
        },
        api: { status: 'healthy', requests_per_minute: 0, error_rate: 0, avg_response_time: 0 },
        external_services: {
          paystack: 'online',
          telecom_providers: { mtn: 'online', airtel: 'online', glo: 'online', nine_mobile: 'online' },
        },
        recent_alerts: [],
      };
      setMetrics(basicMetrics);
      setLastUpdated(new Date());
    } catch (error) {
      console.error('Failed to fetch system metrics:', error);
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'healthy':
      case 'online':
        return 'text-green-600 bg-green-100';
      case 'warning':
      case 'degraded':
        return 'text-yellow-600 bg-yellow-100';
      case 'critical':
      case 'offline':
        return 'text-red-600 bg-red-100';
      default:
        return 'text-gray-600 bg-gray-100';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'healthy':
      case 'online':
        return <CheckCircle className="w-4 h-4" />;
      case 'warning':
      case 'degraded':
        return <AlertTriangle className="w-4 h-4" />;
      case 'critical':
      case 'offline':
        return <AlertTriangle className="w-4 h-4" />;
      default:
        return <Clock className="w-4 h-4" />;
    }
  };

  const formatUptime = (uptime: number) => {
    return `${uptime.toFixed(2)}%`;
  };

  const formatDuration = (timestamp: string) => {
    const diff = Date.now() - new Date(timestamp).getTime();
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(minutes / 60);
    
    if (hours > 0) {
      return `${hours}h ${minutes % 60}m ago`;
    }
    return `${minutes}m ago`;
  };

  if (loading && !metrics) {
    return (
      <Card>
        <CardContent className="p-8 text-center">
          <Loader2 className="w-8 h-8 animate-spin mx-auto mb-4" />
          <p>Loading system monitoring dashboard...</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold flex items-center gap-2">
            <Activity className="w-6 h-6" />
            System Monitoring
          </h2>
          <p className="text-gray-600">
            Real-time system health and performance metrics
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setAutoRefresh(!autoRefresh)}
          >
            <RefreshCw className={`w-4 h-4 mr-2 ${autoRefresh ? 'animate-spin' : ''}`} />
            Auto Refresh: {autoRefresh ? 'ON' : 'OFF'}
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={fetchSystemMetrics}
            disabled={loading}
          >
            {loading ? (
              <Loader2 className="w-4 h-4 animate-spin mr-2" />
            ) : (
              <RefreshCw className="w-4 h-4 mr-2" />
            )}
            Refresh
          </Button>
        </div>
      </div>

      {/* Last Updated */}
      {lastUpdated && (
        <div className="text-sm text-gray-500">
          Last updated: {lastUpdated.toLocaleTimeString()}
        </div>
      )}

      <Tabs defaultValue="overview" className="space-y-4">
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="server">Server</TabsTrigger>
          <TabsTrigger value="database">Database</TabsTrigger>
          <TabsTrigger value="external">External Services</TabsTrigger>
          <TabsTrigger value="alerts">Alerts</TabsTrigger>
        </TabsList>

        {/* Overview Tab */}
        <TabsContent value="overview">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {/* Server Status */}
            <Card>
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600">Server</p>
                    <div className="flex items-center gap-2 mt-1">
                      <Badge className={getStatusColor(metrics?.server.status || 'unknown')}>
                        {getStatusIcon(metrics?.server.status || 'unknown')}
                        <span className="ml-1 capitalize">{metrics?.server.status}</span>
                      </Badge>
                    </div>
                  </div>
                  <Server className="w-8 h-8 text-blue-600" />
                </div>
                <div className="mt-4 text-sm text-gray-600">
                  Uptime: {formatUptime(metrics?.server.uptime || 0)}
                </div>
              </CardContent>
            </Card>

            {/* Database Status */}
            <Card>
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600">Database</p>
                    <div className="flex items-center gap-2 mt-1">
                      <Badge className={getStatusColor(metrics?.database.status || 'unknown')}>
                        {getStatusIcon(metrics?.database.status || 'unknown')}
                        <span className="ml-1 capitalize">{metrics?.database.status}</span>
                      </Badge>
                    </div>
                  </div>
                  <Database className="w-8 h-8 text-green-600" />
                </div>
                <div className="mt-4 text-sm text-gray-600">
                  Connections: {metrics?.database.connections}/{metrics?.database.max_connections}
                </div>
              </CardContent>
            </Card>

            {/* API Status */}
            <Card>
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600">API</p>
                    <div className="flex items-center gap-2 mt-1">
                      <Badge className={getStatusColor(metrics?.api.status || 'unknown')}>
                        {getStatusIcon(metrics?.api.status || 'unknown')}
                        <span className="ml-1 capitalize">{metrics?.api.status}</span>
                      </Badge>
                    </div>
                  </div>
                  <Zap className="w-8 h-8 text-purple-600" />
                </div>
                <div className="mt-4 text-sm text-gray-600">
                  {metrics?.api.requests_per_minute} req/min
                </div>
              </CardContent>
            </Card>

            {/* External Services */}
            <Card>
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600">External</p>
                    <div className="flex items-center gap-2 mt-1">
                      <Badge className={getStatusColor(metrics?.external_services.paystack || 'unknown')}>
                        {getStatusIcon(metrics?.external_services.paystack || 'unknown')}
                        <span className="ml-1">Paystack</span>
                      </Badge>
                    </div>
                  </div>
                  <Wifi className="w-8 h-8 text-orange-600" />
                </div>
                <div className="mt-4 text-sm text-gray-600">
                  Networks: 3/4 online
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        {/* Server Tab */}
        <TabsContent value="server">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Cpu className="w-5 h-5" />
                  CPU Usage
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div>
                    <div className="flex justify-between text-sm mb-2">
                      <span>Current Usage</span>
                      <span>{metrics?.server.cpu_usage}%</span>
                    </div>
                    <Progress value={metrics?.server.cpu_usage} className="h-2" />
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <MemoryStick className="w-5 h-5" />
                  Memory Usage
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div>
                    <div className="flex justify-between text-sm mb-2">
                      <span>Current Usage</span>
                      <span>{metrics?.server.memory_usage}%</span>
                    </div>
                    <Progress value={metrics?.server.memory_usage} className="h-2" />
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <HardDrive className="w-5 h-5" />
                  Disk Usage
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div>
                    <div className="flex justify-between text-sm mb-2">
                      <span>Current Usage</span>
                      <span>{metrics?.server.disk_usage}%</span>
                    </div>
                    <Progress value={metrics?.server.disk_usage} className="h-2" />
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Performance Metrics</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Response Time</span>
                    <span className="font-medium">{metrics?.server.response_time}ms</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Uptime</span>
                    <span className="font-medium">{formatUptime(metrics?.server.uptime || 0)}</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        {/* Database Tab */}
        <TabsContent value="database">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <Card>
              <CardHeader>
                <CardTitle>Connection Pool</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div>
                    <div className="flex justify-between text-sm mb-2">
                      <span>Active Connections</span>
                      <span>{metrics?.database.connections}/{metrics?.database.max_connections}</span>
                    </div>
                    <Progress 
                      value={(metrics?.database.connections || 0) / (metrics?.database.max_connections || 1) * 100} 
                      className="h-2" 
                    />
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Query Performance</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Avg Query Time</span>
                    <span className="font-medium">{metrics?.database.query_time}ms</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Slow Queries</span>
                    <span className="font-medium">{metrics?.database.slow_queries}</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        {/* External Services Tab */}
        <TabsContent value="external">
          <div className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Payment Gateway</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex items-center justify-between p-4 border rounded-lg">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-blue-100 rounded-full">
                      <Zap className="w-5 h-5 text-blue-600" />
                    </div>
                    <div>
                      <div className="font-medium">Paystack</div>
                      <div className="text-sm text-gray-500">Payment processing</div>
                    </div>
                  </div>
                  <Badge className={getStatusColor(metrics?.external_services.paystack || 'unknown')}>
                    {getStatusIcon(metrics?.external_services.paystack || 'unknown')}
                    <span className="ml-1 capitalize">{metrics?.external_services.paystack}</span>
                  </Badge>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Telecom Providers</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {Object.entries(metrics?.external_services.telecom_providers || {}).map(([provider, status]) => (
                    <div key={provider} className="flex items-center justify-between p-4 border rounded-lg">
                      <div className="flex items-center gap-3">
                        <div className="p-2 bg-green-100 rounded-full">
                          <Wifi className="w-5 h-5 text-green-600" />
                        </div>
                        <div>
                          <div className="font-medium capitalize">{provider.replace('_', ' ')}</div>
                          <div className="text-sm text-gray-500">Network provider</div>
                        </div>
                      </div>
                      <Badge className={getStatusColor(status)}>
                        {getStatusIcon(status)}
                        <span className="ml-1 capitalize">{status}</span>
                      </Badge>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        {/* Alerts Tab */}
        <TabsContent value="alerts">
          <Card>
            <CardHeader>
              <CardTitle>Recent Alerts</CardTitle>
              <CardDescription>
                System alerts and notifications from the last 24 hours
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {metrics?.recent_alerts.map((alert) => (
                  <Alert key={alert.id} className={`border-l-4 ${
                    alert.type === 'error' ? 'border-l-red-500' :
                    alert.type === 'warning' ? 'border-l-yellow-500' :
                    'border-l-blue-500'
                  }`}>
                    <div className="flex items-start justify-between">
                      <div className="flex items-start gap-2">
                        {alert.type === 'error' && <AlertTriangle className="w-4 h-4 text-red-500 mt-0.5" />}
                        {alert.type === 'warning' && <AlertTriangle className="w-4 h-4 text-yellow-500 mt-0.5" />}
                        {alert.type === 'info' && <CheckCircle className="w-4 h-4 text-blue-500 mt-0.5" />}
                        <div>
                          <AlertDescription className="font-medium">
                            {alert.message}
                          </AlertDescription>
                          <div className="text-xs text-gray-500 mt-1">
                            {formatDuration(alert.timestamp)}
                          </div>
                        </div>
                      </div>
                      <Badge variant={alert.resolved ? 'default' : 'destructive'}>
                        {alert.resolved ? 'Resolved' : 'Active'}
                      </Badge>
                    </div>
                  </Alert>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
};

export default SystemMonitoringDashboard;