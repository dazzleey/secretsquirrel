package main

import (
	"math"
	"secretsquirrel/database"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type QueueJob struct {
	Bot     *SecretSquirrel
	User    *database.User
	Context *BotContext
}

type JobQueue struct {
	mu sync.Mutex
	ch chan *QueueJob
}

func NewJobQueue() *JobQueue {
	return &JobQueue{
		mu: sync.Mutex{},
		ch: make(chan *QueueJob),
	}
}

const (
	MaxWorkerPool     = 2
	MaxCallsPerSecond = 25
)

func worker(id int, ch <-chan *QueueJob) {
	minTimeBetweenRuns := time.Duration(math.Ceil(1e9 / (MaxCallsPerSecond / float64(MaxWorkerPool))))

	for j := range ch {
		// check if the message has been deleted from the queue.
		// Only send messages that are still queued.
		if q, ok := j.Bot.MessageQueue.queue[j.Context.CacheMessageID]; ok {

			// Check the rate limit
			lastRun := time.Now()
			timeUntilNextRun := -(time.Since(lastRun) - minTimeBetweenRuns)
			if timeUntilNextRun > 0 {
				time.Sleep(timeUntilNextRun)
			}

			msg := j.Bot.NewMessage(j.Context, j.User, j.Context.CacheMessageID)

			// Keep trying to send the message until successful.
			for {
				sent, err := j.Bot.Api.Send(msg.Config)
				if err != nil {
					// try again if the telegram API is rate-limiting
					if err.(*tgbotapi.Error).Code == 429 {
						time.Sleep(time.Duration(err.(*tgbotapi.Error).ResponseParameters.RetryAfter) * time.Second)
						continue
					}
					break
				}

				j.Bot.Cache.saveMapping(msg.User.ID, j.Context.CacheMessageID, sent.MessageID)
				break
			}

			// delete message from queue if this is the last job
			q.sent++
			if q.sent == q.total {
				j.Bot.MessageQueue.Delete(j.Context.CacheMessageID)
			}

		}
	}
}
