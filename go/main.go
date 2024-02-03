package main

import (
	"log"
	"os"
)

func parseArgs(dbStorage *SqliteStorage) int {
	args := os.Args[1:]
	length := len(args)
	if length == 0 {
		return 0
	}

	i := 0
	for i < length {
		if args[i] == "-a" {
			if i+2 >= length {
				log.Println("Wrong number of arguments supplied for '-a' command")
				return 1
			}
			dbStorage.createAdmin(args[i+1], args[i+2])
			i += 2
			continue
		}
		i++
	}
	return 1
}

func main() {
	setEnv()

	redisSessionStore := newRedisSessionStore()
	defer redisSessionStore.close()

	dbStorage := newSqliteStorage()
	defer dbStorage.close()

	if parseArgs(dbStorage) != 0 {
		return
	}

    redisStorage := newRedisStorage()
    defer redisStorage.close()

	server := initServer(dbStorage, redisSessionStore, redisStorage)
	server.run()
}
