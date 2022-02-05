package main

import (
	"database/sql"
	"fmt"
	"log"
	"secretsquirrel/config"
	"secretsquirrel/crypt"
	"secretsquirrel/database"
	"secretsquirrel/messages"
	"strings"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

const (
	BotVersion = "1.0"
)

type SecretSquirrel struct {
	Api          *tgbotapi.BotAPI
	Db           *gorm.DB
	Users        *UserCache
	Cache        *MessageCache
	UserQueue    *PriorityQueue
	JobQueue     *JobQueue
	MessageQueue *MessageQueue
	Spam         *Scorekeeper
	Scheduler    *gocron.Scheduler
}

type JobQueue struct {
	ch chan *QueueJob
	mu sync.Mutex
}

type MessageQueue struct {
	mu    sync.Mutex
	queue map[int]*QueuedMessage
}

func (mq *MessageQueue) Add(msid int, qm *QueuedMessage) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	mq.queue[msid] = qm
}

func (mq *MessageQueue) Delete(msid int) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	delete(mq.queue, msid)
}

type QueuedMessage struct {
	sent  int
	total int
}

// handleMessage does further checks on the message and user before queueing the job for relaying by workers.
func (bot *SecretSquirrel) handleMessage(ctx *BotContext) error {
	bot.JobQueue.mu.Lock()
	defer bot.JobQueue.mu.Unlock()

	// ignore messages from untracked users.
	if ctx.User == nil {
		bot.sendSystemMessage(ctx.User.ID, messages.UserNotInChatMessage)
		return nil
	}

	// check media limit period
	if (ctx.HasFile() || ctx.IsForward()) && cfg.Limits.MediaLimitPeriod > 0 {
		if int(time.Since(ctx.User.Joined).Hours()) < cfg.Limits.MediaLimitPeriod {
			bot.sendSystemMessage(ctx.User.ID, messages.MediaLimitError)
			return nil
		}
	}

	// check if user is spamming.
	if ok := bot.Spam.increaseSpamScore(ctx.User.ID, calculateSpamScore(ctx)); !ok {
		bot.sendSystemMessage(ctx.User.ID, messages.SpamError)
		return nil
	}

	ctx.CacheMessageID = bot.Cache.newMessage(ctx)

	// check types limits
	if ctx.ContentType == DocumentContentType && !cfg.Limits.AllowDocuments {
		return nil
	}
	if ctx.ContentType == ContactContentType && !cfg.Limits.AllowContacts {
		return nil
	}

	bot.MessageQueue.Add(ctx.CacheMessageID, &QueuedMessage{0, len(bot.UserQueue.items) - 1})

	// echo message to all users.
	for _, uindex := range bot.UserQueue.Get() {
		user := (*bot.Users)[uindex]
		// only resend message back to the sender if debug is enabled.
		// this seems to break replies while debug is enabled. idk why
		if user.ID == ctx.Message.From.ID && !user.DebugEnabled {
			bot.Cache.saveMapping(user.ID, ctx.CacheMessageID, ctx.Message.MessageID)
			continue
		}

		bot.JobQueue.ch <- &QueueJob{Bot: bot, User: &user, Context: ctx}
	}

	return nil
}

// handleUpdate is the first step of message processing. It handles the initial tgbotapi.Update from the API.
// Things like checking if the user left, if the user is banned, or if a command was run are done here before messages themselves are preocessed.
func (bot *SecretSquirrel) handleUpdate(u tgbotapi.Update) {

	// wrap update in new bot context
	ctx := bot.CreateContext(u)

	// if user left/blocked the bot.
	if ctx.UserLeftOrKicked() {
		bot.UpdateUser(ctx.User, "left", sql.NullTime{Time: time.Now(), Valid: true})
		return
	}

	// I'm not sure if any other update events don't contain a Message
	// but if they do ignore them.
	if ctx.Message == nil {
		return
	}

	// check if on cooldown or blacklisted before anything else.
	if ctx.User != nil {
		if ctx.User.IsInCooldown() {
			bot.sendSystemMessage(ctx.User.ID, fmt.Sprintf(messages.CooldownError, ctx.User.CooldownUntil.Time.String()))
			return
		}

		if ctx.User.IsBlacklisted() {
			msgText := fmt.Sprintf(messages.BlacklistedError, ctx.User.BlacklistReason)
			if cfg.Bot.BlacklistContact != "" {
				msgText += fmt.Sprintf("\n\nContact: %s", cfg.Bot.BlacklistContact)
			}
			bot.sendSystemMessage(ctx.User.ID, msgText)
			return
		}
	}

	// If the message is a command, check that it's in the map of valid commands.
	// run the command if it is.
	if cmd, ok := BotCommands[ctx.Message.Command()]; ok {
		// TODO: should probably add error returns to commands.
		cmd(bot, ctx)
		return
	}

	if ctx.IsReply() && strings.TrimSpace(ctx.Message.Text) == "+1" {
		bot.giveKarma(ctx)
		return
	}

	if err := bot.handleMessage(ctx); err != nil {
		fmt.Println(err)
	}
}

func (bot *SecretSquirrel) giveKarma(ctx *BotContext) {
	if !ctx.IsReply() {
		bot.sendSystemMessageReply(ctx.User.ID, messages.NoReplyError, ctx.Message.MessageID)
		return
	}

	cm, err := bot.Cache.getMessage(ctx.ReplyID)
	if err != nil {
		bot.sendSystemMessageReply(ctx.User.ID, messages.NotInCacheError, ctx.Message.MessageID)
		return
	}

	if cm.hasUpvoted(ctx.User.ID) {
		bot.sendSystemMessageReply(ctx.User.ID, messages.AlreadyUpvotedError, ctx.Message.MessageID)
		return
	}

	if cm.userID == ctx.User.ID {
		bot.sendSystemMessageReply(ctx.User.ID, messages.UpvoteOwnMessageError, ctx.Message.MessageID)
		return
	}

	user := (*bot.Users)[cm.userID]

	bot.UpdatesUser(&user, database.User{Karma: user.Karma + cfg.Karma.KarmaPlusOne})
	cm.addUpvote(ctx.User.ID)

	if !user.HideKarma {
		reply, _ := bot.Cache.lookupCacheMessageValue(cm.userID, ctx.ReplyID)
		bot.sendSystemMessageReply(user.ID, messages.UpvoteOwnMessageError, reply)
	}

	bot.sendSystemMessage(ctx.User.ID, messages.KarmaThankMessage)
}

func (bot *SecretSquirrel) AddWarning(cfg config.Config, user *database.User) time.Time {
	var cooldownTime int

	if user.Warnings < len(cfg.Cooldown.CooldownTimeBegin) {
		cooldownTime = cfg.Cooldown.CooldownTimeBegin[user.Warnings]
	} else {
		x := user.Warnings - len(cfg.Cooldown.CooldownTimeBegin)
		cooldownTime = cfg.Cooldown.CooldownTimeLinearM*x + cfg.Cooldown.CooldownTimeLinearB
	}

	bot.UpdatesUser(user, database.User{
		CooldownUntil: sql.NullTime{Time: time.Now().Add(time.Minute * time.Duration(cooldownTime)), Valid: true},
		Karma:         user.Karma - cfg.Karma.KarmaWarnPenalty,
	})

	return user.CooldownUntil.Time
}

func (bot *SecretSquirrel) sendSystemMessage(userID int64, message string) (tgbotapi.Message, error) {
	msg := tgbotapi.MessageConfig{
		BaseChat:              tgbotapi.BaseChat{ChatID: userID},
		Text:                  message,
		ParseMode:             "HTML",
		DisableWebPagePreview: false,
	}

	return bot.Api.Send(&msg)
}

func (bot *SecretSquirrel) sendSystemMessageReply(userID int64, message string, replyID int) (tgbotapi.Message, error) {
	baseChat := tgbotapi.BaseChat{
		ChatID:           userID,
		ReplyToMessageID: replyID,
	}

	msg := tgbotapi.MessageConfig{
		BaseChat:              baseChat,
		Text:                  message,
		ParseMode:             "HTML",
		DisableWebPagePreview: false,
	}

	return bot.Api.Send(&msg)
}

// UpdateUser wraps *gorm.DB.Model().Update() and adds the updated database.User to the bot cache.
func (bot *SecretSquirrel) UpdateUser(user *database.User, column string, value interface{}) {
	bot.Db.Model(user).Update(column, value)
	(*bot.Users)[user.ID] = *user
}

// UpdateUser wraps *gorm.DB.Model().Updates() and adds the updated database.User to the bot cache.
func (bot *SecretSquirrel) UpdatesUser(user *database.User, values interface{}) {
	bot.Db.Model(user).Updates(values)
	(*bot.Users)[user.ID] = *user
}

func initBot() *SecretSquirrel {
	var (
		bot *SecretSquirrel = &SecretSquirrel{}
		err error
	)

	bot.Db = database.InitDB(cfg.Bot.DatabasePath)
	bot.Api, err = tgbotapi.NewBotAPI(cfg.Bot.Token)
	if err != nil {
		log.Panic(err)
	}

	bot.Cache = NewMessageCache()

	// create message queue and start workers.
	bot.JobQueue = &JobQueue{ch: make(chan *QueueJob), mu: sync.Mutex{}}
	for i := 0; i < MaxWorkerPool; i++ {
		go worker(i, bot.JobQueue.ch)
	}

	bot.UserQueue = NewPriorityQueue()
	bot.MessageQueue = &MessageQueue{
		mu:    sync.Mutex{},
		queue: map[int]*QueuedMessage{},
	}

	bot.Users = &UserCache{}
	users, err := database.FindUsers(bot.Db, database.AreJoined)
	if err != nil {
		log.Panic("initApp: db query failed.")
	}

	for _, u := range users {
		(*bot.Users)[u.ID] = u
		bot.UserQueue.Add(u.ID)
	}

	bot.Spam = &Scorekeeper{
		lock:   &sync.Mutex{},
		scores: map[int64]float32{},
	}

	bot.Scheduler = gocron.NewScheduler(time.UTC)
	bot.Scheduler.Every(5).Seconds().Do(bot.Spam.expireTask)
	bot.Scheduler.Every(6).Hours().Do(bot.Cache.expire)
	bot.Scheduler.StartAsync()

	return bot
}

func genTripcode(tripcode string) []string {
	var (
		trimpass string
		salt     string
		final    string
	)

	split := strings.Split(tripcode, "#")
	trname := split[0]
	trpass := split[1]

	if len(trpass) >= 8 {
		trimpass = trpass[:8]
	} else {
		trimpass = trpass
	}

	salt = (trimpass + "H.")[1:3]
	for _, v := range salt {
		final += crypt.Salt(v)
	}

	final = crypt.Crypt(trimpass, final)

	return []string{trname, "!" + final[len(final)-10:]}
}
