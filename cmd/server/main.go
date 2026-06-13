package main

import (
	"log"
	"os"

	"github.com/Sammie156/NotifyQ/internal/handler"
	"github.com/Sammie156/NotifyQ/internal/queue"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	q, err := queue.NewQueue(addr)
	if err != nil {
		log.Fatalf("Queue could not be created! Error: %v", err)
	}
	h := handler.NewHandler(q)

	r.POST("/jobs", h.CreateJob)
	r.GET("/jobs/:id", h.GetJob)
	r.GET("/jobs/failed", h.GetFailedJobs)
	r.GET("/jobs/pending", h.GetPendingJobs)
	r.GET("/jobs/delivered", h.GetDeliveredJobs)
	r.GET("/dashboard", h.Dashboard)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server could not start! Error: %v", err)
	}
}
