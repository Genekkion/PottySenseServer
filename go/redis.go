package main

import (
	"gopkg.in/boj/redistore.v1"
	"log"
	"net/http"
	"os"
)

type RedisStorage struct {
	store *redistore.RediStore
}

func newRedisStorage() *RedisStorage {
	store, err := redistore.NewRediStore(10, "tcp", ":6379",
		os.Getenv("REDIS_PASSWORD"), []byte(os.Getenv("REDIS_SECRET")))
	if err != nil {
		log.Println("db.go - newRedisStore()")
		log.Fatal(err)
	}
	store.SetMaxAge(86400 * 30)
	store.Options.SameSite = http.SameSiteDefaultMode
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = os.Getenv("IS_PROD") == "true"
	log.Println("RedisStorage connected successfully")

	return &RedisStorage{
		store: store,
	}
}

func (redisStorage *RedisStorage) close() {
	redisStorage.store.Close()
}
