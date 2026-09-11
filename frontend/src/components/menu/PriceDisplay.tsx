import type { MenuItem } from "../../types/menu";
import type { CartVariant } from "../../types/cart";

interface PriceDisplayProps {
  item: MenuItem;
  variant?: CartVariant;
}

export function PriceDisplay({ item, variant }: PriceDisplayProps) {
  if (item.price !== null) {
    return <span>₹{item.price}</span>;
  }

  if (variant === "half" && item.price_half !== null) {
    return <span>₹{item.price_half}</span>;
  }

  if (variant === "full" && item.price_full !== null) {
    return <span>₹{item.price_full}</span>;
  }

  return <span>₹{item.price_half} Half · ₹{item.price_full} Full</span>;
}
