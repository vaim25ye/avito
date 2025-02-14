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

	// Инициализируем репозиторий (подключаемся к БД)
	repo, err := repository.NewRepository(dsn)
	if err != nil {
		log.Fatalf("failed to init repository: %v", err)
	}

	// Инициализируем кэш
	c := cache.NewCache()

	// Запускаем горутину, которая каждые 15 секунд обновляет кэш
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cache.StartCacheUpdater(ctx, repo, c, 15*time.Second)

	// Создаём хендлер (передаём repo и cache)
	h := handler.NewHandler(repo, c)

	// Маршруты
	mux := http.NewServeMux()
	mux.HandleFunc("/users", h.CreateUser)       // POST
	mux.HandleFunc("/get_user", h.GetUserByID)   // GET ?id=1
	mux.HandleFunc("/transfer", h.Transfer)      // POST
	mux.HandleFunc("/purchase", h.PurchaseMerch) // POST

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
