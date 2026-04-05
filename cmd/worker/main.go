package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/C0deNe0/otify/internal/model"
	"github.com/C0deNe0/otify/internal/repository"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func main() {
	fmt.Println("🚀 Worker started")

	redisOpts := &redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Username: os.Getenv("REDIS_USER"),
		Password: os.Getenv("REDIS_PASSWORD"),
	}
	rdb := redis.NewClient(redisOpts)


	redisRepo := repository.NewRedisRepo(rdb)

	mongoRepo, err := repository.NewMongoRepo(
		os.Getenv("MONGO_URI"),
		os.Getenv("DB_NAME"),
	)
	if err != nil {
		fmt.Println("❌ Mongo connection failed:", err)
		return
	}

	fmt.Println("✅ Mongo connected")

	for {
		jobID, err := redisRepo.Dequeue(ctx)
		if err != nil {
			fmt.Println("❌ dequeue error:", err)
			continue
		}

		fmt.Println("📦 Got job:", jobID)

		err = redisRepo.UpdateJob(ctx, jobID, map[string]interface{}{
			"status": string(model.StatusProcessing),
		})
		if err != nil {
			fmt.Println("❌ update error:", err)
		}

		job, err := redisRepo.GetJob(ctx, jobID)
		if err != nil {
			fmt.Println("❌ get job error:", err)
			continue
		}

		file := job.Name + ".mp3"

		fmt.Println("⬇️ Downloading:", job.URL)

		cmd := exec.Command("yt-dlp", "-x", "--audio-format", "mp3", "-o", file, job.URL)

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err = cmd.Run()
		if err != nil {
			fmt.Println("❌ yt-dlp failed:", stderr.String())

			redisRepo.UpdateJob(ctx, jobID, map[string]interface{}{
				"status": string(model.StatusFailed),
				"error":  stderr.String(),
			})
			continue
		}

		fmt.Println("✅ Downloaded:", file)

		fileID, err := mongoRepo.UploadFile(ctx, file, file)
		if err != nil {
			fmt.Println("❌ Mongo upload failed:", err)
			continue
		}

		fmt.Println("✅ Uploaded to Mongo:", fileID)

		err = redisRepo.UpdateJob(ctx, jobID, map[string]interface{}{
			"status":  string(model.StatusDone),
			"file_id": fileID,
		})
		if err != nil {
			fmt.Println("❌ update error:", err)
		}

		os.Remove(file)
		fmt.Println("🧹 cleaned file")
	}
}
