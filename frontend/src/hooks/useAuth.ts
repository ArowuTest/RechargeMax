import { useState, useEffect, useCallback } from 'react';
import { useToast } from '@/hooks/useToast';
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

// Cache TTL: 5 minutes. After this, the profile is always re-fetched from the backend.
// This ensures admin changes (tier updates, point adjustments, bans) are reflected promptly.
const PROFILE_CACHE_TTL_MS = 5 * 60 * 1000;
const CACHE_KEY  = 'rechargemax_user';
const CACHE_TS_KEY = 'rechargemax_user_ts';

function readCache(): User | null {
  try {
    const ts = parseInt(localStorage.getItem(CACHE_TS_KEY) || '0', 10);
    if (Date.now() - ts > PROFILE_CACHE_TTL_MS) {
      localStorage.removeItem(CACHE_KEY);
      localStorage.removeItem(CACHE_TS_KEY);
      return null;
    }
    const raw = localStorage.getItem(CACHE_KEY);
    return raw ? JSON.parse(raw) : null;
  } catch { return null; }
}

function writeCache(user: User) {
  try {
    localStorage.setItem(CACHE_KEY, JSON.stringify(user));
    localStorage.setItem(CACHE_TS_KEY, String(Date.now()));
  } catch { /* storage full or disabled */ }
}

function clearCache() {
  localStorage.removeItem(CACHE_KEY);
  localStorage.removeItem(CACHE_TS_KEY);
}

export const useAuth = () => {
  const [authState, setAuthState] = useState<AuthState>({
    user: null,
    isAuthenticated: false,
    isLoading: true
  });
  const { toast } = useToast();

  useEffect(() => {
    const checkSession = async () => {
      // 1. Show cached data immediately for a fast UI render (avoids blank flash)
      const cached = readCache();
      if (cached) {
        setAuthState({ user: cached, isAuthenticated: true, isLoading: true });
      }

      // 2. ALWAYS verify with the backend — cache is only a render hint, not the source of truth.
      //    This ensures bans, tier changes, and point adjustments are picked up within 5 min.
      try {
        const res = await userApi.getProfile();
        if (res.success && res.data) {
          const userData = res.data as unknown as User;
          writeCache(userData);
          setAuthState({ user: userData, isAuthenticated: true, isLoading: false });
        } else {
          clearCache();
          setAuthState({ user: null, isAuthenticated: false, isLoading: false });
        }
      } catch {
        if (cached) {
          // Backend unreachable but cache is fresh — keep the user logged in with stale data
          setAuthState({ user: cached, isAuthenticated: true, isLoading: false });
        } else {
          clearCache();
          setAuthState({ user: null, isAuthenticated: false, isLoading: false });
        }
      }
    };

    checkSession();
  }, []);

  const login = useCallback((userData: User) => {
    try {
      writeCache(userData);
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
      await authApi.logout();
    } catch { /* ignore */ } finally {
      clearCache();
      setAuthState({ user: null, isAuthenticated: false, isLoading: false });
      toast({ title: "Logged Out", description: "You have been successfully logged out." });
      window.location.href = '/';
    }
  }, [toast]);

  const updateUser = useCallback((updates: Partial<User>) => {
    if (!authState.user) return;
    const updatedUser = { ...authState.user, ...updates };
    writeCache(updatedUser);
    setAuthState(prev => ({ ...prev, user: updatedUser }));
  }, [authState.user]);

  return { ...authState, login, logout, updateUser };
};

export default useAuth;
