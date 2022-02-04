package database

import (
	"database/sql"
	"fmt"
	"math"
	"math/rand"
	"secretsquirrel/util"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type UserRank int

const (
	RankBanned UserRank = iota // 0
	RankUser                   // 1
	RankMod                    // 2
	RankAdmin                  // 3
)

func (r UserRank) String() string {
	switch r {
	case 0:
		return "Banned"
	case 1:
		return "User"
	case 2:
		return "Mod"
	case 3:
		return "Admin"
	default:
		return fmt.Sprintf("%d", r)
	}
}

type User struct {
	ID              int64 `gorm:"primaryKey"`
	UserName        string
	RealName        string
	Rank            UserRank
	Joined          time.Time
	Left            sql.NullTime
	LastActive      time.Time
	CooldownUntil   sql.NullTime
	BlacklistReason string
	Warnings        int
	Karma           int
	HideKarma       bool
	DebugEnabled    bool
	Tripcode        string
	ToggleTripcode  bool
}

func (u *User) GetFormattedUsername() string {
	if u.UserName != "" {
		return fmt.Sprintf("@%s", u.UserName)
	} else {
		return u.RealName
	}
}

func (u *User) IsInCooldown() bool {
	if u.CooldownUntil.Valid && time.Now().Before(u.CooldownUntil.Time) {
		return true
	}
	return false
}

func (u *User) IsPrivileged() bool {
	return u.Rank == RankMod || u.Rank == RankAdmin
}
func (u *User) IsAdmin() bool {
	return u.Rank == RankAdmin
}

func (u *User) IsMod() bool {
	return u.Rank == RankMod
}

func (u *User) IsBlacklisted() bool {
	return u.Rank == RankBanned
}

func (u *User) GetObfuscatedID() string {
	var (
		alpha  string = "0123456789abcdefghijklmnopqrstuv"
		obfsID string
		salt   int
	)

	salt = int(util.ToOrdinal(time.Now()))
	if salt&0xff == 0 {
		salt = salt >> 8
	}

	value := (int(u.ID) * salt) & 0xffffff
	for _, n := range []int{value, value >> 5, value >> 10, value >> 15} {
		obfsID += string(alpha[n%32])
	}
	return obfsID
}

func (u *User) GetObfuscatedKarma() int {
	offset := int(math.Round(math.Abs(float64(u.Karma)*0.2) + 2))
	return u.Karma + rand.Intn(offset+1) - offset
}

func AreJoined(db *gorm.DB) *gorm.DB {
	return db.Not("rank = ?", RankBanned).Where("left IS NULL")
}

func ByUsernameOrID(u string) func(db *gorm.DB) *gorm.DB {

	userID, _ := strconv.ParseInt(u, 10, 64)

	if userID != 0 {
		return func(db *gorm.DB) *gorm.DB {
			return db.Where("id = ?", userID)
		}

	} else {
		return func(db *gorm.DB) *gorm.DB {
			return db.Where("user_name = ?", u)
		}
	}
}

func ByUsername(username string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_name = ?", strings.ToLower(username))
	}
}

func ByID(userID int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", userID)
	}
}

func FindUser(db *gorm.DB, scopes ...func(*gorm.DB) *gorm.DB) (*User, error) {
	var user User

	if err := db.Scopes(scopes...).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func FindUsers(db *gorm.DB, scopes ...func(*gorm.DB) *gorm.DB) ([]User, error) {
	var users []User

	if err := db.Scopes(scopes...).Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func NewUser(db *gorm.DB, tgUser *tgbotapi.User) *User {
	user := User{
		ID:              tgUser.ID,
		UserName:        tgUser.UserName,
		RealName:        tgUser.FirstName + " " + tgUser.LastName,
		Rank:            RankUser,
		Joined:          time.Now(),
		Left:            sql.NullTime{},
		LastActive:      time.Time{},
		CooldownUntil:   sql.NullTime{},
		BlacklistReason: "",
		Warnings:        0,
		Karma:           0,
		HideKarma:       false,
		DebugEnabled:    false,
		Tripcode:        "",
		ToggleTripcode:  false,
	}
	return &user
}
