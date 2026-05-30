package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Sammie156/NotifyQ/internal/job"
	"github.com/redis/go-redis/v9"
)

const defaultQueueKey = "notifyq:jobs"

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
		return err
	}

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
