package main

import (
	"encoding/json"
	"log"
	"urlshortner/internal/db"
	"urlshortner/internal/redis_client"
	"urlshortner/internal/repository"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("error loading .env")
	}

	db.Connect()
	redis_client.Connect()

	log.Println("worker started , waiting for the jobs .........")

	for {
		result, err := redis_client.RDB.BRPop(redis_client.Ctx, 0, "click_events").Result()
		if err != nil {
			log.Println("error reading from the queue", err)
			continue
		}

		var event redis_client.ClickEvent

		if err := json.Unmarshal([]byte(result[1]), &event); err != nil {
			log.Println("failed to unmarshall event", err)
			continue
		}

		err = repository.InsertClick(event.ShortCode, event.UserAgent, event.IPAddress)
		if err != nil {
			log.Println("failed to insert click", err)
			continue
		}

		log.Println("proccessed click events for", event.ShortCode)

	}
}
