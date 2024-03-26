package main

import (
	"log"
	"os"

	"github.com/genekkion/PottySenseServer/internal"
	"github.com/genekkion/PottySenseServer/internal/globals"
	"github.com/genekkion/PottySenseServer/internal/utils"
)

func main() {
	globals.RUN = true
	utils.SetEnv("./.env")

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
