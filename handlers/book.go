package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/schema"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator"

	"github.com/wmolicki/bookler/models"
	"github.com/wmolicki/bookler/views"
)

var validate *validator.Validate
var decoder = schema.NewDecoder()

type BookHandler struct {
	bs       *models.BookService
	as       *models.AuthorService
	listView *views.View
	addView  *views.View
	editView *views.View
}

func NewBookHandler(as *models.AuthorService, bs *models.BookService) *BookHandler {
	listView := views.NewView("bulma", "templates/books.gohtml")
	addView := views.NewView("bulma", "templates/book_add.gohtml")
	editView := views.NewView("bulma", "templates/book_edit.gohtml")

	decoder.IgnoreUnknownKeys(true)

	bs.DestructiveReset()
	return &BookHandler{bs, as, listView, addView, editView}
}

type BooksViewModel struct {
	Books []*models.Book
}

type BookViewModel struct {
	models.Book
}

func (h *BookHandler) List(w http.ResponseWriter, r *http.Request) {
	books, err := h.bs.GetList()
	if err != nil {
		internalServerError(w, fmt.Sprintf("could not load books: %v", err))
		return
	}
	h.listView.Render(w, r, BooksViewModel{Books: books})
}

func (h *BookHandler) Edit(w http.ResponseWriter, r *http.Request) {
	bookId, err := parseIdParam("bookId", r)
	if err != nil {
		badRequest(w, fmt.Sprintf("could not convert param: %v", err))
		return
	}

	book, err := h.bs.GetBookById(bookId)
	if err != nil {
		// TODO: switch on error type
		http.Error(w, "could not get book", http.StatusInternalServerError)
		return
	}
	viewModel := EditBookFormData{
		Name:        book.Name,
		Authors:     book.Authors,
		Description: book.Description,
		Read:        book.Read,
		ID:          book.ID,
	}

	h.editView.Render(w, r, &viewModel)
	return
}

func (h *BookHandler) HandleEdit(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (h *BookHandler) Add(w http.ResponseWriter, r *http.Request) {
	h.addView.Render(w, r, nil)
}

type AddBookFormData struct {
	Name        string `schema:"name,required"`
	Author      string `schema:"author,required"`
	Description string `schema:"description,required"`
}

type EditBookFormData struct {
	ID          uint                 `schema:"id,required"`
	Name        string               `schema:"name,required"`
	Authors     []*models.BookAuthor `schema:"author,required"`
	Description string               `schema:"description,required"`
	Read        bool                 `schema:"read,required"`
}

func (h *BookHandler) HandleAdd(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		badRequest(w, err.Error())
	}
	var data AddBookFormData
	err = decoder.Decode(&data, r.PostForm)
	if err != nil {
		panic(err)
	}

	book, err := h.bs.New(data.Name, data.Description, "", data.Author, false)
	if err != nil {
		http.Error(w, "error creating book", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/books/%d", book.ID), http.StatusFound)

}

func (h *BookHandler) Index(w http.ResponseWriter, r *http.Request) {
	books, err := h.bs.GetList()
	if err != nil {
		internalServerError(w, fmt.Sprintf("could not load books: %v", err))
		return
	}
	//_ := template.New("templates/books.gohtml")

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

	_, err = h.bs.New(data.Name, data.Description, data.Edition, data.Authors[0], data.Read)

	if err != nil {
		internalServerError(w, fmt.Sprintf("error happened mapping book to author: %v", err))
		return
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

func parseIdParam(param string, r *http.Request) (uint, error) {
	idParam := chi.URLParam(r, "bookId")
	parsedBookId, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(parsedBookId), nil
}

func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	bookId, err := parseIdParam("bookId", r)
	if err != nil {
		badRequest(w, fmt.Sprintf("could not convert param: %v", err))
		return
	}

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
