import { useCallback, useEffect, useMemo, useState, type ReactNode } from "react";
import {
  fetchCurrentCustomer,
  logout as logoutCustomer,
  requestOtp as requestCustomerOtp,
  verifyOtp as verifyCustomerOtp,
  type Customer,
} from "../api/auth";
import { AuthContext, type AuthContextValue } from "./auth-context";

export function AuthProvider({ children }: { children: ReactNode }) {
  const [customer, setCustomer] = useState<Customer | null>(null);
  const [loading, setLoading] = useState(true);

  const refreshCustomer = useCallback(async () => {
    try {
      setCustomer(await fetchCurrentCustomer());
    } catch {
      setCustomer(null);
    }
  }, []);

  useEffect(() => {
    let cancelled = false;
    void fetchCurrentCustomer()
      .then((currentCustomer) => {
        if (!cancelled) setCustomer(currentCustomer);
      })
      .catch(() => {
        if (!cancelled) setCustomer(null);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, []);

  const requestOtp = useCallback(async (phone: string) => {
    await requestCustomerOtp(phone);
  }, []);

  const verifyOtp = useCallback(async (phone: string, otp: string) => {
    setCustomer(await verifyCustomerOtp(phone, otp));
  }, []);

  const logout = useCallback(async () => {
    try {
      await logoutCustomer();
    } finally {
      setCustomer(null);
    }
  }, []);

  const value = useMemo<AuthContextValue>(() => ({
    customer,
    isAuthenticated: customer !== null,
    loading,
    requestOtp,
    verifyOtp,
    refreshCustomer,
    logout,
  }), [customer, loading, requestOtp, verifyOtp, refreshCustomer, logout]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
