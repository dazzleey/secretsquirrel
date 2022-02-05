package main

import (
	"database/sql"
	"fmt"
	"secretsquirrel/messages"
	"strings"
	"time"

	"secretsquirrel/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

var BotCommands = map[string]func(*SecretSquirrel, *BotContext){
	"start":          cmdStart,
	"users":          cmdUsers,
	"info":           cmdInfo,
	"sign":           cmdSignMessage,
	"s":              cmdSignMessage,
	"tsign":          cmdTSign,
	"t":              cmdTSign,
	"motd":           cmdMotd,
	"setmotd":        cmdSetMotd,
	"modhelp":        cmdModHelp,
	"adminhelp":      cmdAdminHelp,
	"toggledebug":    cmdToggleDebug,
	"toggleKarma":    cmdToggleKarma,
	"toggletripcode": cmdToggleTripcode,
	"tripcode":       cmdTripcode,
	"blacklist":      cmdBlacklist,
	"warn":           cmdWarn,
	"delete":         cmdDelete,
	"remove":         cmdRemove,
	"mod":            cmdPromoteMod,
	"admin":          cmdPromoteAdmin,
	"version":        cmdVersion,
}

func cmdStart(bot *SecretSquirrel, ctx *BotContext) {
	// user already in the db
	if ctx.User != nil {
		// already in the chat.
		if !ctx.User.Left.Valid {
			bot.sendSystemMessage(ctx.Message.From.ID, messages.UserInChatMessage)
			return
		}
		bot.UpdateUser(ctx.User, "left", sql.NullTime{})
		bot.UserQueue.Add(ctx.User.ID)
		cmdMotd(bot, ctx)
		return
	}

	var count int64
	bot.Db.Model(database.User{}).Where("rank = ?", database.RankAdmin).Count(&count)

	user := database.NewUser(bot.Db, ctx.Message.From)
	if count == 0 {
		user.Rank = database.RankAdmin
	}
	bot.Db.Save(&user)
	(*bot.Users)[user.ID] = *user
	bot.UserQueue.Add(user.ID)
	cmdMotd(bot, ctx)
}

func cmdUsers(bot *SecretSquirrel, ctx *BotContext) {
	bot.sendSystemMessage(ctx.User.ID, fmt.Sprintf("<b>%d</b> users", len((*bot.Users))))
}

func cmdInfo(bot *SecretSquirrel, ctx *BotContext) {
	// mod Info
	if ctx.IsReply() && ctx.User.IsPrivileged() {
		cmdModInfo(bot, ctx)
		return
	}

	info, err := messages.UserInfo(ctx.User)
	if err != nil {
		bot.sendSystemMessageReply(ctx.Message.From.ID, err.Error(), ctx.Message.MessageID)
		return
	}

	bot.sendSystemMessageReply(ctx.Message.From.ID, info, ctx.Message.MessageID)
}

func cmdModInfo(bot *SecretSquirrel, ctx *BotContext) {
	if !ctx.User.IsPrivileged() {
		return
	}

	cm, err := bot.Cache.getMessage(ctx.ReplyID)
	if err != nil {
		fmt.Println(err)
		return
	}

	if user, ok := (*bot.Users)[cm.userID]; !ok {
		bot.sendSystemMessageReply(ctx.Message.From.ID, messages.NoUserError, ctx.Message.MessageID)
		return

	} else {
		info, err := messages.ModUserInfo(&user)
		if err != nil {
			bot.sendSystemMessageReply(ctx.Message.From.ID, err.Error(), ctx.Message.MessageID)
			return
		}
		bot.sendSystemMessageReply(ctx.Message.From.ID, info, ctx.Message.MessageID)
	}
}

func cmdSignMessage(bot *SecretSquirrel, ctx *BotContext) {
	if !cfg.Limits.EnableSigning {
		bot.sendSystemMessage(ctx.Message.From.ID, messages.SigningDisabledError)
		return
	}

	ctx.Signed = true
	err := bot.handleMessage(ctx)
	if err != nil {
		fmt.Println(err)
	}
}

func cmdTSign(bot *SecretSquirrel, ctx *BotContext) {
	if !cfg.Limits.EnableSigning {
		bot.sendSystemMessage(ctx.Message.From.ID, messages.SigningDisabledError)
		return
	}

	ctx.Tripcode = true
	err := bot.handleMessage(ctx)
	if err != nil {
		fmt.Println(err)
	}
}

func cmdSetMotd(bot *SecretSquirrel, ctx *BotContext) {
	if !ctx.User.IsAdmin() {
		return
	}

	motd := ctx.Message.CommandArguments()
	err := database.SetMotd(bot.Db, motd)
	if err != nil {
		fmt.Println(err)
	}
}

func cmdMotd(bot *SecretSquirrel, ctx *BotContext) {
	motd := database.GetMotd(bot.Db)

	if motd != "" {
		bot.sendSystemMessage(ctx.Message.From.ID, motd)
	} else {
		bot.sendSystemMessage(ctx.Message.From.ID, "No MOTD is set.")
	}
}

func cmdModHelp(bot *SecretSquirrel, ctx *BotContext) {
	bot.sendSystemMessageReply(ctx.Message.From.ID, messages.ModeratorHelp, ctx.Message.MessageID)
}

func cmdAdminHelp(bot *SecretSquirrel, ctx *BotContext) {
	bot.sendSystemMessageReply(ctx.Message.From.ID, messages.AdminHelp, ctx.Message.MessageID)
}

func cmdToggleDebug(bot *SecretSquirrel, ctx *BotContext) {
	var debugState string

	bot.UpdateUser(ctx.User, "DebugEnabled", !ctx.User.DebugEnabled)

	if ctx.User.DebugEnabled {
		debugState = "enabled"
	} else {
		debugState = "disabled"
	}

	bot.sendSystemMessage(ctx.Message.From.ID, fmt.Sprintf("Debug %s.", debugState))
}

func cmdToggleKarma(bot *SecretSquirrel, ctx *BotContext) {
	var karmaState string

	bot.UpdateUser(ctx.User, "HideKarma", !ctx.User.HideKarma)

	if ctx.User.HideKarma {
		karmaState = "enabled"
	} else {
		karmaState = "disabled"
	}

	bot.sendSystemMessage(ctx.Message.From.ID, fmt.Sprintf("Karma notifications %s.", karmaState))
}

func cmdToggleTripcode(bot *SecretSquirrel, ctx *BotContext) {
	var tripcodeState string

	bot.UpdateUser(ctx.User, "ToggleTripcode", !ctx.User.ToggleTripcode)

	if ctx.User.ToggleTripcode {
		tripcodeState = "Enabled"
	} else {
		tripcodeState = "Disabled"
	}

	bot.sendSystemMessage(ctx.Message.From.ID, fmt.Sprintf("Tripcode Toggle: %s", tripcodeState))
}

func cmdTripcode(bot *SecretSquirrel, ctx *BotContext) {
	var (
		s           = ctx.Message.CommandArguments()
		hasTripcode bool
		trip        []string = make([]string, 2)
	)

	// send tripcode info message
	if len(s) == 0 {
		if ctx.User.Tripcode != "" {
			hasTripcode = true
			trip = genTripcode(ctx.User.Tripcode)
		}

		msg, err := messages.TripcodeMessage(hasTripcode, trip)
		if err != nil {
			bot.sendSystemMessage(ctx.Message.From.ID, err.Error())
		}

		bot.sendSystemMessage(ctx.User.ID, msg)
		return
	}

	pos := strings.Index(s, "#")
	if pos == -1 {
		return
	}

	if !(0 < pos && pos < len(s)-1) {
		return
	}
	if strings.Contains(s, "\n") || len(s) > 30 {
		return
	}

	bot.UpdateUser(ctx.User, "tripcode", s)

	trip = genTripcode(ctx.User.Tripcode)
	msg, err := messages.NewTripcodeMessage(trip)
	if err != nil {
		bot.sendSystemMessage(ctx.Message.From.ID, err.Error())
		return
	}

	bot.sendSystemMessageReply(ctx.User.ID, msg, ctx.Message.MessageID)
}

func cmdPromoteMod(bot *SecretSquirrel, ctx *BotContext) {
	username := strings.Replace(ctx.Message.CommandArguments(), "@", "", -1)
	user, err := database.FindUser(bot.Db, database.ByUsername(username))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			bot.sendSystemMessage(ctx.User.ID, messages.NoUserByNameError)
		}
	}

	bot.UpdateUser(ctx.User, "rank", database.RankMod)
	bot.sendSystemMessage(user.ID, messages.PromotedModMessage)
	bot.sendSystemMessageReply(ctx.User.ID, fmt.Sprintf("User Promoted to Mod: @%s", username), ctx.Message.MessageID)
}

func cmdPromoteAdmin(bot *SecretSquirrel, ctx *BotContext) {
	username := strings.Replace(ctx.Message.CommandArguments(), "@", "", -1)
	user, err := database.FindUser(bot.Db, database.ByUsername(username))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			bot.sendSystemMessage(ctx.User.ID, messages.NoUserByNameError)
		}
	}

	bot.UpdateUser(ctx.User, "rank", database.RankAdmin)
	bot.sendSystemMessage(user.ID, messages.PromotedAdminMessage)
	bot.sendSystemMessageReply(ctx.User.ID, fmt.Sprintf("User Promoted to Admin: @%s", username), ctx.Message.MessageID)
}

func cmdRemove(bot *SecretSquirrel, ctx *BotContext) {
	if !cfg.Limits.AllowRemoveCommand {
		bot.sendSystemMessageReply(ctx.User.ID, messages.CommandDisabledError, ctx.Message.MessageID)
		return
	}

	// must be mod.
	if !ctx.User.IsPrivileged() {
		return
	}

	// reply required
	if !ctx.IsReply() {
		bot.sendSystemMessageReply(ctx.User.ID, messages.NoReplyError, ctx.Message.MessageID)
		return
	}

	cm, err := bot.Cache.getMessage(ctx.ReplyID)
	if err != nil {
		fmt.Println(err)
		return
	}

	bot.sendSystemMessageReply(cm.userID, messages.MessageDeletedMessage, ctx.ReplyID)

	go func() {
		for _, uindex := range bot.UserQueue.Get() {
			var user_replyID int = -1
			user := (*bot.Users)[uindex]

			if ctx.ReplyID != -1 {
				user_replyID, err = bot.Cache.lookupCacheMessageValue(user.ID, ctx.ReplyID)
				if err != nil {
					fmt.Println(err)
					continue
				}
			}

			// delete each message
			// api might get mad if this deletes more than 30 messages.
			if user_replyID != -1 {
				bot.Api.Send(tgbotapi.NewDeleteMessage(user.ID, user_replyID))
			}
		}
		bot.Cache.deleteMappings(ctx.ReplyID)
	}()
}

func cmdDelete(bot *SecretSquirrel, ctx *BotContext) {
	// must be mod.
	if !ctx.User.IsPrivileged() {
		return
	}

	// reply required
	if !ctx.IsReply() {
		bot.sendSystemMessageReply(ctx.User.ID, messages.NoReplyError, ctx.Message.MessageID)
		return
	}

	cm, err := bot.Cache.getMessage(ctx.ReplyID)
	if err != nil {
		fmt.Println(err)
		return
	}

	// message already has a warning.
	if cm.warned {
		bot.sendSystemMessageReply(ctx.User.ID, messages.AlreadyWarnedError, ctx.Message.MessageID)
		return
	}

	user, err := database.FindUser(bot.Db, database.ByID(cm.userID))
	if err != nil {
		fmt.Println(err)
		return
	}

	t := bot.AddWarning(cfg, user)
	cm.warned = true

	replyID, err := bot.Cache.lookupCacheMessageValue(cm.userID, ctx.ReplyID)
	if err != nil {
		fmt.Println(err)
		return
	}

	bot.sendSystemMessageReply(cm.userID, fmt.Sprintf(messages.GivenCooldownMessage, t), replyID)

	go func() {
		// echo message to all users.
		for _, uindex := range bot.UserQueue.Get() {
			var user_replyID int = -1
			user := (*bot.Users)[uindex]

			if ctx.ReplyID != -1 {
				user_replyID, err = bot.Cache.lookupCacheMessageValue(user.ID, ctx.ReplyID)
				if err != nil {
					fmt.Println(err)
					continue
				}
			}

			// delete each message
			if user_replyID != -1 {
				bot.Api.Send(tgbotapi.NewDeleteMessage(user.ID, user_replyID))
			}
		}
		bot.Cache.deleteMappings(ctx.ReplyID)
	}()
}

func cmdWarn(bot *SecretSquirrel, ctx *BotContext) {
	// must be mod.
	if !ctx.User.IsPrivileged() {
		return
	}

	// reply required
	if !ctx.IsReply() {
		bot.sendSystemMessageReply(ctx.User.ID, messages.NoReplyError, ctx.Message.MessageID)
		return
	}

	cm, err := bot.Cache.getMessage(ctx.ReplyID)
	if err != nil {
		fmt.Println(err)
		return
	}

	// message already has a warning.
	if cm.warned {
		bot.sendSystemMessageReply(ctx.User.ID, messages.AlreadyWarnedError, ctx.Message.MessageID)
		return
	}

	user, err := database.FindUser(bot.Db, database.ByID(cm.userID))
	if err != nil {
		fmt.Println(err)
		return
	}

	t := bot.AddWarning(cfg, user)
	cm.warned = true

	replyID, err := bot.Cache.lookupCacheMessageValue(cm.userID, ctx.ReplyID)
	if err != nil {
		fmt.Println(err)
		return
	}

	bot.sendSystemMessageReply(cm.userID, fmt.Sprintf(messages.GivenCooldownMessage, t), replyID)
}

func cmdBlacklist(bot *SecretSquirrel, ctx *BotContext) {
	// must be an admin
	if !ctx.User.IsAdmin() {
		return
	}

	reason := ctx.Message.CommandArguments()

	// reply required
	if !ctx.IsReply() {
		bot.sendSystemMessage(ctx.Message.From.ID, messages.NoReplyError)
		return
	}

	cm, err := bot.Cache.getMessage(ctx.ReplyID)
	if err != nil {
		fmt.Println(err)
		return
	}

	user, err := database.FindUser(bot.Db, database.ByID(cm.userID))
	if err != nil {
		fmt.Println(err)
		return
	}

	cm.warned = true

	bot.UpdatesUser(user, database.User{
		Rank:            database.RankBanned,
		Left:            sql.NullTime{Time: time.Now(), Valid: true},
		BlacklistReason: reason,
	})

	replyText := fmt.Sprintf(messages.BlacklistedError, user.BlacklistReason)
	if cfg.Bot.BlacklistContact != "" {
		replyText += fmt.Sprintf("\n\nContact: %s", cfg.Bot.BlacklistContact)
	}

	bot.sendSystemMessage(user.ID, replyText)
}

func cmdVersion(bot *SecretSquirrel, ctx *BotContext) {
	bot.sendSystemMessage(ctx.User.ID, fmt.Sprintf(messages.VersionMessage, BotVersion))
}
