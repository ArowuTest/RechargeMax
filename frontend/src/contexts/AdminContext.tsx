import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';

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

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1';

export const AdminProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [adminUser, setAdminUser] = useState<AdminUser | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // On mount, check if there is a valid session by calling the backend.
    // The httpOnly cookie is sent automatically — we don't touch localStorage for the token.
    const checkAdminSession = async () => {
      try {
        const storedAdmin = localStorage.getItem('rechargemax_admin_user');
        if (!storedAdmin) {
          setIsLoading(false);
          return;
        }

        // Validate session against backend (cookie sent automatically)
        const res = await fetch(`${API_BASE_URL}/admin/dashboard`, {
          credentials: 'include', // send httpOnly cookie
        });

        if (res.ok) {
          setAdminUser(JSON.parse(storedAdmin));
        } else {
          // Cookie expired or invalid — clear stale profile cache
          localStorage.removeItem('rechargemax_admin_user');
        }
      } catch {
        // Network error — restore session optimistically to avoid locking out admins
        const storedAdmin = localStorage.getItem('rechargemax_admin_user');
        if (storedAdmin) setAdminUser(JSON.parse(storedAdmin));
      } finally {
        setIsLoading(false);
      }
    };

    checkAdminSession();
  }, []);

  const adminLogin = async (credentials: { username?: string; email?: string; password: string }): Promise<boolean> => {
    try {
      setIsLoading(true);

      const response = await fetch(`${API_BASE_URL}/admin/login`, {
        method: 'POST',
        credentials: 'include', // receive + store httpOnly cookie
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: credentials.email || credentials.username,
          password: credentials.password,
        }),
      });

      const data = await response.json();

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
      await fetch(`${API_BASE_URL}/admin/logout`, {
        method: 'POST',
        credentials: 'include',
      });
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
    sessionToken: null, // Token is in httpOnly cookie — not exposed to JS
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
