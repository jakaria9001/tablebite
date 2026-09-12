import { apiFetch } from "./client";
import { clearStoredAdminToken, setStoredAdminToken } from "./session";

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
  return apiFetch<{ email: string; role: string }>('/api/v1/auth/me');
}
