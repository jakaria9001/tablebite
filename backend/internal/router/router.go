package router

import (
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jakaria9001/tablebite/internal/handler"
	appmw "github.com/jakaria9001/tablebite/internal/middleware"
)

func New(healthHandler http.HandlerFunc, menuHandler *handler.MenuHandler, authHandler *handler.AuthHandler, corsOrigin string) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(appmw.SlogLogger)
	r.Use(appmw.JSONErrors)
	r.Use(chimw.Recoverer)
	r.Use(appmw.CORS(corsOrigin))
	r.Use(appmw.RequestIDHeader)

	r.Get("/api/health", healthHandler)

	r.Route("/api/v1", func(api chi.Router) {
		api.Use(appmw.CSRF(os.Getenv("CSRF_SECRET")))
		api.Use(appmw.RateLimit(20, time.Minute))

		api.Get("/menu", menuHandler.PublicMenu)
		api.Get("/categories", menuHandler.Categories)
		api.Get("/settings", menuHandler.GetRestaurantSettings)
		api.Get("/gallery", menuHandler.PublicGallery)
		api.Get("/banners", menuHandler.PublicBanners)
		api.Get("/reviews", menuHandler.PublicReviews)
		api.Post("/auth/login", authHandler.Login)
		api.Post("/auth/logout", authHandler.Logout)
		api.Get("/auth/csrf", authHandler.CSRFToken)

		api.Group(func(admin chi.Router) {
			admin.Use(appmw.WithAdminUser)
			admin.Use(appmw.RequireRole("admin"))
			admin.Get("/auth/me", authHandler.Me)
			admin.Post("/auth/change-password", authHandler.ChangePassword)
			admin.Get("/admin/menu", menuHandler.AdminMenu)
			admin.Post("/admin/menu", menuHandler.CreateMenuItem)
			admin.Put("/admin/menu/{id}", menuHandler.UpdateMenuItem)
			admin.Delete("/admin/menu/{id}", menuHandler.DeleteMenuItem)
			admin.Get("/admin/categories", menuHandler.AdminCategories)
			admin.Post("/admin/categories", menuHandler.CreateCategory)
			admin.Put("/admin/categories/{id}", menuHandler.UpdateCategory)
			admin.Post("/admin/categories/reorder", menuHandler.ReorderCategories)
			admin.Get("/admin/gallery", menuHandler.AdminGallery)
			admin.Post("/admin/gallery", menuHandler.CreateGalleryImage)
			admin.Put("/admin/gallery/{id}", menuHandler.UpdateGalleryImage)
			admin.Delete("/admin/gallery/{id}", menuHandler.DeleteGalleryImage)
			admin.Post("/admin/gallery/reorder", menuHandler.ReorderGalleryImages)
			admin.Get("/admin/settings", menuHandler.AdminGetRestaurantSettings)
			admin.Post("/admin/upload", menuHandler.UploadImage)
			admin.Put("/admin/settings", menuHandler.UpdateRestaurantSettings)
			admin.Get("/admin/banners", menuHandler.AdminBanners)
			admin.Post("/admin/banners", menuHandler.CreateBanner)
			admin.Put("/admin/banners/{id}", menuHandler.UpdateBanner)
			admin.Delete("/admin/banners/{id}", menuHandler.DeleteBanner)
			admin.Post("/admin/banners/reorder", menuHandler.ReorderBanners)
			admin.Get("/admin/reviews", menuHandler.AdminReviews)
			admin.Post("/admin/reviews", menuHandler.CreateReview)
			admin.Put("/admin/reviews/{id}", menuHandler.UpdateReview)
			admin.Delete("/admin/reviews/{id}", menuHandler.DeleteReview)
			admin.Post("/admin/reviews/reorder", menuHandler.ReorderReviews)
			admin.Get("/admin/tables", menuHandler.AdminTables)
			admin.Post("/admin/tables", menuHandler.CreateTable)
			admin.Put("/admin/tables/{id}", menuHandler.UpdateTable)
			admin.Delete("/admin/tables/{id}", menuHandler.DeleteTable)
			admin.Post("/admin/tables/reorder", menuHandler.ReorderTables)
		})

	})

	return r
}
