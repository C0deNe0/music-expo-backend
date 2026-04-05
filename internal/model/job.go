package model

type JobStatus string

const (
	StatusPending    JobStatus = "pending"
	StatusProcessing JobStatus = "processing"
	StatusDone       JobStatus = "done"
	StatusFailed     JobStatus = "failed"
)

type Job struct {
	ID     string    `json:"id"`
	URL    string    `json:"url"`
	Name   string    `json:"name"`
	Status JobStatus `json:"status"`
	FileID string    `json:"file_id"`
	Error  string    `json:"error"`
}