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
	as *AuthorService
	ba *BookAuthorService
}

func NewBookService(db *sqlx.DB, as *AuthorService, ba *BookAuthorService) *BookService {
	return &BookService{db: db, as: as, ba: ba}
}

func (bs *BookService) New(name, description, edition string, authors []string) (*Book, error) {
	b := Book{Name: name, Description: description, Edition: edition}
	book, err := bs.insert(&b)

	if err != nil {
		return nil, fmt.Errorf("error happened during creating book: %v", err)
	}

	for _, authorName := range authors {
		// TODO: Maybe this author is already mapped to this book (duplicates?)
		author, err := bs.as.GetByName(authorName)
		if err != nil {
			switch err {
			case ErrorEntityNotFound:
				a := Author{Name: authorName}
				author, err = bs.as.Create(&a)
			default:
				return nil, fmt.Errorf("error happened during retrieving author: %v", err)
			}
		}

		if err := bs.addAuthor(book, author); err != nil {
			return nil, fmt.Errorf("error happened mapping book to author: %v", err)
		}
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
	insertedBook, err := bs.GetBookById(id)
	return insertedBook, err
}

func (bs *BookService) addAuthor(book *Book, author *Author) error {
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
	query := "SELECT id, created_at, updated_at, name, edition, description FROM books WHERE id = ?;"
	row := bs.db.QueryRowx(query, id)

	if err := first(&book, row); err != nil {
		return nil, err
	}

	authors, err := bs.ba.BookAuthors(book.ID)
	if err != nil {
		return nil, err
	}
	book.Authors = authors

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
