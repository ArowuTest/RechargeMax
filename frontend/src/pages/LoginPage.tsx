import React, { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useOTP } from '@/hooks/useOTP';
import { useAuthContext } from '@/contexts/AuthContext';
import { validateNigerianPhone, displayPhoneNumber } from '@/lib/utils';
import { Loader2, Phone, Shield, ArrowLeft, Zap } from 'lucide-react';

export const LoginPage: React.FC = () => {
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
      // Auth cookie set by backend on verifyOTP - get user profile
      const storedUser = localStorage.getItem('rechargemax_user');
      if (storedUser) {
        const userData = JSON.parse(storedUser);
        login(userData);
        resetOTP();
        setPhoneNumber('');
        
        // Redirect to dashboard
        window.location.href = '/dashboard';
      } else {
        // This shouldn't happen, but handle it gracefully
        console.error('User data or token not found after OTP verification');
        // Create fallback user data
        const fallbackUser = {
          id: crypto.randomUUID(),
          msisdn: phone,
          full_name: '',
          email: '',
          loyalty_tier: 'Bronze',
          total_points: 0,
          total_recharges: 0
        };
        
        login(fallbackUser);
        resetOTP();
        setPhoneNumber('');
        
        // Redirect to dashboard
        window.location.href = '/dashboard';
      }
    }
  };

  const handleGoBack = () => {
    resetOTP();
    setPhoneNumber('');
  };

  const remainingTime = getRemainingTime();

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-purple-50 flex items-center justify-center p-4">
      <div className="w-full max-w-md space-y-6">
        {/* Header */}
        <div className="text-center space-y-4">
          <div className="flex items-center justify-center gap-2">
            <div className="p-3 bg-blue-600 rounded-full">
              <Zap className="w-8 h-8 text-white" />
            </div>
            <h1 className="text-3xl font-bold text-gray-900">
              Recharge<span className="text-blue-600">Max</span>
            </h1>
          </div>
          <p className="text-gray-600">
            Login to access your rewards dashboard
          </p>
        </div>

        {/* Login Card */}
        <Card className="w-full">
          <CardHeader className="text-center">
            <div className="mx-auto w-12 h-12 bg-primary/10 rounded-full flex items-center justify-center mb-4">
              {isOTPSent ? <Shield className="w-6 h-6 text-primary" /> : <Phone className="w-6 h-6 text-primary" />}
            </div>
            <CardTitle>
              {isOTPSent ? 'Verify Phone Number' : 'Login with Phone Number'}
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
                    Send Verification Code
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
                    Verify & Login
                  </Button>
                  <Button 
                    variant="outline" 
                    onClick={canResend ? () => sendOTP(phone) : undefined}
                    disabled={!canResend || isSending}
                  >
                    {isSending ? <Loader2 className="w-4 h-4 animate-spin" /> : 'Resend'}
                  </Button>
                </div>
                
                <Button variant="ghost" onClick={handleGoBack} className="w-full">
                  <ArrowLeft className="w-4 h-4 mr-2" />
                  Use Different Number
                </Button>
              </>
            )}
          </CardContent>
        </Card>

        {/* Back to Home */}
        <div className="text-center">
          <Button 
            variant="ghost" 
            onClick={() => window.location.href = '/'}
            className="text-gray-600 hover:text-gray-900"
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Home
          </Button>
        </div>

        {/* Features Preview */}
        <Card className="bg-gradient-to-r from-blue-50 to-purple-50 border-blue-200">
          <CardContent className="p-6">
            <h3 className="font-semibold mb-3 text-center">What you get with RechargeMax:</h3>
            <div className="space-y-2 text-sm text-gray-600">
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 bg-blue-600 rounded-full"></div>
                <span>Instant airtime and data recharge</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 bg-purple-600 rounded-full"></div>
                <span>Daily draws with cash prizes up to ₦500,000</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 bg-yellow-600 rounded-full"></div>
                <span>Spin wheel rewards on ₦1,000+ recharges</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 bg-green-600 rounded-full"></div>
                <span>Loyalty tiers with exclusive benefits</span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default LoginPage;