package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/alextanhongpin/base62"
	"github.com/gocql/gocql"
	"github.com/redis/go-redis/v9"

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

		exists, err := rdb.Get(ctx, "forward:"+request.LongURL).Result()
		if err != nil && err != redis.Nil {
			fmt.Println("failed to check if long URL exists", err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"status":"error","error":"failed to check if long URL exists"}`))
			return
		}
		if exists != "" {
			shortURL := strings.TrimPrefix(exists, "forward:")
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

		currentCount, err := rdb.Incr(ctx, "currentCount").Result()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"status":"error","error":"failed to get current count"}`))
			return
		}

		b62 := base62.New(base62.DEFAULT_CHARS)
		shortURL := b62.Encode(uint64(currentCount))

		err = rdb.Set(ctx, "forward:"+request.LongURL, shortURL, 0).Err()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"status":"error","error":"failed to set short URL"}`))
			return
		}

		err = rdb.Set(ctx, "backward:"+shortURL, request.LongURL, 0).Err()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"status":"error","error":"failed to set long URL"}`))
			return
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
