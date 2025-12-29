package main

import (
	"context"
	"log"
	"os"
	"net/http"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/joho/godotenv"

	"shortenit/internal/httpserver"
)

func main() {

	ctx := context.Background()

	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize Redis client
	redisAddr := os.Getenv("REDIS_ADDR")
	redisUsername := os.Getenv("REDIS_USERNAME")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB_NO"))
	if err != nil {
		log.Fatalf("Error converting REDIS_DB to int: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Username: redisUsername,
		Password: redisPassword,
		DB:       redisDB,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}

	redisDefaultCount := os.Getenv("REDIS_COUNTER_INIT")
	
	if currentCount, err := rdb.Get(ctx, "currentCount").Int() ;err == nil {
		log.Printf("current count: %d", currentCount)
	} else {
		if err := rdb.Set(ctx, "currentCount", redisDefaultCount, 0).Err(); err != nil {
			log.Fatalf("failed to set current count: %v", err)
		}
		log.Printf("set current count to %s", redisDefaultCount)
	}

	// Intializing the db cluster

	scyllaNode0 := os.Getenv("SCYLLA_NODE_0")
	scyllaNode1 := os.Getenv("SCYLLA_NODE_1")
	scyllaNode2 := os.Getenv("SCYLLA_NODE_2")
	scyllaUsername := os.Getenv("SCYLLA_USERNAME")
	scyllaPassword := os.Getenv("SCYLLA_PASSWORD")
	scyllaRegion := os.Getenv("SCYLLA_REGION")

	cluster := gocql.NewCluster(scyllaNode0, scyllaNode1, scyllaNode2)

    cluster.Authenticator = gocql.PasswordAuthenticator{Username: scyllaUsername, Password: scyllaPassword}
	cluster.PoolConfig.HostSelectionPolicy = gocql.DCAwareRoundRobinPolicy(scyllaRegion)

	session, err := gocqlx.WrapSession(cluster.CreateSession())

	if err != nil {
		panic("DB connection fail")
	}

	// Initializing the actual router

	router := httpserver.NewRouter(rdb, session)
	log.Println("shorten.it API listening on :8080 (Redis connected)")

	log.Fatal(http.ListenAndServe(":8080", router))
}


