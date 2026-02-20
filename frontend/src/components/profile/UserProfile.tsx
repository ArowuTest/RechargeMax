import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { useAuthContext } from '@/contexts/AuthContext';
import { useToast } from '@/hooks/use-toast';
import { getUserDashboard } from '@/lib/api';
import { 
  User, 
  Mail, 
  Phone, 
  Edit3, 
  Save, 
  X, 
  CheckCircle,
  AlertCircle,
  Trophy,
  TrendingUp,
  Calendar,
  Shield,
  Loader2
} from 'lucide-react';

interface UserProfileProps {
  className?: string;
}

interface ProfileData {
  user: {
    id: string;
    msisdn: string;
    full_name: string;
    email: string;
    loyalty_tier: string;
    total_points: number;
    total_recharges: number;
    created_at: string;
    updated_at: string;
  };
  summary: {
    total_transactions: number;
    total_prizes: number;
    total_amount_recharged: number;
  };
}

export const UserProfile: React.FC<UserProfileProps> = ({ className = "" }) => {
  const { user, isAuthenticated } = useAuthContext();
  const { toast } = useToast();
  const [profileData, setProfileData] = useState<ProfileData | null>(null);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(false);
  const [editForm, setEditForm] = useState({
    full_name: '',
    email: ''
  });

  useEffect(() => {
    if (isAuthenticated && user) {
      fetchProfileData();
    } else {
      setLoading(false);
    }
  }, [isAuthenticated, user]);

  const fetchProfileData = async () => {
    if (!user?.msisdn) return;

    try {
      setLoading(true);
      
      const result = await getUserDashboard(user.msisdn);
      
      if (result.success) {
        setProfileData(result.data);
        setEditForm({
          full_name: result.data.user.full_name || '',
          email: result.data.user.email || ''
        });
      } else {
        throw new Error(result.error);
      }
    } catch (error) {
      console.error('Failed to fetch profile data:', error);
      toast({
        title: "Error",
        description: "Failed to load profile information",
        variant: "destructive"
      });
    } finally {
      setLoading(false);
    }
  };

  const handleSaveProfile = async () => {
    if (!user?.msisdn || !profileData) return;

    try {
      // For now, just update local state since we don't have an update endpoint
      // In a real app, you'd call an update API here
      toast({
        title: "Profile Updated",
        description: "Your profile information has been saved",
      });
      
      setEditing(false);
      
      // Update local state
      setProfileData(prev => prev ? {
        ...prev,
        user: {
          ...prev.user,
          full_name: editForm.full_name,
          email: editForm.email
        }
      } : null);
      
    } catch (error) {
      console.error('Failed to update profile:', error);
      toast({
        title: "Update Failed",
        description: "Failed to update profile information",
        variant: "destructive"
      });
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-NG', {
      style: 'currency',
      currency: 'NGN'
    }).format(amount);
  };

  const getTierColor = (tier: string) => {
    switch (tier.toLowerCase()) {
      case 'bronze': return 'bg-amber-100 text-amber-800 border-amber-200';
      case 'silver': return 'bg-gray-100 text-gray-800 border-gray-200';
      case 'gold': return 'bg-yellow-100 text-yellow-800 border-yellow-200';
      case 'platinum': return 'bg-purple-100 text-purple-800 border-purple-200';
      default: return 'bg-blue-100 text-blue-800 border-blue-200';
    }
  };

  if (!isAuthenticated) {
    return (
      <div className={`min-h-screen bg-gradient-to-br from-blue-50 to-purple-50 flex items-center justify-center p-4 ${className}`}>
        <Card className="w-full max-w-md text-center">
          <CardHeader>
            <CardTitle className="flex items-center justify-center gap-2">
              <User className="w-6 h-6" />
              Login Required
            </CardTitle>
            <CardDescription>
              Please login to view your profile
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button onClick={() => window.location.href = '/#/login'} className="w-full">
              <User className="w-4 h-4 mr-2" />
              Go to Login
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (loading) {
    return (
      <div className={`min-h-screen bg-gradient-to-br from-blue-50 to-purple-50 flex items-center justify-center ${className}`}>
        <Card className="w-full max-w-md text-center">
          <CardContent className="p-8">
            <Loader2 className="w-8 h-8 animate-spin mx-auto mb-4" />
            <p>Loading your profile...</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!profileData) {
    return (
      <div className={`min-h-screen bg-gradient-to-br from-blue-50 to-purple-50 flex items-center justify-center p-4 ${className}`}>
        <Card className="w-full max-w-md text-center">
          <CardContent className="p-8">
            <AlertCircle className="w-8 h-8 text-red-500 mx-auto mb-4" />
            <p>Failed to load profile information</p>
            <Button onClick={fetchProfileData} className="mt-4">
              Try Again
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className={`min-h-screen bg-gradient-to-br from-blue-50 to-purple-50 p-4 ${className}`}>
      <div className="max-w-4xl mx-auto space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Profile</h1>
            <p className="text-gray-600">Manage your account information</p>
          </div>
          <Button variant="outline" onClick={() => window.location.href = '/#/'}>
            Back to Home
          </Button>
        </div>

        {/* Profile Information */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle className="flex items-center gap-2">
                  <User className="w-6 h-6" />
                  Personal Information
                </CardTitle>
                <CardDescription>
                  Your account details and preferences
                </CardDescription>
              </div>
              {!editing ? (
                <Button variant="outline" onClick={() => setEditing(true)}>
                  <Edit3 className="w-4 h-4 mr-2" />
                  Edit Profile
                </Button>
              ) : (
                <div className="flex gap-2">
                  <Button onClick={handleSaveProfile}>
                    <Save className="w-4 h-4 mr-2" />
                    Save
                  </Button>
                  <Button variant="outline" onClick={() => {
                    setEditing(false);
                    setEditForm({
                      full_name: profileData.user.full_name || '',
                      email: profileData.user.email || ''
                    });
                  }}>
                    <X className="w-4 h-4 mr-2" />
                    Cancel
                  </Button>
                </div>
              )}
            </div>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {/* Phone Number */}
              <div>
                <Label className="flex items-center gap-2 mb-2">
                  <Phone className="w-4 h-4" />
                  Phone Number
                </Label>
                <Input 
                  value={profileData.user.msisdn} 
                  disabled 
                  className="bg-gray-50"
                />
                <p className="text-xs text-gray-500 mt-1">Phone number cannot be changed</p>
              </div>

              {/* Full Name */}
              <div>
                <Label className="flex items-center gap-2 mb-2">
                  <User className="w-4 h-4" />
                  Full Name
                </Label>
                {editing ? (
                  <Input
                    value={editForm.full_name}
                    onChange={(e) => setEditForm(prev => ({ ...prev, full_name: e.target.value }))}
                    placeholder="Enter your full name"
                  />
                ) : (
                  <Input 
                    value={profileData.user.full_name || 'Not set'} 
                    disabled 
                    className="bg-gray-50"
                  />
                )}
              </div>

              {/* Email */}
              <div>
                <Label className="flex items-center gap-2 mb-2">
                  <Mail className="w-4 h-4" />
                  Email Address
                </Label>
                {editing ? (
                  <Input
                    type="email"
                    value={editForm.email}
                    onChange={(e) => setEditForm(prev => ({ ...prev, email: e.target.value }))}
                    placeholder="Enter your email address"
                  />
                ) : (
                  <Input 
                    value={profileData.user.email || 'Not set'} 
                    disabled 
                    className="bg-gray-50"
                  />
                )}
              </div>

              {/* Loyalty Tier */}
              <div>
                <Label className="flex items-center gap-2 mb-2">
                  <Trophy className="w-4 h-4" />
                  Loyalty Tier
                </Label>
                <div>
                  <Badge className={getTierColor(profileData.user.loyalty_tier)}>
                    {profileData.user.loyalty_tier}
                  </Badge>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Account Statistics */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card>
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600">Total Recharges</p>
                  <p className="text-2xl font-bold">{formatCurrency(profileData.user.total_recharges)}</p>
                </div>
                <TrendingUp className="w-8 h-8 text-green-600" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600">Total Points</p>
                  <p className="text-2xl font-bold">{profileData.user.total_points}</p>
                </div>
                <Trophy className="w-8 h-8 text-yellow-600" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600">Total Transactions</p>
                  <p className="text-2xl font-bold">{profileData.summary.total_transactions}</p>
                </div>
                <Shield className="w-8 h-8 text-blue-600" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600">Prizes Won</p>
                  <p className="text-2xl font-bold">{profileData.summary.total_prizes}</p>
                </div>
                <CheckCircle className="w-8 h-8 text-purple-600" />
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Account Information */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Calendar className="w-6 h-6" />
              Account Information
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <Label className="text-sm font-medium text-gray-600">Member Since</Label>
                <p className="text-lg">{formatDate(profileData.user.created_at)}</p>
              </div>
              <div>
                <Label className="text-sm font-medium text-gray-600">Last Updated</Label>
                <p className="text-lg">{formatDate(profileData.user.updated_at)}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};