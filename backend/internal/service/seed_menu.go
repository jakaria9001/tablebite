package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jakaria9001/tablebite/internal/model"
)

type MenuSeeder struct{ db *pgxpool.Pool }

func NewMenuSeeder(db *pgxpool.Pool) *MenuSeeder { return &MenuSeeder{db: db} }

func (s *MenuSeeder) Seed(ctx context.Context, data model.SeedFile) (int, int, error) {
	if strings.TrimSpace(data.RestaurantName) == "" {
		return 0, 0, fmt.Errorf("restaurant_name is required")
	}
	if len(data.Categories) == 0 {
		return 0, 0, fmt.Errorf("categories are required")
	}
	if len(data.MenuItems) == 0 {
		return 0, 0, fmt.Errorf("menu_items are required")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback(ctx)

	var restaurantID int64
	slug := slugify(data.RestaurantName)
	currency := data.Currency
	if currency == "" {
		currency = "INR"
	}
	if err := tx.QueryRow(ctx, `
        INSERT INTO restaurants(name, slug, currency, is_active)
        VALUES ($1, $2, $3, TRUE)
        ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name, currency = EXCLUDED.currency, updated_at = NOW()
        RETURNING id`, data.RestaurantName, slug, currency).Scan(&restaurantID); err != nil {
		return 0, 0, fmt.Errorf("upsert restaurant: %w", err)
	}

	for _, c := range data.Categories {
		cslug := slugify(c.Name)
		if _, err := tx.Exec(ctx, `
            INSERT INTO categories(restaurant_id, source_id, name, slug, display_order, image_url, is_active)
            VALUES ($1,$2,$3,$4,$5,$6,TRUE)
            ON CONFLICT (restaurant_id, source_id) DO UPDATE SET name=EXCLUDED.name, slug=EXCLUDED.slug, display_order=EXCLUDED.display_order, image_url=EXCLUDED.image_url, is_active=TRUE, updated_at=NOW()`, restaurantID, c.ID, c.Name, cslug, c.DisplayOrder, nullIfEmpty(c.ImageURL)); err != nil {
			return 0, 0, fmt.Errorf("upsert category %d: %w", c.ID, err)
		}
	}

	for _, item := range data.MenuItems {
		if _, err := tx.Exec(ctx, `
            INSERT INTO menu_items(restaurant_id, category_source_id, source_id, name, description, price, price_half, price_full, currency, is_veg, is_available, sort_order, image_url, is_active)
            VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,TRUE)
            ON CONFLICT (restaurant_id, source_id) DO UPDATE SET
                category_source_id=EXCLUDED.category_source_id,
                name=EXCLUDED.name,
                description=EXCLUDED.description,
                price=EXCLUDED.price,
                price_half=EXCLUDED.price_half,
                price_full=EXCLUDED.price_full,
                currency=EXCLUDED.currency,
                is_veg=EXCLUDED.is_veg,
                is_available=EXCLUDED.is_available,
                sort_order=EXCLUDED.sort_order,
                image_url=EXCLUDED.image_url,
                is_active=TRUE,
                updated_at=NOW()`, restaurantID, item.CategoryID, item.ID, item.Name, item.Description, item.Price, item.PriceHalf, item.PriceFull, item.Currency, item.IsVeg, item.IsAvailable, item.SortOrder, nullIfEmpty(item.ImageURL)); err != nil {
			return 0, 0, fmt.Errorf("upsert item %d: %w", item.ID, err)
		}
	}

	// Existing records are intentionally left active; destructive synchronization belongs in a future import mode.

	if err := tx.Commit(ctx); err != nil {
		return 0, 0, err
	}
	return len(data.Categories), len(data.MenuItems), nil
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	dash := false
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			dash = false
			continue
		}
		if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.Trim(b.String(), "-")
}
