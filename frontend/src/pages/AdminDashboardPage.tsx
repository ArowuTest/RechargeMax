import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useEffect, useState } from 'react';
import { useAdminContext } from '@/contexts/AdminContext';
import { adminApi } from '@/lib/api-client';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  LayoutDashboard,
  Users,
  Gift,
  Settings,
  TrendingUp,
  FileText,
  Phone,
  DollarSign,
  Trophy,
  Layers,
  BarChart3,
  Shield,
  LogOut,
  AlertTriangle,
  CheckCircle,
  Activity,
  Star,
  Zap,
  Clock,
} from 'lucide-react';

interface AdminModule {
  id: string;
  title: string;
  description: string;
  icon: React.ReactNode;
  path: string;
  color: string;
  permissions?: string[];
}

export const AdminDashboardPage: React.FC = () => {
  const navigate = useNavigate();
  const { adminUser, adminLogout, hasPermission } = useAdminContext();
  const [stats, setStats] = useState<any>(null);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const response = await adminApi.getStats();
        // API returns { success: true, data: { total_users, active_draws, ... } }
        if (response && response.success && response.data) {
          setStats(response.data);
        } else if (response && response.total_users !== undefined) {
          // Direct data object
          setStats(response);
        } else {
          setStats(response);
        }
      } catch (err) {
        console.error('Failed to fetch admin stats:', err);
      }
    };
    fetchStats();
  }, []);

  const adminModules: AdminModule[] = [
    {
      id: 'comprehensive',
      title: 'Comprehensive Portal',
      description: 'All-in-one admin management dashboard',
      icon: <LayoutDashboard className="w-6 h-6" />,
      path: '/admin/comprehensive',
      color: 'bg-blue-500',
    },
    {
      id: 'draws',
      title: 'Draw Management',
      description: 'Manage draws, entries, and winners',
      icon: <Gift className="w-6 h-6" />,
      path: '/admin/draws',
      color: 'bg-purple-500',
    },
    {
      id: 'winners',
      title: 'Winner Claims',
      description: 'Process and approve winner claims',
      icon: <Trophy className="w-6 h-6" />,
      path: '/admin/winners',
      color: 'bg-yellow-500',
    },
    {
      id: 'wheel-prizes',
      title: 'Wheel Prizes',
      description: 'Configure spin wheel prizes and probabilities',
      icon: <Settings className="w-6 h-6" />,
      path: '/admin/wheel-prizes',
      color: 'bg-green-500',
    },
    {
      id: 'subscriptions',
      title: 'Subscription Tiers',
      description: 'Manage subscription tiers and benefits',
      icon: <Layers className="w-6 h-6" />,
      path: '/admin/subscriptions',
      color: 'bg-indigo-500',
    },
    {
      id: 'pricing',
      title: 'Pricing Config',
      description: 'Configure subscription pricing',
      icon: <DollarSign className="w-6 h-6" />,
      path: '/admin/pricing',
      color: 'bg-emerald-500',
    },
    {
      id: 'daily-subs',
      title: 'Daily Subscriptions',
      description: 'Monitor daily subscription activity',
      icon: <Clock className="w-6 h-6" />,
      path: '/admin/daily-subscriptions',
      color: 'bg-cyan-500',
    },
    {
      id: 'ussd',
      title: 'USSD Monitoring',
      description: 'Track USSD recharge transactions',
      icon: <Phone className="w-6 h-6" />,
      path: '/admin/ussd',
      color: 'bg-orange-500',
    },
    {
      id: 'affiliates',
      title: 'Affiliate Management',
      description: 'Manage affiliate program and commissions',
      icon: <TrendingUp className="w-6 h-6" />,
      path: '/admin/affiliates',
      color: 'bg-pink-500',
    },
    {
      id: 'csv',
      title: 'CSV Management',
      description: 'Import/export draw data via CSV',
      icon: <FileText className="w-6 h-6" />,
      path: '/admin/csv',
      color: 'bg-violet-500',
    },
    {
      id: 'monitoring',
      title: 'System Monitoring',
      description: 'Monitor system health and performance',
      icon: <BarChart3 className="w-6 h-6" />,
      path: '/admin/monitoring',
      color: 'bg-red-500',
    },
    {
      id: 'recharge-monitoring',
      title: 'Recharge Monitoring',
      description: 'Track recharge transactions in real time',
      icon: <Activity className="w-6 h-6" />,
      path: '/admin/recharge-monitoring',
      color: 'bg-teal-500',
    },
    {
      id: 'commissions',
      title: 'Commission Reconciliation',
      description: 'Reconcile and release affiliate commissions',
      icon: <CheckCircle className="w-6 h-6" />,
      path: '/admin/commissions',
      color: 'bg-lime-500',
    },
    {
      id: 'failed-provisions',
      title: 'Failed Provisions',
      description: 'Retry failed prize fulfilment attempts',
      icon: <AlertTriangle className="w-6 h-6" />,
      path: '/admin/failed-provisions',
      color: 'bg-rose-500',
    },
    {
      id: 'unclaimed-prizes',
      title: 'Unclaimed Prizes',
      description: 'View and chase up unclaimed prize winners',
      icon: <Star className="w-6 h-6" />,
      path: '/admin/unclaimed-prizes',
      color: 'bg-amber-500',
    },
    {
      id: 'spin-tiers',
      title: 'Spin Tiers',
      description: 'Configure spin eligibility tiers',
      icon: <Zap className="w-6 h-6" />,
      path: '/admin/spin-tiers',
      color: 'bg-sky-500',
    },
    {
      id: 'prize-fulfillment',
      title: 'Prize Fulfillment',
      description: 'Manage spin prize fulfilment configuration',
      icon: <Gift className="w-6 h-6" />,
      path: '/admin/prize-fulfillment',
      color: 'bg-fuchsia-500',
    },
    {
      id: 'validation-stats',
      title: 'Validation Stats',
      description: 'Phone number validation analytics',
      icon: <Shield className="w-6 h-6" />,
      path: '/admin/validation-stats',
      color: 'bg-slate-500',
    },
  ];

  const handleLogout = () => {
    adminLogout();
    navigate('/admin/login');
  };

  const handleModuleClick = (path: string) => {
    navigate(path);
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100">
      {/* Header */}
      <header className="bg-white border-b shadow-sm">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-blue-600 rounded-lg flex items-center justify-center">
                <Shield className="w-6 h-6 text-white" />
              </div>
              <div>
                <h1 className="text-xl font-bold">RechargeMax Admin</h1>
                <p className="text-sm text-muted-foreground">Management Portal</p>
              </div>
            </div>

            <div className="flex items-center gap-4">
              <div className="text-right">
                <p className="font-medium">{adminUser?.username}</p>
                <Badge variant="secondary" className="text-xs">
                  {adminUser?.role?.replace('_', ' ').toUpperCase()}
                </Badge>
              </div>
              <Button variant="outline" size="sm" onClick={handleLogout}>
                <LogOut className="w-4 h-4 mr-2" />
                Logout
              </Button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="container mx-auto px-4 py-8">
        {/* Welcome Section */}
        <div className="mb-8">
          <h2 className="text-3xl font-bold mb-2">
            Welcome back, {adminUser?.username}!
          </h2>
          <p className="text-muted-foreground">
            Select a module below to manage your platform
          </p>
        </div>

        {/* Quick Stats */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          <Card>
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">Total Users</p>
                  <p className="text-2xl font-bold">{stats?.total_users ?? '-'}</p>
                </div>
                <Users className="w-8 h-8 text-blue-500" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">Active Draws</p>
                  <p className="text-2xl font-bold">{stats?.active_draws ?? '-'}</p>
                </div>
                <Gift className="w-8 h-8 text-purple-500" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">Pending Claims</p>
                  <p className="text-2xl font-bold">{stats?.pending_claims ?? '-'}</p>
                </div>
                <Trophy className="w-8 h-8 text-yellow-500" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">Revenue</p>
                  <p className="text-2xl font-bold">₦{stats?.total_revenue ? (stats.total_revenue / 100).toLocaleString() : '0'}</p>
                </div>
                <DollarSign className="w-8 h-8 text-green-500" />
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Admin Modules Grid */}
        <div>
          <h3 className="text-xl font-semibold mb-4">Admin Modules</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {adminModules.map((module) => (
              <Card
                key={module.id}
                className="hover:shadow-lg transition-all cursor-pointer group"
                onClick={() => handleModuleClick(module.path)}
              >
                <CardHeader>
                  <div className="flex items-start justify-between">
                    <div className={`w-12 h-12 ${module.color} rounded-lg flex items-center justify-center text-white group-hover:scale-110 transition-transform`}>
                      {module.icon}
                    </div>
                    {module.permissions && (
                      <Badge variant="outline" className="text-xs">
                        Restricted
                      </Badge>
                    )}
                  </div>
                  <CardTitle className="text-lg mt-4">{module.title}</CardTitle>
                  <CardDescription>{module.description}</CardDescription>
                </CardHeader>
              </Card>
            ))}
          </div>
        </div>

        {/* Footer */}
        <div className="mt-12 text-center text-sm text-muted-foreground">
          <p>© 2024 RechargeMax. All rights reserved.</p>
          <p className="mt-1">Last login: {adminUser?.last_login ? new Date(adminUser.last_login).toLocaleString() : 'N/A'}</p>
        </div>
      </main>
    </div>
  );
};
