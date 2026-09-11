import { apiFetch, apiGet } from "./client";

export interface Review {
  id: number;
  author_name: string;
  rating: number;
  review_text: string;
  source: string;
  source_url: string;
  display_order: number;
  is_featured: boolean;
  is_active: boolean;
}

export interface ReviewPayload {
  id?: number;
  author_name: string;
  rating: number;
  review_text: string;
  source: string;
  source_url: string;
  display_order?: number;
  is_featured: boolean;
  is_active: boolean;
}

export function getPublicReviews(limit = 6): Promise<{ reviews: Review[] }> {
  return apiGet<{ reviews: Review[] }>(`/api/v1/reviews?limit=${limit}`);
}

export function getAdminReviews(): Promise<{ reviews: Review[] }> {
  return apiGet<{ reviews: Review[] }>('/api/v1/admin/reviews');
}

export function createReview(payload: ReviewPayload): Promise<Review> {
  return apiFetch<Review>('/api/v1/admin/reviews', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function updateReview(id: number, payload: ReviewPayload): Promise<Review> {
  return apiFetch<Review>(`/api/v1/admin/reviews/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  });
}

export function deleteReview(id: number): Promise<void> {
  return apiFetch<void>(`/api/v1/admin/reviews/${id}`, { method: 'DELETE' });
}

export function reorderReviews(reviewIds: number[]): Promise<void> {
  return apiFetch<void>('/api/v1/admin/reviews/reorder', {
    method: 'POST',
    body: JSON.stringify(reviewIds),
  });
}
