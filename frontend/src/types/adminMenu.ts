export interface AdminCategory {
  id: number;
  name: string;
  slug: string;
  display_order: number;
  image_url: string | null;
}

export interface AdminMenuItem {
  id: number;
  category_id: number;
  name: string;
  description: string;
  price: number | null;
  price_half: number | null;
  price_full: number | null;
  currency: string;
  is_veg: boolean;
  is_available: boolean;
  is_featured: boolean;
  is_bestseller: boolean;
  is_top10: boolean;
  sort_order: number;
  image_url: string | null;
}

export interface AdminMenuData {
  categories: AdminCategory[];
  items: AdminMenuItem[];
}

export interface AdminMenuItemPayload {
  id?: number;
  name: string;
  description: string;
  category_id: number;
  price: number | null;
  price_half: number | null;
  price_full: number | null;
  is_veg: boolean;
  is_available: boolean;
  is_featured: boolean;
  is_bestseller: boolean;
  is_top10: boolean;
  sort_order: number;
  image_url: string | null;
}
