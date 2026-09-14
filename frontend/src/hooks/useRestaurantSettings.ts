import { useEffect, useState } from "react";
import { getRestaurantSettings } from "../api/settings";

const SETTINGS_STORAGE_KEY = "tablebite-restaurant-settings";

function readCachedSettings(): Record<string, string> {
  if (typeof window === "undefined") {
    return {};
  }

  try {
    const cached = JSON.parse(window.localStorage.getItem(SETTINGS_STORAGE_KEY) ?? "null") as unknown;
    if (!cached || typeof cached !== "object" || Array.isArray(cached)) {
      return {};
    }

    return Object.entries(cached).reduce<Record<string, string>>((settings, [key, value]) => {
      if (typeof value === "string") {
        settings[key] = value;
      }
      return settings;
    }, {});
  } catch {
    return {};
  }
}

export function useRestaurantSettings() {
  const [settings, setSettings] = useState<Record<string, string>>(() => readCachedSettings());
  const [loading, setLoading] = useState(() => Object.keys(readCachedSettings()).length === 0);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const load = async () => {
      try {
        const data = await getRestaurantSettings();
        setSettings(data);
        window.localStorage.setItem(SETTINGS_STORAGE_KEY, JSON.stringify(data));
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
