import { useEffect, useRef, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { ApiError } from "../api/client";
import { getOrder, type Order } from "../api/orders";
import { Footer } from "../components/layout/Footer";
import { Header } from "../components/layout/Header";
import { useCart } from "../hooks/useCart";
import { getCartVariantLabel } from "../types/cart";

export default function OrderDetailsPage() {
  const { orderId } = useParams();
  const { activeTable } = useCart();
  const [order, setOrder] = useState<Order | null>(null);
  const [error, setError] = useState<string | null>(null);
  const parsedOrderId = Number(orderId);
  const invalidOrderId = !Number.isInteger(parsedOrderId) || parsedOrderId <= 0;
  const statusRef = useRef<string | null>(null);
  const statuses = ["new", "accepted", "preparing", "ready", "served", "completed"];
  const isTerminal = (status: string | null) => status === "completed" || status === "cancelled";

  useEffect(() => {
    if (invalidOrderId) {
      return;
    }

    let cancelled = false;
    const loadOrder = async () => {
      try {
        const result = await getOrder(parsedOrderId);
        if (!cancelled) {
          statusRef.current = result.status;
          setOrder(result);
          setError(null);
        }
      } catch (cause) {
        if (!cancelled) {
          setError(cause instanceof ApiError ? cause.message : "Unable to load this order.");
        }
      }
    };

    void loadOrder();
    const interval = window.setInterval(() => {
      if (!isTerminal(statusRef.current)) {
        void loadOrder();
      }
    }, 10000);
    return () => {
      cancelled = true;
      window.clearInterval(interval);
    };
  }, [invalidOrderId, parsedOrderId]);

  return (
    <div className="min-h-screen bg-[#fffaf5]">
      <Header />
      <main className="mx-auto max-w-2xl px-5 py-12 pb-20 lg:px-8">
        {invalidOrderId || error ? (
          <section className="rounded-[28px] border border-rose-100 bg-white p-8 text-center shadow-sm">
            <p className="font-semibold text-rose-700">{error ?? "Order not found."}</p>
            <Link to="/menu" className="mt-6 inline-flex rounded-full bg-slate-950 px-5 py-3 text-sm font-bold text-white">Back to menu</Link>
          </section>
        ) : order ? (
          <>
            <p className="text-sm font-bold uppercase tracking-[0.2em] text-orange-600">Order confirmed</p>
            <h1 className="mt-2 text-4xl font-black text-slate-950">Thank you!</h1>
            <section className="mt-6 rounded-[28px] bg-slate-950 p-6 text-white shadow-xl">
              <p className="text-sm font-bold uppercase tracking-[0.18em] text-orange-300">Order number</p>
              <p className="mt-2 text-3xl font-black">{order.orderNumber}</p>
              <p className="mt-4 text-sm text-slate-300">
                {activeTable?.tableName || `Table ${activeTable?.tableNumber ?? order.tableId}`}
              </p>
              <p className="mt-2 inline-flex rounded-full bg-orange-500/20 px-3 py-1 text-sm font-bold capitalize text-orange-200">{order.status}</p>
            </section>
            <section className="mt-6 rounded-[28px] border border-slate-200 bg-white p-5 shadow-sm">
              <div className="flex items-center justify-between gap-3">
                <h2 className="text-xl font-black text-slate-950">Order status</h2>
                {!isTerminal(order.status) ? <span className="text-xs font-bold uppercase tracking-[0.16em] text-slate-400">Updates automatically</span> : null}
              </div>
              {order.status === "cancelled" ? (
                <p className="mt-4 rounded-2xl bg-rose-50 px-4 py-3 font-bold text-rose-700">This order was cancelled.</p>
              ) : (
                <div className="mt-5 grid grid-cols-3 gap-3 sm:grid-cols-6">
                  {statuses.map((status, index) => {
                    const currentIndex = statuses.indexOf(order.status);
                    const reached = currentIndex >= index;
                    return (
                      <div key={status} className="text-center">
                        <div className={`mx-auto flex h-9 w-9 items-center justify-center rounded-full text-sm font-black ${reached ? "bg-orange-600 text-white" : "bg-slate-100 text-slate-400"}`}>
                          {index + 1}
                        </div>
                        <p className={`mt-2 text-[10px] font-black uppercase tracking-wide ${reached ? "text-orange-700" : "text-slate-400"}`}>{status}</p>
                      </div>
                    );
                  })}
                </div>
              )}
              {order.statusHistory.length > 0 ? (
                <div className="mt-6 border-t border-slate-100 pt-4">
                  <h3 className="text-sm font-black uppercase tracking-[0.16em] text-slate-500">Updates</h3>
                  <div className="mt-3 space-y-3">
                    {order.statusHistory.map((entry) => (
                      <div key={`${entry.newStatus}-${entry.changedAt}`} className="flex items-start justify-between gap-4 text-sm">
                        <div>
                          <p className="font-bold uppercase text-slate-800">{entry.newStatus}</p>
                          {entry.note ? <p className="text-slate-500">{entry.note}</p> : null}
                        </div>
                        <time className="shrink-0 text-xs text-slate-400">{new Date(entry.changedAt).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}</time>
                      </div>
                    ))}
                  </div>
                </div>
              ) : null}
            </section>
            <section className="mt-6 rounded-[28px] border border-slate-200 bg-white p-5 shadow-sm">
              <h2 className="text-xl font-black text-slate-950">Your items</h2>
              <div className="mt-4 space-y-3">
                {order.items.map((item) => (
                  <div key={item.id} className="flex items-center justify-between gap-4 border-b border-slate-100 pb-3 last:border-0 last:pb-0">
                    <div>
                      <p className="font-bold text-slate-900">{item.itemName}</p>
                      <p className="text-sm text-slate-500">{getCartVariantLabel(item.variant)} · Quantity {item.quantity}</p>
                    </div>
                    <p className="font-black text-slate-900">₹{item.lineTotal}</p>
                  </div>
                ))}
              </div>
              <div className="mt-5 flex items-center justify-between border-t border-slate-200 pt-4 text-lg font-black">
                <span>Total</span>
                <span>₹{order.subtotal}</span>
              </div>
            </section>
            <Link to="/menu" className="mt-6 flex w-full justify-center rounded-full bg-orange-600 px-5 py-4 text-sm font-black uppercase tracking-[0.2em] text-white">Continue browsing</Link>
          </>
        ) : (
          <div className="rounded-[28px] border border-slate-200 bg-white p-8 text-center font-semibold text-slate-600">Loading your order...</div>
        )}
      </main>
      <Footer />
    </div>
  );
}
