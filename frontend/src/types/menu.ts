export interface Category {
  id: number;
  name: string;
  slug: string;
  display_order: number;
  image_url: string | null;
}

export interface MenuItem {
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

export interface MenuResponse {
  restaurant: {
    id: number;
    name: string;
    currency: string;
  };
  categories: Category[];
  items: MenuItem[];
}
