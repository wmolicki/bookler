package config

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Env struct {
	DB *sqlx.DB
}

func NewEnv() *Env {
	db := sqlx.MustConnect("sqlite3", "./books.db")

	e := &Env{DB: db}

	return e
}
