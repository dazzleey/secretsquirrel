package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Bot      BotConfig
	Limits   LimitsConfig
	Cooldown CooldownConfig
	Karma    KarmaConfig
	Spam     SpamConfig
}

type BotConfig struct {
	Token            string
	DatabasePath     string
	BlacklistContact string
}

type LimitsConfig struct {
	AllowContacts      bool
	AllowDocuments     bool
	AllowRemoveCommand bool
	EnableSigning      bool
	SignLimitInterval  int
	MediaLimitPeriod   int
}

type CooldownConfig struct {
	CooldownTimeBegin   []int
	CooldownTimeLinearM int
	CooldownTimeLinearB int
	WarnExpireHours     int
}

type KarmaConfig struct {
	KarmaPlusOne     int
	KarmaWarnPenalty int
}

type SpamConfig struct {
	SpamLimit           int
	SpamLimitHit        int
	SpamIntervalSeconds int

	ScoreSticker       float32
	ScoreBaseMessage   float32
	ScoreBaseForward   float32
	ScoreTextCharacter float32
	ScoreTextLineBreak float32
}

func LoadConfig(cfg *Config) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	err := viper.Unmarshal(cfg)
	if err != nil {
		log.Fatalf("Error unmarshalling config file, %s", err)
	}
}
