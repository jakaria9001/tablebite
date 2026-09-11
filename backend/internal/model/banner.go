package model

type Banner struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
	ImageURL     string `json:"image_url"`
	CTALabel     string `json:"cta_label"`
	CTAURL       string `json:"cta_url"`
	DisplayOrder int    `json:"display_order"`
	StartsAt     string `json:"starts_at"`
	EndsAt       string `json:"ends_at"`
	IsActive     bool   `json:"is_active"`
}

type BannerPayload struct {
	ID           int64  `json:"id,omitempty"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
	ImageURL     string `json:"image_url"`
	CTALabel     string `json:"cta_label"`
	CTAURL       string `json:"cta_url"`
	DisplayOrder int    `json:"display_order"`
	StartsAt     string `json:"starts_at"`
	EndsAt       string `json:"ends_at"`
	IsActive     bool   `json:"is_active"`
}
