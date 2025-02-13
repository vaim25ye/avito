package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/vaim25ye/avito/internal/handler"
	"github.com/vaim25ye/avito/internal/repository"
)

func main() {
	// Можно брать строку подключения из переменной окружения
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5432/avito?sslmode=disable"
		// Замените user, password, dbname на свои
	}

	repo, err := repository.NewRepository(dsn)
	if err != nil {
		log.Fatalf("failed to init repository: %v", err)
	}

	h := handler.NewHandler(repo)

	mux := http.NewServeMux()

	mux.HandleFunc("/users", h.CreateUser)       // POST
	mux.HandleFunc("/get_user", h.GetUserByID)   // GET ?id=1
	mux.HandleFunc("/transfer", h.Transfer)      // POST
	mux.HandleFunc("/purchase", h.PurchaseMerch) // POST

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	log.Println("Starting server on :8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
