package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/jakaria9001/tablebite/internal/database"
	"github.com/jakaria9001/tablebite/internal/handler"
	"github.com/jakaria9001/tablebite/internal/repository"
	"github.com/jakaria9001/tablebite/internal/router"
)

func main() {
	loadDotEnv()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.NewPool(ctx)
	if err != nil {
		slog.Error("database startup failed", "error", err)
		return
	}
	defer db.Close()

	menuRepo := repository.NewMenuRepository(db)
	menuHandler := handler.NewMenuHandler(menuRepo)
	adminRepo := repository.NewAdminRepository(db)
	authHandler := handler.NewAuthHandler(adminRepo)
	mux := router.New(handler.Health, menuHandler, authHandler, getenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173"))

	port := getenv("PORT", "8080")
	srv := &http.Server{Addr: ":" + port, Handler: mux, ReadHeaderTimeout: 5 * time.Second}

	go func() {
		slog.Info("TableBite API listening", "address", "http://localhost:"+port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown failed", "error", err)
	} else {
		slog.Info("server stopped")
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func loadDotEnv() {
	candidates := []string{".env", filepath.Join("..", ".env"), filepath.Join("..", "..", ".env")}
	for _, candidate := range candidates {
		if err := loadEnvFile(candidate); err == nil {
			return
		}
	}
}

func loadEnvFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if _, exists := os.LookupEnv(key); !exists && value != "" {
			_ = os.Setenv(key, value)
		}
	}

	return nil
}
