package handler

import (
	"context"
	"net/url"

	"github.com/C0deNe0/otify/internal/model"
	"github.com/C0deNe0/otify/internal/repository"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	redis *repository.RedisRepo
	mongo *repository.MongoRepo
}

func NewHandler(r *repository.RedisRepo, m *repository.MongoRepo) *Handler {
	return &Handler{redis: r, mongo: m}
}

type ExtractRequest struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

func (h *Handler) Extract(c echo.Context) error {
	ctx := context.Background()

	var req ExtractRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, map[string]string{"error": err.Error()})
	}

	u, err := url.Parse(req.URL)
	if err != nil || u.Host == "" {
		return c.JSON(400, map[string]string{"error": "invalid url"})
	}

	jobID := uuid.New().String()
	if req.Name == "" {
		req.Name = jobID
	}

	job := model.Job{
		ID:     jobID,
		URL:    req.URL,
		Name:   req.Name,
		Status: model.StatusPending,
	}

	h.redis.CreateJob(ctx, job)
	h.redis.Enqueue(ctx, jobID)

	return c.JSON(200, map[string]string{"job_id": jobID})
}

func (h *Handler) GetJob(c echo.Context) error {
	ctx := context.Background()

	id := c.QueryParam("id")
	job, _ := h.redis.GetJob(ctx, id)

	return c.JSON(200, job)
}

func (h *Handler) Download(c echo.Context) error {
	ctx := context.Background()

	id := c.QueryParam("id")
	job, _ := h.redis.GetJob(ctx, id)

	c.Response().Header().Set("Content-Type", "audio/mpeg")
	return h.mongo.DownloadFile(ctx, job.Name+".mp3", c.Response().Writer)
}
