package models

import (
	"github.com/wmolicki/bookler/config"
	"gorm.io/gorm"
)

type Author struct {
	gorm.Model
	Name  string  `gorm:"uniqueIndex"`
	Books []*Book `gorm:"many2many:book_author"`
}

type AuthorService struct {
	db *gorm.DB
}

func NewAuthorService(env *config.Env) *AuthorService {
	return &AuthorService{db: env.DB}
}

func (as *AuthorService) GetAuthors() (*[]Author, error) {
	var authors []Author
	db := as.db.Preload("Books").Find(&authors)
	if err := db.Error; err != nil {
		return nil, err
	}
	return &authors, nil
}

func (as *AuthorService) DestructiveReset() {
	as.db.Migrator().DropTable(&Author{})
	as.db.Migrator().AutoMigrate(&Author{})
}
