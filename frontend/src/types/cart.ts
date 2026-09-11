import type { MenuItem } from "./menu";

export type CartVariant = "regular" | "half" | "full";

export type CartItem = {
  menuItemId: number;
  name: string;
  imageUrl: string;
  variant: CartVariant;
  unitPrice: number;
  quantity: number;
};

export function getAvailableCartVariants(item: MenuItem): CartVariant[] {
  const variants: CartVariant[] = [];

  if (item.price !== null) {
    variants.push("regular");
  }

  if (item.price_half !== null) {
    variants.push("half");
  }

  if (item.price_full !== null) {
    variants.push("full");
  }

  return variants;
}

export function getCartVariantPrice(item: MenuItem, variant: CartVariant): number {
  if (variant === "half") {
    return item.price_half ?? item.price ?? 0;
  }

  if (variant === "full") {
    return item.price_full ?? item.price ?? 0;
  }

  return item.price ?? 0;
}

export function getCartVariantLabel(variant: CartVariant): string {
  if (variant === "half") return "Half";
  if (variant === "full") return "Full";
  return "Regular";
}
