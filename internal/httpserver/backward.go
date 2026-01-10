package httpserver

import (
	"log"
	"net/http"

	"github.com/gocql/gocql"
	"github.com/redis/go-redis/v9"
)

func BackwardHandler(rdb *redis.Client, session *gocql.Session) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Backward handler called")

	}
}
