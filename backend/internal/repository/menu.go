package repository

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jakaria9001/tablebite/internal/model"
)

type MenuRepository struct{ db *pgxpool.Pool }

func NewMenuRepository(db *pgxpool.Pool) *MenuRepository { return &MenuRepository{db: db} }

func (r *MenuRepository) GetPublicMenu(ctx context.Context) (model.MenuResponse, error) {
	var out model.MenuResponse

	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return out, err
	}
	out.Restaurant.ID = restaurantID

	var slug string
	if err := r.db.QueryRow(ctx, `SELECT name, currency, slug FROM restaurants WHERE id = $1`, restaurantID).Scan(&out.Restaurant.Name, &out.Restaurant.Currency, &slug); err != nil {
		return out, fmt.Errorf("load restaurant: %w", err)
	}

	if err := r.normalizeCategoryOrders(ctx, restaurantID); err != nil {
		return out, err
	}

	categories, err := r.db.Query(ctx, `SELECT id, name, slug, display_order, COALESCE(image_url, ''), is_active FROM categories WHERE restaurant_id = $1 AND is_active = TRUE ORDER BY display_order, id`, restaurantID)
	if err != nil {
		return out, fmt.Errorf("load categories: %w", err)
	}
	defer categories.Close()
	for categories.Next() {
		var c model.Category
		if err := categories.Scan(&c.ID, &c.Name, &c.Slug, &c.DisplayOrder, &c.ImageURL, &c.IsActive); err != nil {
			return out, fmt.Errorf("scan category: %w", err)
		}
		out.Categories = append(out.Categories, c)
	}
	if err := categories.Err(); err != nil {
		return out, fmt.Errorf("iterate categories: %w", err)
	}

	categoryColumn, err := r.resolveMenuItemCategoryColumn(ctx)
	if err != nil {
		return out, fmt.Errorf("resolve menu item category column: %w", err)
	}

	items, err := r.db.Query(ctx, fmt.Sprintf(`
        SELECT mi.id, mi.%s AS category_id, mi.name, mi.description,
               mi.price::double precision,
               mi.price_half::double precision,
               mi.price_full::double precision,
               mi.currency, mi.is_veg, mi.is_available, mi.is_featured, mi.is_bestseller, mi.is_top10, mi.sort_order, COALESCE(mi.image_url, '')
        FROM menu_items mi
        LEFT JOIN categories c ON c.restaurant_id = mi.restaurant_id AND c.source_id = mi.category_source_id
        WHERE mi.restaurant_id = $1 AND mi.is_active = TRUE AND (c.id IS NULL OR c.is_active = TRUE)
        ORDER BY category_id, mi.sort_order, mi.id`, categoryColumn), restaurantID)
	if err != nil {
		return out, fmt.Errorf("load menu items: %w", err)
	}
	defer items.Close()
	for items.Next() {
		var item model.MenuItem
		if err := items.Scan(&item.ID, &item.CategoryID, &item.Name, &item.Description, &item.Price, &item.PriceHalf, &item.PriceFull, &item.Currency, &item.IsVeg, &item.IsAvailable, &item.IsFeatured, &item.IsBestseller, &item.IsTop10, &item.SortOrder, &item.ImageURL); err != nil {
			return out, fmt.Errorf("scan item: %w", err)
		}
		out.Items = append(out.Items, item)
	}
	if err := items.Err(); err != nil {
		return out, fmt.Errorf("iterate items: %w", err)
	}

	return out, nil
}

func (r *MenuRepository) GetAdminMenuData(ctx context.Context) (model.AdminMenuData, error) {
	var out model.AdminMenuData

	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return out, err
	}

	categories, err := r.db.Query(ctx, `SELECT id, name, slug, display_order, COALESCE(image_url, ''), is_active FROM categories WHERE restaurant_id = $1 AND is_active = TRUE ORDER BY display_order, id`, restaurantID)
	if err != nil {
		return out, fmt.Errorf("load categories: %w", err)
	}
	defer categories.Close()
	for categories.Next() {
		var c model.Category
		if err := categories.Scan(&c.ID, &c.Name, &c.Slug, &c.DisplayOrder, &c.ImageURL, &c.IsActive); err != nil {
			return out, fmt.Errorf("scan category: %w", err)
		}
		out.Categories = append(out.Categories, c)
	}
	if err := categories.Err(); err != nil {
		return out, fmt.Errorf("iterate categories: %w", err)
	}

	items, err := r.db.Query(ctx, `SELECT mi.id, c.id AS category_id, mi.name, mi.description,
			mi.price::double precision,
			mi.price_half::double precision,
			mi.price_full::double precision,
			mi.currency, mi.is_veg, mi.is_available, mi.is_featured, mi.is_bestseller, mi.is_top10, mi.sort_order, COALESCE(mi.image_url, '')
		FROM menu_items mi
		LEFT JOIN categories c ON c.restaurant_id = mi.restaurant_id AND c.source_id = mi.category_source_id
		WHERE mi.restaurant_id = $1 AND mi.is_active = TRUE
		ORDER BY mi.sort_order, mi.id`, restaurantID)
	if err != nil {
		return out, fmt.Errorf("load menu items: %w", err)
	}
	defer items.Close()
	for items.Next() {
		var item model.MenuItem
		if err := items.Scan(&item.ID, &item.CategoryID, &item.Name, &item.Description, &item.Price, &item.PriceHalf, &item.PriceFull, &item.Currency, &item.IsVeg, &item.IsAvailable, &item.IsFeatured, &item.IsBestseller, &item.IsTop10, &item.SortOrder, &item.ImageURL); err != nil {
			return out, fmt.Errorf("scan item: %w", err)
		}
		out.Items = append(out.Items, item)
	}
	if err := items.Err(); err != nil {
		return out, fmt.Errorf("iterate items: %w", err)
	}

	return out, nil
}

func (r *MenuRepository) CreateMenuItem(ctx context.Context, payload model.AdminMenuItemPayload) (model.MenuItem, error) {
	var out model.MenuItem
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return out, err
	}

	var categorySourceID int64
	if err := r.db.QueryRow(ctx, `SELECT source_id FROM categories WHERE id = $1 AND restaurant_id = $2 AND is_active = TRUE`, payload.CategoryID, restaurantID).Scan(&categorySourceID); err != nil {
		return out, fmt.Errorf("resolve category: %w", err)
	}

	sourceID := time.Now().UnixNano()
	if err := r.db.QueryRow(ctx, `INSERT INTO menu_items (
		restaurant_id, category_source_id, source_id, name, description, price, price_half, price_full,
		currency, is_veg, is_available, sort_order, image_url, is_featured, is_bestseller, is_top10
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'INR', $9, $10, $11, $12, $13, $14, $15)
	RETURNING id, category_source_id, name, description, price::double precision, price_half::double precision, price_full::double precision, currency, is_veg, is_available, is_featured, is_bestseller, is_top10, sort_order, COALESCE(image_url, '')`,
		restaurantID, categorySourceID, sourceID, payload.Name, payload.Description, payload.Price, payload.PriceHalf, payload.PriceFull, payload.IsVeg, payload.IsAvailable, payload.SortOrder, payload.ImageURL, payload.IsFeatured, payload.IsBestseller, payload.IsTop10).Scan(&out.ID, &out.CategoryID, &out.Name, &out.Description, &out.Price, &out.PriceHalf, &out.PriceFull, &out.Currency, &out.IsVeg, &out.IsAvailable, &out.IsFeatured, &out.IsBestseller, &out.IsTop10, &out.SortOrder, &out.ImageURL); err != nil {
		return out, fmt.Errorf("create menu item: %w", err)
	}

	return out, nil
}

func (r *MenuRepository) UpdateMenuItem(ctx context.Context, payload model.AdminMenuItemPayload) (model.MenuItem, error) {
	var out model.MenuItem
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return out, err
	}

	var categorySourceID int64
	if err := r.db.QueryRow(ctx, `SELECT source_id FROM categories WHERE id = $1 AND restaurant_id = $2 AND is_active = TRUE`, payload.CategoryID, restaurantID).Scan(&categorySourceID); err != nil {
		return out, fmt.Errorf("resolve category: %w", err)
	}

	if err := r.db.QueryRow(ctx, `UPDATE menu_items SET
		name = $3,
		description = $4,
		category_source_id = $5,
		price = $6,
		price_half = $7,
		price_full = $8,
		is_veg = $9,
		is_available = $10,
		is_featured = $11,
		is_bestseller = $12,
		is_top10 = $13,
		sort_order = $14,
		image_url = $15,
		updated_at = NOW()
	WHERE id = $1 AND restaurant_id = $2 AND is_active = TRUE
	RETURNING id, category_source_id, name, description, price::double precision, price_half::double precision, price_full::double precision, currency, is_veg, is_available, is_featured, is_bestseller, is_top10, sort_order, COALESCE(image_url, '')`, payload.ID, restaurantID, payload.Name, payload.Description, categorySourceID, payload.Price, payload.PriceHalf, payload.PriceFull, payload.IsVeg, payload.IsAvailable, payload.IsFeatured, payload.IsBestseller, payload.IsTop10, payload.SortOrder, payload.ImageURL).Scan(&out.ID, &out.CategoryID, &out.Name, &out.Description, &out.Price, &out.PriceHalf, &out.PriceFull, &out.Currency, &out.IsVeg, &out.IsAvailable, &out.IsFeatured, &out.IsBestseller, &out.IsTop10, &out.SortOrder, &out.ImageURL); err != nil {
		return out, fmt.Errorf("update menu item: %w", err)
	}

	return out, nil
}

func (r *MenuRepository) DeleteMenuItem(ctx context.Context, itemID int64) error {
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, `UPDATE menu_items SET is_active = FALSE, updated_at = NOW() WHERE id = $1 AND restaurant_id = $2 AND is_active = TRUE`, itemID, restaurantID)
	return err
}

func (r *MenuRepository) GetAdminCategories(ctx context.Context) ([]model.Category, error) {
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return nil, err
	}

	if err := r.normalizeCategoryOrders(ctx, restaurantID); err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `SELECT id, name, slug, display_order, COALESCE(image_url, ''), is_active FROM categories WHERE restaurant_id = $1 ORDER BY display_order, id`, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("load categories: %w", err)
	}
	defer rows.Close()

	var categories []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.DisplayOrder, &c.ImageURL, &c.IsActive); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}
	return categories, nil
}

func (r *MenuRepository) CreateCategory(ctx context.Context, payload model.AdminCategoryPayload) (model.Category, error) {
	var out model.Category
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return out, err
	}

	var nextOrder int
	if err := r.db.QueryRow(ctx, `SELECT COALESCE(MAX(display_order), 0) + 1 FROM categories WHERE restaurant_id = $1`, restaurantID).Scan(&nextOrder); err != nil {
		return out, fmt.Errorf("calculate category order: %w", err)
	}

	if err := r.normalizeCategoryOrders(ctx, restaurantID); err != nil {
		return out, err
	}

	if err := r.db.QueryRow(ctx, `INSERT INTO categories (restaurant_id, source_id, name, slug, display_order, image_url, is_active)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id, name, slug, display_order, COALESCE(image_url, ''), is_active`, restaurantID, time.Now().UnixNano(), payload.Name, payload.Slug, nextOrder, payload.ImageURL, payload.IsActive).Scan(&out.ID, &out.Name, &out.Slug, &out.DisplayOrder, &out.ImageURL, &out.IsActive); err != nil {
		return out, fmt.Errorf("create category: %w", err)
	}
	return out, nil
}

func (r *MenuRepository) UpdateCategory(ctx context.Context, payload model.AdminCategoryPayload) (model.Category, error) {
	var out model.Category
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return out, err
	}

	if err := r.db.QueryRow(ctx, `UPDATE categories SET name = $3, slug = $4, image_url = $5, is_active = $6, updated_at = NOW()
	WHERE id = $1 AND restaurant_id = $2
	RETURNING id, name, slug, display_order, COALESCE(image_url, ''), is_active`, payload.ID, restaurantID, payload.Name, payload.Slug, payload.ImageURL, payload.IsActive).Scan(&out.ID, &out.Name, &out.Slug, &out.DisplayOrder, &out.ImageURL, &out.IsActive); err != nil {
		return out, fmt.Errorf("update category: %w", err)
	}
	return out, nil
}

func (r *MenuRepository) UpdateCategoryOrder(ctx context.Context, categoryIDs []int64) error {
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return err
	}

	for index, categoryID := range categoryIDs {
		if _, err := r.db.Exec(ctx, `UPDATE categories SET display_order = $1, updated_at = NOW() WHERE id = $2 AND restaurant_id = $3`, index+1, categoryID, restaurantID); err != nil {
			return fmt.Errorf("update category order: %w", err)
		}
	}
	return nil
}

func (r *MenuRepository) normalizeCategoryOrders(ctx context.Context, restaurantID int64) error {
	rows, err := r.db.Query(ctx, `SELECT id FROM categories WHERE restaurant_id = $1 ORDER BY CASE WHEN display_order > 0 THEN 0 ELSE 1 END, display_order, created_at, id`, restaurantID)
	if err != nil {
		return fmt.Errorf("load category ids: %w", err)
	}
	defer rows.Close()

	var categoryIDs []int64
	for rows.Next() {
		var categoryID int64
		if err := rows.Scan(&categoryID); err != nil {
			return fmt.Errorf("scan category id: %w", err)
		}
		categoryIDs = append(categoryIDs, categoryID)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate category ids: %w", err)
	}

	for index, categoryID := range categoryIDs {
		if _, err := r.db.Exec(ctx, `UPDATE categories SET display_order = $1, updated_at = NOW() WHERE id = $2 AND restaurant_id = $3`, index+1, categoryID, restaurantID); err != nil {
			return fmt.Errorf("normalize category order: %w", err)
		}
	}
	return nil
}

func (r *MenuRepository) restaurantTablesQuery(ctx context.Context) (string, error) {
	columnExists, err := r.columnExists(ctx, "restaurant_tables", "display_order")
	if err != nil {
		return "", err
	}
	if columnExists {
		return `SELECT id, table_number, COALESCE(display_name, ''), qr_token, is_active, COALESCE(display_order, 0) FROM restaurant_tables WHERE restaurant_id = $1 ORDER BY display_order, id`, nil
	}
	return `SELECT id, table_number, COALESCE(display_name, ''), qr_token, is_active, 0 FROM restaurant_tables WHERE restaurant_id = $1 ORDER BY id`, nil
}

func (r *MenuRepository) columnExists(ctx context.Context, tableName, columnName string) (bool, error) {
	var exists bool
	if err := r.db.QueryRow(ctx, `SELECT EXISTS (
		SELECT 1
		FROM information_schema.columns
		WHERE table_schema = current_schema()
		AND table_name = $1
		AND column_name = $2
	)`, tableName, columnName).Scan(&exists); err != nil {
		return false, fmt.Errorf("check %s.%s column: %w", tableName, columnName, err)
	}
	return exists, nil
}

func (r *MenuRepository) getActiveRestaurantID(ctx context.Context) (int64, error) {
	var restaurantID int64
	if err := r.db.QueryRow(ctx, `SELECT id FROM restaurants WHERE is_active = TRUE ORDER BY id LIMIT 1`).Scan(&restaurantID); err != nil {
		return 0, fmt.Errorf("load restaurant: %w", err)
	}
	return restaurantID, nil
}

func (r *MenuRepository) resolveMenuItemCategoryColumn(ctx context.Context) (string, error) {
	var hasCategoryID bool
	if err := r.db.QueryRow(ctx, `SELECT EXISTS (
		SELECT 1
		FROM information_schema.columns
		WHERE table_schema = current_schema()
		  AND table_name = 'menu_items'
		  AND column_name = 'category_id'
	)`).Scan(&hasCategoryID); err != nil {
		return "", fmt.Errorf("check category_id column: %w", err)
	}

	var hasCategorySourceID bool
	if err := r.db.QueryRow(ctx, `SELECT EXISTS (
		SELECT 1
		FROM information_schema.columns
		WHERE table_schema = current_schema()
		  AND table_name = 'menu_items'
		  AND column_name = 'category_source_id'
	)`).Scan(&hasCategorySourceID); err != nil {
		return "", fmt.Errorf("check category_source_id column: %w", err)
	}

	return MenuItemCategoryColumnName(hasCategoryID, hasCategorySourceID)
}

func MenuItemCategoryColumnName(hasCategoryID, hasCategorySourceID bool) (string, error) {
	switch {
	case hasCategoryID:
		return "category_id", nil
	case hasCategorySourceID:
		return "category_source_id", nil
	default:
		return "", fmt.Errorf("menu_items has neither category_id nor category_source_id")
	}
}

func (r *MenuRepository) AdminTables(ctx context.Context) ([]model.Table, error) {
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return nil, err
	}

	query, err := r.restaurantTablesQuery(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, query, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("load tables: %w", err)
	}
	defer rows.Close()

	var tables []model.Table
	for rows.Next() {
		var table model.Table
		if err := rows.Scan(&table.ID, &table.TableNumber, &table.DisplayName, &table.QRCodeToken, &table.IsActive, &table.DisplayOrder); err != nil {
			return nil, fmt.Errorf("scan table: %w", err)
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tables: %w", err)
	}

	return tables, nil
}

func (r *MenuRepository) CreateTable(ctx context.Context, payload model.TablePayload) (model.Table, error) {
	var out model.Table
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return out, err
	}

	tableNumber := strings.TrimSpace(payload.TableNumber)
	if tableNumber == "" {
		tableNumber = FormatTableNumber(1)
	}

	qrToken := strings.TrimSpace(payload.QRCodeToken)
	if qrToken == "" {
		qrToken = GenerateTableToken(restaurantID, 0)
	}

	displayOrder := payload.DisplayOrder
	if displayOrder <= 0 {
		displayOrder = 1
	}

	columnExists, err := r.columnExists(ctx, "restaurant_tables", "display_order")
	if err != nil {
		return out, err
	}
	if columnExists {
		if err := r.db.QueryRow(ctx, `INSERT INTO restaurant_tables (restaurant_id, table_number, display_name, qr_token, is_active, display_order)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, table_number, COALESCE(display_name, ''), qr_token, is_active, COALESCE(display_order, 0)`, restaurantID, tableNumber, strings.TrimSpace(payload.DisplayName), qrToken, payload.IsActive, displayOrder).Scan(&out.ID, &out.TableNumber, &out.DisplayName, &out.QRCodeToken, &out.IsActive, &out.DisplayOrder); err != nil {
			return out, fmt.Errorf("create table: %w", err)
		}
		return out, nil
	}

	if err := r.db.QueryRow(ctx, `INSERT INTO restaurant_tables (restaurant_id, table_number, display_name, qr_token, is_active)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, table_number, COALESCE(display_name, ''), qr_token, is_active, 0`, restaurantID, tableNumber, strings.TrimSpace(payload.DisplayName), qrToken, payload.IsActive).Scan(&out.ID, &out.TableNumber, &out.DisplayName, &out.QRCodeToken, &out.IsActive, &out.DisplayOrder); err != nil {
		return out, fmt.Errorf("create table: %w", err)
	}

	return out, nil
}

func (r *MenuRepository) UpdateTable(ctx context.Context, payload model.TablePayload) (model.Table, error) {
	var out model.Table
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return out, err
	}

	tableNumber := strings.TrimSpace(payload.TableNumber)
	if tableNumber == "" {
		tableNumber = FormatTableNumber(1)
	}
	qrToken := strings.TrimSpace(payload.QRCodeToken)
	if qrToken == "" {
		qrToken = GenerateTableToken(restaurantID, payload.ID)
	}

	columnExists, err := r.columnExists(ctx, "restaurant_tables", "display_order")
	if err != nil {
		return out, err
	}
	if columnExists {
		if err := r.db.QueryRow(ctx, `UPDATE restaurant_tables SET table_number = $3, display_name = $4, qr_token = $5, is_active = $6, display_order = $7, updated_at = NOW()
	WHERE id = $1 AND restaurant_id = $2
	RETURNING id, table_number, COALESCE(display_name, ''), qr_token, is_active, COALESCE(display_order, 0)`, payload.ID, restaurantID, tableNumber, strings.TrimSpace(payload.DisplayName), qrToken, payload.IsActive, payload.DisplayOrder).Scan(&out.ID, &out.TableNumber, &out.DisplayName, &out.QRCodeToken, &out.IsActive, &out.DisplayOrder); err != nil {
			return out, fmt.Errorf("update table: %w", err)
		}
		return out, nil
	}

	if err := r.db.QueryRow(ctx, `UPDATE restaurant_tables SET table_number = $3, display_name = $4, qr_token = $5, is_active = $6, updated_at = NOW()
	WHERE id = $1 AND restaurant_id = $2
	RETURNING id, table_number, COALESCE(display_name, ''), qr_token, is_active, 0`, payload.ID, restaurantID, tableNumber, strings.TrimSpace(payload.DisplayName), qrToken, payload.IsActive).Scan(&out.ID, &out.TableNumber, &out.DisplayName, &out.QRCodeToken, &out.IsActive, &out.DisplayOrder); err != nil {
		return out, fmt.Errorf("update table: %w", err)
	}

	return out, nil
}

func (r *MenuRepository) DeleteTable(ctx context.Context, tableID int64) error {
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, `DELETE FROM restaurant_tables WHERE id = $1 AND restaurant_id = $2`, tableID, restaurantID)
	return err
}

func (r *MenuRepository) ReorderTables(ctx context.Context, tableIDs []int64) error {
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return err
	}

	columnExists, err := r.columnExists(ctx, "restaurant_tables", "display_order")
	if err != nil {
		return err
	}
	if !columnExists {
		return nil
	}

	for index, tableID := range tableIDs {
		if _, err := r.db.Exec(ctx, `UPDATE restaurant_tables SET display_order = $1, updated_at = NOW() WHERE id = $2 AND restaurant_id = $3`, index+1, tableID, restaurantID); err != nil {
			return fmt.Errorf("update table order: %w", err)
		}
	}
	return nil
}

func FormatTableNumber(number int) string {
	if number < 0 {
		return "00"
	}
	return fmt.Sprintf("%02d", number)
}

func GenerateTableToken(restaurantID int64, tableID int64) string {
	seed := fmt.Sprintf("tbl-%d-%d-%d", restaurantID, tableID, time.Now().UnixNano())
	seed = strings.ToLower(strings.ReplaceAll(seed, " ", "-"))
	return seed
}

func (r *MenuRepository) GetRestaurantSettings(ctx context.Context) (map[string]string, error) {
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `SELECT setting_key, setting_value FROM restaurant_settings WHERE restaurant_id = $1 ORDER BY setting_key`, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("load restaurant settings: %w", err)
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key string
		var value *string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("scan restaurant setting: %w", err)
		}
		if value != nil {
			settings[key] = *value
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate restaurant settings: %w", err)
	}
	return settings, nil
}

func (r *MenuRepository) SaveRestaurantSettings(ctx context.Context, settings map[string]string) (map[string]string, error) {
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return nil, err
	}

	for key, value := range settings {
		if _, err := r.db.Exec(ctx, `INSERT INTO restaurant_settings (restaurant_id, setting_key, setting_value)
		VALUES ($1, $2, $3)
		ON CONFLICT (restaurant_id, setting_key) DO UPDATE SET setting_value = EXCLUDED.setting_value`, restaurantID, key, value); err != nil {
			return nil, fmt.Errorf("save restaurant setting %s: %w", key, err)
		}
	}

	return r.GetRestaurantSettings(ctx)
}

func (r *MenuRepository) GetPublicGalleryImages(ctx context.Context, limit int) ([]model.GalleryImage, error) {
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return nil, err
	}

	featuredColumnExists, err := r.galleryFeaturedColumnExists(ctx)
	if err != nil {
		return nil, err
	}

	query := `SELECT id, image_url, COALESCE(caption, ''), display_order, is_active, FALSE AS is_featured
	FROM gallery_images
	WHERE restaurant_id = $1 AND is_active = TRUE
	ORDER BY display_order, id`
	if featuredColumnExists {
		query = `SELECT id, image_url, COALESCE(caption, ''), display_order, is_active, is_featured
	FROM gallery_images
	WHERE restaurant_id = $1 AND is_active = TRUE
	ORDER BY display_order, id`
	}

	rows, err := r.db.Query(ctx, query, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("load gallery images: %w", err)
	}
	defer rows.Close()

	images := make([]model.GalleryImage, 0)
	for rows.Next() {
		var item model.GalleryImage
		if err := rows.Scan(&item.ID, &item.ImageURL, &item.Caption, &item.DisplayOrder, &item.IsActive, &item.IsFeatured); err != nil {
			return nil, fmt.Errorf("scan gallery image: %w", err)
		}
		images = append(images, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate gallery images: %w", err)
	}

	if limit > 0 && len(images) > limit {
		rand.Shuffle(len(images), func(i, j int) { images[i], images[j] = images[j], images[i] })
		images = images[:limit]
	}

	return images, nil
}

func (r *MenuRepository) GetAdminGalleryImages(ctx context.Context) ([]model.GalleryImage, error) {
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return nil, err
	}

	featuredColumnExists, err := r.galleryFeaturedColumnExists(ctx)
	if err != nil {
		return nil, err
	}

	query := `SELECT id, image_url, COALESCE(caption, ''), display_order, is_active, FALSE AS is_featured
	FROM gallery_images
	WHERE restaurant_id = $1
	ORDER BY display_order, id`
	if featuredColumnExists {
		query = `SELECT id, image_url, COALESCE(caption, ''), display_order, is_active, is_featured
	FROM gallery_images
	WHERE restaurant_id = $1
	ORDER BY display_order, id`
	}

	rows, err := r.db.Query(ctx, query, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("load gallery images: %w", err)
	}
	defer rows.Close()

	images := make([]model.GalleryImage, 0)
	for rows.Next() {
		var item model.GalleryImage
		if err := rows.Scan(&item.ID, &item.ImageURL, &item.Caption, &item.DisplayOrder, &item.IsActive, &item.IsFeatured); err != nil {
			return nil, fmt.Errorf("scan gallery image: %w", err)
		}
		images = append(images, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate gallery images: %w", err)
	}
	return images, nil
}

func (r *MenuRepository) CreateGalleryImage(ctx context.Context, payload model.GalleryImagePayload) (model.GalleryImage, error) {
	var out model.GalleryImage
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return out, err
	}

	featuredColumnExists, err := r.galleryFeaturedColumnExists(ctx)
	if err != nil {
		return out, err
	}

	var nextOrder int
	if err := r.db.QueryRow(ctx, `SELECT COALESCE(MAX(display_order), 0) + 1 FROM gallery_images WHERE restaurant_id = $1`, restaurantID).Scan(&nextOrder); err != nil {
		return out, fmt.Errorf("calculate gallery order: %w", err)
	}

	if featuredColumnExists {
		if err := r.db.QueryRow(ctx, `INSERT INTO gallery_images (restaurant_id, image_url, caption, display_order, is_active, is_featured)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, image_url, COALESCE(caption, ''), display_order, is_active, is_featured`, restaurantID, payload.ImageURL, payload.Caption, nextOrder, payload.IsActive, payload.IsFeatured).Scan(&out.ID, &out.ImageURL, &out.Caption, &out.DisplayOrder, &out.IsActive, &out.IsFeatured); err != nil {
			return out, fmt.Errorf("create gallery image: %w", err)
		}
		return out, nil
	}

	if err := r.db.QueryRow(ctx, `INSERT INTO gallery_images (restaurant_id, image_url, caption, display_order, is_active)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, image_url, COALESCE(caption, ''), display_order, is_active`, restaurantID, payload.ImageURL, payload.Caption, nextOrder, payload.IsActive).Scan(&out.ID, &out.ImageURL, &out.Caption, &out.DisplayOrder, &out.IsActive); err != nil {
		return out, fmt.Errorf("create gallery image: %w", err)
	}
	out.IsFeatured = false
	return out, nil
}

func (r *MenuRepository) UpdateGalleryImage(ctx context.Context, payload model.GalleryImagePayload) (model.GalleryImage, error) {
	var out model.GalleryImage
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return out, err
	}

	featuredColumnExists, err := r.galleryFeaturedColumnExists(ctx)
	if err != nil {
		return out, err
	}

	if featuredColumnExists {
		if err := r.db.QueryRow(ctx, `UPDATE gallery_images SET image_url = $3, caption = $4, is_active = $5, is_featured = $6, updated_at = NOW()
	WHERE id = $1 AND restaurant_id = $2
	RETURNING id, image_url, COALESCE(caption, ''), display_order, is_active, is_featured`, payload.ID, restaurantID, payload.ImageURL, payload.Caption, payload.IsActive, payload.IsFeatured).Scan(&out.ID, &out.ImageURL, &out.Caption, &out.DisplayOrder, &out.IsActive, &out.IsFeatured); err != nil {
			return out, fmt.Errorf("update gallery image: %w", err)
		}
		return out, nil
	}

	if err := r.db.QueryRow(ctx, `UPDATE gallery_images SET image_url = $3, caption = $4, is_active = $5, updated_at = NOW()
	WHERE id = $1 AND restaurant_id = $2
	RETURNING id, image_url, COALESCE(caption, ''), display_order, is_active`, payload.ID, restaurantID, payload.ImageURL, payload.Caption, payload.IsActive).Scan(&out.ID, &out.ImageURL, &out.Caption, &out.DisplayOrder, &out.IsActive); err != nil {
		return out, fmt.Errorf("update gallery image: %w", err)
	}
	out.IsFeatured = false
	return out, nil
}

func (r *MenuRepository) DeleteGalleryImage(ctx context.Context, imageID int64) error {
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `UPDATE gallery_images SET is_active = FALSE, updated_at = NOW() WHERE id = $1 AND restaurant_id = $2`, imageID, restaurantID)
	return err
}

func (r *MenuRepository) UpdateGalleryOrder(ctx context.Context, imageIDs []int64) error {
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return err
	}
	for index, imageID := range imageIDs {
		if _, err := r.db.Exec(ctx, `UPDATE gallery_images SET display_order = $1, updated_at = NOW() WHERE id = $2 AND restaurant_id = $3`, index+1, imageID, restaurantID); err != nil {
			return fmt.Errorf("update gallery order: %w", err)
		}
	}
	return nil
}

func (r *MenuRepository) galleryFeaturedColumnExists(ctx context.Context) (bool, error) {
	var exists bool
	if err := r.db.QueryRow(ctx, `SELECT EXISTS (
		SELECT 1
		FROM information_schema.columns
		WHERE table_schema = current_schema()
		AND table_name = 'gallery_images'
		AND column_name = 'is_featured'
	)`).Scan(&exists); err != nil {
		return false, fmt.Errorf("check gallery featured column: %w", err)
	}
	return exists, nil
}

func (r *MenuRepository) GetPublicReviews(ctx context.Context, limit int) ([]model.Review, error) {
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `SELECT id, author_name, rating, review_text, COALESCE(source, 'manual'), COALESCE(source_url, ''), display_order, is_featured, is_active
	FROM reviews
	WHERE restaurant_id = $1 AND is_active = TRUE AND is_featured = TRUE
	ORDER BY display_order, id`, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("load reviews: %w", err)
	}
	defer rows.Close()

	reviews := make([]model.Review, 0)
	for rows.Next() {
		var review model.Review
		if err := rows.Scan(&review.ID, &review.AuthorName, &review.Rating, &review.ReviewText, &review.Source, &review.SourceURL, &review.DisplayOrder, &review.IsFeatured, &review.IsActive); err != nil {
			return nil, fmt.Errorf("scan review: %w", err)
		}
		reviews = append(reviews, review)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reviews: %w", err)
	}

	if limit > 0 && len(reviews) > limit {
		reviews = reviews[:limit]
	}
	return reviews, nil
}

func (r *MenuRepository) GetAdminReviews(ctx context.Context) ([]model.Review, error) {
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `SELECT id, author_name, rating, review_text, COALESCE(source, 'manual'), COALESCE(source_url, ''), display_order, is_featured, is_active
	FROM reviews
	WHERE restaurant_id = $1
	ORDER BY display_order, id`, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("load reviews: %w", err)
	}
	defer rows.Close()

	reviews := make([]model.Review, 0)
	for rows.Next() {
		var review model.Review
		if err := rows.Scan(&review.ID, &review.AuthorName, &review.Rating, &review.ReviewText, &review.Source, &review.SourceURL, &review.DisplayOrder, &review.IsFeatured, &review.IsActive); err != nil {
			return nil, fmt.Errorf("scan review: %w", err)
		}
		reviews = append(reviews, review)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reviews: %w", err)
	}
	return reviews, nil
}

func (r *MenuRepository) CreateReview(ctx context.Context, payload model.ReviewPayload) (model.Review, error) {
	var out model.Review
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return out, err
	}

	var nextOrder int
	if err := r.db.QueryRow(ctx, `SELECT COALESCE(MAX(display_order), 0) + 1 FROM reviews WHERE restaurant_id = $1`, restaurantID).Scan(&nextOrder); err != nil {
		return out, fmt.Errorf("calculate review order: %w", err)
	}

	payload.Rating = normalizeReviewRating(payload.Rating)
	if err := r.db.QueryRow(ctx, `INSERT INTO reviews (restaurant_id, author_name, rating, review_text, source, source_url, display_order, is_featured, is_active)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id, author_name, rating, review_text, COALESCE(source, 'manual'), COALESCE(source_url, ''), display_order, is_featured, is_active`, restaurantID, payload.AuthorName, payload.Rating, payload.ReviewText, payload.Source, payload.SourceURL, nextOrder, payload.IsFeatured, payload.IsActive).Scan(&out.ID, &out.AuthorName, &out.Rating, &out.ReviewText, &out.Source, &out.SourceURL, &out.DisplayOrder, &out.IsFeatured, &out.IsActive); err != nil {
		return out, fmt.Errorf("create review: %w", err)
	}
	return out, nil
}

func (r *MenuRepository) UpdateReview(ctx context.Context, payload model.ReviewPayload) (model.Review, error) {
	var out model.Review
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return out, err
	}

	payload.Rating = normalizeReviewRating(payload.Rating)
	if err := r.db.QueryRow(ctx, `UPDATE reviews SET author_name = $3, rating = $4, review_text = $5, source = $6, source_url = $7, is_featured = $8, is_active = $9, updated_at = NOW()
	WHERE id = $1 AND restaurant_id = $2
	RETURNING id, author_name, rating, review_text, COALESCE(source, 'manual'), COALESCE(source_url, ''), display_order, is_featured, is_active`, payload.ID, restaurantID, payload.AuthorName, payload.Rating, payload.ReviewText, payload.Source, payload.SourceURL, payload.IsFeatured, payload.IsActive).Scan(&out.ID, &out.AuthorName, &out.Rating, &out.ReviewText, &out.Source, &out.SourceURL, &out.DisplayOrder, &out.IsFeatured, &out.IsActive); err != nil {
		return out, fmt.Errorf("update review: %w", err)
	}
	return out, nil
}

func (r *MenuRepository) DeleteReview(ctx context.Context, reviewID int64) error {
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `UPDATE reviews SET is_active = FALSE WHERE id = $1 AND restaurant_id = $2`, reviewID, restaurantID)
	return err
}

func (r *MenuRepository) UpdateReviewOrder(ctx context.Context, reviewIDs []int64) error {
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return err
	}
	for index, reviewID := range reviewIDs {
		if _, err := r.db.Exec(ctx, `UPDATE reviews SET display_order = $1 WHERE id = $2 AND restaurant_id = $3`, index+1, reviewID, restaurantID); err != nil {
			return fmt.Errorf("update review order: %w", err)
		}
	}
	return nil
}

func normalizeReviewRating(rating int) int {
	if rating < 1 {
		return 1
	}
	if rating > 5 {
		return 5
	}
	return rating
}

func (r *MenuRepository) GetPublicBanners(ctx context.Context) ([]model.Banner, error) {
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `SELECT id, title, subtitle, COALESCE(image_url, ''), COALESCE(cta_label, ''), COALESCE(cta_url, ''), display_order, to_char(starts_at, 'YYYY-MM-DD"T"HH24:MI'), to_char(ends_at, 'YYYY-MM-DD"T"HH24:MI'), is_active
	FROM banners
	WHERE restaurant_id = $1 AND is_active = TRUE
	ORDER BY display_order, id`, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("load banners: %w", err)
	}
	defer rows.Close()

	banners := make([]model.Banner, 0)
	for rows.Next() {
		var banner model.Banner
		var startsAt, endsAt *string
		if err := rows.Scan(&banner.ID, &banner.Title, &banner.Subtitle, &banner.ImageURL, &banner.CTALabel, &banner.CTAURL, &banner.DisplayOrder, &startsAt, &endsAt, &banner.IsActive); err != nil {
			return nil, fmt.Errorf("scan banner: %w", err)
		}
		if startsAt != nil {
			banner.StartsAt = *startsAt
		}
		if endsAt != nil {
			banner.EndsAt = *endsAt
		}
		banners = append(banners, banner)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate banners: %w", err)
	}
	return banners, nil
}

func (r *MenuRepository) GetAdminBanners(ctx context.Context) ([]model.Banner, error) {
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `SELECT id, title, subtitle, COALESCE(image_url, ''), COALESCE(cta_label, ''), COALESCE(cta_url, ''), display_order, to_char(starts_at, 'YYYY-MM-DD"T"HH24:MI'), to_char(ends_at, 'YYYY-MM-DD"T"HH24:MI'), is_active
	FROM banners
	WHERE restaurant_id = $1
	ORDER BY display_order, id`, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("load banners: %w", err)
	}
	defer rows.Close()

	banners := make([]model.Banner, 0)
	for rows.Next() {
		var banner model.Banner
		var startsAt, endsAt *string
		if err := rows.Scan(&banner.ID, &banner.Title, &banner.Subtitle, &banner.ImageURL, &banner.CTALabel, &banner.CTAURL, &banner.DisplayOrder, &startsAt, &endsAt, &banner.IsActive); err != nil {
			return nil, fmt.Errorf("scan banner: %w", err)
		}
		if startsAt != nil {
			banner.StartsAt = *startsAt
		}
		if endsAt != nil {
			banner.EndsAt = *endsAt
		}
		banners = append(banners, banner)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate banners: %w", err)
	}
	return banners, nil
}

func (r *MenuRepository) CreateBanner(ctx context.Context, payload model.BannerPayload) (model.Banner, error) {
	var out model.Banner
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return out, err
	}

	var nextOrder int
	if err := r.db.QueryRow(ctx, `SELECT COALESCE(MAX(display_order), 0) + 1 FROM banners WHERE restaurant_id = $1`, restaurantID).Scan(&nextOrder); err != nil {
		return out, fmt.Errorf("calculate banner order: %w", err)
	}

	if err := r.db.QueryRow(ctx, `INSERT INTO banners (restaurant_id, title, subtitle, image_url, cta_label, cta_url, display_order, starts_at, ends_at, is_active)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8::timestamptz, $9::timestamptz, $10)
	RETURNING id, title, subtitle, COALESCE(image_url, ''), COALESCE(cta_label, ''), COALESCE(cta_url, ''), display_order, to_char(starts_at, 'YYYY-MM-DD"T"HH24:MI'), to_char(ends_at, 'YYYY-MM-DD"T"HH24:MI'), is_active`, restaurantID, payload.Title, payload.Subtitle, payload.ImageURL, payload.CTALabel, payload.CTAURL, nextOrder, parseOptionalTimestamp(payload.StartsAt), parseOptionalTimestamp(payload.EndsAt), payload.IsActive).Scan(&out.ID, &out.Title, &out.Subtitle, &out.ImageURL, &out.CTALabel, &out.CTAURL, &out.DisplayOrder, &out.StartsAt, &out.EndsAt, &out.IsActive); err != nil {
		return out, fmt.Errorf("create banner: %w", err)
	}
	return out, nil
}

func (r *MenuRepository) UpdateBanner(ctx context.Context, payload model.BannerPayload) (model.Banner, error) {
	var out model.Banner
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return out, err
	}

	if err := r.db.QueryRow(ctx, `UPDATE banners SET title = $3, subtitle = $4, image_url = $5, cta_label = $6, cta_url = $7, starts_at = $8::timestamptz, ends_at = $9::timestamptz, is_active = $10, updated_at = NOW()
	WHERE id = $1 AND restaurant_id = $2
	RETURNING id, title, subtitle, COALESCE(image_url, ''), COALESCE(cta_label, ''), COALESCE(cta_url, ''), display_order, to_char(starts_at, 'YYYY-MM-DD"T"HH24:MI'), to_char(ends_at, 'YYYY-MM-DD"T"HH24:MI'), is_active`, payload.ID, restaurantID, payload.Title, payload.Subtitle, payload.ImageURL, payload.CTALabel, payload.CTAURL, parseOptionalTimestamp(payload.StartsAt), parseOptionalTimestamp(payload.EndsAt), payload.IsActive).Scan(&out.ID, &out.Title, &out.Subtitle, &out.ImageURL, &out.CTALabel, &out.CTAURL, &out.DisplayOrder, &out.StartsAt, &out.EndsAt, &out.IsActive); err != nil {
		return out, fmt.Errorf("update banner: %w", err)
	}
	return out, nil
}

func (r *MenuRepository) DeleteBanner(ctx context.Context, bannerID int64) error {
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `DELETE FROM banners WHERE id = $1 AND restaurant_id = $2`, bannerID, restaurantID)
	return err
}

func (r *MenuRepository) UpdateBannerOrder(ctx context.Context, bannerIDs []int64) error {
	restaurantID, err := r.getActiveRestaurantID(ctx)
	if err != nil {
		return err
	}
	for index, bannerID := range bannerIDs {
		if _, err := r.db.Exec(ctx, `UPDATE banners SET display_order = $1 WHERE id = $2 AND restaurant_id = $3`, index+1, bannerID, restaurantID); err != nil {
			return fmt.Errorf("update banner order: %w", err)
		}
	}
	return nil
}

func parseOptionalTimestamp(value string) string {
	if value == "" {
		return "NULL"
	}
	return value
}

func (r *MenuRepository) GetCategories(ctx context.Context) ([]model.Category, error) {
	rows, err := r.db.Query(ctx, `SELECT c.id, c.name, c.slug, c.display_order, COALESCE(c.image_url, ''), c.is_active
        FROM categories c JOIN restaurants r ON r.id = c.restaurant_id
        WHERE r.is_active = TRUE AND c.is_active = TRUE
        ORDER BY c.display_order, c.id`)
	if err != nil {
		return nil, fmt.Errorf("load categories: %w", err)
	}
	defer rows.Close()
	var categories []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.DisplayOrder, &c.ImageURL, &c.IsActive); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}
	return categories, nil
}
