package models

import (
	"github.com/jmoiron/sqlx"
	"golang.org/x/oauth2"
)

type ServicesConfig func(*Services) error

func NewServices(configs ...ServicesConfig) (*Services, error) {
	var s Services
	for _, cfg := range configs {
		if err := cfg(&s); err != nil {
			return nil, err
		}
	}
	return &s, nil
}

type Services struct {
	Book        *BookService
	Author      *AuthorService
	User        *UserService
	OauthConfig *oauth2.Config
	Collections *CollectionsService
	db          *sqlx.DB
}

func WithDB(driver, dataSourceName string) ServicesConfig {
	return func(s *Services) error {
		db, err := sqlx.Open(driver, dataSourceName)
		if err != nil {
			return err
		}
		s.db = db
		return nil
	}
}

func WithAuthorService() ServicesConfig {
	return func(s *Services) error {
		as := NewAuthorService(s.db)
		s.Author = as
		return nil
	}
}

func WithBookService() ServicesConfig {
	return func(s *Services) error {
		bs := NewBookService(s.db, s.Author)
		s.Book = bs
		return nil
	}
}

func WithUserService() ServicesConfig {
	return func(s *Services) error {
		us := NewUserService(s.db)
		s.User = us
		return nil
	}
}

func WithOauthConfig(config *oauth2.Config) ServicesConfig {
	return func(s *Services) error {
		s.OauthConfig = config
		return nil
	}
}

func WithCollectionsService() ServicesConfig {
	return func(s *Services) error {
		cs := NewCollectionsService(s.db)
		s.Collections = cs
		return nil
	}
}

// Closes the database connection
func (s *Services) Close() error {
	return s.db.Close()
}
