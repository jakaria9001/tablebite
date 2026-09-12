package main

import (
	"context"
	"log"
	"os"

	"github.com/jakaria9001/tablebite/internal/auth"
	"github.com/jakaria9001/tablebite/internal/database"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	email := auth.NormalizeEmail(os.Getenv("ADMIN_EMAIL"))
	password := os.Getenv("ADMIN_PASSWORD")
	if email == "" || password == "" {
		log.Fatal("set ADMIN_EMAIL and ADMIN_PASSWORD before creating an admin")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.NewPool(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(context.Background(), `
        INSERT INTO admins(email, password_hash, role, is_active)
        VALUES ($1, $2, 'admin', TRUE)
        ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash, is_active = TRUE, updated_at = NOW()`, email, string(hash))
	if err != nil {
		log.Fatalf("create admin: %v", err)
	}

	log.Printf("admin account ready: %s", email)
}
