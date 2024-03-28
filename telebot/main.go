package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func NewRedisStorage(redisAddr string, redisPassword string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})
}

var (
	REQUIRED_ENV = []string{
		"TELEGRAM_BOT_TOKEN",
		"REDIS_ADDR",
		"REDIS_PASSWORD",
	}
)

func main() {
	godotenv.Load(".env")

	for _, env := range REQUIRED_ENV {
		if os.Getenv(env) == "" {
			log.Fatalln("Required env variable\"" + env + "\" not set. Exiting.")
		}
	}

	telegramBotToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	db := NewSqliteStorage("../sqlite.db")
	redisCache := NewRedisStorage(redisAddr, redisPassword)
	err := redisCache.Ping(context.Background()).Err()
	if err != nil {
		log.Fatalln(err)
	}
	bot := NewBot(telegramBotToken, db, redisCache)
	bot.Run()
}
