package model

type Category struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	DisplayOrder int    `json:"display_order"`
	ImageURL     string `json:"image_url"`
	IsActive     bool   `json:"is_active"`
}

type GalleryImage struct {
	ID           int64  `json:"id"`
	ImageURL     string `json:"image_url"`
	Caption      string `json:"caption"`
	DisplayOrder int    `json:"display_order"`
	IsActive     bool   `json:"is_active"`
	IsFeatured   bool   `json:"is_featured"`
}

type GalleryImagePayload struct {
	ID           int64  `json:"id,omitempty"`
	ImageURL     string `json:"image_url"`
	Caption      string `json:"caption"`
	DisplayOrder int    `json:"display_order"`
	IsActive     bool   `json:"is_active"`
	IsFeatured   bool   `json:"is_featured"`
}

type Review struct {
	ID           int64  `json:"id"`
	AuthorName   string `json:"author_name"`
	Rating       int    `json:"rating"`
	ReviewText   string `json:"review_text"`
	Source       string `json:"source"`
	SourceURL    string `json:"source_url"`
	DisplayOrder int    `json:"display_order"`
	IsFeatured   bool   `json:"is_featured"`
	IsActive     bool   `json:"is_active"`
}

type ReviewPayload struct {
	ID           int64  `json:"id,omitempty"`
	AuthorName   string `json:"author_name"`
	Rating       int    `json:"rating"`
	ReviewText   string `json:"review_text"`
	Source       string `json:"source"`
	SourceURL    string `json:"source_url"`
	DisplayOrder int    `json:"display_order"`
	IsFeatured   bool   `json:"is_featured"`
	IsActive     bool   `json:"is_active"`
}

type MenuItem struct {
	ID           int64    `json:"id"`
	CategoryID   int64    `json:"category_id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Price        *float64 `json:"price"`
	PriceHalf    *float64 `json:"price_half"`
	PriceFull    *float64 `json:"price_full"`
	Currency     string   `json:"currency"`
	IsVeg        bool     `json:"is_veg"`
	IsAvailable  bool     `json:"is_available"`
	IsFeatured   bool     `json:"is_featured"`
	IsBestseller bool     `json:"is_bestseller"`
	IsTop10      bool     `json:"is_top10"`
	SortOrder    int      `json:"sort_order"`
	ImageURL     string   `json:"image_url"`
}

type MenuResponse struct {
	Restaurant struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		Currency string `json:"currency"`
	} `json:"restaurant"`
	Categories []Category `json:"categories"`
	Items      []MenuItem `json:"items"`
}

type AdminMenuData struct {
	Categories []Category `json:"categories"`
	Items      []MenuItem `json:"items"`
}

type AdminMenuItemPayload struct {
	ID           int64    `json:"id,omitempty"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	CategoryID   int64    `json:"category_id"`
	Price        *float64 `json:"price"`
	PriceHalf    *float64 `json:"price_half"`
	PriceFull    *float64 `json:"price_full"`
	IsVeg        bool     `json:"is_veg"`
	IsAvailable  bool     `json:"is_available"`
	IsFeatured   bool     `json:"is_featured"`
	IsBestseller bool     `json:"is_bestseller"`
	IsTop10      bool     `json:"is_top10"`
	SortOrder    int      `json:"sort_order"`
	ImageURL     string   `json:"image_url"`
}

type AdminCategoryPayload struct {
	ID           int64  `json:"id,omitempty"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	DisplayOrder int    `json:"display_order"`
	ImageURL     string `json:"image_url"`
	IsActive     bool   `json:"is_active"`
}

type Table struct {
	ID           int64  `json:"id"`
	TableNumber  string `json:"table_number"`
	DisplayName  string `json:"display_name"`
	QRCodeToken  string `json:"qr_token"`
	IsActive     bool   `json:"is_active"`
	DisplayOrder int    `json:"display_order"`
}

type TablePayload struct {
	ID           int64  `json:"id,omitempty"`
	TableNumber  string `json:"table_number"`
	DisplayName  string `json:"display_name"`
	QRCodeToken  string `json:"qr_token"`
	IsActive     bool   `json:"is_active"`
	DisplayOrder int    `json:"display_order"`
}
