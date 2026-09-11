import { useEffect, useState } from "react";
import { getRestaurantSettings } from "../api/settings";

export function useRestaurantSettings() {
  const [settings, setSettings] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const load = async () => {
      try {
        const data = await getRestaurantSettings();
        setSettings(data);
      } catch (cause) {
        setError(cause instanceof Error ? cause.message : "Unable to load restaurant settings");
      } finally {
        setLoading(false);
      }
    };

    void load();
  }, []);

  return { settings, loading, error };
}
