package messages

import (
	"bytes"
	"secretsquirrel/database"
	"text/template"
)

const (
	newTripCode = "Tripcode set. It will appear as: <b>{{ index . 0 }}</b><code>{{ index . 1 }}</code>"

	tripcode = "<b>Tripcode</b>:{{ if .HasTrip }} <b>{{ .TripName }}</b><code>{{ .TripPass }}</code> {{ else }} unset. {{ end }}"

	userInfo = "<b>ID</b>: {{ .GetObfuscatedID }}\n<b>Username</b>: {{ .GetFormattedUsername }}\n<b>Rank</b>: {{ .Rank }}\n" +
		"<b>Karma</b>: {{ .Karma }}\n<b>Warnings</b>: {{ .Warnings}}\n" +
		"<b>Cooldown</b>:{{ if .IsInCooldown }} yes. {{ else }} no. {{ end }}"

	modUserInfo = "<b>ID</b>: {{ .GetObfuscatedID }}\n<b>Karma</b>: {{ .GetObfuscatedKarma }}\n" +
		"<b>Cooldown</b>:{{ if .IsInCooldown }} yes. {{ else }} no. {{ end }}"
)

var (
	userInfoTemplate    *template.Template
	modUserInfoTemplate *template.Template
	newTripcodeTemplate *template.Template
	tripcodeTemplate    *template.Template
)

type TripcodeConfig struct {
	HasTrip  bool
	TripName string
	TripPass string
}

func init() {
	var err error

	userInfoTemplate, err = template.New("userInfo").Parse(userInfo)
	if err != nil {
		panic(err)
	}

	modUserInfoTemplate, err = template.New("modUserInfo").Parse(modUserInfo)
	if err != nil {
		panic(err)
	}

	newTripcodeTemplate, err = template.New("newTripcode").Parse(newTripCode)
	if err != nil {
		panic(err)
	}

	tripcodeTemplate, err = template.New("tripcode").Parse(tripcode)
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
	var tBuffer bytes.Buffer

	err := tripcodeTemplate.Execute(&tBuffer, TripcodeConfig{
		HasTrip:  hasTrip,
		TripName: tripcode[0],
		TripPass: tripcode[1],
	})
	if err != nil {
		return "", err
	}

	return tBuffer.String(), nil
}
