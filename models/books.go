package models

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/wmolicki/bookler/config"
)

type Book struct {
	ID          uint
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	Name        string
	Read        bool
	Edition     string
	Description string
	// Authors     []*Author `gorm:"many2many:book_author"`
}

type BookService struct {
	db *sqlx.DB
}

func NewBookService(env *config.Env) *BookService {
	return &BookService{db: env.DB}
}

func (bs *BookService) Create(book *Book) (*Book, error) {
	query := "INSERT INTO books (name, read, edition, description) VALUES (?, ?, ?, ?)"
	result, err := bs.db.Exec(query, book.Name, book.Read, book.Edition, book.Description)
	if err != nil {
		return nil, err
	}
	bookId, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	id := uint(bookId)
	insertedBook, err := bs.GetBookById(id)
	return insertedBook, err
}

func (bs *BookService) AddAuthor(book *Book, author *Author) error {
	query := "INSERT INTO book_author (book_id, author_id) VALUES (?, ?)"
	_, err := bs.db.Exec(query, book.ID, author.ID)
	return err
}

func (bs *BookService) Update(book *Book) error {
	// return bs.db.Save(book).Error
	return nil
}

func (bs *BookService) GetBookById(id uint) (*Book, error) {
	var book Book
	query := "SELECT id, created_at, updated_at, name, read, edition, description FROM books WHERE id = ?;"
	row := bs.db.QueryRowx(query, id)

	if err := first(&book, row); err != nil {
		return nil, err
	}

	return &book, nil
}

func (bs *BookService) GetList() (*[]Book, error) {
	var books []Book

	query := "SELECT id, created_at, updated_at, name, read, edition, description FROM books;"

	err := bs.db.Select(&books, query)

	if err != nil {
		return nil, err
	}
	return &books, nil
}

func (bs *BookService) DestructiveReset() {
	// as.db.Exec("DROP TABLE books")
}
