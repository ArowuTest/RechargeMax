import React, { useState, useEffect, useCallback } from 'react';

import { getUserDashboard, claimPrize } from '@/lib/api';
import { apiClient } from '@/lib/api-client';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useAuthContext } from '@/contexts/AuthContext';

const NIGERIAN_BANKS = [
  { name: 'Access Bank', code: '044' },
  { name: 'Citibank Nigeria', code: '023' },
  { name: 'Ecobank Nigeria', code: '050' },
  { name: 'Fidelity Bank', code: '070' },
  { name: 'First Bank of Nigeria', code: '011' },
  { name: 'First City Monument Bank (FCMB)', code: '214' },
  { name: 'Globus Bank', code: '00103' },
  { name: 'Guaranty Trust Bank (GTBank)', code: '058' },
  { name: 'Heritage Bank', code: '030' },
  { name: 'Jaiz Bank', code: '301' },
  { name: 'Keystone Bank', code: '082' },
  { name: 'Kuda Microfinance Bank', code: '50211' },
  { name: 'Lotus Bank', code: '303' },
  { name: 'OPay Digital Services', code: '999992' },
  { name: 'Palmpay', code: '999991' },
  { name: 'Parallex Bank', code: '526' },
  { name: 'Polaris Bank', code: '076' },
  { name: 'Providus Bank', code: '101' },
  { name: 'Stanbic IBTC Bank', code: '221' },
  { name: 'Standard Chartered Bank', code: '068' },
  { name: 'Sterling Bank', code: '232' },
  { name: 'Titan Trust Bank', code: '102' },
  { name: 'Union Bank of Nigeria', code: '032' },
  { name: 'United Bank for Africa (UBA)', code: '033' },
  { name: 'Unity Bank', code: '215' },
  { name: 'Wema Bank', code: '035' },
  { name: 'Zenith Bank', code: '057' },
];
import { formatCurrency, formatDate, getNetworkColor } from '@/lib/utils';
import { useToast } from '@/hooks/useToast';
import { useNavigate } from 'react-router-dom';
import { SpinWheel } from '@/components/games/SpinWheel';
import {
  CreditCard,
  Gift,
  TrendingUp,
  Calendar,
  Smartphone,
  Trophy,
  User,
  Loader2,
  CheckCircle,
  Clock,
  AlertCircle,
  ArrowLeft,
  DollarSign,
  Phone,
  Download,
  Search,
  Copy,
  RefreshCw,
  Award
} from 'lucide-react';

interface DashboardData {
  user: {
    id: string;
    msisdn: string;
    first_name?: string;
    last_name?: string;
    email: string;
    loyalty_tier: string;
    total_points: number;
    referral_code: string;
  };
  stats: {
    total_recharges: number;
    total_spins: number;
    total_wins: number;
  };
  summary: {
    total_transactions: number;
    total_prizes: number;
    pending_prizes: number;
    claimed_prizes: number;
    total_amount_recharged: number;
    total_subscriptions: number;
    total_subscription_amount: number;
    total_subscription_entries: number;
    total_subscription_points: number;
  };
  recent_transactions: Array<{
    id: string;
    created_at: string;
    network_provider: string;
    recharge_type: string;
    amount: number;
    points_earned: number;
    status: string;
  }>;
  subscriptions: Array<{
    id: string;
    subscription_code: string;
    transaction_date: string;
    amount: number;
    entries: number;
    points_earned: number;
    status: string;
  }>;
  prizes: Array<{
    id: string;
    prize_name: string;
    prize_value: number;
    prize_type: string;
    won_date: string;
    claim_date?: string;
    status: string;
    fulfillment_mode?: string;
    fulfillment_attempts?: number;
    fulfillment_error?: string;
    claim_reference?: string;
  }>;
}

interface BankDetails {
  account_number: string;
  account_name: string;
  bank_name: string;
  bank_code: string;
}

export const UserDashboard: React.FC = () => {
  const { user, isAuthenticated, logout } = useAuthContext();
  const navigate = useNavigate();
  const { toast } = useToast();
  const [loading, setLoading] = useState(true);
  const [dashboardData, setDashboardData] = useState<DashboardData | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState('overview');
  const [searchTerm, setSearchTerm] = useState('');
  const [claimingPrize, setClaimingPrize] = useState<string | null>(null);
    const [bankDetails, setBankDetails] = useState<BankDetails>({
    account_number: '',
    account_name: '',
    bank_name: '',
    bank_code: '',
  });
  const [showBankForm, setShowBankForm] = useState<string | null>(null);
  const [editingEmail, setEditingEmail] = useState(false);
  const [newEmail, setNewEmail] = useState('');
  const [updatingEmail, setUpdatingEmail] = useState(false);
  const [showSpinWheel, setShowSpinWheel] = useState(false);
  const [availableSpins, setAvailableSpins] = useState(0);
  const [checkingSpins, setCheckingSpins] = useState(false);

  const fetchDashboardData = useCallback(async () => {
    if (!user?.msisdn) {
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError(null);
      
      const response = await getUserDashboard(user.msisdn);

      if (response.success && response.data) {
        setDashboardData(response.data);
      } else {
        setError(!response.success ? response.error : 'Failed to load dashboard');
      }
    } catch (err: any) {
      console.error('Dashboard fetch error:', err);
      setError(err.message || 'An error occurred');
    } finally {
      setLoading(false);
    }
  }, [user?.msisdn]);

  useEffect(() => {
    if (isAuthenticated && user) {
      fetchDashboardData();
    } else {
      setLoading(false);
    }
  }, [isAuthenticated, user, fetchDashboardData]);

  // Check for pending spins after dashboard data is loaded.
  // We use sessionStorage to ensure the auto-popup fires at most ONCE per browser
  // session for a given MSISDN.  The user can always open it manually via the
  // "Prizes" tab or a dashboard button, but we won't interrupt them repeatedly.
  useEffect(() => {
    if (dashboardData && user?.msisdn && !checkingSpins) {
      checkPendingSpins();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dashboardData?.user?.id]); // Only run when dashboard data first loads

  const checkPendingSpins = async () => {
    if (!user?.msisdn || checkingSpins) return;

    // Session-level guard: only auto-popup once per login session per MSISDN.
    // This prevents the wheel from re-appearing every time the dashboard re-mounts
    // (e.g. navigating away and back) when there are legitimate remaining spins.
    const sessionKey = `spin_popup_shown_${user.msisdn}`;
    if (sessionStorage.getItem(sessionKey)) return;

    try {
      setCheckingSpins(true);
      const response = await apiClient.get('/spin/eligibility');

      const data = response.data;

      if (data.success && data.data.eligible && data.data.available_spins > 0) {
        setAvailableSpins(data.data.available_spins);
        // Mark that we have shown (or are about to show) the popup this session
        sessionStorage.setItem(sessionKey, '1');
        // Auto-show spin wheel after a short delay
        setTimeout(() => {
          setShowSpinWheel(true);
        }, 1000);
      }
    } catch (error) {
      console.error('Failed to check pending spins:', error);
    } finally {
      setCheckingSpins(false);
    }
  };

  const handleClaimPrize = async (prizeId: string, prizeType: string) => {
    if (!user?.msisdn) return;

    try {
      setClaimingPrize(prizeId);

      let claimData: any = {
        prize_id: prizeId,
        msisdn: user.msisdn
      };

      // For cash prizes, require bank details
      if (prizeType === 'CASH') {
        if (!bankDetails.account_number || !bankDetails.account_name || !bankDetails.bank_name) {
          setShowBankForm(prizeId);
          setClaimingPrize(null);
          return;
        }
        // Send flat body — backend expects account_number, account_name, bank_name at top level
        claimData.account_number = bankDetails.account_number;
        claimData.account_name   = bankDetails.account_name;
        claimData.bank_name      = bankDetails.bank_name;
        claimData.bank_code      = bankDetails.bank_code;
      }

      const result = await claimPrize(prizeId, claimData);
      if (result.success) {
        toast({
          title: "Prize Claimed!",
          description: prizeType === 'CASH' 
            ? 'Your bank details have been submitted. Admin will process your payment within 24-48 hours.'
            : prizeType === 'AIRTIME' || prizeType === 'DATA'
            ? 'Your prize will be credited to your phone shortly. If not received within 24 hours, our team will process it manually.'
            : 'Prize claimed successfully!',
        });
        
        // Reset bank form
        setBankDetails({ account_number: '', account_name: '', bank_name: '', bank_code: '' });
        setShowBankForm(null);
        
        // Refresh dashboard data
        fetchDashboardData();
      } else {
        throw new Error(result.error);
      }
    } catch (error) {
      console.error('Failed to claim prize:', error);
      toast({
        title: "Claim Failed",
        description: error instanceof Error ? error.message : "Failed to claim prize",
        variant: "destructive"
      });
    } finally {
      setClaimingPrize(null);
    }
  };

  const copyReferralCode = () => {
    if (dashboardData?.user?.referral_code) {
      navigator.clipboard.writeText(dashboardData.user.referral_code);
      toast({
        title: 'Copied!',
        description: 'Referral code copied to clipboard',
      });
    }
  };

  const handleUpdateEmail = async () => {
    if (!newEmail || !user?.msisdn) return;

    // Basic email validation
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(newEmail)) {
      toast({
        title: 'Invalid Email',
        description: 'Please enter a valid email address',
        variant: 'destructive',
      });
      return;
    }

    try {
      setUpdatingEmail(true);
      const response = await apiClient.put('/user/profile', { email: newEmail });

      const data = response.data;

      if (data.success) {
        toast({
          title: 'Email Updated!',
          description: 'Your email has been successfully updated',
        });
        setEditingEmail(false);
        setNewEmail('');
        fetchDashboardData();
      } else {
        throw new Error(data.error || 'Failed to update email');
      }
    } catch (error) {
      console.error('Failed to update email:', error);
      toast({
        title: 'Update Failed',
        description: error instanceof Error ? error.message : 'Failed to update email',
        variant: 'destructive',
      });
    } finally {
      setUpdatingEmail(false);
    }
  };
  if (!isAuthenticated) {
    return (
      <div className="container mx-auto p-6">
        <Card>
          <CardContent className="p-6">
            <p>Please log in to view your dashboard.</p>
            <Button onClick={() => navigate('/login')} className="mt-4">
              Go to Login
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="container mx-auto p-6 flex items-center justify-center min-h-screen">
        <div className="text-center">
          <Loader2 className="w-12 h-12 animate-spin mx-auto mb-4" />
          <p>Loading dashboard...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto p-6">
        <Card>
          <CardContent className="p-6">
            <p className="text-red-600">Error: {error}</p>
            <Button onClick={fetchDashboardData} className="mt-4">
              Retry
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!dashboardData) {
    return (
      <div className="container mx-auto p-6">
        <Card>
          <CardContent className="p-6">
            <p>No dashboard data available.</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  const fullName = `${dashboardData.user.first_name || ''} ${dashboardData.user.last_name || ''}`.trim() || 'User';

  // Filter transactions based on search
  const filteredTransactions = dashboardData.recent_transactions?.filter(tx =>
    tx.network_provider?.toLowerCase().includes(searchTerm.toLowerCase()) ||
    tx.status?.toLowerCase().includes(searchTerm.toLowerCase())
  ) || [];

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Dashboard</h1>
          <p className="text-gray-600">Welcome back, {fullName}!</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => navigate('/')}>
            <ArrowLeft className="w-4 h-4 mr-2" />
            Home
          </Button>
          <Button variant="outline" onClick={fetchDashboardData}>
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </Button>
        </div>
      </div>

      {/* Quick Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Points</CardTitle>
            <Award className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{dashboardData.user.total_points || 0}</div>
            <p className="text-xs text-muted-foreground">
              Tier: {dashboardData.user.loyalty_tier}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Recharges</CardTitle>
            <Smartphone className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{dashboardData.stats?.total_recharges || 0}</div>
            <p className="text-xs text-muted-foreground">
              {formatCurrency(dashboardData.summary?.total_amount_recharged || 0)} total
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Prizes Won</CardTitle>
            <Trophy className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{dashboardData.summary?.total_prizes || 0}</div>
            <p className="text-xs text-muted-foreground">
              {dashboardData.summary?.pending_prizes || 0} pending
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Subscriptions</CardTitle>
            <Calendar className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{dashboardData.summary?.total_subscriptions || 0}</div>
            <p className="text-xs text-muted-foreground">
              {dashboardData.summary?.total_subscription_entries || 0} entries earned
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Custom Tabs */}
      <div className="space-y-4">
        {/* Tab Navigation */}
        <div className="border-b border-gray-200">
          <nav className="-mb-px flex space-x-8">
            {['overview', 'transactions', 'subscriptions', 'prizes', 'profile'].map((tab) => (
              <button
                key={tab}
                onClick={() => setActiveTab(tab)}
                className={`
                  whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm
                  ${activeTab === tab
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                  }
                `}
              >
                {tab.charAt(0).toUpperCase() + tab.slice(1)}
              </button>
            ))}
          </nav>
        </div>

        {/* Tab Content */}
        <div className="mt-4">
          {/* Overview Tab */}
          {activeTab === 'overview' && (
            <div className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2">
                <Card>
                  <CardHeader>
                    <CardTitle>Account Summary</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-2">
                    <div className="flex justify-between">
                      <span className="text-sm text-gray-600">Phone Number</span>
                      <span className="font-semibold">{dashboardData.user.msisdn}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-sm text-gray-600">Email</span>
                      <span className="font-semibold">{dashboardData.user.email || 'Not set'}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-sm text-gray-600">Loyalty Tier</span>
                      <Badge variant="secondary">{dashboardData.user.loyalty_tier}</Badge>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-sm text-gray-600">Total Points</span>
                      <span className="font-semibold">{dashboardData.user.total_points}</span>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <CardTitle>Referral Program</CardTitle>
                    <CardDescription>Share your code and earn rewards</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="flex items-center gap-2">
                      <Input
                        value={dashboardData.user.referral_code || 'N/A'}
                        readOnly
                        className="font-mono"
                      />
                      <Button size="icon" onClick={copyReferralCode}>
                        <Copy className="h-4 w-4" />
                      </Button>
                    </div>
                    <p className="text-sm text-gray-600 mt-2">
                      Earn commission when friends use your code!
                    </p>
                  </CardContent>
                </Card>
              </div>

              <Card>
                <CardHeader>
                  <CardTitle>Recent Activity</CardTitle>
                </CardHeader>
                <CardContent>
                  {dashboardData.recent_transactions && dashboardData.recent_transactions.length > 0 ? (
                    <div className="space-y-2">
                      {dashboardData.recent_transactions.slice(0, 5).map((tx) => (
                        <div key={tx.id} className="flex justify-between items-center p-3 border rounded">
                          <div className="flex items-center gap-3">
                            <div className={`w-10 h-10 rounded-full flex items-center justify-center ${getNetworkColor(tx.network_provider)}`}>
                              <Phone className="w-5 h-5 text-white" />
                            </div>
                            <div>
                              <p className="font-semibold">{tx.network_provider} {tx.recharge_type}</p>
                              <p className="text-sm text-gray-600">{formatDate(tx.created_at)}</p>
                            </div>
                          </div>
                          <div className="text-right">
                            <p className="font-semibold">{formatCurrency(tx.amount)}</p>
                            <div className="flex items-center gap-2">
                              <Badge variant={tx.status === 'SUCCESS' ? 'default' : 'secondary'}>
                                {tx.status}
                              </Badge>
                              {tx.points_earned > 0 && (
                                <span className="text-xs text-green-600">+{tx.points_earned} pts</span>
                              )}
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <p className="text-gray-600">No recent transactions</p>
                  )}
                </CardContent>
              </Card>
            </div>
          )}

          {/* Transactions Tab */}
          {activeTab === 'transactions' && (
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle>Transaction History</CardTitle>
                  <div className="flex gap-2">
                    <div className="relative">
                      <Search className="absolute left-2 top-2.5 h-4 w-4 text-gray-500" />
                      <Input
                        placeholder="Search transactions..."
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                        className="pl-8 w-64"
                      />
                    </div>
                    <Button variant="outline" size="sm">
                      <Download className="h-4 w-4 mr-2" />
                      Export
                    </Button>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Date</TableHead>
                      <TableHead>Network</TableHead>
                      <TableHead>Type</TableHead>
                      <TableHead>Amount</TableHead>
                      <TableHead>Points</TableHead>
                      <TableHead>Status</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredTransactions.length > 0 ? (
                      filteredTransactions.map((tx) => (
                        <TableRow key={tx.id}>
                          <TableCell>{formatDate(tx.created_at)}</TableCell>
                          <TableCell>
                            <Badge className={getNetworkColor(tx.network_provider)}>
                              {tx.network_provider}
                            </Badge>
                          </TableCell>
                          <TableCell>{tx.recharge_type}</TableCell>
                          <TableCell>{formatCurrency(tx.amount)}</TableCell>
                          <TableCell>
                            {tx.points_earned > 0 ? (
                              <span className="text-green-600">+{tx.points_earned}</span>
                            ) : (
                              <span className="text-gray-400">0</span>
                            )}
                          </TableCell>
                          <TableCell>
                            <Badge variant={tx.status === 'SUCCESS' ? 'default' : 'secondary'}>
                              {tx.status}
                            </Badge>
                          </TableCell>
                        </TableRow>
                      ))
                    ) : (
                      <TableRow>
                        <TableCell colSpan={6} className="text-center text-gray-600">
                          No transactions found
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          )}

          {/* Subscriptions Tab */}
          {activeTab === 'subscriptions' && (
            <div className="space-y-4">
              <div className="grid gap-4 md:grid-cols-3">
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Total Subscriptions</CardTitle>
                    <Calendar className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">{dashboardData.summary?.total_subscriptions || 0}</div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Total Entries</CardTitle>
                    <Trophy className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">{dashboardData.summary?.total_subscription_entries || 0}</div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Points Earned</CardTitle>
                    <Award className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">{dashboardData.summary?.total_subscription_points || 0}</div>
                  </CardContent>
                </Card>
              </div>

              <Card>
                <CardHeader>
                  <CardTitle>Subscription History</CardTitle>
                </CardHeader>
                <CardContent>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Date</TableHead>
                        <TableHead>Code</TableHead>
                        <TableHead>Amount</TableHead>
                        <TableHead>Entries</TableHead>
                        <TableHead>Points</TableHead>
                        <TableHead>Status</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {dashboardData.subscriptions && dashboardData.subscriptions.length > 0 ? (
                        dashboardData.subscriptions.map((sub) => (
                          <TableRow key={sub.id}>
                            <TableCell>{formatDate(sub.transaction_date)}</TableCell>
                            <TableCell className="font-mono text-sm">{sub.subscription_code}</TableCell>
                            <TableCell>{formatCurrency(sub.amount)}</TableCell>
                            <TableCell>{sub.entries}</TableCell>
                            <TableCell className="text-green-600">+{sub.points_earned}</TableCell>
                            <TableCell>
                              <Badge variant={sub.status === 'active' ? 'default' : 'secondary'}>
                                {sub.status}
                              </Badge>
                            </TableCell>
                          </TableRow>
                        ))
                      ) : (
                        <TableRow>
                          <TableCell colSpan={6} className="text-center text-gray-600">
                            No subscriptions yet
                          </TableCell>
                        </TableRow>
                      )}
                    </TableBody>
                  </Table>
                </CardContent>
              </Card>

              {/* Quick Subscribe CTA */}
              <Card className="bg-gradient-to-r from-blue-50 to-purple-50 border-blue-200">
                <CardContent className="p-6">
                  <div className="flex items-center justify-between">
                    <div>
                      <h4 className="font-semibold text-gray-900 text-lg">Add More Entries</h4>
                      <p className="text-sm text-gray-600">Subscribe for more daily draw entries</p>
                    </div>
                    <Button 
                      onClick={() => navigate('/subscription')}
                      className="bg-blue-600 hover:bg-blue-700"
                    >
                      Subscribe
                    </Button>
                  </div>
                </CardContent>
              </Card>

              {/* Benefits Section */}
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <Card>
                  <CardContent className="p-6 text-center">
                    <CheckCircle className="w-8 h-8 text-green-600 mx-auto mb-2" />
                    <h4 className="font-semibold mb-1">Guaranteed Entry</h4>
                    <p className="text-sm text-gray-600">1 draw entry every day</p>
                  </CardContent>
                </Card>
                <Card>
                  <CardContent className="p-6 text-center">
                    <Trophy className="w-8 h-8 text-yellow-600 mx-auto mb-2" />
                    <h4 className="font-semibold mb-1">Win Big</h4>
                    <p className="text-sm text-gray-600">Up to ₦500,000 prizes</p>
                  </CardContent>
                </Card>
                <Card>
                  <CardContent className="p-6 text-center">
                    <Clock className="w-8 h-8 text-blue-600 mx-auto mb-2" />
                    <h4 className="font-semibold mb-1">Daily Draws</h4>
                    <p className="text-sm text-gray-600">Multiple draws every day</p>
                  </CardContent>
                </Card>
              </div>
            </div>
          )}

          {/* Prizes Tab */}
          {activeTab === 'prizes' && dashboardData && (
            <div className="space-y-4">
              <div className="grid gap-4 md:grid-cols-3">
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Total Prizes</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">{dashboardData?.summary?.total_prizes || 0}</div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Pending</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">{dashboardData?.summary?.pending_prizes || 0}</div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Claimed</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">{(dashboardData?.summary?.total_prizes || 0) - (dashboardData?.summary?.pending_prizes || 0)}</div>
                  </CardContent>
                </Card>
              </div>

              <Card>
                <CardHeader>
                  <CardTitle>Prize History</CardTitle>
                </CardHeader>
                <CardContent>
                  {dashboardData?.prizes && dashboardData.prizes.length > 0 ? (
                    <div className="space-y-2">
                      {dashboardData.prizes.map((prize, index) => (
                        <div key={prize?.id || index} className="border rounded p-4 space-y-3">
                          <div className="flex justify-between items-start">
                            <div className="flex-1">
                              <p className="font-semibold text-lg">{prize?.prize_name || 'Unknown Prize'}</p>
                              <p className="text-sm text-gray-600">Won on {prize?.won_date ? formatDate(prize.won_date) : 'N/A'}</p>
                              {prize?.claim_date && (
                                <p className="text-sm text-green-600">Claimed on {formatDate(prize.claim_date)}</p>
                              )}
                              
                              {/* Fulfillment Status for Airtime/Data */}
                              {(prize?.prize_type === 'AIRTIME' || prize?.prize_type === 'DATA') && (
                                <div className="mt-2 space-y-1">
                                  {prize?.fulfillment_mode && (
                                    <p className="text-xs text-gray-500">
                                      Mode: <span className="font-medium">{prize.fulfillment_mode}</span>
                                    </p>
                                  )}
                                  {(prize?.fulfillment_attempts ?? 0) > 0 && (
                                    <p className="text-xs text-gray-500">
                                      Provisioning attempts: <span className="font-medium">{prize.fulfillment_attempts}</span>
                                    </p>
                                  )}
                                  {prize?.fulfillment_error && (
                                    <p className="text-xs text-red-600 bg-red-50 p-2 rounded">
                                      ⚠️ {prize.fulfillment_error}
                                    </p>
                                  )}
                                  {prize?.claim_reference && (
                                    <p className="text-xs text-green-600">
                                      Ref: {prize.claim_reference}
                                    </p>
                                  )}
                                </div>
                              )}
                            </div>
                            <div className="text-right">
                              <p className="font-semibold text-lg">{prize?.prize_value ? formatCurrency(prize.prize_value) : 'N/A'}</p>
                              <Badge variant={prize?.status === 'CLAIMED' ? 'default' : 'secondary'}>
                                {prize?.status || 'PENDING'}
                              </Badge>
                            </div>
                          </div>

                          {/* Claim Button for Unclaimed Prizes */}
                          {prize?.status === 'PENDING' && (
                            <div className="space-y-3">
                              {/* Bank Details Form for Cash Prizes */}
                              {prize?.prize_type === 'CASH' && showBankForm === prize.id && (
                                <div className="bg-gray-50 p-4 rounded space-y-3">
                                  <p className="text-sm font-medium">Enter your bank details to claim this cash prize:</p>
                                  <div className="grid gap-3 md:grid-cols-2">
                                    <div>
                                      <label className="text-sm font-medium">Account Name</label>
                                      <Input
                                        value={bankDetails.account_name}
                                        onChange={(e) => setBankDetails(prev => ({ ...prev, account_name: e.target.value }))}
                                        placeholder="John Doe"
                                      />
                                    </div>
                                    <div>
                                      <label className="text-sm font-medium">Account Number</label>
                                      <Input
                                        value={bankDetails.account_number}
                                        onChange={(e) => setBankDetails(prev => ({ ...prev, account_number: e.target.value }))}
                                        placeholder="1234567890"
                                      />
                                    </div>
                                    <div className="md:col-span-2">
                                      <label className="text-sm font-medium">Bank Name</label>
                                      <Select
                                        value={bankDetails.bank_name}
                                        onValueChange={(val) => {
                                          const bank = NIGERIAN_BANKS.find((b) => b.name === val);
                                          setBankDetails(prev => ({
                                            ...prev,
                                            bank_name: val,
                                            bank_code: bank?.code ?? '',
                                          }));
                                        }}
                                      >
                                        <SelectTrigger>
                                          <SelectValue placeholder="Select your bank" />
                                        </SelectTrigger>
                                        <SelectContent>
                                          {NIGERIAN_BANKS.map((b) => (
                                            <SelectItem key={b.code} value={b.name}>
                                              {b.name} ({b.code})
                                            </SelectItem>
                                          ))}
                                        </SelectContent>
                                      </Select>
                                    </div>
                                  </div>
                                  <div className="flex gap-2">
                                    <Button
                                      onClick={() => handleClaimPrize(prize.id, prize.prize_type)}
                                      disabled={claimingPrize === prize.id}
                                    >
                                      {claimingPrize === prize.id ? (
                                        <><Loader2 className="w-4 h-4 animate-spin mr-2" />Submitting...</>
                                      ) : (
                                        'Submit Claim'
                                      )}
                                    </Button>
                                    <Button
                                      variant="outline"
                                      onClick={() => {
                                        setShowBankForm(null);
                                        setBankDetails({ account_number: '', account_name: '', bank_name: '', bank_code: '' });
                                      }}
                                    >
                                      Cancel
                                    </Button>
                                  </div>
                                </div>
                              )}

                              {/* Claim Button */}
                              {showBankForm !== prize.id && (
                                <Button
                                  onClick={() => handleClaimPrize(prize.id, prize.prize_type || 'OTHER')}
                                  disabled={claimingPrize === prize.id}
                                  className="w-full"
                                >
                                  {claimingPrize === prize.id ? (
                                    <><Loader2 className="w-4 h-4 animate-spin mr-2" />Claiming...</>
                                  ) : (
                                    <><Gift className="w-4 h-4 mr-2" />Claim Now</>
                                  )}
                                </Button>
                              )}
                            </div>
                          )}
                        </div>
                      ))}
                    </div>
                  ) : (
                    <p className="text-gray-600 text-center py-4">No prizes yet. Keep playing to win!</p>
                  )}
                </CardContent>
              </Card>
            </div>
          )}

          {/* Profile Tab */}
          {activeTab === 'profile' && dashboardData && (
            <div className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle>Profile Information</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="grid gap-4 md:grid-cols-2">
                    <div>
                      <label className="text-sm font-medium">First Name</label>
                      <Input value={dashboardData.user.first_name || ''} readOnly />
                    </div>
                    <div>
                      <label className="text-sm font-medium">Last Name</label>
                      <Input value={dashboardData.user.last_name || ''} readOnly />
                    </div>
                    <div>
                      <label className="text-sm font-medium">Phone Number</label>
                      <Input value={dashboardData.user.msisdn} readOnly />
                    </div>
                    <div className="md:col-span-2">
                      <label className="text-sm font-medium">Email</label>
                      {editingEmail ? (
                        <div className="flex gap-2">
                          <Input
                            value={newEmail}
                            onChange={(e) => setNewEmail(e.target.value)}
                            placeholder="Enter your email"
                            type="email"
                          />
                          <Button
                            onClick={handleUpdateEmail}
                            disabled={updatingEmail}
                            size="sm"
                          >
                            {updatingEmail ? <Loader2 className="w-4 h-4 animate-spin" /> : 'Save'}
                          </Button>
                          <Button
                            onClick={() => {
                              setEditingEmail(false);
                              setNewEmail('');
                            }}
                            variant="outline"
                            size="sm"
                          >
                            Cancel
                          </Button>
                        </div>
                      ) : (
                        <div className="flex gap-2">
                          <Input value={dashboardData.user.email || 'Not set'} readOnly />
                          <Button
                            onClick={() => {
                              setEditingEmail(true);
                              setNewEmail(dashboardData.user.email || '');
                            }}
                            variant="outline"
                            size="sm"
                          >
                            Edit
                          </Button>
                        </div>
                      )}
                    </div>
                    <div>
                      <label className="text-sm font-medium">Loyalty Tier</label>
                      <Input value={dashboardData.user.loyalty_tier} readOnly />
                    </div>
                    <div>
                      <label className="text-sm font-medium">Total Points</label>
                      <Input value={dashboardData.user.total_points.toString()} readOnly />
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <Button variant="outline">Edit Profile</Button>
                    <Button variant="outline" onClick={logout}>Logout</Button>
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Referral Code</CardTitle>
                  <CardDescription>Share this code with friends to earn rewards</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="flex items-center gap-2">
                    <Input
                      value={dashboardData.user.referral_code || 'N/A'}
                      readOnly
                      className="font-mono text-lg"
                    />
                    <Button size="icon" onClick={copyReferralCode}>
                      <Copy className="h-4 w-4" />
                    </Button>
                  </div>
                </CardContent>
              </Card>
            </div>
          )}
        </div>
      </div>

      {/* Spin Wheel Modal */}
      {showSpinWheel && availableSpins > 0 && (
        <SpinWheel
          isOpen={showSpinWheel}
          onClose={() => {
            setShowSpinWheel(false);
            setAvailableSpins(0);
            // Refresh dashboard to show new prizes
            fetchDashboardData();
          }}
          transactionAmount={1000} // actual spins are managed by the backend
          userPhone={user?.msisdn || ''}
          onPrizeWon={async (_prize) => {
            // After a spin completes, ask the BACKEND how many spins remain.
            // Never trust the client-side decrement alone — the server is the
            // single source of truth for spin counts.
            try {
              const res = await apiClient.get('/spin/eligibility');
              const remaining = res.data?.data?.available_spins ?? 0;
              setAvailableSpins(remaining);
              if (remaining <= 0) {
                setTimeout(() => {
                  setShowSpinWheel(false);
                  fetchDashboardData();
                }, 3000);
              }
            } catch {
              // Fallback: close wheel and refresh
              setAvailableSpins(0);
              setTimeout(() => {
                setShowSpinWheel(false);
                fetchDashboardData();
              }, 3000);
            }
          }}
        />
      )}
    </div>
  );
};
