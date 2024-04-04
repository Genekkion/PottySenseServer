package main

import (
	"log"
	"os"

	"github.com/genekkion/PottySenseServer/internal"
	"github.com/genekkion/PottySenseServer/internal/globals"
	"github.com/genekkion/PottySenseServer/internal/utils"
	"github.com/joho/godotenv"
)

var (
	requiredEnv = []string{

		"SERVER_ADDR",
		"DATABASE_PATH",
		"CSRF_SECRET",
		"GORILLA_SESSION_SECRET",
		"TELEGRAM_BOT_TOKEN",
		"REDIS_ADDR",
		"REDIS_SECRET",
		"SECRET_HEADER",
	}
)

func main() {
	globals.RUN = true
	godotenv.Load("../.env")

	for _, env := range requiredEnv {
		if os.Getenv(env) == "" {
			log.Fatalln("Required environment variable \"" + env + "\" not set. Exiting.")
		}
	}

	// globals.TOILETS_URL[1] = os.Getenv("PI_ADDR")

	redisSessionStore := utils.NewRedisSessionStore()
	defer redisSessionStore.Close()

	dbStorage := utils.NewSqliteStorage(os.Getenv("DATABASE_PATH"))
	defer dbStorage.Close()

	internal.ParseFlags(dbStorage)

	redisStorage := utils.NewRedisStorage()
	defer redisStorage.Close()

	if !globals.RUN {
		log.Println("Exiting program.")
		return
	}
	server := internal.InitServer(dbStorage, redisSessionStore, redisStorage)
	server.Run()
}
