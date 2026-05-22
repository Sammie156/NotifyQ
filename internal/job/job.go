package job

import (
	"time"

	"github.com/google/uuid"
)

// JobStatus keeps track of a job and what is its current status
type JobStatus string

const (
	StatusPending   JobStatus = "pending"
	StatusDelivered JobStatus = "delivered"
	StatusFailed    JobStatus = "failed"
)

type Job struct {
	ID          string    `json:"id"`
	Recipient   string    `json:"recipient"`
	Subject     string    `json:"subject"`
	Body        string    `json:"body"`
	ScheduledAt time.Time `json:"scheduled_at"`
	CreatedAt   time.Time `json:"created_at"`
	Status      JobStatus `json:"status"`
	RetryCount  int       `json:"retry_count"`
	MaxRetries  int       `json:"max_retries"`
	LastError   string    `json:"last_error"`
}

func NewJob(recipient string, subject string, body string, scheduled time.Time) *Job {
	return &Job{
		ID:          uuid.NewString(),
		Recipient:   recipient,
		Subject:     subject,
		Body:        body,
		ScheduledAt: scheduled.UTC(),
		CreatedAt:   time.Now().UTC(),
		Status:      StatusPending,
		RetryCount:  0,
		MaxRetries:  3,
		LastError:   "",
	}
}
