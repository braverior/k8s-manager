import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react';
import { authApi } from '@/api';
import type { User } from '@/types';

interface AuthContextType {
  user: User | null;
  token: string | null;
  loading: boolean;
  isAuthenticated: boolean;
  isAdmin: boolean;
  login: (token: string, user: User) => void;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
  hasClusterPermission: (cluster: string) => boolean;
  hasNamespacePermission: (cluster: string, namespace: string) => boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Auth API uses same-origin (backend is co-located)
function getAuthApiServer(): string {
  return '';
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  // Initialize from localStorage
  useEffect(() => {
    const storedToken = localStorage.getItem('token');
    const storedUser = localStorage.getItem('user');

    if (storedToken && storedUser) {
      setToken(storedToken);
      try {
        setUser(JSON.parse(storedUser));
      } catch {
        localStorage.removeItem('user');
      }
    }
    setLoading(false);
  }, []);

  const login = useCallback((newToken: string, newUser: User) => {
    localStorage.setItem('token', newToken);
    localStorage.setItem('user', JSON.stringify(newUser));
    setToken(newToken);
    setUser(newUser);
  }, []);

  const logout = useCallback(async () => {
    try {
      await authApi.logout(getAuthApiServer());
    } catch {
      // Ignore logout errors
    } finally {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      setToken(null);
      setUser(null);
    }
  }, []);

  const refreshUser = useCallback(async () => {
    if (!token) return;
    try {
      const userData = await authApi.getMe(getAuthApiServer());
      setUser(userData);
      localStorage.setItem('user', JSON.stringify(userData));
    } catch {
      // If refresh fails, logout
      await logout();
    }
  }, [token, logout]);

  const hasClusterPermission = useCallback(
    (cluster: string): boolean => {
      if (!user) return false;
      if (user.is_admin) return true;
      if (!user.permissions) return false;
      return user.permissions.some((p) => p.cluster === cluster);
    },
    [user]
  );

  const hasNamespacePermission = useCallback(
    (cluster: string, namespace: string): boolean => {
      if (!user) return false;
      if (user.is_admin) return true;
      if (!user.permissions) return false;
      const clusterPerm = user.permissions.find((p) => p.cluster === cluster);
      if (!clusterPerm) return false;
      return clusterPerm.namespaces.includes('*') || clusterPerm.namespaces.includes(namespace);
    },
    [user]
  );

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        loading,
        isAuthenticated: !!token && !!user,
        isAdmin: user?.is_admin ?? false,
        login,
        logout,
        refreshUser,
        hasClusterPermission,
        hasNamespacePermission,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
