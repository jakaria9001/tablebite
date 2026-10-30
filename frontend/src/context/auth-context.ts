import { createContext } from "react";
import type { Customer } from "../api/auth";

export interface AuthContextValue {
  customer: Customer | null;
  isAuthenticated: boolean;
  loading: boolean;
  requestOtp(phone: string): Promise<void>;
  verifyOtp(phone: string, otp: string): Promise<void>;
  refreshCustomer(): Promise<void>;
  logout(): Promise<void>;
}

export const AuthContext = createContext<AuthContextValue | undefined>(undefined);
