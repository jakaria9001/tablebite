import { useEffect, useState } from "react";
import { getAdminCategories } from "../api/adminCategories";
import type { AdminCategory } from "../types/adminCategories";

export function useAdminCategories() {
  const [categories, setCategories] = useState<AdminCategory[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refresh = () => {
    setLoading(true);
    getAdminCategories()
      .then((data) => setCategories(data.categories))
      .catch((cause: unknown) => {
        setError(cause instanceof Error ? cause.message : "Unable to load categories");
      })
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    refresh();
  }, []);

  return { categories, loading, error, refresh };
}
