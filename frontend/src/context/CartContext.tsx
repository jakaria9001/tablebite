import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import type { MenuItem } from "../types/menu";
import { getAvailableCartVariants, getCartVariantPrice, type CartItem, type CartVariant } from "../types/cart";

interface ActiveTableInfo {
  id: number | null;
  tableNumber: string;
  tableName: string;
  qrToken: string;
}

interface CartContextValue {
  items: CartItem[];
  itemCount: number;
  subtotal: number;
  activeTable: ActiveTableInfo | null;
  addItem: (item: MenuItem, variant?: CartVariant) => void;
  updateQuantity: (menuItemId: number, variant: CartVariant, quantity: number) => void;
  removeItem: (menuItemId: number, variant: CartVariant) => void;
  clearCart: () => void;
  setTableInfo: (table: ActiveTableInfo | null) => void;
}

const STORAGE_KEY = "tablebite-cart";
const TABLE_STORAGE_KEY = "tablebite-active-table";

const CartContext = createContext<CartContextValue | undefined>(undefined);

function readStoredCart(): CartItem[] {
  if (typeof window === "undefined") {
    return [];
  }

  try {
    const stored = window.localStorage.getItem(STORAGE_KEY);
    if (!stored) {
      return [];
    }

    const parsed = JSON.parse(stored) as CartItem[];
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

function writeStoredCart(items: CartItem[]) {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(items));
}

function readStoredTableInfo(): ActiveTableInfo | null {
  if (typeof window === "undefined") {
    return null;
  }

  try {
    const stored = window.localStorage.getItem(TABLE_STORAGE_KEY);
    if (!stored) {
      return null;
    }

    const parsed = JSON.parse(stored) as ActiveTableInfo;
    return parsed && typeof parsed.tableNumber === "string" ? parsed : null;
  } catch {
    return null;
  }
}

function writeStoredTableInfo(table: ActiveTableInfo | null) {
  if (typeof window === "undefined") {
    return;
  }

  if (!table) {
    window.localStorage.removeItem(TABLE_STORAGE_KEY);
    return;
  }

  window.localStorage.setItem(TABLE_STORAGE_KEY, JSON.stringify(table));
}

export function CartProvider({ children }: { children: ReactNode }) {
  const [items, setItems] = useState<CartItem[]>(() => readStoredCart());
  const [activeTable, setActiveTable] = useState<ActiveTableInfo | null>(() => readStoredTableInfo());

  useEffect(() => {
    writeStoredCart(items);
  }, [items]);

  useEffect(() => {
    writeStoredTableInfo(activeTable);
  }, [activeTable]);

  const itemCount = useMemo(() => items.reduce((total, item) => total + item.quantity, 0), [items]);

  const subtotal = useMemo(() => items.reduce((total, item) => total + item.unitPrice * item.quantity, 0), [items]);

  const addItem = (menuItem: MenuItem, variant?: CartVariant) => {
    const variants = getAvailableCartVariants(menuItem);
    const preferredVariant = variant && variants.includes(variant)
      ? variant
      : variants.includes("regular")
        ? "regular"
        : variants[0] ?? "regular";

    setItems((current) => {
      const existing = current.find((entry) => entry.menuItemId === menuItem.id && entry.variant === preferredVariant);

      if (existing) {
        return current.map((entry) => entry.menuItemId === menuItem.id && entry.variant === preferredVariant
          ? { ...entry, quantity: entry.quantity + 1 }
          : entry);
      }

      return [
        ...current,
        {
          menuItemId: menuItem.id,
          name: menuItem.name,
          imageUrl: menuItem.image_url ?? "/indian_restaurant_logo.jpg",
          variant: preferredVariant,
          unitPrice: getCartVariantPrice(menuItem, preferredVariant),
          quantity: 1,
        },
      ];
    });
  };

  const updateQuantity = (menuItemId: number, variant: CartVariant, quantity: number) => {
    setItems((current) => {
      if (quantity <= 0) {
        return current.filter((entry) => !(entry.menuItemId === menuItemId && entry.variant === variant));
      }

      return current.map((entry) => entry.menuItemId === menuItemId && entry.variant === variant
        ? { ...entry, quantity }
        : entry);
    });
  };

  const removeItem = (menuItemId: number, variant: CartVariant) => {
    setItems((current) => current.filter((entry) => !(entry.menuItemId === menuItemId && entry.variant === variant)));
  };

  const clearCart = () => {
    setItems([]);
  };

  const setTableInfo = (table: ActiveTableInfo | null) => {
    setActiveTable(table);
  };

  const value = useMemo<CartContextValue>(() => ({
    items,
    itemCount,
    subtotal,
    activeTable,
    addItem,
    updateQuantity,
    removeItem,
    clearCart,
    setTableInfo,
  }), [items, itemCount, subtotal, activeTable]);

  return <CartContext.Provider value={value}>{children}</CartContext.Provider>;
}

export function useCart() {
  const context = useContext(CartContext);
  if (!context) {
    throw new Error("useCart must be used within a CartProvider");
  }

  return context;
}
