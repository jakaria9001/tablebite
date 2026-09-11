package main

import (
	"context"
	"log"
	"os"

	"github.com/jakaria9001/tablebite/internal/database"
	"github.com/jakaria9001/tablebite/internal/migrate"
)

func main() {
	ctx := context.Background()
	db, err := database.NewPool(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dir := os.Getenv("MIGRATIONS_DIR")
	if dir == "" {
		dir = "./migrations"
	}
	if err := migrate.Up(ctx, db, dir); err != nil {
		log.Fatal(err)
	}
}
