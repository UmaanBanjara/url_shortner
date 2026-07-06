package handler

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"urlshortner/internal/redis_client"
	"urlshortner/internal/repository"
	"urlshortner/internal/shortner"

	"github.com/go-chi/chi/v5"
)

func CreateURL(w http.ResponseWriter, r *http.Request) {
	var req repository.ShortenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	code, err := shortner.GenerateCode()
	if err != nil {
		http.Error(w, "failed to generate code", http.StatusInternalServerError)
		return
	}

	_, err = repository.CreateURL(code, req.URL)
	if err != nil {
		http.Error(w, "failed to save url", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":      code,
		"short_url": "http://localhost:8080/" + code,
	})
}

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	u, err := repository.GetURLByShortCode(code)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "short code not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	event := redis_client.ClickEvent{
		ShortCode: code,
		UserAgent: r.UserAgent(),
		IPAddress: r.RemoteAddr,
	}

	if err := redis_client.PushClickEvent(event); err != nil {
		log.Println(w, "failed to push click evets", err)
		return
	}
	http.Redirect(w, r, u.OriginalURL, http.StatusFound)
}
