package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jakaria9001/tablebite/internal/model"
	"github.com/jakaria9001/tablebite/internal/repository"
)

type MenuHandler struct{ repo *repository.MenuRepository }

func NewMenuHandler(repo *repository.MenuRepository) *MenuHandler { return &MenuHandler{repo: repo} }

func (h *MenuHandler) PublicMenu(w http.ResponseWriter, r *http.Request) {
	data, err := h.repo.GetPublicMenu(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func (h *MenuHandler) Categories(w http.ResponseWriter, r *http.Request) {
	data, err := h.repo.GetCategories(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"categories": data})
}

func (h *MenuHandler) AdminMenu(w http.ResponseWriter, r *http.Request) {
	data, err := h.repo.GetAdminMenuData(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func (h *MenuHandler) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	var payload model.AdminMenuItemPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	item, err := h.repo.CreateMenuItem(r.Context(), payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(item)
}

func (h *MenuHandler) UpdateMenuItem(w http.ResponseWriter, r *http.Request) {
	itemID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid item id", http.StatusBadRequest)
		return
	}

	var payload model.AdminMenuItemPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	payload.ID = itemID

	item, err := h.repo.UpdateMenuItem(r.Context(), payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(item)
}

func (h *MenuHandler) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	itemID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid item id", http.StatusBadRequest)
		return
	}
	if err := h.repo.DeleteMenuItem(r.Context(), itemID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MenuHandler) AdminCategories(w http.ResponseWriter, r *http.Request) {
	data, err := h.repo.GetAdminCategories(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"categories": data})
}

func (h *MenuHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var payload model.AdminCategoryPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	category, err := h.repo.CreateCategory(r.Context(), payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(category)
}

func (h *MenuHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	categoryID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid category id", http.StatusBadRequest)
		return
	}
	var payload model.AdminCategoryPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	payload.ID = categoryID
	category, err := h.repo.UpdateCategory(r.Context(), payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(category)
}

func (h *MenuHandler) ReorderCategories(w http.ResponseWriter, r *http.Request) {
	var payload []int64
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.repo.UpdateCategoryOrder(r.Context(), payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MenuHandler) PublicGallery(w http.ResponseWriter, r *http.Request) {
	limit := 3
	if value := r.URL.Query().Get("limit"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	data, err := h.repo.GetPublicGalleryImages(r.Context(), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"images": data})
}

func (h *MenuHandler) AdminGallery(w http.ResponseWriter, r *http.Request) {
	data, err := h.repo.GetAdminGalleryImages(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"images": data})
}

func (h *MenuHandler) CreateGalleryImage(w http.ResponseWriter, r *http.Request) {
	var payload model.GalleryImagePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	item, err := h.repo.CreateGalleryImage(r.Context(), payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(item)
}

func (h *MenuHandler) UpdateGalleryImage(w http.ResponseWriter, r *http.Request) {
	imageID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid image id", http.StatusBadRequest)
		return
	}
	var payload model.GalleryImagePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	payload.ID = imageID
	item, err := h.repo.UpdateGalleryImage(r.Context(), payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(item)
}

func (h *MenuHandler) DeleteGalleryImage(w http.ResponseWriter, r *http.Request) {
	imageID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid image id", http.StatusBadRequest)
		return
	}
	if err := h.repo.DeleteGalleryImage(r.Context(), imageID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MenuHandler) ReorderGalleryImages(w http.ResponseWriter, r *http.Request) {
	var payload []int64
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.repo.UpdateGalleryOrder(r.Context(), payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MenuHandler) AdminTables(w http.ResponseWriter, r *http.Request) {
	data, err := h.repo.AdminTables(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"tables": data})
}

func (h *MenuHandler) CreateTable(w http.ResponseWriter, r *http.Request) {
	var payload model.TablePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	table, err := h.repo.CreateTable(r.Context(), payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(table)
}

func (h *MenuHandler) UpdateTable(w http.ResponseWriter, r *http.Request) {
	tableID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid table id", http.StatusBadRequest)
		return
	}
	var payload model.TablePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	payload.ID = tableID
	table, err := h.repo.UpdateTable(r.Context(), payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(table)
}

func (h *MenuHandler) DeleteTable(w http.ResponseWriter, r *http.Request) {
	tableID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid table id", http.StatusBadRequest)
		return
	}
	if err := h.repo.DeleteTable(r.Context(), tableID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MenuHandler) ReorderTables(w http.ResponseWriter, r *http.Request) {
	var payload []int64
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.repo.ReorderTables(r.Context(), payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MenuHandler) GetRestaurantSettings(w http.ResponseWriter, r *http.Request) {
	data, err := h.repo.GetRestaurantSettings(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func (h *MenuHandler) AdminGetRestaurantSettings(w http.ResponseWriter, r *http.Request) {
	data, err := h.repo.GetRestaurantSettings(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func (h *MenuHandler) UpdateRestaurantSettings(w http.ResponseWriter, r *http.Request) {
	var payload map[string]string
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	data, err := h.repo.SaveRestaurantSettings(r.Context(), payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func (h *MenuHandler) PublicBanners(w http.ResponseWriter, r *http.Request) {
	data, err := h.repo.GetPublicBanners(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"banners": data})
}

func (h *MenuHandler) AdminBanners(w http.ResponseWriter, r *http.Request) {
	data, err := h.repo.GetAdminBanners(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"banners": data})
}

func (h *MenuHandler) CreateBanner(w http.ResponseWriter, r *http.Request) {
	var payload model.BannerPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	banner, err := h.repo.CreateBanner(r.Context(), payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(banner)
}

func (h *MenuHandler) UpdateBanner(w http.ResponseWriter, r *http.Request) {
	bannerID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid banner id", http.StatusBadRequest)
		return
	}
	var payload model.BannerPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	payload.ID = bannerID
	banner, err := h.repo.UpdateBanner(r.Context(), payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(banner)
}

func (h *MenuHandler) DeleteBanner(w http.ResponseWriter, r *http.Request) {
	bannerID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid banner id", http.StatusBadRequest)
		return
	}
	if err := h.repo.DeleteBanner(r.Context(), bannerID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MenuHandler) ReorderBanners(w http.ResponseWriter, r *http.Request) {
	var payload []int64
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.repo.UpdateBannerOrder(r.Context(), payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MenuHandler) PublicReviews(w http.ResponseWriter, r *http.Request) {
	limit := 6
	if value := r.URL.Query().Get("limit"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	data, err := h.repo.GetPublicReviews(r.Context(), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"reviews": data})
}

func (h *MenuHandler) AdminReviews(w http.ResponseWriter, r *http.Request) {
	data, err := h.repo.GetAdminReviews(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"reviews": data})
}

func (h *MenuHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	var payload model.ReviewPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	review, err := h.repo.CreateReview(r.Context(), payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(review)
}

func (h *MenuHandler) UpdateReview(w http.ResponseWriter, r *http.Request) {
	reviewID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid review id", http.StatusBadRequest)
		return
	}
	var payload model.ReviewPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	payload.ID = reviewID
	review, err := h.repo.UpdateReview(r.Context(), payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(review)
}

func (h *MenuHandler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	reviewID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid review id", http.StatusBadRequest)
		return
	}
	if err := h.repo.DeleteReview(r.Context(), reviewID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MenuHandler) ReorderReviews(w http.ResponseWriter, r *http.Request) {
	var payload []int64
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.repo.UpdateReviewOrder(r.Context(), payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
