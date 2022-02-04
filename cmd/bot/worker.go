package main

import (
	"fmt"
	"math"
	"secretsquirrel/database"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type QueueJob struct {
	Bot     *SecretSquirrel
	User    *database.User
	Context *BotContext
}

const (
	MaxWorkerPool     = 2
	MaxCallsPerSecond = 25
)

func worker(id int, ch <-chan *QueueJob) {
	minTimeBetweenRuns := time.Duration(math.Ceil(1e9 / (MaxCallsPerSecond / float64(MaxWorkerPool))))

	for j := range ch {
		lastRun := time.Now()
		timeUntilNextRun := -(time.Since(lastRun) - minTimeBetweenRuns)
		if timeUntilNextRun > 0 {
			time.Sleep(timeUntilNextRun)
		}

		msg := j.Bot.NewMessage(j.Context, j.User, j.Context.CacheMessageID)
		for {
			s, err := j.Bot.Api.Send(msg.Config)
			if err != nil {
				fmt.Println(err)
				// try again if rate-limited
				if err.(*tgbotapi.Error).Code == 429 {
					time.Sleep(time.Duration(err.(*tgbotapi.Error).ResponseParameters.RetryAfter) * time.Second)
					continue
				}
				break
			}
			j.Bot.Cache.saveMapping(msg.User.ID, j.Context.CacheMessageID, int(s.MessageID))
			break
		}
	}
}
