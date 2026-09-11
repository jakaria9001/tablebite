import { apiFetch, apiGet } from "./client";

export interface Banner {
  id: number;
  title: string;
  subtitle: string;
  image_url: string;
  cta_label: string;
  cta_url: string;
  display_order: number;
  starts_at: string | null;
  ends_at: string | null;
  is_active: boolean;
}

export interface BannerPayload {
  id?: number;
  title: string;
  subtitle: string;
  image_url: string;
  cta_label: string;
  cta_url: string;
  display_order?: number;
  starts_at: string;
  ends_at: string;
  is_active: boolean;
}

export function getPublicBanners(): Promise<{ banners: Banner[] }> {
  return apiGet<{ banners: Banner[] }>('/api/v1/banners');
}

export function getAdminBanners(): Promise<{ banners: Banner[] }> {
  return apiGet<{ banners: Banner[] }>('/api/v1/admin/banners');
}

export function createBanner(payload: BannerPayload): Promise<Banner> {
  return apiFetch<Banner>('/api/v1/admin/banners', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function updateBanner(id: number, payload: BannerPayload): Promise<Banner> {
  return apiFetch<Banner>(`/api/v1/admin/banners/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  });
}

export function deleteBanner(id: number): Promise<void> {
  return apiFetch<void>(`/api/v1/admin/banners/${id}`, { method: 'DELETE' });
}

export function reorderBanners(bannerIds: number[]): Promise<void> {
  return apiFetch<void>('/api/v1/admin/banners/reorder', {
    method: 'POST',
    body: JSON.stringify(bannerIds),
  });
}
