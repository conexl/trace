import * as React from 'react';
import { setAdminToken } from './api';

interface AuthContextValue {
  token: string;
  isAuthenticated: boolean;
  login: (token: string) => void;
  logout: () => void;
}

const AuthContext = React.createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setToken] = React.useState(() => {
    return typeof window !== 'undefined' ? localStorage.getItem('homelytics-token') ?? '' : '';
  });

  React.useEffect(() => {
    setAdminToken(token);
  }, [token]);

  const login = React.useCallback((newToken: string) => {
    localStorage.setItem('homelytics-token', newToken);
    setToken(newToken);
    setAdminToken(newToken);
  }, []);

  const logout = React.useCallback(() => {
    localStorage.removeItem('homelytics-token');
    setToken('');
    setAdminToken('');
  }, []);

  const value = React.useMemo(
    () => ({
      token,
      isAuthenticated: token.length > 0,
      login,
      logout,
    }),
    [token, login, logout]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const context = React.useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}
