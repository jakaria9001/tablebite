import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { ApiError } from "../api/client";
import { logoutWaiter } from "../api/auth";
import { getStaffOrders, updateStaffOrderStatus } from "../api/staffOrders";
import type { Order } from "../api/orders";
import { getCartVariantLabel } from "../types/cart";

const activeStatuses = ["new", "accepted", "preparing", "ready", "served"];
const pollIntervalMs = Number(import.meta.env.VITE_WAITER_POLL_INTERVAL_MS ?? 5000);

function nextAction(order: Order) {
  if (order.status === "new") return { label: "Accept order", status: "accepted" };
  if (order.status === "accepted") return { label: "Start preparing", status: "preparing" };
  if (order.status === "preparing") return { label: "Mark ready", status: "ready" };
  if (order.status === "ready") return { label: "Mark served", status: "served" };
  if (order.status === "served") return { label: "Complete order", status: "completed" };
  return null;
}

export default function WaiterDashboardPage() {
  const navigate = useNavigate();
  const [orders, setOrders] = useState<Order[]>([]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState<number | null>(null);

  const load = useCallback(async () => {
    try {
      setOrders(await getStaffOrders());
      setError("");
    } catch (cause) {
      if (cause instanceof ApiError && cause.status === 401) {
        navigate("/waiter/login", { replace: true });
        return;
      }
      setError(cause instanceof ApiError ? cause.message : "Unable to load orders.");
    } finally {
      setLoading(false);
    }
  }, [navigate]);

  useEffect(() => {
    const initialLoad = window.setTimeout(() => void load(), 0);
    const timer = window.setInterval(() => void load(), Number.isFinite(pollIntervalMs) && pollIntervalMs >= 1000 ? pollIntervalMs : 5000);
    return () => {
      window.clearTimeout(initialLoad);
      window.clearInterval(timer);
    };
  }, [load]);

  async function transition(order: Order, status: string) {
    setBusy(order.id);
    try {
      const updated = await updateStaffOrderStatus(order.id, status);
      setOrders((current) => current.map((item) => item.id === updated.id ? updated : item).filter((item) => activeStatuses.includes(item.status)));
    } catch (cause) {
      setError(cause instanceof ApiError ? cause.message : "Unable to update order.");
    } finally {
      setBusy(null);
    }
  }

  async function logout() {
    await logoutWaiter().catch(() => undefined);
    navigate("/waiter/login", { replace: true });
  }

  return (
    <div className="min-h-screen bg-slate-100 text-slate-950">
      <header className="sticky top-0 z-10 border-b border-slate-200 bg-white px-5 py-4 shadow-sm">
        <div className="mx-auto flex max-w-7xl items-center justify-between gap-4">
          <div>
            <p className="text-xs font-black uppercase tracking-[0.25em] text-orange-600">TableBite</p>
            <h1 className="text-2xl font-black">Waiter orders</h1>
          </div>
          <button type="button" onClick={() => void logout()} className="rounded-full border border-slate-200 px-4 py-2 text-sm font-bold">Sign out</button>
        </div>
      </header>
      <main className="mx-auto max-w-7xl px-5 py-6">
        {error ? <p role="alert" className="mb-5 rounded-2xl bg-rose-50 px-4 py-3 font-semibold text-rose-700">{error}</p> : null}
        {loading ? <p className="rounded-2xl bg-white p-6 font-semibold">Loading orders...</p> : null}
        {!loading && orders.length === 0 ? <p className="rounded-2xl bg-white p-10 text-center text-lg font-bold text-slate-500">No active orders right now.</p> : null}
        <div className="grid gap-5 lg:grid-cols-2">
          {orders.map((order) => {
            const action = nextAction(order);
            return (
              <article key={order.id} className="rounded-[2rem] border border-slate-200 bg-white p-6 shadow-sm">
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <p className="text-xs font-black uppercase tracking-[0.2em] text-orange-600">{order.status}</p>
                    <h2 className="mt-1 text-2xl font-black">{order.orderNumber}</h2>
                    <p className="mt-1 text-sm font-bold text-slate-500">Table {order.tableNumber || order.tableId} · {new Date(order.createdAt).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}</p>
                  </div>
                  <p className="text-2xl font-black">₹{order.subtotal}</p>
                </div>
                <div className="mt-5 space-y-3">
                  {order.items.map((item) => <div key={item.id} className="flex justify-between gap-4 border-b border-slate-100 pb-3 text-base"><span className="font-bold">{item.itemName} <span className="font-normal text-slate-500">({getCartVariantLabel(item.variant)}) ×{item.quantity}</span></span><span className="font-black">₹{item.lineTotal}</span></div>)}
                </div>
                {order.customerNote ? <p className="mt-4 rounded-2xl bg-amber-50 px-4 py-3 text-sm font-semibold text-amber-900">Note: {order.customerNote}</p> : null}
                {action ? <button type="button" disabled={busy === order.id} onClick={() => void transition(order, action.status)} className="mt-5 min-h-14 w-full rounded-2xl bg-orange-600 px-4 py-3 text-lg font-black text-white disabled:opacity-50">{busy === order.id ? "Updating..." : action.label}</button> : null}
                {order.status === "new" ? <button type="button" disabled={busy === order.id} onClick={() => void transition(order, "cancelled")} className="mt-2 w-full rounded-2xl border border-rose-200 px-4 py-3 font-bold text-rose-700 disabled:opacity-50">Reject order</button> : null}
              </article>
            );
          })}
        </div>
      </main>
    </div>
  );
}
