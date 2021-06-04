package models

import (
	"context"
	"fmt"
	"strings"
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

	Tags    []string
	Authors []*BookAuthor
}

type BookRepo interface {
	ByID(uint) (*Book, error)
	ByName(string) (*Book, error)
	List() ([]*Book, error)

	Create(*Book) error
	Update(*Book) error
	Delete(*Book) error
}

type BookService interface {
	BookRepo
	New(name, description, edition string, authors []string) (*Book, error)
}

type bookService struct {
	BookRepo
}

func NewBookService(db *sqlx.DB) BookService {
	bd := &bookDB{db}
	return &bookService{bd}
}

func (bs *bookService) New(name, description, edition string, authors []string) (*Book, error) {
	b := Book{Name: name, Description: description, Edition: edition}
	err := bs.Create(&b)

	if err != nil {
		return nil, fmt.Errorf("error happened during creating book: %v", err)
	}

	return &b, nil
}

type bookDB struct {
	db *sqlx.DB
}

func (bd *bookDB) Create(book *Book) error {
	query := "INSERT INTO books (name, edition, description) VALUES (?, ?, ?)"
	result, err := bd.db.Exec(query, book.Name, book.Edition, book.Description)
	if err != nil {
		return err
	}
	bookId, err := result.LastInsertId()
	if err != nil {
		return err
	}
	book.ID = uint(bookId)
	return nil
}

func (bd *bookDB) Update(book *Book) error {
	query := `UPDATE books SET name = ?, edition = ?, description = ?, updated_at = ? WHERE id = ?`
	_, err := bd.db.Exec(query, book.Name, book.Edition, book.Description, book.UpdatedAt, book.ID)
	if err != nil {
		return fmt.Errorf("could not update book: %v", err)
	}
	return nil
}

func (bd *bookDB) Delete(book *Book) error {
	ctx := context.Background()
	tx, err := bd.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	query1 := `DELETE FROM book_collection WHERE book_id = ?`
	query2 := `DELETE FROM book_author WHERE book_id = ?`
	query3 := `DELETE FROM user_book WHERE book_id = ?`
	query4 := `DELETE FROM books WHERE id = ?`

	bd.db.Exec(query1, book.ID)
	bd.db.Exec(query2, book.ID)
	bd.db.Exec(query3, book.ID)
	bd.db.Exec(query4, book.ID)

	return tx.Commit()
}

func (bd *bookDB) ByID(id uint) (*Book, error) {
	var queryModel struct {
		RawTags string `db:"tags"`
		Book
	}

	query := "SELECT id, created_at, updated_at, name, edition, description, tags FROM books WHERE id = ?;"
	row := bd.db.QueryRowx(query, id)

	if err := first(&queryModel, row); err != nil {
		return nil, err
	}

	if queryModel.RawTags != "" {
		queryModel.Book.Tags = strings.Split(queryModel.RawTags, ",")
	}

	return &queryModel.Book, nil
}

func (bd *bookDB) ByName(name string) (*Book, error) {
	var book Book
	query := "SELECT id, created_at, updated_at, name, edition, description FROM books WHERE name = ?;"
	row := bd.db.QueryRowx(query, name)

	if err := first(&book, row); err != nil {
		return nil, err
	}

	return &book, nil
}

func (bd *bookDB) List() ([]*Book, error) {
	var books []*Book

	query := "SELECT id, created_at, updated_at, name, edition, description FROM books;"

	err := bd.db.Select(&books, query)

	if err != nil {
		return nil, err
	}
	return books, nil
}

func (bd *bookDB) DestructiveReset() {
	// as.db.Exec("DROP TABLE books")
}
