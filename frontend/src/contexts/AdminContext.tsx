import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { apiClient } from '@/lib/api-client';

interface AdminUser {
  id: string;
  username: string;
  email: string;
  role: 'super_admin' | 'admin' | 'moderator';
  permissions: string[];
  last_login: string;
  created_at: string;
}

interface AdminContextType {
  adminUser: AdminUser | null;
  admin: AdminUser | null;  // Alias for compatibility
  isAdminAuthenticated: boolean;
  isAuthenticated: boolean;  // Alias for compatibility
  sessionToken: string | null;
  adminLogin: (credentials: { username?: string; email?: string; password: string }) => Promise<boolean>;
  login: (username: string, password: string) => Promise<boolean>;  // Alias for compatibility
  adminLogout: () => void;
  logout: () => void;  // Alias for compatibility
  hasPermission: (permission: string) => boolean;
  isLoading: boolean;
}

const AdminContext = createContext<AdminContextType | undefined>(undefined);


export const AdminProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [adminUser, setAdminUser] = useState<AdminUser | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // On mount, restore admin session from localStorage.
    // We use Bearer token auth (cross-domain: httpOnly cookie won't work Vercel→Render).
    const checkAdminSession = async () => {
      try {
        const storedAdmin = localStorage.getItem('rechargemax_admin_user');
        const storedToken = localStorage.getItem('rechargemax_admin_token');
        
        if (!storedAdmin || !storedToken) {
          // No stored session
          setIsLoading(false);
          return;
        }

        // Validate the JWT token is still valid by calling a lightweight admin endpoint
        const res = await apiClient.get('/admin/users?page=1&limit=1');
        if (res.data?.success !== false) {
          setAdminUser(JSON.parse(storedAdmin));
        } else {
          // Token expired — clear session
          localStorage.removeItem('rechargemax_admin_user');
          localStorage.removeItem('rechargemax_admin_token');
        }
      } catch {
        // Network error or 401: clear stale session to force fresh login
        localStorage.removeItem('rechargemax_admin_user');
        localStorage.removeItem('rechargemax_admin_token');
      } finally {
        setIsLoading(false);
      }
    };

    checkAdminSession();
  }, []);

  const adminLogin = async (credentials: { username?: string; email?: string; password: string }): Promise<boolean> => {
    try {
      setIsLoading(true);

      const response = await apiClient.post('/admin/auth/login', {
        email: credentials.email || credentials.username,
        password: credentials.password,
      });

      const data = response.data;

      if (data.success && data.admin) {
        // Token is in the httpOnly cookie — only cache non-sensitive profile data
        const adminData: AdminUser = {
          id: data.admin.id,
          username: data.admin.email,
          email: data.admin.email,
          role: (data.admin.role || 'admin').toLowerCase() as 'super_admin' | 'admin' | 'moderator',
          permissions: data.admin.permissions || [],
          last_login: new Date().toISOString(),
          created_at: data.admin.created_at,
        };
        setAdminUser(adminData);
        localStorage.setItem('rechargemax_admin_user', JSON.stringify(adminData));
        // Store token for cross-domain admin API calls (Bearer token needed since
        // httpOnly cookie is same-domain only and won't work cross-domain on Render→Vercel)
        const jwtToken = data.token;
        if (jwtToken) {
          localStorage.setItem('rechargemax_admin_token', jwtToken);
        }
        return true;
      }

      return false;
    } catch (error) {
      console.error('Admin login error:', error);
      return false;
    } finally {
      setIsLoading(false);
    }
  };

  const adminLogout = async () => {
    try {
      // Ask the backend to clear the httpOnly cookie
      await apiClient.post('/admin/auth/logout');
    } catch {
      // Proceed with local cleanup even if request fails
    }
    setAdminUser(null);
    localStorage.removeItem('rechargemax_admin_user');
  };

  const hasPermission = (permission: string): boolean => {
    if (!adminUser) return false;
    if (adminUser.permissions.includes('all')) return true;
    return adminUser.permissions.includes(permission);
  };

  const value: AdminContextType = {
    adminUser,
    admin: adminUser,
    isAdminAuthenticated: !!adminUser,
    isAuthenticated: !!adminUser,
    sessionToken: localStorage.getItem('rechargemax_admin_token'), // Read from localStorage for Bearer auth
    adminLogin,
    login: (username: string, password: string) => adminLogin({ username, password }),
    adminLogout,
    logout: adminLogout,
    hasPermission,
    isLoading,
  };

  return (
    <AdminContext.Provider value={value}>
      {children}
    </AdminContext.Provider>
  );
};

export const useAdminContext = (): AdminContextType => {
  const context = useContext(AdminContext);
  if (context === undefined) {
    throw new Error('useAdminContext must be used within an AdminProvider');
  }
  return context;
};

export default AdminContext;
