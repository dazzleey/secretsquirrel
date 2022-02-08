package main

import (
	"strings"
	"sync"
)

type Scorekeeper struct {
	lock   *sync.Mutex
	scores map[userID]float32
}

func NewScoreKeeper() *Scorekeeper {
	return &Scorekeeper{
		lock:   &sync.Mutex{},
		scores: map[int64]float32{},
	}
}

func (k *Scorekeeper) increaseSpamScore(uid userID, n float32) bool {
	k.lock.Lock()
	defer k.lock.Unlock()

	var (
		score float32
		ok    bool
	)

	score, ok = k.scores[uid]
	if !ok {
		score = 0
	}

	if score > float32(cfg.Spam.SpamLimit) {
		return false
	} else if score+n > float32(cfg.Spam.SpamLimit) {
		k.scores[uid] = float32(cfg.Spam.SpamLimitHit)
		return false
	}

	k.scores[uid] = score + n
	return true
}

func (k *Scorekeeper) expireTask() {
	k.lock.Lock()
	defer k.lock.Unlock()
	for uid := range k.scores {
		newScore := k.scores[uid] - 1
		if newScore <= 0 {
			delete(k.scores, uid)
		} else {
			k.scores[uid] = newScore
		}
	}
}

func calculateSpamScore(ctx *BotContext) float32 {
	var score = cfg.Spam.ScoreBaseMessage

	if ctx.IsForward() {
		score = cfg.Spam.ScoreBaseForward
	}

	switch ctx.ContentType {
	case StickerContentType:
		return cfg.Spam.ScoreSticker
	case MessageContentType:
		score = float32(len(ctx.Message.Text))*cfg.Spam.ScoreTextCharacter +
			float32(strings.Count(ctx.Message.Text, "\n"))*cfg.Spam.ScoreTextLineBreak
	default:
	}

	return score
}
