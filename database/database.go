package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(path string) *gorm.DB {

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Panic("failed to connect to database.")
	}
	db.AutoMigrate(&SystemConfig{}, &User{})

	return db
}
