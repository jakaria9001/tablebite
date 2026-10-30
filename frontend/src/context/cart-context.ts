import { createContext } from "react";
import type { MenuItem } from "../types/menu";
import type { CartItem, CartVariant } from "../types/cart";

export interface ActiveTableInfo {
  id: number | null;
  tableNumber: string;
  tableName: string;
  qrToken: string;
}

export interface CartContextValue {
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

export const CartContext = createContext<CartContextValue | undefined>(undefined);
