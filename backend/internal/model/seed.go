package model

type SeedFile struct {
	RestaurantName string         `json:"restaurant_name"`
	Currency       string         `json:"currency"`
	Categories     []SeedCategory `json:"categories"`
	MenuItems      []SeedMenuItem `json:"menu_items"`
}

type SeedCategory struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	DisplayOrder int    `json:"display_order"`
	ImageURL     string `json:"image_url"`
}

type SeedMenuItem struct {
	ID          int64    `json:"id"`
	CategoryID  int64    `json:"category_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       *float64 `json:"price"`
	PriceHalf   *float64 `json:"price_half"`
	PriceFull   *float64 `json:"price_full"`
	Currency    string   `json:"currency"`
	IsVeg       bool     `json:"is_veg"`
	IsAvailable bool     `json:"is_available"`
	SortOrder   int      `json:"sort_order"`
	ImageURL    string   `json:"image_url"`
}
