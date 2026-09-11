package service

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jakaria9001/tablebite/internal/model"
)

func TestMenuSeederValidatesRequiredDataBeforeDatabaseAccess(t *testing.T) {
	seeder := NewMenuSeeder(nil)
	for name, data := range map[string]model.SeedFile{
		"missing restaurant": {Categories: []model.SeedCategory{{Name: "Coffee"}}, MenuItems: []model.SeedMenuItem{{Name: "Tea"}}},
		"missing categories": {RestaurantName: "Test Restaurant", MenuItems: []model.SeedMenuItem{{Name: "Tea"}}},
		"missing menu items": {RestaurantName: "Test Restaurant", Categories: []model.SeedCategory{{Name: "Coffee"}}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, err := seeder.Seed(context.Background(), data); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	for input, want := range map[string]string{
		" Indian Restaurant & Sweets ": "indian-restaurant-sweets",
		"Soup Menu":                    "soup-menu",
		"":                             "",
	} {
		if got := slugify(input); got != want {
			t.Fatalf("slugify(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestSeedFixtureImportsExpectedMenuShape(t *testing.T) {
	path := filepath.Join("..", "..", "seed", "menu.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read seed fixture: %v", err)
	}

	var seed model.SeedFile
	if err := json.Unmarshal(data, &seed); err != nil {
		t.Fatalf("decode seed fixture: %v", err)
	}
	if seed.RestaurantName == "" || len(seed.Categories) == 0 || len(seed.MenuItems) == 0 {
		t.Fatal("seed fixture must include restaurant, categories, and menu items")
	}

	var hasVeg, hasHalfFull bool
	for _, item := range seed.MenuItems {
		hasVeg = hasVeg || item.IsVeg
		hasHalfFull = hasHalfFull || item.PriceHalf != nil && item.PriceFull != nil
	}
	if !hasVeg || !hasHalfFull {
		t.Fatal("seed fixture must exercise veg and half/full menu data")
	}
}
