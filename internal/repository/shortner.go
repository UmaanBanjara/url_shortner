package repository

import (
	"time"

	"urlshortner/internal/db"
)

type URL struct {
	ID          int64     `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
}

type ShortenRequest struct {
	URL string `json:"url"`
}

func CreateURL(shortCode string, originalURL string) (int64, error) {
	var id int64
	query := `
		INSERT INTO urls (short_code, original_url)
		VALUES ($1, $2)
		RETURNING id
	`
	err := db.DB.QueryRow(query, shortCode, originalURL).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetURLByShortCode(shortCode string) (URL, error) {
	var u URL
	query := `
		SELECT id, short_code, original_url, created_at
		FROM urls
		WHERE short_code = $1
	`
	err := db.DB.QueryRow(query, shortCode).Scan(
		&u.ID,
		&u.ShortCode,
		&u.OriginalURL,
		&u.CreatedAt,
	)
	if err != nil {
		return URL{}, err
	}
	return u, nil
}

func InsertClick(shortCode, userAgent, ipAddress string) error {
	var urlID int64
	err := db.DB.QueryRow(
		"SELECT id FROM urls WHERE short_code = $1",
		shortCode,
	).Scan(&urlID)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO clicks (url_id, user_agent, ip_address)
		VALUES ($1, $2, $3)
	`
	_, err = db.DB.Exec(query, urlID, userAgent, ipAddress)
	return err
}
