# NotifyQ
A distributed notification scheduling system built in Go using Redis.

## Architecture
The system consists of three main components working together:
- A Go HTTPS server, using `Gin`, that accepts job requests.
- A Redis instance that stores and orders job
- A Go worker that processes jobs and sends the email notifications, using goroutines.

The system has one main queue: `notifyq:jobs` where jobs are posted and pushed to. The queue itself is a Sorted Set, where the `ScheduledTime` is used as a score to sort the jobs. Thus, the job that is to be performed first is always at the first of the queue. 
Then there are three seperate queues for each type of jobs: `notifyq:jobs:pending`, `notifyq:jobs:delivered` and `notifyq:jobs:failed`, where only the ID of each job is saved.
Any job can retrieved using `notifyq:jobs:<job_id>`.

## Tech Stack
- Go
- Gin
- Redis
- Docker
- Mailtrap

## Features
- Jobs/Notifications are delivered as per scheduled.
- Async worker processing with goroutines
- Retry Logic with exponential backoff
- Dead-Letter queue for failed jobs
- Real-Time Dashboard
- Fully Containerized with Docker Compose

## Getting Started
Pretty easy, thanks to Docker.
```bash
git clone https://github.com/Sammie156/NotifyQ
cd notifyq
docker-compose up --build
```
And that's it. You can load up the real-time dashboard by going to `localhost:8080/dashboard`.

## API Endpoints
| Method | Endpoint | Utility |
| --- | --- | --- |
| POST | `/jobs` | Create a job|
| GET | `/jobs/:id` | Get status of a job |
| GET | `/jobs/pending` | List pending jobs |
| GET | `/jobs/delivered` | List delivered jobs |
| GET | `/jobs/failed` | List failed jobs |
| GET | `/dashboard` | real-time dashboard |

### Example curl request:
```bash
curl -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/json" \
  -d '{"recipient":"you@email.com","subject":"Reminder","body":"Don't forget!","scheduled_at":"2026-06-15T10:00:00Z"}'
```

## Known Limitations
- Polling interval means jobs may be up to 5 seconds late than required. The optimal solution would be to use a timer + signal channel 
- No authentication on API endpoints