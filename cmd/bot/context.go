package main

import (
	"fmt"
	"secretsquirrel/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ContentType int

const (
	MessageContentType ContentType = iota
	StickerContentType
	AnimationContentType

	PhotoContentType
	VideoContentType
	AudioContentType
	VoiceContentType
	DocumentContentType
	VideoNoteContentType
	ContactContentType
	LocationContentType
	VenueContentType
)

type BotContext struct {
	User           *database.User
	ContentType    ContentType
	Update         *tgbotapi.Update
	Message        *tgbotapi.Message
	ReplyID        int
	CacheMessageID int
	Signed         bool
	Tripcode       bool
}

func (ctx *BotContext) HasFile() bool {
	return ctx.ContentType > 2
}

func (ctx *BotContext) IsReply() bool {
	return ctx.Message.ReplyToMessage != nil
}

func (ctx *BotContext) IsForward() bool {
	return ctx.Message.ForwardFrom != nil || ctx.Message.ForwardFromChat != nil
}

func (ctx *BotContext) UserLeftOrKicked() bool {
	//  update.MyChatMember is only present in updates on join/leave.
	return ctx.Update.MyChatMember != nil && ctx.Update.MyChatMember.NewChatMember.Status == "kicked"
}

func (bot *SecretSquirrel) CreateContext(u tgbotapi.Update) *BotContext {
	var err error
	var user *database.User

	if cacheUser, ok := (*bot.Users)[u.Message.From.ID]; ok {
		user = &cacheUser
	} else {
		user, _ = database.FindUser(bot.Db, database.ByID(u.Message.From.ID))
	}

	ctx := BotContext{
		User:           user,
		ContentType:    MessageContentType,
		ReplyID:        -1,
		CacheMessageID: -1,
		Signed:         false,
		Tripcode:       false,
	}

	if u.Message != nil {
		ctx.Update = &u
		ctx.Message = u.Message

		switch {
		case ctx.Update.Message.Sticker != nil:
			ctx.ContentType = StickerContentType
		case ctx.Update.Message.Animation != nil:
			ctx.ContentType = AnimationContentType
		case ctx.Update.Message.Photo != nil:
			ctx.ContentType = PhotoContentType
		case ctx.Update.Message.Video != nil:
			ctx.ContentType = VideoContentType
		case ctx.Update.Message.Voice != nil:
			ctx.ContentType = VoiceContentType
		case ctx.Update.Message.VideoNote != nil:
			ctx.ContentType = VideoContentType
		case ctx.Update.Message.Audio != nil:
			ctx.ContentType = AudioContentType
		case ctx.Update.Message.Location != nil:
			ctx.ContentType = LocationContentType
		case ctx.Update.Message.Venue != nil:
			ctx.ContentType = VenueContentType
		case ctx.Update.Message.Contact != nil:
			ctx.ContentType = ContactContentType
		case ctx.Update.Message.Document != nil:
			ctx.ContentType = DocumentContentType
		default:
			ctx.ContentType = MessageContentType
		}

		// check if message is a reply.
		// If it is, look it up in the MessageCache
		if ctx.IsReply() {
			ctx.ReplyID, err = bot.Cache.lookupCacheMessageKey(ctx.Message.From.ID, ctx.Message.ReplyToMessage.MessageID)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	return &ctx
}
