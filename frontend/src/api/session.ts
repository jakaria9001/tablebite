const TOKEN_KEY = "tablebite_admin_token";

export function getStoredAdminToken(): string | null {
  try {
    return localStorage.getItem(TOKEN_KEY);
  } catch {
    return null;
  }
}

export function setStoredAdminToken(token: string): void {
  try {
    localStorage.setItem(TOKEN_KEY, token);
  } catch {
    // Ignore storage failures (e.g. private browsing quota limits).
  }
}

export function clearStoredAdminToken(): void {
  try {
    localStorage.removeItem(TOKEN_KEY);
  } catch {
    // Ignore storage failures.
  }
}
