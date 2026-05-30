package main

import (
	"context"
	"log"

	"github.com/Sammie156/NotifyQ/internal/job"
	"github.com/Sammie156/NotifyQ/internal/mailer"
	"github.com/Sammie156/NotifyQ/internal/queue"
)

func main() {
	q, err := queue.NewQueue("localhost:6379")
	if err != nil {
		log.Fatalf("Failed to create queue: %v", err)
	}

	mailer := mailer.NewMailer("sandbox.smtp.mailtrap.io", 2525, "75E2da00872a40", "075e768a542d2b")

	ctx := context.Background()

	for {
		j, err := q.Dequeue(ctx)
		if err != nil {
			log.Printf("Failed to dequeue: %v", err)
			continue
		}

		if err := q.UpdateStatus(ctx, j, job.StatusProcessing); err != nil {
			log.Printf("failed to update job status for job %s: %v", j.ID, err)
		}

		err = mailer.SendJob(j)
		if err != nil {
			j.LastError = err.Error()
			j.RetryCount++

			if j.RetryCount < j.MaxRetries {
				log.Printf("Retrying job %s again, retry number: %d/%d", j.ID, j.RetryCount, j.MaxRetries)
				if err := q.UpdateStatus(ctx, j, job.StatusPending); err != nil {
					log.Printf("failed to update job status for job %s: %v", j.ID, err)
				}
				if err := q.Retry(ctx, j); err != nil {
					log.Printf("failed to retry job %s: %v", j.ID, err)
				}
			} else {
				log.Printf("job %s failed permanently after %d retries", j.ID, j.RetryCount)
				if err := q.UpdateStatus(ctx, j, job.StatusFailed); err != nil {
					log.Printf("failed to update job status for job %s: %v", j.ID, err)
				}
			}
			continue
		}
		if err := q.UpdateStatus(ctx, j, job.StatusDelivered); err != nil {
			log.Printf("failed to update job status for job %s: %v", j.ID, err)
		}
		log.Printf("Job delivered: %s to %s", j.ID, j.Recipient)
	}
}
