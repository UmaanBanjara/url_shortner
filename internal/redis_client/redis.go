package redis_client

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client
var Ctx = context.Background()

func Connect() {
	err := godotenv.Load()
	if err != nil {
		return
	}

	RDB = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	if err = RDB.Ping(Ctx).Err(); err != nil {
		log.Fatal("unable to connect to redis", err)
	}

	log.Println("redis connected successfully")

}
