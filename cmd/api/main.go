package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	gocqlastra "github.com/datastax/gocql-astra"
	"github.com/gocql/gocql"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

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

	if currentCount, err := rdb.Get(ctx, "currentCount").Int(); err == nil {
		log.Printf("current count: %d", currentCount)
	} else {
		if err := rdb.Set(ctx, "currentCount", redisDefaultCount, 0).Err(); err != nil {
			log.Fatalf("failed to set current count: %v", err)
		}
		log.Printf("set current count to %s", redisDefaultCount)
	}

	// Intializing the db cluster
	astraDBID := os.Getenv("ASTRA_DB_ID")
	astraDBRegion := os.Getenv("ASTRA_DB_REGION")
	astraDBKeyspace := os.Getenv("KEYSPACE_NAME")
	astraDBURL := os.Getenv("ASTRA_DB_URL")
	astraDBToken := os.Getenv("ASTRA_DB_TOKEN")

	if astraDBID == "" || astraDBRegion == "" || astraDBKeyspace == "" || astraDBURL == "" || astraDBToken == "" {
		panic("ASTRA_DB_ID, ASTRA_DB_REGION, KEYSPACE_NAME, ASTRA_DB_ENDPOINT, or ASTRA_DB_TOKEN is not set")
	}
	cluster, err := gocqlastra.NewClusterFromURL(astraDBURL, astraDBID, astraDBToken, 10*time.Second)

	if err != nil {
		log.Fatalf("unable to load cluster from astra: %v", err)
	}

	cluster.Timeout = 30 * time.Second
	//setting the session to use the keyspace
	cluster.Keyspace = astraDBKeyspace
	session, err := gocql.NewSession(*cluster)

	if err != nil {
		log.Fatalf("unable to connect session: %v", err)
	}

	log.Println("Checking if the database is connected")

	iter := session.Query("SELECT release_version FROM system.local").Iter()

	var version string

	for iter.Scan(&version) {
		log.Println(version)
	}

	if err = iter.Close(); err != nil {
		log.Fatalf("error running query: %v", err)
	}

	// Initializing the actual router
	router := httpserver.NewRouter(rdb, session)
	log.Println("shorten.it API listening on :8080 (Redis connected)")

	log.Fatal(http.ListenAndServe(":8080", router))
	defer rdb.Close()
	defer session.Close()
}
