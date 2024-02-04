package main

import "log"

var RUN bool

func main() {
	RUN = true
	setEnv()

	redisSessionStore := newRedisSessionStore()
	defer redisSessionStore.close()

	dbStorage := newSqliteStorage()
	defer dbStorage.close()

	parseFlags(*dbStorage)

	redisStorage := newRedisStorage()
	defer redisStorage.close()

	if !RUN {
		log.Println("Exiting program.")
		return
	}
	server := initServer(dbStorage, redisSessionStore, redisStorage)
	server.run()
}
