import { apiFetch } from "./client";

export async function loginAdmin(email: string, password: string) {
  return apiFetch<{ ok: boolean }>('/api/v1/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  });
}

export async function logoutAdmin() {
  return apiFetch<void>('/api/v1/auth/logout', { method: 'POST' });
}

export async function getAdminMe() {
  return apiFetch<{ email: string; role: string }>('/api/v1/auth/me');
}
