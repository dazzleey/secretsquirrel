package messages

import (
	"bytes"
	"secretsquirrel/database"
	"text/template"
)

const (
	UserJoinedMessage        = "<i>You joined the chat!</i>"
	UserLeftMessage          = "<i>You left the chat!</i>"
	UserInChatMessage        = "<i>You are already in the chat.</i>"
	UserNotInChatMessage     = "<i>You are not in the chat yet. Use /start to join!</i>"
	GivenCooldownMessage     = "<i>You've been handed a cooldown of %s for this message (message also deleted)</i>"
	MessageDeletedMessage    = "<i>Your message has been deleted. No cooldown has been given this time, but refrain from posting it again.</i>"
	PromotedModMessage       = "<i>You've been promoted to moderator, run /modhelp for a list of commands.</i>"
	PromotedAdminMessage     = "<i>You've been promoted to admin, run /adminhelp for a list of commands.</i>"
	KarmaThankMessage        = "<i>You just gave this user some sweet karma, awesome!</i>"
	KarmaNotificationMessage = "<i>You've just been given sweet karma! (check /info to see your karma or /toggleKarma to turn these notifications off)</i>"
	VersionMessage           = "Secretsquirrel version %s - https://github.com/dazzleey/secretsquirrel"

	CommandDisabledError   = "This command has been disabled."
	NoReplyError           = "You need to reply to a message to use this command."
	NotInCacheError        = "Message not found in cache... (24h passed or bot was restarted)"
	NoUserError            = "User not found."
	NoUserByNameError      = "No user found by that name."
	NoUserByIdError        = "No user found by that id! Note that all ids rotate every 24 hours."
	CooldownError          = "You're on cooldown. Your cooldown expires at <i>%s</i>"
	AlreadyWarnedError     = "A warning has already been issued for this message."
	NotInCooldownError     = "This user is not in a cooldown right now."
	BlacklistedError       = "You've been blacklisted.  reason: %s"
	AlreadyUpvotedError    = "You have already upvoted this message."
	UpvoteOwnMessageError  = "You can't upvote your own message."
	SpamError              = "Your message has not been sent. Avoid sending messages too fast, try again later."
	SpamSignError          = "Your message has not been sent. Avoid using /sign too often, try again later."
	InvalidTripFormatError = "Given tripcode is not valid, the format is <code>name#pass</code>"
	NoTripcodeError        = "You don't have a tripcode set."
	MediaLimitError        = "You can't send media or forward messages at this time, try again later."
	SigningDisabledError   = "Signing is disabled."

	ModeratorHelp = `<i>Moderators can use the following commands</i>:
	/modhelp - show this text
	/modsay &lt;message&gt; - send an official moderator message

<i>Or reply to a message and use</i>:
	/info - get info about the user that sent this message
	/warn - warn the user that sent this message (cooldown)
	/delete - delete this message and warn the user
	/remove - delete a message without giving a cooldown/warning`

	AdminHelp = `<i>Admins can use the following commands</i>:
	/adminhelp - show this text
	/adminsay &lt;message&gt; - send an official moderator message
	/setmotd &lt;message&gt; - set the welcome message (HTML formatted)
	/uncooldown &lt;id | username&gt - remove a cooldown from a user
	/mod &lt;username&gt; - promote a user to moderator
	/admin &lt;username&gt; - promote a user to admin
	
	/blacklist &lt;reason&gt; - blacklist the user who sent this message`

	// Templates
	newTripCodeMessage = "Tripcode set. It will appear as: <b>{{ index . 0 }}</b><code>{{ index . 1 }}</code>"
	tripcodeMessage    = "<b>Tripcode</b>:{{ if .HasTrip }} <b>{{ .TripName }}</b><code>{{ .TripPass }}</code> {{ else }} unset. {{ end }}"

	userInfoMessage = "<b>ID</b>: {{ .GetObfuscatedID }}\n<b>Username</b>: {{ .GetFormattedUsername }}\n<b>Rank</b>: {{ .Rank }}\n" +
		"<b>Karma</b>: {{ .Karma }}\n<b>Warnings</b>: {{ .Warnings}}\n" +
		"<b>Cooldown</b>:{{ if .IsInCooldown }} yes. {{ else }} no. {{ end }}"

	modUserInfoMessage = "<b>ID</b>: {{ .GetObfuscatedID }}\n<b>Karma</b>: {{ .GetObfuscatedKarma }}\n" +
		"<b>Cooldown</b>:{{ if .IsInCooldown }} yes. {{ else }} no. {{ end }}"
)

type tripcodeTemplateConfig struct {
	HasTrip  bool
	TripName string
	TripPass string
}

var (
	userInfoTemplate    *template.Template
	modUserInfoTemplate *template.Template
	newTripcodeTemplate *template.Template
	tripcodeTemplate    *template.Template
)

func init() {
	var err error

	userInfoTemplate, err = template.New("userInfo").Parse(userInfoMessage)
	if err != nil {
		panic(err)
	}

	modUserInfoTemplate, err = template.New("modUserInfo").Parse(modUserInfoMessage)
	if err != nil {
		panic(err)
	}

	newTripcodeTemplate, err = template.New("newTripcode").Parse(newTripCodeMessage)
	if err != nil {
		panic(err)
	}

	tripcodeTemplate, err = template.New("tripcode").Parse(tripcodeMessage)
	if err != nil {
		panic(err)
	}
}

func UserInfo(user *database.User) (string, error) {
	var tBuffer bytes.Buffer

	err := userInfoTemplate.Execute(&tBuffer, user)
	if err != nil {
		return "", err
	}

	return tBuffer.String(), nil
}

func ModUserInfo(user *database.User) (string, error) {
	var tBuffer bytes.Buffer

	err := modUserInfoTemplate.Execute(&tBuffer, user)
	if err != nil {
		return "", err
	}

	return tBuffer.String(), nil
}

func NewTripcodeMessage(tripcode []string) (string, error) {
	var tBuffer bytes.Buffer

	err := newTripcodeTemplate.Execute(&tBuffer, tripcode)
	if err != nil {
		return "", err
	}

	return tBuffer.String(), nil
}

func TripcodeMessage(hasTrip bool, tripcode []string) (string, error) {
	var (
		tBuffer bytes.Buffer
	)

	err := tripcodeTemplate.Execute(&tBuffer, tripcodeTemplateConfig{
		HasTrip:  hasTrip,
		TripName: tripcode[0],
		TripPass: tripcode[1],
	})
	if err != nil {
		return "", err
	}

	return tBuffer.String(), nil
}
