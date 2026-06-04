package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Sammie156/NotifyQ/internal/job"
	"github.com/redis/go-redis/v9"
)

// Keys for different queues/lists in Redis client
const defaultQueueKey = "notifyq:jobs"
const pendingQueueKey = "notifyq:jobs:pending"
const deliveredQueueKey = "notifyq:jobs:delivered"
const failedQueueKey = "notifyq:jobs:failed"

var ErrJobNotFound = fmt.Errorf("job not found")

type Queue struct {
	client  *redis.Client
	keyName string
}

func NewQueue(address string) (*Queue, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: address,
	})

	ctx := context.Background()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Queue{
		client:  rdb,
		keyName: defaultQueueKey,
	}, nil
}

func (q *Queue) Enqueue(ctx context.Context, j *job.Job) error {
	jsonData, err := json.Marshal(j)
	if err != nil {
		return fmt.Errorf("failed to convert job %s to json: %v", j.ID, err)
	}

	// Using ZAdd because it helps in sorting anything we push based on a score
	// Using the UTC time in miliseconds as the score.
	// Hence jobs are sorted and the first scheduled job is always at the front
	_, err = q.client.ZAdd(ctx, q.keyName, redis.Z{
		Member: jsonData,
		Score:  float64(j.ScheduledAt.Unix()),
	}).Result()
	if err != nil {
		return err
	}

	return nil
}

func (q *Queue) Dequeue(ctx context.Context) (*job.Job, error) {
	jsonData, err := q.client.BZPopMin(ctx, 0, q.keyName).Result()
	if err != nil {
		return nil, err
	}

	var j job.Job
	if err := json.Unmarshal([]byte(jsonData.Member.(string)), &j); err != nil {
		return nil, err
	}

	return &j, nil
}

// Save a particular job with its state after it has been processed
// Need not be successful
func (q *Queue) SaveJob(ctx context.Context, j *job.Job) error {
	jsonData, err := json.Marshal(j)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("notifyq:job:%s", j.ID)
	err = q.client.Set(ctx, key, jsonData, 0).Err()

	if err != nil {
		return err
	}
	return nil
}

func (q *Queue) GetJob(ctx context.Context, id string) (*job.Job, error) {
	key := fmt.Sprintf("notifyq:job:%s", id)

	jsonData, err := q.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, ErrJobNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("getjob: failed to get job id %s: %w", id, err)
	}

	var j job.Job
	if err := json.Unmarshal([]byte(jsonData), &j); err != nil {
		return nil, err
	}

	return &j, nil
}

func (q *Queue) UpdateStatus(ctx context.Context, j *job.Job, status job.JobStatus) error {
	j.Status = status
	err := q.SaveJob(ctx, j)

	if err != nil {
		return err
	}
	return nil
}

func (q *Queue) AddToDeadLetter(ctx context.Context, j *job.Job) error {
	_, err := q.client.LPush(ctx, failedQueueKey, j.ID).Result()
	if err != nil {
		return fmt.Errorf("failed to send job %s to Dead Letter Queue: %v", j.ID, err)
	}

	return nil
}

func (q *Queue) AddToPending(ctx context.Context, j *job.Job) error {
	_, err := q.client.LPush(ctx, pendingQueueKey, j.ID).Result()
	if err != nil {
		return fmt.Errorf("failed to move job %s to pending: %v", j.ID, err)
	}

	return nil
}

func (q *Queue) RemoveFromPending(ctx context.Context, j *job.Job) error {
	key := j.ID
	_, err := q.client.LRem(ctx, pendingQueueKey, 1, key).Result()
	if err != nil {
		return fmt.Errorf("failed to remove job %s from queue: %v", j.ID, err)
	}

	return nil
}

func (q *Queue) AddToDelivered(ctx context.Context, j *job.Job) error {
	_, err := q.client.LPush(ctx, deliveredQueueKey, j.ID).Result()
	if err != nil {
		return fmt.Errorf("failed to move job %s to delivered: %v", j.ID, err)
	}

	return nil
}

func (q *Queue) GetFailedIDs(ctx context.Context) ([]string, error) {
	jobList, err := q.client.LRange(ctx, failedQueueKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("unable to get failed jobs list: %v", err)
	}

	return jobList, nil
}

func (q *Queue) GetPendingIDs(ctx context.Context) ([]string, error) {
	jobList, err := q.client.LRange(ctx, pendingQueueKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve pending jobs: %v", err)
	}

	return jobList, nil
}

func (q *Queue) GetDeliveredIDs(ctx context.Context) ([]string, error) {
	jobList, err := q.client.LRange(ctx, deliveredQueueKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve delivered jobs: %v", err)
	}

	return jobList, nil
}
