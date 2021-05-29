package models

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

func NewCollectionsService(db *sqlx.DB) *CollectionsService {
	return &CollectionsService{db: db}
}

type CollectionsService struct {
	db *sqlx.DB
}

type BookInCollection struct {
	Book
	CollectionId uint `db:"collection_id"`
	Read         *bool
	Rating       int
}

type Collection struct {
	ID        uint
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Name      string

	Books []*BookInCollection
}

func (cs *CollectionsService) AddBook(c *Collection, b *Book) error {
	query := `INSERT INTO book_collection (collection_id, book_id) VALUES (?, ?)`
	_, err := cs.db.Exec(query, c.ID, b.ID)
	if err != nil {
		return err
	}
	return nil
}

func (cs *CollectionsService) GetByID(collectionID uint) (*Collection, error) {
	var collection Collection
	query := `
		SELECT c.id, c.created_at, c.updated_at, c.name FROM collection c 
		WHERE c.id = ?;
	`
	row := cs.db.QueryRowx(query, collectionID)

	if err := first(&collection, row); err != nil {
		return nil, err
	}

	return &collection, nil
}

func (cs *CollectionsService) Create(user *User, name string) (*Collection, error) {
	query := `INSERT INTO collection (name, user_id) VALUES(?, ?)`
	result, err := cs.db.Exec(query, name, user.ID)
	if err != nil {
		return nil, err
	}

	collectionID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	id := uint(collectionID)

	c, err := cs.GetByID(id)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (cs *CollectionsService) Update(c *Collection) (*Collection, error) {
	query := `UPDATE collection SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := cs.db.Exec(query, c.Name, c.ID)
	if err != nil {
		return nil, err
	}

	c, err = cs.GetByID(c.ID)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (cs *CollectionsService) Delete(c *Collection) error {
	ctx := context.Background()
	tx, err := cs.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	query1 := `DELETE FROM book_collection WHERE collection_id = ?`
	query2 := `DELETE FROM collection WHERE id = ?`

	tx.Exec(query1, c.ID)
	tx.Exec(query2, c.ID)

	return tx.Commit()
}

func (cs *CollectionsService) DeleteBook(c *Collection, b *Book) error {
	ctx := context.Background()
	tx, err := cs.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	query1 := `DELETE FROM book_collection WHERE collection_id = ? AND book_id = ?`
	tx.Exec(query1, c.ID, b.ID)

	return tx.Commit()
}

func (cs *CollectionsService) GetWithBooks(collectionID uint) (*Collection, error) {
	c, err := cs.GetByID(collectionID)
	if err != nil {
		return nil, err
	}

	booksQ := `
		SELECT b.id, b.name, b.edition, b.description, b.created_at, b.updated_at, bc.collection_id, 
		       CASE WHEN ub.read IS NULL THEN 0 ELSE ub.read END as read, 
		       CASE WHEN ub.rating IS NULL THEN -1 ELSE ub.rating END as rating
		FROM book_collection bc
		JOIN books b on bc.book_id = b.id
		JOIN collection c on bc.collection_id = c.id
		LEFT JOIN user_book ub on b.id = ub.book_id
		WHERE c.id = ?;
	`

	books := make([]*BookInCollection, 0)

	err = cs.db.Select(&books, booksQ, c.ID)
	if err != nil {
		return nil, err
	}

	c.Books = books

	return c, nil
}

func (cs *CollectionsService) List(user *User) ([]*Collection, error) {
	var collections []*Collection

	query := `
		SELECT c.id, c.created_at, c.updated_at, c.name FROM collection c 
		WHERE c.user_id = ?;
	`

	err := cs.db.Select(&collections, query, user.ID)
	if err != nil {
		return nil, err
	}

	colMap := map[uint]*Collection{}
	for _, c := range collections {
		colMap[c.ID] = c
	}

	booksQ := `
		SELECT b.id, b.name, b.edition, b.description, b.created_at, b.updated_at, bc.collection_id, 
		       CASE WHEN ub.read IS NULL THEN 0 ELSE ub.read END as read, 
		       CASE WHEN ub.rating IS NULL THEN -1 ELSE ub.rating END as rating
		FROM book_collection bc
		JOIN books b on bc.book_id = b.id
		JOIN collection c on bc.collection_id = c.id
		LEFT JOIN user_book ub on b.id = ub.book_id
		WHERE c.user_id = ?;
	`

	books := make([]*BookInCollection, 0)

	err = cs.db.Select(&books, booksQ, user.ID)
	if err != nil {
		return nil, err
	}

	for _, b := range books {
		c, ok := colMap[b.CollectionId]
		if !ok {
			panic(fmt.Sprintf("there is no collectionId %v in returned collections", b.CollectionId))
		}
		c.Books = append(c.Books, b)
	}

	return collections, nil
}
