package models

import (
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var ErrorEntityNotFound = errors.New("entity does not exists")

type Author struct {
	ID        uint
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Name      string
	BookCount int `db:"book_count"`
	// Books []*Book `gorm:"many2many:book_author"`
}

type AuthorService struct {
	db *sqlx.DB
}

func NewAuthorService(db *sqlx.DB) *AuthorService {
	return &AuthorService{db: db}
}

func (as *AuthorService) GetList() ([]*Author, error) {
	var authors []*Author

	query := "SELECT id, created_at, updated_at, name, COUNT(1) as book_count FROM authors JOIN book_author ba on authors.id = ba.author_id GROUP BY author_id;"

	err := as.db.Select(&authors, query)

	if err != nil {
		return nil, err
	}
	return authors, nil
}

func (as *AuthorService) GetByID(id uint) (*Author, error) {
	var author Author
	query := "SELECT id, created_at, updated_at, name, COUNT(1) as book_count FROM authors JOIN book_author ba on authors.id = ba.author_id WHERE id = ? GROUP BY author_id;"
	row := as.db.QueryRowx(query, id)

	if err := first(&author, row); err != nil {
		return nil, err
	}

	return &author, nil
}

func (as *AuthorService) GetBooks(authorID uint) ([]*Book, error) {
	var books []*Book

	query := "SELECT id, created_at, updated_at, name, edition, description FROM books WHERE id = ?;"

	err := as.db.Select(&books, query, authorID)

	if err != nil {
		return nil, err
	}
	return books, nil
}

func (as *AuthorService) GetByName(name string) (*Author, error) {
	var author Author
	query := "SELECT id, created_at, updated_at, name FROM authors WHERE name = ?;"
	row := as.db.QueryRowx(query, name)

	if err := first(&author, row); err != nil {
		return nil, err
	}

	return &author, nil
}

func (as *AuthorService) Create(author *Author) (*Author, error) {
	query := "INSERT INTO authors (name) VALUES (?)"
	result, err := as.db.Exec(query, author.Name)
	if err != nil {
		return nil, err
	}
	authorId, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	id := uint(authorId)
	insertedAuthor, err := as.GetByID(id)
	return insertedAuthor, err
}

func (as *AuthorService) DestructiveReset() {
	// as.db.Exec("DROP TABLE authors")
}
