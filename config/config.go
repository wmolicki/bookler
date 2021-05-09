package config

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"

	_ "github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Env struct {
	DB *sqlx.DB
}

func NewEnv() *Env {
	db, err := sqlx.Open("sqlite3", "./books_v2.db")

	fmt.Println(os.Getwd())

	if err != nil {
		log.Fatalf("could not open db: %v", err)
	}

	e := &Env{DB: db}

	return e
}
