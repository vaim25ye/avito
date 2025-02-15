package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/vaim25ye/avito/internal/cache"
	"github.com/vaim25ye/avito/internal/handler"
	"github.com/vaim25ye/avito/internal/repository"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:1234@localhost:5432/avito?sslmode=disable"
	}

	repo, err := repository.NewRepository(dsn)
	if err != nil {
		log.Fatalf("failed to init repository: %v", err)
	}

	c := cache.NewCache()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cache.StartCacheUpdater(ctx, repo, c, 15*time.Second)

	h := handler.NewHandler(repo, c)

	mux := http.NewServeMux()
	mux.HandleFunc("/users", h.CreateUser)
	mux.HandleFunc("/get_user", h.GetUserByID)
	mux.HandleFunc("/transfer", h.Transfer)
	mux.HandleFunc("/purchase", h.PurchaseMerch)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
