package main

import (
	"fmt"
	"log"
	"net/http"
	"urlshortner/internal/db"
	"urlshortner/internal/redis_client"

	"urlshortner/internal/handler"

	"github.com/go-chi/chi/v5"
	middleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("failed to load .env")
	}

	db.Connect()
	redis_client.Connect()

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("URL Shortner running..."))
	})

	r.Post("/shorten", handler.CreateURL)
	r.Get("/{code}", handler.RedirectHandler)

	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", r)
}
