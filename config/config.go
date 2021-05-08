package config

import (
	"log"

	"gorm.io/gorm/logger"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Env struct {
	DB *gorm.DB
}

func NewEnv() *Env {
	db, err := gorm.Open(sqlite.Open("./books_v2.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("could not open db: %v", err)
	}

	e := &Env{DB: db}

	return e
}
