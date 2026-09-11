package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/jakaria9001/tablebite/internal/database"
	"github.com/jakaria9001/tablebite/internal/model"
	"github.com/jakaria9001/tablebite/internal/service"
)

func main() {
	file := getenv("SEED_MENU_FILE", "./seed/menu.json")
	raw, err := os.ReadFile(file)
	if err != nil {
		log.Fatalf("read seed file: %v", err)
	}

	var data model.SeedFile
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Fatalf("parse seed file: %v", err)
	}

	ctx := context.Background()
	db, err := database.NewPool(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	cats, items, err := service.NewMenuSeeder(db).Seed(ctx, data)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("seed complete: %d categories, %d menu items", cats, items)
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
