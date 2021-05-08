package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator"
	"github.com/wmolicki/bookler/config"
	"github.com/wmolicki/bookler/models"
)

type bookAuthorQuery struct {
	BookId      int       `db:"book_id"`
	BookName    string    `db:"book_name"`
	Read        bool      `db:"book_read"`
	BookAdded   time.Time `db:"book_added"`
	BookEdition *string   `db:"book_edition"`
	AuthorId    int       `db:"author_id"`
	AuthorName  string    `db:"author_name"`
}

var validate *validator.Validate

type BookHandler struct {
	Env *config.Env
}

func (h *BookHandler) Index(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT b.id as book_id, 
		       b.name as book_name, 
		       b.read as book_read,
		       b.created_on as book_added,
		       b.edition as book_edition,
		       a.id as author_id, 
		       a.name as author_name
		FROM book b
				 JOIN book_author ba ON ba.book_id = b.id
				 JOIN author a ON a.id = ba.author_id;
		`
	var baq []bookAuthorQuery
	err := h.Env.DB.Select(&baq, query)
	if err != nil {
		log.Fatalf("could not query, %v", err)
	}

	books := map[int]*models.Book{}

	for _, ba := range baq {
		book, ok := books[ba.BookId]
		if !ok {
			book = &models.Book{BookId: ba.BookId, BookName: ba.BookName, Authors: []models.Author{}, Added: ba.BookAdded, Edition: ba.BookEdition}
			books[ba.BookId] = book
		}
		book.Authors = append(book.Authors, models.Author{Id: ba.AuthorId, Name: ba.AuthorName})
	}

	var bookList []models.Book

	for _, v := range books {
		bookList = append(bookList, *v)
	}

	b, _ := json.Marshal(bookList)

	w.Write(b)
}

type addBookRequestBody struct {
	Name        string   `json:"name" validate:"required"`
	Authors     []string `json:"authors" validate:"required"`
	Read        *bool    `json:"read" validate:"required"`
	Edition     *string  `json:"edition"`
	Description *string  `json:"description"`
}

func getOrCreateBookId(db *sqlx.DB, body *addBookRequestBody) (int, error) {
	bookQuery := `SELECT id FROM book WHERE name = ?;`
	row := db.QueryRow(bookQuery, body.Name)
	var bookId int
	err := row.Scan(&bookId)
	if err == sql.ErrNoRows {
		bookInsert := `INSERT INTO book (name, read, description, edition) VALUES (?, ?, ?, ?);`
		res, err := db.Exec(bookInsert, body.Name, body.Read, body.Description, body.Edition)
		if err != nil {
			return 0, err
		}
		bookId, err := res.LastInsertId()
		if err != nil {
			return 0, err
		}
		return int(bookId), err
	} else if err != nil {
		log.Fatalf("could not read book table: %v", err)
	}

	return bookId, err
}

func getOrCreateAuthorId(db *sqlx.DB, authorName string) (int, error) {
	authorQuery := `SELECT id FROM author WHERE name = ?;`
	row := db.QueryRow(authorQuery, authorName)
	var authorId int
	err := row.Scan(&authorId)
	if err == sql.ErrNoRows {
		authorInsert := `INSERT INTO author (name) VALUES (?);`
		res, err := db.Exec(authorInsert, authorName)
		if err != nil {
			return 0, err
		}
		authorId, err := res.LastInsertId()
		if err != nil {
			return 0, err
		}
		return int(authorId), err
	} else if err != nil {
		log.Fatalf("could not read author table: %v", err)
	}

	return authorId, nil
}

func mapBookWithAuthor(db *sqlx.DB, bookId, authorId int) error {
	authorToBookInsert := `INSERT INTO book_author (book_id, author_id) VALUES (?, ?);`
	_, err := db.Exec(authorToBookInsert, bookId, authorId)
	return err
}

func readRequestData(r *http.Request) (*addBookRequestBody, error) {
	decoder := json.NewDecoder(r.Body)
	var data addBookRequestBody
	err := decoder.Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("malformed json data: %w", err)
	}
	validate = validator.New()
	err = validate.Struct(data)
	if err != nil {
		return nil, fmt.Errorf("invalid json data: %w", err)
	}

	return &data, nil
}

func (h *BookHandler) AddBook(w http.ResponseWriter, r *http.Request) {
	data, err := readRequestData(r)
	if err != nil {
		badRequest(w, fmt.Sprintf("error happened during reading request body: %v", err))
		return
	}
	bookId, err := getOrCreateBookId(h.Env.DB, data)
	if err != nil {
		internalServerError(w, fmt.Sprintf("error happened during creating book: %v", err))
		return
	}

	for _, authorName := range data.Authors {
		authorId, err := getOrCreateAuthorId(h.Env.DB, authorName)

		if err != nil {
			internalServerError(w, fmt.Sprintf("error happened during fetching author: %v", err))
			return
		}

		err = mapBookWithAuthor(h.Env.DB, bookId, authorId)
		if err != nil {
			internalServerError(w, fmt.Sprintf("error happened during mapping book to author: %v", err))
			return
		}
	}

}

func getBookById(db *sqlx.DB, id int) (*models.Book, error) {
	query := "SELECT * FROM book WHERE id = ?"
	var book models.Book
	err := db.Get(&book, query, id)
	if err != nil {
		return nil, err
	}
	return &book, nil
}

type updateBookRequestBody struct {
	Name        *string  `json:"name"`
	Authors     []string `json:"authors"`
	Read        *bool    `json:"read"`
	Edition     *string  `json:"edition"`
	Description *string  `json:"description"`
}

func readUpdateRequestData(r *http.Request) (*updateBookRequestBody, error) {
	decoder := json.NewDecoder(r.Body)
	var data updateBookRequestBody
	err := decoder.Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("malformed json data: %w", err)
	}
	validate = validator.New()
	err = validate.Struct(data)
	if err != nil {
		return nil, fmt.Errorf("invalid json data: %w", err)
	}

	return &data, nil
}

func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	bookIdParam := chi.URLParam(r, "bookId")
	bookId, err := strconv.Atoi(bookIdParam)
	if err != nil {
		badRequest(w, fmt.Sprintf("could not convert param: %v", err))
		return
	}

	book, err := getBookById(h.Env.DB, bookId)
	if err != nil {
		notFound(w, fmt.Sprintf("could not get book by id: %v", err))
		return
	}

	data, err := readUpdateRequestData(r)
	if err != nil {
		badRequest(w, fmt.Sprintf("could not read request data: %v", data))
		return
	}

	if data.Description != nil {
		book.Description = data.Description
	}
	if data.Name != nil {
		book.BookName = *data.Name
	}
	if data.Read != nil {
		book.Read = *data.Read
	}
	if data.Edition != nil {
		book.Edition = data.Edition
	}

	query := "UPDATE book SET name = ?, read = ?, edition = ?, description = ? WHERE id = ?"
	_, err = h.Env.DB.Exec(query, book.BookName, book.Read, book.Edition, book.Description, bookId)
	if err != nil {
		internalServerError(w, fmt.Sprintf("could not update book: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
