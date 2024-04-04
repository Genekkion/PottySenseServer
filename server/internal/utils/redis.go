package utils

import (
	"log"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
	"gopkg.in/boj/redistore.v1"
)

func NewRedisSessionStore() *redistore.RediStore {
	store, err := redistore.NewRediStore(10, "tcp",
		os.Getenv("REDIS_ADDR"),
		os.Getenv("REDIS_PASSWORD"),
		[]byte(os.Getenv("REDIS_SECRET")),
	)

	if err != nil {
		log.Println("redis.go - newRedisStore()")
		log.Fatal(err)
	}
	store.SetMaxAge(86400 * 30)
	store.Options.SameSite = http.SameSiteDefaultMode
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = os.Getenv("IS_PROD") == "true"
	log.Println("RedisStorage connected successfully")

	return store
}

func NewRedisStorage() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	return client
}
