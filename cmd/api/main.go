package main

import (
	"os"

	"github.com/C0deNe0/otify/internal/handler"
	"github.com/C0deNe0/otify/internal/repository"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("DB_NAME")

	redisOpts := &redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Username: os.Getenv("REDIS_USER"),
		Password: os.Getenv("REDIS_PASSWORD"),
	}
	rdb := redis.NewClient(redisOpts)


	redisRepo := repository.NewRedisRepo(rdb)
	mongoRepo, _ := repository.NewMongoRepo(mongoURI, dbName)

	h := handler.NewHandler(redisRepo, mongoRepo)

	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(200, "OK")
	})

	e.POST("/extract", h.Extract)
	e.GET("/job", h.GetJob)
	e.GET("/download", h.Download)

	e.Start(":" + port)
}
