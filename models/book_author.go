package models

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type BookAuthor struct {
	ID     uint `db:"author_id"`
	BookID uint `db:"book_id"`
	Name   string
}

type BookAuthorService struct {
	db *sqlx.DB
	as *AuthorService
	bs *BookService
}

type AuthorWithBookCount struct {
	Author
	BookCount int `db:"book_count"`
}

func NewBookAuthorService(db *sqlx.DB, as *AuthorService, bs *BookService) *BookAuthorService {
	return &BookAuthorService{db: db, bs: bs, as: as}
}

func (ba *BookAuthorService) GetBookByID(bookID uint) (*Book, error) {
	book, err := ba.bs.GetBookByID(bookID)
	if err != nil {
		return nil, err
	}
	authors, err := ba.BookAuthors(bookID)
	if err != nil {
		return nil, err
	}
	book.Authors = authors
	return book, nil
}

func (ba *BookAuthorService) AuthorBooks(authorID uint) ([]*Book, error) {
	var books []*Book

	query := `SELECT b.id, b.created_at, b.updated_at, b.name, b.edition, b.description
		      FROM books b JOIN book_author ba ON ba.book_id = b.id WHERE ba.author_id = ?;`

	err := ba.db.Select(&books, query, authorID)

	if err != nil {
		return nil, err
	}
	return books, nil
}

func (ba *BookAuthorService) BookAuthors(bookID uint) ([]*BookAuthor, error) {
	var authors []*BookAuthor
	query := `SELECT a.id as author_id, ba.book_id as book_id, name
		      FROM authors a JOIN book_author ba ON a.id = ba.author_id WHERE ba.book_id = ?`

	if err := ba.db.Select(&authors, query, bookID); err != nil {
		return nil, err
	}

	return authors, nil
}

func (ba *BookAuthorService) AuthorsWithBookCount() ([]*AuthorWithBookCount, error) {
	var authors []*AuthorWithBookCount

	query := `SELECT id, created_at, updated_at, name, COUNT(ba.book_id) as book_count 
			  FROM authors LEFT OUTER JOIN book_author ba ON authors.id = ba.author_id
              GROUP BY id ORDER BY COUNT(ba.book_id) DESC;`

	err := ba.db.Select(&authors, query)

	if err != nil {
		return nil, err
	}
	return authors, nil
}

func (ba *BookAuthorService) UpdateBookAuthors(book *Book, authors []string) error {
	mappedAuthors := map[string]bool{}
	bookAuthors, err := ba.BookAuthors(book.ID)
	if err != nil {
		return fmt.Errorf("could not get book authors: %v", err)
	}

	newAuthors := map[string]bool{}
	for _, name := range authors {
		newAuthors[name] = true
	}

	var mappingsToRemove []*BookAuthor

	for _, ba := range bookAuthors {
		mappedAuthors[ba.Name] = true

		_, ok := newAuthors[ba.Name]
		if !ok {
			mappingsToRemove = append(mappingsToRemove, ba)
		}
	}

	for _, name := range authors {
		// skip already mapped authors
		_, ok := mappedAuthors[name]
		if ok {
			continue
		}
		author, err := ba.as.GetByName(name)
		if err != nil {
			switch err {
			case ErrorEntityNotFound:
				a := Author{Name: name}
				author, err = ba.as.Create(&a)
			default:
				return fmt.Errorf("error happened during retrieving author: %v", err)
			}
		}

		if err := ba.addAuthor(book, author); err != nil {
			return fmt.Errorf("error happened mapping book to author: %v", err)
		}
	}

	for _, bookAuthor := range mappingsToRemove {
		if err := ba.removeAuthor(bookAuthor); err != nil {
			return fmt.Errorf("error removing author from book: %v", err)
		}
	}

	return nil
}

func (ba *BookAuthorService) addAuthor(book *Book, author *Author) error {
	query := "INSERT INTO book_author (book_id, author_id) VALUES (?, ?)"
	_, err := ba.db.Exec(query, book.ID, author.ID)
	return err
}

func (ba *BookAuthorService) removeAuthor(bookAuthor *BookAuthor) error {
	query := `DELETE FROM book_author WHERE book_id = ? AND author_id = ?`
	_, err := ba.db.Exec(query, bookAuthor.BookID, bookAuthor.ID)
	return err
}
