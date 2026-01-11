package httpserver

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gocql/gocql"
	"github.com/redis/go-redis/v9"

	"shortenit/internal/models"
)

type BackwardRequest struct {
	ShortURL string `json:"short_url"`
}

type BackwardResponse struct {
	LongURL string `json:"long_url"`
}

func BackwardHandler(rdb *redis.Client, session *gocql.Session) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		// parsing the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"status":"error","error":"failed to read body"}`))
			return
		}

		var request BackwardRequest
		if err := json.Unmarshal(body, &request); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"status":"error","error":"failed to unmarshal body"}`))
			return
		}

		// checking if the short URL exists in the redis
		longURL, err := rdb.Get(ctx, "backward:"+request.ShortURL).Result()
		if err != nil && err != redis.Nil {
			log.Println("failed to check if short URL exists in redis", err)
		} else if longURL != "" {
			response := BackwardResponse{
				LongURL: longURL,
			}
			jsonResponse, err := json.Marshal(response)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"status":"error","error":"failed to marshal response"}`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(jsonResponse)
			return
		}

		// if redis miss, check if the short URL exists in the cql db
		dbLongURL, err := models.FetchLongURL(session, request.ShortURL)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"status":"error","error":"failed to fetch long URL"}`))
			return
		} else if dbLongURL != "" {
			response := BackwardResponse{
				LongURL: dbLongURL,
			}
			jsonResponse, err := json.Marshal(response)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"status":"error","error":"failed to marshal response"}`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(jsonResponse)
			return
		}

		// if both redis and cql db miss, return not found
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"not_found"}`))
	}
}
