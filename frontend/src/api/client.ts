import { getStoredAdminToken } from "./session";

export type ApiErrorKind = "offline" | "unauthorized" | "not_found" | "validation" | "server" | "unknown";

export class ApiError extends Error {
  constructor(public readonly kind: ApiErrorKind, message: string, public readonly status?: number, public readonly requestId?: string) {
    super(message);
    this.name = "ApiError";
  }
}

export type ApiFetchOptions = RequestInit & {
  includeAdminAuth?: boolean;
};

let csrfTokenPromise: Promise<string> | null = null;

async function fetchCsrfToken(baseUrl: string): Promise<string> {
  const response = await fetch(`${baseUrl}/api/v1/auth/csrf`, {
    method: "GET",
    credentials: "include",
    headers: { Accept: "application/json" },
  });
  const body = await response.text();
  if (!response.ok) {
    throw new ApiError("server", "Unable to initialize request security.", response.status);
  }

  try {
    const payload = JSON.parse(body) as { token?: string };
    return payload.token ?? "";
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }
    throw new ApiError("server", "Unable to initialize request security.");
  }
}

function getCsrfToken(baseUrl: string): Promise<string> {
  if (!csrfTokenPromise) {
    csrfTokenPromise = fetchCsrfToken(baseUrl).catch((error) => {
      csrfTokenPromise = null;
      throw error;
    });
  }
  return csrfTokenPromise;
}

export async function apiFetch<T>(path: string, options?: ApiFetchOptions): Promise<T> {
  const baseUrl = import.meta.env.VITE_API_BASE_URL ?? "";
  const token = getStoredAdminToken();
  const { includeAdminAuth = true, ...requestOptions } = options ?? {};
  const headers = new Headers(requestOptions.headers);
  headers.set("Content-Type", "application/json");
  if (includeAdminAuth && token) {
    headers.set("Authorization", `Bearer ${token}`);
  }
  const method = (requestOptions.method ?? "GET").toUpperCase();
  let response: Response;
  try {
    if (!["GET", "HEAD", "OPTIONS"].includes(method) && !headers.has("X-CSRF-Token")) {
      const csrfToken = await getCsrfToken(baseUrl);
      if (csrfToken) {
        headers.set("X-CSRF-Token", csrfToken);
      }
    }
    response = await fetch(`${baseUrl}${path}`, {
      ...requestOptions,
      credentials: "include",
      headers,
    });
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }
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
    throw new ApiError("server", "Received an unexpected response from the server.");
  }
}

export async function apiGet<T>(path: string): Promise<T> {
  return apiFetch<T>(path, { method: "GET", headers: { Accept: "application/json" } });
}
