package models

import (
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/wmolicki/bookler/random"
)

func NewUserService(db *sqlx.DB) *UserService {
	return &UserService{db}
}

const AuthCookieName = "bookler_sid"

type User struct {
	ID              uint      `db:"id"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
	Email           string    `db:"email"`
	Name            string    `db:"name"`
	ProfileImageUrl string    `db:"profile_image_url"`
	RememberToken   string    `db:"remember_token"`
}

type UserService struct {
	db *sqlx.DB
}

type UserDB interface {
	ByEmail(string) (*User, error)
	ByID(uint) (*User, error)
	ByRememberToken(string) (*User, error)

	Create(*User) error
}

func (us *UserService) SignIn(user *User) (string, error) {
	token, err := random.RememberToken()
	if err != nil {
		return "", err
	}
	user.RememberToken = token
	_, err = us.Update(user)
	if err != nil {
		return "", err
	}
	return token, err
}

func (us *UserService) ByEmail(email string) (*User, error) {
	var user User
	query := "SELECT id, created_at, updated_at, name, email, profile_image_url FROM user WHERE email = ?;"
	row := us.db.QueryRowx(query, email)

	if err := first(&user, row); err != nil {
		return nil, err
	}

	return &user, nil
}

func (us *UserService) ByRememberToken(token string) (*User, error) {
	var user User
	query := "SELECT id, created_at, updated_at, name, email, profile_image_url, remember_token FROM user WHERE remember_token = ?;"
	row := us.db.QueryRowx(query, token)

	if err := first(&user, row); err != nil {
		return nil, err
	}

	return &user, nil
}

func (us *UserService) ByID(id uint) (*User, error) {
	var user User
	query := "SELECT id, created_at, updated_at, name, email, profile_image_url, remember_token FROM user WHERE id = ?;"
	row := us.db.QueryRowx(query, id)

	if err := first(&user, row); err != nil {
		return nil, err
	}

	return &user, nil
}

func (us *UserService) Create(email, name, profileImageUrl string) (*User, error) {
	user := &User{Email: email, Name: name, ProfileImageUrl: profileImageUrl}
	insertedUser, err := us.insert(user)
	if err != nil {
		return nil, err
	}
	return insertedUser, nil
}

func (us *UserService) Update(user *User) (*User, error) {
	u, err := us.update(user)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (us *UserService) insert(user *User) (*User, error) {
	query := "INSERT INTO user (email, name, profile_image_url) VALUES (?, ?, ?)"
	_, err := us.db.Exec(query, user.Email, user.Name, user.ProfileImageUrl)
	if err != nil {
		return nil, err
	}
	insertedUser, err := us.ByEmail(user.Email)
	return insertedUser, err
}

func (us *UserService) update(user *User) (*User, error) {
	query := "UPDATE user SET email=?, name=?, profile_image_url=?, remember_token=?, updated_at=CURRENT_TIMESTAMP WHERE id = ?;"
	_, err := us.db.Exec(query, user.Email, user.Name, user.ProfileImageUrl, user.RememberToken, user.ID)
	if err != nil {
		return nil, err
	}
	updatedUser, err := us.ByID(user.ID)
	return updatedUser, err
}

//type UserSQLX struct {
//	db *sqlx.DB
//}
