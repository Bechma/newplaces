package main

import (
	"flag"
	"log"

	"github.com/Bechma/newplaces/backend"
	"github.com/go-redis/redis/v8"
)

var (
	reset        = flag.Bool("reset", false, "Reset the canvas to all white")
	redisAddress = flag.String("redis", "127.0.0.1:6379", "redis address")
)

func main() {
	flag.Parse()
	redisClient := redis.NewClient(&redis.Options{Addr: *redisAddress})
	if *reset {
		backend.ResetCanvas(redisClient)
		return
	}
	r, err := backend.SetupRouter(redisClient)
	if err != nil {
		log.Fatal(err.Error())
	}
	// Listen and Server in 0.0.0.0:8080
	if err = r.Run(":8080"); err != nil {
		log.Fatal(err.Error())
	}
}
