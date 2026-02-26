import React, { useState, useEffect } from 'react';
import { getAffiliateDashboard, registerAffiliate, refreshAffiliateLink } from '@/lib/api';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Separator } from '@/components/ui/separator';
import { useAuthContext } from '@/contexts/AuthContext';
import { useToast } from '@/hooks/use-toast';
import { 
  Users, 
  Link2, 
  Copy, 
  DollarSign, 
  TrendingUp, 
  Award,
  CheckCircle,
  Star,
  Target,
  Gift,
  Smartphone,
  CreditCard,
  UserPlus,
  Mail,
  Phone,
  User,
  ArrowRight,
  Info,
  Loader2,
  AlertCircle,
  Clock,
  Send,
  RefreshCw
} from 'lucide-react';

interface AffiliateData {
  id: string;
  affiliate_code: string;
  full_name: string;
  email: string;
  phone_number: string;
  status: string;
  commission_tier: string;
  commission_rate: number;
  total_clicks: number;
  total_referrals: number;
  total_commission: number;
  pending_commission: number;
  paid_commission: number;
  is_active: boolean;
  conversion_rate: number;
  referral_link: string;
  created_at: string;
  approved_at?: string;
}

interface AffiliateRegistration {
  full_name: string;
  email: string;
  phone_number: string;
  bank_name: string;
  account_number: string;
  account_name: string;
  referral_source: string;
}

export const AffiliateDashboard: React.FC = () => {
  const { user, isAuthenticated } = useAuthContext();
  const { toast } = useToast();
  const [loading, setLoading] = useState(false);
  const [registering, setRegistering] = useState(false);
  const [affiliateData, setAffiliateData] = useState<AffiliateData | null>(null);
  const [affiliateStatus, setAffiliateStatus] = useState<string>('');
  const [dashboardLoading, setDashboardLoading] = useState(true);
  const [showRegistrationForm, setShowRegistrationForm] = useState(false);
  const [statistics, setStatistics] = useState<any>(null);
  const [bankAccounts, setBankAccounts] = useState<any[]>([]);
  const [referralLink, setReferralLink] = useState<string>('');
  
  const [registrationForm, setRegistrationForm] = useState<AffiliateRegistration>({
    full_name: '',
    email: '',
    phone_number: '',
    bank_name: '',
    account_number: '',
    account_name: '',
    referral_source: ''
  });

  // Fetch affiliate data for authenticated users
  useEffect(() => {
    if (isAuthenticated && user?.msisdn) {
      fetchAffiliateData();
    } else {
      setDashboardLoading(false);
    }
  }, [isAuthenticated, user]);

  const fetchAffiliateData = async () => {
    if (!user?.msisdn) return;

    try {
      setDashboardLoading(true);
      
      const response = await getAffiliateDashboard(user.msisdn);

      if (response.success) {
        setAffiliateData(response.data.affiliate);
        setStatistics(response.data.statistics);
        setBankAccounts(response.data.bank_accounts || []);
        setReferralLink(response.data.referral_link || '');
        setAffiliateStatus('APPROVED');
      } else {
        // Check if it's a status issue (pending approval)
        if (response.status) {
          setAffiliateStatus(String(response.status));
        } else {
          setAffiliateStatus('NOT_FOUND');
        }
        console.log('Affiliate fetch result:', response.error);
      }
    } catch (error) {
      console.error('Failed to fetch affiliate data:', error);
      setAffiliateStatus('ERROR');
    } finally {
      setDashboardLoading(false);
    }
  };

  const handleCopyLink = async () => {
    if (!referralLink) return;
    
    try {
      await navigator.clipboard.writeText(referralLink);
      toast({
        title: "Link Copied!",
        description: "Referral link copied to clipboard",
      });
    } catch (error) {
      toast({
        title: "Copy Failed",
        description: "Please copy the link manually",
        variant: "destructive"
      });
    }
  };

  const handleRefreshLink = async () => {
    if (!user?.msisdn) return;
    
    try {
      setDashboardLoading(true);
      
      const response = await refreshAffiliateLink();
      
      if (response.success) {
        // Update the referral link with the new URL
        setReferralLink(response.data.referral_link);
        
        toast({
          title: "Link Updated!",
          description: "Your referral link has been updated to the current domain",
        });
      } else {
        throw new Error(response.error || 'Failed to refresh link');
      }
    } catch (error) {
      console.error('Failed to refresh affiliate link:', error);
      toast({
        title: "Refresh Failed",
        description: "Please try again later",
        variant: "destructive"
      });
    } finally {
      setDashboardLoading(false);
    }
  };

  const handleAffiliateRegistration = async () => {
    if (!registrationForm.full_name || !registrationForm.email || !registrationForm.phone_number) {
      toast({
        title: "Missing Information",
        description: "Please fill in all required fields",
        variant: "destructive"
      });
      return;
    }

    if (!registrationForm.bank_name || !registrationForm.account_number || !registrationForm.account_name) {
      toast({
        title: "Missing Bank Information",
        description: "Please provide complete bank account details",
        variant: "destructive"
      });
      return;
    }

    setRegistering(true);
    
    try {
      const result = await registerAffiliate(registrationForm);
      if (result.success) {
        toast({
          title: "Registration Successful!",
          description: result.message,
        });
        
        // Reset form
        setRegistrationForm({
          full_name: '',
          email: '',
          phone_number: '',
          bank_name: '',
          account_number: '',
          account_name: '',
          referral_source: ''
        });
        
        // Hide registration form and set status to pending
        setShowRegistrationForm(false);
        setAffiliateStatus('PENDING');
      } else {
        toast({
          title: "Registration Failed",
          description: result.error,
          variant: "destructive"
        });
      }
      
    } catch (error) {
      toast({
        title: "Registration Failed",
        description: "Please try again later",
        variant: "destructive"
      });
    } finally {
      setRegistering(false);
    }
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-NG', {
      style: 'currency',
      currency: 'NGN'
    }).format(amount);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  };

  // For non-authenticated users, show affiliate program information and registration
  if (!isAuthenticated) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-purple-50 p-4">
        <div className="max-w-6xl mx-auto space-y-8">
          {/* Hero Section */}
          <div className="text-center space-y-4">
            <div className="flex items-center justify-center gap-2 mb-4">
              <Users className="w-12 h-12 text-blue-600" />
            </div>
            <h1 className="text-4xl font-bold text-gray-900">Join Our Affiliate Program</h1>
            <p className="text-xl text-gray-600 max-w-2xl mx-auto">
              Earn up to 15% commission on every recharge made by your referrals. 
              Start building your passive income today!
            </p>
          </div>

          <Tabs defaultValue="overview" className="space-y-6">
            <TabsList className="grid w-full grid-cols-3">
              <TabsTrigger value="overview">Program Overview</TabsTrigger>
              <TabsTrigger value="benefits">Benefits & Commission</TabsTrigger>
              <TabsTrigger value="register">Join Now</TabsTrigger>
            </TabsList>

            {/* Program Overview */}
            <TabsContent value="overview">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                      <Target className="w-6 h-6 text-blue-600" />
                      How It Works
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="flex items-start gap-3">
                      <div className="w-8 h-8 bg-blue-100 text-blue-600 rounded-full flex items-center justify-center font-bold text-sm">1</div>
                      <div>
                        <h4 className="font-semibold">Get Your Unique Link</h4>
                        <p className="text-sm text-gray-600">Receive a personalized referral link after approval</p>
                      </div>
                    </div>
                    <div className="flex items-start gap-3">
                      <div className="w-8 h-8 bg-blue-100 text-blue-600 rounded-full flex items-center justify-center font-bold text-sm">2</div>
                      <div>
                        <h4 className="font-semibold">Share & Promote</h4>
                        <p className="text-sm text-gray-600">Share your link on social media, WhatsApp, or with friends</p>
                      </div>
                    </div>
                    <div className="flex items-start gap-3">
                      <div className="w-8 h-8 bg-blue-100 text-blue-600 rounded-full flex items-center justify-center font-bold text-sm">3</div>
                      <div>
                        <h4 className="font-semibold">Earn Commissions</h4>
                        <p className="text-sm text-gray-600">Get paid for every recharge made by your referrals</p>
                      </div>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                      <Info className="w-6 h-6 text-green-600" />
                      Requirements
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-3">
                    <div className="flex items-center gap-2">
                      <CheckCircle className="w-5 h-5 text-green-600" />
                      <span>Valid Nigerian phone number</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <CheckCircle className="w-5 h-5 text-green-600" />
                      <span>Active social media presence</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <CheckCircle className="w-5 h-5 text-green-600" />
                      <span>Nigerian bank account for payouts</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <CheckCircle className="w-5 h-5 text-green-600" />
                      <span>Commitment to promote responsibly</span>
                    </div>
                  </CardContent>
                </Card>
              </div>
            </TabsContent>

            {/* Benefits & Commission */}
            <TabsContent value="benefits">
              <div className="space-y-6">
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                      <DollarSign className="w-6 h-6 text-green-600" />
                      Commission Structure
                    </CardTitle>
                    <CardDescription>Earn more as you refer more customers</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                      <div className="text-center p-4 border rounded-lg">
                        <div className="text-2xl font-bold text-blue-600">5%</div>
                        <div className="text-sm text-gray-600">1-50 Referrals</div>
                        <div className="text-xs text-gray-500">Starter Level</div>
                      </div>
                      <div className="text-center p-4 border rounded-lg bg-blue-50">
                        <div className="text-2xl font-bold text-blue-600">10%</div>
                        <div className="text-sm text-gray-600">51-200 Referrals</div>
                        <div className="text-xs text-gray-500">Pro Level</div>
                      </div>
                      <div className="text-center p-4 border rounded-lg bg-yellow-50">
                        <div className="text-2xl font-bold text-yellow-600">15%</div>
                        <div className="text-sm text-gray-600">200+ Referrals</div>
                        <div className="text-xs text-gray-500">Elite Level</div>
                      </div>
                    </div>
                  </CardContent>
                </Card>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <Card>
                    <CardHeader>
                      <CardTitle className="flex items-center gap-2">
                        <Gift className="w-6 h-6 text-purple-600" />
                        Additional Benefits
                      </CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-3">
                      <div className="flex items-center gap-2">
                        <Star className="w-5 h-5 text-yellow-500" />
                        <span>Monthly performance bonuses</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <Star className="w-5 h-5 text-yellow-500" />
                        <span>Exclusive promotional materials</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <Star className="w-5 h-5 text-yellow-500" />
                        <span>Priority customer support</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <Star className="w-5 h-5 text-yellow-500" />
                        <span>Real-time analytics dashboard</span>
                      </div>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader>
                      <CardTitle className="flex items-center gap-2">
                        <CreditCard className="w-6 h-6 text-green-600" />
                        Payout Information
                      </CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-3">
                      <div className="flex justify-between">
                        <span>Minimum Payout:</span>
                        <span className="font-semibold">₦3,000</span>
                      </div>
                      <div className="flex justify-between">
                        <span>Payment Schedule:</span>
                        <span className="font-semibold">Weekly</span>
                      </div>
                      <div className="flex justify-between">
                        <span>Payment Method:</span>
                        <span className="font-semibold">Bank Transfer</span>
                      </div>
                      <div className="flex justify-between">
                        <span>Processing Time:</span>
                        <span className="font-semibold">1-3 Business Days</span>
                      </div>
                    </CardContent>
                  </Card>
                </div>
              </div>
            </TabsContent>

            {/* Registration Form */}
            <TabsContent value="register">
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <UserPlus className="w-6 h-6 text-blue-600" />
                    Become an Affiliate Partner
                  </CardTitle>
                  <CardDescription>
                    Fill out the form below to apply for our affiliate program. 
                    Approval typically takes 24-48 hours.
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-6">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {/* Personal Information */}
                    <div className="space-y-4">
                      <h3 className="font-semibold text-lg flex items-center gap-2">
                        <User className="w-5 h-5" />
                        Personal Information
                      </h3>
                      
                      <div>
                        <Label htmlFor="full_name">Full Name *</Label>
                        <Input
                          id="full_name"
                          value={registrationForm.full_name}
                          onChange={(e) => setRegistrationForm(prev => ({ ...prev, full_name: e.target.value }))}
                          placeholder="John Doe"
                        />
                      </div>

                      <div>
                        <Label htmlFor="email">Email Address *</Label>
                        <Input
                          id="email"
                          type="email"
                          value={registrationForm.email}
                          onChange={(e) => setRegistrationForm(prev => ({ ...prev, email: e.target.value }))}
                          placeholder="john@example.com"
                        />
                      </div>

                      <div>
                        <Label htmlFor="phone_number">Phone Number *</Label>
                        <Input
                          id="phone_number"
                          value={registrationForm.phone_number}
                          onChange={(e) => setRegistrationForm(prev => ({ ...prev, phone_number: e.target.value }))}
                          placeholder="08012345678"
                        />
                      </div>
                    </div>

                    {/* Bank Information */}
                    <div className="space-y-4">
                      <h3 className="font-semibold text-lg flex items-center gap-2">
                        <CreditCard className="w-5 h-5" />
                        Bank Information
                      </h3>
                      
                      <div>
                        <Label htmlFor="bank_name">Bank Name *</Label>
                        <Input
                          id="bank_name"
                          value={registrationForm.bank_name}
                          onChange={(e) => setRegistrationForm(prev => ({ ...prev, bank_name: e.target.value }))}
                          placeholder="First Bank of Nigeria"
                        />
                      </div>

                      <div>
                        <Label htmlFor="account_number">Account Number *</Label>
                        <Input
                          id="account_number"
                          value={registrationForm.account_number}
                          onChange={(e) => setRegistrationForm(prev => ({ ...prev, account_number: e.target.value }))}
                          placeholder="1234567890"
                        />
                      </div>

                      <div>
                        <Label htmlFor="account_name">Account Name *</Label>
                        <Input
                          id="account_name"
                          value={registrationForm.account_name}
                          onChange={(e) => setRegistrationForm(prev => ({ ...prev, account_name: e.target.value }))}
                          placeholder="John Doe"
                        />
                      </div>
                    </div>
                  </div>

                  <div>
                    <Label htmlFor="referral_source">How did you hear about us? (Optional)</Label>
                    <Input
                      id="referral_source"
                      value={registrationForm.referral_source}
                      onChange={(e) => setRegistrationForm(prev => ({ ...prev, referral_source: e.target.value }))}
                      placeholder="Social media, friend referral, online search, etc."
                    />
                  </div>

                  <div className="flex gap-4">
                    <Button 
                      onClick={handleAffiliateRegistration}
                      disabled={registering}
                      className="flex-1"
                    >
                      {registering ? (
                        <Loader2 className="w-4 h-4 animate-spin mr-2" />
                      ) : (
                        <UserPlus className="w-4 h-4 mr-2" />
                      )}
                      {registering ? 'Submitting Application...' : 'Apply to Join Program'}
                    </Button>
                    
                    <Button variant="outline" onClick={() => window.location.href = '/#/login'}>
                      <User className="w-4 h-4 mr-2" />
                      Already an Affiliate? Login
                    </Button>
                  </div>

                  <div className="text-sm text-gray-500 text-center">
                    By applying, you agree to our affiliate terms and conditions. 
                    All applications are reviewed within 24-48 hours.
                  </div>
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </div>
      </div>
    );
  }

  // Loading state for authenticated users
  if (dashboardLoading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-purple-50 flex items-center justify-center">
        <Card className="w-full max-w-md text-center">
          <CardContent className="p-8">
            <Loader2 className="w-8 h-8 animate-spin mx-auto mb-4" />
            <p>Loading affiliate dashboard...</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Handle different affiliate statuses
  if (affiliateStatus === 'PENDING') {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-purple-50 flex items-center justify-center p-4">
        <Card className="w-full max-w-md text-center">
          <CardContent className="p-8">
            <Clock className="w-12 h-12 text-orange-500 mx-auto mb-4" />
            <h2 className="text-xl font-bold mb-2">Application Under Review</h2>
            <p className="text-gray-600 mb-4">
              Your affiliate application is being reviewed by our team. 
              You'll receive approval notification within 24-48 hours.
            </p>
            <Button onClick={() => window.location.href = '/#/'}>
              Back to Home
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (affiliateStatus === 'REJECTED') {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-purple-50 flex items-center justify-center p-4">
        <Card className="w-full max-w-md text-center">
          <CardContent className="p-8">
            <AlertCircle className="w-12 h-12 text-red-500 mx-auto mb-4" />
            <h2 className="text-xl font-bold mb-2">Application Rejected</h2>
            <p className="text-gray-600 mb-4">
              Unfortunately, your affiliate application was not approved. 
              Please contact support for more information.
            </p>
            <Button onClick={() => window.location.href = '/#/'}>
              Back to Home
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (affiliateStatus === 'NOT_FOUND') {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-purple-50 flex items-center justify-center p-4">
        <Card className="w-full max-w-md text-center">
          <CardContent className="p-8">
            <Users className="w-12 h-12 text-blue-500 mx-auto mb-4" />
            <h2 className="text-xl font-bold mb-2">Not an Affiliate Yet</h2>
            <p className="text-gray-600 mb-4">
              You haven't applied for our affiliate program yet. 
              Join now to start earning commissions!
            </p>
            <div className="flex gap-2">
              <Button onClick={() => setShowRegistrationForm(true)} className="flex-1">
                Apply as New Affiliate
              </Button>
              <Button variant="outline" onClick={() => window.location.href = '/#/'}>
                Back to Home
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  // For approved affiliates, show full dashboard
  if (!affiliateData) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-purple-50 flex items-center justify-center p-4">
        <Card className="w-full max-w-md text-center">
          <CardContent className="p-8">
            <AlertCircle className="w-8 h-8 text-red-500 mx-auto mb-4" />
            <p>Failed to load affiliate data</p>
            <Button onClick={fetchAffiliateData} className="mt-4">
              Try Again
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-purple-50 p-4">
      <div className="max-w-6xl mx-auto space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Affiliate Dashboard</h1>
            <p className="text-gray-600">Track your referrals and earnings</p>
            <Badge className="mt-2" variant={affiliateData.status === 'APPROVED' ? 'default' : 'secondary'}>
              {affiliateData.status} • {affiliateData.commission_tier} ({affiliateData.commission_rate}%)
            </Badge>
          </div>
          <Button variant="outline" onClick={() => window.location.href = '/#/'}>
            Back to Home
          </Button>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card>
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600">Total Clicks</p>
                  <p className="text-2xl font-bold">{statistics?.total_clicks?.toLocaleString() || '0'}</p>
                </div>
                <Link2 className="w-8 h-8 text-blue-600" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600">Referrals</p>
                  <p className="text-2xl font-bold">{statistics?.total_referrals || '0'}</p>
                </div>
                <Users className="w-8 h-8 text-green-600" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600">Total Earned</p>
                  <p className="text-2xl font-bold">{formatCurrency(statistics?.total_commission || 0)}</p>
                </div>
                <DollarSign className="w-8 h-8 text-yellow-600" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600">Conversion Rate</p>
                  <p className="text-2xl font-bold">{(statistics?.conversion_rate || 0).toFixed(1)}%</p>
                </div>
                <TrendingUp className="w-8 h-8 text-purple-600" />
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Referral Link */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Link2 className="w-6 h-6" />
              Your Referral Link
            </CardTitle>
            <CardDescription>Share this link to earn commissions</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex gap-2">
              <Input value={referralLink} readOnly className="flex-1" />
              <Button onClick={handleCopyLink}>
                <Copy className="w-4 h-4 mr-2" />
                Copy
              </Button>
              <Button 
                onClick={handleRefreshLink} 
                variant="outline"
                disabled={dashboardLoading}
                title="Update link to current domain"
              >
                {dashboardLoading ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  <RefreshCw className="w-4 h-4" />
                )}
              </Button>
            </div>
            <p className="text-xs text-gray-500 mt-2">
              💡 Click refresh if your link points to an old domain
            </p>
          </CardContent>
        </Card>

        {/* Commission Summary */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <Card>
            <CardHeader>
              <CardTitle>Commission Summary</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex justify-between">
                <span>Pending Commission:</span>
                <span className="font-semibold text-orange-600">{formatCurrency(statistics?.pending_commission || 0)}</span>
              </div>
              <div className="flex justify-between">
                <span>Paid Commission:</span>
                <span className="font-semibold text-green-600">{formatCurrency(statistics?.paid_commission || 0)}</span>
              </div>
              <Separator />
              <div className="flex justify-between text-lg font-bold">
                <span>Total Earned:</span>
                <span>{formatCurrency(statistics?.total_commission || 0)}</span>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Account Information</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex justify-between">
                <span>Affiliate Code:</span>
                <span className="font-semibold">{affiliateData.affiliate_code}</span>
              </div>
              <div className="flex justify-between">
                <span>Member Since:</span>
                <span className="font-semibold">{formatDate(affiliateData.created_at)}</span>
              </div>
              {affiliateData.approved_at && (
                <div className="flex justify-between">
                  <span>Approved On:</span>
                  <span className="font-semibold">{formatDate(affiliateData.approved_at)}</span>
                </div>
              )}
              <div className="flex justify-between">
                <span>Status:</span>
                <Badge variant={affiliateData.is_active ? 'default' : 'secondary'}>
                  {affiliateData.is_active ? 'Active' : 'Inactive'}
                </Badge>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Quick Actions */}
        <Card>
          <CardHeader>
            <CardTitle>Quick Actions</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <Button className="w-full" onClick={handleCopyLink}>
                <Copy className="w-4 h-4 mr-2" />
                Copy Referral Link
              </Button>
              <Button variant="outline" className="w-full">
                <Mail className="w-4 h-4 mr-2" />
                Download Marketing Materials
              </Button>
              <Button 
                variant="outline" 
                className="w-full"
                disabled={!statistics?.pending_commission || statistics.pending_commission < 3000}
              >
                <DollarSign className="w-4 h-4 mr-2" />
                Request Payout ({formatCurrency(statistics?.pending_commission || 0)})
              </Button>
            </div>
            {(statistics?.pending_commission || 0) < 3000 && (
              <p className="text-sm text-gray-500 text-center">
                Minimum payout amount is ₦3,000. Current pending: {formatCurrency(statistics?.pending_commission || 0)}
              </p>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
};