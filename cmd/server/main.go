package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Sammie156/NotifyQ/internal/job"
	"github.com/Sammie156/NotifyQ/internal/queue"
)

func main() {
	j := job.NewJob("foo", "bar", "Working?", time.Now().UTC().Add(time.Minute*5))

	q, err := queue.NewQueue("localhost:6379")
	if err != nil {
		log.Fatalf("Failed to create queue: %v", err)
	}

	ctx := context.Background()

	if err := q.Enqueue(ctx, j); err != nil {
		log.Fatalf("Failed to queue job: %v", err)
	}

	fmt.Println("Job Queued!")
}
