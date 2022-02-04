package main

import (
	"errors"
	"fmt"
	"secretsquirrel/database"
	"sync"
	"time"

	"github.com/zoumo/goset"
)

type Counter struct {
	count int
}

func (c *Counter) Next() *int {
	n := c.count
	c.count++
	return &n
}

type (
	messageID  = int
	userID     = int64
	MessageMap = map[messageID]messageID
)

type UserMap map[userID]MessageMap

func (um *UserMap) In(uid userID) bool {
	for k := range *um {
		if k == uid {
			return true
		}
	}
	return false
}

type UserCache map[userID]database.User

// CachedMessage
type CachedMessage struct {
	userID  userID
	time    time.Time
	warned  bool
	upvoted goset.Set
}

func (cm *CachedMessage) isExpired() bool {
	return time.Now().After(cm.time.Add(24 * time.Hour))
}

func (cm *CachedMessage) hasUpvoted(user int64) bool {
	return cm.upvoted.Contains(user)
}

func (cm *CachedMessage) addUpvote(user int64) {
	cm.upvoted.Add(user)
}

// MessageCache
type MessageCache struct {
	mu       sync.RWMutex
	counter  Counter
	messages map[messageID]*CachedMessage
	userMap  UserMap
}

func (ch *MessageCache) newMessage(ctx *BotContext) int {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	count := ch.counter.Next()

	ch.messages[*count] = &CachedMessage{
		userID:  ctx.Message.From.ID,
		time:    time.Now(),
		warned:  false,
		upvoted: goset.NewSet(),
	}

	return *count
}

func (ch *MessageCache) getMessage(msid messageID) (*CachedMessage, error) {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	var cm *CachedMessage

	if cm, ok := ch.messages[msid]; ok {
		return cm, nil
	}
	return cm, errors.New("cached Message not found")
}

func (ch *MessageCache) saveMapping(uid userID, msid messageID, data int) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if !ch.userMap.In(uid) {
		ch.userMap[uid] = MessageMap{}
	}
	ch.userMap[uid][msid] = data
}

func (ch *MessageCache) deleteMappings(msid messageID) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	for _, v := range ch.userMap {
		delete(v, msid)
	}
}

// LookupCacheMessage takes a messageID and returns the value for that key from MessageCache.UserMap
func (ch *MessageCache) lookupCacheMessageValue(uid userID, msid messageID) (int, error) {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	var (
		mappedValue int = -1
		ok          bool
	)

	if !ch.userMap.In(uid) {
		return -1, errors.New("lookupCacheMessageValue: user not in map")
	}

	if mappedValue, ok = ch.userMap[uid][msid]; !ok {
		return -1, errors.New("lookupCacheMessageValue: value not in userMap")
	}

	return mappedValue, nil
}

// LookupMapMessage takes a messageID and returns the key for that value from MessageCache.UserMap. The reverse of lookupCacheMessageValue.
func (ch *MessageCache) lookupCacheMessageKey(uid userID, msid messageID) (int, error) {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	var mappedKey int = -1

	if !ch.userMap.In(uid) {
		return mappedKey, errors.New("lookupCacheMessageKey: user not in map")
	}

	for k, v := range ch.userMap[uid] {
		if v == msid {
			mappedKey = k
			break
		}
	}

	if mappedKey == -1 {
		return mappedKey, errors.New("lookupCacheMessageKey: key not in userMap")
	}

	return mappedKey, nil
}

// TODO: scheduling for expire()
func (ch *MessageCache) expire() goset.Set {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	expired := goset.NewSet()

	for k, v := range ch.messages {
		if !v.isExpired() {
			continue
		}

		expired.Add(k)
		delete(ch.messages, k)
		ch.deleteMappings(k)
	}

	if l := expired.Len(); l > 0 {
		fmt.Printf("Expired %d entries from cache", l)
	}

	return expired
}

func NewMessageCache() *MessageCache {
	return &MessageCache{
		mu:       sync.RWMutex{},
		counter:  Counter{},
		messages: make(map[int]*CachedMessage),
		userMap:  make(UserMap),
	}
}
