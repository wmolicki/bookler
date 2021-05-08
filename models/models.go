package models

import "time"

//type Book struct {
//	AutoId
//	BookName string `db:"name"`
//	Read     bool
//}
//
type Author struct {
	Id   int
	Name string
}

//type BookAuthor struct {
//	BookId   int `db:"book_id"`
//	AuthorId int `db:"author_id"`
//}

type Book struct {
	BookId      int    `db:"id"`
	BookName    string `db:"name"`
	Edition     *string
	Description *string
	Read        bool
	Authors     []Author
	Added       time.Time `db:"created_on"`
}
