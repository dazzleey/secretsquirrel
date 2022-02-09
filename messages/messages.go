package messages

const (
	UserJoined        = "<i>You joined the chat!</i>"
	UserLeft          = "<i>You left the chat!</i>"
	UserAlreadyInChat = "<i>You are already in the chat.</i>"
	UserNotInChat     = "<i>You are not in the chat yet. Use /start to join!</i>"
	GivenCooldown     = "<i>You've been handed a cooldown of %s for this message (message also deleted)</i>"
	MessageDeleted    = "<i>Your message has been deleted. No cooldown has been given this time, but refrain from posting it again.</i>"
	PromotedMod       = "<i>You've been promoted to moderator, run /modhelp for a list of commands.</i>"
	PromotedAdmin     = "<i>You've been promoted to admin, run /adminhelp for a list of commands.</i>"
	KarmaThank        = "<i>You just gave this user some sweet karma, awesome!</i>"
	KarmaReceived     = "<i>You've just been given sweet karma! (check /info to see your karma or /toggleKarma to turn these notifications off)</i>"
	Version           = "Secretsquirrel version %s - https://github.com/dazzleey/secretsquirrel"
)

// Errors
const (
	CommandDisabled   = "<i>This command has been disabled.</i>"
	NoReply           = "<i>You need to reply to a message to use this command.</i>"
	NotInCache        = "<i>Message not found in cache... (24h passed or the bot was restarted)</i>"
	NoUser            = "<i>User not found.</i>"
	NoUserByName      = "<i>No user found by that name.</i>"
	NoUserById        = "<i>No user found by that id! Note that all ids rotate every 24 hours.</i>"
	Cooldown          = "<i>You're on cooldown. Your cooldown expires at %s</i>"
	AlreadyWarned     = "<i>A warning has already been issued for this message.</i>"
	NotInCooldown     = "<i>This user is not in a cooldown right now.</i>"
	Blacklisted       = "<i>You've been blacklisted.  reason: %s</i>"
	AlreadyUpvoted    = "<i>You have already upvoted this message.</i>"
	UpvoteOwnMessage  = "<i>You can't upvote your own message.</i>"
	Spam              = "<i>Your message has not been sent. Avoid sending messages too fast, try again later.</i>"
	SpamSign          = "<i>Your message has not been sent. Avoid using /sign too often, try again later.</i>"
	InvalidTripFormat = "<i>Given tripcode is not valid, the format is <code>name#pass</code></i>"
	NoTripcode        = "<i>You don't have a tripcode set.</i>"
	MediaLimit        = "<i>You can't send media or forward messages at this time, try again later.</i>"
	SigningDisabled   = "<i>Signing is disabled.</i>"
)

// Help
const (
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
)
