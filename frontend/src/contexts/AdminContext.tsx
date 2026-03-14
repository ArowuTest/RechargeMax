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

interface AdminProviderProps {
  children: ReactNode;
}

export const AdminProvider: React.FC<AdminProviderProps> = ({ children }) => {
  const [adminUser, setAdminUser] = useState<AdminUser | null>(null);
  const [sessionToken, setSessionToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Check for existing admin session and validate the token with the backend
    const checkAdminSession = async () => {
      try {
        const storedAdmin = localStorage.getItem('rechargemax_admin_user');
        const storedToken = localStorage.getItem('rechargemax_admin_token');

        if (storedAdmin && storedToken) {
          // Validate token with backend to ensure it hasn't expired
          const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || '/api/v1';
          try {
            const res = await fetch(`${apiBaseUrl}/admin/dashboard`, {
              headers: { Authorization: `Bearer ${storedToken}` },
            });

            if (res.ok) {
              // Token is valid — restore session
              const adminData = JSON.parse(storedAdmin);
              setAdminUser(adminData);
              setSessionToken(storedToken);
            } else {
              // Token expired or invalid — clear stale session
              localStorage.removeItem('rechargemax_admin_user');
              localStorage.removeItem('rechargemax_admin_token');
            }
          } catch {
            // Network error — restore session optimistically to avoid locking out admins
            const adminData = JSON.parse(storedAdmin);
            setAdminUser(adminData);
            setSessionToken(storedToken);
          }
        }
      } catch (error) {
        console.error('Error checking admin session:', error);
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

      // Call real backend API
      const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || '/api/v1';
      const response = await fetch(`${apiBaseUrl}/admin/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email: credentials.email || credentials.username,
          password: credentials.password,
        }),
      });

      const data = await response.json();

      if (data.success && data.token) {
        // Store token
        localStorage.setItem('rechargemax_admin_token', data.token);

        // Store admin user data
        const adminData: AdminUser = {
          id: data.admin.id,
          username: data.admin.email,
          email: data.admin.email,
          role: data.admin.role.toLowerCase() as 'super_admin' | 'admin' | 'moderator',
          permissions: data.admin.permissions || [],
          last_login: new Date().toISOString(),
          created_at: data.admin.created_at,
        };

        setAdminUser(adminData);
        setSessionToken(data.token);
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

  const adminLogout = () => {
    setAdminUser(null);
    setSessionToken(null);
    localStorage.removeItem('rechargemax_admin_user');
    localStorage.removeItem('rechargemax_admin_token');
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
    sessionToken,
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
