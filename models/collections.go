package models

import (
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
		       CASE WHEN ub.read IS NULL THEN 0 ELSE ub.read END as read, ub.rating
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
