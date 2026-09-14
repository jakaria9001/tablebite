import { useState } from "react";
import { FallbackImage } from "../common/FallbackImage";
import { useCart } from "../../context/CartContext";
import { getAvailableCartVariants, getCartVariantLabel, type CartVariant } from "../../types/cart";
import type { MenuItem } from "../../types/menu";
import { MenuBadge } from "./MenuBadge";
import { PriceDisplay } from "./PriceDisplay";
import { VegIndicator } from "./VegIndicator";

interface Props {
  item: MenuItem;
}

export function MenuCard({ item }: Props) {
  const { addItem, items, updateQuantity } = useCart();
  const variants = getAvailableCartVariants(item);
  const [selectedVariant, setSelectedVariant] = useState<CartVariant>(() => variants.includes("regular") ? "regular" : variants[0] ?? "regular");

  const cartEntry = items.find((entry) => entry.menuItemId === item.id && entry.variant === selectedVariant);
  const quantity = cartEntry?.quantity ?? 0;

  function rememberItem() {
    try {
      const stored = JSON.parse(window.localStorage.getItem("tablebite-recent-items") ?? "[]") as number[];
      const recent = [item.id, ...stored.filter((id) => id !== item.id)].slice(0, 8);
      window.localStorage.setItem("tablebite-recent-items", JSON.stringify(recent));
    } catch {
      // Local storage can be unavailable in private browsing or embedded webviews.
    }
  }

  const badges = [
    item.is_bestseller ? { label: "★ Bestseller", tone: "accent" as const } : null,
    item.is_featured ? { label: "⭐ Chef's Pick", tone: "success" as const } : null,
    item.is_top10 ? { label: "# Top 10", tone: "default" as const } : null,
    !item.is_available ? { label: "Sold Out", tone: "default" as const } : null,
  ].filter((badge): badge is { label: string; tone: "default" | "accent" | "success" } => badge !== null);

  return (
    <article onClick={rememberItem} className={`group overflow-hidden rounded-3xl border border-[var(--line)] bg-white shadow-[var(--shadow-soft)] transition duration-200 hover:-translate-y-1 hover:shadow-[var(--shadow-lifted)] ${!item.is_available ? "opacity-70" : ""}`}>
      <div className="relative overflow-hidden">
        {item.image_url ? (
          <FallbackImage
            src={item.image_url}
            alt={item.name}
            loading="lazy"
            className={`aspect-[4/3] w-full object-cover transition duration-500 group-hover:scale-105 ${!item.is_available ? "grayscale" : ""}`}
          />
        ) : (
          <FallbackImage
            src="/indian_restaurant_logo.jpg"
            alt={item.name}
            loading="lazy"
            className={`aspect-[4/3] w-full object-cover transition duration-500 group-hover:scale-105 ${!item.is_available ? "grayscale" : ""}`}
          />
        )}
        {badges.length > 0 ? (
          <div className="absolute left-3 top-3 flex flex-wrap gap-2">
            {badges.map((badge) => (
              <MenuBadge key={badge.label} label={badge.label} tone={badge.tone} />
            ))}
          </div>
        ) : null}
        {!item.is_available && (
          <div className="absolute inset-0 grid place-items-center bg-black/35">
            <span className="rounded-xl bg-white px-4 py-2 text-sm font-bold">Sold Out</span>
          </div>
        )}
      </div>
      <div className="p-5">
        <div className="mb-2 flex items-start justify-between gap-3">
          <h3 className="font-extrabold leading-tight">{item.name}</h3>
          <VegIndicator isVeg={item.is_veg} />
        </div>
        <p className="min-h-12 text-sm leading-5 text-[var(--muted)]">{item.description}</p>
        <div className="mt-4 flex flex-col gap-3">
          {variants.length > 1 ? (
            <div className="flex flex-wrap gap-2">
              {variants.map((variant) => (
                <button
                  key={variant}
                  type="button"
                  onClick={() => setSelectedVariant(variant)}
                  className={`rounded-full px-3 py-1.5 text-xs font-bold uppercase tracking-[0.2em] ${selectedVariant === variant ? "bg-slate-950 text-white" : "bg-slate-100 text-slate-700"}`}
                >
                  {getCartVariantLabel(variant)}
                </button>
              ))}
            </div>
          ) : null}
          <div className="flex items-center justify-between gap-4">
            <strong className="text-lg"><PriceDisplay item={item} variant={selectedVariant} /></strong>
            {item.is_available ? (
              quantity > 0 ? (
                <div className="flex items-center rounded-full border border-slate-200 bg-white p-1 shadow-sm">
                  <button
                    type="button"
                    onClick={() => updateQuantity(item.id, selectedVariant, quantity - 1)}
                    className="h-9 w-9 rounded-full text-lg font-bold text-slate-700 transition hover:bg-slate-100"
                  >
                    −
                  </button>
                  <span className="min-w-8 text-center text-sm font-black text-slate-900">{quantity}</span>
                  <button
                    type="button"
                    onClick={() => addItem(item, selectedVariant)}
                    className="h-9 w-9 rounded-full text-lg font-bold text-slate-700 transition hover:bg-slate-100"
                  >
                    +
                  </button>
                </div>
              ) : (
                <button
                  type="button"
                  onClick={() => addItem(item, selectedVariant)}
                  className="min-h-11 rounded-full bg-[var(--maroon-800)] px-5 py-2 text-sm font-bold text-white transition hover:bg-[var(--maroon-700)] active:scale-95"
                >
                  Add
                </button>
              )
            ) : (
              <button
                type="button"
                disabled
                className="rounded-full bg-orange-200 px-4 py-2 text-sm font-bold text-orange-900 disabled:cursor-not-allowed"
              >
                Sold Out
              </button>
            )}
          </div>
        </div>
      </div>
    </article>
  );
}
