package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/schema"

	"github.com/go-playground/validator"

	"github.com/wmolicki/bookler/context"
	"github.com/wmolicki/bookler/helpers"
	"github.com/wmolicki/bookler/models"
	"github.com/wmolicki/bookler/views"
)

const maxMultiPartMemory = 5 * 1 << 20 // 5 MB

var validate *validator.Validate
var decoder = schema.NewDecoder()

type BookHandler struct {
	bs models.BookService
	ba *models.BookAuthorService
	as *models.AuthorService
	ub *models.UserBookService

	listView    *views.View
	addView     *views.View
	editView    *views.View
	detailsView *views.View
}

func NewBookHandler(as *models.AuthorService, ba *models.BookAuthorService, bs models.BookService, ub *models.UserBookService) *BookHandler {
	listView := views.NewView("bulma", "templates/books.gohtml")
	addView := views.NewView("bulma", "templates/book_add.gohtml")
	editView := views.NewView("bulma", "templates/book_edit.gohtml")
	detailsView := views.NewView("bulma", "templates/book_details.gohtml")

	decoder.IgnoreUnknownKeys(true)

	// bs.DestructiveReset()
	return &BookHandler{bs, ba, as, ub, listView, addView, editView, detailsView}
}

type BooksViewModel struct {
	Books []*models.Book
}

type BookViewModel struct {
	models.Book
}

func (h *BookHandler) List(w http.ResponseWriter, r *http.Request) {
	books, err := h.bs.List()
	if err != nil {
		internalServerError(w, fmt.Sprintf("could not load books: %v", err))
		return
	}
	h.listView.Render(w, r, BooksViewModel{Books: books})
}

func (h *BookHandler) Details(w http.ResponseWriter, r *http.Request) {
	bookId, err := helpers.ParseUintParam(r, "bookId")
	if err != nil {
		badRequest(w, fmt.Sprintf("could not convert param: %v", err))
		return
	}

	book, err := h.ba.GetBookByID(bookId)
	if err != nil {
		// TODO: switch on error type
		log.Errorf("error getting book: %v", err)
		http.Error(w, "could not get book", http.StatusInternalServerError)
		return
	}

	authorsSl := make([]string, 0, len(book.Authors))
	for _, a := range book.Authors {
		authorsSl = append(authorsSl, a.Name)
	}
	authors := strings.Join(authorsSl, ", ")

	viewModel := struct {
		EditBookFormData
		Name        string
		Authors     string
		Description string
		ID          uint
		Tags        []string
		Image       string
		Ratings     []int
	}{
		Ratings:     []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		Name:        book.Name,
		Authors:     authors,
		Description: book.Description,
		ID:          book.ID,
		Tags:        book.Tags,
		Image:       book.Image,
	}

	if viewModel.Image == "" {
		viewModel.Image = "320x480.png"
	}

	user := context.User(r.Context())
	if user != nil {
		userBook, err := h.ub.GetUserBook(book, user)

		switch err {
		case nil:
			viewModel.Read = userBook.Read
			var rating int
			if userBook.Rating != nil {
				rating = *userBook.Rating
			} else {
				rating = -1
			}
			viewModel.Rating = rating
		case models.ErrorEntityNotFound:
			// TODO: ugly hack for now (default (0) value should mean no rating, instead
			// of having -1... maybe change no rating value to 0, and have minimal rating as 1/10
			viewModel.Rating = -1
		default:
			log.Errorf("could not load user book profile: %v", err)
			internalServerError(w, "could not load user book profile")
			return
		}
	}

	h.detailsView.Render(w, r, &viewModel)
	return
}

func (h *BookHandler) Edit(w http.ResponseWriter, r *http.Request) {
	bookId, err := helpers.ParseUintParam(r, "bookId")
	if err != nil {
		badRequest(w, fmt.Sprintf("could not convert param: %v", err))
		return
	}

	book, err := h.ba.GetBookByID(bookId)
	if err != nil {
		// TODO: switch on error type
		log.Errorf("error getting book: %v", err)
		http.Error(w, "could not get book", http.StatusInternalServerError)
		return
	}

	authorsSl := make([]string, 0, len(book.Authors))
	for _, a := range book.Authors {
		authorsSl = append(authorsSl, a.Name)
	}
	authors := strings.Join(authorsSl, ", ")

	viewModel := struct {
		EditBookFormData
		Name        string
		Authors     string
		Description string
		ID          uint
		Tags        string
		Image       string
		Ratings     []int
	}{
		Ratings:     []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		Name:        book.Name,
		Authors:     authors,
		Description: book.Description,
		ID:          book.ID,
		Tags:        strings.Join(book.Tags, ","),
	}

	user := context.User(r.Context())
	if user != nil {
		userBook, err := h.ub.GetUserBook(book, user)

		switch err {
		case nil:
			viewModel.Read = userBook.Read
			var rating int
			if userBook.Rating != nil {
				rating = *userBook.Rating
			} else {
				rating = -1
			}
			viewModel.Rating = rating
		case models.ErrorEntityNotFound:
			// TODO: ugly hack for now (default (0) value should mean no rating, instead
			// of having -1... maybe change no rating value to 0, and have minimal rating as 1/10
			viewModel.Rating = -1
		default:
			log.Errorf("could not load user book profile: %v", err)
			internalServerError(w, "could not load user book profile")
			return
		}
	}

	h.editView.Render(w, r, &viewModel)
	return
}

func (h *BookHandler) HandleEdit(w http.ResponseWriter, r *http.Request) {
	bookId, err := helpers.ParseUintParam(r, "bookId")
	if err != nil {
		badRequest(w, fmt.Sprintf("could not convert param: %v", err))
		return
	}

	book, err := h.bs.ByID(bookId)
	if err != nil {
		// TODO: switch on error type
		log.Errorf("could not get book: %v", err)
		http.Error(w, "could not get book", http.StatusInternalServerError)
		return
	}

	user := context.User(r.Context())
	if user == nil {
		panic("signed in only!")
	}

	err = r.ParseMultipartForm(maxMultiPartMemory)
	if err != nil {
		badRequest(w, err.Error())
	}

	formF, fh, err := r.FormFile("book-img")
	if err == http.ErrMissingFile {
		// do nothing
	} else if err != nil {
		log.Errorf("error getting file from form: %v", err)
		http.Error(w, "error processing the form", http.StatusInternalServerError)
		return
	} else {
		hashed, err := helpers.StoreFile(fh.Filename, formF)
		book.Image = hashed
		if err != nil {
			log.Errorf("error writing file to disk: %v", err)
			http.Error(w, "error processing the form", http.StatusInternalServerError)
			return
		}
	}

	var data EditBookFormData
	err = decoder.Decode(&data, r.PostForm)
	if err != nil {
		panic(err)
	}

	book.Name = data.Name
	book.Description = data.Description
	book.Edition = data.Edition
	book.Tags = helpers.CommaSplit(data.Tags)

	err = h.bs.Update(book)
	if err != nil {
		log.Errorf("error updating book: %v", err)
		http.Error(w, "error updating book", http.StatusInternalServerError)
		return
	}

	authors := splitAuthors(data.Authors)
	err = h.ba.UpdateBookAuthors(book, authors)
	if err != nil {
		log.Errorf("could not load update book authors: %v", err)
		http.Error(w, "could not update book authors", http.StatusInternalServerError)
		return
	}

	if _, err := h.ub.Rate(book, user, data.Rating); err != nil {
		log.Errorf("could not rate book: %v", err)
		http.Error(w, "could not rate book", http.StatusInternalServerError)
		return
	}

	if _, err = h.ub.Read(book, user, data.Read); err != nil {
		log.Errorf("could not mark book as read: %v", err)
		http.Error(w, "could not mark book as read", http.StatusInternalServerError)
		return
	}

	views.FlashSuccess(w, "Book edited successfully.")
	http.Redirect(w, r, fmt.Sprintf("/books/%d", bookId), http.StatusFound)
}

func (h *BookHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	bookId, err := helpers.ParseUintParam(r, "bookId")
	if err != nil {
		badRequest(w, fmt.Sprintf("could not convert param: %v", err))
		return
	}

	book, err := h.bs.ByID(bookId)
	if err != nil {
		// TODO: switch on error type
		log.Errorf("could not get book: %v", err)
		http.Error(w, "could not get book", http.StatusInternalServerError)
		return
	}

	err = h.bs.Delete(book)
	if err != nil {
		log.Errorf("could not delete book: %v", err)
		http.Error(w, "could not delete book", http.StatusInternalServerError)
		return
	}

	views.FlashSuccess(w, "Book deleted successfully.")
	http.Redirect(w, r, "/books", http.StatusFound)
}

func (h *BookHandler) Add(w http.ResponseWriter, r *http.Request) {
	h.addView.Render(w, r, nil)
}

type AddBookFormData struct {
	Name        string `schema:"name,required"`
	Authors     string `schema:"authors,required"`
	Description string `schema:"description"`
	Tags        string `schema:"tags"`
}

type EditBookFormData struct {
	ID          uint   `schema:"id,required"`
	Name        string `schema:"name,required"`
	Authors     string `schema:"authors,required"`
	Description string `schema:"description"`
	Edition     string `schema:"edition"`
	Rating      int    `schema:"rating"`
	Read        bool   `schema:"read"`
	Tags        string `schema:"tags"`
	Image       string `schema:"img"`
}

func splitAuthors(authors string) []string {
	authorsSl := strings.Split(authors, ",")
	for i, a := range authorsSl {
		authorsSl[i] = strings.TrimSpace(a)
	}
	return authorsSl
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

	authors := splitAuthors(data.Authors)

	book, err := h.bs.New(data.Name, data.Description, "", authors)
	if err != nil {
		log.Errorf("could not create book: %v", err)
		http.Error(w, "error creating book", http.StatusInternalServerError)
		return
	}

	if err := h.ba.UpdateBookAuthors(book, authors); err != nil {
		log.Errorf("could not map book too authors: %v", err)
		http.Error(w, "error happened mapping book to authors", http.StatusInternalServerError)
		return
	}

	views.FlashSuccess(w, "Book added successfully.")
	http.Redirect(w, r, fmt.Sprintf("/books/%d", book.ID), http.StatusFound)
}

func (h *BookHandler) Index(w http.ResponseWriter, r *http.Request) {
	books, err := h.bs.List()
	if err != nil {
		log.Errorf("could not load books: %v", err)
		internalServerError(w, "could not load books")
		return
	}
	//_ := template.New("templates/books.gohtml")

	b, _ := json.Marshal(books)

	w.Write(b)
}

type addBookRequestBody struct {
	Name        string `validate:"required"`
	Authors     string `json:"authors" validate:"required"`
	Description string `validate:"required"`
	ImageURL    string `json:"image_url" validate:"required"`
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

	authors := splitAuthors(data.Authors)
	book, err := h.bs.New(data.Name, data.Description, "", authors)
	if err != nil {
		handleCreateBookError(w, err)
		return
	}

	resp, err := http.Get(data.ImageURL)
	if err != nil {
		handleCreateBookError(w, err)
		return
	}
	defer resp.Body.Close()
	filename := path.Base(data.ImageURL)
	hashed, err := helpers.StoreFile(filename, resp.Body)
	if err != nil {
		handleCreateBookError(w, err)
		return
	}

	book.Image = hashed
	err = h.bs.Update(book)
	if err != nil {
		handleCreateBookError(w, err)
		return
	}
	err = h.ba.UpdateBookAuthors(book, authors)

	if err != nil {
		log.Errorf("could not map book to authors: %v", err)
		internalServerError(w, "error happened mapping book to authors")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func handleCreateBookError(w http.ResponseWriter, err error) {
	log.Errorf("error creating book: %v", err)
	internalServerError(w, "error happened creating book")
}

type updateBookRequestBody struct {
	Name        string   `json:"name"`
	Authors     []string `json:"authors"`
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
	bookId, err := helpers.ParseUintParam(r, "bookId")
	if err != nil {
		badRequest(w, fmt.Sprintf("could not convert param: %v", err))
		return
	}

	book, err := h.bs.ByID(bookId)
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
	book.Edition = data.Edition
	book.Description = data.Description

	err = h.bs.Update(book)
	if err != nil {
		internalServerError(w, fmt.Sprintf("could not update book: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
