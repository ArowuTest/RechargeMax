import React, { useState, useEffect, useCallback } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Progress } from '@/components/ui/progress';
import { formatCurrency, formatRelativeTime } from '@/lib/utils';
import { useAuthContext } from '@/contexts/AuthContext';
import { Clock, Gift, Trophy, Zap, Loader2, AlertCircle, RefreshCw } from 'lucide-react';
import { getActiveDraws, getRecentWinners } from '@/lib/api';
import { useToast } from '@/hooks/use-toast';

interface Draw {
  id: string;
  name: string;
  prize_amount: number;
  end_time: string;
  total_entries: number;
  user_entries: number;
  status: 'active' | 'ended' | 'upcoming';
}

interface Winner {
  id: string;
  phone: string;
  prize: string;
  amount: number;
  won_at: string;
}

interface DrawsListProps {
  onLoginRequired?: () => void;
}

export const DrawsList: React.FC<DrawsListProps> = ({ onLoginRequired }) => {
  const { isAuthenticated, user } = useAuthContext();
  const { toast } = useToast();
  
  // State management
  const [draws, setDraws] = useState<Draw[]>([]);
  const [winners, setWinners] = useState<Winner[]>([]);
  const [isLoadingDraws, setIsLoadingDraws] = useState(true);
  const [isLoadingWinners, setIsLoadingWinners] = useState(true);
  const [drawsError, setDrawsError] = useState<string | null>(null);
  const [winnersError, setWinnersError] = useState<string | null>(null);
  const [retryCount, setRetryCount] = useState(0);

  // Fetch active draws from backend
  const fetchDraws = useCallback(async () => {
    setIsLoadingDraws(true);
    setDrawsError(null);
    
    try {
      const response = await getActiveDraws();
      
      if (response && response.success && response.data) {
        setDraws(response.data);
      } else {
        setDraws([]);
      }
    } catch (error: any) {
      console.error('Failed to fetch draws:', error);
      setDrawsError(error.message || 'Failed to load draws');
      
      // Show toast notification for error
      toast({
        title: "Error Loading Draws",
        description: "Unable to fetch active draws. Please try again.",
        variant: "destructive",
      });
    } finally {
      setIsLoadingDraws(false);
    }
  }, [toast]);

  // Fetch recent winners from backend
  const fetchWinners = useCallback(async () => {
    setIsLoadingWinners(true);
    setWinnersError(null);
    
    try {
      const response = await getRecentWinners(10);
      
      if (response && response.data) {
        setWinners(response.data);
      } else {
        setWinners([]);
      }
    } catch (error: any) {
      console.error('Failed to fetch winners:', error);
      setWinnersError(error.message || 'Failed to load winners');
    } finally {
      setIsLoadingWinners(false);
    }
  }, []);

  // Initial data fetch
  useEffect(() => {
    fetchDraws();
    fetchWinners();
  }, [fetchDraws, fetchWinners]);

  // Auto-refresh every 60 seconds
  useEffect(() => {
    const interval = setInterval(() => {
      fetchDraws();
      fetchWinners();
    }, 60000);

    return () => clearInterval(interval);
  }, [fetchDraws, fetchWinners]);

  // Retry handler
  const handleRetry = () => {
    setRetryCount(prev => prev + 1);
    fetchDraws();
    fetchWinners();
  };

  // Time remaining calculator
  const getTimeRemaining = (endTime: string) => {
    const now = new Date().getTime();
    const end = new Date(endTime).getTime();
    const diff = end - now;

    if (diff <= 0) return 'Ended';

    const days = Math.floor(diff / (1000 * 60 * 60 * 24));
    const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

    if (days > 0) return `${days}d ${hours}h ${minutes}m`;
    if (hours > 0) return `${hours}h ${minutes}m`;
    return `${minutes}m`;
  };

  // Winning chance calculator
  const getWinningChance = (userEntries: number, totalEntries: number) => {
    if (userEntries === 0 || totalEntries === 0) return 0;
    return (userEntries / totalEntries) * 100;
  };

  // Format relative time for winners
  const getRelativeTime = (timestamp: string) => {
    const now = new Date().getTime();
    const then = new Date(timestamp).getTime();
    const diff = now - then;

    const minutes = Math.floor(diff / (1000 * 60));
    const hours = Math.floor(diff / (1000 * 60 * 60));
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));

    if (days > 0) return `${days} day${days > 1 ? 's' : ''} ago`;
    if (hours > 0) return `${hours} hour${hours > 1 ? 's' : ''} ago`;
    if (minutes > 0) return `${minutes} minute${minutes > 1 ? 's' : ''} ago`;
    return 'Just now';
  };

  // Mask phone number for privacy
  const maskPhone = (phone: string) => {
    if (phone.length < 8) return phone;
    const start = phone.substring(0, 4);
    const end = phone.substring(phone.length - 3);
    return `${start}****${end}`;
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="text-center space-y-2">
        <h2 className="text-3xl font-bold">Active Draws</h2>
        <p className="text-muted-foreground">
          Participate in daily draws and win amazing prizes!
        </p>
      </div>

      {/* How it Works */}
      <Card className="bg-gradient-to-r from-blue-50 to-purple-50 border-blue-200">
        <CardContent className="p-6">
          <h3 className="font-semibold mb-3 flex items-center gap-2">
            <Gift className="w-5 h-5 text-blue-600" />
            How to Enter Draws
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
            <div className="flex items-start gap-2">
              <div className="w-6 h-6 bg-blue-600 text-white rounded-full flex items-center justify-center text-xs font-bold">1</div>
              <div>
                <p className="font-medium">Recharge Your Phone</p>
                <p className="text-muted-foreground">Every ₦200 recharge = 1 draw entry</p>
              </div>
            </div>
            <div className="flex items-start gap-2">
              <div className="w-6 h-6 bg-blue-600 text-white rounded-full flex items-center justify-center text-xs font-bold">2</div>
              <div>
                <p className="font-medium">Subscribe Daily (Optional)</p>
                <p className="text-muted-foreground">₦20/day for guaranteed daily entries</p>
              </div>
            </div>
            <div className="flex items-start gap-2">
              <div className="w-6 h-6 bg-blue-600 text-white rounded-full flex items-center justify-center text-xs font-bold">3</div>
              <div>
                <p className="font-medium">Win Amazing Prizes</p>
                <p className="text-muted-foreground">Cash prizes, airtime, and data bundles</p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Loading State for Draws */}
      {isLoadingDraws && (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="w-8 h-8 animate-spin text-blue-600" />
          <span className="ml-3 text-muted-foreground">Loading active draws...</span>
        </div>
      )}

      {/* Error State for Draws */}
      {!isLoadingDraws && drawsError && (
        <Card className="border-red-200 bg-red-50">
          <CardContent className="p-6 text-center">
            <AlertCircle className="w-12 h-12 text-red-600 mx-auto mb-3" />
            <h3 className="font-semibold text-red-900 mb-2">Failed to Load Draws</h3>
            <p className="text-red-700 mb-4">{drawsError}</p>
            <Button onClick={handleRetry} variant="outline" className="border-red-300">
              <RefreshCw className="w-4 h-4 mr-2" />
              Retry
            </Button>
          </CardContent>
        </Card>
      )}

      {/* Empty State for Draws */}
      {!isLoadingDraws && !drawsError && draws.length === 0 && (
        <Card>
          <CardContent className="p-12 text-center">
            <Trophy className="w-16 h-16 text-muted-foreground mx-auto mb-4" />
            <h3 className="font-semibold text-lg mb-2">No Active Draws</h3>
            <p className="text-muted-foreground mb-4">
              There are currently no active draws. Check back soon for new opportunities to win!
            </p>
            <Button onClick={handleRetry} variant="outline">
              <RefreshCw className="w-4 h-4 mr-2" />
              Refresh
            </Button>
          </CardContent>
        </Card>
      )}

      {/* Draws Grid */}
      {!isLoadingDraws && !drawsError && draws.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {draws.map((draw) => {
            const timeRemaining = getTimeRemaining(draw.end_time);
            const winningChance = getWinningChance(draw.user_entries, draw.total_entries);
            
            return (
              <Card key={draw.id} className="relative overflow-hidden hover:shadow-lg transition-shadow">
                {/* Prize Badge */}
                <div className="absolute top-4 right-4">
                  <Badge variant="secondary" className="bg-yellow-100 text-yellow-800">
                    <Trophy className="w-3 h-3 mr-1" />
                    {formatCurrency(draw.prize_amount)}
                  </Badge>
                </div>

                <CardHeader className="pb-3">
                  <CardTitle className="text-xl">{draw.name}</CardTitle>
                  <CardDescription className="flex items-center gap-2">
                    <Clock className="w-4 h-4" />
                    Ends in {timeRemaining}
                  </CardDescription>
                </CardHeader>

                <CardContent className="space-y-4">
                  {/* Entry Stats */}
                  <div className="space-y-2">
                    <div className="flex justify-between text-sm">
                      <span>Total Entries</span>
                      <span className="font-medium">{draw.total_entries.toLocaleString()}</span>
                    </div>
                    
                    {isAuthenticated ? (
                      <>
                        <div className="flex justify-between text-sm">
                          <span>Your Entries</span>
                          <span className="font-medium text-blue-600">{draw.user_entries}</span>
                        </div>
                        
                        {draw.user_entries > 0 && (
                          <div className="space-y-1">
                            <div className="flex justify-between text-xs">
                              <span>Winning Chance</span>
                              <span className="font-medium">{winningChance.toFixed(4)}%</span>
                            </div>
                            <Progress value={winningChance} className="h-2" />
                          </div>
                        )}
                      </>
                    ) : (
                      <div className="bg-muted/50 rounded-lg p-3 text-center">
                        <p className="text-sm text-muted-foreground mb-2">
                          Login to see your entries
                        </p>
                        <Button size="sm" onClick={onLoginRequired}>
                          Login Now
                        </Button>
                      </div>
                    )}
                  </div>

                  {/* Action Button */}
                  <Button className="w-full" variant="outline" onClick={() => window.location.href = '/recharge'}>
                    <Zap className="w-4 h-4 mr-2" />
                    Recharge to Enter
                  </Button>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}

      {/* Recent Winners */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Trophy className="w-5 h-5 text-yellow-500" />
            Recent Winners
          </CardTitle>
          <CardDescription>
            Congratulations to our latest prize winners!
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoadingWinners ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="w-6 h-6 animate-spin text-blue-600" />
              <span className="ml-2 text-sm text-muted-foreground">Loading winners...</span>
            </div>
          ) : winnersError ? (
            <div className="text-center py-8">
              <AlertCircle className="w-8 h-8 text-muted-foreground mx-auto mb-2" />
              <p className="text-sm text-muted-foreground">Failed to load recent winners</p>
            </div>
          ) : winners.length === 0 ? (
            <div className="text-center py-8">
              <Trophy className="w-12 h-12 text-muted-foreground mx-auto mb-2" />
              <p className="text-sm text-muted-foreground">No recent winners yet</p>
            </div>
          ) : (
            <div className="space-y-3">
              {winners.map((winner) => (
                <div key={winner.id} className="flex items-center justify-between py-2 border-b last:border-b-0">
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 bg-yellow-100 rounded-full flex items-center justify-center">
                      <Trophy className="w-4 h-4 text-yellow-600" />
                    </div>
                    <div>
                      <p className="font-medium">{maskPhone(winner.phone)}</p>
                      <p className="text-sm text-muted-foreground">{winner.prize}</p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="font-bold text-green-600">{formatCurrency(winner.amount)}</p>
                    <p className="text-xs text-muted-foreground">{getRelativeTime(winner.won_at)}</p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Subscription CTA */}
      {isAuthenticated && (
        <Card className="bg-gradient-to-r from-green-50 to-blue-50 border-green-200">
          <CardContent className="p-6 text-center">
            <h3 className="font-semibold mb-2">Never Miss a Draw!</h3>
            <p className="text-muted-foreground mb-4">
              Subscribe for just ₦20/day and get guaranteed entries into every daily draw
            </p>
            <Button className="bg-green-600 hover:bg-green-700" onClick={() => window.location.href = '/subscription'}>
              Subscribe Now
            </Button>
          </CardContent>
        </Card>
      )}
    </div>
  );
};
