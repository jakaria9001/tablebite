import { apiFetch } from "./client";
import type { CartVariant } from "../types/cart";

export interface OrderItemInput {
  menuItemId: number;
  variant: CartVariant;
  quantity: number;
}

export interface OrderItem {
  id: number;
  menuItemId: number;
  itemName: string;
  variant: CartVariant;
  quantity: number;
  unitPrice: number;
  lineTotal: number;
}

export interface OrderStatusHistory {
  previousStatus?: string;
  newStatus: string;
  changedAt: string;
  note?: string;
}

export interface Order {
  id: number;
  orderNumber: string;
  customerId: number;
  tableId: number;
  tableSessionId: number;
  tableNumber?: string;
  status: string;
  subtotal: number;
  currency: string;
  customerNote?: string;
  createdAt: string;
  updatedAt: string;
  items: OrderItem[];
  statusHistory: OrderStatusHistory[];
}

export async function createOrder(input: {
  tableSessionId: number;
  customerNote?: string;
  items: OrderItemInput[];
}, idempotencyKey: string): Promise<Order> {
  const response = await apiFetch<{ order: Order }>("/api/v1/orders", {
    method: "POST",
    headers: {
      "Idempotency-Key": idempotencyKey,
    },
    body: JSON.stringify(input),
    includeAdminAuth: false,
  });
  return response.order;
}

export async function getOrder(orderId: number): Promise<Order> {
  const response = await apiFetch<{ order: Order }>(`/api/v1/orders/${orderId}`, {
    method: "GET",
    headers: { Accept: "application/json" },
    includeAdminAuth: false,
  });
  return response.order;
}
