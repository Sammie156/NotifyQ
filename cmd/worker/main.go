package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Sammie156/NotifyQ/internal/queue"
)

func main() {
	q, err := queue.NewQueue("localhost:6379")
	if err != nil {
		log.Fatalf("Failed to create queue: %v", err)
	}

	ctx := context.Background()

	for {
		j, err := q.Dequeue(ctx)
		if err != nil {
			log.Printf("Failed to dequeue: %v", err)
			continue
		}

		fmt.Println(j.ID)
	}
}
