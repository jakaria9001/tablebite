import { useEffect, useState } from "react";
import { getMenu } from "../api/menu";
import type { MenuResponse } from "../types/menu";

export function useMenu() {
  const [data, setData] = useState<MenuResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    setLoading(true);
    setError(null);
    getMenu()
      .then(setData)
      .catch((cause: unknown) => {
        setError(cause instanceof Error ? cause.message : "Unable to load menu");
      })
      .finally(() => setLoading(false));
  }, [attempt]);

  return { data, loading, error, retry: () => setAttempt((value) => value + 1) };
}
