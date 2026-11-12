import { apiFetch } from "./client";
import type { Order } from "./orders";

export async function getStaffOrders(): Promise<Order[]> {
  const response = await apiFetch<{ orders: Order[] }>("/api/v1/staff/orders", {
    method: "GET",
    headers: { Accept: "application/json" },
    includeAdminAuth: false,
  });
  return response.orders;
}

export async function updateStaffOrderStatus(orderId: number, status: string, note?: string): Promise<Order> {
  const response = await apiFetch<{ order: Order }>(`/api/v1/staff/orders/${orderId}/status`, {
    method: "POST",
    body: JSON.stringify({ status, note }),
    includeAdminAuth: false,
  });
  return response.order;
}

export async function provisionWaiter(phone: string, name: string) {
  const response = await apiFetch<{ waiter: { id: number; customerId: number; name: string; phone: string } }>("/api/v1/admin/waiters", {
    method: "POST",
    body: JSON.stringify({ phone, name }),
  });
  return response.waiter;
}
