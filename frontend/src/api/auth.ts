import { ApiError, apiFetch } from "./client";
import { clearStoredAdminToken, setStoredAdminToken } from "./session";

export type Customer = {
  id: number;
  phoneNumber: string;
  phoneVerified: boolean;
};

type CustomerResponse = {
  customer: {
    id: number;
    phone?: string;
    phoneNumber?: string;
    phone_verified?: boolean;
    phoneVerified?: boolean;
  };
};

function parseCustomerResponse(response: CustomerResponse): Customer {
  const customer = response.customer;
  const phoneNumber = customer.phoneNumber ?? customer.phone;
  const phoneVerified = customer.phoneVerified ?? customer.phone_verified;

  if (typeof customer.id !== "number" || typeof phoneNumber !== "string" || typeof phoneVerified !== "boolean") {
    throw new ApiError("unauthorized", "Customer authentication is unavailable.");
  }

  return { id: customer.id, phoneNumber, phoneVerified };
}

export async function loginAdmin(email: string, password: string) {
  const result = await apiFetch<{ ok: boolean; token?: string }>('/api/v1/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email: email.trim(), password: password.trim() }),
  });
  if (result.token) {
    setStoredAdminToken(result.token);
  }
  return result;
}

export async function logoutAdmin() {
  try {
    return await apiFetch<void>('/api/v1/auth/logout', { method: 'POST' });
  } finally {
    clearStoredAdminToken();
  }
}

export async function getAdminMe() {
  const result = await apiFetch<{ email?: string; role?: string }>("/api/v1/auth/me");
  if (typeof result.email !== "string" || typeof result.role !== "string") {
    throw new ApiError("unauthorized", "Admin authentication is required.");
  }
  return { email: result.email, role: result.role };
}

export async function requestOtp(phone: string): Promise<void> {
  await apiFetch<void>("/api/v1/auth/otp/request", {
    method: "POST",
    body: JSON.stringify({ phone }),
    includeAdminAuth: false,
  });
}

export async function verifyOtp(phone: string, otp: string): Promise<Customer> {
  const result = await apiFetch<CustomerResponse>("/api/v1/auth/otp/verify", {
    method: "POST",
    body: JSON.stringify({ phone, otp }),
    includeAdminAuth: false,
  });
  return parseCustomerResponse(result);
}

export async function fetchCurrentCustomer(): Promise<Customer> {
  const result = await apiFetch<CustomerResponse>("/api/v1/auth/me", {
    method: "GET",
    headers: { Accept: "application/json" },
    includeAdminAuth: false,
  });
  return parseCustomerResponse(result);
}

export async function logout(): Promise<void> {
  await apiFetch<void>("/api/v1/auth/logout", {
    method: "POST",
    includeAdminAuth: false,
  });
}
