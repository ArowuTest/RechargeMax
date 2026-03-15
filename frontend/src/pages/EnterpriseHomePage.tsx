import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { Separator } from '@/components/ui/separator';
import { useAuthContext } from '@/contexts/AuthContext';
import { LoginModal } from '@/components/auth/LoginModal';
import { DrawsList } from '@/components/draws/DrawsList';
import { formatCurrency, formatRelativeTime } from '@/lib/utils';
import { 
  Zap, 
  Gift, 
  Trophy, 
  TrendingUp, 
  Users, 
  Star, 
  Smartphone, 
  CreditCard,
  Target,
  Clock,
  CheckCircle,
  ArrowRight,
  Sparkles,
  Crown,
  DollarSign,
  Phone,
  Wifi,
  Shield,
  Award,
  Calendar,
  BarChart3,
  Gamepad2,
  Heart,
  Coins
} from 'lucide-react';

interface HomePageProps {
  className?: string;
}

export const EnterpriseHomePage: React.FC<HomePageProps> = ({ className = "" }) => {
  const { isAuthenticated, user } = useAuthContext();
  const [showLoginModal, setShowLoginModal] = useState(false);
  const [currentTime, setCurrentTime] = useState(new Date());

  // Update time every minute for draw countdown
  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentTime(new Date());
    }, 60000);

    return () => clearInterval(timer);
  }, []);

  const handleLoginRequired = () => {
    setShowLoginModal(true);
  };

  const getNextDrawTime = () => {
    const now = new Date();
    const nextDraw = new Date();
    nextDraw.setHours(20, 0, 0, 0); // 8 PM daily draw
    
    if (now.getHours() >= 20) {
      nextDraw.setDate(nextDraw.getDate() + 1);
    }
    
    const diff = nextDraw.getTime() - now.getTime();
    const hours = Math.floor(diff / (1000 * 60 * 60));
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
    
    return `${hours}h ${minutes}m`;
  };

  return (
    <div className={`min-h-screen bg-gradient-to-br from-blue-50 via-white to-purple-50 ${className}`}>
      {/* Hero Section */}
      <section className="relative overflow-hidden bg-gradient-to-r from-blue-600 via-purple-600 to-blue-800 text-white">
        <div className="absolute inset-0 bg-black/20"></div>
        <div className="relative max-w-7xl mx-auto px-4 py-20 sm:py-24">
          <div className="text-center space-y-8">
            <div className="flex items-center justify-center gap-3 mb-6">
              <div className="p-3 bg-white/10 rounded-full backdrop-blur-sm">
                <Zap className="w-8 h-8" />
              </div>
              <h1 className="text-5xl sm:text-6xl font-bold">
                Recharge<span className="text-yellow-300">Max</span>
              </h1>
            </div>
            
            <p className="text-xl sm:text-2xl text-blue-100 max-w-3xl mx-auto leading-relaxed">
              Nigeria's Premier Rewards Platform - Recharge, Win, Earn!
            </p>
            
            <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
              <Button 
                size="lg" 
                className="bg-yellow-500 hover:bg-yellow-600 text-black font-semibold px-8 py-4 text-lg"
                onClick={() => window.location.href = '/#/recharge'}
              >
                <Smartphone className="w-5 h-5 mr-2" />
                Start Recharging
              </Button>
              
              <Button 
                size="lg" 
                variant="outline" 
                className="border-white text-white hover:bg-white hover:text-blue-600 px-8 py-4 text-lg"
                onClick={() => window.location.href = '/#/draws'}
              >
                <Trophy className="w-5 h-5 mr-2" />
                View Draws
              </Button>
            </div>

            {/* Quick Stats */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-6 mt-12 max-w-4xl mx-auto">
              <div className="text-center">
                <div className="text-3xl font-bold text-yellow-300">₦500K</div>
                <div className="text-blue-100">Daily Jackpot</div>
              </div>
              <div className="text-center">
                <div className="text-3xl font-bold text-yellow-300">15%</div>
                <div className="text-blue-100">Max Commission</div>
              </div>
              <div className="text-center">
                <div className="text-3xl font-bold text-yellow-300">24/7</div>
                <div className="text-blue-100">Service</div>
              </div>
              <div className="text-center">
                <div className="text-3xl font-bold text-yellow-300">100K+</div>
                <div className="text-blue-100">Happy Users</div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-20 bg-white">
        <div className="max-w-7xl mx-auto px-4">
          <div className="text-center mb-16">
            <h2 className="text-4xl font-bold text-gray-900 mb-4">Why Choose RechargeMax?</h2>
            <p className="text-xl text-gray-600 max-w-3xl mx-auto">
              Experience the future of mobile recharging with rewards, games, and incredible prizes
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            {/* Instant Recharge */}
            <Card className="group hover:shadow-xl transition-all duration-300 border-2 hover:border-blue-200">
              <CardHeader className="text-center">
                <div className="w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-4 group-hover:bg-blue-200 transition-colors">
                  <Zap className="w-8 h-8 text-blue-600" />
                </div>
                <CardTitle className="text-xl">Instant Recharge</CardTitle>
                <CardDescription>
                  Lightning-fast airtime and data top-ups for all Nigerian networks
                </CardDescription>
              </CardHeader>
              <CardContent>
                <ul className="space-y-2 text-sm text-gray-600">
                  <li className="flex items-center gap-2">
                    <CheckCircle className="w-4 h-4 text-green-500" />
                    MTN, Airtel, Glo, 9mobile
                  </li>
                  <li className="flex items-center gap-2">
                    <CheckCircle className="w-4 h-4 text-green-500" />
                    Instant delivery
                  </li>
                  <li className="flex items-center gap-2">
                    <CheckCircle className="w-4 h-4 text-green-500" />
                    Secure payments
                  </li>
                </ul>
              </CardContent>
            </Card>

            {/* Daily Draws */}
            <Card className="group hover:shadow-xl transition-all duration-300 border-2 hover:border-purple-200">
              <CardHeader className="text-center">
                <div className="w-16 h-16 bg-purple-100 rounded-full flex items-center justify-center mx-auto mb-4 group-hover:bg-purple-200 transition-colors">
                  <Trophy className="w-8 h-8 text-purple-600" />
                </div>
                <CardTitle className="text-xl">Daily Draws</CardTitle>
                <CardDescription>
                  Win cash prizes up to ₦500,000 in our daily jackpot draws
                </CardDescription>
              </CardHeader>
              <CardContent>
                <ul className="space-y-2 text-sm text-gray-600">
                  <li className="flex items-center gap-2">
                    <CheckCircle className="w-4 h-4 text-green-500" />
                    Multiple draws daily
                  </li>
                  <li className="flex items-center gap-2">
                    <CheckCircle className="w-4 h-4 text-green-500" />
                    Automatic entries
                  </li>
                  <li className="flex items-center gap-2">
                    <CheckCircle className="w-4 h-4 text-green-500" />
                    Real cash prizes
                  </li>
                </ul>
              </CardContent>
            </Card>

            {/* Spin Wheel */}
            <Card className="group hover:shadow-xl transition-all duration-300 border-2 hover:border-yellow-200">
              <CardHeader className="text-center">
                <div className="w-16 h-16 bg-yellow-100 rounded-full flex items-center justify-center mx-auto mb-4 group-hover:bg-yellow-200 transition-colors">
                  <Gamepad2 className="w-8 h-8 text-yellow-600" />
                </div>
                <CardTitle className="text-xl">Spin Wheel</CardTitle>
                <CardDescription>
                  Spin to win instant rewards on recharges of ₦1,000 and above
                </CardDescription>
              </CardHeader>
              <CardContent>
                <ul className="space-y-2 text-sm text-gray-600">
                  <li className="flex items-center gap-2">
                    <CheckCircle className="w-4 h-4 text-green-500" />
                    Instant rewards
                  </li>
                  <li className="flex items-center gap-2">
                    <CheckCircle className="w-4 h-4 text-green-500" />
                    Multiple spins
                  </li>
                  <li className="flex items-center gap-2">
                    <CheckCircle className="w-4 h-4 text-green-500" />
                    Bonus prizes
                  </li>
                </ul>
              </CardContent>
            </Card>

            {/* Loyalty Program */}
            <Card className="group hover:shadow-xl transition-all duration-300 border-2 hover:border-green-200">
              <CardHeader className="text-center">
                <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4 group-hover:bg-green-200 transition-colors">
                  <Crown className="w-8 h-8 text-green-600" />
                </div>
                <CardTitle className="text-xl">Loyalty Tiers</CardTitle>
                <CardDescription>
                  Unlock better rewards as you recharge more with our tier system
                </CardDescription>
              </CardHeader>
              <CardContent>
                <ul className="space-y-2 text-sm text-gray-600">
                  <li className="flex items-center gap-2">
                    <CheckCircle className="w-4 h-4 text-green-500" />
                    Bronze to Diamond
                  </li>
                  <li className="flex items-center gap-2">
                    <CheckCircle className="w-4 h-4 text-green-500" />
                    Better rewards
                  </li>
                  <li className="flex items-center gap-2">
                    <CheckCircle className="w-4 h-4 text-green-500" />
                    Exclusive benefits
                  </li>
                </ul>
              </CardContent>
            </Card>

            {/* Affiliate Program */}
            <Card className="group hover:shadow-xl transition-all duration-300 border-2 hover:border-red-200">
              <CardHeader className="text-center">
                <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4 group-hover:bg-red-200 transition-colors">
                  <Users className="w-8 h-8 text-red-600" />
                </div>
                <CardTitle className="text-xl">Affiliate Program</CardTitle>
                <CardDescription>
                  Earn up to 15% commission by referring friends and family
                </CardDescription>
              </CardHeader>
              <CardContent>
                <ul className="space-y-2 text-sm text-gray-600">
                  <li className="flex items-center gap-2">
                    <CheckCircle className="w-4 h-4 text-green-500" />
                    Up to 15% commission
                  </li>
                  <li className="flex items-center gap-2">
                    <CheckCircle className="w-4 h-4 text-green-500" />
                    Weekly payouts
                  </li>
                  <li className="flex items-center gap-2">
                    <CheckCircle className="w-4 h-4 text-green-500" />
                    Real-time tracking
                  </li>
                </ul>
              </CardContent>
            </Card>

            {/* Daily Subscription */}
            <Card className="group hover:shadow-xl transition-all duration-300 border-2 hover:border-indigo-200">
              <CardHeader className="text-center">
                <div className="w-16 h-16 bg-indigo-100 rounded-full flex items-center justify-center mx-auto mb-4 group-hover:bg-indigo-200 transition-colors">
                  <Calendar className="w-8 h-8 text-indigo-600" />
                </div>
                <CardTitle className="text-xl">Daily Subscription</CardTitle>
                <CardDescription>
                  Subscribe for guaranteed daily draw entries at just ₦20 per entry
                </CardDescription>
              </CardHeader>
              <CardContent>
                <ul className="space-y-2 text-sm text-gray-600">
                  <li className="flex items-center gap-2">
                    <CheckCircle className="w-4 h-4 text-green-500" />
                    Guaranteed entries
                  </li>
                  <li className="flex items-center gap-2">
                    <CheckCircle className="w-4 h-4 text-green-500" />
                    Affordable pricing
                  </li>
                  <li className="flex items-center gap-2">
                    <CheckCircle className="w-4 h-4 text-green-500" />
                    Never miss a draw
                  </li>
                </ul>
              </CardContent>
            </Card>
          </div>
        </div>
      </section>

      {/* How It Works */}
      <section className="py-20 bg-gray-50">
        <div className="max-w-7xl mx-auto px-4">
          <div className="text-center mb-16">
            <h2 className="text-4xl font-bold text-gray-900 mb-4">How It Works</h2>
            <p className="text-xl text-gray-600">Simple steps to start winning</p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
            <div className="text-center">
              <div className="w-20 h-20 bg-blue-600 text-white rounded-full flex items-center justify-center mx-auto mb-6 text-2xl font-bold">
                1
              </div>
              <h3 className="text-xl font-semibold mb-3">Recharge Your Phone</h3>
              <p className="text-gray-600">
                Top up your airtime or data on any Nigerian network
              </p>
            </div>

            <div className="text-center">
              <div className="w-20 h-20 bg-purple-600 text-white rounded-full flex items-center justify-center mx-auto mb-6 text-2xl font-bold">
                2
              </div>
              <h3 className="text-xl font-semibold mb-3">Earn Entries</h3>
              <p className="text-gray-600">
                Get automatic draw entries for every ₦200 you recharge
              </p>
            </div>

            <div className="text-center">
              <div className="w-20 h-20 bg-yellow-600 text-white rounded-full flex items-center justify-center mx-auto mb-6 text-2xl font-bold">
                3
              </div>
              <h3 className="text-xl font-semibold mb-3">Play & Win</h3>
              <p className="text-gray-600">
                Spin the wheel on ₦1,000+ recharges and join daily draws
              </p>
            </div>

            <div className="text-center">
              <div className="w-20 h-20 bg-green-600 text-white rounded-full flex items-center justify-center mx-auto mb-6 text-2xl font-bold">
                4
              </div>
              <h3 className="text-xl font-semibold mb-3">Get Rewarded</h3>
              <p className="text-gray-600">
                Win cash prizes, bonuses, and climb loyalty tiers
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Live Draws Section */}
      <section className="py-20 bg-white">
        <div className="max-w-7xl mx-auto px-4">
          <div className="text-center mb-12">
            <h2 className="text-4xl font-bold text-gray-900 mb-4">Live Draws</h2>
            <p className="text-xl text-gray-600">Join thousands of players in our daily draws</p>
          </div>

          <DrawsList onLoginRequired={handleLoginRequired} />
        </div>
      </section>

      {/* Next Draw Countdown */}
      <section className="py-16 bg-gradient-to-r from-purple-600 to-blue-600 text-white">
        <div className="max-w-4xl mx-auto px-4 text-center">
          <h2 className="text-3xl font-bold mb-4">Next Daily Draw</h2>
          <div className="text-6xl font-bold mb-4">{getNextDrawTime()}</div>
          <p className="text-xl text-purple-100 mb-8">
            Don't miss your chance to win ₦500,000!
          </p>
          
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Button 
              size="lg" 
              className="bg-yellow-500 hover:bg-yellow-600 text-black font-semibold"
              onClick={() => window.location.href = '/#/recharge'}
            >
              <Zap className="w-5 h-5 mr-2" />
              Recharge Now
            </Button>
            
            <Button 
              size="lg" 
              variant="outline" 
              className="border-white text-white hover:bg-white hover:text-purple-600"
              onClick={() => window.location.href = '/#/subscription'}
            >
              <Calendar className="w-5 h-5 mr-2" />
              Subscribe Daily
            </Button>
          </div>
        </div>
      </section>

      {/* Recent Winners */}
      <section className="py-20 bg-gray-50">
        <div className="max-w-7xl mx-auto px-4">
          <div className="text-center mb-12">
            <h2 className="text-4xl font-bold text-gray-900 mb-4">Recent Winners</h2>
            <p className="text-xl text-gray-600">Congratulations to our latest prize winners!</p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {[
              { phone: '0803****567', prize: 'Daily Jackpot', amount: 500000, time: '2 hours ago', tier: 'Gold' },
              { phone: '0807****321', prize: 'Spin Wheel', amount: 50000, time: '4 hours ago', tier: 'Silver' },
              { phone: '0809****876', prize: 'Daily Draw', amount: 100000, time: '6 hours ago', tier: 'Platinum' },
              { phone: '0806****210', prize: 'Weekly Mega', amount: 1000000, time: '1 day ago', tier: 'Diamond' },
              { phone: '0812****543', prize: 'Spin Wheel', amount: 25000, time: '1 day ago', tier: 'Bronze' },
              { phone: '0815****987', prize: 'Daily Draw', amount: 75000, time: '2 days ago', tier: 'Gold' },
            ].map((winner, index) => (
              <Card key={index} className="hover:shadow-lg transition-shadow">
                <CardContent className="p-6">
                  <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center gap-3">
                      <div className="w-12 h-12 bg-yellow-100 rounded-full flex items-center justify-center">
                        <Trophy className="w-6 h-6 text-yellow-600" />
                      </div>
                      <div>
                        <p className="font-semibold">{winner.phone}</p>
                        <Badge variant="secondary" className="text-xs">
                          {winner.tier} Member
                        </Badge>
                      </div>
                    </div>
                    <div className="text-right">
                      <p className="font-bold text-green-600 text-lg">
                        {formatCurrency(winner.amount)}
                      </p>
                      <p className="text-xs text-gray-500">{winner.time}</p>
                    </div>
                  </div>
                  <p className="text-sm text-gray-600 text-center">
                    Won in <span className="font-medium">{winner.prize}</span>
                  </p>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-20 bg-gradient-to-r from-blue-600 to-purple-600 text-white">
        <div className="max-w-4xl mx-auto px-4 text-center">
          <h2 className="text-4xl font-bold mb-4">Ready to Start Winning?</h2>
          <p className="text-xl text-blue-100 mb-8">
            Join thousands of Nigerians who are already earning rewards with every recharge
          </p>
          
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            {!isAuthenticated ? (
              <>
                <Button 
                  size="lg" 
                  className="bg-yellow-500 hover:bg-yellow-600 text-black font-semibold px-8 py-4"
                  onClick={() => setShowLoginModal(true)}
                >
                  <Phone className="w-5 h-5 mr-2" />
                  Login to Start
                </Button>
                
                <Button 
                  size="lg" 
                  variant="outline" 
                  className="border-white text-white hover:bg-white hover:text-blue-600 px-8 py-4"
                  onClick={() => window.location.href = '/#/recharge'}
                >
                  <Zap className="w-5 h-5 mr-2" />
                  Recharge Now
                </Button>
              </>
            ) : (
              <>
                <Button 
                  size="lg" 
                  className="bg-yellow-500 hover:bg-yellow-600 text-black font-semibold px-8 py-4"
                  onClick={() => window.location.href = '/#/dashboard'}
                >
                  <BarChart3 className="w-5 h-5 mr-2" />
                  View Dashboard
                </Button>
                
                <Button 
                  size="lg" 
                  variant="outline" 
                  className="border-white text-white hover:bg-white hover:text-blue-600 px-8 py-4"
                  onClick={() => window.location.href = '/#/recharge'}
                >
                  <Zap className="w-5 h-5 mr-2" />
                  Recharge & Win
                </Button>
              </>
            )}
          </div>
        </div>
      </section>

      {/* Login Modal */}
      <LoginModal 
        isOpen={showLoginModal}
        onClose={() => setShowLoginModal(false)}
        onSuccess={() => {
          setShowLoginModal(false);
          // Optionally redirect to dashboard
        }}
      />
    </div>
  );
};

export default EnterpriseHomePage;