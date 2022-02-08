package main

import (
	"database/sql"
	"fmt"
	"log"
	"secretsquirrel/crypt"
	"secretsquirrel/database"
	"secretsquirrel/messages"
	"secretsquirrel/util"
	"strings"
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

// handleMessage does further checks on the message and user before queueing the job for relaying by workers.
func (bot *SecretSquirrel) handleMessage(ctx *BotContext) error {
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

	// check types limits
	if ctx.ContentType == DocumentContentType && !cfg.Limits.AllowDocuments {
		return nil
	}
	if ctx.ContentType == ContactContentType && !cfg.Limits.AllowContacts {
		return nil
	}

	bot.sendMessage(ctx)
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
			bot.sendSystemMessage(ctx.User.ID, fmt.Sprintf(messages.CooldownError, ctx.User.CooldownUntil.Time.Format(time.RFC822)))
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

func (bot *SecretSquirrel) giveWarning(ctx *BotContext, cm *CachedMessage) {
	var (
		user             = (*bot.Users)[cm.userID]
		cooldownDuration time.Duration
	)

	if user.Warnings < len(cfg.Cooldown.CooldownTimeBegin) {
		cooldownDuration = time.Duration(cfg.Cooldown.CooldownTimeBegin[user.Warnings]) * time.Minute
	} else {
		x := user.Warnings - len(cfg.Cooldown.CooldownTimeBegin)
		cooldownDuration = time.Duration(cfg.Cooldown.CooldownTimeLinearM*x+cfg.Cooldown.CooldownTimeLinearB) * time.Minute
	}

	bot.UpdatesUser(&user, database.User{
		CooldownUntil: sql.NullTime{Time: time.Now().Add(cooldownDuration), Valid: true},
		Karma:         user.Karma - cfg.Karma.KarmaWarnPenalty,
	})

	cm.warned = true
	replyID, err := bot.Cache.lookupCacheMessageValue(cm.userID, ctx.ReplyID)
	if err != nil {
		fmt.Println(err)
		return
	}

	bot.sendSystemMessageReply(cm.userID, fmt.Sprintf(messages.GivenCooldownMessage, util.TimeStr(cooldownDuration)), replyID)
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
	msg := tgbotapi.MessageConfig{
		BaseChat:              tgbotapi.BaseChat{ChatID: userID, ReplyToMessageID: replyID},
		Text:                  message,
		ParseMode:             "HTML",
		DisableWebPagePreview: false,
	}

	return bot.Api.Send(&msg)
}

func (bot *SecretSquirrel) sendMessage(ctx *BotContext) {
	bot.JobQueue.mu.Lock()
	defer bot.JobQueue.mu.Unlock()

	ctx.CacheMessageID = bot.Cache.newMessage(ctx)

	bot.MessageQueue.Add(ctx.CacheMessageID, &MessageQueueItem{0, len(bot.UserQueue.items) - 1})

	// queue job for all users.
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

}

func (bot *SecretSquirrel) deleteMessage(ctx *BotContext, cm *CachedMessage) {
	// cancel unsent queued messages
	bot.MessageQueue.Delete(ctx.ReplyID)

	for _, uindex := range bot.UserQueue.Get() {
		if uindex != cm.userID {
			var (
				user         = (*bot.Users)[uindex]
				user_replyID = -1
				err          error
			)

			if ctx.ReplyID != -1 {
				user_replyID, err = bot.Cache.lookupCacheMessageValue(user.ID, ctx.ReplyID)
				if err != nil {
					fmt.Println(err)
					continue
				}
			}

			// delete each message
			// api might get mad if this deletes more than 30 messages a second. I'm not sure
			if user_replyID != -1 {
				bot.Api.Send(tgbotapi.NewDeleteMessage(user.ID, user_replyID))
			}
		}
	}
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
	bot.JobQueue = NewJobQueue()
	for i := 0; i < MaxWorkerPool; i++ {
		go worker(i, bot.JobQueue.ch)
	}

	bot.UserQueue = NewPriorityQueue()
	bot.MessageQueue = NewMessageQueue()

	bot.Users = &UserCache{}
	users, err := database.FindUsers(bot.Db, database.AreJoined)
	if err != nil {
		log.Panic("initApp: db query failed.")
	}

	for _, u := range users {
		(*bot.Users)[u.ID] = u
		bot.UserQueue.Add(u.ID)
	}

	bot.Spam = NewScoreKeeper()

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
