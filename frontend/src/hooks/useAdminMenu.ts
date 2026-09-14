import { useEffect, useState } from "react";
import { getAdminMenuData } from "../api/adminMenu";
import type { AdminMenuData } from "../types/adminMenu";

export function useAdminMenu() {
  const [data, setData] = useState<AdminMenuData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refresh = () => {
    setLoading(true);
    getAdminMenuData()
      .then(setData)
      .catch((cause: unknown) => {
        setError(cause instanceof Error ? cause.message : "Unable to load menu data");
      })
      .finally(() => setLoading(false));
  };

  const updateItem = (updatedItem: AdminMenuData["items"][number]) => {
    setData((current) => current ? {
      ...current,
      items: current.items.map((item) => item.id === updatedItem.id ? updatedItem : item),
    } : current);
  };

  useEffect(() => {
    refresh();
  }, []);

  return { data, loading, error, refresh, updateItem };
}
