package database

import (
	"errors"

	"gorm.io/gorm"
)

type SystemConfig struct {
	Name  string `gorm:"primaryKey"`
	Value string
}

func GetMotd(db *gorm.DB) string {
	var motd SystemConfig

	err := db.Where("name = ?", "motd").First(&motd).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ""
		}
	}
	return motd.Value
}

func SetMotd(db *gorm.DB, text string) error {
	var motd SystemConfig

	err := db.Where("name = ?", "motd").First(&motd).Error
	if err != nil {
		// create the db row for 'motd' if it doesn't exist
		if errors.Is(err, gorm.ErrRecordNotFound) {
			motd := SystemConfig{
				Name:  "motd",
				Value: text,
			}
			db.Create(&motd)
			return nil
		}
		return err
	}

	motd.Value = text
	db.Save(&motd)
	return nil
}
