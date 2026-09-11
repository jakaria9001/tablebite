import { apiFetch, apiGet } from "./client";

export interface GalleryImage {
  id: number;
  image_url: string;
  caption: string;
  display_order: number;
  is_active: boolean;
  is_featured: boolean;
}

export interface GalleryImagePayload {
  id?: number;
  image_url: string;
  caption: string;
  display_order?: number;
  is_active: boolean;
  is_featured: boolean;
}

export function getPublicGalleryImages(limit = 3): Promise<{ images: GalleryImage[] }> {
  return apiGet<{ images: GalleryImage[] }>(`/api/v1/gallery?limit=${limit}`);
}

export function getAdminGalleryImages(): Promise<{ images: GalleryImage[] }> {
  return apiGet<{ images: GalleryImage[] }>('/api/v1/admin/gallery');
}

export function createGalleryImage(payload: GalleryImagePayload): Promise<GalleryImage> {
  return apiFetch<GalleryImage>('/api/v1/admin/gallery', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function updateGalleryImage(id: number, payload: GalleryImagePayload): Promise<GalleryImage> {
  return apiFetch<GalleryImage>(`/api/v1/admin/gallery/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  });
}

export function deleteGalleryImage(id: number): Promise<void> {
  return apiFetch<void>(`/api/v1/admin/gallery/${id}`, { method: 'DELETE' });
}

export function reorderGalleryImages(imageIds: number[]): Promise<void> {
  return apiFetch<void>('/api/v1/admin/gallery/reorder', {
    method: 'POST',
    body: JSON.stringify(imageIds),
  });
}
