import { apiFetch } from "./client";

export interface AdminTable {
  id: number;
  table_number: string;
  display_name: string;
  qr_token: string;
  is_active: boolean;
  display_order: number;
}

export interface AdminTablePayload {
  id?: number;
  table_number: string;
  display_name: string;
  qr_token: string;
  is_active: boolean;
  display_order?: number;
}

export async function getAdminTables(): Promise<AdminTable[]> {
  const data = await apiFetch<{ tables: AdminTable[] }>('/api/v1/admin/tables');
  return data.tables ?? [];
}

export async function createAdminTable(payload: AdminTablePayload): Promise<AdminTable> {
  return apiFetch<AdminTable>('/api/v1/admin/tables', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export async function updateAdminTable(id: number, payload: AdminTablePayload): Promise<AdminTable> {
  return apiFetch<AdminTable>(`/api/v1/admin/tables/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  });
}

export async function deleteAdminTable(id: number): Promise<void> {
  await apiFetch<void>(`/api/v1/admin/tables/${id}`, { method: 'DELETE' });
}

export async function reorderAdminTables(tableIds: number[]): Promise<void> {
  await apiFetch<void>('/api/v1/admin/tables/reorder', {
    method: 'POST',
    body: JSON.stringify(tableIds),
  });
}
