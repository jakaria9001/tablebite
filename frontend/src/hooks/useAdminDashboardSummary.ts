import { useEffect, useState } from "react";
import { getAdminDashboardSummary } from "../api/admin";
import type { AdminDashboardSummary } from "../types/admin";

export function useAdminDashboardSummary() {
  const [data, setData] = useState<AdminDashboardSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getAdminDashboardSummary()
      .then(setData)
      .catch((cause: unknown) => {
        setError(cause instanceof Error ? cause.message : "Unable to load dashboard summary");
      })
      .finally(() => setLoading(false));
  }, []);

  return { data, loading, error };
}
