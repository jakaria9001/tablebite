import { apiFetch, apiGet } from "./client";
import type { AdminMenuData, AdminMenuItem, AdminMenuItemPayload } from "../types/adminMenu";

export function getAdminMenuData(): Promise<AdminMenuData> {
  return apiGet<AdminMenuData>("/api/v1/admin/menu");
}

export function createAdminMenuItem(payload: AdminMenuItemPayload): Promise<AdminMenuItem> {
  return apiFetch<AdminMenuItem>("/api/v1/admin/menu", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function updateAdminMenuItem(id: number, payload: AdminMenuItemPayload): Promise<AdminMenuItem> {
  return apiFetch<AdminMenuItem>(`/api/v1/admin/menu/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload),
  });
}

export function deleteAdminMenuItem(id: number): Promise<void> {
  return apiFetch<void>(`/api/v1/admin/menu/${id}`, {
    method: "DELETE",
  });
}
