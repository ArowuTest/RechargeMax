import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { rechargeApi, apiClient } from '@/lib/api-client';
import { SpinWheel } from '@/components/games/SpinWheel';
import { useAuthContext } from '@/contexts/AuthContext';
import { toast } from '@/components/ui/sonner';
import { logError, logPerformance } from '@/lib/api';
import { useAffiliateTracking } from '@/hooks/useAffiliateTracking';
import { 
  Phone, 
  Wifi, 
  CreditCard, 
  Wallet,
  Sparkles,
  CheckCircle,
  AlertTriangle,
  Loader2,
  Gift,
  TrendingUp,
  Shield,
  Users,
  Info
} from 'lucide-react';

interface PremiumRechargeFormProps {
  className?: string;
  onRechargeSuccess?: (result: any) => void;
  showRewards?: boolean;
  compactMode?: boolean;
}

interface FormData {
  phoneNumber: string;
  networkProvider: string;
  rechargeType: 'AIRTIME' | 'DATA';
  amount: number;
  dataBundle?: string;
  paymentMethod: 'CARD' | 'BANK_TRANSFER' | 'WALLET';
  useWallet: boolean;
  customerEmail?: string;
  customerName?: string;
}

interface ServicePricing {
  service_code: string;
  service_name: string;
  selling_price: number;
  profit_margin: number;
}

interface WalletInfo {
  balance: number;
  total_funded: number;
  total_spent: number;
}

interface NetworkValidationResult {
  valid: boolean;
  network: string;
  validation_source: string;
  confidence: string;
  message: string;
  actual_network?: string;
}

interface CachedNetworkResult {
  cached: boolean;
  network?: string;
  last_recharged?: string;
  message?: string;
  confidence?: string;
}

interface DataBundle {
  id: string;
  name: string;
  network: string;
  price: number;
  data_size: string;
  validity: string;
  description: string;
}

// Preset airtime amounts in Naira
const PRESET_AIRTIME_AMOUNTS = [100, 200, 500, 1000, 2000, 5000];

export const PremiumRechargeForm: React.FC<PremiumRechargeFormProps> = ({ 
  className,
  onRechargeSuccess,
  showRewards = true,
  compactMode = false
}) => {
  const { user } = useAuthContext();
  const { getAffiliateCode } = useAffiliateTracking();
  // toast is from sonner — global, works outside React render cycle

  const [formData, setFormData] = useState<FormData>({
    phoneNumber: '',
    networkProvider: '',
    rechargeType: 'AIRTIME',
    amount: 0,
    paymentMethod: 'CARD',
    useWallet: false
  });

  const [isProcessing, setIsProcessing] = useState(false);
  const [isValidating, setIsValidating] = useState(false);
  const [isLoadingDataPlans, setIsLoadingDataPlans] = useState(false);
  const [processingStep, setProcessingStep] = useState('');
  const [processingProgress, setProcessingProgress] = useState(0);
  const [validationErrors, setValidationErrors] = useState<Record<string, string>>({});
  const [rechargeResult, setRechargeResult] = useState<any>(null);
  const [showSpinWheel, setShowSpinWheel] = useState(false);
  
  // Network validation states
  const [networkSuggestion, setNetworkSuggestion] = useState<CachedNetworkResult | null>(null);
  const [networkValidation, setNetworkValidation] = useState<NetworkValidationResult | null>(null);
  const [showNetworkWarning, setShowNetworkWarning] = useState(false);
  
  // Wallet and pricing data
  const [walletInfo, setWalletInfo] = useState<WalletInfo>({ balance: 0, total_funded: 0, total_spent: 0 });
  const [dataBundles, setDataBundles] = useState<DataBundle[]>([]);
  const [selectedPresetAmount, setSelectedPresetAmount] = useState<number | null>(null);

  useEffect(() => {
    if (user) {
      loadWalletInfo();
    }
  }, [user]);

  // Load data plans when network and recharge type change
  useEffect(() => {
    if (formData.networkProvider && formData.rechargeType === 'DATA') {
      loadDataPlans();
    } else {
      setDataBundles([]);
    }
  }, [formData.networkProvider, formData.rechargeType]);

  // Check for cached network when phone number is entered
  useEffect(() => {
    const checkCachedNetwork = async () => {
      if (formData.phoneNumber.length === 11 && /^0[789][01]\d{8}$/.test(formData.phoneNumber)) {
        try {
          const result = await apiClient.post('/networks/cached', { phone_number: formData.phoneNumber });
          if (result.data?.success && result.data.data?.cached) {
            setNetworkSuggestion(result.data.data);
            if (!formData.networkProvider) {
              setFormData(prev => ({ ...prev, networkProvider: result.data.data.network }));
              toast('Network Detected', { description: result.data.data.message, duration: 3000 });
            }
          }
        } catch (error) {
          console.error('Failed to check cached network:', error);
        }
      }
    };

    checkCachedNetwork();
  }, [formData.phoneNumber]);

  // Validate network selection in real-time
  useEffect(() => {
    const validateNetworkSelection = async () => {
      if (formData.phoneNumber.length === 11 && formData.networkProvider) {
        setIsValidating(true);
        try {
          const result = await apiClient.post('/networks/validate-selection', {
            phone_number: formData.phoneNumber,
            selected_network: formData.networkProvider,
          });
          
          if (result.data?.success) {
            // Validation passed
            setNetworkValidation(result.data.data);
            setShowNetworkWarning(false);
            setValidationErrors(prev => {
              const { networkProvider, ...rest } = prev;
              return rest;
            });
          } else {
            // Validation failed - check if it's high-confidence (real HLR mismatch)
            const validationData = result.data.data;
            setNetworkValidation(validationData);
            // Only warn if high-confidence mismatch (not just "prefix unavailable")
            const isHighConfidence = validationData?.confidence !== 'low' && 
              validationData?.validation_source !== 'user_selection';
            setShowNetworkWarning(isHighConfidence);
            if (isHighConfidence) {
              setValidationErrors(prev => ({
                ...prev,
                networkProvider: result.data?.error?.message || 'Network mismatch detected'
              }));
            }
          }
        } catch (error) {
          console.error('Network validation error:', error);
        } finally {
          setIsValidating(false);
        }
      }
    };

    // Debounce validation
    const timer = setTimeout(validateNetworkSelection, 500);
    return () => clearTimeout(timer);
  }, [formData.phoneNumber, formData.networkProvider]);

  const loadWalletInfo = async () => {
    // Wallet functionality disabled - using direct payment only
    setWalletInfo({ balance: 0, total_funded: 0, total_spent: 0 });
  };

  const loadDataPlans = async () => {
    if (!formData.networkProvider) return;
    
    setIsLoadingDataPlans(true);
    try {
      const result = await apiClient.get(`/networks/${formData.networkProvider}/bundles`);
      if (result.data?.success && result.data.data) {
        setDataBundles(result.data.data);
      } else {
        console.error('Failed to load data plans:', result.data?.error);
        toast.error("Error Loading Data Plans", { description: "Could not load data plans. Please try again." });
      }
    } catch (error) {
      console.error('Data plans loading error:', error);
      toast.error("Error Loading Data Plans", { description: "Could not load data plans. Please try again." });
    } finally {
      setIsLoadingDataPlans(false);
    }
  };

  const handlePresetAmountClick = (amount: number) => {
    setSelectedPresetAmount(amount);
    setFormData(prev => ({ ...prev, amount }));
  };

  const handleCustomAmountChange = (value: string) => {
    const amount = parseInt(value) || 0;
    setSelectedPresetAmount(null); // Deselect preset when custom amount is entered
    setFormData(prev => ({ ...prev, amount }));
  };

  const validateForm = (): boolean => {
    const errors: Record<string, string> = {};
    
    if (!formData.phoneNumber) {
      errors.phoneNumber = 'Phone number is required';
    } else if (!/^0[789][01]\d{8}$/.test(formData.phoneNumber.replace(/\s/g, ''))) {
      errors.phoneNumber = 'Invalid Nigerian phone number';
    }

    if (!formData.networkProvider) {
      errors.networkProvider = 'Please select a network provider';
    }

    // Only block if there's a CONFIRMED high-confidence network mismatch
    // Low confidence / user_selection results are always accepted (HLR unavailable)
    const isHighConfidenceMismatch = networkValidation !== null && 
      !networkValidation?.valid && 
      networkValidation?.confidence !== 'low' &&
      networkValidation?.validation_source !== 'user_selection';
    if (isHighConfidenceMismatch) {
      errors.networkProvider = networkValidation?.message || 'Please verify network selection';
    }

    if (formData.rechargeType === 'AIRTIME') {
      if (!formData.amount || formData.amount <= 0) {
        errors.amount = 'Please select or enter an amount';
      } else if (formData.amount % 1 !== 0) {
        errors.amount = 'Amount must be a whole number (no decimals)';
      } else if (formData.amount < 50) {
        errors.amount = 'Minimum amount is ₦50';
      } else if (formData.amount > 50000) {
        errors.amount = 'Maximum amount is ₦50,000';
      }
    }

    if (formData.rechargeType === 'DATA' && !formData.dataBundle) {
      errors.dataBundle = 'Please select a data bundle';
    }

    if (formData.useWallet && walletInfo.balance < formData.amount) {
      errors.wallet = `Insufficient wallet balance. Available: ₦${walletInfo.balance.toLocaleString()}`;
    }

    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const startTime = Date.now();
    
    if (!validateForm()) {
      await logError({
        message: 'Form validation failed',
        component: 'recharge_form',
        severity: 'LOW',
        errors: validationErrors,
        formData: { ...formData, phoneNumber: 'REDACTED' }
      });
      
      toast.error("Validation Error", { description: "Please fix the errors and try again" });
      return;
    }

    setIsProcessing(true);
    setProcessingStep('Initializing...');
    setProcessingProgress(0);

    try {
      // Step 1: Final network validation
      setProcessingStep('Validating network...');
      setProcessingProgress(20);

      // Step 2: Initialize recharge
      setProcessingStep('Initializing recharge...');
      setProcessingProgress(40);

      let response;
      if (formData.rechargeType === 'AIRTIME') {
        response = await rechargeApi.initiateAirtimeRecharge({
          phone_number: formData.phoneNumber,
          network: formData.networkProvider,
          amount: formData.amount // Send in naira (backend will convert to kobo for Paystack)
        });
      } else {
        response = await rechargeApi.initiateDataRecharge({
          phone_number: formData.phoneNumber,
          network: formData.networkProvider,
          bundle_id: formData.dataBundle!
        });
      }

      setProcessingProgress(60);

      if (!response.success) {
        throw new Error(response.error || 'Recharge initiation failed');
      }

      // Step 3: Redirect to payment gateway
      setProcessingStep('Redirecting to secure payment...');
      setProcessingProgress(90);

      // Navigate to Paystack checkout. Page will come back to /?payment=success after payment.
      if (response.data?.payment_url) {
        window.location.href = response.data.payment_url;
        return; // Stop execution — page is navigating away
      }

      setProcessingProgress(100);
      
      toast.success("Success!", { description: "Recharge initiated successfully" });

      if (onRechargeSuccess) {
        onRechargeSuccess(response.data);
      }

    } catch (error: any) {
      console.error('Recharge error:', error);
      const duration = Date.now() - startTime;
      
      await logError({
        error,
        component: 'recharge_form',
        severity: 'HIGH',
        formData: { ...formData, phoneNumber: 'REDACTED' },
        processingStep,
        duration,
        userType: user ? 'authenticated' : 'guest',
        networkValidation: networkValidation
      });
      
      toast.error("Recharge Failed", { description: error.message || 'Please try again later' });
    } finally {
      const duration = Date.now() - startTime;
      
      await logPerformance({
        endpoint: '/api/recharge/initialize',
        method: 'POST',
        duration,
        status: isProcessing ? 200 : 500,
        metadata: {
          rechargeType: formData.rechargeType,
          networkProvider: formData.networkProvider,
          amount: formData.amount,
          validationSource: networkValidation?.validation_source
        }
      });
      
      setIsProcessing(false);
      setProcessingStep('');
      setProcessingProgress(0);
    }
  };

  return (
    <div className={className}>
      <Card className="w-full max-w-2xl mx-auto">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Sparkles className="w-6 h-6 text-primary" />
            Quick Recharge
          </CardTitle>
          <CardDescription>
            Recharge airtime or data and earn rewards
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Phone Number Input */}
            <div className="space-y-2">
              <Label htmlFor="phoneNumber">
                <Phone className="w-4 h-4 inline mr-2" />
                Phone Number
              </Label>
              <Input
                id="phoneNumber"
                type="tel"
                placeholder="0801234567"
                value={formData.phoneNumber}
                onChange={(e) => setFormData({ ...formData, phoneNumber: e.target.value })}
                className={validationErrors.phoneNumber ? 'border-red-500' : ''}
              />
              {validationErrors.phoneNumber && (
                <p className="text-sm text-red-500">{validationErrors.phoneNumber}</p>
              )}
              
              {/* Network Suggestion */}
              {networkSuggestion?.cached && (
                <Alert className="mt-2">
                  <Info className="h-4 w-4" />
                  <AlertDescription>
                    {networkSuggestion.message}
                  </AlertDescription>
                </Alert>
              )}
            </div>

            {/* Network Provider Select */}
            <div className="space-y-2">
              <Label htmlFor="networkProvider">
                <Wifi className="w-4 h-4 inline mr-2" />
                Network Provider
                {isValidating && (
                  <Loader2 className="w-4 h-4 inline ml-2 animate-spin" />
                )}
              </Label>
              <Select
                value={formData.networkProvider}
                onValueChange={(value) => setFormData({ ...formData, networkProvider: value })}
              >
                <SelectTrigger className={validationErrors.networkProvider ? 'border-red-500' : ''}>
                  <SelectValue placeholder="Select network" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="MTN">MTN</SelectItem>
                  <SelectItem value="GLO">GLO</SelectItem>
                  <SelectItem value="AIRTEL">Airtel</SelectItem>
                  <SelectItem value="9MOBILE">9mobile</SelectItem>
                </SelectContent>
              </Select>
              
              {/* Network Validation Feedback */}
              {networkValidation && !isValidating && (
                <div className="mt-2">
                  {networkValidation.valid ? (
                    networkValidation.confidence === 'low' || networkValidation.validation_source === 'user_selection' ? (
                      // Low confidence = HLR unavailable, accepted based on user selection
                      <Alert className="border-blue-300 bg-blue-50">
                        <CheckCircle className="h-4 w-4 text-blue-500" />
                        <AlertDescription className="text-blue-700">
                          Network accepted: {networkValidation.network || formData.networkProvider}
                          <Badge variant="outline" className="ml-2 text-xs text-blue-600">User Selected</Badge>
                        </AlertDescription>
                      </Alert>
                    ) : (
                      <Alert className="border-green-500 bg-green-50">
                        <CheckCircle className="h-4 w-4 text-green-600" />
                        <AlertDescription className="text-green-700">
                          {networkValidation.message}
                          {networkValidation.validation_source === 'hlr_api' && (
                            <Badge variant="outline" className="ml-2 text-xs">
                              Verified
                            </Badge>
                          )}
                        </AlertDescription>
                      </Alert>
                    )
                  ) : (
                    networkValidation.confidence === 'low' || networkValidation.validation_source === 'user_selection' ? (
                      // Low confidence failure = still accept, just inform
                      <Alert className="border-blue-300 bg-blue-50">
                        <CheckCircle className="h-4 w-4 text-blue-500" />
                        <AlertDescription className="text-blue-700">
                          Network accepted: {formData.networkProvider}
                          <Badge variant="outline" className="ml-2 text-xs text-blue-600">User Selected</Badge>
                        </AlertDescription>
                      </Alert>
                    ) : (
                      <Alert variant="destructive">
                        <AlertTriangle className="h-4 w-4" />
                        <AlertDescription>
                          {networkValidation.message}
                          {networkValidation.actual_network && (
                            <div className="mt-2">
                              <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                onClick={() => setFormData({ ...formData, networkProvider: networkValidation.actual_network! })}
                              >
                                Switch to {networkValidation.actual_network}
                              </Button>
                            </div>
                          )}
                        </AlertDescription>
                      </Alert>
                    )
                  )}
                </div>
              )}
              
              {validationErrors.networkProvider && !networkValidation && (
                <p className="text-sm text-red-500">{validationErrors.networkProvider}</p>
              )}
            </div>

            {/* Recharge Type */}
            <div className="space-y-2">
              <Label>Recharge Type</Label>
              <div className="flex gap-4">
                <Button
                  type="button"
                  variant={formData.rechargeType === 'AIRTIME' ? 'default' : 'outline'}
                  onClick={() => setFormData({ ...formData, rechargeType: 'AIRTIME', dataBundle: undefined })}
                  className="flex-1"
                >
                  <Phone className="w-4 h-4 mr-2" />
                  Airtime
                </Button>
                <Button
                  type="button"
                  variant={formData.rechargeType === 'DATA' ? 'default' : 'outline'}
                  onClick={() => setFormData({ ...formData, rechargeType: 'DATA', amount: 0 })}
                  className="flex-1"
                >
                  <Wifi className="w-4 h-4 mr-2" />
                  Data
                </Button>
              </div>
            </div>

            {/* Airtime Amount Selection */}
            {formData.rechargeType === 'AIRTIME' && (
              <div className="space-y-3">
                <Label>Select Amount</Label>
                <div className="grid grid-cols-3 gap-2">
                  {PRESET_AIRTIME_AMOUNTS.map((amount) => (
                    <Button
                      key={amount}
                      type="button"
                      variant={selectedPresetAmount === amount ? 'default' : 'outline'}
                      onClick={() => handlePresetAmountClick(amount)}
                      className="w-full"
                    >
                      ₦{amount.toLocaleString()}
                    </Button>
                  ))}
                </div>
                
                <div className="space-y-2">
                  <Label htmlFor="customAmount">Or Enter Custom Amount (₦)</Label>
                  <Input
                    id="customAmount"
                    type="number"
                    placeholder="Enter amount"
                    value={formData.amount || ''}
                    onChange={(e) => handleCustomAmountChange(e.target.value)}
                    className={validationErrors.amount ? 'border-red-500' : ''}
                  />
                  {validationErrors.amount && (
                    <p className="text-sm text-red-500">{validationErrors.amount}</p>
                  )}
                </div>
              </div>
            )}

            {/* Data Bundle Selection */}
            {formData.rechargeType === 'DATA' && (
              <div className="space-y-2">
                <Label htmlFor="dataBundle">
                  <Wifi className="w-4 h-4 inline mr-2" />
                  Select Data Plan
                  {isLoadingDataPlans && (
                    <Loader2 className="w-4 h-4 inline ml-2 animate-spin" />
                  )}
                </Label>
                <Select
                  value={formData.dataBundle}
                  onValueChange={(value) => {
                    const selectedBundle = dataBundles.find(b => b.id === value);
                    setFormData({ 
                      ...formData, 
                      dataBundle: value,
                      amount: selectedBundle ? selectedBundle.price : 0
                    });
                  }}
                  disabled={!formData.networkProvider || isLoadingDataPlans}
                >
                  <SelectTrigger className={validationErrors.dataBundle ? 'border-red-500' : ''}>
                    <SelectValue placeholder={
                      !formData.networkProvider 
                        ? "Select network first" 
                        : isLoadingDataPlans 
                        ? "Loading plans..." 
                        : "Select data plan"
                    } />
                  </SelectTrigger>
                  <SelectContent>
                    {dataBundles.map((bundle) => (
                      <SelectItem key={bundle.id} value={bundle.id}>
                        {bundle.name} - {bundle.data_size} - ₦{bundle.price.toLocaleString()}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                {validationErrors.dataBundle && (
                  <p className="text-sm text-red-500">{validationErrors.dataBundle}</p>
                )}
                
                {/* Show selected bundle details */}
                {formData.dataBundle && (
                  <div className="mt-2 p-3 bg-blue-50 rounded-md">
                    <p className="text-sm text-blue-900">
                      {dataBundles.find(b => b.id === formData.dataBundle)?.description}
                    </p>
                  </div>
                )}
              </div>
            )}

            {/* Submit Button */}
            <Button
              type="submit"
              className="w-full"
              disabled={isProcessing || isValidating || showNetworkWarning}
            >
              {isProcessing ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  {processingStep || 'Processing...'}
                </>
              ) : (
                <>
                  <Shield className="w-4 h-4 mr-2" />
                  Proceed to Payment {formData.amount > 0 && `₦${formData.amount.toLocaleString()}`}
                </>
              )}
            </Button>

            {/* Processing Progress */}
            {isProcessing && (
              <div className="space-y-2">
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div
                    className="bg-primary h-2 rounded-full transition-all duration-300"
                    style={{ width: `${processingProgress}%` }}
                  />
                </div>
                <p className="text-sm text-center text-gray-600">{processingStep}</p>
              </div>
            )}
          </form>
        </CardContent>
      </Card>
    </div>
  );
};
