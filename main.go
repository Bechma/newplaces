package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Bechma/newplaces/backend"
	"github.com/go-redis/redis/v8"
)

func main() {
	redisAddress := os.Getenv("NEWPLACES_REDIS_ADDRESS")
	if redisAddress == "" {
		log.Fatal("You need to specify a redis address")
	}
	redisPassword := os.Getenv("NEWPLACES_REDIS_PASSWORD")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	redisClient := redis.NewClient(&redis.Options{Addr: redisAddress, Password: redisPassword})
	r, err := backend.SetupRouter(redisClient)
	if err != nil {
		log.Fatal(err.Error())
	}
	// Listen and Server in 0.0.0.0:8080
	if err = r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatal(err.Error())
	}
}
