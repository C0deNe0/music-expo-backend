package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb *redis.Client

type ExtractRequest struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	http.HandleFunc("/extract", extractHandler(logger))
	http.HandleFunc("/job", jobHandler(logger))

	logger.Info("API running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func extractHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ExtractRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		u, err := url.Parse(req.URL)
		if err != nil || u.Host == "" {
			http.Error(w, "invalid url", http.StatusBadRequest)
			return
		}

		jobID := uuid.New().String()

		err = rdb.HSet(ctx, "job:"+jobID, map[string]any{
			"url":        req.URL,
			"name":       req.Name,
			"status":     "pending",
			"file_id":    "",
			"error":      "",
			"created_at": time.Now().String(),
		}).Err()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = rdb.LPush(ctx, "queue", jobID).Err()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Info("job created", "job_id", jobID)

		json.NewEncoder(w).Encode(map[string]string{
			"job_id": jobID,
		})
	}
}
func jobHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobID := r.URL.Query().Get("id")
		if jobID == "" {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}

		data, err := rdb.HGetAll(ctx, "job:"+jobID).Result()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(data) == 0 {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(data)
	}
}
