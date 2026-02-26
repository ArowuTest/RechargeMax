import React, { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useOTP } from '@/hooks/useOTP';
import { useAuthContext } from '@/contexts/AuthContext';
import { validateNigerianPhone, displayPhoneNumber } from '@/lib/utils';
import { Loader2, Phone, Shield } from 'lucide-react';

interface LoginModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess?: () => void;
}

export const LoginModal: React.FC<LoginModalProps> = ({ isOpen, onClose, onSuccess }) => {
  const [phoneNumber, setPhoneNumber] = useState('');
  const { login } = useAuthContext();
  const {
    phone,
    otp,
    isOTPSent,
    isVerifying,
    isSending,
    sendOTP,
    verifyOTP,
    resetOTP,
    setOTP,
    getRemainingTime,
    canResend
  } = useOTP();

  const handleSendOTP = async () => {
    const success = await sendOTP(phoneNumber);
    if (!success) {
      setPhoneNumber('');
    }
  };

  const handleVerifyOTP = async () => {
    const success = await verifyOTP(otp);
    if (success) {
      try {
        // Fetch real user data from database
        const response = await fetch('getUserDashboard', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            msisdn: phone
          })
        });

        const result = await response.json();
        
        if (result.success && result.data.user) {
          const userData = {
            id: String(result.data.user.id),
            msisdn: String(result.data.user.msisdn),
            full_name: String(result.data.user.full_name || ''),
            email: String(result.data.user.email || ''),
            loyalty_tier: String(result.data.user.loyalty_tier),
            total_points: Number(result.data.user.total_points || 0),
            total_recharges: Number(result.data.user.total_recharges || 0)
          };
          
          login(userData);
          resetOTP();
          setPhoneNumber('');
          onSuccess?.();
          onClose();
        } else {
          throw new Error('Failed to fetch user data');
        }
      } catch (error) {
        console.error('Login error:', error);
        // Fallback: create basic user data
        const userData = {
          id: crypto.randomUUID(),
          msisdn: phone,
          full_name: '',
          email: '',
          loyalty_tier: 'Bronze',
          total_points: 0,
          total_recharges: 0
        };
        
        login(userData);
        resetOTP();
        setPhoneNumber('');
        onSuccess?.();
        onClose();
      }
    }
  };

  const handleClose = () => {
    resetOTP();
    setPhoneNumber('');
    onClose();
  };

  const remainingTime = getRemainingTime();

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="mx-auto w-12 h-12 bg-primary/10 rounded-full flex items-center justify-center mb-4">
            {isOTPSent ? <Shield className="w-6 h-6 text-primary" /> : <Phone className="w-6 h-6 text-primary" />}
          </div>
          <CardTitle>
            {isOTPSent ? 'Verify Phone Number' : 'Login to RechargeMax'}
          </CardTitle>
          <CardDescription>
            {isOTPSent 
              ? `Enter the 6-digit code sent to ${displayPhoneNumber(phone)}`
              : 'Enter your phone number to receive a verification code'
            }
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {!isOTPSent ? (
            <>
              <div className="space-y-2">
                <Label htmlFor="phone">Phone Number</Label>
                <Input
                  id="phone"
                  type="tel"
                  placeholder="0803 123 4567"
                  value={phoneNumber}
                  onChange={(e) => setPhoneNumber(e.target.value)}
                  disabled={isSending}
                />
              </div>
              <div className="flex gap-2">
                <Button 
                  onClick={handleSendOTP}
                  disabled={!validateNigerianPhone(phoneNumber) || isSending}
                  className="flex-1"
                >
                  {isSending && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
                  Send Code
                </Button>
                <Button variant="outline" onClick={handleClose}>
                  Cancel
                </Button>
              </div>
            </>
          ) : (
            <>
              <div className="space-y-2">
                <Label htmlFor="otp">Verification Code</Label>
                <Input
                  id="otp"
                  type="text"
                  placeholder="123456"
                  value={otp}
                  onChange={(e) => setOTP(e.target.value.replace(/\D/g, '').slice(0, 6))}
                  disabled={isVerifying}
                  className="text-center text-lg tracking-widest"
                />
              </div>
              
              {remainingTime > 0 && (
                <p className="text-sm text-muted-foreground text-center">
                  Code expires in {Math.floor(remainingTime / 60)}:{(remainingTime % 60).toString().padStart(2, '0')}
                </p>
              )}

              <div className="flex gap-2">
                <Button 
                  onClick={handleVerifyOTP}
                  disabled={otp.length !== 6 || isVerifying}
                  className="flex-1"
                >
                  {isVerifying && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
                  Verify
                </Button>
                <Button 
                  variant="outline" 
                  onClick={canResend ? () => sendOTP(phone) : undefined}
                  disabled={!canResend || isSending}
                >
                  {isSending ? <Loader2 className="w-4 h-4 animate-spin" /> : 'Resend'}
                </Button>
              </div>
              
              <Button variant="ghost" onClick={handleClose} className="w-full">
                Use Different Number
              </Button>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
};