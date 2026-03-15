import React, { useState, useEffect } from 'react';
import { apiClient } from '@/lib/api-client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { PremiumRechargeForm } from '@/components/recharge/PremiumRechargeForm';
import { DrawsList } from '@/components/draws/DrawsList';
import { SpinWheel } from '@/components/games/SpinWheel';
import { DailySpinProgress } from '@/components/spin/DailySpinProgress';
import { useToast } from '@/hooks/useToast';
import { useAuthContext } from '@/contexts/AuthContext';
// Removed Supabase - using Go backend API
import { formatCurrency } from '@/lib/utils';
import { getAvailableSpins, getPlatformStatistics, getRecentWinners } from '@/lib/api';
import { 
  Smartphone, 
  Trophy, 
  Users, 
  Zap, 
  Gift, 
  Star, 
  TrendingUp, 
  Shield, 
  Clock, 
  Target, 
  Sparkles,
  ArrowRight,
  CheckCircle,
  Crown,
  Flame,
  Award,
  DollarSign,
  Calendar,
  BarChart3
} from 'lucide-react';

interface Stats {
  totalUsers: number;
  totalRecharges: number;
  totalPrizes: number;
  activeDraw: {
    name: string;
    prizePool: number;
    endTime: string;
    entries: number;
  } | null;
}

interface RecentWinner {
  id: string;
  name: string;
  prize: string;
  amount: number;
  time: string;
  network: string;
}

export const EnterpriseHomePage: React.FC = () => {
  const { user, isAuthenticated } = useAuthContext();
  const { toast } = useToast();
  const [rechargeSuccess, setRechargeSuccess] = useState<any>(null);
  const [stats, setStats] = useState<Stats>({
    totalUsers: 0,
    totalRecharges: 0,
    totalPrizes: 0,
    activeDraw: null
  });
  
  const [recentWinners, setRecentWinners] = useState<RecentWinner[]>([]);

  const [showSpinWheel, setShowSpinWheel] = useState(false);
  const [availableSpins, setAvailableSpins] = useState(0);
  const [userPhone, setUserPhone] = useState<string>('');

  // Handle recharge success
  const handleRechargeSuccess = (result: any) => {
    setRechargeSuccess(result);
    if (result.amount >= 1000) {
      setShowSpinWheel(true);
    }
    // Refresh platform data after successful recharge
    fetchPlatformData();
  };

  // Fetch real data from backend
  useEffect(() => {
    fetchPlatformData();
    
    // Check for payment success parameters
    // Read URL search parameters (BrowserRouter standard query string)
    const hash = window.location.hash; // kept for legacy Paystack redirect compatibility
    const hashQueryString = hash.includes('?') ? hash.split('?')[1] : '';
    const hashParams = new URLSearchParams(hashQueryString);
    const urlParams = new URLSearchParams(window.location.search);
    
    // Combine both parameter sources
    const allParams = new URLSearchParams();
    hashParams.forEach((value, key) => allParams.set(key, value));
    urlParams.forEach((value, key) => allParams.set(key, value));
    
    
    // Check for payment success (both formats)
    const paymentStatus = allParams.get('payment');
    const paymentSuccess = allParams.get('payment_success') === 'true' || paymentStatus === 'success';
    const subscriptionSuccess = allParams.get('subscription_success') === 'true';
    const reference = allParams.get('reference') || allParams.get('ref');
    
    // Handle simple callback format
    if (paymentSuccess && reference) {
      
      // Clear URL immediately
      window.history.replaceState({}, document.title, window.location.pathname);
      
      // Trigger backend callback to process VTU
      apiClient.get(`/payment/callback?reference=${reference}&gateway=paystack`)
        .catch(err => console.error('Backend callback failed:', err));
      
      // Poll transaction status with retry mechanism
      const pollTransaction = (attempt = 0, maxAttempts = 20) => {
        const pollInterval = 3000; // Poll every 3 seconds
        const maxWaitTime = 60000; // Max 60 seconds
        
        if (attempt === 0) {
        }
        
        setTimeout(() => {
          apiClient.get(`/recharge/reference/${reference}`)
            .then(res => res.data)
            .then(response => {
              // Extract transaction data from nested response
              const txn = response.data || response;
              
              if (txn.status === 'SUCCESS' || txn.status === 'COMPLETED') {
                const amount = txn.amount / 100; // Convert from kobo to naira
                const points = txn.points_earned || 0;
                const spinEligible = txn.spin_eligible || amount >= 1000;
                
                // Store user phone for spin system
                if (txn.msisdn && txn.msisdn !== 'null') {
                  setUserPhone(txn.msisdn);
                }
                
                setRechargeSuccess({
                  amount,
                  points,
                  spinEligible,
                  phone: txn.msisdn || 'Guest User',
                  network: txn.network_provider || 'MTN',
                  transactionReference: reference
                });
                
                // Show success popup
                const message = `🎉 Recharge Successful!\n\nAmount: ₦${amount.toLocaleString()}\nPhone: ${txn.msisdn}\nNetwork: ${txn.network_provider}\nPoints Earned: ${points}\n\n${spinEligible ? '🎰 Wheel Spin Activated!\nYou can now spin the wheel to win prizes!' : 'Recharge ₦1,000+ to unlock wheel spin'}`;
                alert(message);
                
                if (spinEligible) {
                  // Set available spins to 1 for this transaction
                  setAvailableSpins(1);
                  // Show wheel after a short delay
                  setTimeout(() => {
                    setShowSpinWheel(true);
                  }, 500);
                }
              } else if (txn.status === 'FAILED') {
                // Transaction failed
                alert(`❌ Recharge Failed\n\nReason: ${txn.failure_reason || 'Unknown error'}\nReference: ${reference}`);
              } else if (attempt < maxAttempts - 1) {
                // Still processing, poll again
                pollTransaction(attempt + 1, maxAttempts);
              } else {
                // Max attempts reached
                alert(`⏳ Transaction is taking longer than expected.\n\nReference: ${reference}\n\nPlease check your transaction history in a few minutes.`);
              }
            })
            .catch(err => {
              console.error(`Polling attempt ${attempt + 1} failed:`, err);
              if (attempt < maxAttempts - 1) {
                // Retry on error
                pollTransaction(attempt + 1, maxAttempts);
              } else {
                alert('Unable to verify transaction status. Please check your transaction history.');
              }
            });
        }, attempt === 0 ? 3000 : pollInterval); // First poll after 3s, then every 3s
      };
      
      // Start polling
      pollTransaction();
      
      return; // Exit early after starting polling
    }
    
    // Handle subscription success
    if (subscriptionSuccess && reference) {
      
      // Clear URL immediately
      window.history.replaceState({}, document.title, window.location.pathname);
      
      // Trigger backend callback to process subscription
      apiClient.get(`/payment/callback?reference=${reference}&gateway=paystack`)
        .catch(err => console.error('Backend callback failed:', err));
      
      const amount = parseFloat(allParams.get('amount') || '0');
      const entries = parseInt(allParams.get('entries') || '0');
      const totalEntries = parseInt(allParams.get('totalEntries') || '0');
      const totalPoints = parseInt(allParams.get('totalPoints') || '0');
      const isAdditional = totalEntries > entries;
      
      // Show subscription success popup
      setTimeout(() => {
        const message = isAdditional 
          ? `🎉 Subscription Added!\n\nSuccessfully added ${entries} entries for ₦${amount.toLocaleString()}!\n\nYour total daily entries: ${totalEntries}\nTotal points: ${totalPoints}\n\nGood luck in today's draw!`
          : `🎉 Daily Subscription Activated!\n\nYou have ${entries} entries for ₦${amount.toLocaleString()}\nPoints earned: ${totalPoints}\n\nGood luck in today's draw!`;
        
        alert(message);
      }, 1000);
      
      return; // Exit early for subscription callback
    }
    
    // Check for recharge success
    const subscriptionStatus = allParams.get('subscription');
    const subscriptionRecorded = allParams.get('subscription_recorded') === 'true';
    const subscriptionAmount = parseFloat(allParams.get('amount') || '0');
    const subscriptionEntries = parseInt(allParams.get('entries') || '0');
    const subscriptionMsisdn = allParams.get('msisdn');
    const subscriptionRef = allParams.get('ref');
    
    if (subscriptionStatus === 'success' && subscriptionRecorded && subscriptionAmount > 0) {
      // Subscription success detected
      
      // Clear URL immediately
      window.history.replaceState({}, document.title, window.location.pathname);
      
      // Show subscription success message
      setTimeout(() => {
        toast({
          title: "Subscription Successful! 🎉",
          description: `You've subscribed with ${subscriptionMsisdn} for ${subscriptionEntries} ${subscriptionEntries === 1 ? 'entry' : 'entries'}. Login with your phone number to view your subscription.`,
          duration: 8000,
        });
      }, 1000);
      
      return; // Exit early for subscription callback
    }
    
    // Handle subscription failure
    if (subscriptionStatus === 'failed') {
      
      // Clear URL immediately
      window.history.replaceState({}, document.title, window.location.pathname);
      
      // Show subscription failure message
      setTimeout(() => {
        toast({
          title: "Subscription Failed",
          description: "There was an issue processing your subscription. Please try again.",
          variant: "destructive",
          duration: 6000,
        });
      }, 1000);
      
      return; // Exit early for subscription failure
    }
    
    // Enhanced hash-based parameter handling with new spin system
    const amount = parseFloat(allParams.get('amount') || '0');
    const spinEligible = allParams.get('spin') === 'true';
    const points = parseInt(allParams.get('points') || '0');
    const phone = allParams.get('phone');
    const network = allParams.get('network');
    
    // New spin system parameters
    const spinEarned = parseInt(allParams.get('spinEarned') || '0');
    const totalSpins = parseInt(allParams.get('totalSpins') || '0');
    const newTier = allParams.get('newTier');
    const cumulativeAmount = parseFloat(allParams.get('cumulativeAmount') || '0');
    
    if (paymentStatus === 'success' && reference) {
      const transactionRef = reference || 'CALLBACK_SUCCESS';
      
      // Fetch transaction details from backend
      fetchTransactionDetails(transactionRef);
      
      async function fetchTransactionDetails(ref: string) {
        // Calculate expected points (1 point per ₦200)
        const expectedPoints = Math.floor((amount || 0) / 200);
        
        // Show processing message with expected points
        alert(`⏳ Processing your recharge...\n\nAmount: ₦${(amount || 0).toLocaleString()}\nExpected Points: ${expectedPoints}\n\nPlease wait while we complete your transaction.`);
        
        const maxAttempts = 20; // Poll for up to 20 seconds
        const pollInterval = 1000; // Check every 1 second
        
        for (let attempt = 0; attempt < maxAttempts; attempt++) {
          try {
            const response = await apiClient.get(`/recharge/reference/${ref}`);
            const result = response.data;
            const txn = result.data;
            
            // Check if transaction is completed
            if (txn.status === 'COMPLETED' || txn.status === 'SUCCESS') {
              // Transaction completed successfully
              const actualAmount = parseFloat(txn.amount || amount || '0');
              const actualPhone = txn.msisdn || phone || 'null';
              const actualNetwork = txn.network_provider || network || 'null';
              const actualPoints = parseInt(txn.points_earned || points || '0');
              
              proceedWithRechargeSuccess(actualAmount, actualPhone, actualNetwork, actualPoints);
              return;
            } else if (txn.status === 'FAILED' || txn.status === 'REFUNDED') {
              // Transaction failed
              alert(`❌ Recharge Failed\n\n${txn.failure_reason || 'Transaction could not be completed'}\n\nYour payment has been ${txn.status === 'REFUNDED' ? 'refunded' : 'will be refunded'}.`);
              return;
            }
            
            // Still pending/processing, wait and try again
            await new Promise(resolve => setTimeout(resolve, pollInterval));
          } catch (error) {
            console.error('Error fetching transaction:', error);
            // Continue polling on error
            await new Promise(resolve => setTimeout(resolve, pollInterval));
          }
        }
        
        // Timeout - show message with current status
        alert('⏳ Your recharge is still being processed.\n\nThis may take a few minutes. You will receive an SMS confirmation once completed.');
      }
      
      function proceedWithRechargeSuccess(txnAmount?: number, txnPhone?: string, txnNetwork?: string, txnPoints?: number) {
        const finalAmount = txnAmount || amount || 0;
        const finalPhone = txnPhone || phone || 'null';
        const finalNetwork = txnNetwork || network || 'null';
        const finalPoints = txnPoints || points || 0;
        // CRITICAL: Clear URL parameters immediately to prevent multiple spins
        const cleanUrl = window.location.pathname;
        window.history.replaceState({}, document.title, cleanUrl);
        
        // Set recharge success data
        setRechargeSuccess({
          amount: finalAmount,
          points: finalPoints,
          spinEligible,
          phone: finalPhone,
          network: finalNetwork,
          transactionReference: transactionRef
        });
        
        // Store user phone for spin system
        if (finalPhone && finalPhone !== 'null') {
          setUserPhone(finalPhone);
        }
        
        // Show enhanced success popup with new spin system info
        setTimeout(() => {
          let message = `🎉 Recharge Successful!\n\nAmount: ₦${finalAmount.toLocaleString()}\nPhone: ${finalPhone}\nNetwork: ${finalNetwork}\nPoints Earned: ${finalPoints}`;
          
          if (newTier && cumulativeAmount > 0) {
            message += `\n\n🏆 Daily Progress:\nTotal Today: ₦${cumulativeAmount.toLocaleString()}\nCurrent Tier: ${newTier}`;
            
            if (spinEarned > 0) {
              message += `\n\n🎆 New Spins Earned: ${spinEarned}\nTotal Available: ${totalSpins}`;
            }
          }
          
          if (totalSpins > 0) {
            message += `\n\n🎰 Wheel Spin Available!\nYou have ${totalSpins} spins ready!`;
          } else {
            message += `\n\nRecharge more today to unlock wheel spins!`;
          }
          
          alert(message);
          
          // Update available spins and show wheel if spins available
          setAvailableSpins(totalSpins);
          if (totalSpins > 0) {
            setTimeout(() => {
              setShowSpinWheel(true);
            }, 1000);
          }
        }, 1000);
      }
    }
  }, []);

  const fetchPlatformData = async () => {
    try {
      // Get platform statistics from Go backend
      const statsResponse = await getPlatformStatistics();
      
      if ('success' in statsResponse && statsResponse.success && statsResponse.data) {
        setStats({
          totalUsers: statsResponse.data.totalUsers || 0,
          totalRecharges: statsResponse.data.totalTransactions || 0,
          totalPrizes: statsResponse.data.totalPrizes || 0,
          activeDraw: statsResponse.data.activeDraw || null
        });
      }

      // Get recent winners from Go backend
      const winnersResponse = await getRecentWinners(4);

      if ('success' in winnersResponse && winnersResponse.success && winnersResponse.data && Array.isArray(winnersResponse.data) && winnersResponse.data.length > 0) {
        setRecentWinners(winnersResponse.data.map((winner: any, index: number) => ({
          id: `real_${index}`,
          name: winner.full_name,
          prize: winner.prize_description,
          amount: winner.prize_value,
          time: new Date(winner.created_at).toLocaleString(),
          network: winner.network_provider || 'MTN'
        })));
      }
    } catch (error) {
      console.error('Error fetching platform data:', error);
    }
  };
  const [timeRemaining, setTimeRemaining] = useState('');

  useEffect(() => {
    if (!stats.activeDraw) return;

    const updateTimer = () => {
      const now = new Date().getTime();
      const endTime = new Date(stats.activeDraw!.endTime).getTime();
      const difference = endTime - now;

      if (difference > 0) {
        const hours = Math.floor((difference % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
        const minutes = Math.floor((difference % (1000 * 60 * 60)) / (1000 * 60));
        const seconds = Math.floor((difference % (1000 * 60)) / 1000);
        setTimeRemaining(`${hours}h ${minutes}m ${seconds}s`);
      } else {
        setTimeRemaining('Draw Ended');
      }
    };

    updateTimer();
    const timer = setInterval(updateTimer, 1000);
    return () => clearInterval(timer);
  }, [stats.activeDraw]);
  const features = [
    {
      icon: <Smartphone className="w-8 h-8" />,
      title: "Instant Recharge",
      description: "Quick and secure mobile recharge for all Nigerian networks",
      color: "bg-blue-500"
    },
    {
      icon: <Trophy className="w-8 h-8" />,
      title: "Daily Draws",
      description: "Win cash prizes up to ₦500,000 in our daily prize draws",
      color: "bg-yellow-500"
    },
    {
      icon: <Sparkles className="w-8 h-8" />,
      title: "Spin & Win",
      description: "Unlock instant prizes with recharges of ₦1,000 or more",
      color: "bg-purple-500"
    },
    {
      icon: <Users className="w-8 h-8" />,
      title: "Refer & Earn",
      description: "Earn commission by referring friends to RechargeMax",
      color: "bg-green-500"
    }
  ];

  const benefits = [
    "Every ₦200 recharge = 1 draw entry",
    "Recharge ₦1,000+ to unlock Spin Wheel",
    "Daily subscription for guaranteed entries",
    "Earn points and climb loyalty tiers",
    "Refer friends and earn commissions",
    "Secure payments with top providers"
  ];

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-purple-50">
      {/* Hero Section */}
      <section className="relative overflow-hidden bg-gradient-to-r from-blue-600 to-blue-800 text-white">
        <div className="absolute inset-0 bg-black/10"></div>
        <div className="relative container mx-auto px-4 py-16 md:py-24">
          <div className="grid lg:grid-cols-2 gap-12 items-center">
            <div className="space-y-8">
              <div className="space-y-4">
                <Badge className="bg-white/20 text-white border-white/30 hover:bg-white/30">
                  <Flame className="w-4 h-4 mr-2" />
                  Nigeria's #1 Gamified Recharge Platform
                </Badge>
                <h1 className="text-4xl md:text-6xl font-bold leading-tight">
                  Recharge & 
                  <span className="bg-gradient-to-r from-yellow-300 to-orange-300 bg-clip-text text-transparent">
                    Win Big
                  </span>
                </h1>
                <p className="text-xl md:text-2xl text-blue-100 leading-relaxed">
                  Turn every mobile recharge into a chance to win amazing prizes. 
                  Join thousands of Nigerians winning daily!
                </p>
              </div>
              
              <div className="flex flex-wrap gap-4">
                <Button 
                  size="lg" 
                  className="bg-white text-blue-600 hover:bg-blue-50 font-semibold px-8 py-4 text-lg"
                  onClick={() => document.getElementById('recharge-form')?.scrollIntoView({ behavior: 'smooth' })}
                >
                  <Zap className="w-5 h-5 mr-2" />
                  Start Recharging
                </Button>
                <Button 
                  size="lg" 
                  variant="outline" 
                  className="border-white/30 text-white hover:bg-white/10 font-semibold px-8 py-4 text-lg"
                >
                  <Trophy className="w-5 h-5 mr-2" />
                  View Prizes
                </Button>
              </div>
              
              <div className="grid grid-cols-3 gap-6 pt-8">
                <div className="text-center">
                  <div className="text-3xl font-bold">{stats.totalUsers.toLocaleString()}+</div>
                  <div className="text-blue-200 text-sm">Happy Users</div>
                </div>
                <div className="text-center">
                  <div className="text-3xl font-bold">{formatCurrency(stats.totalPrizes * 1000)}</div>
                  <div className="text-blue-200 text-sm">Prizes Won</div>
                </div>
                <div className="text-center">
                  <div className="text-3xl font-bold">{stats.totalRecharges.toLocaleString()}+</div>
                  <div className="text-blue-200 text-sm">Recharges</div>
                </div>
              </div>
            </div>
            
            <div className="relative">
              <div className="absolute inset-0 bg-gradient-to-r from-yellow-400 to-orange-500 rounded-3xl blur-3xl opacity-30 animate-pulse"></div>
              <img 
                src="/images/recharge_hero_banner_20251109_182859.png" 
                alt="RechargeMax Hero" 
                className="relative z-10 w-full h-auto rounded-2xl shadow-2xl"
              />
            </div>
          </div>
        </div>
      </section>

      {/* Live Stats Bar */}
      <section className="bg-white border-b shadow-sm">
        <div className="container mx-auto px-4 py-4">
          <div className="flex flex-wrap items-center justify-between gap-4">
            <div className="flex items-center gap-6">
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 bg-green-500 rounded-full animate-pulse"></div>
                <span className="text-sm font-medium">Live Draw:</span>
                <span className="text-sm text-green-600 font-semibold">
                  {stats.activeDraw?.name} - {formatCurrency(stats.activeDraw?.prizePool || 0)}
                </span>
              </div>
              <div className="flex items-center gap-2">
                <Clock className="w-4 h-4 text-orange-500" />
                <span className="text-sm font-medium text-orange-600">{timeRemaining}</span>
              </div>
            </div>
            <div className="flex items-center gap-4 text-sm">
              <span><strong>{stats.activeDraw?.entries.toLocaleString()}</strong> entries</span>
              <Button size="sm" variant="outline">
                Join Now
                <ArrowRight className="w-4 h-4 ml-1" />
              </Button>
            </div>
          </div>
        </div>
      </section>

      <div className="container mx-auto px-4 py-12 space-y-16">
        {/* Success Message */}
        {rechargeSuccess && (
          <Alert className="border-green-200 bg-green-50 animate-slide-up">
            <CheckCircle className="h-5 w-5 text-green-600" />
            <AlertDescription className="text-green-800">
              <strong>Recharge Successful!</strong> Transaction completed successfully!
            </AlertDescription>
          </Alert>
        )}

        {/* Main Recharge Form */}
        <section id="recharge-form" className="scroll-mt-20">
          <div className="text-center mb-8">
            <h2 className="text-3xl font-bold mb-4">Quick Recharge</h2>
            <p className="text-muted-foreground text-lg max-w-2xl mx-auto">
              Recharge your phone in seconds and automatically enter our daily prize draws
            </p>
          </div>
          <PremiumRechargeForm onRechargeSuccess={handleRechargeSuccess} />
          
          {/* Daily Subscription CTA */}
          <div className="mt-8 p-6 bg-gradient-to-r from-blue-50 to-purple-50 rounded-xl border border-blue-200">
            <div className="text-center space-y-4">
              <div className="flex items-center justify-center gap-2">
                <Calendar className="w-6 h-6 text-blue-600" />
                <h3 className="text-xl font-bold text-gray-900">Daily Subscription - Only ₦20!</h3>
              </div>
              <p className="text-gray-600 max-w-md mx-auto">
                Subscribe daily for guaranteed draw entries. Never miss a chance to win!
              </p>
              <Button 
                onClick={() => window.location.href = '/subscription'}
                className="bg-blue-600 hover:bg-blue-700 text-white px-8 py-3 text-lg"
              >
                <Gift className="w-5 h-5 mr-2" />
                Subscribe Now - ₦20
              </Button>
            </div>
          </div>
          
          {/* Spin Wheel Section */}
          <div className="mt-8 text-center">
            <Card className="max-w-md mx-auto">
              <CardHeader>
                <CardTitle className="flex items-center justify-center gap-2">
                  <Gift className="w-5 h-5" />
                  Spin Wheel
                </CardTitle>
                <CardDescription>
                  Recharge ₦1,000+ to unlock the spin wheel and win instant prizes!
                </CardDescription>
              </CardHeader>
              <CardContent>
                {/* Daily Spin Progress */}
                {userPhone && (
                  <div className="mb-6">
                    <DailySpinProgress 
                      msisdn={userPhone} 
                      onSpinsUpdate={(spins) => setAvailableSpins(spins)}
                    />
                  </div>
                )}
                
                <div className="relative">
                  <div className={`w-32 h-32 mx-auto rounded-full border-4 border-dashed border-gray-300 flex items-center justify-center ${availableSpins > 0 ? 'animate-pulse border-orange-500' : 'opacity-50'}`}>
                    <div className="text-4xl">
                      {availableSpins > 0 ? '🎰' : '🎯'}
                    </div>
                  </div>
                  {availableSpins > 0 ? (
                    <Button 
                      onClick={() => setShowSpinWheel(true)}
                      className="mt-4 w-full bg-gradient-to-r from-orange-500 to-red-500 hover:from-orange-600 hover:to-red-600"
                    >
                      🎰 Spin Now!
                    </Button>
                  ) : (
                    <Button disabled className="mt-4 w-full opacity-50">
                      Recharge ₦1,000+ to unlock
                    </Button>
                  )}
                </div>
              </CardContent>
            </Card>
          </div>
        </section>

        {/* Features Grid */}
        <section>
          <div className="text-center mb-12">
            <h2 className="text-3xl font-bold mb-4">Why Choose RechargeMax?</h2>
            <p className="text-muted-foreground text-lg max-w-3xl mx-auto">
              We've revolutionized mobile recharging by adding excitement, rewards, and real value to every transaction
            </p>
          </div>
          
          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
            {features.map((feature, index) => (
              <Card key={index} className="group hover:shadow-lg transition-all duration-300 hover:-translate-y-1">
                <CardContent className="p-6 text-center">
                  <div className={`w-16 h-16 ${feature.color} rounded-full flex items-center justify-center text-white mx-auto mb-4 group-hover:scale-110 transition-transform duration-300`}>
                    {feature.icon}
                  </div>
                  <h3 className="font-semibold text-lg mb-2">{feature.title}</h3>
                  <p className="text-muted-foreground text-sm">{feature.description}</p>
                </CardContent>
              </Card>
            ))}
          </div>
        </section>

        {/* How It Works */}
        <section className="bg-gradient-to-r from-blue-50 to-purple-50 rounded-3xl p-8 md:p-12">
          <div className="text-center mb-12">
            <h2 className="text-3xl font-bold mb-4">How It Works</h2>
            <p className="text-muted-foreground text-lg">
              Simple steps to start winning with every recharge
            </p>
          </div>
          
          <div className="grid md:grid-cols-3 gap-8">
            <div className="text-center space-y-4">
              <div className="w-16 h-16 bg-blue-500 rounded-full flex items-center justify-center text-white text-2xl font-bold mx-auto">
                1
              </div>
              <h3 className="font-semibold text-lg">Recharge Your Phone</h3>
              <p className="text-muted-foreground">Choose your network, enter amount, and complete payment securely</p>
            </div>
            
            <div className="text-center space-y-4">
              <div className="w-16 h-16 bg-purple-500 rounded-full flex items-center justify-center text-white text-2xl font-bold mx-auto">
                2
              </div>
              <h3 className="font-semibold text-lg">Earn Entries & Points</h3>
              <p className="text-muted-foreground">Every ₦200 = 1 draw entry. Collect points for loyalty rewards</p>
            </div>
            
            <div className="text-center space-y-4">
              <div className="w-16 h-16 bg-green-500 rounded-full flex items-center justify-center text-white text-2xl font-bold mx-auto">
                3
              </div>
              <h3 className="font-semibold text-lg">Win Amazing Prizes</h3>
              <p className="text-muted-foreground">Daily draws, instant spin prizes, and exclusive rewards</p>
            </div>
          </div>
        </section>

        {/* Benefits List */}
        <section>
          <div className="grid lg:grid-cols-2 gap-12 items-center">
            <div>
              <h2 className="text-3xl font-bold mb-6">Everything You Need</h2>
              <div className="space-y-4">
                {benefits.map((benefit, index) => (
                  <div key={index} className="flex items-center gap-3">
                    <div className="w-6 h-6 bg-green-100 rounded-full flex items-center justify-center flex-shrink-0">
                      <CheckCircle className="w-4 h-4 text-green-600" />
                    </div>
                    <span className="text-muted-foreground">{benefit}</span>
                  </div>
                ))}
              </div>
              
              <div className="mt-8 p-6 bg-gradient-nigeria rounded-xl text-white">
                <h3 className="font-bold text-lg mb-2">🇳🇬 Proudly Nigerian</h3>
                <p className="text-green-100">
                  Built specifically for Nigerian mobile users with local payment methods, 
                  network integrations, and customer support.
                </p>
              </div>
            </div>
            
            <div className="space-y-6">
              <Card className="p-6 border-2 border-yellow-200 bg-yellow-50">
                <div className="flex items-center gap-4">
                  <div className="w-12 h-12 bg-yellow-500 rounded-full flex items-center justify-center">
                    <DollarSign className="w-6 h-6 text-white" />
                  </div>
                  <div>
                    <h3 className="font-bold text-lg">Daily Prize Pool</h3>
                    <p className="text-2xl font-bold text-yellow-600">
                      {formatCurrency(stats.activeDraw?.prizePool || 0)}
                    </p>
                  </div>
                </div>
              </Card>
              
              <Card className="p-6 border-2 border-blue-200 bg-blue-50">
                <div className="flex items-center gap-4">
                  <div className="w-12 h-12 bg-blue-500 rounded-full flex items-center justify-center">
                    <Users className="w-6 h-6 text-white" />
                  </div>
                  <div>
                    <h3 className="font-bold text-lg">Active Users</h3>
                    <p className="text-2xl font-bold text-blue-600">
                      {stats.totalUsers.toLocaleString()}+
                    </p>
                  </div>
                </div>
              </Card>
              
              <Card className="p-6 border-2 border-green-200 bg-green-50">
                <div className="flex items-center gap-4">
                  <div className="w-12 h-12 bg-green-500 rounded-full flex items-center justify-center">
                    <Award className="w-6 h-6 text-white" />
                  </div>
                  <div>
                    <h3 className="font-bold text-lg">Prizes Distributed</h3>
                    <p className="text-2xl font-bold text-green-600">
                      {stats.totalPrizes.toLocaleString()}+
                    </p>
                  </div>
                </div>
              </Card>
            </div>
          </div>
        </section>

        {/* Active Draws */}
        <section>
          <div className="text-center mb-8">
            <h2 className="text-3xl font-bold mb-4">Active Prize Draws</h2>
            <p className="text-muted-foreground text-lg">
              Join these exciting draws and win amazing prizes
            </p>
          </div>
          <DrawsList />
        </section>

        {/* CTA Section */}
        <section className="text-center bg-gradient-primary rounded-3xl p-12 text-white">
          <div className="max-w-3xl mx-auto space-y-6">
            <h2 className="text-4xl font-bold">Ready to Start Winning?</h2>
            <p className="text-xl text-blue-100">
              Join thousands of Nigerians who are already turning their mobile recharges into exciting prizes!
            </p>
            <div className="flex flex-wrap gap-4 justify-center">
              <Button 
                size="lg" 
                className="bg-white text-blue-600 hover:bg-blue-50 font-semibold px-8 py-4 text-lg"
                onClick={() => document.getElementById('recharge-form')?.scrollIntoView({ behavior: 'smooth' })}
              >
                <Smartphone className="w-5 h-5 mr-2" />
                Recharge Now
              </Button>
              {!isAuthenticated && (
                <Button 
                  size="lg" 
                  variant="outline" 
                  className="border-white/30 text-white hover:bg-white/10 font-semibold px-8 py-4 text-lg"
                >
                  <Users className="w-5 h-5 mr-2" />
                  Create Account
                </Button>
              )}
            </div>
          </div>
        </section>
      </div>
      
      {/* Enhanced Spin Wheel Modal - Backend-Driven Prize Selection */}
      {showSpinWheel && availableSpins > 0 && (
        <SpinWheel
          isOpen={showSpinWheel}
          onClose={() => {
            setShowSpinWheel(false);
            setAvailableSpins(0); // Reset available spins after closing
          }}
          transactionAmount={rechargeSuccess?.amount || 1000}
          userPhone={userPhone || ''}
          onPrizeWon={(prize) => {
            // Prize already recorded in backend by /spin/play endpoint
            // No need for additional API calls - backend handles everything
          }}
        />
      )}
    </div>
  );
};

export default EnterpriseHomePage;