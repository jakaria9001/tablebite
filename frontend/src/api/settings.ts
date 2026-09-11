import { apiFetch, apiGet } from "./client";

export interface RestaurantSettings {
  [key: string]: string;
}

export function getRestaurantSettings(): Promise<RestaurantSettings> {
  return apiGet<RestaurantSettings>("/api/v1/settings");
}

export function getAdminRestaurantSettings(): Promise<RestaurantSettings> {
  return apiGet<RestaurantSettings>("/api/v1/admin/settings");
}

export function updateAdminRestaurantSettings(payload: RestaurantSettings): Promise<RestaurantSettings> {
  return apiFetch<RestaurantSettings>("/api/v1/admin/settings", {
    method: "PUT",
    body: JSON.stringify(payload),
  });
}
