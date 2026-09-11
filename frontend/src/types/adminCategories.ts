export interface AdminCategory {
  id: number;
  name: string;
  slug: string;
  display_order: number;
  image_url: string | null;
  is_active: boolean;
}

export interface AdminCategoryPayload {
  id?: number;
  name: string;
  slug: string;
  display_order?: number;
  image_url: string | null;
  is_active: boolean;
}
