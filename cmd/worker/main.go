package main

import (
	"context"
	"log"
	"math"
	"os"
	"time"

	"github.com/Sammie156/NotifyQ/internal/job"
	"github.com/Sammie156/NotifyQ/internal/mailer"
	"github.com/Sammie156/NotifyQ/internal/queue"
)

func main() {
	log.Printf("worker starting, REDIS_ADDR=%s", os.Getenv("REDIS_ADDR"))
	addr := os.Getenv("REDIS_ADDR")
	q, err := queue.NewQueue(addr)
	if err != nil {
		log.Fatalf("Failed to create queue: %v", err)
	}
	host := os.Getenv("MAILTRAP_HOST")
	port := 2525
	user := os.Getenv("MAILTRAP_USER")
	pass := os.Getenv("MAILTRAP_PASS")

	mailer := mailer.NewMailer(host, port, user, pass)

	ctx := context.Background()

	log.Println("Worker started")

	for {
		j, err := q.Dequeue(ctx)
		if err != nil {
			log.Printf("Failed to dequeue: %v", err)
			continue
		}

		if time.Now().UTC().Before(j.ScheduledAt) {
			q.Enqueue(ctx, j)
			time.Sleep(5 * time.Second)
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
				backoff := time.Duration(math.Pow(2, float64(j.RetryCount))) * time.Second
				j.ScheduledAt = time.Now().UTC().Add(backoff)
				if err := q.Enqueue(ctx, j); err != nil {
					log.Printf("failed to retry job %s: %v", j.ID, err)
				}
			} else {
				log.Printf("job %s failed permanently after %d retries", j.ID, j.RetryCount)
				if err := q.RemoveFromPending(ctx, j); err != nil {
					log.Printf("failed to remove job %s from pending: %v", j.ID, err)
				}
				if err := q.UpdateStatus(ctx, j, job.StatusFailed); err != nil {
					log.Printf("failed to update job status for job %s: %v", j.ID, err)
				}
				if err := q.AddToDeadLetter(ctx, j); err != nil {
					log.Printf("failed to push job %s to Dead Letter Queue: %v", j.ID, err)
				}
			}
			continue
		}
		if err := q.RemoveFromPending(ctx, j); err != nil {
			log.Printf("failed to remove job %s from pending: %v", j.ID, err)
		}
		if err := q.AddToDelivered(ctx, j); err != nil {
			log.Printf("failed to add job %s to delivered: %v", j.ID, err)
		}
		if err := q.UpdateStatus(ctx, j, job.StatusDelivered); err != nil {
			log.Printf("failed to update job status for job %s: %v", j.ID, err)
		}
		log.Printf("Job delivered: %s to %s", j.ID, j.Recipient)
	}
}
