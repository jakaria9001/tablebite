import { apiGet, apiFetch } from "./client";

export interface TableSession {
  id: number;
  tableId: number;
  tableNumber: string;
  tableName: string;
  qrToken: string;
  openedAt: string;
  closedAt: string | null;
  active: boolean;
}

export async function resolveTableSession(tableToken: string): Promise<TableSession> {
  const response = await apiGet<{ session: TableSession }>(`/api/v1/tables/resolve/${encodeURIComponent(tableToken)}`);
  return response.session;
}

export async function getTableSession(sessionId: number): Promise<TableSession> {
  const response = await apiGet<{ session: TableSession }>(`/api/v1/tables/session/${sessionId}`);
  return response.session;
}

export async function openTableSession(tableId: number): Promise<TableSession> {
  const response = await apiFetch<{ session: TableSession }>(`/api/v1/admin/tables/${tableId}/sessions/open`, {
    method: "POST",
  });
  return response.session;
}

export async function closeTableSession(tableId: number): Promise<TableSession> {
  const response = await apiFetch<{ session: TableSession }>(`/api/v1/admin/tables/${tableId}/sessions/close`, {
    method: "POST",
  });
  return response.session;
}
