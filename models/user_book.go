package models

import (
	"github.com/jmoiron/sqlx"
)

type UserBook struct {
	Read   bool
	Rating *int

	UserID uint `db:"user_id"`
	BookID uint `db:"book_id"`
}

type UserBookService struct {
	db *sqlx.DB
}

func NewUserBookService(db *sqlx.DB) *UserBookService {
	return &UserBookService{db: db}
}

func (ub *UserBookService) GetUserBook(b *Book, u *User) (*UserBook, error) {
	return ub.getByIDs(b.ID, u.ID)
}

func (ub *UserBookService) Read(b *Book, u *User, read bool) (*UserBook, error) {
	userBook, err := ub.getOrCreateUserBook(b, u)
	if err != nil {
		return nil, err
	}
	userBook.Read = read
	userBook, err = ub.update(userBook)
	return userBook, err
}

// Rate will add rating to the book - rating set to -1 means it's not rated yet
func (ub *UserBookService) Rate(b *Book, u *User, rating int) (*UserBook, error) {
	userBook, err := ub.getOrCreateUserBook(b, u)
	if err != nil {
		return nil, err
	}
	userBook.Rating = &rating
	userBook, err = ub.update(userBook)
	return userBook, err
}

func (ub *UserBookService) getByIDs(bookId, userId uint) (*UserBook, error) {
	var userBook UserBook

	query := "SELECT read, rating, user_id, book_id FROM user_book WHERE book_id = ? AND user_id = ?"

	row := ub.db.QueryRowx(query, bookId, userId)

	if err := first(&userBook, row); err != nil {
		return nil, err
	}
	return &userBook, nil
}

func (ub *UserBookService) getOrCreateUserBook(b *Book, u *User) (*UserBook, error) {
	var userBook *UserBook
	var err error

	if userBook, err = ub.GetUserBook(b, u); err == ErrorEntityNotFound {
		userBook, err = ub.insertUserBook(b, u)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return userBook, err
}

func (ub *UserBookService) insertUserBook(b *Book, u *User) (*UserBook, error) {
	query := "INSERT INTO user_book (user_id, book_id) VALUES (?, ?)"

	_, err := ub.db.Exec(query, u.ID, b.ID)
	if err != nil {
		return nil, err
	}
	userBook, err := ub.GetUserBook(b, u)
	return userBook, err
}

func (ub *UserBookService) update(userBook *UserBook) (*UserBook, error) {
	query := "UPDATE user_book SET read = ?, rating = ? WHERE book_id = ? AND user_id = ?"
	_, err := ub.db.Exec(query, userBook.Read, userBook.Rating, userBook.BookID, userBook.UserID)

	if err != nil {
		return nil, err
	}

	return ub.getByIDs(userBook.BookID, userBook.UserID)
}
