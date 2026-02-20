import { useState, useEffect, useCallback } from 'react';
import { useToast } from '@/hooks/use-toast';
import { authApi } from '@/lib/api-client';

interface User {
  id: string;
  msisdn: string;
  full_name: string;
  email: string;
  loyalty_tier: string;
  total_points: number;
  total_recharges: number;
}

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

export const useAuth = () => {
  const [authState, setAuthState] = useState<AuthState>({
    user: null,
    isAuthenticated: false,
    isLoading: true
  });
  const { toast } = useToast();

  // Check for existing session on mount
  useEffect(() => {
    const checkSession = () => {
      try {
        const storedUser = localStorage.getItem('rechargemax_user');
        const storedToken = localStorage.getItem('rechargemax_token');
        
        if (storedUser && storedToken) {
          const userData = JSON.parse(storedUser);
          setAuthState({
            user: userData,
            isAuthenticated: true,
            isLoading: false
          });
        } else {
          setAuthState(prev => ({ ...prev, isLoading: false }));
        }
      } catch (error) {
        console.error('Error checking session:', error);
        localStorage.removeItem('rechargemax_user');
        localStorage.removeItem('rechargemax_token');
        setAuthState(prev => ({ ...prev, isLoading: false }));
      }
    };

    checkSession();
  }, []);

  const login = useCallback((userData: User, token?: string) => {
    try {
      // Store user data
      localStorage.setItem('rechargemax_user', JSON.stringify(userData));
      
      // Store token if provided (for backward compatibility)
      if (token) {
        localStorage.setItem('rechargemax_token', token);
      }
      
      setAuthState({
        user: userData,
        isAuthenticated: true,
        isLoading: false
      });

      toast({
        title: "Login Successful! 🎉",
        description: `Welcome back, ${userData.full_name || 'User'}!`,
      });
    } catch (error) {
      console.error('Login error:', error);
      toast({
        title: "Login Error",
        description: "Failed to save user session",
        variant: "destructive"
      });
    }
  }, [toast]);

  const logout = useCallback(async () => {
    try {
      // Call logout API
      await authApi.logout();
      
      // Clear local storage
      localStorage.removeItem('rechargemax_user');
      localStorage.removeItem('rechargemax_token');
      localStorage.removeItem('auth_token'); // Also clear old token key
      
      setAuthState({
        user: null,
        isAuthenticated: false,
        isLoading: false
      });

      toast({
        title: "Logged Out",
        description: "You have been successfully logged out",
      });
      
      // Redirect to home
      window.location.href = '/#/';
    } catch (error) {
      console.error('Logout error:', error);
      
      // Force logout even if API call fails
      localStorage.removeItem('rechargemax_user');
      localStorage.removeItem('rechargemax_token');
      localStorage.removeItem('auth_token');
      
      setAuthState({
        user: null,
        isAuthenticated: false,
        isLoading: false
      });
      
      window.location.href = '/#/';
    }
  }, [toast]);

  const updateUser = useCallback((updates: Partial<User>) => {
    if (!authState.user) return;

    const updatedUser = { ...authState.user, ...updates };
    
    try {
      localStorage.setItem('rechargemax_user', JSON.stringify(updatedUser));
      setAuthState(prev => ({
        ...prev,
        user: updatedUser
      }));
    } catch (error) {
      console.error('Update user error:', error);
      toast({
        title: "Update Error",
        description: "Failed to update user information",
        variant: "destructive"
      });
    }
  }, [authState.user, toast]);

  return {
    ...authState,
    login,
    logout,
    updateUser
  };
};

export default useAuth;
