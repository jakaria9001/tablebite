import { useCallback, useEffect, useState } from "react";
import { getAdminTables } from "../api/tables";
import type { AdminTable } from "../api/tables";

export function useAdminTables() {
  const [tables, setTables] = useState<AdminTable[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await getAdminTables();
      setTables(data);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Unable to load tables");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    const timer = window.setTimeout(() => void refresh(), 0);
    return () => window.clearTimeout(timer);
  }, [refresh]);

  return { tables, loading, error, refresh };
}
