package queue

import (
	"context"
	"encoding/json"

	"github.com/Sammie156/NotifyQ/internal/job"
	"github.com/redis/go-redis/v9"
)

const defaultQueueKey = "notifyq:jobs"

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

	_, err = q.client.LPush(ctx, q.keyName, jsonData).Result()
	if err != nil {
		return err
	}

	return nil
}

func (q *Queue) Dequeue(ctx context.Context) (*job.Job, error) {
	jsonData, err := q.client.BRPop(ctx, 0, q.keyName).Result()
	if err != nil {
		return nil, err
	}

	var j job.Job
	if err := json.Unmarshal([]byte(jsonData[1]), &j); err != nil {
		return nil, err
	}

	return &j, nil
}
