import { useEffect, useState } from "react";
import { FallbackImage } from "../common/FallbackImage";
import { useCart } from "../../context/CartContext";
import { useRestaurantSettings } from "../../hooks/useRestaurantSettings";
import { getCartVariantLabel } from "../../types/cart";

export function CartPanel() {
  const { items, itemCount, subtotal, activeTable, updateQuantity, removeItem, clearCart } = useCart();
  const { settings } = useRestaurantSettings();
  const restaurantName = settings.restaurant_name?.trim() || "Restaurant";
  const [showNotice, setShowNotice] = useState(false);
  const [showCartSummary, setShowCartSummary] = useState(false);

  useEffect(() => {
    if (!showNotice) {
      return;
    }

    const timer = window.setTimeout(() => {
      setShowNotice(false);
    }, 5000);

    return () => window.clearTimeout(timer);
  }, [showNotice]);

  return (
    <section id="cart" className="mb-8 rounded-[28px] border border-black/5 bg-white p-5 shadow-sm">
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <div>
          <p className="text-sm font-bold uppercase tracking-[0.2em] text-orange-600">Your cart</p>
          <h2 className="text-2xl font-black">{itemCount > 0 ? `${itemCount} item${itemCount > 1 ? "s" : ""} ready` : "Your cart is empty"}</h2>
        </div>
        {items.length > 0 ? (
          <button type="button" onClick={clearCart} className="text-sm font-semibold text-slate-500 transition hover:text-orange-600">
            Clear all
          </button>
        ) : null}
      </div>

      {items.length === 0 ? (
        <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 p-5 text-sm text-slate-600">
          Add something delicious from our 
          <a href="/menu" className="ml-1 font-semibold text-blue-600 underline underline-offset-2 transition hover:text-blue-700">
            menu
          </a>
        </div>
      ) : (
        <>
          <div className="space-y-3">
            {items.map((item) => (
              <div key={`${item.menuItemId}-${item.variant}`} className="flex flex-col gap-3 rounded-2xl border border-slate-200 bg-slate-50 p-3 sm:flex-row sm:items-center">
                <FallbackImage src={item.imageUrl} alt={item.name} className="h-16 w-16 rounded-2xl object-cover" />
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-start justify-between gap-2">
                    <div>
                      <p className="font-extrabold text-slate-900">{item.name}</p>
                      <p className="text-sm text-slate-500">{getCartVariantLabel(item.variant)}</p>
                    </div>
                    <p className="font-black text-slate-900">₹{item.unitPrice * item.quantity}</p>
                  </div>
                  <div className="mt-3 flex flex-wrap items-center justify-between gap-3">
                    <div className="flex items-center rounded-full border border-slate-200 bg-white p-1">
                      <button type="button" onClick={() => updateQuantity(item.menuItemId, item.variant, item.quantity - 1)} className="h-8 w-8 rounded-full text-lg font-bold transition hover:bg-slate-100">−</button>
                      <span className="min-w-8 text-center text-sm font-bold">{item.quantity}</span>
                      <button type="button" onClick={() => updateQuantity(item.menuItemId, item.variant, item.quantity + 1)} className="h-8 w-8 rounded-full text-lg font-bold transition hover:bg-slate-100">+</button>
                    </div>
                    <button type="button" onClick={() => removeItem(item.menuItemId, item.variant)} className="text-sm font-semibold text-rose-600 transition hover:text-rose-700">
                      Remove
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
          <div className="mt-5 rounded-2xl bg-slate-950 px-4 py-4 text-white">
            <div className="flex items-center justify-between">
              <span className="text-sm font-semibold uppercase tracking-[0.2em] text-slate-300">Subtotal</span>
              <span className="text-xl font-black">₹{subtotal}</span>
            </div>
            <div className="mt-4 flex flex-col gap-2">
              <button
                type="button"
                onClick={() => setShowCartSummary((value) => !value)}
                className="w-full rounded-full border border-white/20 bg-white/10 px-4 py-3 text-sm font-black uppercase tracking-[0.2em] text-white transition hover:bg-white/20"
              >
                {showCartSummary ? "Hide Cart" : "Show Cart"}
              </button>
              <button
                type="button"
                onClick={() => setShowNotice(true)}
                className="w-full rounded-full bg-orange-600 px-4 py-3 text-sm font-black uppercase tracking-[0.2em] text-white transition hover:bg-orange-500"
              >
                Order
              </button>
            </div>
          </div>
        </>
      )}

      {showCartSummary ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/70 px-4">
          <div className="w-full max-w-md rounded-[28px] border border-white/10 bg-white p-6 shadow-2xl">
            <div className="text-center">
              <p className="text-sm font-black uppercase tracking-[0.2em] text-orange-600">{restaurantName}</p>
              <h3 className="mt-2 text-xl font-black text-slate-900">YOUR CART</h3>
            </div>
            {activeTable ? (
              <div className="mt-4 flex items-center justify-between rounded-2xl border border-orange-100 bg-orange-50 px-4 py-3 text-sm text-orange-700">
                <div className="font-semibold">Table Details</div>
                <div className="ml-3 text-right font-black text-orange-800">
                  {activeTable.tableName || `Table ${activeTable.tableNumber}`} · #{activeTable.tableNumber}
                </div>
              </div>
            ) : null}
            <div className="mt-5 space-y-3 text-sm text-slate-700">
              {items.map((item) => (
                <div key={`summary-${item.menuItemId}-${item.variant}`} className="flex items-center justify-between gap-3">
                  <span className="flex-1 text-left">
                    {item.name}
                    {item.variant !== "regular" ? ` — ${getCartVariantLabel(item.variant)}` : ""}
                  </span>
                  <span className="font-semibold">×{item.quantity}</span>
                </div>
              ))}
            </div>
            <div className="mt-5 border-t border-slate-200 pt-4 text-sm text-slate-700">
              <div className="flex items-center justify-between font-black">
                <span>TOTAL</span>
                <span>₹{subtotal}</span>
              </div>
            </div>
            <button
              type="button"
              onClick={() => setShowCartSummary(false)}
              className="mt-6 w-full rounded-full bg-slate-950 px-5 py-2.5 text-sm font-bold text-white"
            >
              Close
            </button>
          </div>
        </div>
      ) : null}

      {showNotice ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/70 px-4">
          <div className="w-full max-w-md rounded-[28px] border border-white/10 bg-white p-6 text-center shadow-2xl">
            <p className="text-lg font-black uppercase tracking-[0.2em] text-orange-600">Coming Soon</p>
            <div className="mt-4 space-y-2 text-sm leading-6 text-slate-600">
              <p>Online ordering support is coming soon.</p>
              <p>Please call a waiter and show them</p>
              <p>your cart. Thanks!</p>
              <p className="mt-3 font-semibold text-slate-800">
                Your cart: {itemCount} item{itemCount > 1 ? "s" : ""} · ₹{subtotal}
              </p>
            </div>
            <button
              type="button"
              onClick={() => setShowNotice(false)}
              className="mt-6 rounded-full bg-slate-950 px-5 py-2.5 text-sm font-bold text-white"
            >
              Continue Browsing
            </button>
          </div>
        </div>
      ) : null}
    </section>
  );
}
