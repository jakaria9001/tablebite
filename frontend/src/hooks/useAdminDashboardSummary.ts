import { useEffect, useState } from "react";
import { getAdminDashboardSummary } from "../api/admin";
import type { AdminDashboardSummary } from "../types/admin";
import { invalidateMenuCache } from "./useMenu";

const SUMMARY_CACHE_KEY = "tablebite-admin-dashboard-summary";
const SUMMARY_VERSION_KEY = "tablebite-admin-data-version";
const SUMMARY_UPDATED_EVENT = "tablebite-admin-data-updated";

interface SummaryCache {
  version: string;
  data: AdminDashboardSummary;
}

function readSummaryCache(): SummaryCache | null {
  try {
    const cached = JSON.parse(localStorage.getItem(SUMMARY_CACHE_KEY) ?? "null") as SummaryCache | null;
    if (!cached?.data || typeof cached.version !== "string") return null;
    return cached;
  } catch {
    return null;
  }
}

function readCurrentVersion() {
  return localStorage.getItem(SUMMARY_VERSION_KEY) ?? "0";
}

export function markAdminDataUpdated() {
  const version = new Date().toISOString();
  try {
    localStorage.setItem(SUMMARY_VERSION_KEY, version);
    window.dispatchEvent(new CustomEvent(SUMMARY_UPDATED_EVENT));
  } catch {
    // Dashboard caching is best effort and must not mask a successful mutation.
  }
  invalidateMenuCache();
}

export function useAdminDashboardSummary() {
  const cached = typeof window === "undefined" ? null : readSummaryCache();
  const hasFreshCache = Boolean(cached && cached.version === readCurrentVersion());
  const [data, setData] = useState<AdminDashboardSummary | null>(cached?.data ?? null);
  const [loading, setLoading] = useState(!hasFreshCache);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    const load = async () => {
      try {
        const freshData = await getAdminDashboardSummary();
        if (cancelled) return;
        const version = readCurrentVersion();
        localStorage.setItem(SUMMARY_CACHE_KEY, JSON.stringify({ version, data: freshData } satisfies SummaryCache));
        setData(freshData);
        setError(null);
      } catch (cause) {
        if (!cancelled) setError(cause instanceof Error ? cause.message : "Unable to load dashboard summary");
      } finally {
        if (!cancelled) setLoading(false);
      }
    };

    if (!hasFreshCache) void load();
    const handleUpdate = () => {
      setLoading(false);
      void load();
    };
    const handleStorageUpdate = (event: StorageEvent) => {
      if (event.key === SUMMARY_VERSION_KEY) handleUpdate();
    };
    window.addEventListener(SUMMARY_UPDATED_EVENT, handleUpdate);
    window.addEventListener("storage", handleStorageUpdate);
    return () => {
      cancelled = true;
      window.removeEventListener(SUMMARY_UPDATED_EVENT, handleUpdate);
      window.removeEventListener("storage", handleStorageUpdate);
    };
  }, [hasFreshCache]);

  return { data, loading, error };
}
