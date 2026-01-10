package httpserver

import (
	"context"
	"net/http"

	"github.com/redis/go-redis/v9"
	"github.com/gocql/gocql"
)

func NewRouter(rdb *redis.Client, session *gocql.Session) http.Handler {
	ctx := context.Background()
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"service":"shorten.it","status":"ok"}`))
	})

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := rdb.Ping(ctx).Err(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unhealthy","error":"redis_unreachable"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	})

	mux.Handle("POST /forward", ForwardHandler(rdb, session))

	return mux
}