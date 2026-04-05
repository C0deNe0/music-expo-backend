package repository

import (
	"context"

	"github.com/C0deNe0/otify/internal/model"
	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	client *redis.Client
}

func NewRedisRepo(client *redis.Client) *RedisRepo {
	return &RedisRepo{client: client}
}

func (r *RedisRepo) CreateJob(ctx context.Context, job model.Job) {
	r.client.HSet(ctx, "job:"+job.ID, map[string]interface{}{
		"url":     job.URL,
		"name":    job.Name,
		"status":  string(job.Status),
		"file_id": "",
		"error":   "",
	})
}

func (r *RedisRepo) GetJob(ctx context.Context, id string) (model.Job, error) {
	data, err := r.client.HGetAll(ctx, "job:"+id).Result()
	if err != nil {
		return model.Job{}, err
	}

	return model.Job{
		ID:     id,
		URL:    data["url"],
		Name:   data["name"],
		Status: model.JobStatus(data["status"]),
		FileID: data["file_id"],
		Error:  data["error"],
	}, nil
}

func (r *RedisRepo) UpdateJob(ctx context.Context, id string, updates map[string]interface{}) error {
	return r.client.HSet(ctx, "job:"+id, updates).Err()
}

func (r *RedisRepo) Enqueue(ctx context.Context, jobID string) error {
	return r.client.LPush(ctx, "queue", jobID).Err()
}

func (r *RedisRepo) Dequeue(ctx context.Context) (string, error) {
	res, err := r.client.BRPop(ctx, 0, "queue").Result()
	if err != nil {
		return "", err
	}
	return res[1], nil
}
