import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { 
  Users, 
  DollarSign, 
  TrendingUp, 
  Award,
  Star,
  Gift,
  Target,
  Clock,
  CheckCircle,
  ArrowRight,
  Copy,
  Share2,
  Smartphone,
  Mail,
  MessageCircle
} from 'lucide-react';

interface AffiliateStatsProps {
  totalClicks: number;
  totalReferrals: number;
  totalCommission: number;
  pendingCommission: number;
  conversionRate: number;
  commissionTier: string;
  commissionRate: number;
  referralLink: string;
  onCopyLink: () => void;
}

export const AffiliateStatsCards: React.FC<AffiliateStatsProps> = ({
  totalClicks,
  totalReferrals,
  totalCommission,
  pendingCommission,
  conversionRate,
  commissionTier,
  commissionRate,
  referralLink,
  onCopyLink
}) => {
  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-NG', {
      style: 'currency',
      currency: 'NGN'
    }).format(amount);
  };

  const getNextTierInfo = () => {
    if (commissionTier === 'STARTER') {
      return { nextTier: 'PRO', referralsNeeded: 51 - totalReferrals, nextRate: 10 };
    } else if (commissionTier === 'PRO') {
      return { nextTier: 'ELITE', referralsNeeded: 200 - totalReferrals, nextRate: 15 };
    }
    return null;
  };

  const nextTier = getNextTierInfo();

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      {/* Total Clicks */}
      <Card className="hover:shadow-lg transition-shadow">
        <CardContent className="p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Total Clicks</p>
              <p className="text-2xl font-bold text-blue-600">{totalClicks.toLocaleString()}</p>
              <p className="text-xs text-gray-500 mt-1">Link visits</p>
            </div>
            <div className="w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center">
              <Users className="w-6 h-6 text-blue-600" />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Total Referrals */}
      <Card className="hover:shadow-lg transition-shadow">
        <CardContent className="p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Referrals</p>
              <p className="text-2xl font-bold text-green-600">{totalReferrals}</p>
              <p className="text-xs text-gray-500 mt-1">Successful signups</p>
            </div>
            <div className="w-12 h-12 bg-green-100 rounded-full flex items-center justify-center">
              <Target className="w-6 h-6 text-green-600" />
            </div>
          </div>
          {nextTier && (
            <div className="mt-3">
              <div className="flex justify-between text-xs text-gray-500 mb-1">
                <span>{commissionTier}</span>
                <span>{nextTier.nextTier}</span>
              </div>
              <Progress 
                value={(totalReferrals / (commissionTier === 'STARTER' ? 50 : 200)) * 100} 
                className="h-2"
              />
              <p className="text-xs text-gray-500 mt-1">
                {nextTier.referralsNeeded} more for {nextTier.nextRate}% rate
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Total Commission */}
      <Card className="hover:shadow-lg transition-shadow">
        <CardContent className="p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Total Earned</p>
              <p className="text-2xl font-bold text-yellow-600">{formatCurrency(totalCommission)}</p>
              <p className="text-xs text-gray-500 mt-1">All time earnings</p>
            </div>
            <div className="w-12 h-12 bg-yellow-100 rounded-full flex items-center justify-center">
              <DollarSign className="w-6 h-6 text-yellow-600" />
            </div>
          </div>
          <div className="mt-3 flex justify-between text-xs">
            <span className="text-orange-600">Pending: {formatCurrency(pendingCommission)}</span>
            <span className="text-green-600">Paid: {formatCurrency(totalCommission - pendingCommission)}</span>
          </div>
        </CardContent>
      </Card>

      {/* Conversion Rate */}
      <Card className="hover:shadow-lg transition-shadow">
        <CardContent className="p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Conversion Rate</p>
              <p className="text-2xl font-bold text-purple-600">{conversionRate.toFixed(1)}%</p>
              <p className="text-xs text-gray-500 mt-1">Clicks to signups</p>
            </div>
            <div className="w-12 h-12 bg-purple-100 rounded-full flex items-center justify-center">
              <TrendingUp className="w-6 h-6 text-purple-600" />
            </div>
          </div>
          <div className="mt-3">
            <Badge variant={conversionRate >= 5 ? 'default' : 'secondary'} className="text-xs">
              {conversionRate >= 10 ? 'Excellent' : conversionRate >= 5 ? 'Good' : 'Needs Improvement'}
            </Badge>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

interface AffiliatePromotionToolsProps {
  referralLink: string;
  affiliateCode: string;
  onCopyLink: () => void;
}

export const AffiliatePromotionTools: React.FC<AffiliatePromotionToolsProps> = ({
  referralLink,
  affiliateCode,
  onCopyLink
}) => {
  const shareToWhatsApp = () => {
    const message = `🎉 Join RechargeMax and win amazing prizes with every mobile recharge! 💰\n\n✨ Get instant rewards\n🏆 Daily prize draws\n📱 All Nigerian networks\n\nJoin now: ${referralLink}`;
    window.open(`https://wa.me/?text=${encodeURIComponent(message)}`, '_blank');
  };

  const shareToTwitter = () => {
    const message = `🎉 Turn your mobile recharges into winning opportunities! Join @RechargeMax and win amazing prizes daily! 💰🏆 #RechargeAndWin #Nigeria`;
    window.open(`https://twitter.com/intent/tweet?text=${encodeURIComponent(message)}&url=${encodeURIComponent(referralLink)}`, '_blank');
  };

  const shareViaEmail = () => {
    const subject = 'Join RechargeMax - Win Prizes with Every Recharge!';
    const body = `Hi there!\n\nI wanted to share something exciting with you - RechargeMax!\n\nIt's a platform where you can recharge your phone and automatically enter daily prize draws to win amazing rewards. Every recharge gives you a chance to win cash prizes up to ₦500,000!\n\nFeatures:\n✨ Instant mobile recharge for all networks\n🏆 Daily prize draws\n💰 Cash rewards and bonuses\n📱 Easy and secure payments\n\nJoin using my referral link: ${referralLink}\n\nStart winning today!\n\nBest regards`;
    
    window.open(`mailto:?subject=${encodeURIComponent(subject)}&body=${encodeURIComponent(body)}`, '_blank');
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Share2 className="w-6 h-6" />
          Promotion Tools
        </CardTitle>
        <CardDescription>Share your referral link and start earning</CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Referral Link */}
        <div>
          <label className="text-sm font-medium text-gray-700 mb-2 block">Your Referral Link</label>
          <div className="flex gap-2">
            <input 
              value={referralLink} 
              readOnly 
              className="flex-1 p-3 border border-gray-300 rounded-lg bg-gray-50 text-sm"
            />
            <Button onClick={onCopyLink} size="sm">
              <Copy className="w-4 h-4 mr-2" />
              Copy
            </Button>
          </div>
        </div>

        {/* Quick Share Buttons */}
        <div>
          <label className="text-sm font-medium text-gray-700 mb-3 block">Quick Share</label>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
            <Button 
              onClick={shareToWhatsApp}
              variant="outline" 
              className="w-full bg-green-50 hover:bg-green-100 border-green-200"
            >
              <MessageCircle className="w-4 h-4 mr-2 text-green-600" />
              WhatsApp
            </Button>
            
            <Button 
              onClick={shareToTwitter}
              variant="outline" 
              className="w-full bg-blue-50 hover:bg-blue-100 border-blue-200"
            >
              <Share2 className="w-4 h-4 mr-2 text-blue-600" />
              Twitter
            </Button>
            
            <Button 
              onClick={shareViaEmail}
              variant="outline" 
              className="w-full bg-purple-50 hover:bg-purple-100 border-purple-200"
            >
              <Mail className="w-4 h-4 mr-2 text-purple-600" />
              Email
            </Button>
          </div>
        </div>

        {/* Affiliate Code */}
        <div className="bg-gray-50 p-4 rounded-lg">
          <div className="flex justify-between items-center">
            <div>
              <p className="text-sm font-medium text-gray-700">Your Affiliate Code</p>
              <p className="text-lg font-bold text-blue-600">{affiliateCode}</p>
            </div>
            <Badge variant="secondary">Active</Badge>
          </div>
        </div>

        {/* Marketing Tips */}
        <div className="bg-blue-50 p-4 rounded-lg">
          <h4 className="font-semibold text-blue-900 mb-2">💡 Marketing Tips</h4>
          <ul className="text-sm text-blue-800 space-y-1">
            <li>• Share on your social media stories</li>
            <li>• Post in WhatsApp groups (with permission)</li>
            <li>• Tell friends and family about the rewards</li>
            <li>• Highlight the daily prize draws</li>
            <li>• Mention the convenience of instant recharge</li>
          </ul>
        </div>
      </CardContent>
    </Card>
  );
};

interface AffiliateCommissionTiersProps {
  currentTier: string;
  currentRate: number;
  totalReferrals: number;
}

export const AffiliateCommissionTiers: React.FC<AffiliateCommissionTiersProps> = ({
  currentTier,
  currentRate,
  totalReferrals
}) => {
  const tiers = [
    {
      name: 'STARTER',
      rate: 5,
      minReferrals: 1,
      maxReferrals: 50,
      color: 'blue',
      icon: <Star className="w-5 h-5" />,
      benefits: ['5% commission', 'Basic support', 'Monthly payouts']
    },
    {
      name: 'PRO',
      rate: 10,
      minReferrals: 51,
      maxReferrals: 200,
      color: 'green',
      icon: <Award className="w-5 h-5" />,
      benefits: ['10% commission', 'Priority support', 'Weekly payouts', 'Marketing materials']
    },
    {
      name: 'ELITE',
      rate: 15,
      minReferrals: 201,
      maxReferrals: null,
      color: 'yellow',
      icon: <Gift className="w-5 h-5" />,
      benefits: ['15% commission', 'Dedicated manager', 'Daily payouts', 'Custom materials', 'Performance bonuses']
    }
  ];

  const getProgressToNextTier = () => {
    if (currentTier === 'STARTER') {
      return (totalReferrals / 50) * 100;
    } else if (currentTier === 'PRO') {
      return ((totalReferrals - 50) / 150) * 100;
    }
    return 100;
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <TrendingUp className="w-6 h-6" />
          Commission Tiers
        </CardTitle>
        <CardDescription>Earn more as you refer more customers</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {tiers.map((tier) => {
            const isCurrentTier = tier.name === currentTier;
            const isAchieved = totalReferrals >= tier.minReferrals;
            
            return (
              <div 
                key={tier.name}
                className={`p-4 rounded-lg border-2 transition-all ${
                  isCurrentTier 
                    ? 'border-blue-500 bg-blue-50' 
                    : isAchieved 
                      ? 'border-green-200 bg-green-50' 
                      : 'border-gray-200 bg-gray-50'
                }`}
              >
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-3">
                    <div className={`w-10 h-10 rounded-full flex items-center justify-center ${
                      isCurrentTier 
                        ? 'bg-blue-500 text-white' 
                        : isAchieved 
                          ? 'bg-green-500 text-white' 
                          : 'bg-gray-300 text-gray-600'
                    }`}>
                      {tier.icon}
                    </div>
                    <div>
                      <h3 className="font-semibold text-lg">{tier.name}</h3>
                      <p className="text-sm text-gray-600">
                        {tier.minReferrals}+ referrals • {tier.rate}% commission
                      </p>
                    </div>
                  </div>
                  
                  {isCurrentTier && (
                    <Badge className="bg-blue-500">Current Tier</Badge>
                  )}
                  {isAchieved && !isCurrentTier && (
                    <CheckCircle className="w-6 h-6 text-green-500" />
                  )}
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <h4 className="font-medium text-sm text-gray-700 mb-2">Benefits</h4>
                    <ul className="text-sm text-gray-600 space-y-1">
                      {tier.benefits.map((benefit, index) => (
                        <li key={index} className="flex items-center gap-2">
                          <CheckCircle className="w-3 h-3 text-green-500" />
                          {benefit}
                        </li>
                      ))}
                    </ul>
                  </div>
                  
                  {isCurrentTier && tier.maxReferrals && (
                    <div>
                      <h4 className="font-medium text-sm text-gray-700 mb-2">Progress to Next Tier</h4>
                      <Progress value={getProgressToNextTier()} className="h-2 mb-2" />
                      <p className="text-xs text-gray-600">
                        {tier.maxReferrals - totalReferrals} more referrals needed
                      </p>
                    </div>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
};