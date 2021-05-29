package models

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Book struct {
	ID          uint
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	Name        string
	Edition     string
	Description string
	Authors     []*BookAuthor
}

type BookService struct {
	db *sqlx.DB
}

func NewBookService(db *sqlx.DB) *BookService {
	return &BookService{db: db}
}

func (bs *BookService) New(name, description, edition string, authors []string) (*Book, error) {
	b := Book{Name: name, Description: description, Edition: edition}
	book, err := bs.insert(&b)

	if err != nil {
		return nil, fmt.Errorf("error happened during creating book: %v", err)
	}

	return book, nil
}

func (bs *BookService) insert(book *Book) (*Book, error) {
	query := "INSERT INTO books (name, edition, description) VALUES (?, ?, ?)"
	result, err := bs.db.Exec(query, book.Name, book.Edition, book.Description)
	if err != nil {
		return nil, err
	}
	bookId, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	id := uint(bookId)
	insertedBook, err := bs.GetBookByID(id)
	return insertedBook, err
}

func (bs *BookService) Update(book *Book) (*Book, error) {
	query := `UPDATE books SET name = ?, edition = ?, description = ?, updated_at = ? WHERE id = ?`
	_, err := bs.db.Exec(query, book.Name, book.Edition, book.Description, book.UpdatedAt, book.ID)
	if err != nil {
		return nil, fmt.Errorf("could not update book: %v", err)
	}
	book, err = bs.GetBookByID(book.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get book: %v", err)
	}
	return book, nil
}

func (bs *BookService) GetBookByID(id uint) (*Book, error) {
	var book Book
	query := "SELECT id, created_at, updated_at, name, edition, description FROM books WHERE id = ?;"
	row := bs.db.QueryRowx(query, id)

	if err := first(&book, row); err != nil {
		return nil, err
	}

	return &book, nil
}

func (bs *BookService) GetList() ([]*Book, error) {
	var books []*Book

	query := "SELECT id, created_at, updated_at, name, edition, description FROM books;"

	err := bs.db.Select(&books, query)

	if err != nil {
		return nil, err
	}
	return books, nil
}

func (bs *BookService) DestructiveReset() {
	// as.db.Exec("DROP TABLE books")
}
