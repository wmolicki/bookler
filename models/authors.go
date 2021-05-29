package models

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Author struct {
	ID        uint
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Name      string
}

type AuthorService struct {
	db *sqlx.DB
}

func NewAuthorService(db *sqlx.DB) *AuthorService {
	return &AuthorService{db: db}
}

func (as *AuthorService) GetList() ([]*Author, error) {
	var authors []*Author

	query := "SELECT id, created_at, updated_at, name FROM authors"

	err := as.db.Select(&authors, query)

	if err != nil {
		return nil, err
	}
	return authors, nil
}

func (as *AuthorService) GetByID(id uint) (*Author, error) {
	var author Author
	query := "SELECT id, created_at, updated_at, name FROM authors WHERE id = ?;"
	row := as.db.QueryRowx(query, id)

	if err := first(&author, row); err != nil {
		return nil, err
	}

	return &author, nil
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

func (as *AuthorService) Delete(author *Author) error {
	ctx := context.Background()
	tx, err := as.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	query1 := `DELETE from book_author WHERE author_id = ?`
	query := `DELETE from authors WHERE id = ?`
	tx.Exec(query1, author.ID)
	tx.Exec(query, author.ID)

	return tx.Commit()
}

func (as *AuthorService) DestructiveReset() {
	// as.db.Exec("DROP TABLE authors")
}
