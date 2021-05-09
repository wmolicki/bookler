package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/wmolicki/bookler/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator"
	"github.com/wmolicki/bookler/models"
)

var validate *validator.Validate

type BookHandler struct {
	bs *models.BookService
	as *models.AuthorService
}

func NewBookHandler(env *config.Env) *BookHandler {
	bs := models.NewBookService(env)
	as := models.NewAuthorService(env)
	bs.DestructiveReset()
	return &BookHandler{bs, as}
}

func (h *BookHandler) Index(w http.ResponseWriter, r *http.Request) {
	books, err := h.bs.GetList()
	if err != nil {
		internalServerError(w, fmt.Sprintf("could not load books: %v", err))
		return
	}

	b, _ := json.Marshal(books)

	w.Write(b)
}

type addBookRequestBody struct {
	Name        string   `validate:"required"`
	Authors     []string `json:"authors" validate:"required"`
	Read        bool     `validate:"required"`
	Edition     string   `validate:"required"`
	Description string   `validate:"required"`
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
	// TODO: transaction ?
	data, err := readRequestData(r)
	if err != nil {
		badRequest(w, fmt.Sprintf("error happened during reading request body: %v", err))
		return
	}

	b := models.Book{Name: data.Name, Read: data.Read, Description: data.Description, Edition: data.Edition}
	book, err := h.bs.Create(&b)

	if err != nil {
		internalServerError(w, fmt.Sprintf("error happened during creating book: %v", err))
		return
	}

	for _, authorName := range data.Authors {
		author, err := h.as.GetByName(authorName)
		if err != nil {
			switch err {
			case models.ErrorEntityNotFound:
				a := models.Author{Name: authorName}
				author, err = h.as.Create(&a)
			default:
				internalServerError(w, fmt.Sprintf("error happened during retrieving author: %v", err))
				return
			}
		}

		if err := h.bs.AddAuthor(book, author); err != nil {
			internalServerError(w, fmt.Sprintf("error happened mapping book to author: %v", err))
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
}

type updateBookRequestBody struct {
	Name        string   `json:"name"`
	Authors     []string `json:"authors"`
	Read        bool     `json:"read"`
	Edition     string   `json:"edition"`
	Description string   `json:"description"`
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
	parsedBookId, err := strconv.ParseUint(bookIdParam, 10, 64)
	if err != nil {
		badRequest(w, fmt.Sprintf("could not convert param: %v", err))
		return
	}
	bookId := uint(parsedBookId)

	book, err := h.bs.GetBookById(bookId)
	if err != nil {
		notFound(w, fmt.Sprintf("could not get book by id: %v", err))
		return
	}

	data, err := readUpdateRequestData(r)
	if err != nil {
		badRequest(w, fmt.Sprintf("could not read request data: %v", data))
		return
	}

	book.Name = data.Name
	book.Read = data.Read
	book.Edition = data.Edition
	book.Description = data.Description

	err = h.bs.Update(book)
	if err != nil {
		internalServerError(w, fmt.Sprintf("could not update book: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
