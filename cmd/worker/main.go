package main

import (
	"context"
	"log"

	"github.com/Sammie156/NotifyQ/internal/mailer"
	"github.com/Sammie156/NotifyQ/internal/queue"
)

func main() {
	q, err := queue.NewQueue("localhost:6379")
	if err != nil {
		log.Fatalf("Failed to create queue: %v", err)
	}

	mailer := mailer.NewMailer("sandbox.smtp.mailtrap.io", 2525, "75e2da00872a40", "075e768a542d2b")

	ctx := context.Background()

	for {
		j, err := q.Dequeue(ctx)
		if err != nil {
			log.Printf("Failed to dequeue: %v", err)
			continue
		}

		err = mailer.SendJob(j)
		if err != nil {
			log.Fatalf("Failed to send job: %v", err)
			continue
		}
		log.Printf("Job delivered: %s to %s", j.ID, j.Recipient)
	}
}
