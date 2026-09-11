import { apiFetch, apiGet } from "./client";
import type { AdminCategory, AdminCategoryPayload } from "../types/adminCategories";

export function getAdminCategories(): Promise<{ categories: AdminCategory[] }> {
  return apiGet<{ categories: AdminCategory[] }>("/api/v1/admin/categories");
}

export function createAdminCategory(payload: AdminCategoryPayload): Promise<AdminCategory> {
  return apiFetch<AdminCategory>("/api/v1/admin/categories", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function updateAdminCategory(id: number, payload: AdminCategoryPayload): Promise<AdminCategory> {
  return apiFetch<AdminCategory>(`/api/v1/admin/categories/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload),
  });
}

export function reorderAdminCategories(categoryIds: number[]): Promise<void> {
  return apiFetch<void>("/api/v1/admin/categories/reorder", {
    method: "POST",
    body: JSON.stringify(categoryIds),
  });
}
