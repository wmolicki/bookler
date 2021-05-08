package models

import (
	"github.com/wmolicki/bookler/config"
	"gorm.io/gorm"
)

type Book struct {
	gorm.Model
	Name        string `gorm:"unique_index"`
	Read        bool
	Edition     string
	Description string
	Authors     []*Author `gorm:"many2many:book_author"`
}

type BookService struct {
	db *gorm.DB
}

func NewBookService(env *config.Env) *BookService {
	return &BookService{db: env.DB}
}

func (bs *BookService) Create(book *Book) error {
	return bs.db.Create(book).Error
}

func (bs *BookService) Update(book *Book) error {
	return bs.db.Save(book).Error
}

func (bs *BookService) GetBookById(id uint) (*Book, error) {
	var book Book
	if db := bs.db.First(&book, id); db.Error != nil {
		return nil, db.Error
	}
	return &book, nil
}

func (bs *BookService) GetBooks() (*[]Book, error) {
	var books []Book
	db := bs.db.Preload("Authors").Find(&books)
	if err := db.Error; err != nil {
		return nil, err
	}
	return &books, nil
}

func (bs *BookService) DestructiveReset() {
	bs.db.Migrator().DropTable(&Book{})
	bs.db.Migrator().AutoMigrate(&Book{})
}
