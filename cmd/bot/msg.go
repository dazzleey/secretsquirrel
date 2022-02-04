package main

import (
	"fmt"
	"secretsquirrel/database"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotMessage struct {
	User      *database.User
	MessageID int
	ReplyID   int
	Config    tgbotapi.Chattable
}

func (bot *SecretSquirrel) NewMessage(ctx *BotContext, user *database.User, msid int) *BotMessage {
	var (
		builder  strings.Builder
		err      error
		baseChat tgbotapi.BaseChat = tgbotapi.BaseChat{ChatID: user.ID}
		msg      BotMessage        = BotMessage{User: user, MessageID: msid}
	)

	if ctx.IsForward() {
		msg.Config = tgbotapi.ForwardConfig{
			BaseChat:   baseChat,
			FromChatID: ctx.Message.ForwardFrom.ID,
			MessageID:  ctx.Message.MessageID,
		}
		return &msg
	}

	// get reply ID for each user from cache based on the original Reply ID
	if ctx.ReplyID != -1 {
		msg.ReplyID, err = bot.Cache.lookupCacheMessageValue(msg.User.ID, ctx.ReplyID)
		if err != nil {
			fmt.Println(err)
		}
		baseChat.ReplyToMessageID = msg.ReplyID
	}

	// Build the text for the message.
	if ctx.Tripcode || ctx.User.ToggleTripcode {
		trip := genTripcode(ctx.User.Tripcode)
		builder.WriteString(fmt.Sprintf("<b>%s</b> <code>%s</code>\n", trip[0], trip[1]))
	}

	if ctx.Message.IsCommand() {
		builder.WriteString(ctx.Message.CommandArguments())
	} else {
		builder.WriteString(ctx.Message.Text)
	}

	if ctx.Signed {
		fmt.Fprintf(&builder, " <a href=\"tg://user?id=%d\">~~%s</a>", msg.User.ID, msg.User.GetFormattedUsername())
	}

	// create appropriate message config based on content type
	switch ctx.ContentType {
	case MessageContentType:
		msg.Config = tgbotapi.MessageConfig{
			BaseChat:              baseChat,
			Text:                  builder.String(),
			ParseMode:             "HTML",
			DisableWebPagePreview: false,
		}
	case PhotoContentType:
		msg.Config = tgbotapi.PhotoConfig{
			BaseFile: tgbotapi.BaseFile{
				BaseChat: baseChat,
				File:     tgbotapi.FileID(ctx.Message.Photo[len(ctx.Message.Photo)-1].FileID),
			},
			ParseMode: "HTML",
			Caption:   builder.String(),
		}
	case AnimationContentType:
		msg.Config = tgbotapi.AnimationConfig{
			BaseFile: tgbotapi.BaseFile{
				BaseChat: baseChat,
				File:     tgbotapi.FileID(ctx.Message.Animation.FileID),
			},
			ParseMode: "HTML",
			Caption:   builder.String(),
		}
	case VideoContentType:
		msg.Config = tgbotapi.VideoConfig{
			BaseFile: tgbotapi.BaseFile{
				BaseChat: baseChat,
				File:     tgbotapi.FileID(ctx.Message.Video.FileID),
			},
			ParseMode: "HTML",
			Caption:   builder.String(),
		}
	case AudioContentType:
		msg.Config = tgbotapi.AudioConfig{
			BaseFile: tgbotapi.BaseFile{
				BaseChat: baseChat,
				File:     tgbotapi.FileID(ctx.Message.Audio.FileID),
			},
			ParseMode: "HTML",
			Caption:   builder.String(),
		}
	case VoiceContentType:
		msg.Config = tgbotapi.VoiceConfig{
			BaseFile: tgbotapi.BaseFile{
				BaseChat: baseChat,
				File:     tgbotapi.FileID(ctx.Message.Voice.FileID),
			},
			ParseMode: "HTML",
			Caption:   builder.String(),
		}

	case DocumentContentType:
		msg.Config = tgbotapi.DocumentConfig{
			BaseFile: tgbotapi.BaseFile{
				BaseChat: baseChat,
				File:     tgbotapi.FileID(ctx.Message.Document.FileID),
			},
			ParseMode: "HTML",
			Caption:   builder.String(),
		}
	// captionless types
	case StickerContentType:
		msg.Config = tgbotapi.StickerConfig{
			BaseFile: tgbotapi.BaseFile{
				BaseChat: baseChat,
				File:     tgbotapi.FileID(ctx.Message.Sticker.FileID),
			},
		}
	case VideoNoteContentType:
		msg.Config = tgbotapi.VideoNoteConfig{
			BaseFile: tgbotapi.BaseFile{
				BaseChat: baseChat,
				File:     tgbotapi.FileID(ctx.Message.VideoNote.FileID),
			},
			Duration: ctx.Message.VideoNote.Duration,
			Length:   ctx.Message.VideoNote.Length,
		}
	case ContactContentType:
		msg.Config = tgbotapi.ContactConfig{
			BaseChat:    baseChat,
			PhoneNumber: ctx.Message.Contact.PhoneNumber,
			FirstName:   ctx.Message.Contact.FirstName,
			LastName:    ctx.Message.Contact.LastName,
		}
	case LocationContentType:
		msg.Config = tgbotapi.LocationConfig{
			BaseChat:  baseChat,
			Latitude:  ctx.Message.Location.Latitude,
			Longitude: ctx.Message.Location.Longitude,
		}
	case VenueContentType:
		msg.Config = tgbotapi.VenueConfig{
			BaseChat:  baseChat,
			Title:     ctx.Message.Venue.Title,
			Address:   ctx.Message.Venue.Address,
			Latitude:  ctx.Message.Venue.Location.Latitude,
			Longitude: ctx.Message.Venue.Location.Longitude,
		}
	}

	return &msg
}
