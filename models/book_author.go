package models

import "github.com/jmoiron/sqlx"

type BookAuthor struct {
	ID   uint
	Name string
}

type BookAuthorService struct {
	db *sqlx.DB
}

type AuthorWithBookCount struct {
	Author
	BookCount int `db:"book_count"`
}

func NewBookAuthorService(db *sqlx.DB) *BookAuthorService {
	return &BookAuthorService{db: db}
}

func (ba *BookAuthorService) AuthorBooks(authorID uint) ([]*Book, error) {
	var books []*Book

	query := "SELECT id, created_at, updated_at, name, edition, description FROM books WHERE id = ?;"

	err := ba.db.Select(&books, query, authorID)

	if err != nil {
		return nil, err
	}
	return books, nil
}

func (ba *BookAuthorService) BookAuthors(bookID uint) ([]*BookAuthor, error) {
	var authors []*BookAuthor
	query := "SELECT id, name FROM authors a JOIN book_author ba ON a.id = ba.author_id WHERE ba.book_id = ?"

	if err := ba.db.Select(&authors, query, bookID); err != nil {
		return nil, err
	}

	return authors, nil
}

func (ba *BookAuthorService) AuthorsWithBookCount() ([]*AuthorWithBookCount, error) {
	var authors []*AuthorWithBookCount

	query := "SELECT id, created_at, updated_at, name, COUNT(1) as book_count FROM authors JOIN book_author ba on authors.id = ba.author_id GROUP BY author_id;"

	err := ba.db.Select(&authors, query)

	if err != nil {
		return nil, err
	}
	return authors, nil
}
