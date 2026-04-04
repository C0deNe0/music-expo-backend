package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	logger.Info("worker started")

	for {
		// wait for job
		res, err := rdb.BRPop(ctx, 0*time.Second, "queue").Result()
		if err != nil {
			logger.Error("redis error", "err", err)
			continue
		}

		jobID := res[1]
		logger.Info("processing job", "job_id", jobID)

		processJob(rdb, logger, jobID)
	}
}

func processJob(rdb *redis.Client, logger *slog.Logger, jobID string) {
	key := "job:" + jobID

	// mark processing
	rdb.HSet(ctx, key, "status", "processing")

	jobData, err := rdb.HGetAll(ctx, key).Result()
	if err != nil {
		logger.Error("failed to get job", "err", err)
		return
	}

	url := jobData["url"]
	name := jobData["name"]

	if name == "" {
		name = jobID
	}

	output := fmt.Sprintf("downloads/%s.%%(ext)s", name)

	cmd := exec.Command(
		"yt-dlp",
		"-x",
		"--audio-format", "mp3",
		"--audio-quality", "0",
		"-o", output,
		url,
	)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		logger.Error("yt-dlp failed",
			"err", err,
			"stderr", stderr.String(),
		)

		rdb.HSet(ctx, key,
			"status", "failed",
			"error", stderr.String(),
		)
		return
	}

	fileName := name + ".mp3"

	rdb.HSet(ctx, key,
		"status", "done",
		"file", fileName,
	)

	logger.Info("job done", "job_id", jobID)
}
