# SecretSquirrel

An attempt to rewrite [secretlounge-ng](https://github.com/secretlounge/secretlounge-ng) in Go. 

A bot to make an anonymous group chat on Telegram.

This is very much still a WIP. I'm sure there are still bugs. This isn't a drop-in replacement for secretlounge-ng. For now, use at your own risk.

# Setup

### 1. Edit the config
Copy the config example and edit it with your favorite text editor.
```
$ cp config.yaml.example config.yaml
```

### 2. @BotFather Setup
Message [@BotFather](https://t.me/BotFather) and configure your bot as follows:

* `/setprivacy`: enabled
* `/setjoingroups`: disabled
* `/setcommands`: paste the command list below:
```
start - Join the chat (start receiving messages)
users - Find out how many users are in the chat
info - Get info about your account
sign - Sign a message with your username
s - Alias of sign
tsign - Sign a message with your tripcode
t - Alias of tsign
motd - Show the welcome message
setmotd - Sets the welcome message.
version - Get version & source code of this bot
modhelp - Show commands available to moderators
adminhelp - Show commands available to admins
toggledebug - Toggle debug mode (sends back all messages to you)
togglekarma - Toggle karma notifications
toggletripcode - toggle automatic tripcodes for all your messages. 
tripcode - Show or set a tripcode for your messages
```

### 3. Run
After that just run the executable.
```
./secretsquirrel
```

## CLI
Secretsquirrel also comes with a basic CLI for user mangement that supports multiple databases.

# Build from Source

### 1. Install golang.
You can find instructions at https://go.dev/doc/install

### 2. Clone the repo.
```
git clone https://github.com/dazzleey/secretsquirrel.git
cd ./secretsquirrel
```
### 3. install the dependencies
```
go get -u github.com/go-co-op/gocron
go get -u github.com/go-telegram-bot-api/telegram-bot-api/v5
go get -u github.com/spf13/cobra
go get -u github.com/spf13/viper
go get -u gorm.io/gorm 
```

### 4. Build Executables
```
go build -o secretsquirrel ./cmd/bot
go build -o secretsqli ./cmd/cli
```