import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { SPIN_PRIZES } from '@/lib/constants';
import { formatCurrency } from '@/lib/utils';
import { useToast } from '@/hooks/use-toast';
import { Gift, Zap, RotateCcw } from 'lucide-react';
import apiClient from '@/lib/api-client';

interface SpinWheelProps {
  isOpen: boolean;
  onClose: () => void;
  transactionAmount: number;
  userPhone: string; // User's phone number for guest spin
  onPrizeWon?: (prize: any) => void;
}

export const SpinWheel: React.FC<SpinWheelProps> = ({ 
  isOpen, 
  onClose, 
  transactionAmount,
  userPhone,
  onPrizeWon 
}) => {
  const { toast } = useToast();
  const [isSpinning, setIsSpinning] = useState(false);
  const [rotation, setRotation] = useState(0);
  const [selectedPrize, setSelectedPrize] = useState<any>(null);
  const [hasSpun, setHasSpun] = useState(false);

  // Calculate segment angles
  const segmentAngle = 360 / SPIN_PRIZES.length;
  
  const spinWheel = async () => {
    if (isSpinning || hasSpun) return;
    
    setIsSpinning(true);
    
    try {
      // Call backend API to play spin - SECURITY: Prize determined server-side
      const response = await apiClient.post('/spin/play', {
        msisdn: userPhone
      });
      
      if (!response.data.success) {
        throw new Error(response.data.error || 'Failed to spin');
      }
      
      const spinResult = response.data.data;
      
      // Find the matching prize from SPIN_PRIZES based on backend response
      const winningPrize = SPIN_PRIZES.find(p => 
        p.type === spinResult.prize_type && 
        p.value === spinResult.prize_value
      ) || SPIN_PRIZES.find(p => p.name === spinResult.prize_won);
      
      if (!winningPrize) {
        console.error('Prize not found in SPIN_PRIZES:', spinResult);
        throw new Error('Prize configuration error');
      }
      
      // Calculate rotation to land on winning prize
      const prizeIndex = SPIN_PRIZES.findIndex(p => p.name === winningPrize.name);
      const targetAngle = (prizeIndex * segmentAngle) + (segmentAngle / 2);
      const spins = 5 + Math.random() * 3; // 5-8 full rotations for visual effect
      const finalRotation = (spins * 360) + (360 - targetAngle);
      
      setRotation(prev => prev + finalRotation);
      
      // Show result after animation
      setTimeout(() => {
        setIsSpinning(false);
        setSelectedPrize({ ...winningPrize, claimStatus: spinResult.claim_status });
        setHasSpun(true);
        
        // Show enhanced toast notification
        // PERF-002: Backend returns PROVISIONING for async airtime/data prizes.
        // Show "being processed" copy instead of "immediately credited" to set correct expectations.
        const isProvisioning = spinResult.claim_status === 'PROVISIONING';
        const claimInstructions = winningPrize.type === 'AIRTIME' || winningPrize.type === 'DATA'
          ? isProvisioning
            ? 'Your prize is being processed — it will be credited to your phone within 5-10 minutes. Check Dashboard for status updates.'
            : 'Login with your phone number to claim. Prize will be automatically credited within 5-10 minutes.'
          : winningPrize.type === 'CASH'
          ? 'Login with your phone number, then go to Dashboard → Prize Claims to complete bank details form.'
          : 'Login to see your updated account.';
        
        toast({
          title: "🎉 Congratulations! You Won!",
          description: `${winningPrize.name}! ${claimInstructions}`,
          duration: 10000, // Show for 10 seconds
        });
        
        // Call the prize won callback
        onPrizeWon?.(winningPrize);
      }, 4000);
      
    } catch (error: any) {
      console.error('Spin error:', error);
      setIsSpinning(false);
      
      toast({
        title: "Spin Failed",
        description: error.response?.data?.error || error.message || 'Failed to spin the wheel. Please try again.',
        variant: "destructive",
        duration: 5000,
      });
    }
  };

  const handleClose = () => {
    setRotation(0);
    setSelectedPrize(null);
    setHasSpun(false);
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4">
      <Card className="w-full max-w-lg">
        <CardHeader className="text-center">
          <CardTitle className="flex items-center justify-center gap-2">
            <Zap className="w-6 h-6 text-yellow-500" />
            Spin the Wheel!
          </CardTitle>
          <CardDescription>
            You've unlocked a free spin for recharging {formatCurrency(transactionAmount)}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Wheel Container */}
          <div className="relative mx-auto w-80 h-80">
            {/* Wheel */}
            <div 
              className="w-full h-full rounded-full border-4 border-gray-300 relative overflow-hidden transition-transform duration-4000 ease-out"
              style={{ 
                transform: `rotate(${rotation}deg)`,
                background: `conic-gradient(${SPIN_PRIZES.map((prize, index) => 
                  `${prize.color} ${index * segmentAngle}deg ${(index + 1) * segmentAngle}deg`
                ).join(', ')})`
              }}
            >
              {/* Prize Labels */}
              {SPIN_PRIZES.map((prize, index) => {
                const angle = (index * segmentAngle) + (segmentAngle / 2);
                const radian = (angle * Math.PI) / 180;
                const x = Math.cos(radian) * 120;
                const y = Math.sin(radian) * 120;
                
                return (
                  <div
                    key={prize.name}
                    className="absolute text-white text-xs font-bold text-center"
                    style={{
                      left: `calc(50% + ${x}px - 30px)`,
                      top: `calc(50% + ${y}px - 10px)`,
                      width: '60px',
                      transform: `rotate(${angle}deg)`,
                      textShadow: '1px 1px 2px rgba(0,0,0,0.8)'
                    }}
                  >
                    {prize.name.split(' ').map((word, i) => (
                      <div key={i}>{word}</div>
                    ))}
                  </div>
                );
              })}
            </div>
            
            {/* Center Spin Button */}
            <div className="absolute inset-0 flex items-center justify-center">
              <Button
                onClick={spinWheel}
                disabled={isSpinning || hasSpun}
                className="w-20 h-20 rounded-full bg-white text-primary border-4 border-primary hover:bg-gray-50 disabled:opacity-50"
                size="lg"
              >
                {isSpinning ? (
                  <RotateCcw className="w-8 h-8 animate-spin" />
                ) : (
                  <span className="font-bold text-lg">SPIN</span>
                )}
              </Button>
            </div>
            
            {/* Pointer */}
            <div className="absolute top-0 left-1/2 transform -translate-x-1/2 -translate-y-2">
              <div className="w-0 h-0 border-l-4 border-r-4 border-b-8 border-l-transparent border-r-transparent border-b-red-500"></div>
            </div>
          </div>

          {/* Prize Result */}
          {selectedPrize && (
            <div className="text-center space-y-4">
              <div className="bg-gradient-to-r from-yellow-400 to-orange-500 text-white p-4 rounded-lg">
                <h3 className="text-xl font-bold flex items-center justify-center gap-2">
                  <Gift className="w-6 h-6" />
                  🎉 Congratulations! You Won!
                </h3>
                <p className="text-2xl font-bold mt-2">{selectedPrize.name}</p>
                <Badge variant="secondary" className="mt-2">
                  {selectedPrize.type === 'AIRTIME' && 'Airtime Prize'}
                  {selectedPrize.type === 'DATA' && 'Data Prize'}
                  {selectedPrize.type === 'CASH' && 'Cash Prize'}
                  {selectedPrize.type === 'DRAW_TICKETS' && 'Extra Draw Entries'}
                </Badge>
              </div>
              
              {/* Prize Claiming Instructions - Enhanced */}
              <div className="bg-gradient-to-r from-blue-50 to-green-50 border-2 border-blue-300 p-6 rounded-xl text-left shadow-lg animate-pulse">
                <h4 className="font-bold text-lg text-blue-900 mb-4 flex items-center gap-2">
                  🎁 <span className="bg-yellow-200 px-2 py-1 rounded">IMPORTANT: How to Claim Your Prize</span>
                </h4>
                {selectedPrize.type === 'CASH' && (
                  <div className="text-sm text-blue-700 space-y-1">
                    <p>💰 <strong>Cash Prize:</strong> {selectedPrize.name}</p>
                    <p>1. <strong>Login</strong> with your phone number (MSISDN)</p>
                    <p>2. Go to your <strong>Dashboard</strong> → <strong>Prize Claims</strong></p>
                    <p>3. Complete the <strong>Bank Details Form</strong></p>
                    <p>4. Cash will be transferred within <strong>24-48 hours</strong></p>
                  </div>
                )}
                {selectedPrize.type === 'AIRTIME' && (
                  <div className="text-sm text-blue-700 space-y-1">
                    <p>📱 <strong>Airtime Prize:</strong> {selectedPrize.name}</p>
                    {selectedPrize.claimStatus === 'PROVISIONING' ? (
                      <>
                        <p>⏳ Your prize is <strong>being processed</strong></p>
                        <p>1. Airtime will be credited within <strong>5-10 minutes</strong></p>
                        <p>2. <strong>Login</strong> and check <strong>Dashboard → My Prizes</strong> for status</p>
                      </>
                    ) : (
                      <>
                        <p>1. <strong>Login</strong> with your phone number (MSISDN)</p>
                        <p>2. Prize will be <strong>automatically credited</strong> to your phone</p>
                        <p>3. Check your <strong>Dashboard</strong> for claim status</p>
                        <p>4. Airtime credited within <strong>5-10 minutes</strong></p>
                      </>
                    )}
                  </div>
                )}
                {selectedPrize.type === 'DATA' && (
                  <div className="text-sm text-blue-700 space-y-1">
                    <p>📶 <strong>Data Prize:</strong> {selectedPrize.name}</p>
                    {selectedPrize.claimStatus === 'PROVISIONING' ? (
                      <>
                        <p>⏳ Your prize is <strong>being processed</strong></p>
                        <p>1. Data will be credited within <strong>5-10 minutes</strong></p>
                        <p>2. <strong>Login</strong> and check <strong>Dashboard → My Prizes</strong> for status</p>
                      </>
                    ) : (
                      <>
                        <p>1. <strong>Login</strong> with your phone number (MSISDN)</p>
                        <p>2. Data will be <strong>automatically credited</strong> to your phone</p>
                        <p>3. Check your <strong>Dashboard</strong> for claim status</p>
                        <p>4. Data credited within <strong>5-10 minutes</strong></p>
                      </>
                    )}
                  </div>
                )}
                {selectedPrize.type === 'DRAW_TICKETS' && (
                  <div className="text-sm text-blue-700 space-y-1">
                    <p>🎫 <strong>Draw Entries:</strong> {selectedPrize.value} extra entries</p>
                    <p>1. Entries <strong>automatically added</strong> to your account</p>
                    <p>2. <strong>Login</strong> to see updated entry count</p>
                    <p>3. Check <strong>Daily Draws</strong> for upcoming draws</p>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Action Buttons */}
          <div className="flex gap-3">
            {!hasSpun ? (
              <>
                <Button onClick={spinWheel} disabled={isSpinning} className="flex-1">
                  {isSpinning ? 'Spinning...' : 'Spin Now!'}
                </Button>
                <Button variant="outline" onClick={handleClose}>
                  Skip
                </Button>
              </>
            ) : (
              <div className="space-y-3 w-full">
                {/* Prominent Login/Claim Button */}
                <Button 
                  onClick={() => {
                    // Redirect to login page or show login modal
                    window.location.href = '/login';
                  }}
                  className="w-full bg-green-600 hover:bg-green-700 text-white font-bold py-3"
                >
                  📱 Login to Claim Your Prize
                </Button>
                <Button onClick={handleClose} variant="outline" className="w-full">
                  Close (Claim Later)
                </Button>
              </div>
            )}
          </div>

          {/* Prizes List */}
          <div className="border-t pt-4">
            <h4 className="font-semibold mb-3 text-center">Possible Prizes</h4>
            <div className="grid grid-cols-2 gap-2 text-sm">
              {SPIN_PRIZES.map((prize) => (
                <div key={prize.name} className="flex items-center gap-2">
                  <div 
                    className="w-3 h-3 rounded-full" 
                    style={{ backgroundColor: prize.color }}
                  />
                  <span>{prize.name}</span>
                </div>
              ))}
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};
