import { getStoredAdminToken } from "./session";

export type ApiErrorKind = "offline" | "unauthorized" | "not_found" | "validation" | "server" | "unknown";

export class ApiError extends Error {
  constructor(public readonly kind: ApiErrorKind, message: string, public readonly status?: number, public readonly requestId?: string) {
    super(message);
    this.name = "ApiError";
  }
}

export async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const baseUrl = import.meta.env.VITE_API_BASE_URL ?? "";
  const token = getStoredAdminToken();
  let response: Response;
  try {
    response = await fetch(`${baseUrl}${path}`, {
      ...options,
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...options?.headers,
      },
    });
  } catch {
    throw new ApiError("offline", navigator.onLine ? "Unable to reach the server." : "You appear to be offline.");
  }

  const body = await response.text();

  if (!response.ok) {
    let message = body || `Request failed with status ${response.status}`;
    let requestId: string | undefined;
    try {
      const payload = JSON.parse(body) as { error?: string; request_id?: string };
      message = payload.error || message;
      requestId = payload.request_id;
    } catch {
      // Preserve non-JSON server responses.
    }
    const kind: ApiErrorKind = response.status === 401 ? "unauthorized" : response.status === 404 ? "not_found" : response.status >= 500 ? "server" : response.status >= 400 ? "validation" : "unknown";
    throw new ApiError(kind, message, response.status, requestId);
  }

  if (!body) {
    return undefined as T;
  }

  try {
    return JSON.parse(body) as T;
  } catch {
    // A non-JSON 200 body usually means the request never reached the API (e.g. a routing misconfiguration).
    throw new ApiError("server", "Received an unexpected response from the server.");
  }
}

export async function apiGet<T>(path: string): Promise<T> {
  return apiFetch<T>(path, { method: "GET", headers: { Accept: "application/json" } });
}
