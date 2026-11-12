import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { ApiError } from "../api/client";
import { createOrder } from "../api/orders";
import { Footer } from "../components/layout/Footer";
import { Header } from "../components/layout/Header";
import { useAuth } from "../hooks/useAuth";
import { useCart } from "../hooks/useCart";
import { useRestaurantSettings } from "../hooks/useRestaurantSettings";
import { getCartVariantLabel } from "../types/cart";

function createIdempotencyKey() {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }
  return `${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

export default function OrderReviewPage() {
  const { items, subtotal, activeTable, clearCart } = useCart();
  const { isAuthenticated } = useAuth();
  const { settings } = useRestaurantSettings();
  const navigate = useNavigate();
  const [customerNote, setCustomerNote] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [placing, setPlacing] = useState(false);
  const [idempotencyKey, setIdempotencyKey] = useState<string | null>(null);
  const restaurantName = settings.restaurant_name?.trim() || "Restaurant";

  async function handlePlaceOrder() {
    if (!isAuthenticated) {
      navigate("/cart");
      return;
    }
    if (!activeTable?.sessionId) {
      setError("Scan a valid table QR code before placing an order.");
      return;
    }
    if (items.length === 0) {
      setError("Your cart is empty.");
      return;
    }

    setError(null);
    setPlacing(true);
    const submissionKey = idempotencyKey ?? createIdempotencyKey();
    setIdempotencyKey(submissionKey);
    try {
      const order = await createOrder({
        tableSessionId: activeTable.sessionId,
        customerNote: customerNote.trim() || undefined,
        items: items.map((item) => ({
          menuItemId: item.menuItemId,
          variant: item.variant,
          quantity: item.quantity,
        })),
      }, submissionKey);
      clearCart();
      navigate(`/orders/${order.id}`);
    } catch (cause) {
      setError(cause instanceof ApiError ? cause.message : "We could not place your order. Please try again.");
    } finally {
      setPlacing(false);
    }
  }

  return (
    <div className="min-h-screen bg-[#fffaf5]">
      <Header />
      <main className="mx-auto max-w-3xl px-5 py-10 pb-20 lg:px-8">
        <p className="text-sm font-bold uppercase tracking-[0.2em] text-orange-600">Order review</p>
        <h1 className="mt-2 text-4xl font-black text-slate-950">Ready to place your order?</h1>
        <div className="mt-6 rounded-[28px] border border-orange-100 bg-orange-50 p-5">
          <p className="text-sm font-bold uppercase tracking-[0.18em] text-orange-700">{restaurantName}</p>
          <p className="mt-2 text-lg font-black text-orange-950">
            {activeTable ? `${activeTable.tableName || `Table ${activeTable.tableNumber}`} · Table ${activeTable.tableNumber}` : "No active table"}
          </p>
        </div>

        <section className="mt-6 rounded-[28px] border border-slate-200 bg-white p-5 shadow-sm">
          <h2 className="text-xl font-black text-slate-950">Items</h2>
          <div className="mt-4 space-y-3">
            {items.map((item) => (
              <div key={`${item.menuItemId}-${item.variant}`} className="flex items-center justify-between gap-4 border-b border-slate-100 pb-3 last:border-0 last:pb-0">
                <div>
                  <p className="font-bold text-slate-900">{item.name}</p>
                  <p className="text-sm text-slate-500">{getCartVariantLabel(item.variant)} · Quantity {item.quantity}</p>
                </div>
                <p className="font-black text-slate-900">₹{item.unitPrice * item.quantity}</p>
              </div>
            ))}
          </div>
          <div className="mt-5 flex items-center justify-between border-t border-slate-200 pt-4 text-lg font-black">
            <span>Subtotal</span>
            <span>₹{subtotal}</span>
          </div>
        </section>

        <label className="mt-6 block rounded-[28px] border border-slate-200 bg-white p-5 shadow-sm">
          <span className="font-black text-slate-950">Special instructions</span>
          <textarea
            value={customerNote}
            onChange={(event) => setCustomerNote(event.target.value)}
            maxLength={500}
            rows={4}
            placeholder="Anything the kitchen or waiter should know?"
            className="mt-3 w-full rounded-2xl border border-slate-200 px-4 py-3 text-sm outline-none ring-orange-200 focus:ring-2"
          />
        </label>

        {error ? <p role="alert" className="mt-4 rounded-2xl bg-rose-50 px-4 py-3 text-sm font-semibold text-rose-700">{error}</p> : null}
        <button
          type="button"
          disabled={placing || items.length === 0}
          onClick={() => void handlePlaceOrder()}
          className="mt-6 w-full rounded-full bg-orange-600 px-5 py-4 text-sm font-black uppercase tracking-[0.2em] text-white transition hover:bg-orange-500 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {placing ? "Placing order..." : "Place order"}
        </button>
        <button type="button" onClick={() => navigate("/cart")} className="mt-3 w-full rounded-full border border-slate-300 px-5 py-3 text-sm font-bold text-slate-700">
          Back to cart
        </button>
      </main>
      <Footer />
    </div>
  );
}
