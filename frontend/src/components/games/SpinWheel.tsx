import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { formatCurrency } from '@/lib/utils';
import { useToast } from '@/hooks/useToast';
import { Gift, Zap, RotateCcw, Loader2 } from 'lucide-react';
import apiClient from '@/lib/api-client';
import { useAuthContext } from '@/contexts/AuthContext';

// Fallback prizes used only when /spin/prizes is unreachable
const FALLBACK_PRIZES = [
  { name: '₦100 Airtime',   type: 'AIRTIME', value: 10000,  probability: 25, color: '#10b981' },
  { name: '₦200 Airtime',   type: 'AIRTIME', value: 20000,  probability: 20, color: '#3b82f6' },
  { name: '500MB Data',     type: 'DATA',    value: 50000,  probability: 15, color: '#8b5cf6' },
  { name: '1GB Data',       type: 'DATA',    value: 100000, probability: 15, color: '#f59e0b' },
  { name: '₦100 Cash',      type: 'CASH',    value: 10000,  probability: 10, color: '#ef4444' },
  { name: '₦200 Cash',      type: 'CASH',    value: 20000,  probability: 8,  color: '#ec4899' },
  { name: '₦500 Cash',      type: 'CASH',    value: 50000,  probability: 5,  color: '#fbbf24' },
  { name: '₦1000 Cash',     type: 'CASH',    value: 100000, probability: 2,  color: '#6b7280' },
];

interface WheelPrize {
  name: string;
  type: string;
  value: number;
  probability: number;
  color: string;
}

interface SpinWheelProps {
  isOpen: boolean;
  onClose: () => void;
  transactionAmount: number;
  userPhone: string;
  onPrizeWon?: (prize: any) => void;
}

export const SpinWheel: React.FC<SpinWheelProps> = ({
  isOpen,
  onClose,
  transactionAmount,
  userPhone,
  onPrizeWon,
}) => {
  const { toast } = useToast();
  const { isAuthenticated } = useAuthContext();
  const [prizes, setPrizes] = useState<WheelPrize[]>(FALLBACK_PRIZES);
  const [loadingPrizes, setLoadingPrizes] = useState(true);
  const [isSpinning, setIsSpinning] = useState(false);
  const [rotation, setRotation] = useState(0);
  const [selectedPrize, setSelectedPrize] = useState<any>(null);
  const [hasSpun, setHasSpun] = useState(false);

  // Fetch live prizes from backend when the wheel opens
  useEffect(() => {
    if (!isOpen) return;
    setLoadingPrizes(true);
    apiClient
      .get('/spin/prizes')
      .then((res) => {
        const raw: any[] = res.data?.data ?? [];
        if (raw.length > 0) {
          const mapped: WheelPrize[] = raw
            .filter((p) => p.is_active !== false)
            .map((p) => ({
              name:        p.prize_name ?? p.name ?? 'Prize',
              type:        (p.prize_type ?? p.type ?? 'AIRTIME').toUpperCase(),
              value:       Number(p.prize_value ?? p.value ?? 0),
              probability: Number(p.probability ?? 0),
              color:       p.color_scheme ?? p.color ?? '#6b7280',
            }));
          if (mapped.length > 0) setPrizes(mapped);
        }
      })
      .catch(() => {
        // Silently fall back to FALLBACK_PRIZES; wheel still works
      })
      .finally(() => setLoadingPrizes(false));
  }, [isOpen]);

  const segmentAngle = 360 / prizes.length;

  const spinWheel = async () => {
    if (isSpinning || hasSpun) return;
    setIsSpinning(true);

    try {
      // SECURITY: prize is ALWAYS determined server-side.
      // When the user is logged in, the JWT (sent automatically via the
      // Authorization header in api-client.ts) identifies them — no MSISDN
      // needed in the body.
      // When the user is a guest (not logged in), we send the MSISDN from the
      // recharge form so the backend can validate a qualifying transaction
      // exists within the last 4 hours for that number.
      const spinBody = isAuthenticated ? {} : { msisdn: userPhone };
      const response = await apiClient.post('/spin/play', spinBody);

      if (!response.data.success) {
        throw new Error(response.data.error || 'Failed to spin');
      }

      const spinResult = response.data.data;

      // Match backend result to a wheel segment — fall back to first prize if nothing matches
      const winningPrize: WheelPrize =
        prizes.find(
          (p) => p.type === spinResult.prize_type && p.value === spinResult.prize_value,
        ) ??
        prizes.find((p) => p.name === spinResult.prize_won) ??
        prizes[0] ?? { name: spinResult.prize_won ?? 'Prize', type: spinResult.prize_type ?? 'AIRTIME', value: 0, probability: 0, color: '#6b7280' };

      // Animate to the winning segment
      const prizeIndex = prizes.findIndex((p) => p.name === winningPrize.name);
      const targetAngle = prizeIndex * segmentAngle + segmentAngle / 2;
      const spins = 5 + Math.random() * 3;
      const finalRotation = spins * 360 + (360 - targetAngle);
      setRotation((prev) => prev + finalRotation);

      setTimeout(() => {
        setIsSpinning(false);
        setSelectedPrize({ ...winningPrize, claimStatus: spinResult.claim_status });
        setHasSpun(true);

        const isProvisioning = spinResult.claim_status === 'PROVISIONING';
        const claimInstructions =
          winningPrize.type === 'AIRTIME' || winningPrize.type === 'DATA'
            ? isProvisioning
              ? 'Your prize is being processed — it will be credited within 5-10 minutes.'
              : 'Login and check Dashboard → My Prizes for status.'
            : winningPrize.type === 'CASH'
            ? 'Login, then go to Dashboard → Prize Claims to submit your bank details.'
            : 'Login to see your updated account.';

        toast({
          title: '🎉 Congratulations! You Won!',
          description: `${winningPrize.name}! ${claimInstructions}`,
          duration: 10000,
        });

        onPrizeWon?.(winningPrize);
      }, 4000);
    } catch (error: any) {
      setIsSpinning(false);
      // error.response.data.error is an object {code, message} — extract .message to avoid React #31 crash
      const errMsg: string =
        error.response?.data?.error?.message ??
        error.response?.data?.message ??
        error.message ??
        'Failed to spin the wheel. Please try again.';
      toast({
        title: 'Spin Failed',
        description: errMsg,
        variant: 'destructive',
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
          {loadingPrizes ? (
            <div className="flex justify-center items-center h-80">
              <Loader2 className="w-10 h-10 animate-spin text-primary" />
            </div>
          ) : (
            <div className="relative mx-auto w-80 h-80">
              {/* Spinning disc */}
              <div
                className="w-full h-full rounded-full border-4 border-gray-300 relative overflow-hidden transition-transform duration-[4000ms] ease-out"
                style={{
                  transform: `rotate(${rotation}deg)`,
                  background: `conic-gradient(${prizes
                    .map(
                      (prize, i) =>
                        `${prize.color} ${i * segmentAngle}deg ${(i + 1) * segmentAngle}deg`,
                    )
                    .join(', ')})`,
                }}
              >
                {prizes.map((prize, index) => {
                  const angle = index * segmentAngle + segmentAngle / 2;
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
                        textShadow: '1px 1px 2px rgba(0,0,0,0.8)',
                      }}
                    >
                      {prize.name.split(' ').map((word, i) => (
                        <div key={i}>{word}</div>
                      ))}
                    </div>
                  );
                })}
              </div>

              {/* Centre spin button */}
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
                <div className="w-0 h-0 border-l-4 border-r-4 border-b-8 border-l-transparent border-r-transparent border-b-red-500" />
              </div>
            </div>
          )}

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

              <div className="bg-gradient-to-r from-blue-50 to-green-50 border-2 border-blue-300 p-6 rounded-xl text-left shadow-lg">
                <h4 className="font-bold text-lg text-blue-900 mb-4 flex items-center gap-2">
                  🎁{' '}
                  <span className="bg-yellow-200 px-2 py-1 rounded">
                    IMPORTANT: How to Claim Your Prize
                  </span>
                </h4>
                {(selectedPrize.type === 'AIRTIME' || selectedPrize.type === 'DATA') && (
                  <div className="text-sm text-blue-700 space-y-1">
                    <p>
                      {selectedPrize.type === 'AIRTIME' ? '📱' : '📶'}{' '}
                      <strong>{selectedPrize.type === 'AIRTIME' ? 'Airtime' : 'Data'} Prize:</strong>{' '}
                      {selectedPrize.name}
                    </p>
                    {selectedPrize.claimStatus === 'PROVISIONING' ? (
                      <>
                        <p>⏳ Your prize is <strong>being processed</strong></p>
                        <p>1. It will be credited within <strong>5-10 minutes</strong></p>
                        <p>2. <strong>Login</strong> and check <strong>Dashboard → My Prizes</strong> for status</p>
                      </>
                    ) : (
                      <>
                        <p>1. <strong>Login</strong> with your phone number</p>
                        <p>2. Prize will be <strong>automatically credited</strong> to your phone</p>
                        <p>3. Credited within <strong>5-10 minutes</strong></p>
                      </>
                    )}
                  </div>
                )}
                {selectedPrize.type === 'CASH' && (
                  <div className="text-sm text-blue-700 space-y-1">
                    <p>💰 <strong>Cash Prize:</strong> {selectedPrize.name}</p>
                    <p>1. <strong>Login</strong> with your phone number</p>
                    <p>2. Go to <strong>Dashboard → Prize Claims</strong></p>
                    <p>3. Complete the <strong>Bank Details Form</strong></p>
                    <p>4. Cash transferred within <strong>24-48 hours</strong></p>
                  </div>
                )}
                {selectedPrize.type === 'DRAW_TICKETS' && (
                  <div className="text-sm text-blue-700 space-y-1">
                    <p>🎫 <strong>Draw Entries:</strong> {selectedPrize.value} extra entries</p>
                    <p>1. Entries <strong>automatically added</strong> to your account</p>
                    <p>2. <strong>Login</strong> to see updated entry count</p>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Action Buttons */}
          <div className="flex gap-3">
            {!hasSpun ? (
              <>
                <Button onClick={spinWheel} disabled={isSpinning || loadingPrizes} className="flex-1">
                  {isSpinning ? 'Spinning...' : 'Spin Now!'}
                </Button>
                <Button variant="outline" onClick={handleClose}>
                  Skip
                </Button>
              </>
            ) : (
              <div className="space-y-3 w-full">
                <Button
                  onClick={() => (window.location.href = '/login')}
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

          {/* Prizes list */}
          <div className="border-t pt-4">
            <h4 className="font-semibold mb-3 text-center">Possible Prizes</h4>
            <div className="grid grid-cols-2 gap-2 text-sm">
              {prizes.map((prize) => (
                <div key={prize.name} className="flex items-center gap-2">
                  <div className="w-3 h-3 rounded-full" style={{ backgroundColor: prize.color }} />
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
