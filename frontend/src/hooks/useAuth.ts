import { useState, useEffect, useCallback } from 'react';
import { useToast } from '@/hooks/use-toast';
import { authApi, userApi } from '@/lib/api-client';

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

  // On mount: verify session by calling /user/profile (uses httpOnly cookie)
  // Non-sensitive profile data may be cached in localStorage for UI speed
  useEffect(() => {
    const checkSession = async () => {
      try {
        // Try to restore from localStorage cache (non-sensitive profile data only)
        const storedUser = localStorage.getItem('rechargemax_user');
        if (storedUser) {
          const userData = JSON.parse(storedUser);
          setAuthState({ user: userData, isAuthenticated: true, isLoading: false });
          return;
        }

        // Verify with backend via cookie
        const res = await userApi.getProfile();
        if (res.success && res.data) {
          const userData = res.data as unknown as User;
          localStorage.setItem('rechargemax_user', JSON.stringify(userData));
          setAuthState({ user: userData, isAuthenticated: true, isLoading: false });
        } else {
          setAuthState(prev => ({ ...prev, isLoading: false }));
        }
      } catch {
        // No valid session
        localStorage.removeItem('rechargemax_user');
        setAuthState({ user: null, isAuthenticated: false, isLoading: false });
      }
    };

    checkSession();
  }, []);

  const login = useCallback((userData: User) => {
    try {
      // Store only non-sensitive profile data in localStorage (token is in httpOnly cookie)
      localStorage.setItem('rechargemax_user', JSON.stringify(userData));

      setAuthState({ user: userData, isAuthenticated: true, isLoading: false });

      toast({
        title: "Login Successful! 🎉",
        description: `Welcome back, ${userData.full_name || 'User'}!`,
      });
    } catch (error) {
      console.error('Login error:', error);
    }
  }, [toast]);

  const logout = useCallback(async () => {
    try {
      await authApi.logout(); // clears httpOnly cookie via Set-Cookie
    } catch { /* ignore */ } finally {
      localStorage.removeItem('rechargemax_user');
      setAuthState({ user: null, isAuthenticated: false, isLoading: false });

      toast({ title: "Logged Out", description: "You have been successfully logged out." });
      window.location.href = '/#/';
    }
  }, [toast]);

  const updateUser = useCallback((updates: Partial<User>) => {
    if (!authState.user) return;
    const updatedUser = { ...authState.user, ...updates };
    try {
      localStorage.setItem('rechargemax_user', JSON.stringify(updatedUser));
    } catch { /* ignore */ }
    setAuthState(prev => ({ ...prev, user: updatedUser }));
  }, [authState.user]);

  return { ...authState, login, logout, updateUser };
};

export default useAuth;
