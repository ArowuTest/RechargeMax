/**
 * Enterprise React Hook - Enhanced Recharge Management
 * Implements Clean Architecture principles with comprehensive error handling
 */

import { useState, useCallback, useRef, useEffect } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

import { 
  RechargeRequest, 
  RechargeResult, 
  NetworkProvider, 
  RechargeType, 
  PaymentMethod,
  Result,
  PhoneNumber,
  Money,
  TransactionStatus
} from '@/domain/types';

import { RechargeService } from '@/application/services/RechargeService';
import { logger } from '@/infrastructure/logging/Logger';
import { businessMetrics } from '@/infrastructure/monitoring/Metrics';
import { useAuthContext } from '@/contexts/AuthContext';

// ============================================================================
// TYPES AND INTERFACES
// ============================================================================

export interface RechargeFormData {
  phoneNumber: string;
  networkProvider: NetworkProvider | '';
  rechargeType: RechargeType;
  amount: number;
  dataBundle?: string;
  paymentMethod: PaymentMethod;
  paymentReference?: string;
  affiliateCode?: string;
}

export interface RechargeValidationErrors {
  phoneNumber?: string;
  networkProvider?: string;
  amount?: string;
  dataBundle?: string;
  paymentMethod?: string;
  general?: string;
}

export interface RechargeState {
  // Form state
  formData: RechargeFormData;
  validationErrors: RechargeValidationErrors;
  
  // Processing state
  isProcessing: boolean;
  processingStep: string;
  processingProgress: number;
  
  // Result state
  lastResult: RechargeResult | null;
  transactionHistory: RechargeResult[];
  
  // UI state
  showSpinWheel: boolean;
  showAccountPrompt: boolean;
}

export interface RechargeHookOptions {
  onSuccess?: (result: RechargeResult) => void;
  onError?: (error: Error) => void;
  autoValidate?: boolean;
  enableMetrics?: boolean;
}

// ============================================================================
// VALIDATION UTILITIES
// ============================================================================

class RechargeValidator {
  static validatePhoneNumber(phoneNumber: string): string | null {
    try {
      PhoneNumber.create(phoneNumber);
      return null;
    } catch (error) {
      return error instanceof Error ? error.message : 'Invalid phone number';
    }
  }

  static validateAmount(amount: number, rechargeType: RechargeType): string | null {
    if (amount < 50) {
      return 'Minimum recharge amount is ₦50';
    }
    
    if (amount > 50000) {
      return 'Maximum recharge amount is ₦50,000';
    }
    
    if (rechargeType === RechargeType.AIRTIME && amount > 20000) {
      return 'Maximum airtime recharge is ₦20,000';
    }
    
    return null;
  }

  static validateNetworkProvider(provider: string): string | null {
    if (!provider) {
      return 'Please select a network provider';
    }
    
    const validProviders = Object.values(NetworkProvider);
    if (!validProviders.includes(provider as NetworkProvider)) {
      return 'Invalid network provider';
    }
    
    return null;
  }

  static validateDataBundle(bundle: string | undefined, rechargeType: RechargeType): string | null {
    if (rechargeType === RechargeType.DATA && !bundle) {
      return 'Please select a data bundle';
    }
    
    return null;
  }

  static validateForm(formData: RechargeFormData): RechargeValidationErrors {
    const errors: RechargeValidationErrors = {};
    
    const phoneError = this.validatePhoneNumber(formData.phoneNumber);
    if (phoneError) errors.phoneNumber = phoneError;
    
    const networkError = this.validateNetworkProvider(formData.networkProvider);
    if (networkError) errors.networkProvider = networkError;
    
    const amountError = this.validateAmount(formData.amount, formData.rechargeType);
    if (amountError) errors.amount = amountError;
    
    const bundleError = this.validateDataBundle(formData.dataBundle, formData.rechargeType);
    if (bundleError) errors.dataBundle = bundleError;
    
    return errors;
  }
}

// ============================================================================
// RECHARGE HOOK IMPLEMENTATION
// ============================================================================

export const useRecharge = (options: RechargeHookOptions = {}) => {
  const { 
    onSuccess, 
    onError, 
    autoValidate = true, 
    enableMetrics = true 
  } = options;
  
  const { user, isAuthenticated } = useAuthContext();
  const queryClient = useQueryClient();
  const correlationIdRef = useRef<string>();
  
  // ============================================================================
  // STATE MANAGEMENT
  // ============================================================================
  
  const [state, setState] = useState<RechargeState>({
    formData: {
      phoneNumber: user?.phoneNumber?.toString() || '',
      networkProvider: '',
      rechargeType: RechargeType.AIRTIME,
      amount: 0,
      paymentMethod: PaymentMethod.CARD
    },
    validationErrors: {},
    isProcessing: false,
    processingStep: '',
    processingProgress: 0,
    lastResult: null,
    transactionHistory: [],
    showSpinWheel: false,
    showAccountPrompt: false
  });
  
  // ============================================================================
  // FORM MANAGEMENT
  // ============================================================================
  
  const updateFormData = useCallback((updates: Partial<RechargeFormData>) => {
    setState(prev => {
      const newFormData = { ...prev.formData, ...updates };
      
      // Auto-validate if enabled
      const validationErrors = autoValidate 
        ? RechargeValidator.validateForm(newFormData)
        : {};
      
      return {
        ...prev,
        formData: newFormData,
        validationErrors
      };
    });
  }, [autoValidate]);
  
  const validateForm = useCallback((): boolean => {
    const errors = RechargeValidator.validateForm(state.formData);
    
    setState(prev => ({
      ...prev,
      validationErrors: errors
    }));
    
    return Object.keys(errors).length === 0;
  }, [state.formData]);
  
  const resetForm = useCallback(() => {
    setState(prev => ({
      ...prev,
      formData: {
        phoneNumber: user?.phoneNumber?.toString() || '',
        networkProvider: '',
        rechargeType: RechargeType.AIRTIME,
        amount: 0,
        paymentMethod: PaymentMethod.CARD
      },
      validationErrors: {},
      processingStep: '',
      processingProgress: 0
    }));
  }, [user]);
  
  // ============================================================================
  // RECHARGE PROCESSING
  // ============================================================================
  
  const rechargeService = new RechargeService(
    // Dependencies would be injected here in a real implementation
    {} as any, {} as any, {} as any, {} as any, {} as any, {} as any, logger, businessMetrics
  );
  
  const processRecharge = useMutation({
    mutationFn: async (formData: RechargeFormData): Promise<RechargeResult> => {
      // Generate correlation ID for tracking
      correlationIdRef.current = `recharge_${Date.now()}_${Math.random().toString(36).substring(2)}`;
      
      logger.info('Starting recharge process', {
        correlationId: correlationIdRef.current,
        phoneNumber: formData.phoneNumber,
        networkProvider: formData.networkProvider,
        amount: formData.amount
      });
      
      // Record metrics
      if (enableMetrics) {
        businessMetrics.recordRechargeRequest(
          formData.networkProvider, 
          formData.rechargeType
        );
      }
      
      // Simulate processing steps
      const steps = [
        { step: 'Validating request...', progress: 10 },
        { step: 'Processing payment...', progress: 30 },
        { step: 'Contacting network provider...', progress: 60 },
        { step: 'Confirming recharge...', progress: 80 },
        { step: 'Calculating rewards...', progress: 95 },
        { step: 'Complete!', progress: 100 }
      ];
      
      for (const { step, progress } of steps) {
        setState(prev => ({
          ...prev,
          processingStep: step,
          processingProgress: progress
        }));
        
        // Simulate processing time
        await new Promise(resolve => setTimeout(resolve, 800));
      }
      
      // Create recharge request
      const request: RechargeRequest = {
        phoneNumber: formData.phoneNumber,
        networkProvider: formData.networkProvider as NetworkProvider,
        rechargeType: formData.rechargeType,
        amount: formData.amount,
        paymentMethod: formData.paymentMethod,
        paymentReference: formData.paymentReference,
        userId: user?.id,
        affiliateCode: formData.affiliateCode,
        dataBundle: formData.dataBundle,
        metadata: {
          correlationId: correlationIdRef.current,
          userAgent: navigator.userAgent,
          timestamp: new Date().toISOString()
        }
      };
      
      // Process recharge (this would call the actual service)
      const result = await rechargeService.processRecharge(request);
      
      if (result.isFailure()) {
        throw result.error;
      }
      
      return result.value;
    },
    
    onMutate: () => {
      setState(prev => ({
        ...prev,
        isProcessing: true,
        processingStep: 'Initializing...',
        processingProgress: 0,
        validationErrors: {}
      }));
    },
    
    onSuccess: (result: RechargeResult) => {
      const duration = Date.now() - parseInt(correlationIdRef.current?.split('_')[1] || '0');
      
      logger.info('Recharge completed successfully', {
        correlationId: correlationIdRef.current,
        transactionId: result.transactionId.toString(),
        duration
      });
      
      // Record success metrics
      if (enableMetrics) {
        businessMetrics.recordRechargeSuccess(
          state.formData.networkProvider,
          state.formData.rechargeType,
          duration
        );
      }
      
      setState(prev => ({
        ...prev,
        isProcessing: false,
        lastResult: result,
        transactionHistory: [result, ...prev.transactionHistory.slice(0, 9)], // Keep last 10
        showSpinWheel: result.spinEligible,
        showAccountPrompt: !isAuthenticated && !prev.showAccountPrompt
      }));
      
      // Show success notification
      toast.success('🎉 Recharge Successful!', {
        description: `Transaction ${result.transactionId.toString().substring(0, 8)}... completed. You earned ${result.drawEntries} draw entries!`,
        duration: 5000
      });
      
      // Reset form
      resetForm();
      
      // Invalidate related queries
      queryClient.invalidateQueries({ queryKey: ['transactions'] });
      queryClient.invalidateQueries({ queryKey: ['user-stats'] });
      
      // Call success callback
      onSuccess?.(result);
    },
    
    onError: (error: Error) => {
      const duration = Date.now() - parseInt(correlationIdRef.current?.split('_')[1] || '0');
      
      logger.error('Recharge failed', {
        correlationId: correlationIdRef.current,
        error: error.message,
        duration
      });
      
      // Record failure metrics
      if (enableMetrics) {
        businessMetrics.recordRechargeFailure(
          state.formData.networkProvider,
          state.formData.rechargeType,
          error.constructor.name
        );
      }
      
      setState(prev => ({
        ...prev,
        isProcessing: false,
        validationErrors: {
          general: error.message || 'An unexpected error occurred'
        }
      }));
      
      // Show error notification
      toast.error('Recharge Failed', {
        description: error.message || 'Please try again later',
        duration: 5000
      });
      
      // Call error callback
      onError?.(error);
    }
  });
  
  // ============================================================================
  // TRANSACTION HISTORY
  // ============================================================================
  
  const { data: transactionHistory, isLoading: isLoadingHistory } = useQuery({
    queryKey: ['transactions', user?.id],
    queryFn: async () => {
      // This would fetch from the actual API
      return [];
    },
    enabled: isAuthenticated,
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000 // 10 minutes
  });
  
  // ============================================================================
  // UTILITY FUNCTIONS
  // ============================================================================
  
  const calculateRewards = useCallback((amount: number) => {
    const points = amount; // 1:1 ratio
    const drawEntries = Math.floor(points / 200);
    const spinEligible = amount >= 1000;
    
    return { points, drawEntries, spinEligible };
  }, []);
  
  const getNetworkFromPhone = useCallback((phoneNumber: string): NetworkProvider | null => {
    const cleaned = phoneNumber.replace(/\\D/g, '');
    const prefix = cleaned.substring(cleaned.length - 10, cleaned.length - 7);
    
    // Nigerian network prefixes
    const prefixMap: Record<string, NetworkProvider> = {
      '803': NetworkProvider.MTN,
      '806': NetworkProvider.MTN,
      '813': NetworkProvider.MTN,
      '816': NetworkProvider.MTN,
      '810': NetworkProvider.MTN,
      '814': NetworkProvider.MTN,
      '903': NetworkProvider.MTN,
      '906': NetworkProvider.MTN,
      
      '802': NetworkProvider.AIRTEL,
      '808': NetworkProvider.AIRTEL,
      '812': NetworkProvider.AIRTEL,
      '901': NetworkProvider.AIRTEL,
      '902': NetworkProvider.AIRTEL,
      '904': NetworkProvider.AIRTEL,
      '907': NetworkProvider.AIRTEL,
      
      '805': NetworkProvider.GLO,
      '807': NetworkProvider.GLO,
      '811': NetworkProvider.GLO,
      '815': NetworkProvider.GLO,
      '905': NetworkProvider.GLO,
      
      '809': NetworkProvider.NINE_MOBILE,
      '817': NetworkProvider.NINE_MOBILE,
      '818': NetworkProvider.NINE_MOBILE,
      '908': NetworkProvider.NINE_MOBILE,
      '909': NetworkProvider.NINE_MOBILE
    };
    
    return prefixMap[prefix] || null;
  }, []);
  
  // ============================================================================
  // EFFECTS
  // ============================================================================
  
  // Auto-detect network from phone number
  useEffect(() => {
    if (state.formData.phoneNumber && !state.formData.networkProvider) {
      const detectedNetwork = getNetworkFromPhone(state.formData.phoneNumber);
      if (detectedNetwork) {
        updateFormData({ networkProvider: detectedNetwork });
      }
    }
  }, [state.formData.phoneNumber, state.formData.networkProvider, getNetworkFromPhone, updateFormData]);
  
  // ============================================================================
  // RETURN INTERFACE
  // ============================================================================
  
  return {
    // State
    formData: state.formData,
    validationErrors: state.validationErrors,
    isProcessing: state.isProcessing,
    processingStep: state.processingStep,
    processingProgress: state.processingProgress,
    lastResult: state.lastResult,
    showSpinWheel: state.showSpinWheel,
    showAccountPrompt: state.showAccountPrompt,
    
    // Actions
    updateFormData,
    validateForm,
    resetForm,
    processRecharge: processRecharge.mutate,
    
    // Utilities
    calculateRewards,
    getNetworkFromPhone,
    
    // Query states
    isLoadingHistory,
    transactionHistory: transactionHistory || state.transactionHistory,
    
    // Computed values
    isFormValid: Object.keys(state.validationErrors).length === 0,
    canSubmit: !state.isProcessing && Object.keys(state.validationErrors).length === 0 && state.formData.amount >= 50,
    
    // UI helpers
    dismissSpinWheel: () => setState(prev => ({ ...prev, showSpinWheel: false })),
    dismissAccountPrompt: () => setState(prev => ({ ...prev, showAccountPrompt: false }))
  };
};

// ============================================================================
// HOOK VARIANTS
// ============================================================================

// Simplified hook for basic recharge functionality
export const useSimpleRecharge = () => {
  return useRecharge({
    autoValidate: true,
    enableMetrics: false
  });
};

// Hook with enhanced error handling for admin interfaces
export const useAdminRecharge = () => {
  return useRecharge({
    autoValidate: true,
    enableMetrics: true,
    onError: (error) => {
      // Enhanced error reporting for admin users
      console.error('Admin recharge error:', error);
    }
  });
};

export default useRecharge;