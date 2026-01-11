package httpserver

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/alextanhongpin/base62"
	"github.com/gocql/gocql"
	"github.com/redis/go-redis/v9"

	"shortenit/internal/models"
	"shortenit/internal/utils"
)

type ForwardRequest struct {
	LongURL string `json:"long_url"`
}

type ForwardResponse struct {
	ShortURL string `json:"short_url"`
}

func ForwardHandler(rdb *redis.Client, session *gocql.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		// parsing the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"status":"error","error":"failed to read body"}`))
			return
		}

		var request ForwardRequest
		if err := json.Unmarshal(body, &request); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"status":"error","error":"failed to unmarshal body"}`))
			return
		}

		// validating the long url
		// TODO: explore validation libraries
		if request.LongURL == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"status":"error","error":"long_url is required"}`))
			return
		}

		if !utils.ValidateLongURL(request.LongURL) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"status":"error","error":"invalid long_url"}`))
			return
		}

		// checking if the long url exists in the redis
		redisExists, err := rdb.Get(ctx, "forward:"+request.LongURL).Result()
		if err != nil && err != redis.Nil {
			log.Println("failed to check if long URL exists in redis", err)
		} else if redisExists != "" {
			// redis hit, return the short url
			shortURL := strings.TrimPrefix(redisExists, "forward:")
			response := ForwardResponse{
				ShortURL: shortURL,
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

		// checking if the long url exists in the cql db
		dbShortURL, err := models.FetchShortURL(session, request.LongURL)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"status":"error","error":"failed to fetch short URL"}`))
			return
		} else if dbShortURL != "" {
			response := ForwardResponse{
				ShortURL: dbShortURL,
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

		// If both redis and cql db miss, we need to generate a new short URL
		currentCount, err := rdb.Incr(ctx, "currentCount").Result()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"status":"error","error":"failed to get current count"}`))
			return
		}

		b62 := base62.New(base62.DEFAULT_CHARS)
		shortURL := b62.Encode(uint64(currentCount))

		err = models.InsertNewURL(session, request.LongURL, shortURL)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"status":"error","error":"failed to insert new URL"}`))
			return
		}
		err = models.SetNewURLRedis(rdb, request.LongURL, shortURL)
		if err != nil {
			log.Println("failed to set new URL in redis", err)
		}

		response := ForwardResponse{
			ShortURL: shortURL,
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
	}
}
