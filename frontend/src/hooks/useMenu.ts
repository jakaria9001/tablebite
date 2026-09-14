import { useEffect, useState } from "react";
import { getMenu } from "../api/menu";
import type { MenuResponse } from "../types/menu";

const MENU_CACHE_KEY = "tablebite-public-menu";
const MENU_CACHE_TTL = 5 * 60 * 1000;
const MENU_CACHE_UPDATED_EVENT = "tablebite-public-menu-updated";

interface MenuCache {
  savedAt: number;
  data: MenuResponse;
}

function readMenuCache(): MenuCache | null {
  if (typeof window === "undefined") return null;

  try {
    const cached = JSON.parse(window.localStorage.getItem(MENU_CACHE_KEY) ?? "null") as MenuCache | null;
    if (!cached?.data || typeof cached.savedAt !== "number") return null;
    return cached;
  } catch {
    return null;
  }
}

let menuRequest: Promise<MenuResponse> | null = null;

export function invalidateMenuCache() {
  try {
    window.localStorage.removeItem(MENU_CACHE_KEY);
    window.dispatchEvent(new CustomEvent(MENU_CACHE_UPDATED_EVENT));
  } catch {
    // Cache invalidation is best effort and must not fail an admin mutation.
  }
}

function fetchMenu() {
  if (!menuRequest) {
    menuRequest = getMenu().finally(() => {
      menuRequest = null;
    });
  }
  return menuRequest;
}

export function useMenu() {
  const cached = readMenuCache();
  const [data, setData] = useState<MenuResponse | null>(cached?.data ?? null);
  const [loading, setLoading] = useState(cached === null);
  const [error, setError] = useState<string | null>(null);

  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    if (attempt === 0 && cached && Date.now() - cached.savedAt < MENU_CACHE_TTL) {
      return;
    }

    setLoading(!data);
    setError(null);
    fetchMenu()
      .then((freshData) => {
        setData(freshData);
        try {
          window.localStorage.setItem(MENU_CACHE_KEY, JSON.stringify({ savedAt: Date.now(), data: freshData } satisfies MenuCache));
        } catch {
          // Continue using the in-memory response when storage is unavailable.
        }
      })
      .catch((cause: unknown) => {
        if (!data) {
          setError(cause instanceof Error ? cause.message : "Unable to load menu");
        }
      })
      .finally(() => setLoading(false));
  }, [attempt]);

  return { data, loading, error, retry: () => setAttempt((value) => value + 1) };
}
