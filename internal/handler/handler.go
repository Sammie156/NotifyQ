package handler

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Sammie156/NotifyQ/internal/job"
	"github.com/Sammie156/NotifyQ/internal/queue"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	queue *queue.Queue
}

type CreateJobRequest struct {
	Recipient   string    `json:"recipient" binding:"required"`
	Subject     string    `json:"subject" binding:"required"`
	Body        string    `json:"body" binding:"required"`
	ScheduledAt time.Time `json:"scheduled_at" binding:"required"`
}

type CreateJobResponse struct {
	ID      string        `json:"id"`
	Status  job.JobStatus `json:"status"`
	Message string        `json:"message"`
}

func NewHandler(q *queue.Queue) *Handler {
	return &Handler{queue: q}
}

func (h *Handler) CreateJob(c *gin.Context) {
	var req CreateJobRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ScheduledAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scheduled_at must be in the future"})
		return
	}

	j := job.NewJob(req.Recipient, req.Subject, req.Body, req.ScheduledAt)

	if err := h.queue.Enqueue(c.Request.Context(), j); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.queue.AddToPending(c.Request.Context(), j); err != nil {
		log.Printf("failed to add job %s to pending: %v", j.ID, err)
	}

	c.JSON(http.StatusCreated, CreateJobResponse{
		ID:      j.ID,
		Status:  j.Status,
		Message: "success",
	})
}

func (h *Handler) GetJob(c *gin.Context) {
	id := c.Param("id")

	j, err := h.queue.GetJob(context.Background(), id)
	if err != nil {
		if err == queue.ErrJobNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, CreateJobResponse{
		ID:      j.ID,
		Status:  j.Status,
		Message: "success to retrieve job",
	})
}

func (h *Handler) GetFailedJobs(c *gin.Context) {
	jobIDs, err := h.queue.GetFailedIDs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var jobs []*job.Job

	for _, value := range jobIDs {
		j, err := h.queue.GetJob(c.Request.Context(), value)
		if err != nil {
			log.Printf("Failed to get job %s: %v", value, err)
			continue
		}

		jobs = append(jobs, j)
	}

	c.JSON(http.StatusOK, gin.H{"jobs": jobs})
}

func (h *Handler) GetPendingJobs(c *gin.Context) {
	jobIDs, err := h.queue.GetPendingIDs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var jobs []*job.Job

	for _, value := range jobIDs {
		j, err := h.queue.GetJob(c.Request.Context(), value)
		if err != nil {
			log.Printf("Failed to get job %s: %v", value, err)
			continue
		}

		jobs = append(jobs, j)
	}

	c.JSON(http.StatusOK, gin.H{"jobs": jobs})
}

func (h *Handler) GetDeliveredJobs(c *gin.Context) {
	jobIDs, err := h.queue.GetDeliveredIDs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var jobs []*job.Job

	for _, value := range jobIDs {
		j, err := h.queue.GetJob(c.Request.Context(), value)
		if err != nil {
			log.Printf("failed to get job %s: %v", value, err)
			continue
		}

		jobs = append(jobs, j)
	}

	c.JSON(http.StatusOK, gin.H{"jobs": jobs})
}

func (h *Handler) Dashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", nil)
}
